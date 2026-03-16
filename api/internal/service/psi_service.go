// api/internal/service/psi_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
// Este archivo contiene las operaciones centrales relacionadas con los psicólogos colegiados,
// incluyendo la importación masiva desde CSV, la gestión de perfiles públicos y la autenticación.

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"golang.org/x/crypto/bcrypt"
)

// PsiService es la estructura que implementa la lógica de negocio relacionada con los psicólogos.
type PsiService struct {
	repo        domain.PsiUserRepository
	s3Client    *s3.S3Client
	mailService IMailService       // Inyectamos el servicio de correo para notificaciones
	sanitizer   *bluemonday.Policy // Política de sanitización para biografías (XSS Protection)
}

// NewPsiService es el constructor de PsiService, inyectando las dependencias necesarias.
func NewPsiService(repo domain.PsiUserRepository, s3Client *s3.S3Client, mailService IMailService) *PsiService {
	policy := bluemonday.UGCPolicy()
	policy.AllowStyles("text-align").OnElements("p", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "li", "blockquote")

	return &PsiService{
		repo:        repo,
		s3Client:    s3Client,
		mailService: mailService,
		sanitizer:   policy,
	}
}

// =========================================================================
// GESTIÓN MASIVA (CSV IMPORT)
// =========================================================================

// ImportFromCSV procesa la carga masiva de agremiados desde un flujo de datos TSV/CSV.
// Implementa geo-normalización, hashing de credenciales y escritura transaccional
// atómica por cada registro. Los fallos individuales no interrumpen el lote completo.
//
// Columnas esperadas (separadas por tab, índice base 0):
//
//	[0]  UserName                  [15] PublicPhone
//	[1]  Email                     [16] ShowPublicPhone
//	[2]  Password                  [17] ServiceAddress
//	[3]  FirstName                 [18] ShowPublicServiceAddress
//	[4]  SecondName                [19] Solvent
//	[5]  LastName                  [20] MunicipalityCarabobo
//	[6]  SecondLastName            [21] PhoneCarabobo
//	[7]  FPV                       [22] CelPhoneCarabobo
//	[8]  CI                        [23] State (fuera de Carabobo)
//	[9]  Letter (ignorado)         [24] MunicipalityOutSideCarabobo
//	[10] Nationality               [25] PhoneOutSideCarabobo
//	[11] BornDate                  [26] CelPhoneOutSideCarabobo
//	[12] Genre                     [27] UniversityUndergraduate
//	[13] ContactEmail              [28] GraduateDate
//	[14] ShowContactEmail          [29] MentionUndergraduate
//	[30] RegisterTitleState        [37] GuildCollaborator
//	[31] RegisterTitleDate         [38] PublicEmployee
//	[32] RegisterNumber            [39] UniversityProfessor
//	[33] RegisterFolio             [40] DateOfLastSolvency
//	[34] RegisterTome              [41] DoubleGuild
//	[35] GuildDirector             [42] CPSM
//	[36] SixtyFiveOrPlus
func (s *PsiService) ImportFromCSV(ctx context.Context, reader io.Reader, adminID uuid.UUID) (int, []map[string]string) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = '\t'
	csvReader.LazyQuotes = true
	_, _ = csvReader.Read() // saltar cabeceras

	successCount := 0
	var failedRecords []map[string]string

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(record) < 43 {
			failedRecords = append(failedRecords, map[string]string{
				"fila":   safeGet(record, 0),
				"nombre": safeGet(record, 3) + " " + safeGet(record, 5),
				"ci":     safeGet(record, 8),
				"fpv":    safeGet(record, 7),
				"error":  fmt.Sprintf("columnas insuficientes: se esperaban 43, se recibieron %d", len(record)),
			})
			continue
		}

		// ── Geo-normalización ─────────────────────────────────────────────
		// Los valores "-" o vacíos se omiten para no disparar error de validación.
		var municipioCarabobo string
		if raw := strings.TrimSpace(record[20]); raw != "" && raw != "-" {
			mun, ok := utils.NormalizeMunicipioCarabobo(raw)
			if !ok {
				failedRecords = append(failedRecords, map[string]string{
					"fila":   record[0],
					"nombre": record[3] + " " + record[5],
					"ci":     record[8],
					"fpv":    record[7],
					"error":  fmt.Sprintf("municipio de Carabobo inválido: %q", raw),
				})
				continue
			}
			municipioCarabobo = mun
		}

		var estadoOutside string
		if raw := strings.TrimSpace(record[23]); raw != "" && raw != "-" {
			estado, ok := utils.NormalizeEstadoVenezuela(raw)
			if !ok {
				// Si no es estado venezolano, se trata como país extranjero (exterior)
				// Se guarda tal cual sin validación de catálogo.
				estadoOutside = raw
			} else {
				estadoOutside = estado
			}
		}

		// ── Hash de contraseña ────────────────────────────────────────────
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(record[2]), bcrypt.DefaultCost)
		if err != nil {
			failedRecords = append(failedRecords, map[string]string{
				"fila":   record[0],
				"nombre": record[3] + " " + record[5],
				"ci":     record[8],
				"fpv":    record[7],
				"error":  "error al procesar seguridad: " + err.Error(),
			})
			continue
		}

		psiID := uuid.New()
		audit := domain.AuditModel{
			CreateById: &adminID,
			CreateBy:   "Admin_CSV_Import",
			UpdateById: &adminID,
			UpdateBy:   "Admin_CSV_Import",
		}

		// ── Modelo Principal ──────────────────────────────────────────────
		psi := &domain.PsiUserModel{
			ID:         psiID,
			Key:        uuid.New().String(),
			AuditModel: audit,

			// Credenciales
			Username: record[0],
			Email:    record[1],
			Password: string(hashedPassword),

			// Identidad Legal
			FirstName:      record[3],
			SecondName:     cleanDash(record[4]),
			LastName:       record[5],
			SecondLastName: cleanDash(record[6]),
			FPV:            parseInt(record[7]),
			CI:             parseInt(record[8]),
			Nationality:    record[10],
			BornDate:       parseDate(record[11]),
			Genre:          record[12],

			// Contacto
			ContactEmail:             record[13],
			ShowContactEmail:         parseBool(record[14]),
			PublicPhone:              cleanDash(record[15]),
			ShowPublicPhone:          parseBool(record[16]),
			ServiceAddress:           cleanDash(record[17]),
			ShowPublicServiceAddress: parseBool(record[18]),

			// Estatus
			Solvent: parseBool(record[19]),

			// Ubicación: Carabobo
			MunicipalityCarabobo: municipioCarabobo,
			PhoneCarabobo:        cleanDash(record[21]),
			CelPhoneCarabobo:     cleanDash(record[22]),

			// Ubicación: Fuera de Carabobo / Venezuela / Exterior
			StateOutside:                cleanDash(estadoOutside),
			MunicipalityOutSideCarabobo: cleanDash(record[24]),
			PhoneOutSideCarabobo:        cleanDash(record[25]),
			CelPhoneOutSideCarabobo:     cleanDash(record[26]),
		}

		// ── Datos Colegiales ──────────────────────────────────────────────
		colData := &domain.PsiUserColData{
			ID:             uuid.New(),
			PsiUserModelID: psiID,
			AuditModel:     audit,

			// Pregrado
			UniversityUndergraduate: record[27],
			GraduateDate:            parseDate(record[28]),
			MentionUndergraduate:    record[29],

			// Registro legal del título
			RegisterTitleState: record[30],
			RegisterTitleDate:  parseDate(record[31]),
			RegisterNumber:     parseInt(record[32]),
			RegisterFolio:      cleanDash(record[33]),
			RegisterTome:       cleanDash(record[34]),

			// Flags gremiales
			GuildDirector:       parseBool(record[35]),
			SixtyFiveOrPlus:     parseBool(record[36]),
			GuildCollaborator:   parseBool(record[37]),
			PublicEmployee:      parseBool(record[38]),
			UniversityProfessor: parseBool(record[39]),

			// Historial gremial
			DateOfLastSolvency: parseDate(record[40]),
			DoubleGuild:        parseBool(record[41]),
			CPSM:               parseBool(record[42]),
		}

		// ── Persistencia transaccional ────────────────────────────────────
		if err := s.repo.CreateWithColData(ctx, psi, colData); err != nil {
			failedRecords = append(failedRecords, map[string]string{
				"fila":   record[0],
				"nombre": record[3] + " " + record[5],
				"ci":     record[8],
				"fpv":    record[7],
				"error":  MapDBError(err).Error(),
			})
			continue
		}

		// ── Notificación de bienvenida (no bloqueante) ────────────────────
		mailData := map[string]interface{}{
			"Name":     psi.Username,
			"Email":    psi.Email,
			"Password": record[2], // contraseña en claro — solo en el correo inicial
		}
		if err := s.mailService.SendEmail(psi.Email, "Bienvenido a la plataforma Colegio de Psicólogos", "welcome_psi", mailData); err != nil {
			log.Printf("⚠️ Error al enviar correo de bienvenida [%s]: %v", psi.Username, err)
		}

		successCount++
	}

	return successCount, failedRecords
}

// ── Helpers de parseo ─────────────────────────────────────────────────────────

// safeGet retorna el elemento en el índice dado, o "" si el slice es más corto.
func safeGet(record []string, i int) string {
	if i < len(record) {
		return record[i]
	}
	return ""
}

// cleanDash convierte el placeholder "-" (o variantes con espacios) en string vacío.
func cleanDash(s string) string {
	if strings.TrimSpace(s) == "-" {
		return ""
	}
	return strings.TrimSpace(s)
}

// =========================================================================
// AUTOGESTIÓN Y PRIVACIDAD (SELF-MANAGEMENT)
// =========================================================================

// UpdateProfileSelf permite al psicólogo actualizar sus datos de contacto y visibilidad.
// Implementa "Lazy Loading" para ColData: solo consulta y actualiza la tabla de datos
// colegiales si el usuario solicita cambios en esos campos específicos.
// UpdateProfileSelf permite al psicólogo actualizar sus datos de contacto y visibilidad.
// UpdateProfileSelf procesa la autogestión del perfil del psicólogo.
// Implementa actualizaciones parciales, manejo seguro de binarios y sanitización XSS.
func (s *PsiService) UpdateProfileSelf(
	ctx context.Context,
	psi *domain.PsiUserModel,
	id uuid.UUID,
	req request_structs.PsiUserUpdateRequestSelf,
	profilePic *multipart.FileHeader,
	titleImgOne *multipart.FileHeader,
	titleImgTwo *multipart.FileHeader,
	titleImgThree *multipart.FileHeader,
) (*domain.PsiUserModel, error) {

	// 1. Validación de contraseña actual (Security First)
	if err := bcrypt.CompareHashAndPassword([]byte(psi.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("contraseña actual incorrecta")
	}

	// 2. Cambio de contraseña (si se solicitó)
	if req.NewPassword1 != nil && *req.NewPassword1 != "" {
		if req.NewPassword2 == nil || *req.NewPassword1 != *req.NewPassword2 {
			return nil, errors.New("las nuevas contraseñas no coinciden")
		}
		if !utils.IsStrongPassword(*req.NewPassword1) {
			return nil, errors.New("la nueva contraseña no cumple los requisitos de seguridad")
		}
		hashed, _ := bcrypt.GenerateFromPassword([]byte(*req.NewPassword1), bcrypt.DefaultCost)
		psi.Password = string(hashed)
		psi.Key = uuid.New().String()
	}

	var uploadedS3Keys []string
	var oldS3KeysToDelete []string

	// 3. Imagen de perfil
	if profilePic != nil {
		src, _ := profilePic.Open()
		defer src.Close()
		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return nil, err
		}
		filename := fmt.Sprintf("%s%s", psi.ID.String(), ext)
		newKey, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "avatars", filename, contentType)
		if err != nil {
			return nil, err
		}
		uploadedS3Keys = append(uploadedS3Keys, newKey)
		if psi.ProfilePictureS3Key != "" && psi.ProfilePictureS3Key != newKey {
			oldS3KeysToDelete = append(oldS3KeysToDelete, psi.ProfilePictureS3Key)
		}
		psi.ProfilePictureS3Key = newKey
	}

	// 4. Auditoría
	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID

	// 4a. Credenciales
	if req.Username != nil {
		validate_username := strings.ToLower(*req.Username)
		err := s.repo.ValidateUniqueCredentials(ctx, validate_username, "", psi.ID)
		if err != nil {
			return nil, err
		}
		psi.Username = *req.Username
	}
	if req.Email != nil {
		validate_email, err := utils.ParseAndValidateEmail(*req.Email)
		if err != nil {
			return nil, err
		}
		err = s.repo.ValidateUniqueCredentials(ctx, "", validate_email, psi.ID)
		if err != nil {
			return nil, err
		}
		psi.Email = validate_email
	}

	// 4b. Contacto
	if req.ContactEmail != nil {
		validate_email, err := utils.ParseAndValidateEmail(*req.ContactEmail)
		if err != nil {
			return nil, err
		}
		psi.ContactEmail = validate_email
	}

	if req.PublicPhone != nil {
		psi.PublicPhone = *req.PublicPhone
	}
	if req.ServiceAddress != nil {
		psi.ServiceAddress = *req.ServiceAddress
	}

	if v := req.ShowContactEmail(); v != nil {
		psi.ShowContactEmail = *v
	}
	if v := req.ShowPublicPhone(); v != nil {
		psi.ShowPublicPhone = *v
	}
	if v := req.ShowPublicServiceAddress(); v != nil {
		psi.ShowPublicServiceAddress = *v
	}

	// 4c. Ubicación: Carabobo — geo-validación antes de asignar
	if req.MunicipalityCarabobo != nil {
		mun, ok := utils.NormalizeMunicipioCarabobo(*req.MunicipalityCarabobo)
		if !ok {
			return nil, fmt.Errorf("municipio de Carabobo inválido: %q", *req.MunicipalityCarabobo)
		}
		psi.MunicipalityCarabobo = mun
	}
	if req.PhoneCarabobo != nil {
		psi.PhoneCarabobo = *req.PhoneCarabobo
	}
	if req.CelPhoneCarabobo != nil {
		psi.CelPhoneCarabobo = *req.CelPhoneCarabobo
	}

	// 4d. Ubicación: Fuera de Carabobo (Venezuela)
	if req.StateOutside != nil {
		estado, ok := utils.NormalizeEstadoVenezuela(*req.StateOutside)
		if !ok {
			return nil, fmt.Errorf("estado venezolano inválido o no permitido: %q", *req.StateOutside)
		}
		psi.StateOutside = estado
	}
	if req.MunicipalityOutSideCarabobo != nil {
		psi.MunicipalityOutSideCarabobo = *req.MunicipalityOutSideCarabobo
	}
	if req.PhoneOutSideCarabobo != nil {
		psi.PhoneOutSideCarabobo = *req.PhoneOutSideCarabobo
	}
	if req.CelPhoneOutSideCarabobo != nil {
		psi.CelPhoneOutSideCarabobo = *req.CelPhoneOutSideCarabobo
	}
	if req.ServiceAddressOutSideCarabobo != nil {
		psi.ServiceAddressOutSideCarabobo = *req.ServiceAddressOutSideCarabobo
	}

	if v := req.ShowPhoneOutSideCarabobo(); v != nil {
		psi.ShowPhoneOutSideCarabobo = *v
	}
	if v := req.ShowCellPhoneOutSideCarabobo(); v != nil {
		psi.ShowCellPhoneOutSideCarabobo = *v
	}
	if v := req.ShowPublicServiceAddressOutSideCarabobo(); v != nil {
		psi.ShowPublicServiceAddressOutSideCarabobo = *v
	}

	// 4e. Ubicación: Fuera de Venezuela
	// Country no tiene catálogo restringido — se acepta cualquier valor (ISO libre)
	if req.Country != nil {
		psi.Country = *req.Country
	}
	if req.PhoneOutSideVenezuela != nil {
		psi.PhoneOutSideVenezuela = *req.PhoneOutSideVenezuela
	}
	if req.ServiceAddressOutSideVenezuela != nil {
		psi.ServiceAddressOutSideVenezuela = *req.ServiceAddressOutSideVenezuela
	}

	if v := req.ShowPhoneOutSideVenezuela(); v != nil {
		psi.ShowPhoneOutSideVenezuela = *v
	}
	if v := req.ShowCellPhoneOutSideVenezuela(); v != nil {
		psi.ShowCellPhoneOutSideVenezuela = *v
	}
	if v := req.ShowPublicServiceAddressOutSideVenezuela(); v != nil {
		psi.ShowPublicServiceAddressOutSideVenezuela = *v
	}

	// 4f. Perfil profesional
	if req.PrimarySpecialty != nil {
		psi.PrimarySpecialty = *req.PrimarySpecialty
	}
	if req.SecondarySpecialty != nil {
		psi.SecondarySpecialty = *req.SecondarySpecialty
	}
	if req.MiniBio != nil {
		runes := []rune(*req.MiniBio)
		if len(runes) > 250 {
			psi.MiniBio = string(runes[:250])
		} else {
			psi.MiniBio = *req.MiniBio
		}
	}

	// 5. Biografía extensa (sanitización XSS)
	var bioTextToUpdate *domain.TextModel
	if req.FullBio != nil {
		cleanHTML := s.sanitizer.Sanitize(*req.FullBio)
		if psi.BioTextID != uuid.Nil {
			psi.FullBio.Content = cleanHTML
			psi.FullBio.UpdateBy = psi.Username
			psi.FullBio.UpdateById = &psi.ID
		} else {
			psi.FullBio = domain.TextModel{
				ID:      uuid.New(),
				Content: cleanHTML,
				AuditModel: domain.AuditModel{
					CreateBy: psi.Username, CreateById: &psi.ID,
					UpdateBy: psi.Username, UpdateById: &psi.ID,
				},
			}
			psi.BioTextID = psi.FullBio.ID
		}
		bioTextToUpdate = &psi.FullBio
	}

	// 6. Datos colegiales y títulos (Lazy Load)
	var colDataToUpdate *domain.PsiUserColData
	hasColDataChanges := req.ShowUniversityUndergraduateRaw != "" ||
		req.ShowGraduateDateRaw != "" ||
		req.ShowMentionUndergraduateRaw != "" ||
		titleImgOne != nil || titleImgTwo != nil || titleImgThree != nil

	if hasColDataChanges {
		currentColData, err := s.repo.GetPsiUserColData(ctx, psi.ID)
		if err != nil {
			return nil, err
		}

		if v := req.ShowUniversityUndergraduate(); v != nil {
			currentColData.ShowUniversityUndergraduate = *v
		}
		if v := req.ShowGraduateDate(); v != nil {
			currentColData.ShowGraduateDate = *v
		}
		if v := req.ShowMentionUndergraduate(); v != nil {
			currentColData.ShowMentionUndergraduate = *v
		}

		processTitleImage := func(file *multipart.FileHeader, orderNum string, oldKey string) (string, error) {
			if file == nil {
				return oldKey, nil
			}
			src, err := file.Open()
			if err != nil {
				return "", err
			}
			defer src.Close()
			cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
			if err != nil {
				return "", err
			}
			shortUUID := uuid.New().String()[:6]
			filename := fmt.Sprintf("%s_title_%s_%s%s", psi.ID.String(), orderNum, shortUUID, ext)
			newKey, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "titles", filename, contentType)
			if err != nil {
				return "", err
			}
			uploadedS3Keys = append(uploadedS3Keys, newKey)
			if oldKey != "" && oldKey != newKey {
				oldS3KeysToDelete = append(oldS3KeysToDelete, oldKey)
			}
			return newKey, nil
		}

		if newKey, err := processTitleImage(titleImgOne, "1", currentColData.TitleImageOneS3Key); err == nil {
			currentColData.TitleImageOneS3Key = newKey
		} else {
			return nil, err
		}
		if newKey, err := processTitleImage(titleImgTwo, "2", currentColData.TitleImageTwoS3Key); err == nil {
			currentColData.TitleImageTwoS3Key = newKey
		} else {
			return nil, err
		}
		if newKey, err := processTitleImage(titleImgThree, "3", currentColData.TitleImageThreeS3Key); err == nil {
			currentColData.TitleImageThreeS3Key = newKey
		} else {
			return nil, err
		}

		currentColData.UpdateBy = psi.Username
		currentColData.UpdateById = &psi.ID
		colDataToUpdate = currentColData
	}

	// 7. Persistencia — rollback S3 si falla la DB
	if err := s.repo.UpdatePublicProfile(ctx, psi, colDataToUpdate, bioTextToUpdate); err != nil {
		for _, key := range uploadedS3Keys {
			_ = s.s3Client.DeleteFile(context.Background(), key)
		}
		return nil, err
	}

	// Limpiar archivos S3 reemplazados solo tras confirmar la persistencia
	for _, oldKey := range oldS3KeysToDelete {
		_ = s.s3Client.DeleteFile(context.Background(), oldKey)
	}

	return psi, nil
}

// =========================================================================
// CONSULTA PÚBLICA (DIRECTORY & PROFILE)
// =========================================================================

// GetPublicDirectory devuelve una lista paginada de mini-perfiles.
// Aplica normalización de género y oculta datos sensibles de solvencia al público.
func (s *PsiService) GetPublicDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) (interface{}, error) {
	// Normalizar paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 12
	}

	filter.Gender = strings.ToUpper(strings.TrimSpace(filter.Gender))
	if filter.Gender != "M" && filter.Gender != "F" {
		filter.Gender = ""
	}

	users, total, err := s.repo.SearchDirectory(ctx, filter)
	if err != nil {
		return nil, err
	}

	list := make([]request_structs.PsiMiniProfileDTO, 0, len(users))
	for _, u := range users {
		mini := request_structs.PsiMiniProfileDTO{
			ID:             u.ID,
			FirstName:      u.FirstName,
			LastName:       u.LastName,
			CI:             u.CI,
			FPV:            u.FPV,
			ProfilePicture: u.ProfilePictureS3Key,
			MiniBio:        u.MiniBio,
			// Solvent:        u.Solvent, // no mostrar al publico
		}

		// Añadimos las especialidades al mini perfil para que aparezcan en las "cards"
		mini.Specialties = []string{}
		if u.PrimarySpecialty != "" {
			mini.Specialties = append(mini.Specialties, u.PrimarySpecialty)
		}
		if u.SecondarySpecialty != "" {
			mini.Specialties = append(mini.Specialties, u.SecondarySpecialty)
		}

		list = append(list, mini)
	}

	return fiber.Map{
		"data":        list,
		"total":       total,
		"page":        filter.Page,
		"limit":       filter.Limit,
		"total_pages": (total + int64(filter.Limit) - 1) / int64(filter.Limit),
	}, nil
}

// GetPublicProfile construye la ficha técnica del psicólogo aplicando el "Escudo de Privacidad".
// Los datos personales (Email, Teléfono) y académicos (Postgrados) se ocultan dinámicamente
// según la configuración del usuario y su estatus de solvencia institucional.
// GetPublicProfile construye la ficha técnica del psicólogo aplicando el "Escudo de Privacidad"
// y la "Restricción de Solvencia": si el psicólogo no está al día con sus cuotas, el perfil
// público solo expone los datos esenciales de identidad y su universidad de pregrado.
func (s *PsiService) GetPublicProfile(ctx context.Context, id int) (*request_structs.PsiFullProfileDTO, uuid.UUID, error) {
	// 1. Obtener datos crudos de la DB
	psi, err := s.repo.GetByFPV(ctx, id)
	if err != nil {
		return nil, uuid.Nil, errors.New("psicólogo no encontrado")
	}

	// 2. Verificar si está activo
	if !psi.IsActive {
		return nil, uuid.Nil, errors.New("perfil no disponible")
	}

	// 3. RESTRICCIÓN DE SOLVENCIA — early return con datos mínimos
	if !psi.Solvent {
		return &request_structs.PsiFullProfileDTO{
			FirstName:      psi.FirstName,
			SecondName:     psi.SecondName,
			LastName:       psi.LastName,
			SecondLastName: psi.SecondLastName,
			FPV:            psi.FPV,
			CI:             psi.CI,
			Gender:         psi.Genre,
			ProfilePicture: psi.ProfilePictureS3Key,
			Solvent:        false,
			Undergraduate: request_structs.UndergraduateDTO{
				University: psi.ColData.UniversityUndergraduate,
			},
			Specialties:    make([]string, 0),
			PostGrades:     make([]request_structs.PostGradeDTO, 0),
			SocialNetworks: make([]request_structs.SocialNetworkDTO, 0),
		}, uuid.Nil, nil
	}

	// A partir de aquí: psicólogo solvente — perfil completo con Privacy Shield

	// 4. Obtener biografía extensa
	fullBio, err := s.repo.GetTextContentByID(ctx, psi.BioTextID)
	if err != nil {
		log.Printf("⚠️ Error al obtener la biografía extensa del psicólogo %d: %v", id, err)
	}

	// 5. Inicializar DTO
	dto := &request_structs.PsiFullProfileDTO{
		FirstName:      psi.FirstName,
		SecondName:     psi.SecondName,
		LastName:       psi.LastName,
		SecondLastName: psi.SecondLastName,
		FPV:            psi.FPV,
		CI:             psi.CI,
		Gender:         psi.Genre,
		ProfilePicture: psi.ProfilePictureS3Key,
		Solvent:        true,
		MiniBio:        psi.MiniBio,
		FullBioContent: fullBio,
		Specialties:    make([]string, 0),
		PostGrades:     make([]request_structs.PostGradeDTO, 0),
		SocialNetworks: make([]request_structs.SocialNetworkDTO, 0),
		Undergraduate:  request_structs.UndergraduateDTO{},
	}

	// ── Privacy Shield: Contacto principal ───────────────────────────────
	if psi.ShowContactEmail {
		dto.Email = psi.ContactEmail
	}
	if psi.ShowPublicPhone {
		dto.Phone = psi.PublicPhone
	}
	if psi.ShowPublicServiceAddress {
		dto.Address = psi.ServiceAddress
	}

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	if psi.MunicipalityCarabobo != "" {
		loc := &request_structs.PsiLocationCaraboboDTO{
			Municipality: psi.MunicipalityCarabobo,
		}
		if psi.ShowPhoneOutSideCarabobo {
			loc.Phone = psi.PhoneCarabobo
		}
		if psi.ShowCellPhoneOutSideCarabobo {
			loc.CellPhone = psi.CelPhoneCarabobo
		}
		if psi.ShowPublicServiceAddress {
			loc.Address = psi.ServiceAddress
		}
		dto.Location.Carabobo = loc
	}

	// ── Ubicación: Fuera de Carabobo (Venezuela) ──────────────────────────
	if psi.StateOutside != "" && (psi.ShowPublicServiceAddressOutSideCarabobo || psi.ShowCellPhoneOutSideCarabobo || psi.ShowPhoneOutSideCarabobo) {
		loc := &request_structs.PsiLocationVenezuelaDTO{
			State:        psi.StateOutside,
			Municipality: psi.MunicipalityOutSideCarabobo,
		}
		if psi.ShowPhoneOutSideCarabobo {
			loc.Phone = psi.PhoneOutSideCarabobo
		}
		if psi.ShowCellPhoneOutSideCarabobo {
			loc.CellPhone = psi.CelPhoneOutSideCarabobo
		}
		if psi.ShowPublicServiceAddressOutSideCarabobo {
			loc.Address = psi.ServiceAddressOutSideCarabobo
		}
		dto.Location.Venezuela = loc
	}

	// ── Ubicación: Exterior ───────────────────────────────────────────────
	if psi.Country != "" && (psi.ShowPublicServiceAddressOutSideVenezuela || psi.ShowPhoneOutSideVenezuela) {
		loc := &request_structs.PsiLocationExteriorDTO{
			Country: psi.Country,
		}
		if psi.ShowPhoneOutSideVenezuela {
			loc.Phone = psi.PhoneOutSideVenezuela
		}
		if psi.ShowPublicServiceAddressOutSideVenezuela {
			loc.Address = psi.ServiceAddressOutSideVenezuela
		}
		dto.Location.Exterior = loc
	}

	// ── Especialidades ────────────────────────────────────────────────────
	if psi.PrimarySpecialty != "" {
		dto.Specialties = append(dto.Specialties, psi.PrimarySpecialty)
	}
	if psi.SecondarySpecialty != "" {
		dto.Specialties = append(dto.Specialties, psi.SecondarySpecialty)
	}

	// ── Privacy Shield: Pregrado ──────────────────────────────────────────
	if psi.ColData.ShowUniversityUndergraduate {
		dto.Undergraduate.University = psi.ColData.UniversityUndergraduate
	}
	if psi.ColData.ShowGraduateDate {
		dto.Undergraduate.Date = psi.ColData.GraduateDate.Format("2006-01-02")
	}
	if psi.ColData.ShowMentionUndergraduate {
		dto.Undergraduate.Mention = psi.ColData.MentionUndergraduate
	}
	dto.Undergraduate.TitleImageOneURL = psi.ColData.TitleImageOneS3Key
	dto.Undergraduate.TitleImageTwoURL = psi.ColData.TitleImageTwoS3Key
	dto.Undergraduate.TitleImageThreeURL = psi.ColData.TitleImageThreeS3Key

	// ── Redes Sociales ────────────────────────────────────────────────────
	for _, sn := range psi.SocialNetworks {
		dto.SocialNetworks = append(dto.SocialNetworks, request_structs.SocialNetworkDTO{
			Name: sn.Name,
			URL:  sn.URL,
		})
	}

	// ── Postgrados (solo activos) ─────────────────────────────────────────
	for _, pg := range psi.PostGrades {
		if pg.Active {
			dto.PostGrades = append(dto.PostGrades, request_structs.PostGradeDTO{
				Title:       pg.Title,
				University:  pg.University,
				Year:        pg.GraduationYear,
				Description: pg.Description,
				PicOneURL:   pg.PicOneS3Key,
				PicTwoURL:   pg.PicTwoS3Key,
				PicThreeURL: pg.PicThreeS3Key,
			})
		}
	}

	log.Printf("##### #### Undergraduate data being sent: University=%v, Date=%v, Mention=%v, Images=[%v, %v, %v]",
		dto.Undergraduate.University,
		dto.Undergraduate.Date,
		dto.Undergraduate.Mention,
		dto.Undergraduate.TitleImageOneURL,
		dto.Undergraduate.TitleImageTwoURL,
		dto.Undergraduate.TitleImageThreeURL,
	)

	return dto, psi.ID, nil
}

// =========================================================================
// AUTENTICACIÓN Y SESIÓN
// =========================================================================

// Login gestiona el acceso de psicólogos implementando "Key Rotation".
// Al iniciar sesión, se genera un nuevo secreto de firma que invalida físicamente
// cualquier token previo del usuario en otros dispositivos.
func (s *PsiService) Login(ctx context.Context, identifier, password string) (string, *domain.PsiUserModel, error) {
	// 1. Buscar usuario
	psi, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		return "", nil, errors.New("credenciales inválidas")
	}

	// 2. Verificar si está activo (Soft delete o ban)
	if !psi.IsActive {
		return "", nil, errors.New("cuenta inactiva o suspendida")
	}

	// 3. Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(psi.Password), []byte(password)); err != nil {
		return "", nil, errors.New("credenciales inválidas")
	}

	// 4. ROTACIÓN DE SESIÓN (Seguridad Senior)
	// Generamos una nueva llave. Esto invalida tokens anteriores en otros dispositivos.
	newKey := uuid.New().String()
	psi.Key = newKey

	// Auditoría automática de login
	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID

	if err := s.repo.UpdateKey(ctx, psi); err != nil {
		return "", nil, errors.New("error de sistema al iniciar sesión")
	}

	// 5. Generar JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": psi.ID.String(),
		"role":    "psi", // Rol específico para el middleware híbrido
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	// Notificacion de Login
	mailData := map[string]interface{}{
		"Name":      psi.Username,
		"Email":     psi.Email,
		"LoginTime": time.Now().Format(time.RFC1123),
	}

	// Invocación dinámica y no-bloqueante
	if err := s.mailService.SendEmail(psi.Email, "Colegio de Psicólogos de Carabobo - Inicio de sesión en la plataforma.", "login_psi", mailData); err != nil {
		log.Printf("⚠️ Error al preparar el correo (pero el psicólogo se logueó): %v", err)
	}

	// Firmar con la llave personal del usuario
	signed, err := token.SignedString([]byte(newKey))
	return signed, psi, err
}

// Logout
func (s *PsiService) Logout(ctx context.Context, psi *domain.PsiUserModel) error {
	// Rotar la key invalida físicamente el token actual
	// — cualquier request posterior con el JWT viejo fallará en validateToken
	psi.Key = uuid.New().String()
	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID
	return s.repo.UpdateKey(ctx, psi)
}

// =========================================================================
// MÓDULO ACADÉMICO (CERTIFICADOS)
// =========================================================================

// AddPostGrade registra un título y gestiona la subida de hasta 3 documentos a S3.
// Implementa saneamiento de imágenes para eliminar metadatos y scripts maliciosos.
func (s *PsiService) AddPostGrade(ctx context.Context, psi *domain.PsiUserModel, req request_structs.CreatePostGradeRequest, files []*multipart.FileHeader) error {

	// Estructura base
	postGrade := &domain.PsiUserPostGrade{
		AuditModel: domain.AuditModel{
			CreateById: &psi.ID,
			CreateBy:   psi.Username,
			UpdateById: &psi.ID,
			UpdateBy:   psi.Username,
		},
		PsiUserID:      psi.ID,
		Title:          req.Title,
		University:     req.University,
		GraduationYear: req.GraduationYear,
		Description:    req.Description,
		Active:         true,
	}

	// Helper interno para procesar subidas
	uploadHelper := func(fh *multipart.FileHeader) (string, error) {
		if fh == nil {
			return "", nil
		}

		// 1. Abrir y Sanitizar (Re-encoding a JPEG/PNG limpio)
		src, err := fh.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return "", fmt.Errorf("error en imagen: %v", err)
		}

		// 2. Subir a S3
		filename := uuid.New().String() + ext
		// Guardamos en la carpeta 'certificates'
		return s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "certificates", filename, contentType)
	}

	// Procesar las 3 imágenes posibles
	// Asumimos que el slice 'files' viene en orden [pic1, pic2, pic3] desde el handler
	var err error
	if len(files) > 0 {
		postGrade.PicOneS3Key, err = uploadHelper(files[0])
	}
	if err != nil {
		return err
	}

	if len(files) > 1 {
		postGrade.PicTwoS3Key, err = uploadHelper(files[1])
	}
	if err != nil {
		return err
	}

	if len(files) > 2 {
		postGrade.PicThreeS3Key, err = uploadHelper(files[2])
	}
	if err != nil {
		return err
	}

	// Persistir en DB
	return s.repo.CreatePostGrade(ctx, postGrade)
}

// UpdatePostGrade permite editar un título y reemplazar sus imágenes.
// Implementa un reemplazo inteligente: si se sube una nueva imagen, se borra la anterior de S3 para evitar acumulación de archivos huérfanos.
// fileMap: mapa donde la clave es el campo (ej: "pic_one") y el valor es el archivo.
func (s *PsiService) UpdatePostGrade(ctx context.Context, psi *domain.PsiUserModel, pgID uuid.UUID, req request_structs.UpdatePostGradeRequest, fileMap map[string]*multipart.FileHeader) error {

	// 1. Obtener el registro actual
	pg, err := s.repo.GetPostGradeByID(ctx, pgID)
	if err != nil {
		return errors.New("título académico no encontrado")
	}

	// 2. SEGURIDAD: Verificar Propiedad (Ownership Check)
	// Impedir que el Psicólogo A edite el título del Psicólogo B
	if pg.PsiUserID != psi.ID {
		return errors.New("no tienes permiso para editar este registro")
	}

	// 3. Auditoría
	pg.UpdateBy = psi.Username
	pg.UpdateById = &psi.ID
	pg.UpdatedAt = time.Now()

	// 4. Actualización de Campos de Texto (Si vienen en el request)
	if req.Title != nil {
		pg.Title = *req.Title
	}
	if req.University != nil {
		pg.University = *req.University
	}
	if req.GraduationYear != nil {
		pg.GraduationYear = *req.GraduationYear
	}
	if req.Description != nil {
		pg.Description = *req.Description
	}

	// 5. GESTIÓN DE IMÁGENES (Reemplazo Inteligente)
	// Helper para no repetir código: Sube nueva -> Borra vieja -> Retorna nueva Key
	replaceImage := func(newFile *multipart.FileHeader, oldKey string) (string, error) {
		// A. Sanitizar y leer
		src, err := newFile.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return "", err
		}

		// B. Subir nueva
		filename := uuid.New().String() + ext
		newKey, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "certificates", filename, contentType)
		if err != nil {
			return "", err
		}

		// C. Borrar vieja (Si existía)
		if oldKey != "" {
			// No bloqueamos si falla el borrado, solo logueamos (idealmente)
			_ = s.s3Client.DeleteFile(ctx, oldKey)
		}
		return newKey, nil
	}

	// Aplicar para cada slot si viene un archivo
	if file, ok := fileMap["pic_one"]; ok {
		pg.PicOneS3Key, err = replaceImage(file, pg.PicOneS3Key)
		if err != nil {
			return err
		}
	}
	if file, ok := fileMap["pic_two"]; ok {
		pg.PicTwoS3Key, err = replaceImage(file, pg.PicTwoS3Key)
		if err != nil {
			return err
		}
	}
	if file, ok := fileMap["pic_three"]; ok {
		pg.PicThreeS3Key, err = replaceImage(file, pg.PicThreeS3Key)
		if err != nil {
			return err
		}
	}

	// 6. Persistir cambios
	return s.repo.UpdatePostGrade(ctx, pg)
}

// =========================================================================
// --- HELPERS DE CONVERSIÓN (Privados) ---
// =========================================================================
func parseInt(s string) int {
	val, _ := strconv.Atoi(strings.TrimSpace(s))
	return val
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "v" || s == "s"
}

func parseDate(s string) time.Time {
	layout := "2006-01-02" // Formato estándar del CSV que pasaste
	t, err := time.Parse(layout, strings.TrimSpace(s))
	if err != nil {
		return time.Time{} // Fecha cero si falla
	}
	return t
}

func (s *PsiService) GetPsiBioByID(ctx context.Context, id uuid.UUID) (string, error) {
	bio, err := s.repo.GetTextContentByID(ctx, id)
	if err != nil {
		return "", err
	}
	return bio, nil
}

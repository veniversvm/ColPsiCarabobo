// api/internal/service/psi_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
// Este archivo contiene las operaciones centrales relacionadas con los psicólogos colegiados,
// incluyendo la importación masiva desde CSV, la gestión de perfiles públicos y la autenticación.

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
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
	"github.com/xuri/excelize/v2"
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
	// 1. Logs (Igual que antes)
	_ = os.Mkdir("logs", 0755)
	logFileName := fmt.Sprintf("logs/import_%s.log", time.Now().Format("2006-01-02_15-04-05"))
	logFile, _ := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer logFile.Close()
	auditLogger := log.New(logFile, "", log.LstdFlags)

	// 2. Abrir Excel
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return 0, []map[string]string{{"error": "archivo inválido"}}
	}
	defer f.Close()

	rows, _ := f.Rows("BD ColPsiCarabobo 2026")

	// 3. Preparación
	successCount := 0
	var failedRecords []map[string]string
	defaultPassword := "Colpsi2025!"
	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	hashedPassword := string(hashedPasswordBytes)

	audit := domain.AuditModel{
		CreateById: &adminID, CreateBy: "Admin_XLSX_Import",
		UpdateById: &adminID, UpdateBy: "Admin_XLSX_Import",
	}

	rowIdx := 0
	for rows.Next() {
		rowIdx++
		row, _ := rows.Columns()
		if rowIdx <= 2 {
			continue
		}

		// Captura de datos...
		rawFPV := getValorSeguro(row, 3)
		rawCI := getValorSeguro(row, 6)
		firstName := getValorSeguro(row, 7)
		lastName := getValorSeguro(row, 9)
		fullName := firstName + " " + lastName

		fpvInt := parseInt(rawFPV)
		ciInt := parseInt(rawCI)

		if fpvInt == 0 || ciInt == 0 || firstName == "" {
			failedRecords = append(failedRecords, map[string]string{"fila": strconv.Itoa(rowIdx), "nombre": fullName, "error": "Datos incompletos"})
			continue
		}

		// Failsafe correo...
		email := getValorSeguro(row, 15)
		emailToProcess := email
		validEmail := true
		if email == "" || !strings.Contains(email, "@") {
			emailToProcess = fmt.Sprintf("%d.sincorreo@colpsi.com", fpvInt)
			validEmail = false
		}

		// Generación única de UUIDs v7
		psiID := uuid.Must(uuid.NewV7())
		sessionKey := uuid.Must(uuid.NewV7()).String()

		// Modelo...
		psi := &domain.PsiUserModel{
			ID: psiID, Key: sessionKey, AuditModel: audit,
			Username: generateSecureUsername(emailToProcess, strconv.Itoa(fpvInt), firstName),
			Email:    emailToProcess, Password: hashedPassword,
			FirstName: firstName, LastName: lastName,
			FPV: fpvInt, CI: ciInt, BornDate: parseDate(getValorSeguro(row, 11)),
			Genre: getValorSeguro(row, 13), IsActive: getValorSeguro(row, 45) == "Activo",
			Solvent:              getValorSeguro(row, 45) == "Activo",
			ProofOfLife:          strings.ToLower(getValorSeguro(row, 14)) != "fallecido",
			ContactPhone:         cleanDash(getValorSeguro(row, 17)),
			ContactCellPhone:     cleanDash(getValorSeguro(row, 18)),
			ContactEmail:         email, // El original
			MunicipalityCarabobo: getValorSeguro(row, 16),
		}

		colData := &domain.PsiUserColData{
			ID: uuid.Must(uuid.NewV7()), PsiUserModelID: psiID, AuditModel: audit,
			GuildInscriptionDate:    parseDate(getValorSeguro(row, 4)),
			UniversityUndergraduate: getValorSeguro(row, 25),
			GraduateDate:            parseDate(getValorSeguro(row, 26)),
			MentionUndergraduate:    getValorSeguro(row, 27),
			RegisterNumber:          parseInt(getValorSeguro(row, 30)),
			DateOfLastSolvency:      parseDate(getValorSeguro(row, 44)),
		}

		solvency := domain.PsiUSerSolvency{
			ID: uuid.Must(uuid.NewV7()), PsiUserModelID: psiID, AuditModel: audit, Date: colData.DateOfLastSolvency,
		}

		// 5. PERSISTENCIA
		if err := s.repo.CreateWithColData(ctx, psi, colData, solvency, []domain.PsiUserPostGrade{}); err != nil {
			humanError := mapDBErrorToHuman(err)
			auditLogger.Printf("❌ FILA %d | %s | %v", rowIdx, fullName, humanError)
			failedRecords = append(failedRecords, map[string]string{"fila": strconv.Itoa(rowIdx), "nombre": fullName, "error": humanError})
			continue
		}

		// 6. ENVÍO DE EMAIL NO BLOQUEANTE
		// Al estar en una goroutine, el bucle principal de la DB sigue a toda velocidad
		if psi.ProofOfLife && validEmail {
			go s.mailService.SendEmail(
				psi.Email,
				"Bienvenido(a) a la plataforma COLPSI Carabobo",
				"welcome_psi",
				map[string]interface{}{
					"Name":     psi.FirstName,
					"Email":    psi.Email,
					"Password": defaultPassword,
				},
			)
		}

		successCount++
	}

	return successCount, failedRecords
}

func mapDBErrorToHuman(err error) string {
	errStr := err.Error()

	// Errores de Duplicidad (Postgres / GORM)
	if strings.Contains(errStr, "idx_psi_users_ci") || strings.Contains(errStr, "psi_users_ci_key") {
		return "La Cédula de Identidad ya está registrada en el sistema."
	}
	if strings.Contains(errStr, "idx_psi_users_fpv") || strings.Contains(errStr, "psi_users_fpv_key") {
		return "El número de FPV ya está registrado en el sistema."
	}
	if strings.Contains(errStr, "idx_psi_users_email") || strings.Contains(errStr, "psi_users_email_key") {
		return "El correo electrónico ya está en uso por otro perfil."
	}
	if strings.Contains(errStr, "idx_psi_users_username") || strings.Contains(errStr, "psi_users_username_key") {
		return "El nombre de usuario generado ya existe."
	}

	// Errores de Longitud
	if strings.Contains(errStr, "value too long for type character varying(25)") {
		return "El nombre de usuario generado excede el límite de 25 caracteres."
	}
	if strings.Contains(errStr, "value too long") {
		return "Un campo (Nombre, Apellido o Dirección) es demasiado largo para la base de datos."
	}

	// Otros
	if strings.Contains(errStr, "invalid input syntax for type uuid") {
		return "Error interno: ID de sistema inválido."
	}

	return "Error de base de datos: " + errStr
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
func cleanDash(val string) string {
	if val == "-" || val == "0" {
		return ""
	}
	return val
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
	req request_structs.PsiUserUpdateRequestSelf, // Asegúrate de actualizar este struct con los nuevos campos
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
		psi.Key = uuid.Must(uuid.NewV7()).String()
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

	// 4b. Contacto (Se actualiza a ContactPhone/ContactCellPhone en lugar del antiguo PublicPhone)
	if req.ContactEmail != nil {
		validate_email, err := utils.ParseAndValidateEmail(*req.ContactEmail)
		if err != nil {
			return nil, err
		}
		psi.ContactEmail = validate_email
	}
	if req.ContactPhone != nil {
		psi.ContactPhone = *req.ContactPhone
	}
	if req.ContactCellPhone != nil {
		psi.ContactCellPhone = *req.ContactCellPhone
	}
	if req.ServiceAddress != nil {
		psi.ServiceAddress = *req.ServiceAddress
	}

	if v := req.ShowContactEmail(); v != nil {
		psi.ShowContactEmail = *v
	}
	if v := req.ShowPublicServiceAddress(); v != nil {
		psi.ShowPublicServiceAddress = *v
	}

	// 4c. Ubicación: Carabobo
	if req.MunicipalityCarabobo != nil {
		mun, ok := utils.NormalizeMunicipioCarabobo(*req.MunicipalityCarabobo)
		if !ok {
			return nil, fmt.Errorf("municipio de Carabobo inválido: %q", *req.MunicipalityCarabobo)
		}
		psi.MunicipalityCarabobo = mun
	}
	// (Recuerda arreglar tu modelo en DB para que ShowMunicipalityCarabobo sea bool)
	if v := req.ShowMunicipalityCarabobo(); v != nil {
		// Ajusta este casting dependiendo de si arreglas el modelo en string o bool
		// Si lo dejas como string en BD, debes guardar *v convertido a string. Si lo pasas a bool, déjalo como *v
		psi.ShowMunicipalityCarabobo = *v
	}
	if req.PhoneCarabobo != nil {
		psi.PhoneCarabobo = *req.PhoneCarabobo
	}
	if req.CelPhoneCarabobo != nil {
		psi.CelPhoneCarabobo = *req.CelPhoneCarabobo
	}
	if v := req.ShowPhoneCarabobo(); v != nil {
		psi.ShowPhoneCarabobo = *v
	}
	if v := req.ShowCelPhoneCarabobo(); v != nil {
		psi.ShowCelPhoneCarabobo = *v
	}

	// 4d. Ubicación: Fuera de Carabobo (Venezuela)
	if req.StateOutside != nil {
		estado, ok := utils.NormalizeEstadoVenezuela(*req.StateOutside)
		if !ok {
			return nil, fmt.Errorf("estado venezolano inválido o no permitido: %q", *req.StateOutside)
		}
		psi.StateOutside = estado
	}
	if v := req.ShowStateOutside(); v != nil {
		psi.ShowStateOutside = *v
	}
	if req.MunicipalityOutSideCarabobo != nil {
		psi.MunicipalityOutSideCarabobo = *req.MunicipalityOutSideCarabobo
	}
	if v := req.ShowMunicipalityOutSideCarabobo(); v != nil {
		psi.ShowMunicipalityOutSideCarabobo = *v
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
	if req.Country != nil {
		psi.Country = *req.Country
	}
	if req.PhoneOutSideVenezuela != nil {
		psi.PhoneOutSideVenezuela = *req.PhoneOutSideVenezuela
	}
	if req.CellPhoneOutSideVenezuela != nil {
		psi.CellPhoneOutSideVenezuela = *req.CellPhoneOutSideVenezuela
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

	// 4f. Perfil profesional (ahora usando WorkArea en lugar de Specialty)
	if req.PrimaryWorkArea != nil {
		psi.PrimaryWorkArea = *req.PrimaryWorkArea
	}
	if req.SecondaryWorkArea != nil {
		psi.SecondaryWorkArea = *req.SecondaryWorkArea
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
				ID:      uuid.Must(uuid.NewV7()),
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
			shortUUID := uuid.Must(uuid.NewV7()).String()[:6]
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
		if u.PrimaryWorkArea != "" {
			mini.Specialties = append(mini.Specialties, u.PrimaryWorkArea)
		}
		if u.SecondaryWorkArea != "" {
			mini.Specialties = append(mini.Specialties, u.SecondaryWorkArea)
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
			// Specialties:    make([]string, 0),
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
		// Specialties:    make([]string, 0), // Llenado dinámicamente más abajo con WorkAreas
		PrimaryWorkArea:   psi.PrimaryWorkArea,
		SecondaryWorkArea: psi.SecondaryWorkArea,
		PostGrades:        make([]request_structs.PostGradeDTO, 0),
		SocialNetworks:    make([]request_structs.SocialNetworkDTO, 0),
		Undergraduate:     request_structs.UndergraduateDTO{},
	}

	// ── Privacy Shield: Contacto principal ───────────────────────────────
	if psi.ShowContactEmail {
		dto.Email = psi.ContactEmail
	}

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	hasCaraboboData := false
	locCarabobo := &request_structs.PsiLocationCaraboboDTO{}

	if psi.ShowMunicipalityCarabobo && psi.MunicipalityCarabobo != "" {
		locCarabobo.Municipality = psi.MunicipalityCarabobo
		hasCaraboboData = true
	}
	if psi.ShowPhoneCarabobo && psi.PhoneCarabobo != "" {
		locCarabobo.Phone = psi.PhoneCarabobo
		hasCaraboboData = true
	}
	if psi.ShowCelPhoneCarabobo && psi.CelPhoneCarabobo != "" {
		locCarabobo.CellPhone = psi.CelPhoneCarabobo
		hasCaraboboData = true
	}
	if psi.ShowPublicServiceAddress && psi.ServiceAddress != "" {
		locCarabobo.Address = psi.ServiceAddress
		hasCaraboboData = true
	}
	if hasCaraboboData {
		dto.Location.Carabobo = locCarabobo
	}

	// ── Ubicación: Fuera de Carabobo (Venezuela) ──────────────────────────
	hasVenezuelaData := false
	locVenezuela := &request_structs.PsiLocationVenezuelaDTO{}

	if psi.ShowStateOutside && psi.StateOutside != "" {
		locVenezuela.State = psi.StateOutside
		hasVenezuelaData = true
	}
	if psi.ShowMunicipalityOutSideCarabobo && psi.MunicipalityOutSideCarabobo != "" {
		locVenezuela.Municipality = psi.MunicipalityOutSideCarabobo
		hasVenezuelaData = true
	}
	if psi.ShowPhoneOutSideCarabobo && psi.PhoneOutSideCarabobo != "" {
		locVenezuela.Phone = psi.PhoneOutSideCarabobo
		hasVenezuelaData = true
	}
	if psi.ShowCellPhoneOutSideCarabobo && psi.CelPhoneOutSideCarabobo != "" {
		locVenezuela.CellPhone = psi.CelPhoneOutSideCarabobo
		hasVenezuelaData = true
	}
	if psi.ShowPublicServiceAddressOutSideCarabobo && psi.ServiceAddressOutSideCarabobo != "" {
		locVenezuela.Address = psi.ServiceAddressOutSideCarabobo
		hasVenezuelaData = true
	}
	if hasVenezuelaData {
		dto.Location.Venezuela = locVenezuela
	}

	// ── Ubicación: Exterior ───────────────────────────────────────────────
	hasExteriorData := false
	locExterior := &request_structs.PsiLocationExteriorDTO{}

	if psi.Country != "" {
		locExterior.Country = psi.Country
		hasExteriorData = true
	}
	if psi.ShowPhoneOutSideVenezuela && psi.PhoneOutSideVenezuela != "" {
		locExterior.Phone = psi.PhoneOutSideVenezuela
		hasExteriorData = true
	}
	if psi.ShowCellPhoneOutSideVenezuela && psi.CellPhoneOutSideVenezuela != "" {
		locExterior.CellPhone = psi.CellPhoneOutSideVenezuela
		hasExteriorData = true
	}
	if psi.ShowPublicServiceAddressOutSideVenezuela && psi.ServiceAddressOutSideVenezuela != "" {
		locExterior.Address = psi.ServiceAddressOutSideVenezuela
		hasExteriorData = true
	}
	if hasExteriorData {
		dto.Location.Exterior = locExterior
	}

	// // ── Áreas de Trabajo (Mapeadas al DTO Specialties) ────────────────────
	// if psi.PrimaryWorkArea != "" {
	// 	dto.Specialties = append(dto.Specialties, psi.PrimaryWorkArea)
	// }
	// if psi.SecondaryWorkArea != "" {
	// 	dto.Specialties = append(dto.Specialties, psi.SecondaryWorkArea)
	// }

	// ── Privacy Shield: Pregrado ──────────────────────────────────────────
	if psi.ColData.ShowUniversityUndergraduate {
		dto.Undergraduate.University = psi.ColData.UniversityUndergraduate
	}
	if psi.ColData.ShowGraduateDate && !psi.ColData.GraduateDate.IsZero() {
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
				Type:        string(pg.Type), // 👈 Añadido el tipo (Especialización, Doctorado, etc)
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
	newKey := uuid.Must(uuid.NewV7()).String()
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
	psi.Key = uuid.Must(uuid.NewV7()).String()
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
		filename := uuid.Must(uuid.NewV7()).String() + ext
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
		filename := uuid.Must(uuid.NewV7()).String() + ext
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
func parseInt(val string) int {
	if val == "" || val == "-" {
		return 0
	}
	// ELIMINAR CUALQUIER CARÁCTER QUE NO SEA NÚMERO
	// Esto limpia comas de miles, puntos, espacios y decimales accidentales
	clean := ""
	for _, r := range val {
		if r >= '0' && r <= '9' {
			clean += string(r)
		} else if r == ',' || r == '.' {
			// Si detectamos un separador, simplemente lo ignoramos para unir los números
			// Ej: "20,493" -> "20493"
			continue
		}
	}

	i, _ := strconv.Atoi(clean)
	return i
}

func generateSecureUsername(email, fpv, name string) string {
	base := ""
	if strings.Contains(email, "@") {
		base = strings.Split(email, "@")[0]
	} else {
		base = strings.ReplaceAll(strings.ToLower(name), " ", "")
	}
	combined := base + fpv
	if len(combined) > 25 {
		maxBase := 25 - len(fpv)
		if maxBase > 0 {
			combined = base[:maxBase] + fpv
		} else {
			combined = combined[:25]
		}
	}
	return combined
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "v" || s == "s"
}

func parseDate(val string) time.Time {
	if val == "" || val == "-" || val == "0" {
		return time.Time{}
	}
	// Excel a veces devuelve números seriales (ej: 45123). Excelize suele convertirlos,
	// pero por si acaso intentamos varios formatos comunes.
	layouts := []string{
		"02/01/2006", "02-01-2006", "2006-01-02",
		"1/2/06", "1-2-06", "01-02-06",
	}
	for _, l := range layouts {
		t, err := time.Parse(l, val)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

func (s *PsiService) GetPsiBioByID(ctx context.Context, id uuid.UUID) (string, error) {
	bio, err := s.repo.GetTextContentByID(ctx, id)
	if err != nil {
		return "", err
	}
	return bio, nil
}

func (s *PsiService) GetPsiSOlvency(ctx context.Context, id uuid.UUID) ([]domain.PsiUSerSolvency, error) {
	bio, err := s.repo.GetSolvencies(ctx, id)
	if err != nil {
		return []domain.PsiUSerSolvency{}, err
	}
	return bio, nil
}

func (s *PsiService) GetSitemapPsis(ctx context.Context) (interface{}, error) {
	// Solo traemos los campos mínimos para el sitemap
	return s.repo.GetSitemapData(ctx)
}

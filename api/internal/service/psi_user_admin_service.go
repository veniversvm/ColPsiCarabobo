// api/internal/service/psi_user_admin_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
// Este archivo contiene las operaciones administrativas de alto nivel para la gestión
// de los expedientes de los psicólogos colegiados.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// GESTIÓN DE EXPEDIENTES (LECTURA DETALLADA)
// =========================================================================

// GetPsiByIDAdmin retorna el expediente completo de un psicólogo sin restricciones de privacidad.
// Está diseñado exclusivamente para el uso del personal administrativo autorizado.
// Implementa una barrera de permisos RBAC interna para asegurar que solo personal de gestión acceda.
func (s *PsiService) GetPsiByIDAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) (*domain.PsiUserModel, error) {
	// Como es una operación de solo lectura para el panel interno,
	// verificamos que sea Sudo o que al menos tenga un permiso administrativo básico.
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, errors.New("permisos insuficientes para ver expedientes detallados")
	}

	psi, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return nil, errors.New("psicólogo no encontrado")
	}

	return psi, nil
}

// =========================================================================
// REGISTRO INDIVIDUAL (ESCRITURA ATÓMICA)
// =========================================================================

// CreatePsiByAdmin orquesta la creación manual de un nuevo colegiado.
// Realiza validación de permisos, geo-normalización, hashing de credenciales,
// parseo de fechas y escritura transaccional atómica del perfil y datos colegiales.
func (s *PsiService) CreatePsiByAdmin(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreatePsiAdminRequest) error {
	// 1. Validar Permisos
	if !admin.CanCreatePsi && !admin.Sudo {
		return errors.New("no tienes permiso para registrar psicólogos")
	}

	// 2. Geo-validación y normalización
	// Los campos opcionales solo se validan si vienen en el request.
	var municipioCarabobo string
	if req.MunicipalityCarabobo != "" {
		mun, ok := utils.NormalizeMunicipioCarabobo(req.MunicipalityCarabobo)
		if !ok {
			return fmt.Errorf("municipio de Carabobo inválido: %q", req.MunicipalityCarabobo)
		}
		municipioCarabobo = mun
	}

	var estadoOutside string
	if req.StateOutside != "" {
		estado, ok := utils.NormalizeEstadoVenezuela(req.StateOutside)
		if !ok {
			return fmt.Errorf("estado venezolano inválido o no permitido: %q", req.StateOutside)
		}
		estadoOutside = estado
	}

	// 3. Hash de Password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error al procesar seguridad")
	}

	// 4. Parsear fechas
	bornDate, _ := time.Parse("2006-01-02", req.BornDate)
	gradDate, _ := time.Parse("2006-01-02", req.GraduateDate)
	solvDate, _ := time.Parse("2006-01-02", req.DateOfLastSolvency)
	regDate, _ := time.Parse("2006-01-02", req.RegisterTitleDate)

	// 5. Mapeo de Identidades (UUIDs frescos)
	psiID := uuid.New()
	colDataID := uuid.New()

	psi := &domain.PsiUserModel{
		ID:  psiID,
		Key: uuid.New().String(),
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},

		// ── Credenciales ──────────────────────────────────────────────────
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashed),

		// ── Identidad Legal ───────────────────────────────────────────────
		FirstName:      req.FirstName,
		SecondName:     req.SecondName,
		LastName:       req.LastName,
		SecondLastName: req.SecondLastName,
		CI:             req.CI,
		FPV:            req.FPV,
		BornDate:       bornDate,
		Genre:          req.Genre,
		Nationality:    req.Nationality,

		// ── Estatus Administrativo ────────────────────────────────────────
		Solvent:     req.Solvent,
		ProofOfLife: req.ProofOfLife,
		IsActive:    req.IsActive,

		// ── Contacto Público ──────────────────────────────────────────────
		ContactEmail:   req.ContactEmail,
		PublicPhone:    req.PublicPhone,
		ServiceAddress: req.ServiceAddress,

		// ── Ubicación: Carabobo ───────────────────────────────────────────
		MunicipalityCarabobo: municipioCarabobo, // normalizado
		PhoneCarabobo:        req.PhoneCarabobo,
		CelPhoneCarabobo:     req.CelPhoneCarabobo,

		// ── Ubicación: Fuera de Carabobo (Venezuela) ─────────────────────
		StateOutside:                  estadoOutside, // normalizado, sin Carabobo
		MunicipalityOutSideCarabobo:   req.MunicipalityOutSideCarabobo,
		PhoneOutSideCarabobo:          req.PhoneOutSideCarabobo,
		CelPhoneOutSideCarabobo:       req.CelPhoneOutSideCarabobo,
		ServiceAddressOutSideCarabobo: req.ServiceAddressOutSideCarabobo,

		// ── Ubicación: Fuera de Venezuela ─────────────────────────────────
		Country:                        req.Country,
		PhoneOutSideVenezuela:          req.PhoneOutSideVenezuela,
		ServiceAddressOutSideVenezuela: req.ServiceAddressOutSideVenezuela,

		// ── Perfil Profesional ────────────────────────────────────────────
		PrimarySpecialty:   req.PrimarySpecialty,
		SecondarySpecialty: req.SecondarySpecialty,
	}

	colData := &domain.PsiUserColData{
		ID:             colDataID,
		PsiUserModelID: psiID,
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},

		// ── Pregrado ──────────────────────────────────────────────────────
		UniversityUndergraduate: req.UniversityUndergraduate,
		GraduateDate:            gradDate,
		MentionUndergraduate:    req.MentionUndergraduate,

		// ── Registro Legal del Título ─────────────────────────────────────
		RegisterTitleState: req.RegisterTitleState,
		RegisterTitleDate:  regDate,
		RegisterNumber:     req.RegisterNumber,
		RegisterFolio:      req.RegisterFolio,
		RegisterTome:       req.RegisterTome,

		// ── Flags Gremiales ───────────────────────────────────────────────
		GuildDirector:       req.GuildDirector,
		SixtyFiveOrPlus:     req.SixtyFiveOrPlus,
		GuildCollaborator:   req.GuildCollaborator,
		PublicEmployee:      req.PublicEmployee,
		UniversityProfessor: req.UniversityProfessor,

		// ── Historial Gremial ─────────────────────────────────────────────
		DateOfLastSolvency: solvDate,
		DoubleGuild:        req.DoubleGuild,
		CPSM:               req.CPSM,
	}

	// 6. Persistencia transaccional — un solo punto de fallo
	if err := s.repo.CreateWithColData(ctx, psi, colData); err != nil {
		return MapDBError(err)
	}

	// 7. Notificación de bienvenida — no bloqueante, fallo silencioso intencional
	mailData := map[string]interface{}{
		"Name":     psi.Username,
		"Email":    psi.Email,
		"Password": req.Password,
	}
	if err := s.mailService.SendEmail(psi.Email, "Bienvenido", "welcome_psi", mailData); err != nil {
		log.Printf("⚠️ Error al enviar correo de bienvenida (psi creado correctamente): %v", err)
	}

	return nil
}

// =========================================================================
// ACTUALIZACIÓN MAESTRA (CONTROL TOTAL)
// =========================================================================

// UpdatePsiByAdmin permite a un administrador modificar íntegramente el expediente.
// Soporta actualizaciones parciales (PATCH) mediante el uso de punteros en el DTO,
// garantizando que solo los campos enviados en el JSON sean alterados en la base de datos.
func (s *PsiService) UpdatePsiByAdmin(
	ctx context.Context,
	admin *domain.UserAdmin,
	targetID uuid.UUID,
	req request_structs.UpdatePsiAdminRequest,
	profilePic *multipart.FileHeader,
	titleImgOne *multipart.FileHeader,
	titleImgTwo *multipart.FileHeader,
	titleImgThree *multipart.FileHeader,
) error {
	// 1. VALIDACIÓN DE PERMISOS (RBAC)
	if !admin.CanUpdatePsi && !admin.Sudo {
		return errors.New("no tienes permiso para editar registros de psicólogos")
	}

	// 2. OBTENER REGISTRO ACTUAL
	psi, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("error al recuperar el psicólogo: %w", err)
	}

	// Helper local para parsear fechas
	parseDate := func(dateStr *string) time.Time {
		if dateStr == nil || *dateStr == "" {
			return time.Time{}
		}
		t, _ := time.Parse("2006-01-02", *dateStr)
		return t
	}

	// fmt.Printf("DEBUG MunicipalityCarabobo: %v\n", *req.MunicipalityCarabobo)
	// fmt.Printf("DEBUG StateOutside: %v\n", *req.StateOutside)
	if req.StateOutside != nil {
		fmt.Printf("DEBUG StateOutside value: %q\n", *req.StateOutside)
		estado, ok := utils.NormalizeEstadoVenezuela(*req.StateOutside)
		fmt.Printf("DEBUG NormalizeEstadoVenezuela result: %q, ok: %v\n", estado, ok)
		if !ok {
			return fmt.Errorf("estado venezolano inválido o no permitido: %q", *req.StateOutside)
		}
		psi.StateOutside = estado
	}

	var uploadedS3Keys []string
	var oldS3KeysToDelete []string

	// 3. IMAGEN DE PERFIL
	if profilePic != nil {
		src, _ := profilePic.Open()
		defer src.Close()
		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return err
		}
		filename := fmt.Sprintf("%s%s", psi.ID.String(), ext)
		newKey, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "avatars", filename, contentType)
		if err != nil {
			return err
		}
		uploadedS3Keys = append(uploadedS3Keys, newKey)
		if psi.ProfilePictureS3Key != "" && psi.ProfilePictureS3Key != newKey {
			oldS3KeysToDelete = append(oldS3KeysToDelete, psi.ProfilePictureS3Key)
		}
		psi.ProfilePictureS3Key = newKey
	}

	// 4. MAPEO DE TABLA PRINCIPAL (PsiUserModel)

	// 4a. Identidad y Filiación (solo admin)
	if req.FirstName != nil {
		psi.FirstName = *req.FirstName
	}
	if req.SecondName != nil {
		psi.SecondName = *req.SecondName
	}
	if req.LastName != nil {
		psi.LastName = *req.LastName
	}
	if req.SecondLastName != nil {
		psi.SecondLastName = *req.SecondLastName
	}
	if req.Email != nil {
		validate_email, err := utils.ParseAndValidateEmail(*req.Email)
		if err != nil {
			return err
		}
		err = s.repo.ValidateUniqueCredentials(ctx, "", validate_email, psi.ID)
		if err != nil {
			return err
		}
		psi.Email = validate_email
	}
	if req.Username != nil {
		validate_username := strings.ToLower(*req.Username)
		err := s.repo.ValidateUniqueCredentials(ctx, validate_username, "", psi.ID)
		if err != nil {
			return err
		}
		psi.Username = validate_username
	}
	if req.CI != nil {
		psi.CI = *req.CI
	}
	if req.FPV != nil {
		psi.FPV = *req.FPV
	}
	if req.Genre != nil {
		psi.Genre = *req.Genre
	}
	if req.Nationality != nil {
		psi.Nationality = *req.Nationality
	}
	if req.BornDate != nil {
		psi.BornDate = parseDate(req.BornDate)
	}

	// 4b. Estatus Administrativo (solo admin)
	if req.Solvent != nil {
		psi.Solvent = *req.Solvent
	}
	if req.ProofOfLife != nil {
		psi.ProofOfLife = *req.ProofOfLife
	}
	if req.IsActive != nil {
		psi.IsActive = *req.IsActive
	}

	// 4c. Contacto y Privacidad
	if req.ContactEmail != nil {
		validate_email, err := utils.ParseAndValidateEmail(*req.ContactEmail)
		if err != nil {
			return err
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

	// 4d. Ubicación: Carabobo
	if req.MunicipalityCarabobo != nil {
		mun, ok := utils.NormalizeMunicipioCarabobo(*req.MunicipalityCarabobo)
		if !ok {
			return fmt.Errorf("municipio de Carabobo inválido: %q", *req.MunicipalityCarabobo)
		}
		psi.MunicipalityCarabobo = mun
	}
	if req.PhoneCarabobo != nil {
		psi.PhoneCarabobo = *req.PhoneCarabobo
	}
	if req.CelPhoneCarabobo != nil {
		psi.CelPhoneCarabobo = *req.CelPhoneCarabobo
	}

	// 4e. Ubicación: Fuera de Carabobo (Venezuela)
	if req.StateOutside != nil {
		estado, ok := utils.NormalizeEstadoVenezuela(*req.StateOutside)
		if !ok {
			return fmt.Errorf("estado venezolano inválido o no permitido: %q", *req.StateOutside)
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

	// 4f. Ubicación: Fuera de Venezuela
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

	// 4g. Perfil Profesional
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

	// 4h. Biografía extensa (sanitización XSS)
	var bioTextToUpdate *domain.TextModel
	if req.FullBio != nil {
		cleanHTML := s.sanitizer.Sanitize(*req.FullBio)
		if psi.BioTextID != uuid.Nil {
			psi.FullBio.Content = cleanHTML
			psi.FullBio.UpdateBy = admin.Username
			psi.FullBio.UpdateById = &admin.ID
		} else {
			psi.FullBio = domain.TextModel{
				ID:      uuid.New(),
				Content: cleanHTML,
				AuditModel: domain.AuditModel{
					CreateBy: admin.Username, CreateById: &admin.ID,
					UpdateBy: admin.Username, UpdateById: &admin.ID,
				},
			}
			psi.BioTextID = psi.FullBio.ID
		}
		bioTextToUpdate = &psi.FullBio
	}

	// 5. MAPEO DE TABLA RELACIONADA (PsiUserColData)
	hasColDataChanges := req.ShowUniversityUndergraduateRaw != "" ||
		req.ShowGraduateDateRaw != "" ||
		req.ShowMentionUndergraduateRaw != "" ||
		req.UniversityUndergraduate != nil ||
		req.GraduateDate != nil ||
		req.MentionUndergraduate != nil ||
		req.RegisterNumber != nil ||
		req.RegisterTitleState != nil ||
		req.RegisterTitleDate != nil ||
		req.RegisterFolio != nil ||
		req.RegisterTome != nil ||
		req.GuildDirector != nil ||
		req.SixtyFiveOrPlus != nil ||
		req.GuildCollaborator != nil ||
		req.PublicEmployee != nil ||
		req.UniversityProfessor != nil ||
		req.DoubleGuild != nil ||
		req.CPSM != nil ||
		req.DateOfLastSolvency != nil ||
		titleImgOne != nil || titleImgTwo != nil || titleImgThree != nil

	var colDataToUpdate *domain.PsiUserColData
	if hasColDataChanges {
		currentColData, err := s.repo.GetPsiUserColData(ctx, psi.ID)
		if err != nil {
			return err
		}

		// Visibilidad colegial
		if v := req.ShowUniversityUndergraduate(); v != nil {
			currentColData.ShowUniversityUndergraduate = *v
		}
		if v := req.ShowGraduateDate(); v != nil {
			currentColData.ShowGraduateDate = *v
		}
		if v := req.ShowMentionUndergraduate(); v != nil {
			currentColData.ShowMentionUndergraduate = *v
		}

		// Datos académicos (solo admin)
		if req.UniversityUndergraduate != nil {
			currentColData.UniversityUndergraduate = *req.UniversityUndergraduate
		}
		if req.GraduateDate != nil {
			currentColData.GraduateDate = parseDate(req.GraduateDate)
		}
		if req.MentionUndergraduate != nil {
			currentColData.MentionUndergraduate = *req.MentionUndergraduate
		}

		// Registro de título (solo admin)
		if req.RegisterNumber != nil {
			currentColData.RegisterNumber = *req.RegisterNumber
		}
		if req.RegisterTitleState != nil {
			currentColData.RegisterTitleState = *req.RegisterTitleState
		}
		if req.RegisterTitleDate != nil {
			currentColData.RegisterTitleDate = parseDate(req.RegisterTitleDate)
		}
		if req.RegisterFolio != nil {
			currentColData.RegisterFolio = *req.RegisterFolio
		}
		if req.RegisterTome != nil {
			currentColData.RegisterTome = *req.RegisterTome
		}

		// Banderas profesionales (solo admin)
		if req.GuildDirector != nil {
			currentColData.GuildDirector = *req.GuildDirector
		}
		if req.SixtyFiveOrPlus != nil {
			currentColData.SixtyFiveOrPlus = *req.SixtyFiveOrPlus
		}
		if req.GuildCollaborator != nil {
			currentColData.GuildCollaborator = *req.GuildCollaborator
		}
		if req.PublicEmployee != nil {
			currentColData.PublicEmployee = *req.PublicEmployee
		}
		if req.UniversityProfessor != nil {
			currentColData.UniversityProfessor = *req.UniversityProfessor
		}
		if req.DoubleGuild != nil {
			currentColData.DoubleGuild = *req.DoubleGuild
		}
		if req.CPSM != nil {
			currentColData.CPSM = *req.CPSM
		}
		if req.DateOfLastSolvency != nil {
			currentColData.DateOfLastSolvency = parseDate(req.DateOfLastSolvency)
		}

		// Imágenes de títulos
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
			return err
		}
		if newKey, err := processTitleImage(titleImgTwo, "2", currentColData.TitleImageTwoS3Key); err == nil {
			currentColData.TitleImageTwoS3Key = newKey
		} else {
			return err
		}
		if newKey, err := processTitleImage(titleImgThree, "3", currentColData.TitleImageThreeS3Key); err == nil {
			currentColData.TitleImageThreeS3Key = newKey
		} else {
			return err
		}

		currentColData.UpdateBy = admin.Username
		currentColData.UpdateById = &admin.ID
		colDataToUpdate = currentColData
	}

	// 6. AUDITORÍA
	psi.UpdateBy = admin.Username
	psi.UpdateById = &admin.ID

	// 7. PERSISTENCIA — rollback S3 si falla la DB
	if err := s.repo.Update(ctx, psi, colDataToUpdate, bioTextToUpdate); err != nil {
		for _, key := range uploadedS3Keys {
			_ = s.s3Client.DeleteFile(context.Background(), key)
		}
		return fmt.Errorf("error al persistir los cambios: %w", err)
	}

	// Limpiar S3 reemplazados solo tras confirmar persistencia
	for _, oldKey := range oldS3KeysToDelete {
		_ = s.s3Client.DeleteFile(context.Background(), oldKey)
	}

	return nil
}

// =========================================================================
// DESTRUCCIÓN LÓGICA (SOFT DELETE)
// =========================================================================

// DeletePsiByAdmin realiza un borrado lógico del registro.
// El sistema preserva la integridad referencial para propósitos legales históricos,
// pero el usuario deja de existir para el motor de búsqueda y la autenticación.
func (s *PsiService) DeletePsiByAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) error {
	// 1. Validación de permisos (RBAC)
	if !admin.CanDeletePsi && !admin.Sudo {
		return errors.New("no tienes permiso para eliminar psicólogos")
	}

	// 2. Ejecutar borrado
	if err := s.repo.Delete(ctx, targetID); err != nil {
		return fmt.Errorf("error al eliminar el registro: %w", err)
	}

	return nil
}

// =========================================================================
// DASHBOARD Y LISTADOS
// =========================================================================

// GetAdminDirectory proporciona una vista de "Rayos X" del listado de miembros.
// Ignora filtros de solvencia y visibilidad pública, permitiendo al staff administrativo
// gestionar la morosidad y estados de actividad.
func (s *PsiService) GetAdminDirectory(ctx context.Context, admin *domain.UserAdmin, filter request_structs.PsiDirectoryFilterDTO) (interface{}, error) {
	// Seguridad: Validar que sea administrador autorizado
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, errors.New("permisos insuficientes para listar agremiados")
	}

	// Normalizar paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 12
	}

	users, total, err := s.repo.SearchAdmin(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Mapeo al DTO Administrativo
	list := make([]request_structs.PsiAdminListDTO, 0, len(users))
	for _, u := range users {
		list = append(list, request_structs.PsiAdminListDTO{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			CI:        u.CI,
			FPV:       u.FPV,
			Email:     u.Email,
			Solvent:   u.Solvent,
			IsActive:  u.IsActive,
		})
	}

	return fiber.Map{
		"data":        list,
		"total":       total,
		"page":        filter.Page,
		"limit":       filter.Limit,
		"total_pages": (total + int64(filter.Limit) - 1) / int64(filter.Limit),
	}, nil
}

// api/internal/service/psi_user_admin_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
// Este archivo contiene las operaciones administrativas de alto nivel (High-Privilege Operations)
// para la gestión integral de los expedientes de los psicólogos colegiados.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
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

// GetPsiByIDAdmin retorna el expediente completo de un psicólogo.
//
// Arquitectura de Visión de Rayos X (Bypass del Privacy Shield):
// A diferencia del perfil público, el personal del Colegio (Staff) requiere acceso total
// al expediente (teléfonos privados, correos, historial de solvencias) para tareas
// operativas y legales. Este método omite intencionalmente los filtros de privacidad.
func (s *PsiService) GetPsiByIDAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) (*domain.PsiUserModel, error) {
	// Como es una operación de solo lectura para el panel interno,
	// verificamos que sea Sudo o que al menos tenga un permiso administrativo básico.
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, domain.ErrInsufficientPerms
	}

	psi, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return nil, domain.ErrPsiNotFound
	}

	// Recuperación del historial financiero asociado al expediente
	solvencies, err := s.repo.GetSolvencies(ctx, targetID)

	psi.Solvencies = solvencies

	// Convertir S3 keys a URLs públicas antes de retornar al handler
	s.ResolvePsiModelURLs(psi)

	return psi, nil
}

// =========================================================================
// REGISTRO INDIVIDUAL (ESCRITURA ATÓMICA)
// =========================================================================

// CreatePsiByAdmin orquesta la creación manual de un nuevo colegiado por parte del staff.
//
// Integridad de Datos (ACID):
// Realiza validación de permisos, geo-normalización, hashing de credenciales, parseo de fechas
// y escritura transaccional. Si alguna validación falla, nada se inserta en la base de datos.
func (s *PsiService) CreatePsiByAdmin(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreatePsiAdminRequest) error {
	// 1. Validar Permisos (Gatekeeping)
	if !admin.CanCreatePsi && !admin.Sudo {
		return errors.New("no tienes permiso para registrar psicólogos")
	}

	// 2. Validación de fuerza de contraseña
	if !utils.IsStrongPassword(req.Password) {
		return errors.New("la contraseña no cumple con los estándares de seguridad")
	}

	// 3. Geo-validación y normalización
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

	// 4. Hash de Password (Criptografía)
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error al procesar seguridad")
	}

	// 5. Parsear fechas
	bornDate, _ := time.Parse("2006-01-02", req.BornDate)
	gradDate, _ := time.Parse("2006-01-02", req.GraduateDate)
	solvDate, _ := time.Parse("2006-01-02", req.DateOfLastSolvency)
	regDate, _ := time.Parse("2006-01-02", req.RegisterTitleDate)

	// 6. Mapeo de Identidades (UUIDs frescos v7 para indexación optimizada en B-Trees)
	psiID := uuid.Must(uuid.NewV7())
	colDataID := uuid.Must(uuid.NewV7())

	// 7. Crear un audit model (Inmutabilidad Forense)
	audit_moodel := domain.AuditModel{
		CreateBy:   admin.Username,
		CreateById: &admin.ID,
		UpdateBy:   admin.Username,
		UpdateById: &admin.ID,
	}

	// Delegación a Constructores (Patrón Factory)
	psi := createPsiUSerModel(req, psiID, audit_moodel, hashed, municipioCarabobo, estadoOutside, bornDate)

	//── Solvencias ────────────────────────────────────────────
	solvencies := buildSolvencyHistory(solvDate, 0, psi.ID, audit_moodel)

	//── Datos colegiales ────────────────────────────────────────────
	colData := createColdata(req, psiID, colDataID, audit_moodel, gradDate, regDate, solvDate)

	// 8. Persistencia transaccional — un solo punto de fallo.
	// MapDBError actúa como escudo para no filtrar metadatos de Postgres al cliente.
	if err := s.repo.CreateWithColData(ctx, psi, colData, solvencies, []domain.PsiUserPostGrade{}); err != nil {
		return MapDBError(err)
	}

	// 9. Notificación de bienvenida — no bloqueante, fallo silencioso intencional
	// Si el SMTP falla, la creación del registro no hace Rollback.
	mailData := map[string]interface{}{
		"Name":     psi.Username,
		"Email":    psi.Email,
		"Password": req.Password,
	}
	if s.mailService != nil {
		if err := s.mailService.SendEmail(psi.Email, "Bienvenido", "welcome_psi", mailData); err != nil {
			log.Warn().Err(err).Str("component", "psi_user_admin_service").Msg("Error al enviar correo de bienvenida (psi creado correctamente)")
		}
	}

	// 10. Aprovisionar cuenta en Audiobookshelf si el agremiado nace solvente.
	// No bloqueante: un fallo de ABS no revierte la creación del expediente.
	if psi.Solvent {
		s.EnsureAudiobookshelf(ctx, psi)
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

	// Estructuras para Transacción Distribuida (Saga/Rollback manual de S3)
	var uploadedS3Keys []string
	var oldS3KeysToDelete []string

	// 3. IMAGEN DE PERFIL (Sanitización y Carga S3)
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
		if psi.ProfilePictureS3Key != "" && psi.ProfilePictureS3Key == newKey {
			// La key del avatar es estable (avatars/{psiID}.webp): la subida la sobrescribe
			// in-place. No debe tratarse como un objeto nuevo, porque si la persistencia
			// falla el rollback la borraría dejando la URL que la DB ya referencia en 404.
			uploadedS3Keys = uploadedS3Keys[:len(uploadedS3Keys)-1]
		}
		if psi.ProfilePictureS3Key != "" && psi.ProfilePictureS3Key != newKey {
			oldS3KeysToDelete = append(oldS3KeysToDelete, psi.ProfilePictureS3Key)
		}
		psi.ProfilePictureS3Key = newKey
	}

	// 4. MAPEO DE TABLA PRINCIPAL (PsiUserModel)

	// 4a. Credenciales de acceso
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

	// 4b. Identidad legal
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
	if req.FPV != nil {
		psi.FPV = *req.FPV
	}
	if req.CI != nil {
		psi.CI = *req.CI
	}
	if req.BornDate != nil {
		psi.BornDate = parseDate(req.BornDate)
	}
	if req.Genre != nil {
		psi.Genre = *req.Genre
	}
	if req.Nationality != nil {
		psi.Nationality = *req.Nationality
	}

	// 4c. Estado gremial
	if req.Solvent != nil {
		psi.Solvent = *req.Solvent
	}
	if req.ProofOfLife != nil {
		psi.ProofOfLife = *req.ProofOfLife
	}
	if req.IsActive != nil {
		psi.IsActive = *req.IsActive
	}

	// 4d. Contacto interno del gremio
	if req.ContactPhone != nil {
		psi.ContactPhone = *req.ContactPhone
	}
	if req.ContactCellPhone != nil {
		psi.ContactCellPhone = *req.ContactCellPhone
	}

	// 4e. Contacto público y privacidad
	if req.ContactEmail != nil {
		validate_email, err := utils.ParseAndValidateEmail(*req.ContactEmail)
		if err != nil {
			return err
		}
		psi.ContactEmail = validate_email
	}
	if v := req.ShowContactEmail(); v != nil {
		psi.ShowContactEmail = *v
	}
	if req.ServiceAddress != nil {
		psi.ServiceAddress = *req.ServiceAddress
	}
	if v := req.ShowPublicServiceAddress(); v != nil {
		psi.ShowPublicServiceAddress = *v
	}

	// 4f. Ubicación: Carabobo
	if req.MunicipalityCarabobo != nil {
		val := strings.TrimSpace(*req.MunicipalityCarabobo)
		if val == "" {
			psi.MunicipalityCarabobo = "" // Si el admin lo borra, se limpia en la DB
		} else {
			mun, ok := utils.NormalizeMunicipioCarabobo(val)
			if !ok {
				return fmt.Errorf("municipio de Carabobo inválido: %q", val)
			}
			psi.MunicipalityCarabobo = mun
		}
	}

	if v := req.ShowMunicipalityCarabobo(); v != nil {
		psi.ShowMunicipalityCarabobo = *v
	}
	if req.PhoneCarabobo != nil {
		psi.PhoneCarabobo = *req.PhoneCarabobo
	}
	if v := req.ShowPhoneCarabobo(); v != nil {
		psi.ShowPhoneCarabobo = *v
	}
	if req.CellPhoneCarabobo != nil {
		psi.CelPhoneCarabobo = *req.CellPhoneCarabobo
	}
	if v := req.ShowCelPhoneCarabobo(); v != nil {
		psi.ShowCelPhoneCarabobo = *v
	}

	// 4g. Ubicación: Fuera de Carabobo (Venezuela)
	if req.StateOutside != nil {
		val := strings.TrimSpace(*req.StateOutside)
		if val == "" {
			psi.StateOutside = "" // Si el admin lo borra, se limpia
		} else {
			estado, ok := utils.NormalizeEstadoVenezuela(val)
			if !ok {
				return fmt.Errorf("estado venezolano inválido o no permitido: %q", val)
			}
			psi.StateOutside = estado
		}
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
	if v := req.ShowPhoneOutSideCarabobo(); v != nil {
		psi.ShowPhoneOutSideCarabobo = *v
	}
	if req.CelPhoneOutSideCarabobo != nil {
		psi.CelPhoneOutSideCarabobo = *req.CelPhoneOutSideCarabobo
	}
	if v := req.ShowCellPhoneOutSideCarabobo(); v != nil {
		psi.ShowCellPhoneOutSideCarabobo = *v
	}
	if req.ServiceAddressOutSideCarabobo != nil {
		psi.ServiceAddressOutSideCarabobo = *req.ServiceAddressOutSideCarabobo
	}
	if v := req.ShowPublicServiceAddressOutSideCarabobo(); v != nil {
		psi.ShowPublicServiceAddressOutSideCarabobo = *v
	}

	// 4h. Ubicación: Fuera de Venezuela
	if req.Country != nil {
		psi.Country = *req.Country
	}
	if req.PhoneOutSideVenezuela != nil {
		psi.PhoneOutSideVenezuela = *req.PhoneOutSideVenezuela
	}
	if v := req.ShowPhoneOutSideVenezuela(); v != nil {
		psi.ShowPhoneOutSideVenezuela = *v
	}
	if req.CellPhoneOutSideVenezuela != nil {
		psi.CellPhoneOutSideVenezuela = *req.CellPhoneOutSideVenezuela
	}
	if v := req.ShowCellPhoneOutSideVenezuela(); v != nil {
		psi.ShowCellPhoneOutSideVenezuela = *v
	}
	if req.ServiceAddressOutSideVenezuela != nil {
		psi.ServiceAddressOutSideVenezuela = *req.ServiceAddressOutSideVenezuela
	}
	if v := req.ShowPublicServiceAddressOutSideVenezuela(); v != nil {
		psi.ShowPublicServiceAddressOutSideVenezuela = *v
	}

	// 4i. Perfil Profesional
	if req.PrimaryWorkArea != nil {
		psi.PrimaryWorkArea = *req.PrimaryWorkArea
	}
	if req.SecondaryWorkArea != nil {
		psi.SecondaryWorkArea = *req.SecondaryWorkArea
	}
	if req.PrimarySpecialtyID != nil {
		psi.PrimarySpecialtyID = req.PrimarySpecialtyID
	}
	if req.SecondarySpecialtyID != nil {
		psi.SecondarySpecialtyID = req.SecondarySpecialtyID
	}
	if req.MiniBio != nil {
		runes := []rune(*req.MiniBio)
		if len(runes) > 250 {
			psi.MiniBio = string(runes[:250])
		} else {
			psi.MiniBio = *req.MiniBio
		}
	}

	// 4j. Biografía extensa (sanitización XSS)
	var bioTextToUpdate *domain.TextModel
	if req.FullBio != nil {
		cleanHTML := s.sanitizer.Sanitize(*req.FullBio)
		if psi.BioTextID != uuid.Nil {
			psi.FullBio.Content = cleanHTML
			psi.FullBio.UpdateBy = admin.Username
			psi.FullBio.UpdateById = &admin.ID
		} else {
			psi.FullBio = domain.TextModel{
				ID:      uuid.Must(uuid.NewV7()),
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

	// 4k. Procesamiento de Solvencias (Decodificando el JSON string)
	// Solución de Arquitectura: Peticiones `multipart/form-data` no pueden anidar
	// arrays de objetos de forma limpia en HTTP. Por ello, el cliente envía un String
	// que contiene un JSON en crudo ("SolvenciesRaw"), el cual decodificamos aquí.
	var solvenciesToCreate []domain.PsiUserSolvency
	currentYear := time.Now().Year()

	if req.SolvenciesRaw != "" {
		// 1. Decodificar el string JSON a un slice de structs
		var incomingSolvencies []request_structs.SolvenciesUpdate
		if err := json.Unmarshal([]byte(req.SolvenciesRaw), &incomingSolvencies); err != nil {
			log.Error().Err(err).Str("component", "psi_admin").Msg("Error al decodificar solvencias JSON")
			// Podrías retornar error aquí o simplemente loguear y continuar
		}

		if len(incomingSolvencies) > 0 {
			// 2. Obtener solvencias actuales para evitar duplicados
			currentSolvencies, err := s.repo.GetSolvencies(ctx, psi.ID)
			if err != nil {
				return fmt.Errorf("error al obtener historial de solvencias: %w", err)
			}

			existingDates := make(map[string]bool)
			for _, s := range currentSolvencies {
				existingDates[s.Date.Format("2006-01-02")] = true
			}

			for _, incoming := range incomingSolvencies {
				if incoming.Date == "" {
					continue
				}

				// Intentar parsear la fecha (Soporta ISO y Estándar Simple)
				t, err := time.Parse(time.RFC3339, incoming.Date)
				if err != nil {
					t, err = time.Parse("2006-01-02", incoming.Date)
				}
				if err != nil {
					continue
				}

				dateKey := t.Format("2006-01-02")

				// 3. Solo añadir si NO existe ya en la DB (Idempotencia Lógica)
				if !existingDates[dateKey] {
					newSolvency := domain.PsiUserSolvency{
						ID:             uuid.Must(uuid.NewV7()),
						PsiUserModelID: psi.ID,
						Date:           t,
						AuditModel: domain.AuditModel{
							CreateBy:   admin.Username,
							CreateById: &admin.ID,
							UpdateBy:   admin.Username,
							UpdateById: &admin.ID,
						},
					}
					solvenciesToCreate = append(solvenciesToCreate, newSolvency)
					existingDates[dateKey] = true

					// Automatización de Estado: Si paga el año actual, se activa la bandera de solvente
					if t.Year() == currentYear {
						psi.Solvent = true
					}

					// Actualizar fecha de última solvencia en ColData si es más reciente
					if t.After(psi.ColData.DateOfLastSolvency) {
						fechaStr := t.Format("2006-01-02")
						req.DateOfLastSolvency = &fechaStr
					}
				}
			}
		}
	}

	// 5. MAPEO DE TABLA RELACIONADA (PsiUserColData)
	// Lazy Load: Solo preparamos la actualización si el payload tocó campos colegiales
	hasColDataChanges := req.ShowUniversityUndergraduateRaw != "" ||
		req.ShowGraduateDateRaw != "" ||
		req.ShowMentionUndergraduateRaw != "" ||
		req.GuildInscriptionDate != nil ||
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
		req.Discapacity != nil ||
		req.UniversityProfessor != nil ||
		req.DoubleGuild != nil ||
		req.DoubleGuildLocation != nil ||
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

		// Fecha de inscripción gremial
		if req.GuildInscriptionDate != nil {
			currentColData.GuildInscriptionDate = parseDate(req.GuildInscriptionDate)
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
		if req.Discapacity != nil {
			currentColData.Discapacity = *req.Discapacity
		}
		if req.UniversityProfessor != nil {
			currentColData.UniversityProfessor = *req.UniversityProfessor
		}
		if req.DoubleGuild != nil {
			currentColData.DoubleGuild = *req.DoubleGuild
		}
		if req.DoubleGuildLocation != nil {
			currentColData.DoubleGuildLocation = *req.DoubleGuildLocation
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

	// 6. AUDITORÍA (Blindaje Ciego)
	psi.UpdateBy = admin.Username
	psi.UpdateById = &admin.ID

	// 7. PERSISTENCIA — Rollback Distribuido si falla la DB
	err = s.repo.Update(ctx, psi, colDataToUpdate, bioTextToUpdate, solvenciesToCreate)
	if err != nil {
		for _, key := range uploadedS3Keys {
			_ = s.s3Client.DeleteFile(context.Background(), key)
		}
		return fmt.Errorf("error al persistir los cambios: %w", err)
	}

	// 👇 NUEVA SINCRONIZACIÓN ASÍNCRONA TRAS ÉXITO EN DB (Eventual Consistency) 👇
	var absUsername *string
	var absEmail *string

	// Validamos si el admin solicitó cambiar las credenciales
	if req.Username != nil {
		absUsername = req.Username
	}
	if req.Email != nil {
		absEmail = req.Email
	}

	// Notificación de Microservicio: Si hubo cambios en credenciales base,
	// se dispara una actualización hacia la biblioteca virtual.
	if absUsername != nil || absEmail != nil {
		// Pasamos nil en la contraseña porque el admin no la está modificando en este panel
		if absErr := s.actualizarEnAudiobookshelf(ctx, psi.AudioBookShellId, absUsername, nil, absEmail); absErr != nil {
			// Logueamos el error interno pero no bloqueamos el retorno exitoso de la petición
			// (Degradación Elegante)
			log.Warn().Err(absErr).Str("component", "psi_user_admin_service").Msg("Error al sincronizar actualización del administrador con Audiobookshelf")
		}
	}

	// Limpiar S3 reemplazados (Garbage Collection) solo tras confirmar persistencia DB
	for _, oldKey := range oldS3KeysToDelete {
		_ = s.s3Client.DeleteFile(context.Background(), oldKey)
	}

	return nil
}

// =========================================================================
// ELIMINACIÓN DE IMAGEN DE PERFIL (ADMIN)
// =========================================================================

// DeleteProfilePictureByAdmin elimina la foto de perfil de un psicólogo:
// borra el objeto del bucket S3 y limpia la referencia en la base de datos.
// Es idempotente por diseño: si el psicólogo no tiene foto, no falla.
func (s *PsiService) DeleteProfilePictureByAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) error {
	// 1. Validación de permisos (RBAC)
	if !admin.CanUpdatePsi && !admin.Sudo {
		return errors.New("no tienes permiso para editar registros de psicólogos")
	}

	// 2. Obtener registro actual
	psi, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("error al recuperar el psicólogo: %w", err)
	}

	// 3. Sin foto: no-op
	if psi.ProfilePictureS3Key == "" {
		return nil
	}

	// 4. Eliminar el objeto del bucket (best-effort: el fallo se loguea pero no
	//    bloquea la limpieza de la DB; el avatar tiene key estable y se sobrescribe).
	if err := s.s3Client.DeleteFile(ctx, psi.ProfilePictureS3Key); err != nil {
		log.Warn().Err(err).Str("component", "psi_user_admin_service").
			Str("key", psi.ProfilePictureS3Key).
			Msg("No se pudo eliminar el objeto de avatar en S3")
	}

	// 5. Limpiar la referencia en la DB
	psi.ProfilePictureS3Key = ""
	psi.UpdateBy = admin.Username
	psi.UpdateById = &admin.ID

	return s.repo.Update(ctx, psi, nil, nil, nil)
}

// =========================================================================
// DESTRUCCIÓN LÓGICA (SOFT DELETE)
// =========================================================================

// DeletePsiByAdmin realiza un borrado lógico del registro.
// El sistema preserva la integridad referencial para propósitos legales e históricos,
// pero el usuario deja de existir para el motor de búsqueda y el sistema de autenticación.
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
// REINICIO DE CLAVE (ADMIN)
// =========================================================================

// ResetPsiPasswordByAdmin reinicia la clave de acceso de un psicólogo: genera
// una contraseña temporal (12 caracteres aleatorios), la hashea con bcrypt,
// rota la Key de sesión para invalidar los JWT vigentes, fuerza el cambio de
// contraseña en el próximo login y notifica al usuario por correo.
//
// La contraseña temporal jamás se devuelve en la respuesta HTTP: viaja
// únicamente en el correo institucional (canal privado del psicólogo).
func (s *PsiService) ResetPsiPasswordByAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) error {
	// 1. Validación de permisos (RBAC)
	if !admin.CanUpdatePsi && !admin.Sudo {
		return errors.New("no tienes permiso para editar registros de psicólogos")
	}

	// 2. Obtener registro actual
	psi, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return domain.ErrPsiNotFound
	}

	// 3. Generar contraseña temporal (12 caracteres aleatorios)
	tempPassword := utils.GenerateSecureRandomString(12)

	// 4. Hash bcrypt y rotación de la Key de sesión
	hashed, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error al procesar seguridad")
	}

	psi.Password = string(hashed)
	psi.Key = uuid.Must(uuid.NewV7()).String()
	psi.MustChangePassword = true
	psi.UpdateBy = admin.Username
	psi.UpdateById = &admin.ID

	// 5. Persistencia acotada a credenciales
	if err := s.repo.ResetPassword(ctx, psi); err != nil {
		return fmt.Errorf("error al reiniciar la contraseña: %w", err)
	}

	// 6. Notificación al psicólogo — no bloqueante, fallo silencioso intencional
	// Si el SMTP falla, el reinicio de la clave ya quedó persistido.
	if s.mailService != nil {
		mailData := map[string]interface{}{
			"Name":     psi.FirstName,
			"Email":    psi.Email,
			"Password": tempPassword,
		}
		if err := s.mailService.SendEmail(psi.Email, "Contraseña reiniciada", "reset_password_psi", mailData); err != nil {
			log.Warn().Err(err).Str("component", "psi_user_admin_service").Msg("Error al enviar correo de clave reiniciada")
		}
	}

	return nil
}

// =========================================================================
// DASHBOARD Y LISTADOS
// =========================================================================

// GetAdminDirectory proporciona una vista de "Rayos X" del listado de miembros.
//
// Proyección de Datos (Data Projection DTO):
// Ignora los filtros públicos (solvencia, privacidad) entregando la información absoluta
// y re-empaqueta la salida en un `PsiAdminListDTO` ligero para evitar saturar
// la red interna con biografías y metadatos no requeridos para una tabla (DataGrid).
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
			ID:            u.ID,
			FirstName:     u.FirstName,
			LastName:      u.LastName,
			CI:            u.CI,
			FPV:           u.FPV,
			Email:         u.Email,
			ControlNumber: u.ControlNumber,
			Solvent:       u.Solvent,
			IsActive:      u.IsActive,
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

// =========================================================================
// =========================================================================
// AUXILIARY FUNCTIONS (PATRÓN FACTORY)
// =========================================================================
// =========================================================================

// Patrón Factory (Creational):
// Las siguientes funciones abstraen la complejidad de inicialización de los grafos
// de objetos del dominio, previniendo que la función principal de registro
// (CreatePsiByAdmin) se convierta en un monolito inmanejable.

// buildSolvencyHistory genera el historial anual de solvencias desde el año de
// colegiatura (mínimo 2024) hasta el año de la última solvencia conocida.
// Cada registro se fecha el 31 de diciembre de su año (convención del Colegio).
//
// Consistencia:
//   - Sin fecha de última solvencia → slice vacío (sin historial).
//   - Año de solvencia previo a 2024 o posterior al año actual → slice vacío
//     (datos inconsistentes, no se siembra historial).
//   - `guildInscriptionYear` es la fecha de colegiatura; si es 0 o menor a 2024
//     se toma 2024 como límite inferior.
func buildSolvencyHistory(lastSolvencyDate time.Time, guildInscriptionYear int, userID uuid.UUID, audit domain.AuditModel) []domain.PsiUserSolvency {
	if lastSolvencyDate.IsZero() {
		return nil
	}
	currentYear := time.Now().Year()
	lastYear := lastSolvencyDate.Year()
	if lastYear > currentYear || lastYear < 2024 {
		return nil
	}
	startYear := guildInscriptionYear
	if startYear < 2024 {
		startYear = 2024
	}
	if startYear > lastYear {
		startYear = lastYear
	}
	solvencies := make([]domain.PsiUserSolvency, 0, lastYear-startYear+1)
	for y := startYear; y <= lastYear; y++ {
		solvencies = append(solvencies, domain.PsiUserSolvency{
			ID:             uuid.Must(uuid.NewV7()),
			PsiUserModelID: userID,
			AuditModel:     audit,
			Date:           time.Date(y, time.December, 31, 0, 0, 0, 0, time.UTC),
		})
	}
	return solvencies
}

func createPsiUSerModel(req request_structs.CreatePsiAdminRequest, psiID uuid.UUID, audit_moodel domain.AuditModel, hashed []byte, municipioCarabobo, estadoOutside string, bornDate time.Time) *domain.PsiUserModel {

	return &domain.PsiUserModel{
		ID:         psiID,
		AuditModel: audit_moodel,
		Credentials: domain.Credentials{
			Key:                uuid.Must(uuid.NewV7()).String(),
			Username:           req.Username,
			Email:              req.Email,
			Password:           string(hashed),
			IsActive:           req.IsActive,
			MustChangePassword: true,
		},

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

		// ── Contacto Gremial y Público ────────────────────────────────────
		ContactEmail:     req.ContactEmail,
		ContactPhone:     req.ContactPhone,     // Reemplaza a PublicPhone
		ContactCellPhone: req.ContactCellPhone, // Nuevo campo
		ServiceAddress:   req.ServiceAddress,

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
		CellPhoneOutSideVenezuela:      req.CellPhoneOutSideVenezuela, // Nuevo campo
		ServiceAddressOutSideVenezuela: req.ServiceAddressOutSideVenezuela,

		// ── Perfil Profesional ────────────────────────────────────────────
		PrimaryWorkArea:      req.PrimaryWorkArea,
		SecondaryWorkArea:    req.SecondaryWorkArea,
		PrimarySpecialtyID:   req.PrimarySpecialtyID,
		SecondarySpecialtyID: req.SecondarySpecialtyID,
	}
}

func createColdata(req request_structs.CreatePsiAdminRequest, psiID, colDataID uuid.UUID, audit_moodel domain.AuditModel, gradDate, regDate, solvDate time.Time) *domain.PsiUserColData {
	return &domain.PsiUserColData{
		ID:             colDataID,
		PsiUserModelID: psiID,
		AuditModel:     audit_moodel,
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
}

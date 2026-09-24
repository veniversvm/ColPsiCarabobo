package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// UpdateProfileSelf updates a psychologist's own profile including credentials, personal data, images, and bio.
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

	// Snapshot previo para el diff por campo de la bitácora (auto-gestión).
	beforeSnapshot := psiSelfSnapshot(psi)
	var beforeColData map[string]any

	if err := bcrypt.CompareHashAndPassword([]byte(psi.Password), []byte(req.Password)); err != nil {
		return nil, domain.ErrPasswordIncorrect
	}

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

	// Pre-allocate with worst-case capacity: 1 profile pic + 3 title images
	uploadedS3Keys := make([]string, 0, 4)
	oldS3KeysToDelete := make([]string, 0, 4)

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

	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID

	if req.Username != nil {
		validate_username := strings.ToLower(*req.Username)
		err := s.repo.ValidateUniqueCredentials(ctx, validate_username, "", psi.ID)
		if err != nil {
			return nil, err
		}
		psi.Username = validate_username
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

	// ── Modalidad de servicio (auto-gestión) ──────────────────────────────
	if v := req.ServiceModalityPresencial(); v != nil {
		psi.ServiceModalityPresencial = *v
	}
	if v := req.ServiceModalityDistance(); v != nil {
		psi.ServiceModalityDistance = *v
	}
	if v := req.ServiceModalityTelephone(); v != nil {
		psi.ServiceModalityTelephone = *v
	}
	if v := req.ShowServiceModality(); v != nil {
		psi.ShowServiceModality = *v
	}

	if req.MunicipalityCarabobo != nil {
		mun, ok := utils.NormalizeMunicipioCarabobo(*req.MunicipalityCarabobo)
		if !ok {
			return nil, fmt.Errorf("municipio de Carabobo inválido: %q", *req.MunicipalityCarabobo)
		}
		psi.MunicipalityCarabobo = mun
	}
	if v := req.ShowMunicipalityCarabobo(); v != nil {
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

	var colDataToUpdate *domain.PsiUserColData
	hasColDataChanges := req.ShowUniversityUndergraduateRaw != "" ||
		req.ShowGraduateDateRaw != "" ||
		req.ShowMentionUndergraduateRaw != "" ||
		req.BirthdayNotificationRaw != "" ||
		titleImgOne != nil || titleImgTwo != nil || titleImgThree != nil

	if hasColDataChanges {
		currentColData, err := s.repo.GetPsiUserColData(ctx, psi.ID)
		if err != nil {
			return nil, err
		}
		// Snapshot previo de los switches de privacidad (diff de la bitácora).
		beforeColData = colDataPrivacySnapshot(currentColData)

		if v := req.ShowUniversityUndergraduate(); v != nil {
			currentColData.ShowUniversityUndergraduate = *v
		}
		if v := req.ShowGraduateDate(); v != nil {
			currentColData.ShowGraduateDate = *v
		}
		if v := req.ShowMentionUndergraduate(); v != nil {
			currentColData.ShowMentionUndergraduate = *v
		}
		if v := req.BirthdayNotification(); v != nil {
			currentColData.BirthdayNotification = *v
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

	if err := s.repo.UpdatePublicProfile(ctx, psi, colDataToUpdate, bioTextToUpdate); err != nil {
		for _, key := range uploadedS3Keys {
			_ = s.s3Client.DeleteFile(context.Background(), key)
		}
		return nil, err
	}

	// ── Bitácora de cambios: auto-gestión del perfil con diff por campo ──
	evt := auditPsiSelfEvent(psi, domain.AuditActionUpdate)
	evt.Changes = BuildDiff(beforeSnapshot, psiSelfSnapshot(psi))
	if beforeColData != nil && colDataToUpdate != nil {
		for k, v := range BuildDiff(beforeColData, colDataPrivacySnapshot(colDataToUpdate)) {
			evt.Changes[k] = v
		}
	}
	meta := map[string]any{}
	if req.NewPassword1 != nil && *req.NewPassword1 != "" {
		meta["password_changed"] = true
	}
	if profilePic != nil {
		meta["profile_picture_updated"] = true
	}
	if titleImgOne != nil || titleImgTwo != nil || titleImgThree != nil {
		meta["title_images_updated"] = true
	}
	if req.FullBio != nil {
		meta["full_bio_updated"] = true
	}
	if len(meta) > 0 {
		evt.Metadata = meta
	}
	RecordAudit(ctx, evt)

	// La cuenta ABS se identifica por el correo del agremiado (no por el
	// username). Solo un cambio de email se propaga a la biblioteca: renombra
	// la cuenta de ABS para conservar progreso/librerías. La clave de ABS
	// deriva del secreto global y un cambio de username no afecta ABS.
	var absUsername *string
	var absEmail *string

	if req.Email != nil {
		newEmail := strings.ToLower(strings.TrimSpace(*req.Email))
		absUsername = &newEmail
		absEmail = &newEmail
	}

	if absUsername != nil || absEmail != nil {
		if absErr := s.actualizarEnAudiobookshelf(ctx, psi.AudioBookShellId, absUsername, nil, absEmail); absErr != nil {
			log.Warn().Err(absErr).Str("component", "psi_service_self_management").Msg("Error al sincronizar el email con Audiobookshelf")
		}
	}

	for _, oldKey := range oldS3KeysToDelete {
		_ = s.s3Client.DeleteFile(context.Background(), oldKey)
	}

	return psi, nil
}

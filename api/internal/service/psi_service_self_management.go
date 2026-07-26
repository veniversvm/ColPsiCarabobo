package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

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

	if err := s.repo.UpdatePublicProfile(ctx, psi, colDataToUpdate, bioTextToUpdate); err != nil {
		for _, key := range uploadedS3Keys {
			_ = s.s3Client.DeleteFile(context.Background(), key)
		}
		return nil, err
	}

	var absUsername *string
	var absEmail *string
	var absPassword *string

	if req.Username != nil {
		absUsername = req.Username
	}
	if req.Email != nil {
		absEmail = req.Email
	}
	if req.NewPassword1 != nil && *req.NewPassword1 != "" {
		absPassword = req.NewPassword1
	}

	if absUsername != nil || absEmail != nil || absPassword != nil {
		if absErr := s.actualizarEnAudiobookshelf(ctx, psi.AudioBookShellId, absUsername, absPassword, absEmail); absErr != nil {
			log.Printf("WARN: Error al sincronizar actualización con Audiobookshelf: %v", absErr)
		}
	}

	for _, oldKey := range oldS3KeysToDelete {
		_ = s.s3Client.DeleteFile(context.Background(), oldKey)
	}

	return psi, nil
}

package service

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

func (s *PsiService) GetPublicDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) (interface{}, error) {
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
			ProfilePicture: s.publicURL(u.ProfilePictureS3Key),
			MiniBio:        u.MiniBio,
		}

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

func (s *PsiService) GetPublicProfile(ctx context.Context, id int) (*request_structs.PsiFullProfileDTO, uuid.UUID, error) {
	psi, err := s.repo.GetByFPV(ctx, id)
	if err != nil {
		return nil, uuid.Nil, domain.ErrPsiNotFound
	}

	if !psi.IsActive {
		return nil, uuid.Nil, errors.New("perfil no disponible")
	}

	if !psi.Solvent {
		return &request_structs.PsiFullProfileDTO{
			FirstName:      psi.FirstName,
			SecondName:     psi.SecondName,
			LastName:       psi.LastName,
			SecondLastName: psi.SecondLastName,
			FPV:            psi.FPV,
			CI:             psi.CI,
			Gender:         psi.Genre,
			ProfilePicture: s.publicURL(psi.ProfilePictureS3Key),
			Solvent:        false,
			Undergraduate: request_structs.UndergraduateDTO{
				University: psi.ColData.UniversityUndergraduate,
			},
			PostGrades:     make([]request_structs.PostGradeDTO, 0),
			SocialNetworks: make([]request_structs.SocialNetworkDTO, 0),
		}, uuid.Nil, nil
	}

	fullBio, err := s.repo.GetTextContentByID(ctx, psi.BioTextID)
	if err != nil {
		log.Warn().Err(err).Int("psi_id", id).Str("component", "psi_service_directory").Msg("Error al obtener la biografía extensa del psicólogo")
	}

	dto := &request_structs.PsiFullProfileDTO{
		FirstName:      psi.FirstName,
		SecondName:     psi.SecondName,
		LastName:       psi.LastName,
		SecondLastName: psi.SecondLastName,
		FPV:            psi.FPV,
		CI:             psi.CI,
		Gender:         psi.Genre,
		ProfilePicture: s.publicURL(psi.ProfilePictureS3Key),
		Solvent:        true,
		MiniBio:        psi.MiniBio,
		FullBioContent: fullBio,
		PrimaryWorkArea:      psi.PrimaryWorkArea,
		SecondaryWorkArea:    psi.SecondaryWorkArea,
		PrimarySpecialtyID:   psi.PrimarySpecialtyID,
		SecondarySpecialtyID: psi.SecondarySpecialtyID,
		PostGrades:           make([]request_structs.PostGradeDTO, 0),
		SocialNetworks:    make([]request_structs.SocialNetworkDTO, 0),
		Undergraduate:     request_structs.UndergraduateDTO{},
	}

	if psi.ShowContactEmail {
		dto.Email = psi.ContactEmail
	}

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

	if psi.ColData.ShowUniversityUndergraduate {
		dto.Undergraduate.University = psi.ColData.UniversityUndergraduate
	}
	if psi.ColData.ShowGraduateDate && !psi.ColData.GraduateDate.IsZero() {
		dto.Undergraduate.Date = psi.ColData.GraduateDate.Format("2006-01-02")
	}
	if psi.ColData.ShowMentionUndergraduate {
		dto.Undergraduate.Mention = psi.ColData.MentionUndergraduate
	}
	dto.Undergraduate.TitleImageOneURL = s.publicURL(psi.ColData.TitleImageOneS3Key)
	dto.Undergraduate.TitleImageTwoURL = s.publicURL(psi.ColData.TitleImageTwoS3Key)
	dto.Undergraduate.TitleImageThreeURL = s.publicURL(psi.ColData.TitleImageThreeS3Key)

	for _, sn := range psi.SocialNetworks {
		dto.SocialNetworks = append(dto.SocialNetworks, request_structs.SocialNetworkDTO{
			Name: sn.Name,
			URL:  sn.URL,
		})
	}

	for _, pg := range psi.PostGrades {
		if pg.Active {
			dto.PostGrades = append(dto.PostGrades, request_structs.PostGradeDTO{
				Type:        string(pg.Type),
				Title:       pg.Title,
				University:  pg.University,
				Year:        pg.GraduationYear,
				Description: pg.Description,
				PicOneURL:   s.publicURL(pg.PicOneS3Key),
				PicTwoURL:   s.publicURL(pg.PicTwoS3Key),
				PicThreeURL: s.publicURL(pg.PicThreeS3Key),
			})
		}
	}

	log.Debug().Str("component", "psi_service_directory").
		Str("university", dto.Undergraduate.University).
		Str("date", dto.Undergraduate.Date).
		Str("mention", dto.Undergraduate.Mention).
		Str("img_one", dto.Undergraduate.TitleImageOneURL).
		Str("img_two", dto.Undergraduate.TitleImageTwoURL).
		Str("img_three", dto.Undergraduate.TitleImageThreeURL).
		Msg("Undergraduate data being sent")

	return dto, psi.ID, nil
}

func (s *PsiService) GetPsiBioByID(ctx context.Context, id uuid.UUID) (string, error) {
	bio, err := s.repo.GetTextContentByID(ctx, id)
	if err != nil {
		return "", err
	}
	return bio, nil
}

func (s *PsiService) GetPsiSolvency(ctx context.Context, id uuid.UUID) ([]domain.PsiUserSolvency, error) {
	bio, err := s.repo.GetSolvencies(ctx, id)
	if err != nil {
		return []domain.PsiUserSolvency{}, err
	}
	return bio, nil
}

func (s *PsiService) GetSitemapPsis(ctx context.Context) (interface{}, error) {
	return s.repo.GetSitemapData(ctx)
}

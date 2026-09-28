package service

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

func TestPsiService_GetPublicDirectory_Pagination(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Paginación por defecto", func(t *testing.T) {
		repo.SearchDirectoryFunc = func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
			if filter.Page != 1 {
				t.Errorf("Page = %d, want 1", filter.Page)
			}
			if filter.Limit != 12 {
				t.Errorf("Limit = %d, want 12", filter.Limit)
			}
			return []domain.PsiUserModel{}, 0, nil
		}

		_, err := svc.GetPublicDirectory(context.Background(), request_structs.PsiDirectoryFilterDTO{})
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	})

	t.Run("Género normalizado a mayúsculas", func(t *testing.T) {
		repo.SearchDirectoryFunc = func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
			if filter.Gender != "F" {
				t.Errorf("Gender = %q, want F", filter.Gender)
			}
			return []domain.PsiUserModel{}, 0, nil
		}

		_, err := svc.GetPublicDirectory(context.Background(), request_structs.PsiDirectoryFilterDTO{Gender: "f"})
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	})

	t.Run("Género inválido se limpia", func(t *testing.T) {
		repo.SearchDirectoryFunc = func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
			if filter.Gender != "" {
				t.Errorf("Gender = %q, want empty", filter.Gender)
			}
			return []domain.PsiUserModel{}, 0, nil
		}

		_, err := svc.GetPublicDirectory(context.Background(), request_structs.PsiDirectoryFilterDTO{Gender: "X"})
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	})
}

func TestPsiService_GetPublicDirectory_MiniProfile(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Mini-perfil de solvente incluye especialidades", func(t *testing.T) {
		psiID := uuid.Must(uuid.NewV7())
		repo.SearchDirectoryFunc = func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
			return []domain.PsiUserModel{
				{
					ID:                psiID,
					FirstName:         "Ana",
					LastName:          "García",
					CI:                12345,
					FPV:               99999,
					Solvent:           true,
					PrimaryWorkArea:   "Clínica",
					SecondaryWorkArea: "Educación",
					MiniBio:           "Bio breve",
				},
			}, 1, nil
		}

		result, err := svc.GetPublicDirectory(context.Background(), request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 10})
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		fiberMap := result.(fiber.Map)
		data := fiberMap["data"].([]request_structs.PsiMiniProfileDTO)
		if len(data) != 1 {
			t.Fatalf("Expected 1 profile, got %d", len(data))
		}
		if len(data[0].Specialties) != 2 {
			t.Errorf("Expected 2 specialties, got %d", len(data[0].Specialties))
		}
	})

	t.Run("Mini-perfil de insolvente NO incluye especialidades", func(t *testing.T) {
		psiID := uuid.Must(uuid.NewV7())
		repo.SearchDirectoryFunc = func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
			return []domain.PsiUserModel{
				{
					ID:                psiID,
					FirstName:         "Luis",
					LastName:          "Pérez",
					CI:                54321,
					FPV:               88888,
					Solvent:           false,
					PrimaryWorkArea:   "Clínica",
					SecondaryWorkArea: "Educación",
					MiniBio:           "Bio breve",
				},
			}, 1, nil
		}

		result, err := svc.GetPublicDirectory(context.Background(), request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 10})
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		fiberMap := result.(fiber.Map)
		data := fiberMap["data"].([]request_structs.PsiMiniProfileDTO)
		if len(data) != 1 {
			t.Fatalf("Expected 1 profile, got %d", len(data))
		}
		if len(data[0].Specialties) != 0 {
			t.Errorf("Expected 0 specialties for insolvent, got %d", len(data[0].Specialties))
		}
	})

	t.Run("Paginación total_pages calculada correctamente", func(t *testing.T) {
		repo.SearchDirectoryFunc = func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
			return []domain.PsiUserModel{}, 25, nil
		}

		result, err := svc.GetPublicDirectory(context.Background(), request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 10})
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		fiberMap := result.(fiber.Map)
		totalPages := fiberMap["total_pages"].(int64)
		if totalPages != 3 {
			t.Errorf("total_pages = %d, want 3 (ceil(25/10))", totalPages)
		}
	})
}

func TestPsiService_GetPublicProfile(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Solvente: resuelve y expone sus áreas de desempeño", func(t *testing.T) {
		prim := uint32(1)
		sec := uint32(2)
		repo.GetByFPVFunc = func(ctx context.Context, fpv int) (domain.PsiUserModel, error) {
			return domain.PsiUserModel{
				FirstName:            "Ana",
				LastName:             "García",
				FPV:                  99999,
				CI:                   12345,
				Genre:                "F",
				Solvent:              true,
				PrimarySpecialtyID:   &prim,
				SecondarySpecialtyID: &sec,
				Credentials:          domain.Credentials{IsActive: true},
				ColData:              domain.PsiUserColData{},
			}, nil
		}
		repo.ResolveSpecialtyNamesFunc = func(ctx context.Context, ids []uint32, legacy []string) ([]string, error) {
			if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
				t.Errorf("ids = %v, want [1 2]", ids)
			}
			return []string{"Clínica", "Educación"}, nil
		}

		dto, _, err := svc.GetPublicProfile(context.Background(), 99999)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(dto.WorkAreas) != 2 || dto.WorkAreas[0] != "Clínica" || dto.WorkAreas[1] != "Educación" {
			t.Errorf("WorkAreas = %v, want [Clínica Educación]", dto.WorkAreas)
		}
	})

	t.Run("Solvente sin FK: áreas recuperadas por coincidencia exacta del legacy", func(t *testing.T) {
		repo.GetByFPVFunc = func(ctx context.Context, fpv int) (domain.PsiUserModel, error) {
			return domain.PsiUserModel{
				FirstName:         "Carl",
				LastName:          "Jung",
				FPV:               56504,
				CI:                123456,
				Genre:             "M",
				Solvent:           true,
				PrimaryWorkArea:   "Neuropsicología",
				SecondaryWorkArea: "Clínica",
				Credentials:       domain.Credentials{IsActive: true},
				ColData:           domain.PsiUserColData{},
			}, nil
		}
		repo.ResolveSpecialtyNamesFunc = func(ctx context.Context, ids []uint32, legacy []string) ([]string, error) {
			if len(ids) != 0 {
				t.Errorf("ids = %v, want vacío (sin FKs asignadas)", ids)
			}
			if len(legacy) != 2 || legacy[0] != "Neuropsicología" || legacy[1] != "Clínica" {
				t.Errorf("legacy = %v, want [Neuropsicología Clínica]", legacy)
			}
			return []string{"Neuropsicología", "Clínica"}, nil
		}

		dto, _, err := svc.GetPublicProfile(context.Background(), 56504)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(dto.WorkAreas) != 2 || dto.WorkAreas[0] != "Neuropsicología" || dto.WorkAreas[1] != "Clínica" {
			t.Errorf("WorkAreas = %v, want [Neuropsicología Clínica]", dto.WorkAreas)
		}
	})

	t.Run("Insolvente: perfil reducido sin áreas (privacidad de solvencia)", func(t *testing.T) {
		resolveCalled := false
		repo.GetByFPVFunc = func(ctx context.Context, fpv int) (domain.PsiUserModel, error) {
			return domain.PsiUserModel{
				FirstName: "Luis",
				LastName:  "Pérez",
				FPV:       88888,
				CI:        54321,
				Genre:     "M",
				Solvent:   false,
				Credentials: domain.Credentials{
					IsActive: true,
				},
				ColData: domain.PsiUserColData{},
			}, nil
		}
		repo.ResolveSpecialtyNamesFunc = func(ctx context.Context, ids []uint32, legacy []string) ([]string, error) {
			resolveCalled = true
			return []string{"Clínica"}, nil
		}

		dto, _, err := svc.GetPublicProfile(context.Background(), 88888)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if resolveCalled {
			t.Error("ResolveSpecialtyNames no debió invocarse para un insolvente")
		}
		if len(dto.WorkAreas) != 0 {
			t.Errorf("WorkAreas = %v, want vacío para insolvente", dto.WorkAreas)
		}
		if dto.FPV != 88888 || dto.FirstName != "Luis" {
			t.Errorf("La identidad pública debía conservarse, got FPV=%d first=%q", dto.FPV, dto.FirstName)
		}
		// La solvencia no se expone: PsiFullProfileDTO no tiene campo Solvent,
		// por lo que aquí verificamos que el DTO reducido tampoco filtre áreas.
		if dto.MiniBio != "" {
			t.Errorf("Un insolvente no debía exponer mini_bio, got %q", dto.MiniBio)
		}
	})

	t.Run("Inactivo: perfil no disponible", func(t *testing.T) {
		repo.GetByFPVFunc = func(ctx context.Context, fpv int) (domain.PsiUserModel, error) {
			return domain.PsiUserModel{
				FirstName: "Carla",
				LastName:  "Rojas",
				FPV:       77777,
				Solvent:   true,
				Credentials: domain.Credentials{
					IsActive: false,
				},
			}, nil
		}

		_, _, err := svc.GetPublicProfile(context.Background(), 77777)
		if err == nil || err.Error() != "perfil no disponible" {
			t.Errorf("Error = %v, want 'perfil no disponible'", err)
		}
	})
}

func TestPsiService_GetPsiBioByID(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Obtener biografía existente", func(t *testing.T) {
		bioID := uuid.Must(uuid.NewV7())
		repo.GetTextContentByIDFunc = func(ctx context.Context, id uuid.UUID) (string, error) {
			if id != bioID {
				t.Errorf("ID mismatch")
			}
			return "Mi biografía completa", nil
		}

		bio, err := svc.GetPsiBioByID(context.Background(), bioID)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if bio != "Mi biografía completa" {
			t.Errorf("Bio = %q", bio)
		}
	})

	t.Run("Error de DB retorna error", func(t *testing.T) {
		bioID := uuid.Must(uuid.NewV7())
		repo.GetTextContentByIDFunc = func(ctx context.Context, id uuid.UUID) (string, error) {
			return "", domain.ErrPsiNotFound
		}

		_, err := svc.GetPsiBioByID(context.Background(), bioID)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestPsiService_GetSolvencies(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Obtener solvencias", func(t *testing.T) {
		psiID := uuid.Must(uuid.NewV7())
		repo.GetSolvenciesFunc = func(ctx context.Context, id uuid.UUID) ([]domain.PsiUserSolvency, error) {
			return []domain.PsiUserSolvency{{PsiUserModelID: psiID}}, nil
		}

		sol, err := svc.GetPsiSolvency(context.Background(), psiID)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(sol) != 1 {
			t.Errorf("Expected 1 solvency, got %d", len(sol))
		}
	})
}

func TestPsiService_GetSitemapPsis(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Retorna datos del sitemap", func(t *testing.T) {
		repo.GetSitemapDataFunc = func(ctx context.Context) ([]domain.PsiUserModel, error) {
			return []domain.PsiUserModel{{FPV: 99999}}, nil
		}

		result, err := svc.GetSitemapPsis(context.Background())
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil sitemap data")
		}
	})
}

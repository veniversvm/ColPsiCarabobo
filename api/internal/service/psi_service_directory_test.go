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

	t.Run("Mini-perfil incluye especialidades", func(t *testing.T) {
		psiID := uuid.Must(uuid.NewV7())
		repo.SearchDirectoryFunc = func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
			return []domain.PsiUserModel{
				{
					ID:                psiID,
					FirstName:         "Ana",
					LastName:          "García",
					CI:                12345,
					FPV:               99999,
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

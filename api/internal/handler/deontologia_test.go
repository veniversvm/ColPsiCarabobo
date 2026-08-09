package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

func TestListDeontologiaByAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		entry := domain.PsiODeontologia{ID: uuid.New(), PsiUserID: psiID, Content: "Expediente A"}

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return testPsiUser(psiID), nil
			},
			ListDeontologiaByPsiIDFunc: func(_ context.Context, id uuid.UUID) ([]domain.PsiODeontologia, error) {
				return []domain.PsiODeontologia{entry}, nil
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}

		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/:id/deontologia", h.ListDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/"+psiID.String()+"/deontologia", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var body []domain.PsiODeontologia
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		require.Len(t, body, 1)
		require.Equal(t, "Expediente A", body[0].Content)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/:id/deontologia", h.ListDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/not-a-uuid/deontologia", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("psi_not_found", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return nil, domain.ErrPsiNotFound
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/:id/deontologia", h.ListDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/"+psiID.String()+"/deontologia", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("sin_token_es_404", func(t *testing.T) {
		adminRepo := &mockAdminRepo{}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/:id/deontologia", h.ListDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/"+uuid.New().String()+"/deontologia", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		// ProtectedAdmin404 enmascara como 404 (ver AGENTS.md).
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestAddDeontologiaByAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return testPsiUser(psiID), nil
			},
			CreateDeontologiaFunc: func(_ context.Context, entry *domain.PsiODeontologia) error {
				return nil
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodPost, "/psi/:id/deontologia", h.AddDeontologiaByAdmin, adminRepo, psiRepo)

		body := `{"content":"Expediente disciplinario abierto."}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/psi/"+psiID.String()+"/deontologia", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	t.Run("json_invalido", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return testPsiUser(psiID), nil
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodPost, "/psi/:id/deontologia", h.AddDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/psi/"+psiID.String()+"/deontologia", strings.NewReader("{malformed"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("contenido_vacio", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return testPsiUser(psiID), nil
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodPost, "/psi/:id/deontologia", h.AddDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/psi/"+psiID.String()+"/deontologia", strings.NewReader(`{"content":"   "}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("psi_not_found", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return nil, domain.ErrPsiNotFound
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodPost, "/psi/:id/deontologia", h.AddDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/psi/"+psiID.String()+"/deontologia", strings.NewReader(`{"content":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestDeleteDeontologiaByAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		entryID := uuid.New()
		psiRepo := &mockPsiRepo{
			GetDeontologiaByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiODeontologia, error) {
				return &domain.PsiODeontologia{ID: entryID, PsiUserID: psiID}, nil
			},
			DeleteDeontologiaFunc: func(_ context.Context, id uuid.UUID) error {
				return nil
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodDelete, "/psi/:id/deontologia/:entryId", h.DeleteDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/psi/"+psiID.String()+"/deontologia/"+entryID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("invalid_entry_uuid", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		psiRepo := &mockPsiRepo{}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodDelete, "/psi/:id/deontologia/:entryId", h.DeleteDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/psi/"+psiID.String()+"/deontologia/not-a-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("entry_not_found", func(t *testing.T) {
		admin := testAdmin(uuid.New(), true, true)
		psiID := uuid.New()
		entryID := uuid.New()
		psiRepo := &mockPsiRepo{
			GetDeontologiaByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiODeontologia, error) {
				return nil, domain.ErrDeontologiaNotFound
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})
		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodDelete, "/psi/:id/deontologia/:entryId", h.DeleteDeontologiaByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/psi/"+psiID.String()+"/deontologia/"+entryID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

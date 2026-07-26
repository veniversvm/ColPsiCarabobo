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
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

func TestGetSpecialties(t *testing.T) {
	t.Run("public", func(t *testing.T) {
		specs := []domain.PsiSpecialtyModel{
			*testSpecialty(1, "Clinica"),
			*testSpecialty(2, "Educacional"),
		}
		repo := &mockSpecialtyRepo{
			GetAllFunc: func(_ context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
				require.Equal(t, "active", status)
				return specs, nil
			},
		}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		app := setupPublicRoute(fiber.MethodGet, "/specialties", h.GetSpecialties)
		req := httptest.NewRequest(fiber.MethodGet, "/specialties", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result []domain.PsiSpecialtyModel
		json.NewDecoder(resp.Body).Decode(&result)
		require.Len(t, result, 2)
	})

	t.Run("admin_with_status_filter", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)
		admin.CanReadNotifications = true

		repo := &mockSpecialtyRepo{
			GetAllFunc: func(_ context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
				require.Equal(t, "all", status)
				return nil, nil
			},
		}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		app := newTestApp()
		app.Get("/specialties", func(c *fiber.Ctx) error {
			c.Locals("admin", admin)
			return h.GetSpecialties(c)
		})
		req := httptest.NewRequest(fiber.MethodGet, "/specialties?status=all", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestGetSpecialtyByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		spec := testSpecialty(1, "Clinica")
		repo := &mockSpecialtyRepo{
			GetByIDFunc: func(_ context.Context, id uint32, active bool) (*domain.PsiSpecialtyModel, error) {
				return spec, nil
			},
		}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		app := setupPublicRoute(fiber.MethodGet, "/specialties/:id", h.GetSpecialtyByID)
		req := httptest.NewRequest(fiber.MethodGet, "/specialties/1", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.PsiSpecialtyModel
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, "Clinica", result.Name)
	})

	t.Run("not_found", func(t *testing.T) {
		repo := &mockSpecialtyRepo{
			GetByIDFunc: func(_ context.Context, id uint32, active bool) (*domain.PsiSpecialtyModel, error) {
				return nil, domain.ErrPsiNotFound
			},
		}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		app := setupPublicRoute(fiber.MethodGet, "/specialties/:id", h.GetSpecialtyByID)
		req := httptest.NewRequest(fiber.MethodGet, "/specialties/999", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid_id", func(t *testing.T) {
		svc := service.NewSpecialtyService(&mockSpecialtyRepo{})
		h := NewSpecialtyHandler(svc)

		app := setupPublicRoute(fiber.MethodGet, "/specialties/:id", h.GetSpecialtyByID)
		req := httptest.NewRequest(fiber.MethodGet, "/specialties/abc", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestCreateSpecialty(t *testing.T) {
	t.Run("with_permission", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		repo := &mockSpecialtyRepo{
			CreateFunc: func(_ context.Context, s *domain.PsiSpecialtyModel) error {
				return nil
			},
		}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupAdminRoute(fiber.MethodPost, "/specialties", h.CreateSpecialty, adminRepo, &mockPsiRepo{})

		body := `{"name":"Nueva Especialidad","description":"Test"}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/specialties", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	t.Run("insufficient_permission", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, false, false) // No CanCreateTags
		admin.CanCreateTags = false

		repo := &mockSpecialtyRepo{}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupAdminRoute(fiber.MethodPost, "/specialties", h.CreateSpecialty, adminRepo, &mockPsiRepo{})

		body := `{"name":"Test","description":"Test"}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/specialties", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}

func TestUpdateSpecialty(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		repo := &mockSpecialtyRepo{
			GetByAdminIDFunc: func(_ context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
				return testSpecialty(id, "Old Name"), nil
			},
			UpdateFunc: func(_ context.Context, s *domain.PsiSpecialtyModel) error {
				return nil
			},
		}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupAdminRoute(fiber.MethodPatch, "/specialties/:id", h.UpdateSpecialty, adminRepo, &mockPsiRepo{})

		body := `{"name":"Updated Name"}`
		req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/admin/specialties/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestDeleteSpecialty(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		repo := &mockSpecialtyRepo{
			DeleteFunc: func(_ context.Context, id uint32) error {
				return nil
			},
		}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupAdminRoute(fiber.MethodDelete, "/specialties/:id", h.DeleteSpecialty, adminRepo, &mockPsiRepo{})

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/specialties/1", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestCountSpecialties(t *testing.T) {
	t.Run("public", func(t *testing.T) {
		repo := &mockSpecialtyRepo{
			CountFunc: func(_ context.Context, active *bool) (int64, error) {
				require.NotNil(t, active)
				require.True(t, *active)
				return 15, nil
			},
		}
		svc := service.NewSpecialtyService(repo)
		h := NewSpecialtyHandler(svc)

		app := setupPublicRoute(fiber.MethodGet, "/specialties/count", h.CountSpecialties)
		req := httptest.NewRequest(fiber.MethodGet, "/specialties/count", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, float64(15), result["count"])
	})
}

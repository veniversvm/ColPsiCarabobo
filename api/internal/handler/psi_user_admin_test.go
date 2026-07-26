package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// HELPERS
// =========================================================================

func testPsiAdminHandler(psiRepo *mockPsiRepo, adminRepo *mockAdminRepo, analyticsRepo *mockAnalyticsRepo) *PsiHandler {
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)
	mailSvc := &mockMailService{}
	psiSvc := service.NewPsiService(psiRepo, nil, mailSvc)
	return NewPsiHandler(psiSvc, analyticsSvc)
}

func adminTestToken(admin *domain.UserAdmin) string {
	return generateTestToken(admin.ID.String(), "admin", time.Now().Add(24*time.Hour))
}

// setupAdminRouteForTest creates a Fiber app with ProtectedAdmin404 + a single route.
// Routes are registered under /api/v1/admin/<path> to match the real router.
func setupAdminRouteForTest(method, path string, handler fiber.Handler, adminRepo *mockAdminRepo, psiRepo *mockPsiRepo) *fiber.App {
	app := newTestApp()
	mw := buildAuthMiddleware(adminRepo, psiRepo)
	group := app.Group("/api/v1/admin", mw.ProtectedAdmin404())
	group.Add(method, path, handler)
	return app
}

// =========================================================================
// GetPsiByIDAdmin
// =========================================================================

func TestGetPsiByIDAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		psiID := uuid.New()
		psi := testPsiUser(psiID)

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
		}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/:id", h.GetPsiByIDAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/"+psiID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/:id", h.GetPsiByIDAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/not-a-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("not_found", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

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
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/:id", h.GetPsiByIDAdmin, adminRepo, psiRepo)

		targetID := uuid.New()
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/"+targetID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =========================================================================
// CreatePsiByAdmin
// =========================================================================

func TestCreatePsiByAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		psiRepo := &mockPsiRepo{}
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodPost, "/psi/create", h.CreatePsiByAdmin, adminRepo, psiRepo)

		body := `{
			"username":"newpsi@test.com",
			"email":"newpsi@test.com",
			"password":"Strong123!@#",
			"first_name":"Juan",
			"last_name":"Perez",
			"ci":12345678,
			"fpv":99999,
			"nationality":"V",
			"born_date":"1990-01-01",
			"genre":"M"
		}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/psi/create", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	t.Run("invalid_body", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodPost, "/psi/create", h.CreatePsiByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/psi/create", strings.NewReader("not-json"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =========================================================================
// UpdatePsiByAdmin
// =========================================================================

func TestUpdatePsiByAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

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
		app := setupAdminRouteForTest(fiber.MethodPatch, "/psi/:id", h.UpdatePsiByAdmin, adminRepo, psiRepo)

		body := `{"first_name":"UpdatedName"}`
		req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/admin/psi/"+psiID.String(), strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodPatch, "/psi/:id", h.UpdatePsiByAdmin, adminRepo, psiRepo)

		body := `{"first_name":"UpdatedName"}`
		req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/admin/psi/not-a-uuid", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("empty_request_body", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		psiID := uuid.New()

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodPatch, "/psi/:id", h.UpdatePsiByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/admin/psi/"+psiID.String(), strings.NewReader("{}"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =========================================================================
// DeletePsiByAdmin
// =========================================================================

func TestDeletePsiByAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		psiID := uuid.New()

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodDelete, "/psi/:id", h.DeletePsiByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/psi/"+psiID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodDelete, "/psi/:id", h.DeletePsiByAdmin, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/psi/not-a-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =========================================================================
// ListAllPsis
// =========================================================================

func TestListAllPsis(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/list", h.ListAllPsis, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/list?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.True(t, resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusForbidden)
	})

	t.Run("with_query_params", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		psiRepo := &mockPsiRepo{}
		h := testPsiAdminHandler(psiRepo, adminRepo, &mockAnalyticsRepo{})

		token := adminTestToken(admin)
		app := setupAdminRouteForTest(fiber.MethodGet, "/psi/list", h.ListAllPsis, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/psi/list?q=test&location=Valencia&gender=M&specialty=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.True(t, resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusForbidden)
	})
}

// =========================================================================
// LoginLibrary
// =========================================================================

func TestLoginLibrary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("secret1234"), bcrypt.DefaultCost)
		psiID := uuid.New()
		psi := testPsiUser(psiID)
		psi.Password = string(hashedPass)

		psiRepo := &mockPsiRepo{
			GetByIdentifierFunc: func(_ context.Context, id string) (*domain.PsiUserModel, error) {
				return psi, nil
			},
			UpdateKeyFunc: func(_ context.Context, p *domain.PsiUserModel) error {
				return nil
			},
		}
		h, _ := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodPost, "/psi/login-library", h.LoginLibrary)
		body := `{"identifier":"psi@test.com","password":"secret1234"}`
		req := httptest.NewRequest(fiber.MethodPost, "/psi/login-library", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.NotEmpty(t, result["token"])
		require.Equal(t, "Acceso a la biblioteca", result["message"])
	})

	t.Run("invalid_password", func(t *testing.T) {
		psiRepo := &mockPsiRepo{
			GetByIdentifierFunc: func(_ context.Context, id string) (*domain.PsiUserModel, error) {
				psi := testPsiUser(uuid.New())
				psi.Password = "$2a$10$invalidhashwillnevermatch"
				return psi, nil
			},
		}
		h, _ := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodPost, "/psi/login-library", h.LoginLibrary)
		body := `{"identifier":"psi@test.com","password":"wrongpassword"}`
		req := httptest.NewRequest(fiber.MethodPost, "/psi/login-library", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("not_found", func(t *testing.T) {
		psiRepo := &mockPsiRepo{
			GetByIdentifierFunc: func(_ context.Context, id string) (*domain.PsiUserModel, error) {
				return nil, domain.ErrPsiNotFound
			},
		}
		h, _ := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodPost, "/psi/login-library", h.LoginLibrary)
		body := `{"identifier":"nobody@test.com","password":"secret1234"}`
		req := httptest.NewRequest(fiber.MethodPost, "/psi/login-library", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid_body", func(t *testing.T) {
		h, _ := testPsiHandler(&mockPsiRepo{}, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodPost, "/psi/login-library", h.LoginLibrary)
		req := httptest.NewRequest(fiber.MethodPost, "/psi/login-library", strings.NewReader("not-json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

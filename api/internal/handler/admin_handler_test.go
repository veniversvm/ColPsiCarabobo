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

func TestAdminLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("secret1234"), bcrypt.DefaultCost)
		admin := testAdmin(uuid.New(), true, true)
		admin.Password = string(hashedPass)
		adminRepo := &mockAdminRepo{
			GetByIdentifierFunc: func(_ context.Context, id string) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		app := setupPublicRoute(fiber.MethodPost, "/auth/login", h.Login)
		body := `{"identifier":"admin@test.com","password":"secret1234"}`
		req := httptest.NewRequest(fiber.MethodPost, "/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.NotEmpty(t, result["token"], "should return a JWT token")
		require.Equal(t, "Bienvenido al sistema", result["message"])
	})

	t.Run("invalid_credentials", func(t *testing.T) {
		adminRepo := &mockAdminRepo{
			GetByIdentifierFunc: func(_ context.Context, id string) (*domain.UserAdmin, error) {
				return &domain.UserAdmin{
					Credentials: domain.Credentials{
						Password: "$2a$10$invalidhashthatwillnevermatch", // hash of something else
						Key:      testSecret,
						IsActive: true,
					},
				}, nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		app := setupPublicRoute(fiber.MethodPost, "/auth/login", h.Login)
		body := `{"identifier":"admin@test.com","password":"wrongpassword"}`
		req := httptest.NewRequest(fiber.MethodPost, "/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Contains(t, result["error"], "credenciales")
	})

	t.Run("not_found", func(t *testing.T) {
		adminRepo := &mockAdminRepo{
			GetByIdentifierFunc: func(_ context.Context, id string) (*domain.UserAdmin, error) {
				return nil, domain.ErrPsiNotFound
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		app := setupPublicRoute(fiber.MethodPost, "/auth/login", h.Login)
		body := `{"identifier":"nobody@test.com","password":"secret1234"}`
		req := httptest.NewRequest(fiber.MethodPost, "/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid_body", func(t *testing.T) {
		svc := service.NewAdminService(&mockAdminRepo{}, nil)
		h := NewAdminHandler(svc)

		app := setupPublicRoute(fiber.MethodPost, "/auth/login", h.Login)
		req := httptest.NewRequest(fiber.MethodPost, "/auth/login", strings.NewReader("not-json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Contains(t, result["error"], "JSON")
	})
}

func TestAdminLogout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
			UpdateKeyFunc: func(_ context.Context, user *domain.UserAdmin) error {
				return nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodPost, "/logout", h.Logout, adminRepo, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/logout", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, "Sesión cerrada correctamente", result["message"])
	})

	t.Run("no_auth", func(t *testing.T) {
		svc := service.NewAdminService(&mockAdminRepo{}, nil)
		h := NewAdminHandler(svc)

		adminRepo := &mockAdminRepo{}
		app := setupAdminRoute(fiber.MethodPost, "/logout", h.Logout, adminRepo, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/logout", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		// ProtectedAdmin404 returns 404 when no token
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestAdminCreateAdmin(t *testing.T) {
	t.Run("sudo_creates", func(t *testing.T) {
		creator := testAdmin(uuid.New(), true, true)
		creatorRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return creator, nil
			},
			CreateFunc: func(_ context.Context, user *domain.UserAdmin) error {
				return nil
			},
			CountSudosFunc: func(_ context.Context) (int64, error) {
				return 1, nil
			},
		}
		svc := service.NewAdminService(creatorRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(creator.ID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodPost, "/create", h.CreateAdmin, creatorRepo, &mockPsiRepo{})

		body := `{"username":"newadmin@test.com","email":"newadmin@test.com","password":"StrongPass123!","permissions":{"can_create_psi":true}}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Contains(t, result["message"], "creado")
	})

	t.Run("insufficient_permissions", func(t *testing.T) {
		creator := testAdmin(uuid.New(), false, false) // No CanCreateAdmin
		creatorRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return creator, nil
			},
		}
		svc := service.NewAdminService(creatorRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(creator.ID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodPost, "/create", h.CreateAdmin, creatorRepo, &mockPsiRepo{})

		body := `{"username":"newadmin@test.com","email":"newadmin@test.com","password":"StrongPass123!","permissions":{}}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Contains(t, result["error"], "permisos")
	})

	t.Run("invalid_body", func(t *testing.T) {
		creator := testAdmin(uuid.New(), true, true)
		creatorRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return creator, nil
			},
		}
		svc := service.NewAdminService(creatorRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(creator.ID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodPost, "/create", h.CreateAdmin, creatorRepo, &mockPsiRepo{})

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/create", strings.NewReader("not-json"))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestAdminGetAdmins(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)
		admins := []domain.UserAdmin{*admin}

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
			ListFunc: func(_ context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
				return admins, 1, nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodGet, "/list", h.GetAdmins, adminRepo, &mockPsiRepo{})

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/list?page=1&limit=10", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.NotNil(t, result["data"])
		require.Equal(t, float64(1), result["total"])
	})

	t.Run("no_auth", func(t *testing.T) {
		svc := service.NewAdminService(&mockAdminRepo{}, nil)
		h := NewAdminHandler(svc)

		adminRepo := &mockAdminRepo{}
		app := setupAdminRoute(fiber.MethodGet, "/list", h.GetAdmins, adminRepo, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/list", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with_search", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
			ListFunc: func(_ context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
				require.Equal(t, "test", search)
				return nil, 0, nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodGet, "/list", h.GetAdmins, adminRepo, &mockPsiRepo{})

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/list?search=test", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestAdminUpdateAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		updaterID := uuid.New()
		updater := testAdmin(updaterID, true, true)
		targetID := uuid.New()

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				if id == updaterID {
					return updater, nil
				}
				return testAdmin(id, false, false), nil
			},
			UpdateFunc: func(_ context.Context, user *domain.UserAdmin) error {
				return nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(updaterID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodPatch, "/update", h.UpdateAdmin, adminRepo, &mockPsiRepo{})

		body := `{"id":"` + targetID.String() + `","username":"updated_name"}`
		req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/admin/update", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Contains(t, result["message"], "actualizado")
	})
}

func TestAdminDeleteAdmin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		updaterID := uuid.New()
		targetID := uuid.New()

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return testAdmin(id, false, false), nil
			},
			DeleteFunc: func(_ context.Context, id uuid.UUID) error {
				return nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(updaterID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodDelete, "/delete/:id", h.DeleteAdmin, adminRepo, &mockPsiRepo{})

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/delete/"+targetID.String(), nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Contains(t, result["message"], "eliminado")
	})

	t.Run("invalid_id", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodDelete, "/delete/:id", h.DeleteAdmin, adminRepo, &mockPsiRepo{})

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/delete/not-a-uuid", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("cannot_delete_self", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		svc := service.NewAdminService(adminRepo, nil)
		h := NewAdminHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		app := setupAdminRoute(fiber.MethodDelete, "/delete/:id", h.DeleteAdmin, adminRepo, &mockPsiRepo{})

		// Try to delete self
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/admin/delete/"+adminID.String(), nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Contains(t, result["error"], "propia cuenta")
	})
}

// futureTime returns a time 1 hour in the future for valid JWT tokens.
func futureTime() time.Time {
	return time.Now().Add(1 * time.Hour)
}

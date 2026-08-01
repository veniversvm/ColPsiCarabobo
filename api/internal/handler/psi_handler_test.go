package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// MOCK PSI SERVICE (wraps real PsiService with mock repo)
// =========================================================================

// testPsiHandler creates a PsiHandler with mock repos for both PsiService and AnalyticsService.
func testPsiHandler(psiRepo *mockPsiRepo, analyticsRepo *mockAnalyticsRepo) (*PsiHandler, *mockAdminRepo) {
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)
	mailSvc := &mockMailService{}
	psiSvc := service.NewPsiService(psiRepo, nil, mailSvc) // nil S3, mock IMailService
	h := NewPsiHandler(psiSvc, analyticsSvc)
	adminRepo := &mockAdminRepo{}
	return h, adminRepo
}

func TestPsiLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hashedPass, _ := generateBcryptHash("secret1234")
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

		app := setupPublicRoute(fiber.MethodPost, "/psi/login", h.Login)
		body := `{"identifier":"psi@test.com","password":"secret1234"}`
		req := httptest.NewRequest(fiber.MethodPost, "/psi/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.NotEmpty(t, result["token"])
		require.Equal(t, "Bienvenido colega", result["message"])
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

		app := setupPublicRoute(fiber.MethodPost, "/psi/login", h.Login)
		body := `{"identifier":"psi@test.com","password":"wrongpassword"}`
		req := httptest.NewRequest(fiber.MethodPost, "/psi/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})
}

func TestSearchDirectory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		psiRepo := &mockPsiRepo{
			SearchDirectoryFunc: func(_ context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
				return []domain.PsiUserModel{*testPsiUser(uuid.New())}, 1, nil
			},
		}
		h, _ := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodGet, "/psi/directory", h.SearchDirectory)
		req := httptest.NewRequest(fiber.MethodGet, "/psi/directory?q=test&location=Valencia", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("sql_injection_sanitized", func(t *testing.T) {
		psiRepo := &mockPsiRepo{
			SearchDirectoryFunc: func(_ context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
				// The sanitizer should have cleaned the SQL injection attempt
				require.NotContains(t, filter.SearchTerm, "'")
				return nil, 0, nil
			},
		}
		h, _ := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodGet, "/psi/directory", h.SearchDirectory)
		req := httptest.NewRequest(fiber.MethodGet, "/psi/directory?q=%27%3B%20DROP%20TABLE--", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestGetPublicProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		psiRepo := &mockPsiRepo{
			GetByFPVFunc: func(_ context.Context, id int) (domain.PsiUserModel, error) {
				return *testPsiUser(uuid.New()), nil
			},
		}
		h, _ := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodGet, "/psi/:id", h.GetPublicProfile)
		req := httptest.NewRequest(fiber.MethodGet, "/psi/12345", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("not_found", func(t *testing.T) {
		psiRepo := &mockPsiRepo{
			GetByFPVFunc: func(_ context.Context, id int) (domain.PsiUserModel, error) {
				return domain.PsiUserModel{}, domain.ErrPsiNotFound
			},
		}
		h, _ := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodGet, "/psi/:id", h.GetPublicProfile)
		req := httptest.NewRequest(fiber.MethodGet, "/psi/99999", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestPsiGetMe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		psiID := uuid.New()
		psi := testPsiUser(psiID)

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
			GetPsiUserColDataFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserColData, error) {
				return &domain.PsiUserColData{}, nil
			},
		}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		token := generateTestToken(psiID.String(), "psi", futureTime())
		app := setupPsiRoute(fiber.MethodGet, "/", h.GetMe, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/psi/me/", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.PsiUserModel
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, psi.FirstName, result.FirstName)
	})

	t.Run("no_auth", func(t *testing.T) {
		psiRepo := &mockPsiRepo{}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPsiRoute(fiber.MethodGet, "/", h.GetMe, adminRepo, psiRepo)
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/psi/me/", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})
}

func TestGetAudiobookshelfAccess(t *testing.T) {
	// Fake ABS que responde login con un accessToken.
	absSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			w.Write([]byte(`{"user":{"id":"abs-id-123","username":"psi_12345678","accessToken":"tok-abc"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer absSrv.Close()

	absSvc := service.NewAudiobookshelfService(absSrv.URL, "https://abs.public", "admin", "pass", "secret-test")

	t.Run("solvent_returns_url", func(t *testing.T) {
		psi := testPsiUser(uuid.New())
		psi.Solvent = true
		psi.CI = 12345678

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
		}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})
		h.SetAudiobookshelf(absSvc)

		token := generateTestToken(psi.ID.String(), "psi", futureTime())
		app := setupPsiRoute(fiber.MethodGet, "/audiobookshelf", h.GetAudiobookshelfAccess, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/psi/me/audiobookshelf", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		result := decodeBody(resp)
		require.Equal(t, "https://abs.public/login/?accessToken=tok-abc", result["url"])
		require.Equal(t, "psi_12345678", result["username"])
	})

	t.Run("not_solvent_403", func(t *testing.T) {
		psi := testPsiUser(uuid.New())
		psi.Solvent = false

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
		}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})
		h.SetAudiobookshelf(absSvc)

		token := generateTestToken(psi.ID.String(), "psi", futureTime())
		app := setupPsiRoute(fiber.MethodGet, "/audiobookshelf", h.GetAudiobookshelfAccess, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/psi/me/audiobookshelf", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("no_auth_401", func(t *testing.T) {
		h, adminRepo := testPsiHandler(&mockPsiRepo{}, &mockAnalyticsRepo{})
		h.SetAudiobookshelf(absSvc)

		app := setupPsiRoute(fiber.MethodGet, "/audiobookshelf", h.GetAudiobookshelfAccess, adminRepo, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/psi/me/audiobookshelf", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("service_unconfigured_503", func(t *testing.T) {
		psi := testPsiUser(uuid.New())
		psi.Solvent = true

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
		}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		token := generateTestToken(psi.ID.String(), "psi", futureTime())
		app := setupPsiRoute(fiber.MethodGet, "/audiobookshelf", h.GetAudiobookshelfAccess, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/psi/me/audiobookshelf", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	})
}

func TestPsiLogout(t *testing.T) {	t.Run("success", func(t *testing.T) {
		psiID := uuid.New()
		psi := testPsiUser(psiID)

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
			UpdateKeyFunc: func(_ context.Context, p *domain.PsiUserModel) error {
				return nil
			},
		}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		token := generateTestToken(psiID.String(), "psi", futureTime())
		app := setupPsiRoute(fiber.MethodPost, "/logout", h.Logout, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/psi/me/logout", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, "Sesión cerrada correctamente", result["message"])
	})
}

func TestPsiAddSocialNetwork(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		psiID := uuid.New()
		psi := testPsiUser(psiID)

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
			CreateSocialNetworkFunc: func(_ context.Context, sn *domain.PsiUserSocialNetwork) error {
				return nil
			},
			CountSocialNetworksByPsiIDFunc: func(_ context.Context, id uuid.UUID) (int64, error) {
				return 2, nil // under limit
			},
		}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		token := generateTestToken(psiID.String(), "psi", futureTime())
		app := setupPsiRoute(fiber.MethodPost, "/social", h.AddSocialNetwork, adminRepo, psiRepo)

		body := `{"name":"Instagram","url":"https://instagram.com/test"}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/psi/me/social", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})
}

func TestPsiUpdateSocialNetwork(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		psiID := uuid.New()
		psi := testPsiUser(psiID)
		netID := uuid.New()

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
			GetSocialNetworkByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
				return &domain.PsiUserSocialNetwork{
					ID:        netID,
					PsiUserID: psiID,
					Name:      "Instagram",
					URL:       "https://instagram.com/old",
				}, nil
			},
			UpdateSocialNetworkFunc: func(_ context.Context, sn *domain.PsiUserSocialNetwork) error {
				return nil
			},
		}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		token := generateTestToken(psiID.String(), "psi", futureTime())
		app := setupPsiRoute(fiber.MethodPatch, "/social/:id", h.UpdateSocialNetwork, adminRepo, psiRepo)

		body := `{"url":"https://instagram.com/new"}`
		req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/psi/me/social/"+netID.String(), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestPsiDeleteSocialNetwork(t *testing.T) {
	t.Run("success_own_network", func(t *testing.T) {
		psiID := uuid.New()
		psi := testPsiUser(psiID)
		netID := uuid.New()

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
			GetSocialNetworkByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
				return &domain.PsiUserSocialNetwork{
					ID:        netID,
					PsiUserID: psiID,
				}, nil
			},
			DeleteSocialNetworkFunc: func(_ context.Context, id uuid.UUID) error {
				return nil
			},
		}
		h, adminRepo := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		token := generateTestToken(psiID.String(), "psi", futureTime())
		app := setupPsiRoute(fiber.MethodDelete, "/social/:id", h.DeleteSocialNetwork, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/psi/me/social/"+netID.String(), nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestGetSitemapData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		psiRepo := &mockPsiRepo{
			GetSitemapDataFunc: func(_ context.Context) ([]domain.PsiUserModel, error) {
				return []domain.PsiUserModel{*testPsiUser(uuid.New())}, nil
			},
		}
		h, _ := testPsiHandler(psiRepo, &mockAnalyticsRepo{})

		app := setupPublicRoute(fiber.MethodGet, "/psi/public/sitemap-data", h.GetSitemapData)
		req := httptest.NewRequest(fiber.MethodGet, "/psi/public/sitemap-data", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

// =========================================================================
// HELPERS
// =========================================================================

func generateBcryptHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

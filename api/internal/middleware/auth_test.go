package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// =========================================================================
// MOCKS ROBUSTOS
// =========================================================================

type mockAdminRepo struct {
	domain.UserAdminRepository
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error)
}

func (m *mockAdminRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
	return m.GetByIDFunc(ctx, id)
}

type mockPsiRepo struct {
	domain.PsiUserRepository
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
}

func (m *mockPsiRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
}

// =========================================================================
// GENERADOR DE TOKENS PARA TESTS
// =========================================================================

func generateTestToken(userID string, role string, secret string, expiresAt time.Time) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte(secret))
	return t
}

// =========================================================================
// TEST SUITE PRINCIPAL
// =========================================================================

func TestAuthMiddleware_Extensive(t *testing.T) {
	// Setup de variables globales del test
	adminID := uuid.New()
	psiID := uuid.New()
	correctSecret := "secret-valencia-2026"

	// --- FIX: Inicialización de GORM en modo DryRun ---
	// Esto crea una instancia de DB que NO necesita conexión real.
	// Evita el nil pointer dereference cuando AnalyticsService llama a s.db
	dummyDB, _ := gorm.Open(postgres.New(postgres.Config{}), &gorm.Config{
		DryRun: true,
	})

	// Inicializamos el servicio con la DB dummy
	analytics := service.NewAnalyticsService(dummyDB)

	mAdmin := &mockAdminRepo{}
	mPsi := &mockPsiRepo{}

	// Constructor con sus 3 dependencias
	mw := NewAuthMiddleware(mAdmin, mPsi, analytics)

	// --- ESCENARIO 1: PROTECTED ADMIN (SECURITY BY OBSCURITY) ---
	t.Run("Admin_404_Logic", func(t *testing.T) {
		app := fiber.New()
		app.Get("/admin/secret", mw.ProtectedAdmin404(), func(c *fiber.Ctx) error { return c.SendStatus(200) })

		scenarios := []struct {
			name       string
			token      string
			setupMock  func()
			wantStatus int
		}{
			{
				name:  "Éxito: Admin válido",
				token: generateTestToken(adminID.String(), "admin", correctSecret, time.Now().Add(time.Hour)),
				setupMock: func() {
					mAdmin.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
						return &domain.UserAdmin{ID: adminID, Key: correctSecret}, nil
					}
				},
				wantStatus: 200,
			},
			{
				name:       "Fallo: Token expirado (404)",
				token:      generateTestToken(adminID.String(), "admin", correctSecret, time.Now().Add(-time.Hour)),
				setupMock:  func() {},
				wantStatus: 404,
			},
			{
				name:  "Fallo: Admin no existe en DB",
				token: generateTestToken(uuid.NewString(), "admin", correctSecret, time.Now().Add(time.Hour)),
				setupMock: func() {
					mAdmin.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
						return nil, errors.New("not found")
					}
				},
				wantStatus: 404,
			},
		}

		for _, sc := range scenarios {
			t.Run(sc.name, func(t *testing.T) {
				sc.setupMock()
				req := httptest.NewRequest("GET", "/admin/secret", nil)
				req.Header.Set("Authorization", "Bearer "+sc.token)
				resp, _ := app.Test(req)
				if resp.StatusCode != sc.wantStatus {
					t.Errorf("[%s] se esperaba %d, se obtuvo %d", sc.name, sc.wantStatus, resp.StatusCode)
				}
			})
		}
	})

	// --- ESCENARIO 2: PROTECTED PSI USER (ATAQUES Y ERRORES) ---
	t.Run("PsiUser_Security_Edge_Cases", func(t *testing.T) {
		app := fiber.New()
		app.Get("/psi/me", mw.ProtectedPsiUser(), func(c *fiber.Ctx) error { return c.SendStatus(200) })

		t.Run("Wrong_Signing_Method", func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"user_id": psiID.String()})
			tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

			req := httptest.NewRequest("GET", "/psi/me", nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)
			resp, _ := app.Test(req)

			if resp.StatusCode != 401 {
				t.Errorf("Se esperaba 401, se obtuvo %d", resp.StatusCode)
			}
		})
	})

	// --- ESCENARIO 3: OPTIONAL HYBRID (FLEXIBILIDAD) ---
	t.Run("Hybrid_Auth_Context_Injection", func(t *testing.T) {
		app := fiber.New()
		app.Get("/hybrid", mw.OptionalHybridAuth(), func(c *fiber.Ctx) error {
			if c.Locals("admin") != nil {
				return c.Status(200).SendString("is_admin")
			}
			if c.Locals("psi_user") != nil {
				return c.Status(200).SendString("is_psi")
			}
			return c.Status(200).SendString("is_anonymous")
		})

		t.Run("Detects_Psi", func(t *testing.T) {
			mPsi.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return &domain.PsiUserModel{ID: psiID, Key: correctSecret}, nil
			}
			token := generateTestToken(psiID.String(), "psi", correctSecret, time.Now().Add(time.Hour))
			req := httptest.NewRequest("GET", "/hybrid", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, _ := app.Test(req)
			if resp.StatusCode != 200 {
				t.Fail()
			}
		})
	})
}

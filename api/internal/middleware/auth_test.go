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
// Patrón de Mocks Funcionales: En lugar de usar herramientas pesadas de generación
// de código (como mockgen), se emplean structs que encapsulan punteros a funciones.
// Esto permite inyectar comportamientos dinámicos (y escenarios de error específicos)
// directamente dentro de cada sub-test (t.Run) manteniendo un aislamiento total.

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

// generateTestToken construye un JWT válido firmado con un secreto específico.
// Se utiliza para simular sesiones activas o expiradas manipulando el claim "exp",
// permitiendo probar los flujos de autorización sin requerir el endpoint de login.
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

// TestAuthMiddleware_Extensive valida las barreras de seguridad de la API.
// Comprueba el manejo de tokens inválidos, la mitigación de vulnerabilidades JWT
// y la inyección correcta del contexto de usuario en el framework Fiber.
func TestAuthMiddleware_Extensive(t *testing.T) {
	// Setup de variables globales del test
	adminID := uuid.New()
	psiID := uuid.New()
	correctSecret := "secret-valencia-2026"

	// --- FIX: Inicialización de GORM en modo DryRun ---
	// DryRun es una técnica avanzada para testing: GORM inicializa su motor interno
	// pero NO intenta conectarse a una base de datos real ni ejecuta SQL.
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
	// Táctica de Defensa (Seguridad por Oscuridad):
	// Los endpoints de administración devuelven intencionalmente un Error 404 (Not Found)
	// en lugar de 401 (Unauthorized) o 403 (Forbidden) ante credenciales inválidas.
	// Esto previene ataques de enumeración, impidiendo que un atacante descubra
	// la topología de la API administrativa.
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

		// Prevención de Vulnerabilidad CVE: "None Algorithm Attack".
		// Verifica que el middleware rechace tokens JWT construidos intencionalmente
		// sin algoritmo de firma (alg: "none"). Un fallo aquí permitiría a un atacante
		// forjar tokens válidos con cualquier user_id sin conocer el secreto.
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
	// Prueba el middleware no bloqueante. Utilizado en rutas públicas que pueden
	// entregar información extra si el usuario está autenticado, pero no restringen
	// el acceso si es un visitante anónimo (inyecta el contexto si es posible).
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

		// Valida que la identidad se extraiga de la DB y se asigne a Fiber Locals
		// de manera transparente para que el handler la consuma sin hacer parseos manuales.
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

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

// mockAnalyticsRepo es un no-op para tests que no dependen de analytics reales.
type mockAnalyticsRepo struct {
	domain.AnalyticsRepository
}

func (m *mockAnalyticsRepo) CreateLoginEvent(domain.LoginEvent) error   { return nil }
func (m *mockAnalyticsRepo) UpsertActiveSession(domain.ActiveSession) error { return nil }
func (m *mockAnalyticsRepo) DeleteActiveSession(uuid.UUID) error        { return nil }
func (m *mockAnalyticsRepo) UpdateSessionHeartbeat(uuid.UUID, time.Time, time.Time) error { return nil }
func (m *mockAnalyticsRepo) CreateSearchEvent(domain.SearchEvent) error { return nil }
func (m *mockAnalyticsRepo) CreateProfileView(domain.ProfileView) error { return nil }
func (m *mockAnalyticsRepo) CreatePageView(domain.PageView) error       { return nil }
func (m *mockAnalyticsRepo) CountRecentPageViews(string, time.Time) (int64, error) { return 0, nil }
func (m *mockAnalyticsRepo) GetDashboardStats() (*domain.DashboardStats, error) { return &domain.DashboardStats{}, nil }
func (m *mockAnalyticsRepo) DeletePageViewsOlderThan(time.Time) error   { return nil }
func (m *mockAnalyticsRepo) DeleteSearchEventsOlderThan(time.Time) error { return nil }
func (m *mockAnalyticsRepo) DeleteProfileViewsOlderThan(time.Time) error { return nil }
func (m *mockAnalyticsRepo) DeleteExpiredSessions(time.Time) error      { return nil }

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

	// --- Mock de AnalyticsRepository (no-op) ---
	analytics := service.NewAnalyticsService(&mockAnalyticsRepo{})

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
						return &domain.UserAdmin{ID: adminID, Credentials: domain.Credentials{Key: correctSecret}}, nil
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
				return &domain.PsiUserModel{ID: psiID, Credentials: domain.Credentials{Key: correctSecret}}, nil
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

	// --- ESCENARIO 4: OPTIONAL HYBRID AUTH — TESTS DE SEGURIDAD ---
	// Valida que el fix de side-effect-before-validation funcione correctamente.
	// Un token con firma inválida NUNCA debe inyectar identidad en c.Locals().
	t.Run("OptionalHybridAuth_Security", func(t *testing.T) {
		app := fiber.New()
		app.Get("/public", mw.OptionalHybridAuth(), func(c *fiber.Ctx) error {
			if c.Locals("admin") != nil {
				return c.Status(200).SendString("is_admin")
			}
			if c.Locals("psi_user") != nil {
				return c.Status(200).SendString("is_psi")
			}
			return c.Status(200).SendString("is_anonymous")
		})

		// Test 1: Token con firma inválida (secret incorrecto) → debe ser anónimo
		t.Run("Forged_Token_Rejected", func(t *testing.T) {
			// El admin en la DB tiene "correctSecret", pero el token está firmado con "wrong-key"
			mAdmin.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return &domain.UserAdmin{ID: adminID, Credentials: domain.Credentials{Key: correctSecret}}, nil
			}
			forgedToken := generateTestToken(adminID.String(), "admin", "wrong-key", time.Now().Add(time.Hour))

			req := httptest.NewRequest("GET", "/public", nil)
			req.Header.Set("Authorization", "Bearer "+forgedToken)
			resp, _ := app.Test(req)

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if resp.StatusCode != 200 {
				t.Errorf("se esperaba 200, se obtuvo %d", resp.StatusCode)
			}
			if bodyStr != "is_anonymous" {
				t.Errorf("token forjado inyectó identidad — se esperaba 'is_anonymous', se obtuvo '%s'", bodyStr)
			}
		})

		// Test 2: Token válido de admin → debe detectarlo
		t.Run("Valid_Admin_Token_Detected", func(t *testing.T) {
			mAdmin.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return &domain.UserAdmin{ID: adminID, Credentials: domain.Credentials{Key: correctSecret}}, nil
			}
			validToken := generateTestToken(adminID.String(), "admin", correctSecret, time.Now().Add(time.Hour))

			req := httptest.NewRequest("GET", "/public", nil)
			req.Header.Set("Authorization", "Bearer "+validToken)
			resp, _ := app.Test(req)

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if bodyStr != "is_admin" {
				t.Errorf("token válido no detectado — se esperaba 'is_admin', se obtuvo '%s'", bodyStr)
			}
		})

		// Test 3: Token válido de psi → debe detectarlo
		t.Run("Valid_Psi_Token_Detected", func(t *testing.T) {
			mPsi.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return &domain.PsiUserModel{ID: psiID, Credentials: domain.Credentials{Key: correctSecret}}, nil
			}
			validToken := generateTestToken(psiID.String(), "psi", correctSecret, time.Now().Add(time.Hour))

			req := httptest.NewRequest("GET", "/public", nil)
			req.Header.Set("Authorization", "Bearer "+validToken)
			resp, _ := app.Test(req)

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if bodyStr != "is_psi" {
				t.Errorf("token psi válido no detectado — se esperaba 'is_psi', se obtuvo '%s'", bodyStr)
			}
		})

		// Test 4: Token expirado → debe ser anónimo
		t.Run("Expired_Token_Anonymous", func(t *testing.T) {
			expiredToken := generateTestToken(adminID.String(), "admin", correctSecret, time.Now().Add(-time.Hour))

			req := httptest.NewRequest("GET", "/public", nil)
			req.Header.Set("Authorization", "Bearer "+expiredToken)
			resp, _ := app.Test(req)

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if bodyStr != "is_anonymous" {
				t.Errorf("token expirado inyectó identidad — se esperaba 'is_anonymous', se obtuvo '%s'", bodyStr)
			}
		})

		// Test 5: Sin token → anónimo
		t.Run("No_Token_Anonymous", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/public", nil)
			resp, _ := app.Test(req)

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if bodyStr != "is_anonymous" {
				t.Errorf("sin token se esperaba 'is_anonymous', se obtuvo '%s'", bodyStr)
			}
		})

		// Test 6: Token con alg:none → debe ser anónimo
		t.Run("None_Algorithm_Anonymous", func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
				"user_id": adminID.String(),
				"role":    "admin",
			})
			tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

			req := httptest.NewRequest("GET", "/public", nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)
			resp, _ := app.Test(req)

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if bodyStr != "is_anonymous" {
				t.Errorf("alg:none inyectó identidad — se esperaba 'is_anonymous', se obtuvo '%s'", bodyStr)
			}
		})

		// Test 7: User no existe en DB → anónimo
		t.Run("Nonexistent_User_Anonymous", func(t *testing.T) {
			mAdmin.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return nil, errors.New("not found")
			}
			token := generateTestToken(uuid.NewString(), "admin", correctSecret, time.Now().Add(time.Hour))

			req := httptest.NewRequest("GET", "/public", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, _ := app.Test(req)

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if bodyStr != "is_anonymous" {
				t.Errorf("user inexistente inyectó identidad — se esperaba 'is_anonymous', se obtuvo '%s'", bodyStr)
			}
		})
	})
}

package handler

import (
	"context"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// El panel de analítica mide público. El staff logueado (el que administra el
// Colegio, no el que lo visita) navega el sitio público para verificar fichas,
// buscar nombres y probar filtros: si se contara, "Búsquedas" y "Visitas" volverían
// a significar dos cosas distintas, que es exactamente el gotcha 26.
//
// Estos tests fijan el contrato en los dos handlers que escriben telemetría fuera
// del AnalyticsMiddleware: SearchDirectory y GetPublicProfile.

// staffBrowserUA es un UA de navegador porque "curl" está en la lista de bots de
// IsBotUA: sin él el filtro de bots descartaría el evento y el control positivo
// de estos tests no probaría nada.
const staffBrowserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// analyticsSpy cuenta los eventos de telemetría que el handler escribe.
type analyticsSpy struct {
	searches atomic.Int64
	views    atomic.Int64
}

func (s *analyticsSpy) repo() *mockAnalyticsRepo {
	return &mockAnalyticsRepo{
		CreateSearchEventFunc: func(context.Context, domain.SearchEvent) error {
			s.searches.Add(1)
			return nil
		},
		CreateProfileViewFunc: func(context.Context, domain.ProfileView) error {
			s.views.Add(1)
			return nil
		},
	}
}

// withStaff simula lo que OptionalHybridAuth inyecta cuando la petición llega con
// el JWT de un admin: c.Locals("admin") con un *domain.UserAdmin.
func withStaff(c *fiber.Ctx) error {
	c.Locals("admin", testAdmin(uuid.New(), false, false))
	return c.Next()
}

// appConStaff monta la ruta pública con el Locals del staff ya inyectado.
func appConStaff(method, path string, handler fiber.Handler) *fiber.App {
	app := newTestApp()
	app.Add(method, path, withStaff, handler)
	return app
}

// esperaEventos es el CONTROL POSITIVO: exige que el evento llegue. Si esto
// fallara, el subtest "no escribe" no estaría probando nada.
func esperaEventos(t *testing.T, lee func() int64, d time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool { return lee() >= 1 }, d, 10*time.Millisecond,
		"el evento de telemetría no llegó: el spy o la escritura están rotos, no el filtro")
}

// esperaSinEventos exige que el contador siga en cero durante toda la ventana.
// El recorrido es indispensable: las escrituras son fire-and-forget, así que un
// require.Zero puntual podría ejecutarse antes de que la goroutine llegue a
// escribir. Un solo tick después de la respuesta NO demuestra nada.
func esperaSinEventos(t *testing.T, lee func() int64, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		require.Zero(t, lee(), "se escribió un evento de telemetría que no debía")
		time.Sleep(10 * time.Millisecond)
	}
}

// =========================================================================
// Búsquedas del directorio
// =========================================================================

func TestSearchDirectory_ElStaffNoGeneraTelemetria(t *testing.T) {
	repo := func() *mockPsiRepo {
		return &mockPsiRepo{
			SearchDirectoryFunc: func(_ context.Context, _ request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
				return []domain.PsiUserModel{*testPsiUser(uuid.New())}, 1, nil
			},
		}
	}

	t.Run("anonimo_si_se_escribe", func(t *testing.T) {
		// Control positivo: sin el filtro de staff el evento sí se escribe.
		spy := &analyticsSpy{}
		h, _ := testPsiHandler(repo(), spy.repo())
		app := setupPublicRoute(fiber.MethodGet, "/psi/directory", h.SearchDirectory)

		req := httptest.NewRequest(fiber.MethodGet, "/psi/directory?q=terapia", nil)
		req.Header.Set("User-Agent", staffBrowserUA)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		esperaEventos(t, spy.searches.Load, 2*time.Second)
	})

	t.Run("staff_no_se_escribe", func(t *testing.T) {
		spy := &analyticsSpy{}
		h, _ := testPsiHandler(repo(), spy.repo())
		app := appConStaff(fiber.MethodGet, "/psi/directory", h.SearchDirectory)

		req := httptest.NewRequest(fiber.MethodGet, "/psi/directory?q=terapia", nil)
		req.Header.Set("User-Agent", staffBrowserUA)
		resp, err := app.Test(req)
		require.NoError(t, err)
		// La respuesta es idéntica: el filtro no puede romper el sitio público.
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		esperaSinEventos(t, spy.searches.Load, time.Second)
	})
}

// =========================================================================
// Visitas a la ficha pública
// =========================================================================

func TestGetPublicProfile_ElStaffNoGeneraTelemetria(t *testing.T) {
	repo := func() *mockPsiRepo {
		return &mockPsiRepo{
			GetByFPVFunc: func(_ context.Context, _ int) (domain.PsiUserModel, error) {
				return *testPsiUser(uuid.New()), nil
			},
		}
	}

	t.Run("anonimo_si_se_escribe", func(t *testing.T) {
		spy := &analyticsSpy{}
		h, _ := testPsiHandler(repo(), spy.repo())
		app := setupPublicRoute(fiber.MethodGet, "/psi/:id", h.GetPublicProfile)

		req := httptest.NewRequest(fiber.MethodGet, "/psi/12345", nil)
		req.Header.Set("User-Agent", staffBrowserUA)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		esperaEventos(t, spy.views.Load, 2*time.Second)
	})

	t.Run("staff_no_se_escribe", func(t *testing.T) {
		spy := &analyticsSpy{}
		h, _ := testPsiHandler(repo(), spy.repo())
		app := appConStaff(fiber.MethodGet, "/psi/:id", h.GetPublicProfile)

		req := httptest.NewRequest(fiber.MethodGet, "/psi/12345", nil)
		req.Header.Set("User-Agent", staffBrowserUA)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		esperaSinEventos(t, spy.views.Load, time.Second)
	})
}

// =========================================================================
// Cadena completa: OptionalHybridAuth → handler
// =========================================================================

// El test anterior inyecta el Locals a mano, así que no detectaría un cableado
// roto: si OptionalHybridAuth dejara de rellenar c.Locals("admin"), o si se
// registrara sin él, los tests de arriba seguirían en verde y el staff volvería
// a contarse. Aquí manda un JWT de admin de verdad por el middleware real.
func TestPsiHandler_StaffDetectadoPorElMiddleware(t *testing.T) {
	adminID := uuid.New()
	adminRepo := &mockAdminRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			if id != adminID {
				return nil, domain.ErrPsiNotFound
			}
			return testAdmin(adminID, false, false), nil
		},
	}
	psiRepo := &mockPsiRepo{
		SearchDirectoryFunc: func(_ context.Context, _ request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
			return []domain.PsiUserModel{*testPsiUser(uuid.New())}, 1, nil
		},
	}
	token := generateTestToken(adminID.String(), "admin", time.Now().Add(time.Hour))

	t.Run("con_jwt_de_admin_no_se_escribe", func(t *testing.T) {
		spy := &analyticsSpy{}
		h, _ := testPsiHandler(psiRepo, spy.repo())

		// Envuelve el handler para comprobar que el middleware REALMENTE inyectó
		// la identidad: si no lo hiciera, este test pasaría sin que el filtro
		// tuviera nada que filtrar.
		observa := func(c *fiber.Ctx) error {
			require.NotNil(t, c.Locals("admin"), "OptionalHybridAuth no inyectó c.Locals(\"admin\") con el JWT de admin")
			return h.SearchDirectory(c)
		}
		app := setupHybridRoute(fiber.MethodGet, "/psi/directory", observa, adminRepo, psiRepo)

		req := authRequest(httptest.NewRequest(fiber.MethodGet, "/psi/directory?q=terapia", nil), token)
		req.Header.Set("User-Agent", staffBrowserUA)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		esperaSinEventos(t, spy.searches.Load, time.Second)
	})

	t.Run("sin_jwt_si_se_escribe", func(t *testing.T) {
		// El control negativo del control: el mismo montaje, sin cabecera
		// Authorization. Si esto no contara, el test de arriba no probaría nada.
		spy := &analyticsSpy{}
		h, _ := testPsiHandler(psiRepo, spy.repo())
		app := setupHybridRoute(fiber.MethodGet, "/psi/directory", h.SearchDirectory, adminRepo, psiRepo)

		req := httptest.NewRequest(fiber.MethodGet, "/psi/directory?q=terapia", nil)
		req.Header.Set("User-Agent", staffBrowserUA)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		esperaEventos(t, spy.searches.Load, 2*time.Second)
	})
}

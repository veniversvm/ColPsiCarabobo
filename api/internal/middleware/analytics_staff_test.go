package middleware

import (
	"context"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// =========================================================================
// AnalyticsMiddleware: el staff no cuenta como visitante
// =========================================================================

// El paso 3 del AnalyticsMiddleware ("Exclusión de Staff") llevaba tiempo sin un
// test que lo sujetara, y por eso se pudrió en silencio: se puede borrar una
// línea de este middleware y la suite sigue en verde. Estos tests son el candado.
//
// OJO: esta exclusión depende de que la ruta tenga middleware de auth que
// rellene c.Locals("admin") —OptionalHybridAuth en las rutas públicas GET—.
// Si alguien registra una ruta pública sin él, el staff vuelve a contarse y
// solo un test a nivel de router lo detectaría. Por eso el registro por ruta
// lleva su propio comentario.

func TestAnalyticsMiddleware_ElStaffNoRegistraPageView(t *testing.T) {
	// repoDeEspia cuenta las page views escritas.
	repoDeEspia := func(createCount *int64) *mockAnalyticsRepoForMiddleware {
		return &mockAnalyticsRepoForMiddleware{
			CountRecentPageViewsFunc: func(_ context.Context, _ string, _ time.Time) (int64, error) {
				return 0, nil
			},
			CreatePageViewFunc: func(_ context.Context, _ domain.PageView) error {
				atomic.AddInt64(createCount, 1)
				return nil
			},
		}
	}

	// inyectaAdmin simula lo que OptionalHybridAuth deja en c.Locals cuando la
	// petición llega con el JWT de un admin.
	inyectaAdmin := func(c *fiber.Ctx) error {
		c.Locals("admin", &domain.UserAdmin{})
		return c.Next()
	}

	const ruta = "/api/v1/psi/directory"
	okHandler := func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) }

	t.Run("control_anonimo_si_registra", func(t *testing.T) {
		// Control positivo: sin esto, el subtest de abajo no probaría nada.
		var count int64
		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		app.Get(ruta, AnalyticsMiddleware(service.NewAnalyticsService(repoDeEspia(&count))), okHandler)

		req := httptest.NewRequest("GET", ruta, nil)
		req.Header.Set("User-Agent", testBrowserUA)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.Eventually(t, func() bool {
			return atomic.LoadInt64(&count) == 1
		}, 2*time.Second, 10*time.Millisecond, "un visitante anónimo debe registrarse")
	})

	t.Run("staff_no_registra", func(t *testing.T) {
		var count int64
		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		app.Get(ruta, inyectaAdmin, AnalyticsMiddleware(service.NewAnalyticsService(repoDeEspia(&count))), okHandler)

		req := httptest.NewRequest("GET", ruta, nil)
		req.Header.Set("User-Agent", testBrowserUA)
		resp, err := app.Test(req)
		require.NoError(t, err)
		// El staff debe seguir viendo el directorio con normalidad: el filtro es
		// de métricas, jamás de acceso.
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		time.Sleep(300 * time.Millisecond)
		require.Zero(t, atomic.LoadInt64(&count), "el staff no debe registrarse como visitante")
		require.Empty(t, resp.Header.Values("Set-Cookie"), "al staff no debe ponérsele la cookie _sid: tampoco es un visitante")
	})
}

package router

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
)

// =========================================================================
// Guardián del quirk de Fiber v2 en las rutas públicas de /psi
// =========================================================================

// Las GET públicas que alimentan la telemetría llevan OptionalHybridAuth POR
// RUTA (psi_router.go, specialty_router.go) para que el staff no cuente como
// visitante, y con él una petición sin sesión debe seguir entrando con normalidad:
// el filtro es de MÉTRICAS, jamás de acceso.
//
// Este test fija ese invariante, que es la regresión que de verdad importaría
// (un staff con un token caducado, o una futura middleware mal colocada, dejando
// el sitio público en 401). Solo manda peticiones anónimas: con cabecera
// Authorization, OptionalHybridAuth tocaría los repositorios (nil aquí, a
// propósito). La cadena completa con un JWT de admin real la cubre
// handler.TestPsiHandler_StaffDetectadoPorElMiddleware.
//
// ⚠️ NO lo justifiques por el "quirk de Fiber v2" del gotcha 16. Se comprobó
// empíricamente (v2.52.11) que registrar `meGroup` = /psi/me con
// ProtectedPsiUser ANTES que `psiGroup` = /psi NO filtra ese middleware a
// /psi/directory: los prefijos se aíslan. El registro por ruta es una decisión
// de legibilidad y de no heredar la resolución de identidad en silencio, no un
// rodeo. El gotcha 16 sí es real para un 2º Group sobre el MISMO prefijo y para
// un registro directo bajo un prefijo ya usado; para eso sí hay que usar
// prefijos distintos.

func TestPublicasPsi_SiguenAbiertasSinToken(t *testing.T) {
	mw := middleware.NewAuthMiddleware(nil, nil, nil)
	okHandler := func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) }

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Orden como en psi_router.go: el grupo que bloquea va primero.
	meGroup := app.Group("/psi/me", middleware.NoStore(), bloqueaSinToken())
	meGroup.Get("/", okHandler)

	psiGroup := app.Group("/psi")
	psiGroup.Get("/directory", mw.OptionalHybridAuth(), okHandler)
	psiGroup.Get("/:id", mw.OptionalHybridAuth(), okHandler)
	psiGroup.Get("/public/sitemap-data", mw.OptionalHybridAuth(), okHandler)

	specialties := app.Group("/specialties")
	specialties.Get("/count", mw.OptionalHybridAuth(), okHandler)
	specialties.Get("/", mw.OptionalHybridAuth(), okHandler)
	specialties.Get("/:id<int>", mw.OptionalHybridAuth(), okHandler)

	for _, ruta := range []string{
		"/psi/directory",
		"/psi/12345",
		"/psi/public/sitemap-data",
		"/specialties",
		"/specialties/count",
		"/specialties/7",
	} {
		t.Run(ruta, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest("GET", ruta, nil))
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode,
				"una ruta pública sin sesión debe seguir abierta: el filtro de staff es de métricas, nunca de acceso")
		})
	}

	t.Run("autogestion_psi_sigue_bloqueada", func(t *testing.T) {
		// La comprobación complementaria: /psi/me NO debe abrirse por el hecho de
		// que ahora haya un OptionalHybridAuth en el grupo hermano.
		resp, err := app.Test(httptest.NewRequest("GET", "/psi/me", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode,
			"/psi/me es privada de agremiados: no puede quedar accesible")
	})
}

// bloqueaSinToken imita un middleware de bloqueo sin tocar la base, para que el
// test pueda reproducir el apilamiento del stack sin montar la BD.
func bloqueaSinToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("Authorization") == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Next()
	}
}

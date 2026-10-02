package middleware

import (
	"context"
	"io"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// cuerpo lee el body de la respuesta como string.
func cuerpo(respBody io.ReadCloser) string {
	defer respBody.Close()
	b, _ := io.ReadAll(respBody)
	return string(b)
}

// =========================================================================
// OptionalHybridAuth: una fila por petición, no dos
// =========================================================================

// OptionalHybridAuth se registró en las rutas públicas GET (el directorio, la
// ficha, /specialties) para que la telemetría pueda excluir al staff, así que su
// coste pasó de "solo /posts" a "todo el sitio público".
//
// El recorrido anónimo no cuesta nada —cortocircuita en la línea de la cabecera
// Authorization—, pero el del staff sí: la keyFunc de jwt.Parse lee la fila para
// obtener la key de firma y el PASO 3 la volvía a leer para inyectar identidad.
// Dos SELECT idénticos del MISMO uid en la MISMA petición, para el mismo dato.
//
// Este test fija el contrato de "una lectura por petición". Es también el que
// documenta que el anónimo paga cero, que es lo que hace aceptable el cambio.

// contaLecturas mide cuántas veces se piden las filas, por rol.
type contaLecturas struct {
	admins atomic.Int64
	psis   atomic.Int64
}

func TestOptionalHybridAuth_LeeLaFilaUnaSolaVez(t *testing.T) {
	const secret = "clave-de-firma-del-admin"

	nuevoApp := func(c *contaLecturas, idAdmin, idPsi uuid.UUID) *fiber.App {
		mAdmin := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				c.admins.Add(1)
				if id != idAdmin {
					return nil, domain.ErrPsiNotFound
				}
				return &domain.UserAdmin{
					ID:          idAdmin,
					Credentials: domain.Credentials{Key: secret},
				}, nil
			},
		}
		mPsi := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				c.psis.Add(1)
				if id != idPsi {
					return nil, domain.ErrPsiNotFound
				}
				return &domain.PsiUserModel{
					ID:          idPsi,
					Credentials: domain.Credentials{Key: secret},
				}, nil
			},
		}

		mw := NewAuthMiddleware(mAdmin, mPsi, service.NewAnalyticsService(&mockAnalyticsRepo{}))
		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		app.Get("/publico", mw.OptionalHybridAuth(), func(c *fiber.Ctx) error {
			// Devuelve qué identidad quedó inyectada, para poder exigirlo.
			switch {
			case c.Locals("admin") != nil:
				return c.SendString("admin")
			case c.Locals("psi_user") != nil:
				return c.SendString("psi")
			default:
				return c.SendString("anonimo")
			}
		})
		return app
	}

	t.Run("admin_una_sola_lectura", func(t *testing.T) {
		c := &contaLecturas{}
		idAdmin, idPsi := uuid.New(), uuid.New()
		app := nuevoApp(c, idAdmin, idPsi)

		req := httptest.NewRequest("GET", "/publico", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(idAdmin.String(), "admin", secret, time.Now().Add(time.Hour)))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.Equal(t, 1, int(c.admins.Load()), "el admin se leyó más de una vez por petición")
		require.Equal(t, 0, int(c.psis.Load()), "no debía tocarse la tabla de psicólogos")
		body := cuerpo(resp.Body)
		require.Equal(t, "admin", body, "la identidad debe inyectarse aunque se reutilice la fila de la keyFunc")
	})

	t.Run("psi_una_sola_lectura", func(t *testing.T) {
		c := &contaLecturas{}
		idAdmin, idPsi := uuid.New(), uuid.New()
		app := nuevoApp(c, idAdmin, idPsi)

		req := httptest.NewRequest("GET", "/publico", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(idPsi.String(), "psi", secret, time.Now().Add(time.Hour)))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.Equal(t, 1, int(c.psis.Load()), "el agremiado se leyó más de una vez por petición")
		require.Equal(t, 0, int(c.admins.Load()))
		body := cuerpo(resp.Body)
		require.Equal(t, "psi", body)
	})

	t.Run("anonimo_no_consulta_la_base", func(t *testing.T) {
		// Es lo que hace aceptable poner este middleware en las rutas públicas:
		// el visitante sin sesión no paga ni un SELECT.
		c := &contaLecturas{}
		app := nuevoApp(c, uuid.New(), uuid.New())

		req := httptest.NewRequest("GET", "/publico", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.Zero(t, c.admins.Load(), "un visitante anónimo no debe disparar ninguna consulta")
		require.Zero(t, c.psis.Load(), "un visitante anónimo no debe disparar ninguna consulta")
		body := cuerpo(resp.Body)
		require.Equal(t, "anonimo", body)
	})

	t.Run("firma_invalida_no_inyecta_identidad", func(t *testing.T) {
		// El memo solo se reutiliza cuando el token ya pasó la verificación:
		// un token con firma inválida sigue siendo anónimo (gotcha de
		// side-effect-before-validation).
		c := &contaLecturas{}
		idAdmin, idPsi := uuid.New(), uuid.New()
		app := nuevoApp(c, idAdmin, idPsi)

		req := httptest.NewRequest("GET", "/publico", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(idAdmin.String(), "admin", "clave-que-no-es", time.Now().Add(time.Hour)))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode, "el middleware es opcional: nunca bloquea")

		body := cuerpo(resp.Body)
		require.Equal(t, "anonimo", body, "una firma inválida NUNCA debe inyectar identidad")
	})
}

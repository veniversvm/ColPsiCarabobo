// api/internal/middleware/analytics.go

// Package middleware contiene interceptores HTTP que gestionan preocupaciones transversales
// (Cross-Cutting Concerns) como seguridad, telemetría y autenticación, evaluándose
// antes o después de que la petición alcance la capa de controladores.
package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// visitWindow define el umbral de "Debouncing" para el rastreo de visitas.
// Si un mismo identificador de sesión (_sid) navega por múltiples rutas o recarga
// la página dentro de esta ventana de 30 minutos, se considera una única visita continua.
// Esto previene el inflado artificial de métricas y el agotamiento de la base de datos.
const visitWindow = 30 * time.Minute

// skipPaths define una Lista Negra (Blocklist) de prefijos de ruta que no deben
// generar eventos analíticos, evitando que el "ruido" contamine las métricas de negocio.
var skipPaths = []string{
	"/health", "/favicon.ico", "/static/", "/assets/",
	"/_build/", "/metrics",
}

// shouldSkip evalúa rápidamente si la ruta actual coincide con la lista negra.
func shouldSkip(path string) bool {
	for _, s := range skipPaths {
		if strings.HasPrefix(path, s) {
			return true
		}
	}
	return false
}

// AnalyticsMiddleware rastrea la actividad de los usuarios de forma no intrusiva.
//
// Diseño de Rendimiento:
// Funciona de manera Asíncrona (Post-Procesamiento). En lugar de bloquear la
// respuesta HTTP esperando a que la base de datos registre la visita, ejecuta
// un "Fire-and-Forget" mediante Goroutines, garantizando una latencia de red de 0ms
// de impacto para el cliente final.
func AnalyticsMiddleware(analytics *service.AnalyticsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ejecutar la petición (c.Next) hacia el controlador.
		err := c.Next()

		// 2. Filtro de Calidad de Datos (Data Quality)
		if c.Method() != "GET" || shouldSkip(c.Path()) {
			return err
		}
		if c.Response().StatusCode() < 200 || c.Response().StatusCode() >= 300 {
			return err
		}

		// 3. Exclusión de Staff (Métricas Limpias)
		if admin, ok := c.Locals("admin").(*domain.UserAdmin); ok && admin != nil {
			return err
		}

		// 4. Identificación de Entidad
		var userID *uuid.UUID
		if uid, ok := c.Locals("userID").(uuid.UUID); ok {
			userID = &uid
		}

		// 5. Gestión de Sesión Anónima (Tracking Cookie)
		sessionID := c.Cookies("_sid")
		if sessionID == "" {
			sessionID = uuid.Must(uuid.NewV7()).String()
			c.Cookie(&fiber.Cookie{
				Name:     "_sid",
				Value:    sessionID,
				Expires:  time.Now().Add(365 * 24 * time.Hour),
				HTTPOnly: true,
				Secure:   config.Envs.Environment == "production",
				SameSite: "Lax",
			})
		}

		// ── SEGURIDAD DE CONTEXTO (Fiber Lifecycle) ──────────────────────────
		path := c.Path()
		method := c.Method()
		ip := c.IP()
		referer := c.Get("Referer")
		// ─────────────────────────────────────────────────────────────────────

		// 6. Volcado Asíncrono a Base de Datos
		go func() {
			// Ventana de sesión (Debouncing):
			count, _ := analytics.CountRecentPageViews(sessionID, time.Now().Add(-visitWindow))
			if count > 0 {
				return // Ya registrado en esta ventana
			}

			// Si es una visita fresca, se persiste en el log de analíticas.
			analytics.RecordPageView(domain.PageView{
				Path:      path,
				Method:    method,
				UserID:    userID,
				SessionID: sessionID,
				IP:        ip,
				Referer:   referer,
				CreatedAt: time.Now(),
			})
		}()

		return err
	}
}

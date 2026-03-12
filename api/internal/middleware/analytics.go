// api/internal/middleware/analytics.go

package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// Ventana de sesión: si el mismo _sid tiene un PageView en los últimos 30 minutos,
// no se registra una nueva visita. Navegar entre subrutas o refrescar no cuenta.
const visitWindow = 30 * time.Minute

var skipPaths = []string{
	"/health", "/favicon.ico", "/static/", "/assets/",
	"/_build/", "/metrics",
}

func shouldSkip(path string) bool {
	for _, s := range skipPaths {
		if strings.HasPrefix(path, s) {
			return true
		}
	}
	return false
}

// AnalyticsMiddleware registra una visita por sesión cada 30 minutos (no por page view).
func AnalyticsMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		// Solo GET exitosos (2xx) fuera de rutas de sistema
		if c.Method() != "GET" || shouldSkip(c.Path()) {
			return err
		}
		if c.Response().StatusCode() < 200 || c.Response().StatusCode() >= 300 {
			return err
		}

		// Extraer userID si está autenticado
		var userID *uuid.UUID
		if uid, ok := c.Locals("userID").(uuid.UUID); ok {
			userID = &uid
		}

		// Sesión anónima via cookie (persiste 1 año)
		sessionID := c.Cookies("_sid")
		if sessionID == "" {
			sessionID = uuid.NewString()
			c.Cookie(&fiber.Cookie{
				Name:     "_sid",
				Value:    sessionID,
				Expires:  time.Now().Add(365 * 24 * time.Hour),
				HTTPOnly: true,
				SameSite: "Lax",
			})
		}

		if admin, ok := c.Locals("admin").(*domain.UserAdmin); ok && admin != nil {
			return err
		}

		// Capturar valores ANTES de la goroutine — Fiber libera el contexto
		// al terminar el handler y accederlo desde una goroutine da nil pointer
		path := c.Path()
		method := c.Method()
		ip := c.IP()
		referer := c.Get("Referer")

		go func() {
			// ── Ventana de sesión ────────────────────────────────────────────
			// Si este _sid ya tiene un PageView en los últimos 30 minutos,
			// el usuario solo está navegando — no es una visita nueva.
			var count int64
			db.Model(&domain.PageView{}).
				Where("session_id = ? AND created_at >= ?", sessionID, time.Now().Add(-visitWindow)).
				Count(&count)

			if count > 0 {
				return // Ya registrado en esta ventana, no insertar
			}
			// ────────────────────────────────────────────────────────────────

			db.Create(&domain.PageView{
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

// api/internal/middleware/security_headers.go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

// SecurityHeaders configura helmet con valores consistentes para todo el stack
// (main.go y app de tests). HSTS solo se emite sobre HTTPS; Permissions-Policy
// bloquea APIs de hardware/locación que la API no necesita.
func SecurityHeaders() fiber.Handler {
	return helmet.New(helmet.Config{
		HSTSMaxAge:         config.Envs.HSTSMaxAge,
		HSTSPreloadEnabled: config.Envs.HSTSPreloadEnabled,
		PermissionPolicy:   "geolocation=(), microphone=(), camera=(), usb=(), magnetometer=(), accelerometer=(), gyroscope=()",
	})
}

// NoStore evita que proxies/caches guarden respuestas de endpoints sensibles
// (login, admin, datos propios del psicólogo).
func NoStore() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "no-store")
		return c.Next()
	}
}

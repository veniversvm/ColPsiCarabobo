// api/internal/middleware/rate_limiter.go
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// AuthRateLimiter aplica un límite estricto para endpoints de autenticación.
// 10 intentos por IP cada 15 minutos — bloquea fuerza bruta sin afectar uso legítimo.
func AuthRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,
		Expiration: 15 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Usa el IP real aunque haya un proxy/reverse proxy delante
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Demasiados intentos de acceso.",
				"message": "Por seguridad, tu acceso ha sido bloqueado temporalmente. Intenta de nuevo en 15 minutos.",
			})
		},
		// No contar rutas que no sean POST para no gastar cuota en GETs accidentales
		Next: func(c *fiber.Ctx) bool {
			return c.Method() != fiber.MethodPost
		},
	})
}

// AdminAuthRateLimiter es más agresivo para el login de administradores.
// 5 intentos por IP cada 30 minutos.
func AdminAuthRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 30 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Acceso administrativo bloqueado.",
				"message": "Demasiados intentos fallidos. Intenta de nuevo en 30 minutos.",
			})
		},
		Next: func(c *fiber.Ctx) bool {
			return c.Method() != fiber.MethodPost
		},
	})
}

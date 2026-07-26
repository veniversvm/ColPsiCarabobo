// api/internal/middleware/rate_limiter.go

// Package middleware aloja los interceptores de seguridad de la API.
//
// CONCEPTO DE RATE LIMITING (Control de Tráfico):
// Este archivo implementa el patrón de estrangulamiento (Throttling). Funciona como la
// primera barrera defensiva en la capa HTTP para proteger las rutas críticas (Logins)
// contra ataques de Fuerza Bruta (adivinar contraseñas) y Credential Stuffing (uso
// automatizado de credenciales filtradas en otras brechas de seguridad).
//
// Almacenamiento: Si VALKEY_ADDR está configurado, los contadores persisten en Valkey
// (compatible con Redis). Si no, se usa almacenamiento in-memory (Fiber default).
// Valkey soporta despliegue multi-instancia (Docker replicas, K8s).
package middleware

import (
	"github.com/rs/zerolog/log"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage/valkey"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

// rateLimiterStore es una instancia compartida de storage para rate limiting.
// Se inicializa una vez al arrancar la aplicación.
var (
	rateLimiterStore fiber.Storage
	once             sync.Once
)

// newRateLimiterStorage crea o retorna el storage para rate limiting.
// Si VALKEY_ADDR está configurado, usa Valkey (persistente, multi-instancia).
// Si no, retorna nil para usar el default in-memory de Fiber.
func newRateLimiterStorage() fiber.Storage {
	once.Do(func() {
		if config.Envs == nil || config.Envs.ValkeyAddr == "" {
			log.Info().Str("component", "rate-limit").Msg("Modo: in-memory (sin VALKEY_ADDR configurado)")
			return
		}

		// Intentar conectar a Valkey con panic recovery
		// gofiber/storage/valkey panics si no puede conectar
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Warn().Err(fmt.Errorf("%v", r)).Str("component", "rate-limit").Msg("No se pudo conectar a Valkey. Usando in-memory.")
				}
			}()

			store := valkey.New(valkey.Config{
				InitAddress: []string{config.Envs.ValkeyAddr},
			})
			rateLimiterStore = store
			log.Info().Str("component", "rate-limit").Str("addr", config.Envs.ValkeyAddr).Msg("Modo: Valkey — persistente, multi-instancia")
		}()
	})
	return rateLimiterStore
}

// AuthRateLimiter aplica un escudo de protección estándar para los endpoints
// de autenticación del usuario general (psicólogos).
//
// Umbral de Tolerancia:
// Permite 10 intentos por IP en una ventana temporal de 15 minutos. Este límite
// fue diseñado para ser lo suficientemente holgado para un usuario humano legítimo
// que olvidó su contraseña, pero matemáticamente inviable para un script que
// intente romper la seguridad por fuerza bruta.
func AuthRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,
		Expiration: 15 * time.Minute,
		Storage:    newRateLimiterStorage(),

		// KeyGenerator define cómo se identifica a un "actor" único.
		// Arquitectura de Red: c.IP() de Fiber resuelve automáticamente cabeceras como
		// 'X-Forwarded-For' o 'X-Real-IP'. Esto garantiza que si la API está detrás de un
		// Reverse Proxy (como Nginx o Cloudflare), el limitador castigará la IP real
		// del atacante y no la IP interna del proxy.
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},

		// LimitReached estandariza la respuesta de error HTTP 429 (Too Many Requests).
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Demasiados intentos de acceso.",
				"message": "Por seguridad, tu acceso ha sido bloqueado temporalmente. Intenta de nuevo en 15 minutos.",
			})
		},

		// Next funciona como un filtro de excepción (Bypass).
		// Optimización y UX: Evita descontar cuota del limitador si la petición NO es POST.
		// Esto es vital en arquitecturas web modernas, ya que los navegadores envían
		// peticiones 'OPTIONS' (CORS Preflight) antes del POST. Si no se ignoran,
		// un usuario consumiría 2 intentos por cada click en "Ingresar".
		Next: func(c *fiber.Ctx) bool {
			return c.Method() != fiber.MethodPost
		},
	})
}

// AdminAuthRateLimiter aplica una política de "Confianza Cero" (Zero Trust)
// diseñada específicamente para la superficie de ataque del Staff.
//
// Asimetría de Seguridad:
// Las cuentas administrativas son "High-Value Targets" (Objetivos de Alto Valor).
// Por ello, la política es deliberadamente el doble de agresiva que la del usuario
// general: restringe los intentos a la mitad (5) y duplica el tiempo de castigo (30 minutos),
// mitigando vectores de ataque dirigidos directamente al corazón del sistema.
func AdminAuthRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 30 * time.Minute,
		Storage:    newRateLimiterStorage(),
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

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
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
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
func AnalyticsMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ejecutar la petición (c.Next) hacia el controlador.
		// Esto nos permite capturar el código de estado (StatusCode) real de la respuesta.
		err := c.Next()

		// 2. Filtro de Calidad de Datos (Data Quality)
		// Solo rastreamos peticiones GET (lectura de páginas) que hayan sido
		// exitosas (HTTP 2xx) y que no pertenezcan a rutas de sistema (estáticos, healthchecks).
		if c.Method() != "GET" || shouldSkip(c.Path()) {
			return err
		}
		if c.Response().StatusCode() < 200 || c.Response().StatusCode() >= 300 {
			return err
		}

		// 3. Exclusión de Staff (Métricas Limpias)
		// Ignorar el tráfico generado por el propio equipo administrativo para no sesgar las métricas.
		if admin, ok := c.Locals("admin").(*domain.UserAdmin); ok && admin != nil {
			return err
		}

		// 4. Identificación de Entidad
		// Extracción del ID si es un psicólogo autenticado.
		var userID *uuid.UUID
		if uid, ok := c.Locals("userID").(uuid.UUID); ok {
			userID = &uid
		}

		// 5. Gestión de Sesión Anónima (Tracking Cookie)
		// Si el usuario no tiene la cookie de sesión (_sid), se le asigna una.
		sessionID := c.Cookies("_sid")
		if sessionID == "" {
			sessionID = uuid.NewString()
			c.Cookie(&fiber.Cookie{
				Name:     "_sid",
				Value:    sessionID,
				Expires:  time.Now().Add(365 * 24 * time.Hour), // Persistencia de 1 año
				HTTPOnly: true,                                 // Previene robo por XSS (Javascript no puede leerla)
				SameSite: "Lax",                                // Seguridad CSRF básica
			})
		}

		// ── SEGURIDAD DE CONTEXTO (Fiber Lifecycle) ──────────────────────────
		// FastHTTP (motor base de Fiber) recicla activamente el objeto *fiber.Ctx
		// tan pronto como la función actual retorna al cliente, para ahorrar memoria.
		// Si la Goroutine intenta acceder a c.Path() o c.IP() más tarde, la memoria
		// ya habrá sido sobreescrita, provocando un Panic o inyectando basura en la DB.
		// Solución: Copiar los valores primitivos a memoria segura ANTES del go func().
		path := c.Path()
		method := c.Method()
		ip := c.IP()
		referer := c.Get("Referer")
		// ─────────────────────────────────────────────────────────────────────

		// 6. Volcado Asíncrono a Base de Datos
		go func() {
			// Ventana de sesión (Debouncing):
			// Consultamos si este dispositivo ya registró actividad reciente.
			var count int64
			db.Model(&domain.PageView{}).
				Where("session_id = ? AND created_at >= ?", sessionID, time.Now().Add(-visitWindow)).
				Count(&count)

			if count > 0 {
				return // Ya registrado en esta ventana (solo está navegando)
			}

			// Si es una visita fresca, se persiste en el log de analíticas.
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

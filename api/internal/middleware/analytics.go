// api/internal/middleware/analytics.go

// Package middleware contiene interceptores HTTP que gestionan preocupaciones transversales
// (Cross-Cutting Concerns) como seguridad, telemetría y autenticación, evaluándose
// antes o después de que la petición alcance la capa de controladores.
package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// analyticsCtxTimeout acota la vida de la goroutine analítica por request.
// Si la BD está degradada, la goroutine muere sola en vez de colgarse esperando
// una conexión del pool.
const analyticsCtxTimeout = 5 * time.Second

// analyticsSessionCookieDays es la vida de la cookie técnica _sid, que existe
// únicamente para no contar dos veces la misma visita dentro de la ventana de
// deduplicación. Antes duraba 365 días, que es una ventana de seguimiento
// desproporcionada para lo que hace; 30 días mantiene el conteo de visitantes
// únicos igual de útil.
const analyticsSessionCookieDays = 30

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

// botUserAgents son substrings de User-Agent de crawlers/scripts conocidos.
// Se usan SOLO para excluir su ruido de las métricas; no bloquean la respuesta
// HTTP (Googlebot y compañía deben seguir indexando el sitio).
var botUserAgents = []string{
	"googlebot", "bingbot", "duckduckbot", "yandex", "baiduspider",
	"slurp", "msnbot", "semrushbot", "ahrefsbot", "mj12bot", "applebot",
	"twitterbot", "facebookexternalhit", "linkedinbot", "pinterest",
	"telegrambot", "whatsapp", "discordbot", "uptimerobot", "pingdom",
	"archive.org_bot", "gptbot", "ccbot", "bytespider", "perplexitybot",
	"claudebot", "petalbot", "dotbot", "curl", "wget", "python-requests",
}

// isBotUA detecta si un User-Agent pertenece a un bot/crawler/script.
// Un UA vacío o ausente también se considera sospechoso (scripts sin cabecera).
func isBotUA(userAgent string) bool {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" {
		return true
	}
	for _, b := range botUserAgents {
		if strings.Contains(ua, b) {
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
		if isBotUA(c.Get("User-Agent")) {
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
		// Solo sirve para no contar dos veces a la misma persona: la deduplicación
		// de TrackPageView corre en analyticsVisitWindow (30 minutos), así que 30
		// días de vida sobra con creces y es lo que se declara en los Términos.
		sessionID := c.Cookies("_sid")
		if sessionID == "" {
			sessionID = uuid.Must(uuid.NewV7()).String()
			c.Cookie(&fiber.Cookie{
				Name:     "_sid",
				Value:    sessionID,
				Expires:  time.Now().Add(analyticsSessionCookieDays * 24 * time.Hour),
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
		// La goroutine es la única frontera asíncrona: TrackPageView corre de forma
		// síncrona dentro (con su propio debouncing) bajo el semáforo del servicio.
		//
		// OJO: aquí se entrega la DIRECCIÓN en crudo, aunque el campo se llame
		// IPHash. La conversión a huella la hace el servicio a propósito (es el
		// que garantiza que a la base nunca llegue una IP); si te la haces aquí,
		// otro flujo futuro vuelve a escribirla en claro.
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), analyticsCtxTimeout)
			defer cancel()

			analytics.TrackPageView(ctx, domain.PageView{
				Path:      path,
				Method:    method,
				UserID:    userID,
				SessionID: sessionID,
				IPHash:    ip,
				Referer:   referer,
				CreatedAt: time.Now(),
			})
		}()

		return err
	}
}

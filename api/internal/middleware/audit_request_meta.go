// api/internal/middleware/audit_request_meta.go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// AuditRequestMeta inyecta la IP y el User-Agent de la petición en su
// context.Context (service.WithAuditMeta) para que los servicios puedan
// etiquetar los eventos de auditoría sin recibir el fiber.Ctx.
//
// Es una operación puramente aditiva: no bloquea, no transforma la respuesta
// y no interviene en la autorización. Se aplica globalmente en SetupRouter,
// antes del registro de rutas.
func AuditRequestMeta() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := service.WithAuditMeta(c.UserContext(), c.IP(), c.Get("User-Agent"))
		c.SetUserContext(ctx)
		return c.Next()
	}
}

// api/internal/handler/manual_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/manuales"
)

// ManualHandler sirve los manuales PDF del proyecto. Los bytes provienen del
// paquete `manuales` (embebidos en el binario); el acceso lo garantiza el
// middleware ProtectedAnyAuth (sesión admin O psicólogo) del router.
type ManualHandler struct{}

// NewManualHandler construye el handler de manuales.
func NewManualHandler() *ManualHandler {
	return &ManualHandler{}
}

// GetManual devuelve el PDF del manual solicitado.
// @Summary      Obtener manual (PDF)
// @Description  Devuelve el manual solicitado en PDF. Requiere sesión válida de administrador o de psicólogo. Los archivos están embebidos en el binario y nunca se sirven a visitantes anónimos.
// @Tags         Manuales
// @Security     BearerAuth
// @Param        file path string true "Nombre del manual (manual-admin.pdf | manual-psiuser.pdf)"
// @Success      200 {file} binary "PDF del manual"
// @Failure      401 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Router       /manuales/{file} [get]
func (h *ManualHandler) GetManual(c *fiber.Ctx) error {
	name := c.Params("file")
	data, ok := manuales.Get(name)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Manual no encontrado.",
		})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Cache-Control", "no-store")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("Content-Disposition", `inline; filename="`+name+`"`)
	return c.Send(data)
}
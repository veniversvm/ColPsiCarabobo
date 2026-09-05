// api/internal/handler/settings_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// SettingsHandler expone el KV de configuración global al panel administrativo.
type SettingsHandler struct {
	settingsSvc *service.SettingsService
}

// NewSettingsHandler crea el handler de configuración global.
func NewSettingsHandler(settingsSvc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsSvc: settingsSvc}
}

// ReceptionView es el estado completo de los interruptores de recepción.
type ReceptionView = service.ReceptionSwitchesSnapshot

// GetReception godoc
// @Summary      Estado de los interruptores de recepción (admin)
// @Description  Devuelve el estado de recepción de tickets y de inscripciones. Los cambios solo los puede hacer el Sudo.
// @Security     BearerAuth
// @Tags         Admin - Configuración
// @Produce      json
// @Success      200 {object} service.ReceptionSwitchesSnapshot
// @Failure      500 {object} map[string]string "error: fallo interno"
// @Router       /admin/settings/reception [get]
func (h *SettingsHandler) GetReception(c *fiber.Ctx) error {
	if _, err := middleware.GetAuthenticatedAdmin(c); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}
	snapshot, err := h.settingsSvc.GetReceptionSwitches(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al consultar la configuración"})
	}
	return c.JSON(snapshot)
}

// UpdateReception godoc
// @Summary      Actualizar interruptor de recepción (solo Sudo)
// @Description  Activa/desactiva la recepción de tickets o inscripciones con un mensaje público opcional. Requiere rol Sudo.
// @Security     BearerAuth
// @Tags         Admin - Configuración
// @Accept       json
// @Produce      json
// @Param        request body request_structs.UpdateReceptionRequest true "Cambio de recepción"
// @Success      200 {object} service.ReceptionSwitchesSnapshot
// @Failure      400 {object} map[string]string "error: validación"
// @Failure      403 {object} map[string]string "error: solo SUDO"
// @Router       /admin/settings/reception [post]
func (h *SettingsHandler) UpdateReception(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}
	if !admin.Sudo {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "solo el Super Usuario puede modificar esta configuración"})
	}

	var req request_structs.UpdateReceptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "petición inválida"})
	}
	if req.Enabled == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "campo 'enabled' es obligatorio"})
	}
	if err := h.settingsSvc.UpdateReception(c.UserContext(), *admin, req.Key, *req.Enabled, req.Message); err != nil {
		if err == service.ErrInvalidSettingKey {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo actualizar la configuración"})
	}

	snapshot, err := h.settingsSvc.GetReceptionSwitches(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al consultar la configuración"})
	}
	return c.JSON(snapshot)
}
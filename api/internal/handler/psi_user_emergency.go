// api/internal/handler/psi_user_emergency.go
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// PERSONA DE CONTACTO PARA EMERGENCIAS (AUTOGESTIÓN)
// =========================================================================

// AddEmergencyContact godoc
// @Summary      Registrar persona de contacto para emergencias
// @Description  El agremiado declara hasta 3 personas a las que el Colegio puede avisar si no puede ser localizado o ante un accidente. Se exige nombre, parentesco y al menos un teléfono o correo.
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Accept       json
// @Produce      json
// @Param        request body request_structs.CreateEmergencyContactRequest true "Datos del contacto"
// @Success      201 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Router       /psi/me/emergency [post]
func (h *PsiHandler) AddEmergencyContact(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req request_structs.CreateEmergencyContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.AddEmergencyContact(c.UserContext(), psi, req); err != nil {
		return emergencyContactError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Contacto de emergencia registrado"})
}

// UpdateEmergencyContact godoc
// @Summary      Editar persona de contacto para emergencias
// @Description  Actualización parcial (PATCH). La regla de integridad se revalida sobre el resultado: no puede quedar sin nombre, sin parentesco o sin teléfono ni correo.
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Accept       json
// @Produce      json
// @Param        id      path string true "UUID del contacto"
// @Param        request body request_structs.UpdateEmergencyContactRequest true "Campos parciales"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /psi/me/emergency/{id} [patch]
func (h *PsiHandler) UpdateEmergencyContact(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	contactID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateEmergencyContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateEmergencyContact(c.UserContext(), psi, contactID, req); err != nil {
		return emergencyContactError(c, err)
	}

	return c.JSON(fiber.Map{"message": "Contacto de emergencia actualizado"})
}

// DeleteEmergencyContact godoc
// @Summary      Eliminar persona de contacto para emergencias (Soft Delete)
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Param        id path string true "UUID del contacto"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /psi/me/emergency/{id} [delete]
func (h *PsiHandler) DeleteEmergencyContact(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	contactID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.DeleteEmergencyContact(c.UserContext(), "psi", psi.ID, contactID); err != nil {
		return emergencyContactError(c, err)
	}

	return c.JSON(fiber.Map{"message": "Contacto de emergencia eliminado"})
}

// emergencyContactError traduce los errores de dominio del submódulo a códigos
// HTTP coherentes: los errores de validación (400) se distinguen de los de
// autorización (403) y de inexistencia (404) para que el frontend pueda
// reaccionar sin parsear el texto.
//
// Nota de privacidad: los mensajes NO incluyen el contenido de otros registros
// ni nombres de terceros; un 403 por IDOR responde genérico.
func emergencyContactError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrEmergencyContactIncomplete):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, domain.ErrMaxEmergencyContacts):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, domain.ErrEmergencyContactNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, domain.ErrEmergencyContactPermDenied),
		errors.Is(err, domain.ErrEmergencyContactOwnDenied),
		errors.Is(err, domain.ErrInsufficientPerms):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "no tienes permiso para modificar este contacto de emergencia"})
	case errors.Is(err, domain.ErrPsiNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	default:
		// Los errores de validación de canal (correo con formato inválido) llegan
		// como errores planos; se exponen porque son accionables por el usuario.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
}

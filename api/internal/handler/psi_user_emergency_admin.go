// api/internal/handler/psi_user_emergency_admin.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// PERSONA DE CONTACTO PARA EMERGENCIAS (MODERACIÓN ADMIN)
// =========================================================================
//
// Estos endpoints comparten el gate RBAC de la ficha del psicólogo
// (Sudo || CanUpdatePsi || CanCreatePsi || CanReadPsi; el borrado admite además
// CanDeletePsi) y responden 404 enmascarado cuando el admin no está autenticado
// (ProtectedAdmin404), nunca 403 por falta de token.

// AddEmergencyContactByAdmin godoc
// @Summary      Registrar contacto de emergencia (Admin)
// @Description  Añade una persona de contacto para emergencias a la ficha de un psicólogo desde el panel de moderación (por ejemplo, cuando el Colegio la obtiene por teléfono). La bitácora registra al operador administrativo.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Accept       json
// @Produce      json
// @Param        id      path string true "UUID del Psicólogo"
// @Param        request body request_structs.CreateEmergencyContactRequest true "Datos del contacto"
// @Success      201 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /admin/psi/{id}/emergency [post]
func (h *PsiHandler) AddEmergencyContactByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	var req request_structs.CreateEmergencyContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.AddEmergencyContactByAdmin(c.UserContext(), admin, targetID, req); err != nil {
		return emergencyContactError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Contacto de emergencia registrado"})
}

// UpdateEmergencyContactByAdmin godoc
// @Summary      Editar contacto de emergencia (Admin)
// @Description  Edita un contacto de emergencia de un psicólogo. Verifica que el contacto pertenezca a la ficha indicada (prevención de IDOR) y revalida la regla de integridad.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Accept       json
// @Produce      json
// @Param        id        path string true "UUID del Psicólogo"
// @Param        contactId path string true "UUID del contacto de emergencia"
// @Param        request   body request_structs.UpdateEmergencyContactRequest true "Campos parciales"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /admin/psi/{id}/emergency/{contactId} [patch]
func (h *PsiHandler) UpdateEmergencyContactByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	contactID, err := uuid.Parse(c.Params("contactId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID del contacto no es un UUID válido"})
	}

	var req request_structs.UpdateEmergencyContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateEmergencyContactByAdmin(c.UserContext(), admin, targetID, contactID, req); err != nil {
		return emergencyContactError(c, err)
	}

	return c.JSON(fiber.Map{"message": "Contacto de emergencia actualizado"})
}

// DeleteEmergencyContactByAdmin godoc
// @Summary      Eliminar contacto de emergencia (Admin)
// @Description  Elimina lógicamente un contacto de emergencia de un psicólogo. Verifica que el contacto pertenezca a la ficha indicada (prevención de IDOR).
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Param        id        path string true "UUID del Psicólogo"
// @Param        contactId path string true "UUID del contacto de emergencia"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /admin/psi/{id}/emergency/{contactId} [delete]
func (h *PsiHandler) DeleteEmergencyContactByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	contactID, err := uuid.Parse(c.Params("contactId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID del contacto no es un UUID válido"})
	}

	if err := h.service.DeleteEmergencyContactByAdmin(c.UserContext(), admin, targetID, contactID); err != nil {
		return emergencyContactError(c, err)
	}

	return c.JSON(fiber.Map{"message": "Contacto de emergencia eliminado"})
}

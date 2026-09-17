// api/internal/handler/psi_user_social_admin.go
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// AddSocialNetworkByAdmin godoc
// @Summary      Agregar red social (Admin)
// @Description  Vincula una nueva red social al perfil público de un psicólogo desde el panel de moderación. La auditoría registra al operador administrativo.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "UUID del Psicólogo"
// @Param        request body      request_structs.CreateSocialNetworkRequest true "Datos de la red"
// @Success      201     {object}  map[string]string "message: Red social agregada"
// @Failure      400     {object}  map[string]string "error: ID inválido o JSON inválido"
// @Failure      403     {object}  map[string]string "error: Permisos insuficientes"
// @Failure      404     {object}  map[string]string "error: Psicólogo no encontrado"
// @Router       /admin/psi/{id}/social [post]
func (h *PsiHandler) AddSocialNetworkByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	var req request_structs.CreateSocialNetworkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.AddSocialNetworkByAdmin(c.UserContext(), admin, targetID, req); err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientPerms):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrPsiNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Red social agregada"})
}

// UpdateSocialNetworkByAdmin godoc
// @Summary      Actualizar red social (Admin)
// @Description  Edita los datos de una red social de un psicólogo desde el panel de moderación. Verifica que la red pertenezca a la ficha indicada.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "UUID del Psicólogo"
// @Param        socialId  path      string  true  "UUID de la red social"
// @Param        request   body      request_structs.UpdateSocialNetworkRequest true "Campos parciales"
// @Success      200       {object}  map[string]string "message: Red social actualizada"
// @Failure      400       {object}  map[string]string "error: ID inválido o JSON inválido"
// @Failure      403       {object}  map[string]string "error: Permisos insuficientes"
// @Failure      404       {object}  map[string]string "error: Red social no encontrada"
// @Router       /admin/psi/{id}/social/{socialId} [patch]
func (h *PsiHandler) UpdateSocialNetworkByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	socialID, err := uuid.Parse(c.Params("socialId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID de la red social no es un UUID válido"})
	}

	var req request_structs.UpdateSocialNetworkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateSocialNetworkByAdmin(c.UserContext(), admin, targetID, socialID, req); err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientPerms):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"message": "Red social actualizada"})
}

// DeleteSocialNetworkByAdmin godoc
// @Summary      Eliminar red social (Admin)
// @Description  Elimina lógicamente una red social de un psicólogo desde el panel de moderación.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Param        id        path      string  true  "UUID del Psicólogo"
// @Param        socialId  path      string  true  "UUID de la red social"
// @Success      200       {object}  map[string]string "message: Red social eliminada correctamente"
// @Failure      400       {object}  map[string]string "error: ID inválido"
// @Failure      403       {object}  map[string]string "error: Permisos insuficientes"
// @Failure      404       {object}  map[string]string "error: Red social no encontrada"
// @Router       /admin/psi/{id}/social/{socialId} [delete]
func (h *PsiHandler) DeleteSocialNetworkByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	socialID, err := uuid.Parse(c.Params("socialId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID de la red social no es un UUID válido"})
	}

	if err := h.service.DeleteSocialNetworkByAdmin(c.UserContext(), admin, targetID, socialID); err != nil {
		if errors.Is(err, domain.ErrInsufficientPerms) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Red social eliminada correctamente"})
}
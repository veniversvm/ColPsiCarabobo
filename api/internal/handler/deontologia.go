// api/internal/handler/deontologia.go
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// ListDeontologiaByAdmin godoc
// @Summary      Listar expediente deontológico (Admin)
// @Description  Retorna TODAS las entradas deontológicas de un psicólogo. Acceso exclusivo del personal administrativo autorizado; el psicólogo nunca ve su propio expediente.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Produce      json
// @Param        id   path      string  true  "UUID del Psicólogo"
// @Success      200  {array}   domain.PsiODeontologia
// @Failure      400  {object}  map[string]string "ID inválido"
// @Failure      403  {object}  map[string]string "Permisos insuficientes"
// @Failure      404  {object}  map[string]string "Psicólogo no encontrado"
// @Router       /admin/psi/{id}/deontologia [get]
func (h *PsiHandler) ListDeontologiaByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	entries, err := h.service.ListDeontologiaByPsiID(c.UserContext(), admin, targetID)
	if err != nil {
		if errors.Is(err, domain.ErrPsiNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(entries)
}

// AddDeontologiaByAdmin godoc
// @Summary      Crear entrada deontológica (Admin)
// @Description  Registra una nueva entrada del expediente deontológico de un psicólogo. El contenido es texto plano y se sanitiza en la API.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Accept       json
// @Produce      json
// @Param        id      path      string                             true  "UUID del Psicólogo"
// @Param        request body      request_structs.CreateDeontologiaRequest true "Contenido de la entrada"
// @Success      201     {object}  map[string]string "message: Entrada deontológica creada"
// @Failure      400     {object}  map[string]string "error: ID inválido o contenido vacío"
// @Failure      403     {object}  map[string]string "error: Permisos insuficientes"
// @Failure      404     {object}  map[string]string "error: Psicólogo no encontrado"
// @Router       /admin/psi/{id}/deontologia [post]
func (h *PsiHandler) AddDeontologiaByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	var req request_structs.CreateDeontologiaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.AddDeontologiaByAdmin(c.UserContext(), admin, targetID, req); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRequest):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrPsiNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Entrada deontológica creada"})
}

// DeleteDeontologiaByAdmin godoc
// @Summary      Eliminar entrada deontológica (Admin)
// @Description  Elimina lógicamente (Soft Delete) una entrada deontológica de un psicólogo.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Produce      json
// @Param        id       path      string  true  "UUID del Psicólogo"
// @Param        entryId  path      string  true  "UUID de la entrada deontológica"
// @Success      200      {object}  map[string]string "message: Entrada deontológica eliminada"
// @Failure      400      {object}  map[string]string "error: ID inválido"
// @Failure      403      {object}  map[string]string "error: Permisos insuficientes"
// @Failure      404      {object}  map[string]string "error: Entrada no encontrada"
// @Router       /admin/psi/{id}/deontologia/{entryId} [delete]
func (h *PsiHandler) DeleteDeontologiaByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	entryID, err := uuid.Parse(c.Params("entryId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID de la entrada no es un UUID válido"})
	}

	if err := h.service.DeleteDeontologiaByAdmin(c.UserContext(), admin, entryID); err != nil {
		if errors.Is(err, domain.ErrDeontologiaNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Entrada deontológica eliminada"})
}

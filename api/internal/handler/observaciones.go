// api/internal/handler/observaciones.go
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// ListObservacionesByAdmin godoc
// @Summary      Listar observaciones internas (Admin)
// @Description  Retorna TODAS las observaciones internas de un psicólogo. Acceso exclusivo del personal administrativo autorizado; el psicólogo nunca ve sus propias observaciones.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Produce      json
// @Param        id   path      string  true  "UUID del Psicólogo"
// @Success      200  {array}   domain.PsiObservations
// @Failure      400  {object}  map[string]string "ID inválido"
// @Failure      403  {object}  map[string]string "Permisos insuficientes"
// @Failure      404  {object}  map[string]string "Psicólogo no encontrado"
// @Router       /admin/psi/{id}/observaciones [get]
func (h *PsiHandler) ListObservacionesByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	entries, err := h.service.ListObservacionesByPsiID(c.UserContext(), admin, targetID)
	if err != nil {
		if errors.Is(err, domain.ErrPsiNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(entries)
}

// AddObservacionesByAdmin godoc
// @Summary      Crear observación interna (Admin)
// @Description  Registra una nueva observación interna sobre un psicólogo. El contenido es texto plano y se sanitiza en la API.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Accept       json
// @Produce      json
// @Param        id      path      string                               true  "UUID del Psicólogo"
// @Param        request body      request_structs.CreateObservacionesRequest true "Contenido de la observación"
// @Success      201     {object}  map[string]string "message: Observación creada"
// @Failure      400     {object}  map[string]string "error: ID inválido o contenido vacío"
// @Failure      403     {object}  map[string]string "error: Permisos insuficientes"
// @Failure      404     {object}  map[string]string "error: Psicólogo no encontrado"
// @Router       /admin/psi/{id}/observaciones [post]
func (h *PsiHandler) AddObservacionesByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	var req request_structs.CreateObservacionesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.AddObservacionesByAdmin(c.UserContext(), admin, targetID, req); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRequest):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrPsiNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Observación creada"})
}

// UpdateObservacionesByAdmin godoc
// @Summary      Editar observación interna (Admin)
// @Description  Edita el contenido de una observación interna existente. El contenido es texto plano y se sanitiza en la API.
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                               true  "UUID del Psicólogo"
// @Param        entryId  path      string                               true  "UUID de la observación"
// @Param        request  body      request_structs.UpdateObservacionesRequest true "Nuevo contenido de la observación"
// @Success      200      {object}  map[string]string "message: Observación actualizada"
// @Failure      400      {object}  map[string]string "error: ID inválido o contenido vacío"
// @Failure      403      {object}  map[string]string "error: Permisos insuficientes"
// @Failure      404      {object}  map[string]string "error: Observación no encontrada"
// @Router       /admin/psi/{id}/observaciones/{entryId} [patch]
func (h *PsiHandler) UpdateObservacionesByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	entryID, err := uuid.Parse(c.Params("entryId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID de la observación no es un UUID válido"})
	}

	var req request_structs.UpdateObservacionesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateObservacionesByAdmin(c.UserContext(), admin, entryID, req); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRequest):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrObservacionesNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"message": "Observación actualizada"})
}

// api/internal/handler/kanban_handler.go
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// KanbanHandler gestiona el módulo de Proyectos (tableros Kanban) del panel admin.
type KanbanHandler struct {
	service *service.KanbanService
}

// NewKanbanHandler crea una instancia del handler inyectando el servicio.
func NewKanbanHandler(svc *service.KanbanService) *KanbanHandler {
	return &KanbanHandler{service: svc}
}

// parseUUID extrae y valida un UUID de los parámetros de ruta.
func parseUUIDParam(c *fiber.Ctx, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Params(key))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// mapKanbanError traduce errores de dominio a respuestas HTTP.
func mapKanbanError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrInsufficientPerms),
		errors.Is(err, domain.ErrNotProjectMember),
		errors.Is(err, domain.ErrPermissionDenied):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, domain.ErrProjectNotFound),
		errors.Is(err, domain.ErrColumnNotFound),
		errors.Is(err, domain.ErrCardNotFound),
		errors.Is(err, domain.ErrNoteNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, domain.ErrNoteLimitReached),
		errors.Is(err, domain.ErrNoteTooLong),
		errors.Is(err, domain.ErrInvalidMemberRole),
		errors.Is(err, domain.ErrMemberAlreadyExists),
		errors.Is(err, domain.ErrInvalidRequest):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

// =========================================================================
// PROYECTOS
// =========================================================================

// ListProjects godoc
// @Summary      Listar proyectos Kanban
// @Description  Retorna los proyectos accesibles para el admin autenticado (dueños, miembro o master).
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Produce      json
// @Success      200  {array}  domain.KanbanProject
// @Router       /admin/projects [get]
func (h *KanbanHandler) ListProjects(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	projects, err := h.service.ListProjects(c.UserContext(), admin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al listar proyectos"})
	}
	return c.JSON(fiber.Map{"data": projects})
}

// CreateProject godoc
// @Summary      Crear proyecto Kanban
// @Description  Crea un proyecto (el admin autenticado es el dueño) y siembra las columnas por defecto.
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.CreateProjectRequest  true  "Datos del proyecto"
// @Success      201      {object}  domain.KanbanProject
// @Router       /admin/projects [post]
func (h *KanbanHandler) CreateProject(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	var req request_structs.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	project, err := h.service.CreateProject(c.UserContext(), admin, req)
	if err != nil {
		return mapKanbanError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

// GetProject godoc
// @Summary      Tablero Kanban de un proyecto
// @Description  Retorna el tablero completo: proyecto + columnas con sus tarjetas y notas.
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Produce      json
// @Param        id   path      string  true  "UUID del proyecto"
// @Success      200  {object}  map[string]interface{}
// @Router       /admin/projects/{id} [get]
func (h *KanbanHandler) GetProject(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	projectID, ok := parseUUIDParam(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	project, columns, _, err := h.service.GetBoard(c.UserContext(), admin, projectID)
	if err != nil {
		return mapKanbanError(c, err)
	}
	if columns == nil {
		columns = []domain.KanbanColumn{}
	}
	return c.JSON(fiber.Map{"project": project, "columns": columns})
}

// UpdateProject godoc
// @Summary      Actualizar proyecto Kanban
// @Description  Renombra o edita la descripción de un proyecto (solo dueño o master).
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        id       path      string                              true  "UUID del proyecto"
// @Param        request  body      request_structs.UpdateProjectRequest  true  "Campos a modificar"
// @Success      200      {object}  map[string]string
// @Router       /admin/projects/{id} [patch]
func (h *KanbanHandler) UpdateProject(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	projectID, ok := parseUUIDParam(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateProject(c.UserContext(), admin, projectID, req); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Proyecto actualizado"})
}

// DeleteProject godoc
// @Summary      Eliminar proyecto Kanban
// @Description  Elimina definitivamente el proyecto, columnas, tarjetas y notas (solo dueño o master).
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Param        id   path      string  true  "UUID del proyecto"
// @Success      200  {object}  map[string]string
// @Router       /admin/projects/{id} [delete]
func (h *KanbanHandler) DeleteProject(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	projectID, ok := parseUUIDParam(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.DeleteProject(c.UserContext(), admin, projectID); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Proyecto eliminado"})
}

// =========================================================================
// MIEMBROS
// =========================================================================

// ListMembers godoc
// @Summary      Listar miembros de un proyecto
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Produce      json
// @Param        id   path      string  true  "UUID del proyecto"
// @Success      200  {array}   domain.KanbanMember
// @Router       /admin/projects/{id}/members [get]
func (h *KanbanHandler) ListMembers(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	projectID, ok := parseUUIDParam(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	members, err := h.service.ListMembers(c.UserContext(), admin, projectID)
	if err != nil {
		return mapKanbanError(c, err)
	}
	if members == nil {
		members = []domain.KanbanMember{}
	}
	return c.JSON(fiber.Map{"data": members})
}

// AddMember godoc
// @Summary      Invitar administrador a un proyecto
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        id       path      string                            true  "UUID del proyecto"
// @Param        request  body      request_structs.AddMemberRequest  true  "Admin objetivo + rol"
// @Success      201      {object}  map[string]string
// @Router       /admin/projects/{id}/members [post]
func (h *KanbanHandler) AddMember(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	projectID, ok := parseUUIDParam(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.AddMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.AddMember(c.UserContext(), admin, projectID, req); err != nil {
		return mapKanbanError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Miembro añadido"})
}

// UpdateMember godoc
// @Summary      Cambiar rol de un miembro
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        memberId  path      string                             true  "UUID del miembro"
// @Param        request   body      request_structs.UpdateMemberRequest  true  "Nuevo rol"
// @Success      200       {object}  map[string]string
// @Router       /admin/projects/members/{memberId} [patch]
func (h *KanbanHandler) UpdateMember(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	memberID, ok := parseUUIDParam(c, "memberId")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateMember(c.UserContext(), admin, memberID, req); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Rol actualizado"})
}

// RemoveMember godoc
// @Summary      Quitar miembro de un proyecto
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Param        memberId  path      string  true  "UUID del miembro"
// @Success      200       {object}  map[string]string
// @Router       /admin/projects/members/{memberId} [delete]
func (h *KanbanHandler) RemoveMember(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	memberID, ok := parseUUIDParam(c, "memberId")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.RemoveMember(c.UserContext(), admin, memberID); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Miembro eliminado"})
}

// =========================================================================
// COLUMNAS
// =========================================================================

// CreateColumn godoc
// @Summary      Crear columna en el tablero
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        id       path      string                              true  "UUID del proyecto"
// @Param        request  body      request_structs.CreateColumnRequest  true  "Título de la columna"
// @Success      201      {object}  domain.KanbanColumn
// @Router       /admin/projects/{id}/columns [post]
func (h *KanbanHandler) CreateColumn(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	projectID, ok := parseUUIDParam(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.CreateColumnRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	column, err := h.service.CreateColumn(c.UserContext(), admin, projectID, req)
	if err != nil {
		return mapKanbanError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(column)
}

// UpdateColumn godoc
// @Summary      Actualizar columna
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        columnId  path      string                              true  "UUID de la columna"
// @Param        request   body      request_structs.UpdateColumnRequest  true  "Campos a modificar"
// @Success      200       {object}  map[string]string
// @Router       /admin/projects/columns/{columnId} [patch]
func (h *KanbanHandler) UpdateColumn(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	columnID, ok := parseUUIDParam(c, "columnId")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateColumnRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateColumn(c.UserContext(), admin, columnID, req); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Columna actualizada"})
}

// DeleteColumn godoc
// @Summary      Eliminar columna
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Param        columnId  path      string  true  "UUID de la columna"
// @Success      200       {object}  map[string]string
// @Router       /admin/projects/columns/{columnId} [delete]
func (h *KanbanHandler) DeleteColumn(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	columnID, ok := parseUUIDParam(c, "columnId")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.DeleteColumn(c.UserContext(), admin, columnID); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Columna eliminada"})
}

// =========================================================================
// TARJETAS
// =========================================================================

// CreateCard godoc
// @Summary      Crear tarjeta Kanban
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        id       path      string                            true  "UUID del proyecto"
// @Param        request  body      request_structs.CreateCardRequest  true  "Datos de la tarjeta"
// @Success      201      {object}  domain.KanbanCard
// @Router       /admin/projects/{id}/cards [post]
func (h *KanbanHandler) CreateCard(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	projectID, ok := parseUUIDParam(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.CreateCardRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	card, err := h.service.CreateCard(c.UserContext(), admin, projectID, req)
	if err != nil {
		return mapKanbanError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(card)
}

// UpdateCard godoc
// @Summary      Actualizar o mover tarjeta Kanban
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        cardId   path      string                            true  "UUID de la tarjeta"
// @Param        request  body      request_structs.UpdateCardRequest  true  "Campos a modificar"
// @Success      200      {object}  map[string]string
// @Router       /admin/projects/cards/{cardId} [patch]
func (h *KanbanHandler) UpdateCard(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	cardID, ok := parseUUIDParam(c, "cardId")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateCardRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateCard(c.UserContext(), admin, cardID, req); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Tarjeta actualizada"})
}

// DeleteCard godoc
// @Summary      Eliminar tarjeta Kanban
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Param        cardId   path      string  true  "UUID de la tarjeta"
// @Success      200      {object}  map[string]string
// @Router       /admin/projects/cards/{cardId} [delete]
func (h *KanbanHandler) DeleteCard(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	cardID, ok := parseUUIDParam(c, "cardId")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.DeleteCard(c.UserContext(), admin, cardID); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Tarjeta eliminada"})
}

// =========================================================================
// NOTAS
// =========================================================================

// CreateNote godoc
// @Summary      Añadir nota a una tarjeta
// @Description  Cada tarjeta admite hasta 10 notas de máximo 500 caracteres.
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Accept       json
// @Produce      json
// @Param        cardId   path      string                            true  "UUID de la tarjeta"
// @Param        request  body      request_structs.CreateNoteRequest  true  "Contenido de la nota"
// @Success      201      {object}  domain.KanbanNote
// @Router       /admin/projects/cards/{cardId}/notes [post]
func (h *KanbanHandler) CreateNote(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	cardID, ok := parseUUIDParam(c, "cardId")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.CreateNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	note, err := h.service.CreateNote(c.UserContext(), admin, cardID, req)
	if err != nil {
		return mapKanbanError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(note)
}

// DeleteNote godoc
// @Summary      Eliminar nota de una tarjeta
// @Description  Solo el autor de la nota, el dueño del proyecto o un master pueden borrarla.
// @Security     BearerAuth
// @Tags         Administración - Proyectos (Kanban)
// @Param        noteId   path      string  true  "UUID de la nota"
// @Success      200      {object}  map[string]string
// @Router       /admin/projects/notes/{noteId} [delete]
func (h *KanbanHandler) DeleteNote(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	noteID, ok := parseUUIDParam(c, "noteId")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.DeleteNote(c.UserContext(), admin, noteID); err != nil {
		return mapKanbanError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Nota eliminada"})
}

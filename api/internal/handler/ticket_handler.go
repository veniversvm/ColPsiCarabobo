// api/internal/handler/ticket_handler.go
package handler

import (
	"errors"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// TicketHandler gestiona el módulo de Tickets de Solicitudes (portal psi y panel admin).
type TicketHandler struct {
	service *service.TicketService
}

// NewTicketHandler crea una instancia del handler inyectando el servicio.
func NewTicketHandler(svc *service.TicketService) *TicketHandler {
	return &TicketHandler{service: svc}
}

// GetStatus godoc
// @Summary      Estado de recepción de tickets (portal psi)
// @Description  Indica si la recepción de tickets de solicitudes está habilitada. Cuando está desactivada, el campo `message` explica el motivo público.
// @Security     BearerAuth
// @Tags         Tickets - Portal Psi
// @Produce      json
// @Success      200 {object} domain.ReceptionSetting
// @Router       /psi/tickets/status [get]
func (h *TicketHandler) GetStatus(c *fiber.Ctx) error {
	status, err := h.service.ReceptionStatus(c.UserContext())
	if err != nil {
		log.Error().Err(err).Str("component", "tickets").Msg("Error al consultar estado de recepción de tickets")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al consultar el estado"})
	}
	return c.JSON(status)
}

// parseTicketUint extrae y valida un ID numérico (uint) de los parámetros de ruta.
func parseTicketUint(c *fiber.Ctx, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Params(key), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// mapTicketError traduce errores de dominio del módulo a respuestas HTTP.
func mapTicketError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrInsufficientPerms),
		errors.Is(err, domain.ErrTicketNotOwner):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, domain.ErrTicketNotFound),
		errors.Is(err, domain.ErrMotivoNotFound),
		errors.Is(err, domain.ErrEstadoNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, domain.ErrTicketLimitReached),
		errors.Is(err, domain.ErrTicketClosed),
		errors.Is(err, domain.ErrMaxConsecutiveComments),
		errors.Is(err, domain.ErrMensajeTooLong),
		errors.Is(err, domain.ErrMensajeVacio),
		errors.Is(err, domain.ErrCloseReasonRequired),
		errors.Is(err, domain.ErrEstadoNotInMotivo),
		errors.Is(err, domain.ErrMotivoInUse),
		errors.Is(err, domain.ErrEstadoInUse),
		errors.Is(err, domain.ErrMotivoLimitInvalid),
		errors.Is(err, domain.ErrInvalidRequest):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		var rdErr *service.ReceptionDisabledError
		if errors.As(err, &rdErr) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "reception_disabled",
				"message": rdErr.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

// ── Helpers de request (JSON o multipart con anexos) ────────────────────────

// isMultipart detecta si la petición viaja como multipart/form-data.
func isMultipart(c *fiber.Ctx) bool {
	return strings.HasPrefix(c.Get("Content-Type"), "multipart/form-data")
}

// formField lee un campo de un formulario multipart ya parseado.
func formField(form *multipart.Form, key string) string {
	if form == nil {
		return ""
	}
	vals := form.Value[key]
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

// formUint lee un campo numérico de un formulario multipart.
func formUint(form *multipart.Form, key string) (uint, error) {
	s := formField(form, key)
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// parseCreateTicket extrae la carga útil de la creación de un ticket: acepta
// JSON plano o multipart (con el campo "files" para los anexos iniciales).
func parseCreateTicket(c *fiber.Ctx) (request_structs.CreateTicketRequest, []*multipart.FileHeader, error) {
	var req request_structs.CreateTicketRequest
	if isMultipart(c) {
		form, err := c.MultipartForm()
		if err != nil {
			return req, nil, err
		}
		if req.MotivoID, err = formUint(form, "motivo_id"); err != nil {
			return req, nil, err
		}
		req.Title = formField(form, "title")
		req.Description = formField(form, "description")
		return req, form.File["files"], nil
	}
	if err := c.BodyParser(&req); err != nil {
		return req, nil, err
	}
	return req, nil, nil
}

// parseMensaje extrae el comentario de la conversación (JSON o multipart con
// el campo "files" para anexos).
func parseMensaje(c *fiber.Ctx) (string, []*multipart.FileHeader, error) {
	if isMultipart(c) {
		form, err := c.MultipartForm()
		if err != nil {
			return "", nil, err
		}
		return formField(form, "message"), form.File["files"], nil
	}
	var req request_structs.AddMensajeRequest
	if err := c.BodyParser(&req); err != nil {
		return "", nil, err
	}
	return req.Message, nil, nil
}

// =========================================================================
// CONFIGURACIÓN (panel admin y lectura psi)
// =========================================================================

// ListMotivos godoc
// @Summary      Listar motivos de tickets
// @Description  Retorna los motivos con sus estados (configuración del colegio).
// @Security     BearerAuth
// @Tags         Tickets - Configuración
// @Produce      json
// @Success      200  {array}  domain.TicketMotivo
// @Router       /admin/tickets/motivos [get]
func (h *TicketHandler) ListMotivos(c *fiber.Ctx) error {
	motivos, err := h.service.ListMotivosConfig(c.UserContext())
	if err != nil {
		return mapTicketError(c, err)
	}
	if motivos == nil {
		motivos = []domain.TicketMotivo{}
	}
	return c.JSON(fiber.Map{"data": motivos})
}

// CreateMotivo godoc
// @Summary      Crear motivo de tickets
// @Description  Crea un motivo en un área y siembra los estados por defecto.
// @Security     BearerAuth
// @Tags         Tickets - Configuración
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.CreateTicketMotivoRequest  true  "Datos del motivo"
// @Success      201      {object}  domain.TicketMotivo
// @Router       /admin/tickets/motivos [post]
func (h *TicketHandler) CreateMotivo(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	var req request_structs.CreateTicketMotivoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	motivo, err := h.service.CreateMotivo(c.UserContext(), admin, req)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(motivo)
}

// UpdateMotivo godoc
// @Summary      Actualizar motivo de tickets
// @Security     BearerAuth
// @Tags         Tickets - Configuración
// @Accept       json
// @Produce      json
// @Param        id       path      string                               true  "ID del motivo"
// @Param        request  body      request_structs.UpdateTicketMotivoRequest  true  "Campos a modificar"
// @Success      200      {object}  domain.TicketMotivo
// @Router       /admin/tickets/motivos/{id} [patch]
func (h *TicketHandler) UpdateMotivo(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	motivoID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateTicketMotivoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	motivo, err := h.service.UpdateMotivo(c.UserContext(), admin, motivoID, req)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(motivo)
}

// DeleteMotivo godoc
// @Summary      Eliminar motivo de tickets
// @Description  Solo se puede eliminar un motivo sin tickets asociados.
// @Security     BearerAuth
// @Tags         Tickets - Configuración
// @Param        id   path      string  true  "ID del motivo"
// @Success      200  {object}  map[string]string
// @Router       /admin/tickets/motivos/{id} [delete]
func (h *TicketHandler) DeleteMotivo(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	motivoID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.DeleteMotivo(c.UserContext(), admin, motivoID); err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Motivo eliminado"})
}

// ListEstadosConfig godoc
// @Summary      Listar estados de un motivo
// @Security     BearerAuth
// @Tags         Tickets - Configuración
// @Produce      json
// @Param        id   path  string  true  "ID del motivo"
// @Success      200  {array}  domain.TicketEstado
// @Router       /admin/tickets/motivos/{id}/estados [get]
func (h *TicketHandler) ListEstadosConfig(c *fiber.Ctx) error {
	motivoID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	estados, err := h.service.ListEstadosConfig(c.UserContext(), motivoID)
	if err != nil {
		return mapTicketError(c, err)
	}
	if estados == nil {
		estados = []domain.TicketEstado{}
	}
	return c.JSON(fiber.Map{"data": estados})
}

// CreateEstado godoc
// @Summary      Crear estado de un motivo
// @Security     BearerAuth
// @Tags         Tickets - Configuración
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.CreateTicketEstadoRequest  true  "Datos del estado"
// @Success      201      {object}  domain.TicketEstado
// @Router       /admin/tickets/estados [post]
func (h *TicketHandler) CreateEstado(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	var req request_structs.CreateTicketEstadoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	estado, err := h.service.CreateEstado(c.UserContext(), admin, req)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(estado)
}

// UpdateEstadoConfig godoc
// @Summary      Actualizar estado de un motivo
// @Description  Permite renombrar, reordenar o marcar el estado como cerrado.
// @Security     BearerAuth
// @Tags         Tickets - Configuración
// @Accept       json
// @Produce      json
// @Param        id       path      string                               true  "ID del estado"
// @Param        request  body      request_structs.UpdateTicketEstadoRequest  true  "Campos a modificar"
// @Success      200      {object}  domain.TicketEstado
// @Router       /admin/tickets/estados/{id} [patch]
func (h *TicketHandler) UpdateEstadoConfig(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	estadoID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateTicketEstadoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	estado, err := h.service.UpdateEstadoConfig(c.UserContext(), admin, estadoID, req)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(estado)
}

// DeleteEstadoConfig godoc
// @Summary      Eliminar estado de un motivo
// @Description  Solo se puede eliminar un estado que ningún ticket esté usando.
// @Security     BearerAuth
// @Tags         Tickets - Configuración
// @Param        id   path      string  true  "ID del estado"
// @Success      200  {object}  map[string]string
// @Router       /admin/tickets/estados/{id} [delete]
func (h *TicketHandler) DeleteEstadoConfig(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	estadoID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.DeleteEstadoConfig(c.UserContext(), admin, estadoID); err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Estado eliminado"})
}

// =========================================================================
// PORTAL PSI
// =========================================================================

// GetConfigPSI godoc
// @Summary      Configuración de tickets para el psicólogo
// @Description  Retorna los motivos con sus estados (y su límite tickets_per_psi) para que el psi pueda crear un ticket.
// @Security     BearerAuth
// @Tags         Tickets - Portal Psi
// @Produce      json
// @Success      200  {array}  domain.TicketMotivo
// @Router       /psi/tickets/config [get]
func (h *TicketHandler) GetConfigPSI(c *fiber.Ctx) error {
	motivos, err := h.service.ListMotivosConfig(c.UserContext())
	if err != nil {
		return mapTicketError(c, err)
	}
	if motivos == nil {
		motivos = []domain.TicketMotivo{}
	}
	return c.JSON(fiber.Map{"data": motivos})
}

// ListMyTickets godoc
// @Summary      Listar mis tickets
// @Security     BearerAuth
// @Tags         Tickets - Portal Psi
// @Produce      json
// @Param        page   query  int  false  "Página"
// @Param        limit  query  int  false  "Límite"
// @Success      200    {object}  map[string]interface{}
// @Router       /psi/tickets [get]
func (h *TicketHandler) ListMyTickets(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	page, limit := parsePagination(c)
	tickets, total, err := h.service.ListMyTickets(c.UserContext(), psi, page, limit)
	if err != nil {
		return mapTicketError(c, err)
	}
	if tickets == nil {
		tickets = []domain.Ticket{}
	}
	return c.JSON(fiber.Map{"data": tickets, "total": total, "page": page, "limit": limit})
}

// CreateTicket godoc
// @Summary      Crear un ticket de solicitud
// @Description  El psicólogo abre un ticket en un área/motivo con un título, descripción y anexos opcionales (multipart, campo "files").
// @Security     BearerAuth
// @Tags         Tickets - Portal Psi
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.CreateTicketRequest  true  "Datos del ticket (o multipart)"
// @Success      201      {object}  domain.Ticket
// @Router       /psi/tickets [post]
func (h *TicketHandler) CreateTicket(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	req, files, err := parseCreateTicket(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "petición inválida"})
	}

	ticket, err := h.service.CreateTicket(c.UserContext(), psi, req, files)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(ticket)
}

// GetTicketAsPsi godoc
// @Summary      Ver detalle de un ticket propio
// @Security     BearerAuth
// @Tags         Tickets - Portal Psi
// @Produce      json
// @Param        id   path  string  true  "ID del ticket"
// @Success      200  {object}  domain.Ticket
// @Router       /psi/tickets/{id} [get]
func (h *TicketHandler) GetTicketAsPsi(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	ticketID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	ticket, err := h.service.GetTicketAsPsi(c.UserContext(), psi, ticketID)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(ticket)
}

// AddMensajePsi godoc
// @Summary      Comentar en mi ticket
// @Description  Publica un mensaje en la conversación (máx 1000 chars, no más de 3 seguidos). Anexos opcionales en multipart (campo "files").
// @Security     BearerAuth
// @Tags         Tickets - Portal Psi
// @Accept       json
// @Produce      json
// @Param        id       path  string                              true  "ID del ticket"
// @Param        request  body  request_structs.AddMensajeRequest    true  "Mensaje (o multipart)"
// @Success      201      {object}  domain.TicketMensaje
// @Router       /psi/tickets/{id}/mensaje [post]
func (h *TicketHandler) AddMensajePsi(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	ticketID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	message, files, err := parseMensaje(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "petición inválida"})
	}

	mensaje, err := h.service.AddMensajeAsPsi(c.UserContext(), psi, ticketID, message, files)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(mensaje)
}

// CloseTicketPsi godoc
// @Summary      Cerrar mi ticket
// @Description  El psicólogo cierra su propio ticket indicando un motivo de cierre obligatorio.
// @Security     BearerAuth
// @Tags         Tickets - Portal Psi
// @Accept       json
// @Produce      json
// @Param        id       path  string                           true  "ID del ticket"
// @Param        request  body  request_structs.CloseTicketRequest  true  "Motivo de cierre"
// @Success      200      {object}  domain.Ticket
// @Router       /psi/tickets/{id}/cerrar [post]
func (h *TicketHandler) CloseTicketPsi(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	ticketID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.CloseTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	ticket, err := h.service.CloseTicketAsPsi(c.UserContext(), psi, ticketID, req)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(ticket)
}

// =========================================================================
// PANEL ADMINISTRATIVO
// =========================================================================

// ListTicketsAdmin godoc
// @Summary      Listar tickets (cola FIFO)
// @Description  Por defecto lista solo los abiertos, ordenados por orden de llegada. Filtros opcionales: motivo_id, estado_id, psi_id, q, solo_abiertos.
// @Security     BearerAuth
// @Tags         Tickets - Administración
// @Produce      json
// @Param        solo_abiertos  query  bool   false  "Solo abiertos (default true)"
// @Param        motivo_id      query  int    false  "Filtrar por motivo"
// @Param        estado_id      query  int    false  "Filtrar por estado"
// @Param        psi_id         query  string false  "Filtrar por psicólogo"
// @Param        q              query  string false  "Búsqueda en título/descripción"
// @Param        page           query  int    false  "Página"
// @Param        limit          query  int    false  "Límite"
// @Success      200            {object}  map[string]interface{}
// @Router       /admin/tickets [get]
func (h *TicketHandler) ListTicketsAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	page, limit := parsePagination(c)
	filter := domain.TicketFilter{
		SoloAbiertos: c.Query("solo_abiertos", "true") != "false",
		Search:       strings.TrimSpace(c.Query("q")),
		Page:         page,
		Limit:        limit,
	}
	if s := c.Query("cursor"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			cur := uint(v)
			filter.Cursor = &cur
		}
	}
	if s := c.Query("motivo_id"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			id := uint(v)
			filter.MotivoID = &id
		}
	}
	if s := c.Query("estado_id"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			id := uint(v)
			filter.EstadoID = &id
		}
	}
	if s := c.Query("psi_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			filter.PsiUserID = &id
		}
	}

	tickets, total, err := h.service.ListTicketsAdmin(c.UserContext(), admin, filter)
	if err != nil {
		return mapTicketError(c, err)
	}
	if tickets == nil {
		tickets = []domain.Ticket{}
	}
	res := fiber.Map{"data": tickets, "total": total, "page": page, "limit": limit}
	if len(tickets) == limit && len(tickets) > 0 {
		res["next_cursor"] = tickets[len(tickets)-1].ID
	}
	return c.JSON(res)
}

// CountPendientesAdmin godoc
// @Summary      Contar tickets abiertos
// @Description  Número de tickets no cerrados (badge del menú admin).
// @Security     BearerAuth
// @Tags         Tickets - Administración
// @Produce      json
// @Success      200  {object}  map[string]int64
// @Router       /admin/tickets/pendientes-count [get]
func (h *TicketHandler) CountPendientesAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	count, err := h.service.CountPendientesAdmin(c.UserContext(), admin)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(fiber.Map{"pendientes": count})
}

// GetTicketAdmin godoc
// @Summary      Ver detalle de un ticket
// @Description  Retorna el ticket con el área, motivo, estado, conversación completa e historial de cambios.
// @Security     BearerAuth
// @Tags         Tickets - Administración
// @Produce      json
// @Param        id   path  string  true  "ID del ticket"
// @Success      200  {object}  domain.Ticket
// @Router       /admin/tickets/{id} [get]
func (h *TicketHandler) GetTicketAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	ticketID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	ticket, err := h.service.GetTicketAdmin(c.UserContext(), admin, ticketID)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(ticket)
}

// UpdateTicketEstado godoc
// @Summary      Cambiar estado de un ticket
// @Description  El nuevo estado debe pertenecer al motivo del ticket. Si el estado es "cerrado" se registra el cierre.
// @Security     BearerAuth
// @Tags         Tickets - Administración
// @Accept       json
// @Produce      json
// @Param        id       path  string                             true  "ID del ticket"
// @Param        request  body  request_structs.UpdateTicketEstado  true  "Nuevo estado + motivo"
// @Success      200      {object}  domain.Ticket
// @Router       /admin/tickets/{id}/estado [patch]
func (h *TicketHandler) UpdateTicketEstado(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	ticketID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateTicketEstado
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	ticket, err := h.service.UpdateTicketEstado(c.UserContext(), admin, ticketID, req)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(ticket)
}

// AddMensajeAdmin godoc
// @Summary      Responder en la conversación de un ticket
// @Description  El admin responde en el ticket (sin límite de mensajes seguidos). Anexos opcionales en multipart (campo "files").
// @Security     BearerAuth
// @Tags         Tickets - Administración
// @Accept       json
// @Produce      json
// @Param        id       path  string                            true  "ID del ticket"
// @Param        request  body  request_structs.AddMensajeRequest  true  "Mensaje (o multipart)"
// @Success      201      {object}  domain.TicketMensaje
// @Router       /admin/tickets/{id}/mensaje [post]
func (h *TicketHandler) AddMensajeAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	ticketID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	message, files, err := parseMensaje(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "petición inválida"})
	}

	mensaje, err := h.service.AddMensajeAsAdmin(c.UserContext(), admin, ticketID, message, files)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(mensaje)
}

// CloseTicketAdmin godoc
// @Summary      Cerrar un ticket
// @Description  El admin cierra el ticket indicando un motivo de cierre obligatorio. Se notifica al psicólogo.
// @Security     BearerAuth
// @Tags         Tickets - Administración
// @Accept       json
// @Produce      json
// @Param        id       path  string                           true  "ID del ticket"
// @Param        request  body  request_structs.CloseTicketRequest  true  "Motivo de cierre"
// @Success      200      {object}  domain.Ticket
// @Router       /admin/tickets/{id}/cerrar [post]
func (h *TicketHandler) CloseTicketAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no autenticado"})
	}

	ticketID, ok := parseTicketUint(c, "id")
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.CloseTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	ticket, err := h.service.CloseTicketAsAdmin(c.UserContext(), admin, ticketID, req)
	if err != nil {
		return mapTicketError(c, err)
	}
	return c.JSON(ticket)
}
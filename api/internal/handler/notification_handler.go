// api/internal/handler/notification_handler.go
package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// NotificationHandler gestiona los endpoints de notificaciones (admin y agremiado).
type NotificationHandler struct {
	service *service.NotificationService
}

// NewNotificationHandler crea una instancia del handler con su servicio inyectado.
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: svc}
}

// =========================================================================
// ADMIN
// =========================================================================

// PreviewRecipients godoc
// @Summary      Previsualizar destinatarios
// @Description  Retorna los destinatarios potenciales de una notificación sin crearla.
// @Security     BearerAuth
// @Tags         Administración - Notificaciones
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.PreviewNotificationRequest  true  "Datos de la previsualización"
// @Success      200      {object}  map[string]interface{}
// @Failure      403      {object}  map[string]string
// @Router       /notifications/admin/preview [post]
func (h *NotificationHandler) PreviewRecipients(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req request_structs.PreviewNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	res, err := h.service.PreviewRecipients(c.UserContext(), admin, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotificationPermDenied) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}

// CreateNotification godoc
// @Summary      Crear notificación
// @Description  Crea una notificación. Si lleva scheduled_at futuro se programa; si no, se envía inmediatamente.
// @Security     BearerAuth
// @Tags         Administración - Notificaciones
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.CreateNotificationRequest  true  "Datos de la notificación"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Router       /notifications/admin [post]
func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req request_structs.CreateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	res, err := h.service.CreateNotification(c.UserContext(), admin, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotificationPermDenied):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrNotificationInvalidSchedule),
			errors.Is(err, domain.ErrNotificationInvalidTargetType):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

// AttachFile godoc
// @Summary      Adjuntar archivo a notificación
// @Description  Sube una imagen/archivo adjunto a una notificación y la guarda en S3.
// @Security     BearerAuth
// @Tags         Administración - Notificaciones
// @Accept       multipart/form-data
// @Produce      json
// @Param        id      path      string  true  "UUID de la notificación"
// @Param        file    formData  file    true  "Archivo adjunto"
// @Success      201     {object}  map[string]string
// @Failure      403     {object}  map[string]string
// @Router       /notifications/admin/{id}/attach [post]
func (h *NotificationHandler) AttachFile(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "archivo requerido"})
	}

	url, err := h.service.AttachFile(c.UserContext(), admin, id, file)
	if err != nil {
		if errors.Is(err, domain.ErrNotificationPermDenied) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Archivo adjuntado", "url": url})
}

// GetMyNotifications godoc
// @Summary      Listar notificaciones del admin
// @Description  Lista las notificaciones creadas por el administrador, paginadas por fecha desc.
// @Security     BearerAuth
// @Tags         Administración - Notificaciones
// @Produce      json
// @Param        page   query  int  false  "Número de página (Default: 1)"
// @Param        limit  query  int  false  "Items por página (Default: 10)"
// @Success      200    {object}  map[string]interface{}
// @Router       /notifications/admin [get]
func (h *NotificationHandler) GetMyNotifications(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	page, limit := parsePagination(c)
	res, err := h.service.ListMyNotifications(c.UserContext(), admin, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// GetNotificationDetail godoc
// @Summary      Detalle de notificación (admin)
// @Description  Retorna el detalle con estadísticas de lectura (leídos/total).
// @Security     BearerAuth
// @Tags         Administración - Notificaciones
// @Produce      json
// @Param        id  path  string  true  "UUID de la notificación"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /notifications/admin/{id} [get]
func (h *NotificationHandler) GetNotificationDetail(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	res, err := h.service.GetNotificationDetail(c.UserContext(), admin, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotificationNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, domain.ErrNotificationPermDenied) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// CancelNotification godoc
// @Summary      Cancelar notificación programada
// @Description  Cancela una notificación programada pendiente (solo status = pending).
// @Security     BearerAuth
// @Tags         Administración - Notificaciones
// @Produce      json
// @Param        id  path  string  true  "UUID de la notificación"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /notifications/admin/{id} [delete]
func (h *NotificationHandler) CancelNotification(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.CancelNotification(c.UserContext(), admin, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotificationPermDenied):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrNotificationCannotCancel):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrNotificationNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}
	return c.JSON(fiber.Map{"message": "Notificación cancelada"})
}

// GetTargets godoc
// @Summary      Ver destinatarios de notificación
// @Description  Retorna la lista de destinatarios de una notificación con su estado de lectura.
// @Security     BearerAuth
// @Tags         Administración - Notificaciones
// @Produce      json
// @Param        id  path  string  true  "UUID de la notificación"
// @Success      200  {array}  domain.NotificationTarget
// @Failure      403  {object}  map[string]string
// @Router       /notifications/admin/{id}/targets [get]
func (h *NotificationHandler) GetTargets(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	targets, err := h.service.GetTargetsAdmin(c.UserContext(), admin, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotificationPermDenied) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(targets)
}

// =========================================================================
// AGREMIDO
// =========================================================================

// GetMyNotificationsPsi godoc
// @Summary      Listar notificaciones del agremiado
// @Description  Lista las notificaciones del psicólogo autenticado, paginadas por fecha desc.
// @Security     BearerAuth
// @Tags         Agremiado - Notificaciones
// @Produce      json
// @Param        page   query  int  false  "Número de página (Default: 1)"
// @Param        limit  query  int  false  "Items por página (Default: 10)"
// @Success      200    {object}  map[string]interface{}
// @Router       /notifications/psi-user [get]
func (h *NotificationHandler) GetMyNotificationsPsi(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	page, limit := parsePagination(c)
	res, err := h.service.ListUserNotifications(c.UserContext(), psi, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// GetUnreadCount godoc
// @Summary      Contador de notificaciones no leídas
// @Description  Retorna el número de notificaciones no leídas del agremiado (para badge).
// @Security     BearerAuth
// @Tags         Agremiado - Notificaciones
// @Produce      json
// @Success      200  {object}  map[string]int64
// @Router       /notifications/psi-user/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	count, err := h.service.GetUnreadCount(c.UserContext(), psi)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"unread_count": count})
}

// GetNotificationById godoc
// @Summary      Detalle de notificación (agremiado)
// @Description  Retorna el detalle y marca la notificación como leída automáticamente.
// @Security     BearerAuth
// @Tags         Agremiado - Notificaciones
// @Produce      json
// @Param        id  path  string  true  "UUID de la notificación"
// @Success      200  {object}  domain.Notification
// @Failure      404  {object}  map[string]string
// @Router       /notifications/psi-user/{id} [get]
func (h *NotificationHandler) GetNotificationById(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	notification, err := h.service.GetUserNotificationById(c.UserContext(), psi, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotificationNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, domain.ErrNotificationTargetNotOwned) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(notification)
}

// GetNotificationImage godoc
// @Summary      Descargar adjunto de notificación
// @Description  Retorna la URL pública de un adjunto si el agremiado es destinatario.
// @Security     BearerAuth
// @Tags         Agremiado - Notificaciones
// @Produce      json
// @Param        id        path  string  true  "UUID de la notificación"
// @Param        attachId  path  string  true  "UUID del adjunto"
// @Success      200  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /notifications/psi-user/{id}/attach/{attachId} [get]
func (h *NotificationHandler) GetNotificationImage(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}
	attachID, err := uuid.Parse(c.Params("attachId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de adjunto inválido"})
	}

	url, err := h.service.GetAttachmentURL(c.UserContext(), psi, id, attachID)
	if err != nil {
		if errors.Is(err, domain.ErrNotificationTargetNotOwned) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, domain.ErrAttachmentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"url": url})
}

// parsePagination extrae y valida params de paginación (page/limit).
func parsePagination(c *fiber.Ctx) (int, int) {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

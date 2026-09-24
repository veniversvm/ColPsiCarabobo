// api/internal/handler/audit_handler.go

// AuditHandler expone la bitácora de cambios (audit logs) al panel admin.
// Gates: cualquier endpoint requiere admin autenticado (ProtectedAdmin404 de la
// ruta) + `sudo || can_view_logs` (o can_export_logs para el CSV). Sin permiso
// se responde 404 enmascarado (mismo contrato que el resto del panel).
package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// AuditHandler gestiona la consulta de la bitácora de cambios.
type AuditHandler struct {
	svc *service.AuditService
}

// NewAuditHandler construye el handler con el AuditService inyectado.
func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// canViewLogs es el gate de lectura de la bitácora.
func (h *AuditHandler) canViewLogs(admin *domain.UserAdmin) bool {
	return admin != nil && (admin.Sudo || admin.CanViewLogs)
}

// canExportLogs es el gate de exportación (CSV).
func (h *AuditHandler) canExportLogs(admin *domain.UserAdmin) bool {
	return admin != nil && (admin.Sudo || admin.CanExportLogs)
}

// masked404 responde el 404 enmascarado del panel (no revela la existencia).
func masked404(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
	})
}

// parseAuditFilters construye los filtros desde la query string:
// q (texto libre sobre entity_label), suceso (action), entidad (entity),
// actor_id, desde/hasta (fechas ISO), page y limit.
func parseAuditFilters(c *fiber.Ctx) (domain.AuditLogFilters, error) {
	f := domain.AuditLogFilters{
		Q:      strings.TrimSpace(c.Query("q")),
		Action: strings.TrimSpace(c.Query("suceso")),
		Entity: strings.TrimSpace(c.Query("entidad")),
	}

	if v := strings.TrimSpace(c.Query("actor_id")); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return f, errors.New("actor_id inválido")
		}
		f.ActorID = &id
	}

	desde, err := parseAuditDate(c.Query("desde"), false)
	if err != nil {
		return f, err
	}
	hasta, err := parseAuditDate(c.Query("hasta"), true)
	if err != nil {
		return f, err
	}
	f.From = desde
	f.To = hasta

	if p, err := strconv.Atoi(c.Query("page", "1")); err == nil {
		f.Page = p
	} else {
		f.Page = 1
	}
	if l, err := strconv.Atoi(c.Query("limit", "20")); err == nil {
		f.Limit = l
	} else {
		f.Limit = 20
	}
	return f, nil
}

// parseAuditDate interpreta fechas en formato YYYY-MM-DD o RFC3339.
// Si `endOfDay` es true y el formato es solo fecha, extiende al final del día
// para que el filtro "hasta" incluya el día completo.
func parseAuditDate(raw string, endOfDay bool) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, raw); err == nil {
			if endOfDay && layout == "2006-01-02" {
				t = t.Add(24 * time.Hour)
			}
			return &t, nil
		}
	}
	return nil, errors.New("fecha inválida (use YYYY-MM-DD)")
}

// adminOr404 recupera el admin autenticado; si el gate no cumple, responde 404.
func (h *AuditHandler) adminOr404(c *fiber.Ctx, gate func(*domain.UserAdmin) bool) (*domain.UserAdmin, error) {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil || !gate(admin) {
		return nil, errMasked404
	}
	return admin, nil
}

// errMasked404 marca el camino de 404 enmascarado.
var errMasked404 = errors.New("not permitted")

// ListLogs godoc
// @Summary      Buscar en la bitácora de cambios
// @Description  Lista paginada de eventos con filtros combinables (q, suceso, entidad, actor_id, desde, hasta).
// @Tags         Administración - Auditoría
// @Produce      json
// @Param        q        query string false "Texto libre (entity_label)"
// @Param        suceso   query string false "Acción (create, update, delete, login, ...)"
// @Param        entidad  query string false "Entidad (psi, staff, notificacion, ...)"
// @Param        actor_id query string false "UUID del actor"
// @Param        desde    query string false "Fecha desde (YYYY-MM-DD)"
// @Param        hasta    query string false "Fecha hasta (YYYY-MM-DD)"
// @Param        page     query int    false "Página (default 1)"
// @Param        limit    query int    false "Por página (default 20, máx 100)"
// @Success      200      {object} map[string]interface{}
// @Failure      404      {object} map[string]string
// @Router       /admin/audit-logs [get]
func (h *AuditHandler) ListLogs(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil || !h.canViewLogs(admin) {
		return masked404(c)
	}

	f, perr := parseAuditFilters(c)
	if perr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": perr.Error()})
	}

	logs, total, err := h.svc.List(c.UserContext(), f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudieron consultar los registros de auditoría"})
	}
	return c.JSON(fiber.Map{
		"data":  logs,
		"total": total,
		"page":  f.Page,
		"limit": f.Limit,
	})
}

// ListPsiLogs godoc
// @Summary      Bitácora de cambios de un psicólogo
// @Description  Historial completo de lo que le pasó a un agremiado (quién y qué cambió).
// @Tags         Administración - Auditoría
// @Produce      json
// @Param        id   path string true "UUID del psicólogo"
// @Param        page query int    false "Página (default 1)"
// @Param        limit query int   false "Por página (default 20, máx 100)"
// @Success      200  {object} map[string]interface{}
// @Failure      404  {object} map[string]string
// @Router       /admin/audit-logs/psi/{id} [get]
func (h *AuditHandler) ListPsiLogs(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil || !h.canViewLogs(admin) {
		return masked404(c)
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return masked404(c)
	}

	f, perr := parseAuditFilters(c)
	if perr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": perr.Error()})
	}
	f.Entity = domain.AuditEntityPsi
	f.EntityID = targetID.String()

	logs, total, err := h.svc.List(c.UserContext(), f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudieron consultar los registros de auditoría"})
	}
	return c.JSON(fiber.Map{
		"data":  logs,
		"total": total,
		"page":  f.Page,
		"limit": f.Limit,
	})
}

// ExportLogs godoc
// @Summary      Exportar bitácora (CSV)
// @Description  Descarga en CSV los eventos que coinciden con los filtros (requiere can_export_logs o sudo).
// @Tags         Administración - Auditoría
// @Produce      text/csv
// @Param        q        query string false "Texto libre"
// @Param        suceso   query string false "Acción"
// @Param        entidad  query string false "Entidad"
// @Param        actor_id query string false "UUID del actor"
// @Param        desde    query string false "Fecha desde (YYYY-MM-DD)"
// @Param        hasta    query string false "Fecha hasta (YYYY-MM-DD)"
// @Success      200      {file} binary
// @Failure      404      {object} map[string]string
// @Router       /admin/audit-logs/export [get]
func (h *AuditHandler) ExportLogs(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil || !h.canExportLogs(admin) {
		return masked404(c)
	}

	f, perr := parseAuditFilters(c)
	if perr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": perr.Error()})
	}
	// Exportar sin paginación (hasta 10.000 filas).
	f.Page = 1
	f.Limit = 10000

	logs, _, err := h.svc.List(c.UserContext(), f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudieron exportar los registros de auditoría"})
	}

	c.Response().Header.Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header.Set("Content-Disposition", "attachment; filename=audit-logs.csv")

	w := csv.NewWriter(c)
	_ = w.Write([]string{"fecha", "actor", "rol", "entidad", "id_entidad", "etiqueta", "suceso", "ip", "cambios"})
	for _, l := range logs {
		_ = w.Write([]string{
			l.CreatedAt.UTC().Format(time.RFC3339),
			l.ActorUsername,
			l.ActorRole,
			l.Entity,
			l.EntityID,
			l.EntityLabel,
			l.Action,
			l.IP,
			string(l.Changes),
		})
	}
	w.Flush()
	return nil
}

// StatsLogs godoc
// @Summary      Estadísticas de la bitácora
// @Description  Agrega los sucesos por (entidad, acción) desde una fecha.
// @Tags         Administración - Auditoría
// @Produce      json
// @Param        desde query string false "Desde (YYYY-MM-DD); default 30 días atrás"
// @Success      200   {object} map[string]interface{}
// @Failure      404   {object} map[string]string
// @Router       /admin/audit-logs/stats [get]
func (h *AuditHandler) StatsLogs(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil || !h.canViewLogs(admin) {
		return masked404(c)
	}

	since := time.Now().AddDate(0, 0, -30).UTC()
	if t, perr := parseAuditDate(c.Query("desde"), false); perr == nil && t != nil {
		since = *t
	}

	stats, err := h.svc.Stats(c.UserContext(), since)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudieron calcular las estadísticas de auditoría"})
	}
	return c.JSON(fiber.Map{
		"since": since.Format(time.RFC3339),
		"stats": stats,
	})
}
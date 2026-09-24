// api/internal/router/audit_router.go

// SetupAuditRoutes registra los endpoints de lectura de la bitácora de cambios.
// Todos corren bajo /admin/audit-logs con ProtectedAdmin404 + NoStore; los gates
// de permiso (can_view_logs / can_export_logs) viven en el handler.
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// SetupAuditRoutes inicializa el handler y registra las rutas de auditoría.
func SetupAuditRoutes(router fiber.Router, adminRepo domain.UserAdminRepository, psiRepo domain.PsiUserRepository, auditSvc *service.AuditService, analyticsSvc *service.AnalyticsService) {
	h := handler.NewAuditHandler(auditSvc)
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	admin := router.Group("/admin/audit-logs", middleware.NoStore(), authMid.ProtectedAdmin404())

	admin.Get("/", h.ListLogs)           // Búsqueda con filtros (q, suceso, entidad, actor_id, fechas)
	admin.Get("/psi/:id", h.ListPsiLogs) // Historial completo de un psicólogo
	admin.Get("/export", h.ExportLogs)   // CSV (can_export_logs)
	admin.Get("/stats", h.StatsLogs)     // Agregados por entidad/acción
}
// Package router organiza los puntos de entrada de la API segmentados por dominios de negocio.
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// SetupKanbanRoutes registra los endpoints del módulo de Proyectos (tableros Kanban).
// Todas las rutas viven bajo /admin/projects protegidas con ProtectedAdmin404
// (Security by Obscurity) y NoStore. Los permisos de proyecto (master/owner/
// editor/viewer) se resuelven en la capa de servicio.
func SetupKanbanRoutes(router fiber.Router, adminRepo domain.UserAdminRepository, psiRepo domain.PsiUserRepository, kanbanRepo domain.KanbanRepository, analyticsSvc *service.AnalyticsService) {
	svc := service.NewKanbanService(kanbanRepo, adminRepo)
	h := handler.NewKanbanHandler(svc)
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// =========================================================================
	// GRUPO ADMINISTRATIVO (ALTA PRIORIDAD)
	// =========================================================================
	admin := router.Group("/admin/projects", middleware.NoStore(), authMid.ProtectedAdmin404())

	// Proyectos
	admin.Get("/", h.ListProjects)
	admin.Post("/", h.CreateProject)
	admin.Get("/:id", h.GetProject)
	admin.Patch("/:id", h.UpdateProject)
	admin.Delete("/:id", h.DeleteProject)

	// Miembros (rutas planas para evitar colisiones con :id)
	admin.Get("/:id/members", h.ListMembers)
	admin.Post("/:id/members", h.AddMember)
	admin.Patch("/members/:memberId", h.UpdateMember)
	admin.Delete("/members/:memberId", h.RemoveMember)

	// Columnas
	admin.Post("/:id/columns", h.CreateColumn)
	admin.Patch("/columns/:columnId", h.UpdateColumn)
	admin.Delete("/columns/:columnId", h.DeleteColumn)

	// Tarjetas
	admin.Post("/:id/cards", h.CreateCard)
	admin.Patch("/cards/:cardId", h.UpdateCard)
	admin.Delete("/cards/:cardId", h.DeleteCard)

	// Notas
	admin.Post("/cards/:cardId/notes", h.CreateNote)
	admin.Delete("/notes/:noteId", h.DeleteNote)
}

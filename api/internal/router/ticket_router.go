// Package router organiza los puntos de entrada de la API segmentados por dominios de negocio.
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// SetupTicketRoutes registra los endpoints del módulo de Tickets de Solicitudes.
//
// Portal psicólogo (/psi/tickets): crear, listar, ver, comentar y cerrar sus
// tickets, además de leer la configuración (motivos → estados).
//
// Panel administrativo (/admin/tickets): cola FIFO de tickets abiertos con
// filtros, detalle, respuesta, cambio de estado, cierre y conteo de pendientes
// (badge). La configuración (motivos/estados) vive también aquí y se protege
// con el permiso granular CanManageTickets + NoStore.
func SetupTicketRoutes(router fiber.Router, adminRepo domain.UserAdminRepository, psiRepo domain.PsiUserRepository, s3Client *s3.S3Client, ticketRepo domain.TicketRepository, ticketConfigRepo domain.TicketConfigRepository, notificationSvc *service.NotificationService, analyticsSvc *service.AnalyticsService) {
	svc := service.NewTicketService(ticketRepo, ticketConfigRepo, s3Client, notificationSvc)
	h := handler.NewTicketHandler(svc)
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// =========================================================================
	// GRUPO ADMINISTRATIVO (configuración y cola de tickets)
	// =========================================================================
	admin := router.Group("/admin/tickets", middleware.NoStore(), authMid.ProtectedAdmin404())

	// Configuración — rutas estáticas ANTES de :id para evitar el match dinámico
	admin.Get("/motivos", h.ListMotivos)
	admin.Post("/motivos", h.CreateMotivo)
	admin.Patch("/motivos/:id", h.UpdateMotivo)
	admin.Delete("/motivos/:id", h.DeleteMotivo)
	admin.Get("/motivos/:id/estados", h.ListEstadosConfig)

	admin.Post("/estados", h.CreateEstado)
	admin.Patch("/estados/:id", h.UpdateEstadoConfig)
	admin.Delete("/estados/:id", h.DeleteEstadoConfig)

	// Cola administrativa
	admin.Get("/pendientes-count", h.CountPendientesAdmin)
	admin.Get("/", h.ListTicketsAdmin)
	admin.Get("/:id", h.GetTicketAdmin)
	admin.Patch("/:id/estado", h.UpdateTicketEstado)
	admin.Post("/:id/mensaje", h.AddMensajeAdmin)
	admin.Post("/:id/cerrar", h.CloseTicketAdmin)

	// =========================================================================
	// PORTAL PSICÓLOGO
	// =========================================================================
	psi := router.Group("/psi/tickets", middleware.NoStore(), authMid.ProtectedPsiUser())

	psi.Get("/config", h.GetConfigPSI)
	psi.Get("/", h.ListMyTickets)
	psi.Post("/", h.CreateTicket)
	psi.Get("/:id", h.GetTicketAsPsi)
	psi.Post("/:id/mensaje", h.AddMensajePsi)
	psi.Post("/:id/cerrar", h.CloseTicketPsi)
}
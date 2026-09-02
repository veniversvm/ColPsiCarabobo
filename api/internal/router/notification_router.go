// api/internal/router/notification_router.go
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// SetupNotificationRoutes registra las rutas de notificaciones (admin y agremiado).
func SetupNotificationRoutes(
	router fiber.Router,
	adminRepo domain.UserAdminRepository,
	psiRepo domain.PsiUserRepository,
	analyticsSvc *service.AnalyticsService,
	notificationSvc *service.NotificationService,
) {
	h := handler.NewNotificationHandler(notificationSvc)
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// =========================================================================
	// ADMIN — protegidos con ProtectedAdmin404()
	// =========================================================================
	admin := router.Group("/notifications/admin", authMid.ProtectedAdmin404())

	admin.Post("/preview", h.PreviewRecipients)
	admin.Post("/", h.CreateNotification)
	admin.Post("/:id/attach", h.AttachFile)
	admin.Get("/", h.GetMyNotifications)
	admin.Get("/:id", h.GetNotificationDetail)
	admin.Delete("/:id", h.CancelNotification)
	admin.Get("/:id/targets", h.GetTargets)

	// Nota: la ruta /:id GET y /:id/targets GET podrían colisionar en algunos
	// routers; Fiber las resuelve en orden de registro. Colocamos /:id/targets
	// después de /:id para que el matcher más específico tenga prioridad.

	// =========================================================================
	// AGREMIDO — protegidos con ProtectedPsiUser()
	// =========================================================================
	psiUser := router.Group("/notifications/psi-user", authMid.ProtectedPsiUser())

	psiUser.Get("/", h.GetMyNotificationsPsi)
	psiUser.Get("/unread-count", h.GetUnreadCount)
	psiUser.Get("/:id", h.GetNotificationById)
	psiUser.Get("/:id/attach/:attachId", h.GetNotificationImage)
}

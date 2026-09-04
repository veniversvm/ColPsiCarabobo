// api/internal/router/inscription_router.go
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// SetupInscriptionRoutes registra las rutas públicas y administrativas del
// módulo de pre-inscripción de profesionales.
func SetupInscriptionRoutes(router fiber.Router, inscriptionRepo domain.InscriptionRepository, psiRepo domain.PsiUserRepository, adminRepo domain.UserAdminRepository, s3Client *s3.S3Client, mailService service.IMailService, analyticsSvc *service.AnalyticsService) {
	svc := service.NewInscriptionService(inscriptionRepo, psiRepo, s3Client, mailService)
	h := handler.NewInscriptionHandler(svc)
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// =========================================================================
	// ZONA 1: PÚBLICO (sin auth)
	// =========================================================================
	publicGroup := router.Group("/inscripcion", middleware.NoStore())

	publicGroup.Get("/check-ci", h.CheckCI)
	publicGroup.Get("/check-fpv", h.CheckFPV)
	publicGroup.Post("/submit", h.Submit)

	// =========================================================================
	// ZONA 2: ADMIN
	// =========================================================================
	adminGroup := router.Group("/admin/inscripciones", middleware.NoStore(), authMid.ProtectedAdmin404())

	adminGroup.Get("/list", h.List)
	adminGroup.Get("/:id<uuid>", h.Detail)
	adminGroup.Post("/:id<uuid>/approve", h.Approve)
	adminGroup.Delete("/:id<uuid>", h.Reject)
}

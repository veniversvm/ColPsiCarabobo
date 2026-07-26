// api/internal/router/psi.go
package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// SetupPsiRoutes registers psychologist public, self-management, and admin CRUD routes.
func SetupPsiRoutes(router fiber.Router, psiRepo domain.PsiUserRepository, adminRepo domain.UserAdminRepository, s3Client *s3.S3Client, analyticsSvc *service.AnalyticsService, mailService *service.MailService) {
	svc := service.NewPsiService(psiRepo, s3Client, mailService)
	h := handler.NewPsiHandler(svc, analyticsSvc)
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// Store de idempotencia — compartido solo para las rutas que lo necesitan
	idempotencyStore := middleware.NewIdempotencyStore()

	// =========================================================================
	// ZONA 1: ADMINISTRACIÓN
	// =========================================================================
	adminGroup := router.Group("/admin/psi", authMid.ProtectedAdmin404())

	adminGroup.Get("/list", h.ListAllPsis)

	// Idempotencia en crear — evita duplicados por doble click o reintento
	adminGroup.Post(
		"/create",
		middleware.UserScopedIdempotency(idempotencyStore, 30*time.Minute),
		h.CreatePsiByAdmin,
	)

	adminGroup.Post("/upload-csv", h.UploadCsv)
	adminGroup.Get("/:id<uuid>", h.GetPsiByIDAdmin)
	adminGroup.Patch("/:id", h.UpdatePsiByAdmin)
	adminGroup.Delete("/:id", h.DeletePsiByAdmin)

	// =========================================================================
	// ZONA 2: AUTOGESTIÓN
	// =========================================================================
	meGroup := router.Group("/psi/me", authMid.ProtectedPsiUser())

	meGroup.Get("/", h.GetMe)
	meGroup.Patch("/", h.UpdateOwnProfile)
	meGroup.Post("/postgrades", h.AddPostGrade)
	meGroup.Patch("/postgrades/:id", h.UpdatePostGrade)
	meGroup.Post("/social", h.AddSocialNetwork)
	meGroup.Patch("/social/:id", h.UpdateSocialNetwork)
	meGroup.Delete("/social/:id", h.DeleteSocialNetwork)
	meGroup.Post("/logout", h.Logout)

	// =========================================================================
	// ZONA 3: PÚBLICO
	// =========================================================================
	psiGroup := router.Group("/psi")

	// No necesita token de admin porque es para el sitemap público
	psiGroup.Get("/public/sitemap-data", h.GetSitemapData)

	// Login con rate limiting — 10 intentos por IP cada 15 minutos
	psiGroup.Post("/login", middleware.AuthRateLimiter(), h.Login)
	psiGroup.Post("/login-library", middleware.AuthRateLimiter(), h.LoginLibrary)
	psiGroup.Get("/directory", h.SearchDirectory)
	psiGroup.Get("/:id", h.GetPublicProfile)
}

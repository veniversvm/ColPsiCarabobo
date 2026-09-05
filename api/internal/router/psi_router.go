// api/internal/router/psi.go
package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/cache"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// SetupPsiRoutes registers psychologist public, self-management, and admin CRUD routes.
func SetupPsiRoutes(router fiber.Router, psiRepo domain.PsiUserRepository, adminRepo domain.UserAdminRepository, s3Client *s3.S3Client, analyticsSvc *service.AnalyticsService, mailService service.IMailService, appCache *cache.Cache) {
	svc := service.NewPsiService(psiRepo, s3Client, mailService)
	h := handler.NewPsiHandler(svc, analyticsSvc, appCache)
	absSvc := service.NewAudiobookshelfService(
		config.Envs.AbsBaseURL,
		config.Envs.AbsPublicURL,
		config.Envs.AbsAdminUsername,
		config.Envs.AbsAdminPassword,
		config.Envs.AbsPasswordSecret,
	)
	svc.SetAudiobookshelf(absSvc)
	h.SetAudiobookshelf(absSvc)
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// Store de idempotencia — compartido solo para las rutas que lo necesitan
	idempotencyStore := middleware.NewIdempotencyStore()

	// =========================================================================
	// ZONA 1: ADMINISTRACIÓN
	// =========================================================================
	adminGroup := router.Group("/admin/psi", middleware.NoStore(), authMid.ProtectedAdmin404())

	adminGroup.Get("/list", h.ListAllPsis)

	// Aviso de cumpleaños al admin (opt-in del psicólogo)
	adminGroup.Get("/birthdays", h.GetBirthdaysByAdmin)

	// Idempotencia en crear — evita duplicados por doble click o reintento
	adminGroup.Post(
		"/create",
		middleware.UserScopedIdempotency(idempotencyStore, 30*time.Minute),
		h.CreatePsiByAdmin,
	)

	adminGroup.Post("/upload-csv", h.UploadCsv)
	adminGroup.Get("/:id<uuid>", h.GetPsiByIDAdmin)
	adminGroup.Delete("/:id/picture", h.DeleteProfilePictureByAdmin)
	adminGroup.Patch("/:id", h.UpdatePsiByAdmin)
	adminGroup.Delete("/:id", h.DeletePsiByAdmin)
	adminGroup.Post("/:id/reset-password", h.ResetPsiPasswordByAdmin)

	// Expediente deontológico (acceso exclusivo admin)
	adminGroup.Get("/:id<uuid>/deontologia", h.ListDeontologiaByAdmin)
	adminGroup.Post("/:id<uuid>/deontologia", h.AddDeontologiaByAdmin)
	adminGroup.Patch("/:id<uuid>/deontologia/:entryId<uuid>", h.UpdateDeontologiaByAdmin)

	// Observaciones internas (acceso exclusivo admin)
	adminGroup.Get("/:id<uuid>/observaciones", h.ListObservacionesByAdmin)
	adminGroup.Post("/:id<uuid>/observaciones", h.AddObservacionesByAdmin)
	adminGroup.Patch("/:id<uuid>/observaciones/:entryId<uuid>", h.UpdateObservacionesByAdmin)

	// Registro digital de documentos (gestión exclusiva admin; el psi solo lee)
	adminGroup.Get("/:id<uuid>/documents", h.ListDocumentsByAdmin)
	adminGroup.Post("/:id<uuid>/documents", h.AddDocumentByAdmin)
	adminGroup.Patch("/:id<uuid>/documents/:docId<uuid>", h.UpdateDocumentByAdmin)
	adminGroup.Delete("/:id<uuid>/documents/:docId<uuid>", h.DeleteDocumentByAdmin)

	// =========================================================================
	// ZONA 2: AUTOGESTIÓN
	// =========================================================================
	meGroup := router.Group("/psi/me", middleware.NoStore(), authMid.ProtectedPsiUser())

	meGroup.Get("/", h.GetMe)
	meGroup.Get("/audiobookshelf", h.GetAudiobookshelfAccess)
	meGroup.Patch("/", h.UpdateOwnProfile)
	meGroup.Post("/postgrades", h.AddPostGrade)
	meGroup.Patch("/postgrades/:id", h.UpdatePostGrade)
	meGroup.Post("/social", h.AddSocialNetwork)
	meGroup.Patch("/social/:id", h.UpdateSocialNetwork)
	meGroup.Delete("/social/:id", h.DeleteSocialNetwork)
	meGroup.Post("/logout", h.Logout)
	meGroup.Get("/validate", h.ValidateSession)

	// Mis documentos digitales — SOLO LECTURA (el psicólogo no puede editarlos)
	meGroup.Get("/documents", h.GetMyDocuments)

	// =========================================================================
	// ZONA 3: PÚBLICO
	// =========================================================================
	psiGroup := router.Group("/psi")

	// No necesita token de admin porque es para el sitemap público
	psiGroup.Get("/public/sitemap-data", h.GetSitemapData)

	// Login con rate limiting — 10 intentos por IP cada 15 minutos
	psiGroup.Post("/login", middleware.NoStore(), middleware.AuthRateLimiter(), h.Login)
	psiGroup.Post("/login-library", middleware.NoStore(), middleware.AuthRateLimiter(), h.LoginLibrary)
	psiGroup.Get("/directory", h.SearchDirectory)
	psiGroup.Get("/:id", h.GetPublicProfile)
}

// Package router organiza los puntos de entrada de la API segmentados por dominios de negocio.
package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

// SetupPostRoutes inicializa las dependencias y registra los endpoints del módulo de noticias y publicaciones.
// Implementa un sistema de visibilidad dual: público para visitantes y extendido para miembros autenticados.
func SetupPostRoutes(router fiber.Router, db *gorm.DB, s3Client *s3.S3Client, analyticsSvc *service.AnalyticsService) {
	adminRepo := postgres.NewAdminRepository(db)
	psiRepo := postgres.NewPsiRepository(db)
	postRepo := postgres.NewPostRepository(db)

	svc := service.NewPostService(postRepo, s3Client)
	h := handler.NewPostHandler(svc)

	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// Store propio — independiente del de psi.go para no mezclar dominios
	idempotencyStore := middleware.NewIdempotencyStore()

	// =========================================================================
	// GRUPO DE CONSULTA (ACCESO HÍBRIDO)
	// =========================================================================
	posts := router.Group("/posts")

	posts.Get("/", authMid.OptionalHybridAuth(), h.ListPosts)
	posts.Get("/public/sitemap-posts", h.GetSiteMapHandler)
	posts.Get("/:id", authMid.OptionalHybridAuth(), h.GetPost)

	// =========================================================================
	// GRUPO ADMINISTRATIVO (ALTA SEGURIDAD)
	// =========================================================================
	admin := router.Group("/admin/posts", authMid.ProtectedAdmin404())

	// Idempotencia en crear — evita publicaciones duplicadas por doble click o reintento
	admin.Post(
		"/",
		middleware.UserScopedIdempotency(idempotencyStore, 30*time.Minute),
		h.CreatePost,
	)

	admin.Patch("/:id", h.UpdatePost)

	// admin.Delete("/:id", h.DeletePost)
}

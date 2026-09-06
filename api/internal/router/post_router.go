// Package router organiza los puntos de entrada de la API segmentados por dominios de negocio.
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

// SetupPostRoutes inicializa las dependencias y registra los endpoints del módulo de noticias y publicaciones.
// Implementa un sistema de visibilidad dual: público para visitantes y extendido para miembros autenticados.
func SetupPostRoutes(router fiber.Router, adminRepo domain.UserAdminRepository, psiRepo domain.PsiUserRepository, postRepo domain.PostRepository, s3Client *s3.S3Client, analyticsSvc *service.AnalyticsService) {
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
}

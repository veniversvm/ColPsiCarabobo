package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

func SetupPostRoutes(router fiber.Router, db *gorm.DB, s3Client *s3.S3Client) {
	repo := postgres.NewAdminRepository(db)
	psiRepo := postgres.NewPsiRepository(db)
	postRepo := postgres.NewPostRepository(db)
	svc := service.NewPostService(postRepo, s3Client)
	h := handler.NewPostHandler(svc)

	// Middleware que intenta leer token pero no falla si no hay
	// Necesita soportar tanto Admin como PsiUser
	authMid := middleware.NewAuthMiddleware(repo, psiRepo)

	posts := router.Group("/posts")

	// GET /posts -> Middleware "Híbrido" que detecta si es Admin, Psi o Nadie
	// GET /posts -> Middleware "Híbrido"
	// Si envía token válido -> Ve contenido privado.
	// Si no envía token -> Ve contenido público.
	posts.Get("/", authMid.OptionalHybridAuth(), h.ListPosts)
	posts.Get("/:id", authMid.OptionalHybridAuth(), h.GetPost)

	// Admin Only
	admin := router.Group("/admin/posts", authMid.ProtectedAdmin404())
	admin.Post("/", h.CreatePost)
	admin.Patch("/:id", h.UpdatePost)
	// admin.Delete("/:id", h.DeletePost)
}

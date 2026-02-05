package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware" // Importante para el auth
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

func SetupPsiRoutes(router fiber.Router, db *gorm.DB, s3Client *s3.S3Client) {
	repo := postgres.NewPsiRepository(db)
	adminRepo := postgres.NewAdminRepository(db)

	svc := service.NewPsiService(repo, s3Client)
	h := handler.NewPsiHandler(svc)

	// Instanciar Middleware de protección
	authMid := middleware.NewAuthMiddleware(adminRepo, repo)

	psiGroup := router.Group("/psi")

	// Usar el middleware de "obscuridad" 404
	psiGroup.Post("/upload-csv", authMid.ProtectedAdmin404(), h.UploadCsv)
}

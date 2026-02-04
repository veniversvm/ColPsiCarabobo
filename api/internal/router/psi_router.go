package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

// SetupPsiRoutes inyecta las dependencias necesarias para el dominio de Psicólogos
func SetupPsiRoutes(router fiber.Router, db *gorm.DB, s3Client *s3.S3Client) {
	repo := postgres.NewPsiRepository(db)
	svc := service.NewPsiService(repo, s3Client)

	// Ahora esto funcionará porque NewPsiHandler retorna un puntero
	h := handler.NewPsiHandler(svc)

	psiGroup := router.Group("/psi")

	psiGroup.Post("/upload-csv", h.UploadCsv)

}

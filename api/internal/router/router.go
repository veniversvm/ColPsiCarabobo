package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

// SetupRouter centraliza la configuración de todos los dominios
func SetupRouter(app *fiber.App, db *gorm.DB, s3Client *s3.S3Client) {
	// Middleware global para esta rama de rutas
	app.Use(logger.New())

	// Agrupación principal de la API
	api := app.Group("/api/v1")

	// Registro de sub-routers por dominio
	SetupPsiRoutes(api, db, s3Client)
	// SetupAdminRoutes(api, db) // Se irá agregando según avancemos
	SetupAdminRoutes(api, db)
	// SetupPostRoutes(api, db, s3Client)
}

package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

func SetupRouter(app *fiber.App, db *gorm.DB, s3Client *s3.S3Client) {
	api := app.Group("/api/v1")

	// Llamamos a las funciones con sus nombres correctos
	SetupAdminRoutes(api, db)
	SetupPsiRoutes(api, db, s3Client)
}

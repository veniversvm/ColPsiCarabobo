package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

func SetupRouter(app *fiber.App, db *gorm.DB, s3Client *s3.S3Client) {
	api := app.Group("/api/v1")

	// Llamamos a las funciones con sus nombres correctos
	SetupAdminRoutes(api, db)
	SetupPsiRoutes(api, db, s3Client)

	// =========================================================================
	// 2. DEFAULT 404 HANDLER (CATCH-ALL)
	// =========================================================================
	// IMPORTANTE: Este bloque debe ir al final de SetupRouter o en el main
	// después de registrar todos los dominios.
	app.Use(func(c *fiber.Ctx) error {
		// Retornamos el mismo formato que usamos en el middleware de auth
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})
}

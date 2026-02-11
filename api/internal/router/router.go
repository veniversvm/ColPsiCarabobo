package router

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"

	// [!] CRÍTICO: Este import carga los archivos JSON generados por 'swag init'
	// Asegúrate de que la ruta coincida con el nombre de tu módulo en go.mod
	_ "github.com/veniversvm/ColPsiCarabobo/api/docs"
)

func SetupRouter(app *fiber.App, db *gorm.DB, s3Client *s3.S3Client) {
	// Agrupación principal
	api := app.Group("/api/v1")

	// OpenAPI: Documentación Swagger (Ahora dentro de /api/v1)
	// URL: http://localhost:8080/api/v1/swagger/index.html
	if config.Envs.Environment == "development" {
		log.Println("=== OPEN API ===")
		api.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Rutas de dominio
	SetupAdminRoutes(api, db)
	SetupPsiRoutes(api, db, s3Client)

	// =========================================================================
	// 2. DEFAULT 404 HANDLER (CATCH-ALL)
	// =========================================================================
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})
}

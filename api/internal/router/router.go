package router

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"

	// [!] CRÍTICO: Este import carga los archivos JSON generados por 'swag init'
	// Asegúrate de que la ruta coincida con el nombre de tu módulo en go.mod
	_ "github.com/veniversvm/ColPsiCarabobo/api/docs"
)

func SetupRouter(app *fiber.App, db *gorm.DB, s3Client *s3.S3Client) {

	// ── Analytics: instanciar servicio y registrar middleware global ──────────
	// El middleware captura TODOS los GET exitosos antes de que lleguen a las rutas
	analyticsSvc := service.NewAnalyticsService(db)
	app.Use(middleware.AnalyticsMiddleware(db))

	// Agrupación principal
	api := app.Group("/api/v1")

	// OpenAPI: Documentación Swagger
	if config.Envs.Environment == "development" {
		log.Println("=== OPEN API ===")
		api.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Rutas de dominio — se pasa analyticsSvc solo a quienes lo necesitan
	SetupAdminRoutes(api, db, analyticsSvc)
	SetupPsiRoutes(api, db, s3Client, analyticsSvc)
	SetupSpecialtyRoutes(api, db, analyticsSvc)
	SetupPostRoutes(api, db, s3Client, analyticsSvc)

	// =========================================================================
	// DEFAULT 404 HANDLER (CATCH-ALL)
	// =========================================================================
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})
}

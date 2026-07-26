package router

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"

	// [!] CRÍTICO: Este import carga los archivos JSON generados por 'swag init'
	// Asegúrate de que la ruta coincida con el nombre de tu módulo en go.mod
	_ "github.com/veniversvm/ColPsiCarabobo/api/docs"
)

func SetupRouter(app *fiber.App, db *gorm.DB, s3Client *s3.S3Client) {

	// ── Analytics: instanciar repo, servicio y registrar middleware global ────
	analyticsRepo := postgres.NewAnalyticsRepository(db)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)
	app.Use(middleware.AnalyticsMiddleware(analyticsSvc))

	// ── Repositories: instanciar una sola vez para todos los routers ─────────
	adminRepo := postgres.NewAdminRepository(db)
	psiRepo := postgres.NewPsiRepository(db)
	postRepo := postgres.NewPostRepository(db)
	specialtyRepo := postgres.NewSpecialtyRepository(db)

	// ── MailService: una sola instancia compartida entre todos los routers ────
	mailSvc, err := service.NewMailService()
	if err != nil {
		log.Printf("[WARN] Advertencia: No se pudo conectar al servidor SMTP: %v", err)
	}

	// Agrupación principal
	api := app.Group("/api/v1")

	// OpenAPI: Documentación Swagger
	if config.Envs.Environment == "development" {
		log.Println("=== OPEN API ===")
		api.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Rutas de dominio — se pasan repos, analyticsSvc y mailSvc
	SetupAdminRoutes(api, adminRepo, psiRepo, analyticsSvc, mailSvc)
	SetupPsiRoutes(api, psiRepo, adminRepo, s3Client, analyticsSvc, mailSvc)
	SetupSpecialtyRoutes(api, psiRepo, adminRepo, specialtyRepo, analyticsSvc)
	SetupPostRoutes(api, adminRepo, psiRepo, postRepo, s3Client, analyticsSvc)

	// =========================================================================
	// DEFAULT 404 HANDLER (CATCH-ALL)
	// =========================================================================
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})
}

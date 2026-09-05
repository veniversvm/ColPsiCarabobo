// Package router configures all HTTP routes for the ColPsiCarabobo API.
package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/cache"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"

	// [!] CRÍTICO: Este import carga los archivos JSON generados por 'swag init'
	// Asegúrate de que la ruta coincida con el nombre de tu módulo en go.mod
	_ "github.com/veniversvm/ColPsiCarabobo/api/docs"
)

// SetupRouter initializes all API routes, middleware, and dependency injection.
func SetupRouter(app *fiber.App, db *gorm.DB, s3Client *s3.S3Client, appCache *cache.Cache, mailSvc service.IMailService, notificationSvc *service.NotificationService) {

	// ── Analytics: instanciar repo, servicio y registrar middleware global ────
	analyticsRepo := postgres.NewAnalyticsRepository(db)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)
	app.Use(middleware.AnalyticsMiddleware(analyticsSvc))

	// ── Repositories: instanciar una sola vez para todos los routers ─────────
	adminRepo := postgres.NewAdminRepository(db)
	psiRepo := postgres.NewPsiRepository(db)
	postRepo := postgres.NewPostRepository(db)
	specialtyRepo := postgres.NewSpecialtyRepository(db)
	inscriptionRepo := postgres.NewInscriptionRepository(db)
	kanbanRepo := postgres.NewKanbanRepository(db)
	ticketRepo := postgres.NewTicketRepository(db)
	ticketConfigRepo := postgres.NewTicketConfigRepository(db)

	// Agrupación principal
	api := app.Group("/api/v1")

	// OpenAPI: Documentación Swagger
	if config.Envs.Environment == "development" {
		log.Info().Str("component", "router").Msg("=== OPEN API ===")
		api.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Rutas de dominio — se pasan repos, analyticsSvc y mailSvc
	// IMPORTANTE: SetupTicketRoutes va ANTES de SetupPsiRoutes. En gofiber, a
	// igual especificidad gana la ruta registrada primero; el `/psi/:id` público
	// (SetupPsiRoutes) se tragaría la raíz estática `/psi/tickets` del portal
	// psicólogo (ListMyTickets) si se registrara después.
	SetupAdminRoutes(api, adminRepo, psiRepo, analyticsSvc, mailSvc)
	SetupTicketRoutes(api, adminRepo, psiRepo, s3Client, ticketRepo, ticketConfigRepo, notificationSvc, analyticsSvc)
	SetupPsiRoutes(api, psiRepo, adminRepo, s3Client, analyticsSvc, mailSvc, appCache)
	SetupSpecialtyRoutes(api, psiRepo, adminRepo, specialtyRepo, analyticsSvc)
	SetupPostRoutes(api, adminRepo, psiRepo, postRepo, s3Client, analyticsSvc)
	SetupNotificationRoutes(api, adminRepo, psiRepo, analyticsSvc, notificationSvc)
	SetupInscriptionRoutes(api, inscriptionRepo, psiRepo, adminRepo, s3Client, mailSvc, analyticsSvc)
	SetupKanbanRoutes(api, adminRepo, psiRepo, kanbanRepo, analyticsSvc)
	SetupTicketRoutes(api, adminRepo, psiRepo, s3Client, ticketRepo, ticketConfigRepo, notificationSvc, analyticsSvc)

	// =========================================================================
	// DEFAULT 404 HANDLER (CATCH-ALL)
	// =========================================================================
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})
}

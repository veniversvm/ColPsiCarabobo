// api/internal/router/admin.go
package router

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"gorm.io/gorm"
)

func SetupAdminRoutes(router fiber.Router, db *gorm.DB, analyticsSvc *service.AnalyticsService) {
	repo := postgres.NewAdminRepository(db)
	psiRepo := postgres.NewPsiRepository(db)

	mailSvc, err := service.NewMailService()
	if err != nil {
		log.Printf("⚠️  Advertencia: No se pudo conectar al servidor SMTP: %v", err)
	}

	svc := service.NewAdminService(repo, mailSvc)
	h := handler.NewAdminHandler(svc)
	authMid := middleware.NewAuthMiddleware(repo, psiRepo, analyticsSvc)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)

	// =========================================================================
	// RUTAS DE DESARROLLO
	// =========================================================================
	router.Get("/debug-monitor", monitor.New(monitor.Config{Title: "DEV ONLY - Monitor"}))

	// =========================================================================
	// RUTAS PÚBLICAS
	// =========================================================================
	auth := router.Group("/auth")

	// Login de admin con rate limiting — 5 intentos por IP cada 30 minutos
	auth.Post("/login", middleware.AdminAuthRateLimiter(), h.Login)

	// =========================================================================
	// RUTAS PROTEGIDAS
	// =========================================================================
	admin := router.Group("/admin", authMid.ProtectedAdmin404())

	admin.Get("/metrics", monitor.New(monitor.Config{
		Title: "ColPsiCarabobo - Panel de Control Administrativo",
	}))
	admin.Get("/dashboard/stats", analyticsHandler.GetDashboardStats)

	admin.Post("/create", h.CreateAdmin)
	admin.Get("/list", h.GetAdmins)
	admin.Patch("/update", h.UpdateAdmin)
	admin.Delete("/delete/:id", h.DeleteAdmin)
}

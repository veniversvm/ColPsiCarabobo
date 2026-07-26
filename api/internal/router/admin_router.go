// api/internal/router/admin.go
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

func SetupAdminRoutes(router fiber.Router, adminRepo domain.UserAdminRepository, psiRepo domain.PsiUserRepository, analyticsSvc *service.AnalyticsService, mailSvc *service.MailService) {
	svc := service.NewAdminService(adminRepo, mailSvc)
	h := handler.NewAdminHandler(svc)
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)

	// =========================================================================
	// RUTAS DE DESARROLLO
	// =========================================================================
	if config.Envs.Environment == "development" {
		router.Get("/debug-monitor", monitor.New(monitor.Config{Title: "DEV ONLY - Monitor"}))
	}

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

	// Logout — requiere sesión activa
	admin.Post("/logout", h.Logout)

	admin.Get("/metrics", monitor.New(monitor.Config{
		Title: "ColPsiCarabobo - Panel de Control Administrativo",
	}))
	admin.Get("/dashboard/stats", analyticsHandler.GetDashboardStats)

	admin.Post("/create", h.CreateAdmin)
	admin.Get("/list", h.GetAdmins)
	admin.Patch("/update", h.UpdateAdmin)
	admin.Delete("/delete/:id", h.DeleteAdmin)
}

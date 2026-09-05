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

// SetupAdminRoutes registers admin authentication, management, and dashboard routes.
func SetupAdminRoutes(router fiber.Router, adminRepo domain.UserAdminRepository, psiRepo domain.PsiUserRepository, analyticsSvc *service.AnalyticsService, mailSvc service.IMailService) {
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
	auth := router.Group("/auth", middleware.NoStore())

	// Login de admin con rate limiting — 5 intentos por IP cada 30 minutos
	auth.Post("/login", middleware.AdminAuthRateLimiter(), h.Login)

	// =========================================================================
	// RUTAS PROTEGIDAS
	// =========================================================================
	admin := router.Group("/admin", middleware.NoStore(), authMid.ProtectedAdmin404())

	// Logout — requiere sesión activa
	admin.Post("/logout", h.Logout)

	admin.Get("/metrics", monitor.New(monitor.Config{
		Title: "ColPsiCarabobo - Panel de Control Administrativo",
	}))
	admin.Get("/dashboard/stats", analyticsHandler.GetDashboardStats)

	admin.Post("/create", h.CreateAdmin)
	admin.Get("/list", h.GetAdmins)
	admin.Get("/roles/presets", h.GetRolePresets)
	admin.Patch("/update", h.UpdateAdmin)
	admin.Delete("/delete/:id", h.DeleteAdmin)

	// =========================================================================
	// VALIDACIÓN DE SESIÓN (retorna 401 explícito, no 404 enmascarado)
	// =========================================================================
	// Grupo aparte con ProtectedAdmin() para que el frontend pueda distinguir
	// "sesión inválida" (401) de "ruta inexistente" (404).
	adminValidate := router.Group("/admin", middleware.NoStore(), authMid.ProtectedAdmin())
	adminValidate.Get("/validate", h.ValidateSession)
}

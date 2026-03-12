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

	// Handler de analytics (solo lectura del dashboard)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)

	// =========================================================================
	// RUTAS DE DESARROLLO
	// =========================================================================
	router.Get("/debug-monitor", monitor.New(monitor.Config{Title: "DEV ONLY - Monitor"}))

	// =========================================================================
	// RUTAS PÚBLICAS
	// =========================================================================
	auth := router.Group("/auth")
	auth.Post("/login", h.Login) // Login de admin — sin analytics de login por ahora

	// =========================================================================
	// RUTAS PROTEGIDAS (Admin Staff Only)
	// =========================================================================
	admin := router.Group("/admin", authMid.ProtectedAdmin404())

	admin.Get("/metrics", monitor.New(monitor.Config{
		Title: "ColPsiCarabobo - Panel de Control Administrativo",
	}))

	// ── Dashboard de analytics ───────────────────────────────────────────────
	// GET /api/v1/admin/dashboard/stats
	// Devuelve el JSON completo con todos los contadores, tendencias y tops
	admin.Get("/dashboard/stats", analyticsHandler.GetDashboardStats)

	// CRUD de Administradores (sin cambios)
	admin.Post("/create", h.CreateAdmin)
	admin.Get("/list", h.GetAdmins)
	admin.Patch("/update", h.UpdateAdmin)
	admin.Delete("/delete/:id", h.DeleteAdmin)
}

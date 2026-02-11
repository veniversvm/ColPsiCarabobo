package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"gorm.io/gorm"
)

func SetupAdminRoutes(router fiber.Router, db *gorm.DB) {
	// 1. Configuración de dependencias
	repo := postgres.NewAdminRepository(db)
	psiRepo := postgres.NewPsiRepository(db)
	svc := service.NewAdminService(repo)
	h := handler.NewAdminHandler(svc)
	authMid := middleware.NewAuthMiddleware(repo, psiRepo)

	// =========================================================================
	// RUTAS DE DESARROLLO (DEBUGGING)
	// =========================================================================

	// [!] ELIMINAR O COMENTAR ESTA LÍNEA EN PRODUCCIÓN
	// Permite acceso directo desde el navegador en: http://localhost:8080/api/v1/debug-monitor
	router.Get("/debug-monitor", monitor.New(monitor.Config{Title: "DEV ONLY - Monitor"}))

	// =========================================================================
	// RUTAS PÚBLICAS
	// =========================================================================
	auth := router.Group("/auth")
	auth.Post("/login", h.Login)

	// =========================================================================
	// RUTAS PROTEGIDAS (Admin Staff Only)
	// =========================================================================
	admin := router.Group("/admin", authMid.ProtectedAdmin404())

	// Versión profesional del monitor (Requiere JWT)
	admin.Get("/metrics", monitor.New(monitor.Config{
		Title: "ColPsiCarabobo - Panel de Control Administrativo",
	}))

	// CRUD de Administradores
	admin.Post("/create", h.CreateAdmin)
	admin.Get("/list", h.GetAdmins)
	admin.Patch("/update", h.UpdateAdmin)
	admin.Delete("/delete/:id", h.DeleteAdmin)
}

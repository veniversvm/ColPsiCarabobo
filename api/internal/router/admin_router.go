package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"gorm.io/gorm"
)

// CAMBIA EL NOMBRE AQUÍ A SetupAdminRoutes
func SetupAdminRoutes(router fiber.Router, db *gorm.DB) {
	repo := postgres.NewAdminRepository(db)
	psiRepo := postgres.NewPsiRepository(db)
	svc := service.NewAdminService(repo)
	h := handler.NewAdminHandler(svc)

	adminGroup := router.Group("/admin-routes")
	adminGroup.Post("/login", h.Login)

	authMid := middleware.NewAuthMiddleware(repo, psiRepo)

	adminGroup.Post("/create", authMid.ProtectedAdmin404(), h.CreateAdmin)
	adminGroup.Get("/list", authMid.ProtectedAdmin404(), h.GetAdmins)
	adminGroup.Patch("/update", authMid.ProtectedAdmin404(), h.UpdateAdmin)
	adminGroup.Delete("/delete/:id", authMid.ProtectedAdmin404(), h.DeleteAdmin)
}

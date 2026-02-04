package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"gorm.io/gorm"
)

func SetupAdminRoutes(router fiber.Router, db *gorm.DB) {
	repo := postgres.NewAdminRepository(db)
	svc := service.NewAdminService(repo)
	h := handler.NewAdminHandler(svc)

	adminGroup := router.Group("/admin-routes")
	adminGroup.Post("/login", h.Login)
}

package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"gorm.io/gorm"
)

func SetupSpecialtyRoutes(router fiber.Router, db *gorm.DB) {
	repo := postgres.NewPsiRepository(db)
	adminRepo := postgres.NewAdminRepository(db)
	specialtyRepo := postgres.NewSpecialtyRepository(db)
	svc := service.NewSpecialtyService(specialtyRepo)
	h := handler.NewSpecialtyHandler(svc)

	specialties := router.Group("/specialties")

	// RUTAS PÚBLICA (Accesible por pacientes/visitantes)
	specialties.Get("/", h.GetSpecialties)
	specialties.Get("/:id", h.GetSpecialtyByID)
	specialties.Get("/count", h.CountSpecialties)

	// RUTAS PROTEGIDAS (Solo Admin)
	authMid := middleware.NewAuthMiddleware(adminRepo, repo)
	admin := router.Group("/specialties", authMid.ProtectedAdmin404())
	admin.Post("/", h.CreateSpecialty)
	admin.Patch("/:id", h.UpdateSpecialty)
	admin.Delete("/:id", h.DeleteSpecialty)
}

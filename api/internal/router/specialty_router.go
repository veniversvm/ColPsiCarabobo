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

	authMid := middleware.NewAuthMiddleware(adminRepo, repo)

	// --- GRUPO ADMINISTRATIVO (ALTA PRIORIDAD) ---
	// Usamos un prefijo distinto (/admin/specialties) para evitar colisiones con los IDs
	// y cumplir con la semántica de tu documentación Swagger.
	admin := router.Group("/admin/specialties", authMid.ProtectedAdmin404())

	admin.Get("/all", h.GetAllAdmin) // URL: /api/v1/admin/specialties/all
	admin.Post("/", h.CreateSpecialty)
	admin.Patch("/:id", h.UpdateSpecialty)
	admin.Delete("/:id", h.DeleteSpecialty)

	// --- GRUPO PÚBLICO (BAJA PRIORIDAD) ---
	specialties := router.Group("/specialties")

	// Las rutas estáticas (/count) siempre deben ir ANTES de las dinámicas (/:id)
	specialties.Get("/count", h.CountSpecialties)
	specialties.Get("/", h.GetSpecialties)

	// La ruta con parámetro va al final para que no "se coma" a las anteriores
	// Tip Senior: Puedes añadir <int> para que Fiber solo coincida si el ID es numérico
	specialties.Get("/:id<int>", h.GetSpecialtyByID)
}

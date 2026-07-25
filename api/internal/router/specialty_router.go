// Package router organiza los puntos de entrada de la API segmentados por dominios de negocio.
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// SetupSpecialtyRoutes inicializa las dependencias y registra los endpoints del catálogo de especialidades.
// Aplica una arquitectura de capas inyectando Repositorios -> Servicios -> Handlers.
func SetupSpecialtyRoutes(router fiber.Router, psiRepo domain.PsiUserRepository, adminRepo domain.UserAdminRepository, specialtyRepo domain.SpecialtyRepository, analyticsSvc *service.AnalyticsService) {
	svc := service.NewSpecialtyService(specialtyRepo)
	h := handler.NewSpecialtyHandler(svc)

	// Configuración del middleware de autenticación dinámica.
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// =========================================================================
	// GRUPO ADMINISTRATIVO (ALTA PRIORIDAD)
	// =========================================================================
	// Usamos el prefijo /admin/specialties para separar físicamente las operaciones
	// sensibles. Se aplica ProtectedAdmin404 para ocultar los endpoints (Security by Obscurity).
	admin := router.Group("/admin/specialties", authMid.ProtectedAdmin404())

	admin.Get("/all", h.GetAllAdmin)        // Lista total (activas + inactivas)
	admin.Post("/", h.CreateSpecialty)      // Creación de nueva especialidad
	admin.Patch("/:id", h.UpdateSpecialty)  // Actualización parcial
	admin.Delete("/:id", h.DeleteSpecialty) // Desactivación lógica

	// =========================================================================
	// GRUPO PÚBLICO (BAJA PRIORIDAD)
	// =========================================================================
	// Estas rutas son accesibles por visitantes o pacientes. El orden de registro
	// aquí es crítico para evitar colisiones de ruteo.
	specialties := router.Group("/specialties")

	// Precedencia Estática: Registramos /count ANTES que /:id para evitar que
	// la palabra "count" sea interpretada erróneamente como un identificador.
	specialties.Get("/count", h.CountSpecialties)

	// Listado público filtrado (Solo registros activos).
	specialties.Get("/", h.GetSpecialties)

	// Precedencia Dinámica con Restricción: Usamos el constraint <int> para que
	// Fiber solo coincida con esta ruta si el parámetro es numérico.
	// Esto protege contra colisiones accidentales y mejora el rendimiento.
	specialties.Get("/:id<int>", h.GetSpecialtyByID)
}

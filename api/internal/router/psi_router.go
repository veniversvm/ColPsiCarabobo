package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware" // Importante para el auth
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

func SetupPsiRoutes(router fiber.Router, db *gorm.DB, s3Client *s3.S3Client) {
	// 1. Inicialización de dependencias
	repo := postgres.NewPsiRepository(db)
	adminRepo := postgres.NewAdminRepository(db)

	svc := service.NewPsiService(repo, s3Client)
	h := handler.NewPsiHandler(svc)

	// Instanciar Middleware de protección
	authMid := middleware.NewAuthMiddleware(adminRepo, repo)

	// =========================================================================
	// ZONA 1: ADMINISTRACIÓN (Requiere Token de Admin)
	// =========================================================================
	// Usamos un grupo separado para la gestión interna.
	adminGroup := router.Group("/admin/psi", authMid.ProtectedAdmin404())

	// URL: POST /api/v1/admin/psi/upload-csv
	adminGroup.Post("/upload-csv", h.UploadCsv)

	// =========================================================================
	// ZONA 2: AUTOGESTIÓN (Requiere Token de Psicólogo)
	// =========================================================================
	// El middleware ProtectedPsiUser valida la sesión del psicólogo
	meGroup := router.Group("/psi/me", authMid.ProtectedPsiUser())

	meGroup.Get("/", h.GetMe)              // Ver mi perfil completo
	meGroup.Patch("/", h.UpdateOwnProfile) // Actualizar mis datos permitidos
	meGroup.Post("/postgrades", h.AddPostGrade)
	// // Añadir un postgrado a mi cuenta
	meGroup.Patch("/postgrades/:id", h.UpdatePostGrade) // Actualizar un postgrado existente

	// =========================================================================
	// ZONA 3: PÚBLICO (Sin autenticación)
	// =========================================================================
	psiGroup := router.Group("/psi")

	// Login público para agremiados
	psiGroup.Post("/login", h.Login) // <--- NUEVA RUTA

	// URL: GET /api/v1/psi/directory (Listado paginado)
	// Nota: Siempre colocar las rutas estáticas antes de las dinámicas (/:id)
	psiGroup.Get("/directory", h.SearchDirectory)

	// URL: GET /api/v1/psi/a1b2c3d4... (Detalle público)
	// Tip Senior: Usamos <uuid> para asegurar que solo entre aquí si es un UUID válido.
	// Esto evita que Fiber confunda "/psi/directory" con un ID si algo falla en el orden.
	psiGroup.Get("/:id<uuid>", h.GetPublicProfile)
}

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
	mailService, err := service.NewMailService()
	if err != nil {
		panic("Error al inicializar el servicio de correo: " + err.Error())
	}

	svc := service.NewPsiService(repo, s3Client, mailService)
	h := handler.NewPsiHandler(svc)

	// Instanciar Middleware de protección
	authMid := middleware.NewAuthMiddleware(adminRepo, repo)

	// =========================================================================
	// ZONA 1: ADMINISTRACIÓN (Requiere Token de Admin)
	// =========================================================================
	// Usamos un grupo separado para la gestión interna.
	adminGroup := router.Group("/admin/psi", authMid.ProtectedAdmin404())

	adminGroup.Get("/list", h.ListAllPsis)
	adminGroup.Post("/create", h.CreatePsiByAdmin) // Creación individual
	adminGroup.Post("/upload-csv", h.UploadCsv)    // Creación masiva vía CSV
	adminGroup.Get("/:id<uuid>", h.GetPsiByIDAdmin)
	adminGroup.Patch("/:id", h.UpdatePsiByAdmin)  // Edición total
	adminGroup.Delete("/:id", h.DeletePsiByAdmin) // Borrado lógico

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

	// URL: GET /api/v1/psi/123456... (Detalle público)
	// usamos el fpv para la busqueda
	psiGroup.Get("/:id", h.GetPublicProfile)

	meGroup.Post("/social", h.AddSocialNetwork)
	meGroup.Patch("/social/:id", h.UpdateSocialNetwork)
	meGroup.Delete("/social/:id", h.DeleteSocialNetwork)
}

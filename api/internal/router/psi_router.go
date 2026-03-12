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

func SetupPsiRoutes(router fiber.Router, db *gorm.DB, s3Client *s3.S3Client, analyticsSvc *service.AnalyticsService) {
	repo := postgres.NewPsiRepository(db)
	adminRepo := postgres.NewAdminRepository(db)

	mailService, err := service.NewMailService()
	if err != nil {
		panic("Error al inicializar el servicio de correo: " + err.Error())
	}

	svc := service.NewPsiService(repo, s3Client, mailService)

	// ── PsiHandler ahora recibe analyticsSvc ─────────────────────────────────
	// Ver siguiente paso: actualizar NewPsiHandler y los métodos Login,
	// SearchDirectory y GetPublicProfile para llamar al servicio de analytics
	h := handler.NewPsiHandler(svc, analyticsSvc)

	authMid := middleware.NewAuthMiddleware(adminRepo, repo, analyticsSvc)

	// =========================================================================
	// ZONA 1: ADMINISTRACIÓN (sin cambios)
	// =========================================================================
	adminGroup := router.Group("/admin/psi", authMid.ProtectedAdmin404())

	adminGroup.Get("/list", h.ListAllPsis)
	adminGroup.Post("/create", h.CreatePsiByAdmin)
	adminGroup.Post("/upload-csv", h.UploadCsv)
	adminGroup.Get("/:id<uuid>", h.GetPsiByIDAdmin)
	adminGroup.Patch("/:id", h.UpdatePsiByAdmin)
	adminGroup.Delete("/:id", h.DeletePsiByAdmin)

	// Rutas de social media para admin
	// adminGroup.Post("/:id/social", h.AdminAddSocialNetwork)
	// adminGroup.Delete("/:id/social/:socialId", h.AdminDeleteSocialNetwork)

	// =========================================================================
	// ZONA 2: AUTOGESTIÓN (sin cambios)
	// =========================================================================
	meGroup := router.Group("/psi/me", authMid.ProtectedPsiUser())

	meGroup.Get("/", h.GetMe)
	meGroup.Patch("/", h.UpdateOwnProfile)
	meGroup.Post("/postgrades", h.AddPostGrade)
	meGroup.Patch("/postgrades/:id", h.UpdatePostGrade)
	meGroup.Post("/social", h.AddSocialNetwork)
	meGroup.Patch("/social/:id", h.UpdateSocialNetwork)
	meGroup.Delete("/social/:id", h.DeleteSocialNetwork)
	meGroup.Post("/logout", h.Logout)

	// =========================================================================
	// ZONA 3: PÚBLICO — aquí viven los 3 puntos de analytics del psicólogo
	// =========================================================================
	psiGroup := router.Group("/psi")

	// RecordLogin se llama dentro de h.Login tras autenticar exitosamente
	psiGroup.Post("/login", h.Login)

	// RecordSearch se llama dentro de h.SearchDirectory tras ejecutar la query
	psiGroup.Get("/directory", h.SearchDirectory)

	// RecordProfileView se llama dentro de h.GetPublicProfile tras obtener el perfil
	psiGroup.Get("/:id", h.GetPublicProfile)
}

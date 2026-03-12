// Package router organiza los puntos de entrada de la API segmentados por dominios de negocio.
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"gorm.io/gorm"
)

// SetupPostRoutes inicializa las dependencias y registra los endpoints del módulo de noticias y publicaciones.
// Implementa un sistema de visibilidad dual: público para visitantes y extendido para miembros autenticados.
func SetupPostRoutes(router fiber.Router, db *gorm.DB, s3Client *s3.S3Client, analyticsSvc *service.AnalyticsService) {
	// 1. INYECCIÓN DE DEPENDENCIAS
	// Inicializamos repositorios necesarios para la lógica de negocio y la validación de seguridad.
	adminRepo := postgres.NewAdminRepository(db)
	psiRepo := postgres.NewPsiRepository(db)
	postRepo := postgres.NewPostRepository(db)

	// Inicializamos el servicio con integración de S3 para el manejo de imágenes de portada.
	svc := service.NewPostService(postRepo, s3Client)
	h := handler.NewPostHandler(svc)

	// Configuración del middleware de autenticación que soporta múltiples roles (Admin/Psicólogo).
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	// =========================================================================
	// GRUPO DE CONSULTA (ACCESO HÍBRIDO)
	// =========================================================================
	// Estas rutas utilizan 'OptionalHybridAuth', lo que permite que el endpoint sea público
	// pero que "sepa" quién lo llama para mostrar contenido restringido si el token es válido.
	posts := router.Group("/posts")

	// Listado de noticias:
	// - Público: Ve solo noticias marcadas como 'public'.
	// - Psicólogo: Ve noticias 'public' + 'psi'.
	// - Admin: Ve todo (incluyendo borradores).
	posts.Get("/", authMid.OptionalHybridAuth(), h.ListPosts)

	// Detalle de noticia: Valida visibilidad según el rol inyectado por el middleware.
	posts.Get("/:id", authMid.OptionalHybridAuth(), h.GetPost)

	// =========================================================================
	// GRUPO ADMINISTRATIVO (ALTA SEGURIDAD)
	// =========================================================================
	// Usamos el prefijo /admin para separar la gestión del contenido.
	// Se aplica ProtectedAdmin404 para ocultar la existencia de estos endpoints a usuarios no autorizados.
	admin := router.Group("/admin/posts", authMid.ProtectedAdmin404())

	// Registro de nueva publicación con soporte para carga de imagen a S3.
	admin.Post("/", h.CreatePost)

	// Actualización parcial (PATCH) de metadatos o contenido HTML.
	admin.Patch("/:id", h.UpdatePost)

	// Nota de Arquitectura: El borrado se mantiene comentado hasta definir si será
	// físico o un Soft-Delete adicional al del modelo.
	// admin.Delete("/:id", h.DeletePost)
}

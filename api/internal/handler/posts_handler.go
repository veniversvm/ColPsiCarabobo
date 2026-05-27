// api/internal/handler/posts_handler.go
package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

type PostHandler struct {
	service *service.PostService
}

func NewPostHandler(svc *service.PostService) *PostHandler {
	return &PostHandler{service: svc}
}

// ListPosts godoc
// @Summary      Listar noticias y publicaciones
// @Description  Obtiene noticias. El contenido varía según el usuario: Público (solo publicados públicos), Psicólogo (públicos + gremiales), Admin (todos los estados).
// @Tags         Publicaciones
// @Produce      json
// @Param        page   query  int  false  "Número de página (Default: 1)"
// @Param        limit  query  int  false  "Items por página (Default: 10)"
// @Success      200    {object}  map[string]interface{} "data, total, page"
// @Failure      500    {object}  map[string]string
// @Router       /posts [get]
func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
	role := "public"
	if _, ok := c.Locals("admin").(*domain.UserAdmin); ok {
		role = "admin"
	} else if _, ok := c.Locals("psi_user").(*domain.PsiUserModel); ok {
		role = "psi"
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	res, err := h.service.GetPostsList(c.UserContext(), page, limit, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al recuperar publicaciones"})
	}

	return c.JSON(res)
}

// CreatePost godoc
// @Summary      Crear nueva publicación
// @Description  Crea un post con contenido HTML y opcionalmente una imagen. El campo 'status' controla el ciclo de vida.
// @Security     BearerAuth
// @Tags         Administración - Publicaciones
// @Accept       multipart/form-data
// @Produce      json
// @Param        title             formData  string  true   "Título del post (max 100 chars)"
// @Param        short_description formData  string  false  "Resumen para el feed (max 250 chars)"
// @Param        content           formData  string  true   "Contenido HTML/Texto"
// @Param        type              formData  string  true   "Visibilidad: public | psi"
// @Param        status            formData  string  true   "Estado: draft | published | archived | scheduled"
// @Param        publish_at        formData  string  false  "Fecha ISO8601 de publicación — obligatorio si status=scheduled"
// @Param        image             formData  file    false  "Imagen de portada"
// @Success      201               {object}  map[string]string
// @Failure      400               {object}  map[string]string
// @Failure      500               {object}  map[string]string
// @Router       /admin/posts [post]
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	admin := c.Locals("admin").(*domain.UserAdmin)

	var req request_structs.CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos"})
	}

	file, _ := c.FormFile("image")

	if err := h.service.CreatePost(c.UserContext(), admin, req, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Post creado exitosamente"})
}

// GetPost godoc
// @Summary      Obtener publicación por ID
// @Description  Devuelve el detalle completo incluyendo contenido HTML. Aplica ACL según el rol del solicitante.
// @Tags         Publicaciones
// @Produce      json
// @Param        id  path  string  true  "UUID del Post"
// @Success      200  {object}  domain.Post
// @Failure      404  {object}  map[string]string
// @Router       /posts/{id} [get]
func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	idParam := c.Params("id")
	postID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de formato inválido"})
	}

	role := "public"
	if _, ok := c.Locals("admin").(*domain.UserAdmin); ok {
		role = "admin"
	} else if _, ok := c.Locals("psi_user").(*domain.PsiUserModel); ok {
		role = "psi"
	}

	post, err := h.service.GetPostByID(c.UserContext(), postID, role)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post no encontrado"})
	}

	return c.JSON(post)
}

// UpdatePost godoc
// @Summary      Actualizar publicación
// @Description  Edición parcial (PATCH). Solo los campos enviados se modifican. Permite cambiar el estado del ciclo de vida.
// @Security     BearerAuth
// @Tags         Administración - Publicaciones
// @Accept       multipart/form-data
// @Produce      json
// @Param        id                path      string  true   "UUID de la publicación"
// @Param        title             formData  string  false  "Nuevo título"
// @Param        short_description formData  string  false  "Nueva descripción corta"
// @Param        content           formData  string  false  "Nuevo contenido HTML"
// @Param        type              formData  string  false  "Nueva visibilidad: public | psi"
// @Param        status            formData  string  false  "Nuevo estado: draft | published | archived | scheduled"
// @Param        publish_at        formData  string  false  "Nueva fecha de publicación — obligatorio si status=scheduled"
// @Param        image             formData  file    false  "Nueva imagen de portada (reemplaza la anterior)"
// @Success      200               {object}  map[string]string
// @Failure      400               {object}  map[string]string
// @Failure      500               {object}  map[string]string
// @Router       /admin/posts/{id} [patch]
func (h *PostHandler) UpdatePost(c *fiber.Ctx) error {
	admin := c.Locals("admin").(*domain.UserAdmin)

	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID requerido en la ruta"})
	}

	uuidID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de formato inválido"})
	}

	var req request_structs.UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos"})
	}

	file, _ := c.FormFile("image")

	if err := h.service.UpdatePost(c.UserContext(), admin, req, file, uuidID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Publicación actualizada correctamente"})
}

func (h *PostHandler) GetSiteMapHandler(c *fiber.Ctx) error {
	// 1. Llamamos al servicio usando c.UserContext() para respetar timeouts
	data, err := h.service.GetSitemapData(c.UserContext())

	// 2. Si hay error, respondemos con un status 500
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudieron obtener los datos para el sitemap: " + err.Error(),
		})
	}

	// 3. Si todo está bien, enviamos el JSON con los posts
	// Esto es lo que leerá tu archivo sitemap.xml.ts en el frontend
	return c.JSON(data)
}

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
// @Description  Obtiene noticias. El contenido varía según el usuario: Público (solo noticias generales), Psicólogo (generales + gremiales), Admin (todo).
// @Tags         Publicaciones
// @Produce      json
// @Param        page   query     int     false  "Número de página (Default: 1)"
// @Param        limit  query     int     false  "Items por página (Default: 10)"
// @Success      200    {object}  map[string]interface{} "data, total, page"
// @Failure      500    {object}  map[string]string
// @Router       /posts [get]
func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
	// 1. Detección de Rol (Polimorfismo basado en Contexto)
	role := "public"

	// El middleware OptionalHybridAuth ya validó el token e inyectó el struct correcto
	if _, ok := c.Locals("admin").(*domain.UserAdmin); ok {
		role = "admin"
	} else if _, ok := c.Locals("psi_user").(*domain.PsiUserModel); ok {
		role = "psi"
	}

	// 2. Paginación Segura (Fail-safe defaults)
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
	} // Protección contra carga excesiva

	// 3. Llamada al servicio
	res, err := h.service.GetPostsList(c.UserContext(), page, limit, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al recuperar publicaciones"})
	}

	return c.JSON(res)
}

// CreatePost godoc
// @Summary      Publicar nueva noticia
// @Description  Crea un post con contenido HTML y opcionalmente una imagen.
// @Security     BearerAuth
// @Tags         Administración - Publicaciones
// @Accept       multipart/form-data
// @Produce      json
// @Param        title             formData  string  true   "Título del post"
// @Param        short_description formData  string  false  "Resumen para el feed"
// @Param        content           formData  string  true   "Contenido HTML/Texto"
// @Param        type              formData  string  true   "Tipo: public, psi"
// @Param        is_active         formData  bool    true   "Visible inmediatamente"
// @Param        image             formData  file    false  "Imagen de portada"
// @Success      201               {object}  map[string]string
// @Failure      400               {object}  map[string]string
// @Failure      403               {object}  map[string]string
// @Router       /admin/posts [post]
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	admin := c.Locals("admin").(*domain.UserAdmin)

	var req request_structs.CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos"})
	}

	// Recuperar archivo de imagen (opcional)
	file, _ := c.FormFile("image")

	if err := h.service.CreatePost(c.UserContext(), admin, req, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Post publicado exitosamente"})
}

// GetPost godoc
// @Summary      Obtener noticia por ID
// @Description  Devuelve el detalle de la noticia incluyendo el contenido HTML. Valida permisos de visibilidad.
// @Tags         Publicaciones
// @Produce      json
// @Param        id   path      string  true  "UUID del Post"
// @Success      200  {object}  domain.Post
// @Failure      404  {object}  map[string]string
// @Router       /posts/{id} [get]
func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	// 1. Validar UUID
	idParam := c.Params("id")
	postID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de formato inválido"})
	}

	// 2. Detectar Rol (Gracias al middleware OptionalHybridAuth)
	role := "public"
	if _, ok := c.Locals("admin").(*domain.UserAdmin); ok {
		role = "admin"
	} else if _, ok := c.Locals("psi_user").(*domain.PsiUserModel); ok {
		role = "psi"
	}

	// 3. Llamar al servicio
	post, err := h.service.GetPostByID(c.UserContext(), postID, role)
	if err != nil {
		// Usamos 404 para no revelar si existe pero es privado
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post no encontrado"})
	}

	return c.JSON(post)
}

// UpdatePost godoc
// @Summary      Actualizar publicación
// @Description  Actualiza metadatos e imagen. El ID se obtiene de la URL. Los campos son opcionales (PATCH).
// @Security     BearerAuth
// @Tags         Administración - Publicaciones
// @Accept       mpfd
// @Produce      json
// @Param        id                path      string  true  "UUID de la publicación"
// @Param        title             formData  string  false "Nuevo título"
// @Param        short_description formData  string  false "Descripción corta"
// @Param        content           formData  string  false "Contenido detallado"
// @Param        type              formData  string  false "Tipo de post"
// @Param        is_active         formData  boolean false "Estado de activación"
// @Param        image             formData  file    false "Nueva imagen de portada"
// @Success      200               {object}  map[string]string "{"message": "Publicación actualizada"}"
// @Failure      400               {object}  map[string]string "{"error": "Datos inválidos"}"
// @Router       /admin/posts/{id} [patch]
func (h *PostHandler) UpdatePost(c *fiber.Ctx) error {
	admin := c.Locals("admin").(*domain.UserAdmin)

	// 1. Obtener el ID únicamente de la ruta
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID requerido en la ruta"})
	}

	uuidID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de formato inválido"})
	}

	// 2. Parsear el cuerpo (esto llenará el struct con los campos formData)
	var req request_structs.UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos"})
	}

	// 3. Recuperar el archivo (si existe)
	file, _ := c.FormFile("image")

	// 4. Pasar el idParam directamente al servicio junto con los datos y el archivo
	// El servicio debería encargarse de convertir idParam a UUID si es necesario
	if err := h.service.UpdatePost(c.UserContext(), admin, req, file, uuidID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Publicación actualizada correctamente"})
}

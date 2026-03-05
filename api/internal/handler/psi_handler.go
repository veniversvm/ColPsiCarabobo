// api/internal/handler/psi_handler.go

package handler

import (
	"fmt"
	"log"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

type PsiHandler struct {
	service *service.PsiService
}

// NewPsiHandler inyecta el servicio de psicólogos en el controlador
func NewPsiHandler(svc *service.PsiService) *PsiHandler {
	return &PsiHandler{
		service: svc,
	}
}

// UploadCsv godoc
// @Summary      Importar psicólogos masivamente
// @Description  Procesa un archivo CSV y crea registros. Solo accesible para administradores autorizados.
// @Tags         Administración - Psicólogos  <-- Tag corregido para agruparlo bien
// @Accept       multipart/form-data
// @Produce      json
// @Param        csv  formData  file  true  "Archivo CSV a procesar"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string "Retornado si el usuario no tiene permisos (Security by Obscurity)"
// @Router       /admin/psi/upload-csv [post]
func (h *PsiHandler) UploadCsv(c *fiber.Ctx) error {
	fmt.Println("Entrando a UploadCsv")
	// 1. OBTENER EL ADMIN DESDE EL TOKEN (Locals)
	// Esto es lo que inyectó el middleware. Si llegamos aquí, ya está validado.
	admin, ok := c.Locals("admin").(*domain.UserAdmin)
	if !ok || admin == nil {
		// Si no hay admin en locals, algo falló en el middleware. 404 por seguridad.
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Cannot POST /api/v1/psi/upload-csv",
		})
	}

	// 2. BORRA TODO EL BLOQUE QUE HACÍA: c.FormValue("admin_id")
	// Ya no lo necesitamos.

	// 3. PROCESAR EL ARCHIVO CSV
	file, err := c.FormFile("csv")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El archivo CSV es requerido",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No se pudo abrir el archivo"})
	}
	defer src.Close()

	// 4. USAR EL ID DEL ADMIN RECUPERADO DEL TOKEN
	success, failed := h.service.ImportFromCSV(c.UserContext(), src, admin.ID)

	return c.JSON(fiber.Map{
		"success_count": success,
		"failed":        failed,
	})
}

// UpdateOwnProfile godoc
// @Summary      Actualizar mi perfil (Autogestión)
// @Description  Permite al psicólogo actualizar su información. Requiere la contraseña actual para validar cambios. Soporta subida de foto de perfil e imágenes del título.
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Accept       multipart/form-data
// @Produce      json
// @Param        password      formData string true  "Contraseña actual obligatoria"
// @Param        username      formData string false "Nuevo nombre de usuario"
// @Param        profile_picture formData file  false "Imagen de perfil (JPEG/PNG)"
// @Param        title_image_one   formData file false "Imagen del título 1 (JPEG/PNG)"
// @Param        title_image_two   formData file false "Imagen del título 2 (JPEG/PNG)"
// @Param        title_image_three formData file false "Imagen del título 3 (JPEG/PNG)"
// @Param        request       body     request_structs.PsiUserUpdateRequestSelf true "Otros datos (enviar como form-data)"
// @Success      200           {object} map[string]interface{}
// @Failure      401           {object} map[string]string "Contraseña incorrecta"
// @Router       /psi/me [patch]
func (h *PsiHandler) UpdateOwnProfile(c *fiber.Ctx) error {
	// 1. Obtener identidad segura desde el token
	updater := c.Locals("psi_user").(*domain.PsiUserModel)

	// 2. Parsear campos de texto del formulario
	var req request_structs.PsiUserUpdateRequestSelf
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Formato de datos inválido"})
	}

	// 3. Capturar archivos de imagen (Opcionales)
	profilePic, _ := c.FormFile("profile_picture")
	titleImgOne, _ := c.FormFile("title_image_one")
	titleImgTwo, _ := c.FormFile("title_image_two")
	titleImgThree, _ := c.FormFile("title_image_three")

	// 4. Validar contraseña actual presente
	if req.Password == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Se requiere la contraseña actual para confirmar cambios"})
	}

	// 5. Llamar al servicio inyectando los archivos
	profile, err := h.service.UpdateProfileSelf(
		c.UserContext(),
		updater,
		updater.ID,
		req,
		profilePic,
		titleImgOne,
		titleImgTwo,
		titleImgThree,
	)

	if err != nil {
		if err.Error() == "contraseña actual incorrecta" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Perfil actualizado correctamente",
		"id":      profile.ID,
	})
}

// SearchDirectory godoc
// @Summary      Directorio Público de Psicólogos
// @Description  Motor de búsqueda avanzado. Si se usa 'q', busca por identidad (ignora solvencia). Si no se usa 'q', solo muestra solventes y aplica filtros de ubicación/especialidad.
// @Tags         Psicólogos - Público
// @Produce      json
// @Param        q          query     string  false  "Búsqueda por Nombre, Apellido, CI o FPV"
// @Param        specialty  query     int     false  "ID de la Especialidad (Catálogo Maestro)"
// @Param        location   query     string  false  "Municipio o Estado"
// @Param        gender     query     string  false  "Género: M o F"
// @Param        page       query     int     false  "Página (Def: 1)"
// @Param        limit      query     int     false  "Límite (Def: 12)"
// @Success      200        {object}  map[string]interface{}
// @Router       /psi/directory [get]
func (h *PsiHandler) SearchDirectory(c *fiber.Ctx) error {
	// Usamos QueryParser para llenar el DTO automáticamente
	var filter request_structs.PsiDirectoryFilterDTO
	if err := c.QueryParser(&filter); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Parámetros de consulta inválidos"})
	}

	result, err := h.service.GetPublicDirectory(c.UserContext(), filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error interno en la búsqueda"})
	}

	return c.JSON(result)
}

// GetPublicProfile godoc
// @Summary      Ver perfil público de psicólogo
// @Description  Retorna la ficha técnica. Oculta datos privados (teléfono, email) según configuración del usuario.
// @Tags         Psicólogos - Público
// @Produce      json
// @Param        id   path      string  true  "UUID del Psicólogo"
// @Success      200  {object}  request_structs.PsiFullProfileDTO
// @Failure      404  {object}  map[string]string
// @Router       /psi/{id} [get]
func (h *PsiHandler) GetPublicProfile(c *fiber.Ctx) error {
	fvp, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	profile, err := h.service.GetPublicProfile(c.UserContext(), fvp)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(profile)
}

// GetMe godoc
// @Summary Obtener mi propio perfil (Autogestión)
// @Description Retorna toda la información del psicólogo autenticado (sin aplicar filtros de privacidad o solvencia). Ideal para el Dashboard del usuario.
// @Security BearerAuth
// @Tags Psicólogos - Perfil
// @Produce json
// @Success 200 {object} domain.PsiUserModel
// @Failure 401 {object} map[string]interface{} "No autorizado"
// @Router /psi/me [get]
func (h *PsiHandler) GetMe(c *fiber.Ctx) error {
	// Obtenemos el psicólogo inyectado por el middleware ProtectedPsiUser.
	// Esta consulta ya se hizo hace microsegundos y trae TODO gracias al Preload del Repositorio.
	psi, ok := c.Locals("psi_user").(*domain.PsiUserModel)
	if !ok || psi == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Sesión inválida o expirada",
		})
	}

	bio, err := h.service.GetPsiBioByID(c.UserContext(), psi.BioTextID)
	if err != nil {
		log.Printf("---- error al recuperar la BIO ----\n")
		log.Printf("%v\n", err)
	}
	log.Printf("----- BIO ----\n")
	log.Printf("%v", bio)

	psi.FullBio.Content = bio

	// Retornamos el modelo completo.
	// Nota: Password y Key no se envían porque en domain.PsiUserModel tienen `json:"-"`.
	return c.JSON(psi)
}

// Login godoc
// @Summary      Login de Psicólogo
// @Description  Autentica a un agremiado y retorna un token de sesión.
// @Tags         Psicólogos - Auth
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.PsiLoginRequest  true  "Credenciales"  <-- AQUÍ ESTÁ EL CAMBIO
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]string
// @Router       /psi/login [post]
func (h *PsiHandler) Login(c *fiber.Ctx) error {
	var req request_structs.PsiLoginRequest // Esto está bien en Go
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	token, err := h.service.Login(c.UserContext(), req.Identifier, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Bienvenido colega",
		"token":   token,
	})
}

// AddPostGrade godoc
// @Summary      Agregar postgrado con soportes
// @Description  Registra un título académico y permite subir hasta 3 imágenes (título, notas, etc).
// @Security     BearerAuth
// @Tags         Psicólogos - Académico
// @Accept       multipart/form-data   <-- CAMBIO IMPORTANTE
// @Produce      json
// @Param        title            formData  string  true   "Título obtenido"
// @Param        university       formData  string  true   "Universidad"
// @Param        graduation_year  formData  string  true   "Año de graduación"
// @Param        description      formData  string  false  "Descripción opcional"
// @Param        pic_one          formData  file    false  "Imagen del Título"
// @Param        pic_two          formData  file    false  "Imagen de Notas"
// @Param        pic_three        formData  file    false  "Imagen Extra"
// @Success      201              {object}  map[string]string "message"
// @Router       /psi/me/postgrades [post]
func (h *PsiHandler) AddPostGrade(c *fiber.Ctx) error {
	psi := c.Locals("psi_user").(*domain.PsiUserModel)

	// 1. Parsear Texto
	var req request_structs.CreatePostGradeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos de formulario inválidos"})
	}

	// 2. Extraer Archivos
	// Fiber retorna error si el archivo no existe, así que ignoramos el error
	// y verificamos si el puntero es nil en el servicio.
	file1, _ := c.FormFile("pic_one")
	file2, _ := c.FormFile("pic_two")
	file3, _ := c.FormFile("pic_three")

	// Crear un slice para pasar al servicio
	files := []*multipart.FileHeader{file1, file2, file3}

	// 3. Llamar al servicio
	if err := h.service.AddPostGrade(c.UserContext(), psi, req, files); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Postgrado y documentos registrados exitosamente"})
}

// Los demás métodos (ListPsis, GetPsi, etc.) se implementarán siguiendo el mismo patrón...
func (h *PsiHandler) ListPsis(c *fiber.Ctx) error  { return nil }
func (h *PsiHandler) GetPsi(c *fiber.Ctx) error    { return nil }
func (h *PsiHandler) UpdatePsi(c *fiber.Ctx) error { return nil }

// UpdatePostGrade godoc
// @Summary      Actualizar postgrado
// @Description  Edita un título. Permite reemplazar imágenes específicas enviando 'pic_one', 'pic_two' o 'pic_three'.
// @Security     BearerAuth
// @Tags         Psicólogos - Académico
// @Accept       multipart/form-data
// @Produce      json
// @Param        id               path      string  true   "ID del Postgrado"
// @Param        title            formData  string  false  "Título obtenido"
// @Param        university       formData  string  false  "Universidad"
// @Param        graduation_year  formData  string  false  "Año"
// @Param        description      formData  string  false  "Descripción"
// @Param        pic_one          formData  file    false  "Reemplazar Imagen 1"
// @Param        pic_two          formData  file    false  "Reemplazar Imagen 2"
// @Param        pic_three        formData  file    false  "Reemplazar Imagen 3"
// @Success      200              {object}  map[string]string "message"
// @Failure      403              {object}  map[string]string "No es tu registro"
// @Router       /psi/me/postgrades/{id} [patch]
func (h *PsiHandler) UpdatePostGrade(c *fiber.Ctx) error {
	psi := c.Locals("psi_user").(*domain.PsiUserModel)

	// Validar ID del Path
	pgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	// Construir DTO manual para campos de texto (punteros)
	req := request_structs.UpdatePostGradeRequest{}

	if val := c.FormValue("title"); val != "" {
		req.Title = &val
	}
	if val := c.FormValue("university"); val != "" {
		req.University = &val
	}
	if val := c.FormValue("graduation_year"); val != "" {
		req.GraduationYear = &val
	}
	if val := c.FormValue("description"); val != "" {
		req.Description = &val
	}

	// Recolectar archivos en un mapa
	fileMap := make(map[string]*multipart.FileHeader)

	if f, err := c.FormFile("pic_one"); err == nil {
		fileMap["pic_one"] = f
	}
	if f, err := c.FormFile("pic_two"); err == nil {
		fileMap["pic_two"] = f
	}
	if f, err := c.FormFile("pic_three"); err == nil {
		fileMap["pic_three"] = f
	}

	// Llamar servicio
	if err := h.service.UpdatePostGrade(c.UserContext(), psi, pgID, req, fileMap); err != nil {
		// Distinguir error de propiedad
		if err.Error() == "no tienes permiso para editar este registro" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Título actualizado correctamente"})
}

// AddSocialNetwork godoc
// @Summary      Agregar red social
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Accept       json
// @Produce      json
// @Param        request body request_structs.CreateSocialNetworkRequest true "Datos de la red"
// @Success      201 {object} map[string]string
// @Router       /psi/me/social [post]
func (h *PsiHandler) AddSocialNetwork(c *fiber.Ctx) error {
	psi := c.Locals("psi_user").(*domain.PsiUserModel)
	var req request_structs.CreateSocialNetworkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.AddSocialNetwork(c.UserContext(), psi, req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Red social agregada"})
}

// UpdateSocialNetwork godoc
// @Summary      Actualizar red social
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID de la red"
// @Param        request body request_structs.UpdateSocialNetworkRequest true "Campos parciales"
// @Success      200 {object} map[string]string
// @Router       /psi/me/social/{id} [patch]
func (h *PsiHandler) UpdateSocialNetwork(c *fiber.Ctx) error {
	psi := c.Locals("psi_user").(*domain.PsiUserModel)
	netID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateSocialNetworkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateSocialNetwork(c.UserContext(), psi, netID, req); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Red social actualizada"})
}

// DeleteSocialNetwork godoc
// @Summary      Borrar red social (Soft Delete)
// @Description  Puede ser usado por el psicólogo dueño o por un Administrador.
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Param        id path string true "UUID de la red"
// @Success      200 {object} map[string]string
// @Router       /psi/me/social/{id} [delete]
// @Router       /admin/psi/social/{id} [delete]
func (h *PsiHandler) DeleteSocialNetwork(c *fiber.Ctx) error {
	netID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	role := ""
	var execID uuid.UUID

	if admin, ok := c.Locals("admin").(*domain.UserAdmin); ok && admin != nil {
		role = "admin"
		execID = admin.ID
	} else if psi, ok := c.Locals("psi_user").(*domain.PsiUserModel); ok && psi != nil {
		role = "psi"
		execID = psi.ID
	}

	if err := h.service.DeleteSocialNetwork(c.UserContext(), role, execID, netID); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Red social eliminada correctamente"})
}

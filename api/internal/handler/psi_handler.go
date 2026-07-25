// api/internal/handler/psi_handler.go

package handler

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// ── Struct: se añade analytics ────────────────────────────────────────────────
type PsiHandler struct {
	service   *service.PsiService
	analytics *service.AnalyticsService // 👈 NUEVO
}

// ── Constructor: acepta el segundo parámetro ──────────────────────────────────
func NewPsiHandler(svc *service.PsiService, analytics *service.AnalyticsService) *PsiHandler {
	return &PsiHandler{
		service:   svc,
		analytics: analytics, // 👈 NUEVO
	}
}

// UploadCsv godoc
// @Summary      Importar psicólogos masivamente
// @Description  Procesa un archivo CSV y crea registros. Solo accesible para administradores autorizados.
// @Tags         Administración - Psicólogos
// @Accept       multipart/form-data
// @Produce      json
// @Param        csv  formData  file  true  "Archivo CSV a procesar"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /admin/psi/upload-csv [post]
func (h *PsiHandler) UploadCsv(c *fiber.Ctx) error {
	fmt.Println("Entrando a UploadCsv")
	admin, ok := c.Locals("admin").(*domain.UserAdmin)
	if !ok || admin == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Cannot POST /api/v1/psi/upload-csv",
		})
	}

	file, err := c.FormFile("xlsx")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No se pudo abrir el archivo"})
	}
	defer src.Close()

	success, failed_records := h.service.ImportFromCSV(c.UserContext(), src, admin.ID)

	return c.JSON(fiber.Map{
		"imported": success,
		"failed":   len(failed_records),
		"errors":   failed_records,
	})
}

// UpdateOwnProfile godoc
// @Summary      Actualizar mi perfil (Autogestión)
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Accept       multipart/form-data
// @Produce      json
// @Param        password          formData string true  "Contraseña actual obligatoria"
// @Param        username          formData string false "Nuevo nombre de usuario"
// @Param        profile_picture   formData file  false  "Imagen de perfil (JPEG/PNG)"
// @Param        title_image_one   formData file  false  "Imagen del título 1"
// @Param        title_image_two   formData file  false  "Imagen del título 2"
// @Param        title_image_three formData file  false  "Imagen del título 3"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /psi/me [patch]
func (h *PsiHandler) UpdateOwnProfile(c *fiber.Ctx) error {
	updater, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return err
	}

	var req request_structs.PsiUserUpdateRequestSelf
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Formato de datos inválido"})
	}

	profilePic, _ := c.FormFile("profile_picture")
	titleImgOne, _ := c.FormFile("title_image_one")
	titleImgTwo, _ := c.FormFile("title_image_two")
	titleImgThree, _ := c.FormFile("title_image_three")

	if req.Password == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Se requiere la contraseña actual para confirmar cambios"})
	}

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
		if errors.Is(err, domain.ErrPasswordIncorrect) {
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
// @Description  Motor de búsqueda avanzado con filtros de especialidad, ubicación y género.
// @Tags         Psicólogos - Público
// @Produce      json
// @Param        q         query string false "Búsqueda por Nombre, Apellido, CI o FPV"
// @Param        specialty query int    false "ID de la Especialidad"
// @Param        location  query string false "Municipio o Estado"
// @Param        gender    query string false "Género: M o F"
// @Param        page      query int    false "Página (Def: 1)"
// @Param        limit     query int    false "Límite (Def: 12)"
// @Success      200 {object} map[string]interface{}
// @Router       /psi/directory [get]
func (h *PsiHandler) SearchDirectory(c *fiber.Ctx) error {
	var filter request_structs.PsiDirectoryFilterDTO
	if err := c.QueryParser(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Parámetros de consulta inválidos"})
	}

	filter = request_structs.SanitizeDirectoryFilter(filter)

	result, err := h.service.GetPublicDirectory(c.UserContext(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error interno en la búsqueda"})
	}

	// ── Analytics: registrar búsqueda ────────────────────────────────────────
	// Se ejecuta DESPUÉS de obtener los resultados para incluir el conteo real
	var viewerID *uuid.UUID
	if uid, ok := c.Locals("userID").(uuid.UUID); ok {
		viewerID = &uid
	}
	specialtyStr := fmt.Sprintf("%d", filter.SpecialtyID)
	h.analytics.RecordSearch(
		filter.SearchTerm, // "q" en la query string
		specialtyStr,      // SpecialtyID como string
		filter.Location,   // "location" en la query string
		"",                // no tienes state separado en el DTO
		0,                 // result es interface{}, no accesible sin type assertion
		viewerID,
		c.Cookies("_sid"),
		c.IP(),
	)
	// ─────────────────────────────────────────────────────────────────────────

	return c.JSON(result)
}

// GetPublicProfile godoc
// @Summary      Ver perfil público de psicólogo
// @Description  Retorna la ficha técnica. Oculta datos privados según configuración del usuario.
// @Tags         Psicólogos - Público
// @Produce      json
// @Param        id path string true "FPV del Psicólogo"
// @Success      200 {object} request_structs.PsiFullProfileDTO
// @Failure      404 {object} map[string]string
// @Router       /psi/{id} [get]
func (h *PsiHandler) GetPublicProfile(c *fiber.Ctx) error {
	fvp, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	profile, psi_id, err := h.service.GetPublicProfile(c.UserContext(), fvp)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	// ── Analytics: registrar visita al perfil ────────────────────────────────
	var viewerID *uuid.UUID
	if uid, ok := c.Locals("userID").(uuid.UUID); ok {
		viewerID = &uid
	}
	h.analytics.RecordProfileView(
		psi_id, // uuid.UUID del psicólogo visto (ajusta al campo de tu DTO)
		viewerID,
		c.Cookies("_sid"),
		c.IP(),
	)
	// ─────────────────────────────────────────────────────────────────────────

	return c.JSON(profile)
}

// GetMe godoc
// @Summary      Obtener mi propio perfil (Autogestión)
// @Description  Retorna toda la información del psicólogo autenticado sin filtros de privacidad.
// @Security     BearerAuth
// @Tags         Psicólogos - Perfil
// @Produce      json
// @Success      200 {object} domain.PsiUserModel
// @Failure      401 {object} map[string]interface{}
// @Router       /psi/me [get]
func (h *PsiHandler) GetMe(c *fiber.Ctx) error {
	psi, ok := c.Locals("psi_user").(*domain.PsiUserModel)
	if !ok || psi == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Sesión inválida o expirada",
		})
	}

	log.Printf("PSI Profile loaded: id=%s, username=%s", psi.ID, psi.Username)

	bio, err := h.service.GetPsiBioByID(c.UserContext(), psi.BioTextID)
	if err != nil {
		log.Printf("---- error al recuperar la BIO ----\n")
		log.Printf("%v\n", err)
	}

	psi.FullBio.Content = bio

	return c.JSON(psi)
}

// Login godoc
// @Summary      Login de Psicólogo
// @Description  Autentica a un agremiado y retorna un token de sesión.
// @Tags         Psicólogos - Auth
// @Accept       json
// @Produce      json
// @Param        request body request_structs.PsiLoginRequest true "Credenciales"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /psi/login [post]
func (h *PsiHandler) Login(c *fiber.Ctx) error {
	var req request_structs.PsiLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	token, user, err := h.service.Login(c.UserContext(), req.Identifier, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// ── Analytics: registrar login exitoso ───────────────────────────────────
	// Solo se ejecuta si las credenciales fueron válidas (err == nil)
	h.analytics.RecordLogin(
		user.ID,
		user.Username,
		"psi",
		c.IP(),
		c.Get("User-Agent"),
	)
	// ─────────────────────────────────────────────────────────────────────────

	return c.JSON(fiber.Map{
		"message": "Bienvenido colega",
		"token":   token,
	})
}

func (h *PsiHandler) LoginLibrary(c *fiber.Ctx) error {
	var req request_structs.PsiLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	token, err := h.service.LoginLibrary(c.UserContext(), req.Identifier, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Acceso a la biblioteca",
		"token":   token,
	})
}

// Logout godoc
// @Summary      Logout de Psicólogo
// @Description  Invalida el token activo rotando la key de firma. Elimina la sesión activa de las estadísticas.
// @Tags         Psicólogos - Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /psi/logout [post]
func (h *PsiHandler) Logout(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return err
	}

	if err := h.service.Logout(c.UserContext(), psi); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al cerrar sesión",
		})
	}

	// ── Analytics: eliminar sesión activa ────────────────────────────────────
	// Descuenta de active_sessions_now inmediatamente
	h.analytics.RecordLogout(psi.ID)
	// ─────────────────────────────────────────────────────────────────────────

	return c.JSON(fiber.Map{"message": "Sesión cerrada correctamente"})
}

// AddPostGrade godoc
// @Summary      Agregar postgrado con soportes
// @Security     BearerAuth
// @Tags         Psicólogos - Académico
// @Accept       multipart/form-data
// @Produce      json
// @Param        title           formData string true  "Título obtenido"
// @Param        university      formData string true  "Universidad"
// @Param        graduation_year formData string true  "Año de graduación"
// @Param        description     formData string false "Descripción opcional"
// @Param        pic_one         formData file  false  "Imagen del Título"
// @Param        pic_two         formData file  false  "Imagen de Notas"
// @Param        pic_three       formData file  false  "Imagen Extra"
// @Success      201 {object} map[string]string
// @Router       /psi/me/postgrades [post]
func (h *PsiHandler) AddPostGrade(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return err
	}

	var req request_structs.CreatePostGradeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos de formulario inválidos"})
	}

	file1, _ := c.FormFile("pic_one")
	file2, _ := c.FormFile("pic_two")
	file3, _ := c.FormFile("pic_three")
	files := []*multipart.FileHeader{file1, file2, file3}

	if err := h.service.AddPostGrade(c.UserContext(), psi, req, files); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Postgrado y documentos registrados exitosamente"})
}

func (h *PsiHandler) ListPsis(c *fiber.Ctx) error  { return nil }
func (h *PsiHandler) GetPsi(c *fiber.Ctx) error    { return nil }
func (h *PsiHandler) UpdatePsi(c *fiber.Ctx) error { return nil }

// UpdatePostGrade godoc
// @Summary      Actualizar postgrado
// @Security     BearerAuth
// @Tags         Psicólogos - Académico
// @Accept       multipart/form-data
// @Produce      json
// @Param        id              path     string true  "ID del Postgrado"
// @Param        title           formData string false "Título"
// @Param        university      formData string false "Universidad"
// @Param        graduation_year formData string false "Año"
// @Param        description     formData string false "Descripción"
// @Param        pic_one         formData file  false  "Reemplazar Imagen 1"
// @Param        pic_two         formData file  false  "Reemplazar Imagen 2"
// @Param        pic_three       formData file  false  "Reemplazar Imagen 3"
// @Success      200 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Router       /psi/me/postgrades/{id} [patch]
func (h *PsiHandler) UpdatePostGrade(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return err
	}

	pgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

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

	if err := h.service.UpdatePostGrade(c.UserContext(), psi, pgID, req, fileMap); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
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
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return err
	}
	var req request_structs.CreateSocialNetworkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}
	if err := h.service.AddSocialNetwork(c.UserContext(), psi, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Red social agregada"})
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
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return err
	}
	netID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}
	var req request_structs.UpdateSocialNetworkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}
	if err := h.service.UpdateSocialNetwork(c.UserContext(), psi, netID, req); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
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
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Red social eliminada correctamente"})
}

func (h *PsiHandler) GetSitemapData(c *fiber.Ctx) error {
	// Pedimos al repo solo los que deben ser indexados (Activos y Solventes)
	// Usamos el UserContext para que sea compatible con timeouts
	psis, err := h.service.GetSitemapPsis(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(psis)
}

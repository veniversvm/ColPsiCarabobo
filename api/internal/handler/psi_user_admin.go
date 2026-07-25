// api/internal/handler/psi_user_admin.go
package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	utils "github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// GetPsiByIDAdmin godoc
// @Summary      Ver expediente completo (Admin)
// @Description  Retorna TODA la información del psicólogo (Identidad, Privacidad, ColData, Postgrados y RRSS). Ignora configuraciones de privacidad del usuario.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Produce      json
// @Param        id   path      string  true  "UUID del Psicólogo"
// @Success      200  {object}  domain.PsiUserModel
// @Failure      400  {object}  map[string]string "ID inválido"
// @Failure      403  {object}  map[string]string "Permiso denegado"
// @Failure      404  {object}  map[string]string "Psicólogo no encontrado"
// @Router       /admin/psi/{id} [get]
func (h *PsiHandler) GetPsiByIDAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	// 2. Parsear el UUID solicitado
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	// 3. Consultar el servicio
	profile, err := h.service.GetPsiByIDAdmin(c.UserContext(), admin, targetID)
	if err != nil {
		if errors.Is(err, domain.ErrPsiNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	bio, err := h.service.GetPsiBioByID(c.UserContext(), profile.BioTextID)
	if err != nil {
		log.Printf("---- error al recuperar la BIO ----\n")
		log.Printf("%v\n", err)
	}

	solvencies, err := h.service.GetPsiSOlvency(c.UserContext(), targetID)
	if err != nil {
		log.Printf("---- error al recuperar la BIO ----\n")
		log.Printf("%v\n", err)
	}

	profile.FullBio.Content = bio
	profile.Solvencies = solvencies

	// 4. Retornar JSON (Password y Key se ocultan automáticamente por el json:"-" del struct)
	return c.JSON(profile)
}

// CreatePsiByAdmin godoc
// @Summary      Registrar psicólogo (Manual/Individual)
// @Description  Permite a un administrador crear un nuevo perfil de psicólogo junto con sus datos colegiales básicos en una sola transacción.
// @Tags         Administración - Psicólogos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body request_structs.CreatePsiAdminRequest true "Datos de identidad, contacto y registro colegial"
// @Success      201 {object} map[string]string "message: Psicólogo registrado con éxito"
// @Failure      400 {object} map[string]string "error: JSON malformado o validación fallida"
// @Failure      401 {object} map[string]string "error: No autorizado"
// @Failure      500 {object} map[string]string "error: Fallo interno al crear el registro"
// @Router       /admin/psi/create [post]
func (h *PsiHandler) CreatePsiByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return err
	}
	var req request_structs.CreatePsiAdminRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON malformado"})
	}

	if err := h.service.CreatePsiByAdmin(c.UserContext(), admin, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Psicólogo registrado con éxito"})
}

// UpdatePsiByAdmin godoc
// @Summary      Editar perfil completo de psicólogo
// @Description  Permite a un administrador modificar CUALQUIER campo del psicólogo (Identidad legal, solvencia, datos gremiales). Soporta actualizaciones parciales (PATCH).
// @Tags         Administración - Psicólogos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path string true "UUID del Psicólogo"
// @Param        request body request_structs.UpdatePsiAdminRequest true "Campos a modificar (solo enviar los necesarios)"
// @Success      200 {object} map[string]string "message: Perfil actualizado por administración"
// @Failure      400 {object} map[string]string "error: ID inválido o sin campos para actualizar"
// @Failure      403 {object} map[string]string "error: Permisos insuficientes para editar este registro"
// @Failure      404 {object} map[string]string "error: Registro no encontrado"
// @Router       /admin/psi/{id} [patch]
func (h *PsiHandler) UpdatePsiByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de psicólogo inválido"})
	}

	var req request_structs.UpdatePsiAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	// Archivos opcionales: si no se envían, c.FormFile retorna ErrNotFound y el valor es nil.
	// La validación de "request vacío" más abajo verifica que al menos un campo o archivo sea no-nil.
	profilePic, _ := c.FormFile("profile_picture")
	titleImgOne, _ := c.FormFile("title_image_one")
	titleImgTwo, _ := c.FormFile("title_image_two")
	titleImgThree, _ := c.FormFile("title_image_three")

	// Validación de "Request Vacío" para evitar llamadas innecesarias al servicio
	if utils.IsEmptyReq(req) && (profilePic == nil && titleImgOne == nil && titleImgTwo == nil && titleImgThree == nil) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No se proporcionaron campos para actualizar"})
	}

	if err := h.service.UpdatePsiByAdmin(
		c.UserContext(),
		admin,
		targetID,
		req,
		profilePic,
		titleImgOne,
		titleImgTwo,
		titleImgThree,
	); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Perfil actualizado por administración"})
}

// DeletePsiByAdmin godoc
// @Summary      Eliminar psicólogo (Soft Delete)
// @Description  Marca a un psicólogo como eliminado. El registro permanece en la base de datos para auditoría legal pero queda invisible en el sistema.
// @Tags         Administración - Psicólogos
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID del Psicólogo"
// @Success      200 {object} map[string]string "message: Psicólogo eliminado correctamente"
// @Failure      401 {object} map[string]string "error: No autorizado"
// @Failure      403 {object} map[string]string "error: Permisos insuficientes"
// @Failure      404 {object} map[string]string "error: Registro no encontrado"
// @Router       /admin/psi/{id} [delete]
func (h *PsiHandler) DeletePsiByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de psicólogo inválido"})
	}

	if err := h.service.DeletePsiByAdmin(c.UserContext(), admin, targetID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Psicólogo eliminado correctamente"})
}

// ListAllPsis godoc
// @Summary      Listado administrativo de psicólogos
// @Description  Buscador total de psicólogos. Ignora las reglas de solvencia e inactividad. Solo Admin.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Produce      json
// @Param        q          query     string  false  "Búsqueda por Nombre, CI, FPV"
// @Param        specialty  query     int     false  "ID Especialidad"
// @Param        location   query     string  false  "Municipio o Estado"
// @Param        gender     query     string  false  "Género (M/F)"
// @Param        page       query     int     false  "Página (Def: 1)"
// @Param        limit      query     int     false  "Límite (Def: 12)"
// @Success      200        {object}  map[string]interface{}
// @Failure      403        {object}  map[string]string
// @Router       /admin/psi/list [get]
func (h *PsiHandler) ListAllPsis(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	// Construir el filtro desde la URL
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "12"))
	specID, _ := strconv.Atoi(c.Query("specialty", "0"))

	filter := request_structs.PsiDirectoryFilterDTO{
		SearchTerm:  c.Query("q"),
		Location:    c.Query("location"),
		Gender:      c.Query("gender"),
		SpecialtyID: uint32(specID),
		Page:        page,
		Limit:       limit,
	}

	result, err := h.service.GetAdminDirectory(c.UserContext(), admin, filter)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// SpecialtyHandler gestiona el catálogo de especialidades psicológicas.
// Permite el acceso público para consulta y restringido para gestión administrativa.
type SpecialtyHandler struct {
	service *service.SpecialtyService
}

func NewSpecialtyHandler(svc *service.SpecialtyService) *SpecialtyHandler {
	return &SpecialtyHandler{service: svc}
}

// GetSpecialties godoc
// @Summary      Listar especialidades (Público/Admin)
// @Description  Retorna especialidades. Los usuarios ven solo las activas. Admins con permiso pueden filtrar por ?status=active|inactive|all.
// @Tags         Catálogos - Especialidades
// @Produce      json
// @Param        status   query     string  false  "Filtro administrativo: active, inactive, all (Default: active)"
// @Success      200      {array}   domain.PsiSpecialty
// @Failure      500      {object}  map[string]string
// @Router       /specialties [get]
func (h *SpecialtyHandler) GetSpecialties(c *fiber.Ctx) error {
	// A. Detectar nivel de acceso desde el middleware OptionalAdmin
	admin, isLogged := c.Locals("admin").(*domain.UserAdmin)
	isAdminWithPermissions := isLogged && (admin.Sudo || admin.CanReadNotifications)

	// B. Capturar parámetro de filtro
	requestedStatus := c.Query("status", "active")

	// C. Delegar al servicio
	list, err := h.service.GetSpecialties(c.UserContext(), requestedStatus, isAdminWithPermissions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al recuperar especialidades"})
	}

	return c.JSON(list)
}

// GetSpecialtyByID godoc
// @Summary      Obtener especialidad por ID
// @Description  Retorna el detalle de una especialidad específica.
// @Tags         Catálogos - Especialidades
// @Produce      json
// @Param        id   path      int  true  "ID de la especialidad"
// @Success      200  {object}  domain.PsiSpecialty
// @Failure      404  {object}  map[string]string
// @Router       /specialties/{id} [get]
func (h *SpecialtyHandler) GetSpecialtyByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	specialty, err := h.service.GetByID(c.UserContext(), uint32(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "especialidad no encontrada"})
	}

	return c.JSON(specialty)
}

// CreateSpecialty godoc
// @Summary      Crear especialidad
// @Description  Registra una nueva especialidad en el catálogo. Requiere permisos de gestión de etiquetas.
// @Security     BearerAuth
// @Tags         Administración - Especialidades
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.CreateSpecialtyRequest  true  "Datos de la especialidad"
// @Success      201      {object}  map[string]string "message"
// @Failure      403      {object}  map[string]string "error: permiso denegado"
// @Router       /specialties [post]
func (h *SpecialtyHandler) CreateSpecialty(c *fiber.Ctx) error {
	admin := c.Locals("admin").(*domain.UserAdmin)

	var req request_structs.CreateSpecialtyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.Create(c.UserContext(), admin, req); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Especialidad creada exitosamente"})
}

// UpdateSpecialty godoc
// @Summary      Actualizar especialidad
// @Description  Modifica una especialidad existente. Soporta cambios parciales y actualización de estado.
// @Security     BearerAuth
// @Tags         Administración - Especialidades
// @Accept       json
// @Produce      json
// @Param        id       path      int                                     true  "ID de la especialidad"
// @Param        request  body      request_structs.UpdateSpecialtyRequest  true  "Campos a modificar"
// @Success      200      {object}  map[string]string "message"
// @Router       /specialties/{id} [patch]
func (h *SpecialtyHandler) UpdateSpecialty(c *fiber.Ctx) error {
	admin := c.Locals("admin").(*domain.UserAdmin)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var req request_structs.UpdateSpecialtyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.Update(c.UserContext(), admin, uint32(id), req); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Especialidad actualizada correctamente"})
}

// DeleteSpecialty godoc
// @Summary      Eliminar especialidad (Desactivar)
// @Description  Realiza una desactivación lógica de la especialidad para mantener integridad referencial.
// @Security     BearerAuth
// @Tags         Administración - Especialidades
// @Param        id   path      int  true  "ID de la especialidad"
// @Success      200  {object}  map[string]string "message"
// @Router       /specialties/{id} [delete]
func (h *SpecialtyHandler) DeleteSpecialty(c *fiber.Ctx) error {
	admin := c.Locals("admin").(*domain.UserAdmin)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.Delete(c.UserContext(), admin, uint32(id)); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Especialidad desactivada"})
}

// CountSpecialties godoc
// @Summary      Contar total de especialidades
// @Description  Obtiene el conteo. Público: solo activas. Admin: permite filtrar por estado.
// @Tags         Catálogos - Especialidades
// @Produce      json
// @Param        active    query     bool  false  "Filtrar por activas (true) o inactivas (false). Omitir para ver todas (Solo Admins)."
// @Success      200  {object}  map[string]int64 "count"
// @Router       /specialties/count [get]
func (h *SpecialtyHandler) CountSpecialties(c *fiber.Ctx) error {
	var activePtr *bool
	if queryValue := c.Query("active"); queryValue != "" {
		active := c.QueryBool("active")
		activePtr = &active
	}

	// 1. Manejo seguro del local: si no existe, admin será nil
	admin, _ := c.Locals("admin").(*domain.UserAdmin)

	// 2. El servicio ya debe estar preparado para recibir admin == nil
	count, err := h.service.Count(c.UserContext(), activePtr, admin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"count": count})
}

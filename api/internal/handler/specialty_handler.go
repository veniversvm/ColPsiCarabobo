// api/internal/handler/specialty_handler.go
package handler

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// SpecialtyHandler gestiona el catálogo de especialidades psicológicas.
type SpecialtyHandler struct {
	service *service.SpecialtyService
}

func NewSpecialtyHandler(svc *service.SpecialtyService) *SpecialtyHandler {
	return &SpecialtyHandler{service: svc}
}

// GetSpecialties godoc
// @Summary      Listar especialidades (Público/Admin)
// @Description  Retorna el catálogo. Los usuarios ven solo las activas. Admins con permiso pueden filtrar por ?status=active|inactive|all.
// @Tags         Catálogos - Especialidades
// @Produce      json
// @Param        status   query     string  false  "Filtro administrativo: active, inactive, all (Default: active)"
// @Success      200      {array}   domain.PsiSpecialtyModel
// @Failure      500      {object}  map[string]string "error al recuperar especialidades"
// @Router       /specialties [get]
func (h *SpecialtyHandler) GetSpecialties(c *fiber.Ctx) error {
	admin, isLogged := c.Locals("admin").(*domain.UserAdmin)
	isAdminWithPermissions := isLogged && (admin.Sudo || admin.CanReadNotifications)

	requestedStatus := c.Query("status", "active")

	list, err := h.service.GetSpecialties(c.UserContext(), requestedStatus, isAdminWithPermissions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al recuperar especialidades"})
	}

	return c.JSON(list)
}

// GetSpecialtyByID godoc
// @Summary      Obtener especialidad por ID
// @Description  Retorna el detalle completo de una especialidad específica mediante su ID numérico.
// @Tags         Catálogos - Especialidades
// @Produce      json
// @Param        id   path      int  true  "ID único de la especialidad (uint32)"
// @Success      200  {object}  domain.PsiSpecialtyModel
// @Failure      400  {object}  map[string]string "ID inválido"
// @Failure      404  {object}  map[string]string "especialidad no encontrada"
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
// @Summary      Crear nueva especialidad
// @Description  Registra una especialidad. Solo accesible para administradores autorizados.
// @Security     BearerAuth
// @Tags         Administración - Especialidades
// @Accept       json
// @Produce      json
// @Param        request  body      request_structs.CreateSpecialtyRequest  true  "Datos de la nueva especialidad"
// @Success      201      {object}  map[string]string "message: Especialidad creada"
// @Failure      403      {object}  map[string]string "error: permiso denegado"
// @Router       /admin/specialties [post]
func (h *SpecialtyHandler) CreateSpecialty(c *fiber.Ctx) error {
	admin := c.Locals("admin").(*domain.UserAdmin)

	var req request_structs.CreateSpecialtyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.Create(c.UserContext(), admin, req); err != nil {
		if strings.Contains(err.Error(), "permiso") || strings.Contains(err.Error(), "rango") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Especialidad creada exitosamente"})
}

// UpdateSpecialty godoc
// @Summary      Actualizar especialidad
// @Description  Modifica campos de una especialidad. Actualiza automáticamente la auditoría (UpdateBy).
// @Security     BearerAuth
// @Tags         Administración - Especialidades
// @Accept       json
// @Produce      json
// @Param        id       path      int                                     true  "ID de la especialidad"
// @Param        request  body      request_structs.UpdateSpecialtyRequest  true  "Campos parciales a modificar"
// @Success      200      {object}  map[string]string "message: Especialidad actualizada"
// @Failure      403      {object}  map[string]string "error: permiso denegado"
// @Router       /admin/specialties/{id} [patch]
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
// @Summary      Eliminar especialidad (Soft-delete)
// @Description  Marca una especialidad como inactiva. Solo Admin.
// @Security     BearerAuth
// @Tags         Administración - Especialidades
// @Param        id   path      int  true  "ID de la especialidad"
// @Success      200  {object}  map[string]string "message: Especialidad desactivada"
// @Router       /admin/specialties/{id} [delete]
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
// @Description  Obtiene el número total. Público: solo activas. Admin: permite filtrar por ?active=true|false.
// @Tags         Catálogos - Especialidades
// @Produce      json
// @Param        active    query     bool  false  "Filtrar por estado (Solo Admins)"
// @Success      200       {object}  map[string]int64 "count"
// @Router       /specialties/count [get]
func (h *SpecialtyHandler) CountSpecialties(c *fiber.Ctx) error {
	admin, _ := c.Locals("admin").(*domain.UserAdmin)

	var activePtr *bool
	queryVal := c.Query("active")
	if queryVal != "" {
		active := queryVal == "true"
		activePtr = &active
	}

	count, err := h.service.Count(c.UserContext(), activePtr, admin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falló el conteo"})
	}

	return c.JSON(fiber.Map{"count": count})
}

// GetAllAdmin godoc
// @Summary      Listado administrativo total
// @Description  Retorna todas las especialidades sin filtros de estado. Uso para paneles de gestión.
// @Security     BearerAuth
// @Tags         Administración - Especialidades
// @Produce      json
// @Success      200      {array}   domain.PsiSpecialtyModel
// @Failure      403      {object}  map[string]string "error: permisos insuficientes"
// @Router       /admin/specialties/all [get]
func (h *SpecialtyHandler) GetAllAdmin(c *fiber.Ctx) error {
	log.Printf("[DEBUG] GetAllAdmin called")

	list, err := h.service.GetAllAdmin(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(list)
}

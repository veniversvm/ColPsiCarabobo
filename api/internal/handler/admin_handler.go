// api/internal/handler/admin_handler.go
package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// AdminHandler gestiona las peticiones HTTP relacionadas con la administración
// del sistema, incluyendo autenticación y gestión de personal administrativo.
type AdminHandler struct {
	service *service.AdminService
}

// NewAdminHandler inicializa un nuevo controlador de administración con su servicio.
func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{service: svc}
}

// LoginRequest define las credenciales necesarias para el acceso al sistema.
type LoginRequest struct {
	Identifier string `json:"identifier" example:"admin" validate:"required"`
	Password   string `json:"password" example:"admin123" validate:"required"`
}

// Login godoc
// @Summary      Iniciar sesión administrativo
// @Description  Valida las credenciales del administrador (email o username) y retorna un JWT dinámico.
// @Tags         Administración - Auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Credenciales de acceso"
// @Success      200      {object}  map[string]interface{} "message, token"
// @Failure      400      {object}  map[string]string      "error: formato inválido"
// @Failure      401      {object}  map[string]string      "error: credenciales inválidas"
// @Router       /auth/login [post]
func (h *AdminHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El formato del JSON es inválido",
		})
	}

	token, admin, err := h.service.Login(c.UserContext(), req.Identifier, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":              "Bienvenido al sistema",
		"token":                token,
		"must_change_password": admin.MustChangePassword,
	})
}

// Logout godoc
// @Summary      Cerrar sesión administrativo
// @Description  Invalida la sesión del administrador a nivel de servidor (Stateful Logout).
// @Tags         Administración - Auth
// @Security     BearerAuth
// @Success      200      {object}  map[string]string      "message: sesión cerrada"
// @Failure      401      {object}  map[string]string      "error: no autenticado"
// @Router       /auth/logout [post]
func (h *AdminHandler) Logout(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No autenticado"})
	}

	if err := h.service.Logout(c.UserContext(), admin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al cerrar sesión"})
	}

	return c.JSON(fiber.Map{"message": "Sesión cerrada correctamente"})
}

// ValidateSession verifican que la sesión del administrador siga siendo válida.
//
// @Summary      Validar sesión de administrador
// @Description  Devuelve 200 si el token JWT es válido y la sesión sigue activa;
//               devuelve 401 (via el middleware ProtectedAdmin) si fue revocada o expiró.
// @Tags         Administración - Sesión
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Router       /admin/validate [get]
func (h *AdminHandler) ValidateSession(c *fiber.Ctx) error {
	if _, err := middleware.GetAuthenticatedAdmin(c); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Sesión inválida o expirada."})
	}

	return c.JSON(fiber.Map{"valid": true})
}

// GetMe godoc
// @Summary      Estado y permisos del administrador autenticado
// @Description  Retorna identidad, rol y matriz de permisos del administrador actual. La UI lo usa para filtrar el menú; es informativo, nunca autoriza: el backend valida cada operación.
// @Tags         Administración - Sesión
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Router       /admin/me [get]
func (h *AdminHandler) GetMe(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Sesión inválida o expirada."})
	}

	return c.JSON(fiber.Map{
		"id":          admin.ID,
		"username":    admin.Username,
		"email":       admin.Email,
		"is_active":   admin.IsActive,
		"sudo":        admin.Sudo,
		"role":        admin.Role,
		"permissions": service.AdminPermissionSet(admin),
	})
}

// CreateAdmin godoc
// @Summary      Crear un nuevo administrador
// @Description  Registra un nuevo miembro del staff administrativo verificando la jerarquía de permisos.
// @Tags         Administración - Gestión
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      request_structs.CreateAdminRequest  true  "Datos del nuevo administrador y permisos"
// @Success      201      {object}  map[string]string      "message: creado correctamente"
// @Failure      400      {object}  map[string]string      "error: datos inválidos"
// @Failure      403      {object}  map[string]string      "error: violación de jerarquía"
// @Router       /admin/create [post]
func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	creator, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req request_structs.CreateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cuerpo de solicitud inválido"})
	}

	err = h.service.CreateAdmin(c.UserContext(), *creator, req)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Administrador creado correctamente"})
}

// GetAdmins godoc
// @Summary      Listar administradores
// @Description  Retorna una lista paginada de administradores con soporte para búsqueda y caché.
// @Tags         Administración - Gestión
// @Produce      json
// @Security     BearerAuth
// @Param        page    query     int     false  "Número de página (def: 1)"
// @Param        limit   query     int     false  "Registros por página (def: 10)"
// @Param        search  query     string  false  "Búsqueda por username o email"
// @Param        active  query     bool    false  "Filtrar por estado activo (true/false)"
// @Success      200     {object}  map[string]interface{} "data, total, page, limit"
// @Failure      500     {object}  map[string]string      "error: fallo interno"
// @Router       /admin/list [get]
func (h *AdminHandler) GetAdmins(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	var activePtr *bool
	if activeStr := c.Query("active"); activeStr != "" {
		active := activeStr == "true"
		activePtr = &active
	}

	result, err := h.service.GetAdmins(c.UserContext(), activePtr, search, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al recuperar administradores"})
	}

	return c.JSON(result)
}

// GetRolePresets godoc
// @Summary      Listar presets de roles del staff
// @Description  Retorna los perfiles de permisos predeterminados (Secretaría, Comunicación, Soporte, Proyectos, Lector) que la UI aplica como atajo al crear/editar personal. Los roles son solo metadato: la autorización siempre usa los flags booleanos individuales.
// @Tags         Administración - Gestión
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   service.RolePreset
// @Router       /admin/roles/presets [get]
func (h *AdminHandler) GetRolePresets(c *fiber.Ctx) error {
	return c.JSON(h.service.GetRolePresets())
}

// TransferSudo godoc
// @Summary      Transferir el rol de Super Usuario
// @Description  Permite al Sudo actual ceder su rol a un administrador de confianza. Requiere confirmar la contraseña del Sudo. La conmutación es atómica y queda registrada en la auditoría de permisos.
// @Tags         Administración - Gestión
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      request_structs.TransferSudoRequest  true  "Destinatario del Sudo y confirmación de la contraseña"
// @Success      200      {object}  map[string]string      "message: rol transferido"
// @Failure      403      {object}  map[string]string      "error: permiso denegado o contraseña incorrecta"
// @Router       /admin/transfer-sudo [post]
func (h *AdminHandler) TransferSudo(c *fiber.Ctx) error {
	current, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req request_structs.TransferSudoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El formato del JSON es inválido"})
	}

	if err := h.service.TransferSudo(c.UserContext(), current, req.TargetID, req.Password); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Rol de Super Usuario transferido correctamente"})
}

// UpdateAdmin godoc
// @Summary      Actualizar administrador
// @Description  Modifica los datos y permisos de un administrador. Actualiza automáticamente la auditoría.
// @Tags         Administración - Gestión
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      request_structs.UpdateAdminRequest  true  "Campos parciales a actualizar"
// @Success      200      {object}  map[string]string      "message: actualizado correctamente"
// @Failure      403      {object}  map[string]string      "error: permiso denegado"
// @Router       /admin/update [patch]
func (h *AdminHandler) UpdateAdmin(c *fiber.Ctx) error {
	updater, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req request_structs.UpdateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := h.service.UpdateAdmin(c.UserContext(), *updater, req); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Administrador actualizado correctamente"})
}

// DeleteAdmin godoc
// @Summary      Eliminar administrador (Soft-delete)
// @Description  Realiza un borrado lógico del administrador. No permite auto-eliminación ni borrar SUDO.
// @Tags         Administración - Gestión
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "UUID del administrador"
// @Success      200  {object}  map[string]string      "message: eliminado correctamente"
// @Failure      400  {object}  map[string]string      "error: ID inválido"
// @Failure      403  {object}  map[string]string      "error: permisos insuficientes"
// @Router       /admin/delete/{id} [delete]
func (h *AdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	updater, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de administrador inválido"})
	}

	if err := h.service.DeleteAdmin(c.UserContext(), updater, targetID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Administrador eliminado correctamente"})
}

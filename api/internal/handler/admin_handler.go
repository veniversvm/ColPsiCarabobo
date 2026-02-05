package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

type AdminHandler struct {
	service *service.AdminService
}

// NewAdminHandler inicializa el controlador de administración
func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{service: svc}
}

// LoginRequest define la estructura esperada para el inicio de sesión
type LoginRequest struct {
	Identifier string `json:"identifier" example:"admin@example.com"`
	Password   string `json:"password" example:"admin123"`
}

// Login godoc
// @Summary      Iniciar sesión como administrador
// @Description  Valida credenciales y genera un JWT con clave dinámica.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Credenciales de administrador"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /auth/login [post]
func (h *AdminHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	// 1. Parsear cuerpo de la petición
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El formato del JSON es inválido",
		})
	}

	// 2. Llamar al servicio (Ahora devuelve el token string)
	token, err := h.service.Login(c.UserContext(), req.Identifier, req.Password)
	if err != nil {
		// Retornamos 401 Unauthorized para errores de credenciales
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 3. Respuesta exitosa
	// Senior tip: Enviamos el token para Authorization y datos básicos para la UI
	return c.JSON(fiber.Map{
		"message": "Bienvenido al sistema",
		"token":   token, // El cliente debe guardar esto en localStorage/cookies
	})
}

func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	// 1. Obtener al admin que está operando (inyectado por el middleware)
	creator := c.Locals("admin").(*domain.UserAdmin)

	// 2. Parsear el request
	var req request_structs.CreateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "datos inválidos"})
	}

	// 3. Llamar al servicio
	err := h.service.CreateAdmin(c.UserContext(), creator, req)
	if err != nil {
		// Aquí devolvemos 403 Forbidden porque es una violación de permisos
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Administrador creado correctamente"})
}

func (h *AdminHandler) GetAdmins(c *fiber.Ctx) error {
	// Leer query params con valores por defecto
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
		return c.Status(500).JSON(fiber.Map{"error": "error al recuperar administradores"})
	}

	return c.JSON(result)
}

func (h *AdminHandler) UpdateAdmin(c *fiber.Ctx) error {
	updater := c.Locals("admin").(*domain.UserAdmin)

	var req request_structs.UpdateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	// El ID también puede venir del path /:id si prefieres,
	// aquí lo tomamos del body por consistencia con tu esquema.

	if err := h.service.UpdateAdmin(c.UserContext(), updater, req); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Administrador actualizado correctamente"})
}

func (h *AdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	updater := c.Locals("admin").(*domain.UserAdmin)

	// El ID puede venir por Query o por Params. Usaremos Params por estándar REST.
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID de administrador inválido"})
	}

	if err := h.service.DeleteAdmin(c.UserContext(), updater, targetID); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Administrador eliminado (soft-delete) correctamente"})
}

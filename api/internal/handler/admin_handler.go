package handler

import (
	"github.com/gofiber/fiber/v2"
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

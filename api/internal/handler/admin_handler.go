// api/internal/handler/admin_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

type AdminHandler struct {
	service *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{service: svc}
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // Username o Email
	Password   string `json:"password"`
}

func (h *AdminHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "formato de datos inválido"})
	}

	admin, err := h.service.Login(c.UserContext(), req.Identifier, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	// Por ahora devolvemos el admin.
	// El siguiente paso sería generar un JWT aquí.
	return c.JSON(fiber.Map{
		"message": "Login exitoso",
		"user": fiber.Map{
			"id":       admin.ID,
			"username": admin.Username,
			"email":    admin.Email,
			"sudo":     admin.Sudo,
		},
	})
}

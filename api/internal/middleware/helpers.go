package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// GetAuthenticatedAdmin extrae el admin de Fiber Locals de forma segura.
// Retorna error 401 si no hay sesión válida.
func GetAuthenticatedAdmin(c *fiber.Ctx) (*domain.UserAdmin, error) {
	admin, ok := c.Locals("admin").(*domain.UserAdmin)
	if !ok || admin == nil {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Sesión administrativa inválida o expirada",
		})
	}
	return admin, nil
}

// GetAuthenticatedPsi extrae el psicólogo de Fiber Locals de forma segura.
// Retorna error 401 si no hay sesión válida.
func GetAuthenticatedPsi(c *fiber.Ctx) (*domain.PsiUserModel, error) {
	psi, ok := c.Locals("psi_user").(*domain.PsiUserModel)
	if !ok || psi == nil {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Sesión de psicólogo inválida o expirada",
		})
	}
	return psi, nil
}

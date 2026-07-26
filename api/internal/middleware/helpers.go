package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// ErrNotAuthenticated se retorna cuando no hay un usuario autenticado en el contexto.
// A diferencia del patrón anterior, NO escribe respuestas HTTP como side-effect.
// El caller decide qué status retornar.
var ErrNotAuthenticated = errors.New("not authenticated")

// GetAuthenticatedAdmin extrae el admin de Fiber Locals de forma segura.
// Retorna ErrNotAuthenticated si no hay sesión válida (sin escribir HTTP).
func GetAuthenticatedAdmin(c *fiber.Ctx) (*domain.UserAdmin, error) {
	admin, ok := c.Locals("admin").(*domain.UserAdmin)
	if !ok || admin == nil {
		return nil, ErrNotAuthenticated
	}
	return admin, nil
}

// GetAuthenticatedPsi extrae el psicólogo de Fiber Locals de forma segura.
// Retorna ErrNotAuthenticated si no hay sesión válida (sin escribir HTTP).
func GetAuthenticatedPsi(c *fiber.Ctx) (*domain.PsiUserModel, error) {
	psi, ok := c.Locals("psi_user").(*domain.PsiUserModel)
	if !ok || psi == nil {
		return nil, ErrNotAuthenticated
	}
	return psi, nil
}

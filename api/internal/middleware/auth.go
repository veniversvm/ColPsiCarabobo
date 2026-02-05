package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// AuthMiddleware centraliza la validación de tokens dinámicos para diferentes roles
type AuthMiddleware struct {
	adminRepo domain.UserAdminRepository
	psiRepo   domain.PsiUserRepository
}

func NewAuthMiddleware(a domain.UserAdminRepository, p domain.PsiUserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		adminRepo: a,
		psiRepo:   p,
	}
}

// jwtError estandariza las respuestas de error
func jwtError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

// validateToken es una función interna genérica para evitar repetir la lógica de parseo
func (m *AuthMiddleware) validateToken(c *fiber.Ctx, getKeyFunc func(string) (string, error)) (*jwt.Token, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("missing or malformed JWT")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("user_id not found in token")
		}

		// Buscamos la clave dinámica usando la función proveída
		key, err := getKeyFunc(userID)
		if err != nil {
			return nil, err
		}

		return []byte(key), nil
	})
}

// ProtectedAdmin protege rutas que solo el staff administrativo puede usar
func (m *AuthMiddleware) ProtectedAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := m.validateToken(c, func(userID string) (string, error) {
			uid, _ := uuid.Parse(userID)
			admin, err := m.adminRepo.GetByIdentifier(c.UserContext(), uid.String())
			if err != nil {
				return "", err
			}
			// Inyectamos el admin en el contexto para ahorrar consultas en el handler
			c.Locals("admin", admin)
			return admin.Key, nil
		})

		if err != nil || !token.Valid {
			return jwtError(c, fiber.StatusUnauthorized, "Invalid or expired Admin Session")
		}
		return c.Next()
	}
}

// ProtectedPsiUser protege rutas exclusivas para psicólogos colegiados
func (m *AuthMiddleware) ProtectedPsiUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := m.validateToken(c, func(userID string) (string, error) {
			uid, _ := uuid.Parse(userID)
			psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
			if err != nil {
				return "", err
			}
			// Inyectamos el psicólogo en el contexto
			c.Locals("psi_user", psi)
			return psi.Key, nil
		})

		if err != nil || !token.Valid {
			return jwtError(c, fiber.StatusUnauthorized, "Invalid or expired Psychologist Session")
		}
		return c.Next()
	}
}

// ProtectedAdmin404 protege la ruta pero devuelve 404 si falla para ocultar el endpoint
func (m *AuthMiddleware) ProtectedAdmin404() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := m.validateToken(c, func(userID string) (string, error) {
			uid, _ := uuid.Parse(userID)
			admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
			if err != nil {
				return "", err
			}
			// Importante: Guardamos el admin en Locals para el Handler
			c.Locals("admin", admin)
			return admin.Key, nil
		})

		// Si hay error en el token o el admin no existe
		if err != nil || !token.Valid {
			// Devolvemos el mismo formato de error que un 404 estándar de Fiber
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
			})
		}

		return c.Next()
	}
}

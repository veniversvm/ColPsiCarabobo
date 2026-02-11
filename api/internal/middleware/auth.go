package middleware

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

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

func jwtError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

func (m *AuthMiddleware) validateToken(c *fiber.Ctx, getKeyFunc func(string) (string, error)) (*jwt.Token, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		log.Println("[DEBUG AUTH] Error: Cabecera Authorization vacía")
		return nil, errors.New("missing JWT")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		log.Println("[DEBUG AUTH] Error: Formato de cabecera inválido (falta Bearer)")
		return nil, errors.New("malformed JWT")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validar algoritmo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[DEBUG AUTH] Error: Algoritmo inesperado: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("[DEBUG AUTH] Error: No se pudieron parsear los claims")
			return nil, errors.New("invalid claims")
		}

		// IMPORTANTE: Verifica que en el LoginService uses "user_id"
		userID, ok := claims["user_id"].(string)
		if !ok {
			log.Println("[DEBUG AUTH] Error: claim 'user_id' no encontrado en el token")
			return nil, errors.New("user_id not found")
		}

		log.Printf("[DEBUG AUTH] Buscando Key para user_id: %s", userID)
		key, err := getKeyFunc(userID)
		if err != nil {
			log.Printf("[DEBUG AUTH] Error: No se encontró la Key en la DB para el usuario: %v", err)
			return nil, err
		}

		log.Println("[DEBUG AUTH] Key recuperada exitosamente. Verificando firma...")
		return []byte(key), nil
	})
}

func (m *AuthMiddleware) ProtectedAdmin404() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Printf("[DEBUG AUTH] --- Nueva petición a ruta protegida: %s ---", c.Path())

		token, err := m.validateToken(c, func(userID string) (string, error) {
			uid, err := uuid.Parse(userID)
			if err != nil {
				return "", err
			}
			admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
			if err != nil {
				return "", err
			}
			c.Locals("admin", admin)
			return admin.Key, nil
		})

		if err != nil {
			log.Printf("[DEBUG AUTH] Error de validación JWT: %v", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
			})
		}

		if !token.Valid {
			log.Println("[DEBUG AUTH] El token es inválido (expirado o firma incorrecta)")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
			})
		}

		log.Println("[DEBUG AUTH] Acceso concedido.")
		return c.Next()
	}
}

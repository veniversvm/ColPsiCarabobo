//	api/internal/middleware/auth.go

// Package middleware contiene los interceptores de peticiones para lógica transversal
// como seguridad, trazabilidad y manejo de errores.
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

// AuthMiddleware centraliza la lógica de autenticación y autorización.
// Utiliza inyección de dependencias para acceder a los repositorios de usuarios.
type AuthMiddleware struct {
	adminRepo domain.UserAdminRepository
	psiRepo   domain.PsiUserRepository
}

// NewAuthMiddleware inicializa el middleware con los repositorios necesarios.
func NewAuthMiddleware(a domain.UserAdminRepository, p domain.PsiUserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		adminRepo: a,
		psiRepo:   p,
	}
}

// jwtError es un helper interno para estandarizar las respuestas de error en formato JSON.
func jwtError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

// validateToken realiza el "heavy lifting" de la validación JWT.
// 1. Extrae el token del header Authorization.
// 2. Valida el algoritmo de firma (HS256).
// 3. Ejecuta un callback (getKeyFunc) para obtener la clave secreta dinámica del usuario desde la DB.
func (m *AuthMiddleware) validateToken(c *fiber.Ctx, getKeyFunc func(string) (string, error)) (*jwt.Token, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		log.Println("[DEBUG AUTH] Error: Cabecera Authorization vacía")
		return nil, errors.New("missing JWT")
	}

	// El estándar RFC 6750 requiere el prefijo "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		log.Println("[DEBUG AUTH] Error: Formato de cabecera inválido (falta Bearer)")
		return nil, errors.New("malformed JWT")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificación de integridad del método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[DEBUG AUTH] Error: Algoritmo inesperado: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid claims")
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("user_id not found")
		}

		// Invocamos la función de búsqueda para obtener la 'Key' específica de este usuario
		key, err := getKeyFunc(userID)
		if err != nil {
			return nil, err
		}

		return []byte(key), nil
	})
}

// ProtectedAdmin404 protege rutas administrativas.
// Si la autenticación falla, devuelve 404 Not Found en lugar de 401.
// Estrategia: "Security by Obscurity" para evitar el reconocimiento de endpoints privados.
func (m *AuthMiddleware) ProtectedAdmin404() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := m.validateToken(c, func(userID string) (string, error) {
			uid, err := uuid.Parse(userID)
			if err != nil {
				return "", err
			}
			// Propagamos el contexto de usuario para trazabilidad y cancelación
			admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
			if err != nil {
				return "", err
			}
			// Inyectamos el modelo completo para evitar re-consultas en los handlers
			c.Locals("admin", admin)
			return admin.Key, nil
		})

		// Si el token es inválido o el usuario no existe, devolvemos 404 estandarizado
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
			})
		}

		return c.Next()
	}
}

// OptionalHybridAuth es un middleware no-bloqueante.
// Identifica si quien llama es un Admin, un Psicólogo o un visitante anónimo.
// Útil para endpoints públicos donde el contenido se expande si el usuario está logueado.
func (m *AuthMiddleware) OptionalHybridAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next() // Usuario anónimo, continuar
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Intentamos identificar al usuario sin abortar la petición en caso de error
		_, _ = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, nil
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, nil
			}

			userID, _ := claims["user_id"].(string)
			role, _ := claims["role"].(string)
			uid, _ := uuid.Parse(userID)

			if role == "admin" {
				admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
				if err == nil {
					c.Locals("admin", admin)
					return []byte(admin.Key), nil
				}
			} else {
				psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
				if err == nil {
					c.Locals("psi_user", psi)
					return []byte(psi.Key), nil
				}
			}
			return nil, nil
		})

		return c.Next()
	}
}

// ProtectedPsiUser garantiza que la petición provenga de un psicólogo autenticado.
// Utiliza la Key dinámica almacenada en la tabla de psicólogos para la validación.
func (m *AuthMiddleware) ProtectedPsiUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := m.validateToken(c, func(userID string) (string, error) {
			uid, err := uuid.Parse(userID)
			if err != nil {
				return "", err
			}

			psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
			if err != nil {
				return "", err
			}

			c.Locals("psi_user", psi)
			return psi.Key, nil
		})

		if err != nil || !token.Valid {
			return jwtError(c, fiber.StatusUnauthorized, "Sesión inválida o expirada.")
		}

		return c.Next()
	}
}

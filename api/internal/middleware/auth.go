// api/internal/middleware/auth.go

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
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// ── Struct: se añade analyticsService ────────────────────────────────────────
type AuthMiddleware struct {
	adminRepo domain.UserAdminRepository
	psiRepo   domain.PsiUserRepository
	analytics *service.AnalyticsService // 👈 NUEVO
}

// ── Constructor: acepta el tercer parámetro ───────────────────────────────────
func NewAuthMiddleware(a domain.UserAdminRepository, p domain.PsiUserRepository, analytics *service.AnalyticsService) *AuthMiddleware {
	return &AuthMiddleware{
		adminRepo: a,
		psiRepo:   p,
		analytics: analytics, // 👈 NUEVO
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

		key, err := getKeyFunc(userID)
		if err != nil {
			return nil, err
		}

		return []byte(key), nil
	})
}

// ProtectedAdmin404 — sin cambios en comportamiento, añade heartbeat al éxito
func (m *AuthMiddleware) ProtectedAdmin404() fiber.Handler {
	return func(c *fiber.Ctx) error {
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

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
			})
		}

		// ── Heartbeat: renovar sesión activa del admin ────────────────────────
		// Solo si la validación fue exitosa — asíncrono, no bloquea la request
		if admin, ok := c.Locals("admin").(*domain.UserAdmin); ok && admin != nil {
			m.analytics.HeartbeatSession(admin.ID)
		}
		// ─────────────────────────────────────────────────────────────────────

		return c.Next()
	}
}

// OptionalHybridAuth — sin cambios, no aplica heartbeat (no es ruta protegida)
func (m *AuthMiddleware) OptionalHybridAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

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

// ProtectedPsiUser — añade heartbeat al éxito
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

		// ── Heartbeat: renovar sesión activa del psicólogo ────────────────────
		// Solo si la validación fue exitosa — asíncrono, no bloquea la request
		if psi, ok := c.Locals("psi_user").(*domain.PsiUserModel); ok && psi != nil {
			m.analytics.HeartbeatSession(psi.ID)
		}
		// ─────────────────────────────────────────────────────────────────────

		return c.Next()
	}
}

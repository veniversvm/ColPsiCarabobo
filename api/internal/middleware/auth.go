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

// OptionalHybridAuth intenta autenticar al usuario (Admin o Psicólogo), pero NO bloquea la petición si falla.
// Es ideal para rutas "mixtas" (ej: Listado de Noticias) donde el contenido varía según quién lo ve.
func (m *AuthMiddleware) OptionalHybridAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		// 1. Si no hay cabecera, es un visitante anónimo. Continuamos sin error.
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Intentamos parsear el token
		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validar algoritmo
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritmo inválido")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, errors.New("claims inválidos")
			}

			// 3. Extraer Identidad y Rol
			userID, _ := claims["user_id"].(string)
			role, _ := claims["role"].(string) // "admin" o "psi"

			uid, err := uuid.Parse(userID)
			if err != nil {
				return nil, err
			}

			// 4. Búsqueda polimórfica según el rol
			if role == "admin" {
				// Es Admin: Buscamos en tabla de admins
				admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
				if err != nil {
					return nil, err
				}

				// Inyectamos y devolvemos su Key
				c.Locals("admin", admin)
				return []byte(admin.Key), nil

			} else {
				// Asumimos Psicólogo (rol="psi" o vacío)
				psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
				if err != nil {
					return nil, err
				}

				// Inyectamos y devolvemos su Key
				c.Locals("psi_user", psi)
				return []byte(psi.Key), nil
			}
		})

		// 5. Decisión final:
		// No nos importa si el token dio error (expirado, firma mala, usuario borrado).
		// Si 'token.Valid' es false, simplemente no habremos inyectado nada en c.Locals.
		// El Handler recibirá la petición como si fuera un usuario anónimo.

		// Opcional: Podrías loguear advertencias aquí si quieres debug
		if token != nil && !token.Valid {
			// log.Println("[AUTH] Token inválido en ruta híbrida, tratando como anónimo")
		}

		return c.Next()
	}
}

// ProtectedPsiUser verifica que el token pertenezca a un psicólogo colegiado.
// Extrae la Key dinámica de la base de datos para validar la firma y previene el uso de sesiones antiguas.
func (m *AuthMiddleware) ProtectedPsiUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Reutilizamos la función genérica validateToken
		token, err := m.validateToken(c, func(userID string) (string, error) {
			uid, err := uuid.Parse(userID)
			if err != nil {
				return "", err
			}

			// Buscamos en el repositorio de Psicólogos (no en el de Admins)
			psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
			if err != nil {
				return "", err
			}

			// Inyectamos el objeto completo del psicólogo en el contexto
			// para que los handlers (/psi/me) no tengan que volver a buscarlo en la BD.
			c.Locals("psi_user", psi)

			// Retornamos su Key personal para verificar la firma del JWT
			return psi.Key, nil
		})

		// Si el token falló la validación, fue alterado o el usuario ya no existe
		if err != nil || !token.Valid {
			log.Printf(" Error de validación JWT: %v", err)
			return jwtError(c, fiber.StatusUnauthorized, "Sesión inválida o expirada. Por favor, inicie sesión nuevamente.")
		}

		// Si todo está bien, pasamos al siguiente Handler
		return c.Next()
	}
}

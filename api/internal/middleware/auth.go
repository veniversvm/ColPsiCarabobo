// api/internal/middleware/auth.go

// Package middleware contiene los interceptores HTTP responsables de la seguridad perimetral.
//
// Actúa como el "Gatekeeper" de la API, validando firmas criptográficas (JWT),
// mitigando ataques de suplantación, inyectando el contexto de identidad (Fiber Locals)
// y disparando eventos de telemetría sin bloquear el hilo principal de ejecución.
package middleware

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// AuthMiddleware encapsula las dependencias necesarias para validar sesiones.
// Utiliza inyección de dependencias para acceder a los repositorios de usuarios
// y al motor de analíticas.
type AuthMiddleware struct {
	adminRepo domain.UserAdminRepository
	psiRepo   domain.PsiUserRepository
	analytics *service.AnalyticsService
}

// NewAuthMiddleware construye el middleware de autenticación.
// Garantiza que los handlers HTTP tengan acceso a la base de datos y a la telemetría
// de forma segura y concurrente.
func NewAuthMiddleware(a domain.UserAdminRepository, p domain.PsiUserRepository, analytics *service.AnalyticsService) *AuthMiddleware {
	return &AuthMiddleware{
		adminRepo: a,
		psiRepo:   p,
		analytics: analytics,
	}
}

// jwtError es un helper interno para estandarizar las respuestas JSON
// ante fallos de autorización, evitando exponer stack traces al cliente.
func jwtError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

// validateToken es el motor criptográfico principal del middleware.
//
// Diseño de Seguridad (Invalidación de Sesiones):
// A diferencia de los JWT tradicionales que usan un único secreto global, este método
// recibe un 'getKeyFunc' que consulta la base de datos para obtener el secreto (Key)
// Específico de ese usuario. Si un usuario cambia su contraseña, su 'Key' en la DB cambia,
// lo que invalida matemática e instantáneamente todos sus JWTs emitidos previamente.
//
// Mitigación CVE (None Algorithm Attack):
// Verifica estrictamente que el token fue firmado con HMAC, previniendo que un
// atacante envíe un token con cabecera `{"alg": "none"}` para saltarse la validación.
func (m *AuthMiddleware) validateToken(c *fiber.Ctx, getKeyFunc func(string) (string, error)) (*jwt.Token, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		if config.Envs != nil && config.Envs.Environment == "development" {
			log.Debug().Str("component", "auth").Msg("Error: Cabecera Authorization vacía")
		}
		return nil, errors.New("missing JWT")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		if config.Envs != nil && config.Envs.Environment == "development" {
			log.Debug().Str("component", "auth").Msg("Error: Formato de cabecera inválido (falta Bearer)")
		}
		return nil, errors.New("malformed JWT")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificación estricta del algoritmo de firma (Defensa contra vulnerabilidades de librerías JWT)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			if config.Envs != nil && config.Envs.Environment == "development" {
				log.Debug().Str("component", "auth").Interface("alg", token.Header["alg"]).Msg("Algoritmo de firma inesperado")
			}
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

		// Resolución Dinámica del Secreto (Permite Revocación de Tokens)
		key, err := getKeyFunc(userID)
		if err != nil {
			return nil, err
		}

		return []byte(key), nil
	})
}

// ProtectedAdmin404 protege las rutas exclusivas del panel de administración.
//
// Táctica de Seguridad por Oscuridad (Security by Obscurity):
// En lugar de devolver un HTTP 401 (Unauthorized), simula que la ruta no existe (404).
// Esto frustra a los escáneres automáticos de vulnerabilidades, impidiendo que
// descubran la topología de las URLs del panel administrativo administrativo.
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
			// Inyección de Contexto: Permite a los controladores acceder al admin
			// sin tener que hacer una segunda consulta a la base de datos.
			c.Locals("admin", admin)
			if admin.Key == "" {
				return "", errors.New("session expired")
			}
			return admin.Key, nil
		})

		if err != nil || !token.Valid {
			// Enmascaramiento de la respuesta (404 Not Found)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
			})
		}

		// ── Heartbeat (Telemetría Asíncrona) ──────────────────────────────────
		// Renueva la estampa de "Última vez visto" (Active Session) del administrador.
		// Al delegarlo al servicio de analíticas, no bloqueamos la petición HTTP actual.
		if admin, ok := c.Locals("admin").(*domain.UserAdmin); ok && admin != nil {
			m.analytics.HeartbeatSession(admin.ID)
		}
		// ─────────────────────────────────────────────────────────────────────

		return c.Next()
	}
}

// ProtectedAdmin protege una ruta exclusiva del panel de administración pero,
// a diferencia de ProtectedAdmin404, devuelve un 401 (Unauthorized) explícito
// en lugar de enmascarar la respuesta como 404.
//
// Está pensada para endpoints de validación de sesión del lado del cliente
// (p. ej. GET /session/validate), donde el frontend necesita distinguir con
// claridad "sesión inválida/revocada" (401) de "ruta inexistente".
func (m *AuthMiddleware) ProtectedAdmin() fiber.Handler {
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
			// Inyección de Contexto: Permite a los controladores acceder al admin
			// sin tener que hacer una segunda consulta a la base de datos.
			c.Locals("admin", admin)
			if admin.Key == "" {
				return "", errors.New("session expired")
			}
			return admin.Key, nil
		})

		if err != nil || !token.Valid {
			return jwtError(c, fiber.StatusUnauthorized, "Sesión inválida o expirada.")
		}

		// ── Heartbeat (Telemetría Asíncrona) ──────────────────────────────────
		// Renueva la estampa de "Última vez visto" (Active Session) del administrador.
		if admin, ok := c.Locals("admin").(*domain.UserAdmin); ok && admin != nil {
			m.analytics.HeartbeatSession(admin.ID)
		}
		// ─────────────────────────────────────────────────────────────────────

		return c.Next()
	}
}

// OptionalHybridAuth implementa un patrón de "Soft Authentication" (Autenticación Híbrida).
//
// Es un middleware no bloqueante. Intenta decodificar el token y extraer la identidad
// (sea de Administrador o de Psicólogo). Si falla o no hay token, permite que la
// petición continúe su curso normal hacia el controlador.
// Es ideal para endpoints públicos que muestran datos adicionales si el visitante
// resulta ser un usuario autenticado (Graceful Degradation).
//
// Seguridad: La validación de firma SEPARADA de los side-effects garantiza que
// un token con firma inválida NUNCA inyecte identidad en c.Locals().
func (m *AuthMiddleware) OptionalHybridAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		// Si no hay cabecera, se trata como un visitante anónimo.
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// La fila que hay que leer para obtener la key de firma es la MISMA que
		// después hay que inyectar en el PASO 3: mismo uid, misma petición. Se
		// memoiza para no hacer dos SELECT idénticos — este middleware ahora
		// corre en las rutas públicas GET (para que la telemetría pueda excluir
		// al staff) y el coste se nota. El anónimo nunca llega aquí: cortocircuita
		// en la cabecera de arriba sin tocar la base.
		//
		// Reutilizar la fila ya verificada además es MÁS consistente que releerla:
		// si la segunda lectura fallara, la identidad NO se inyectaría pese a que
		// la firma se comprobó bien, y el staff contaría como visitante sin que
		// nada fallara.
		var memoAdmin *domain.UserAdmin
		var memoPsi *domain.PsiUserModel

		// PASO 1: Parse + Verificar firma (sin side effects).
		// El keyFunc retorna la clave HMAC pero NO inyecta en c.Locals().
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Method)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, fmt.Errorf("claims inválidos")
			}

			userID, _ := claims["user_id"].(string)
			role, _ := claims["role"].(string)
			uid, err := uuid.Parse(userID)
			if err != nil {
				return nil, fmt.Errorf("user_id inválido: %w", err)
			}

			// Resolución dinámica del secreto (necesario para verificación HMAC).
			// NO se inyecta en c.Locals() aquí — solo se retorna la key.
			if role == "admin" {
				admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
				if err != nil {
					return nil, fmt.Errorf("admin no encontrado: %w", err)
				}
				memoAdmin = admin
				return []byte(admin.Key), nil
			}
			psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
			if err != nil {
				return nil, fmt.Errorf("psi no encontrado: %w", err)
			}
			memoPsi = psi
			return []byte(psi.Key), nil
		})

		// PASO 2: Verificar que el token sea válido ANTES de inyectar identidad.
		if err != nil || !token.Valid {
			return c.Next() // token inválido → proceder como anónimo
		}

		// PASO 3: Ahora SÍ inyectar identidad (el token ya pasó verificación de
		// firma). Se decide por el memo y no por releer el claim "role": el memo
		// es lo que verificó la firma, así que no puede contradecirla, mientras
		// que releer el claim es una segunda noción de "quién es" que deriva sola.
		// Si el token es válido, la keyFunc devolvió una key, así que exactamente
		// uno de los dos memos quedó poblado.
		switch {
		case memoAdmin != nil:
			c.Locals("admin", memoAdmin)
		case memoPsi != nil:
			c.Locals("psi_user", memoPsi)
		}

		return c.Next()
	}
}

// ProtectedAnyAuth protege los recursos compartidos entre ambos paneles
// (p. ej. los manuales PDF). Acepta una sesión válida de ADMINISTRADOR o de
// PSICÓLOGO y responde 401 explícito cuando el token falta, es inválido o
// expiró.
//
// NO reutiliza OptionalHybridAuth: ese middleware termina con c.Next(), y
// anidarlo aquí avanzaría el stack de Fiber hasta el handler y luego mi
// wrapper avanzaría OTRA vez (hasta el catch-all), pisando la respuesta —
// reproducido con Fiber v2.52. Validación autocontenida, igual que
// ProtectedAdmin / ProtectedPsiUser.
func (m *AuthMiddleware) ProtectedAnyAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return jwtError(c, fiber.StatusUnauthorized, "Acceso no autorizado.")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Method)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, fmt.Errorf("claims inválidos")
			}

			userID, _ := claims["user_id"].(string)
			role, _ := claims["role"].(string)
			uid, err := uuid.Parse(userID)
			if err != nil {
				return nil, fmt.Errorf("user_id inválido: %w", err)
			}

			// Resolución dinámica del secreto + inyección de identidad,
			// solo DESPUÉS de verificar la firma (sin side-effects antes).
			if role == "admin" {
				admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
				if err != nil {
					return nil, fmt.Errorf("admin no encontrado: %w", err)
				}
				if admin.Key == "" {
					return nil, errors.New("session expired")
				}
				c.Locals("admin", admin)
				return []byte(admin.Key), nil
			}

			psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
			if err != nil {
				return nil, fmt.Errorf("psi no encontrado: %w", err)
			}
			if psi.Key == "" {
				return nil, errors.New("session expired")
			}
			c.Locals("psi_user", psi)
			return []byte(psi.Key), nil
		})

		if err != nil || !token.Valid {
			return jwtError(c, fiber.StatusUnauthorized, "Acceso no autorizado.")
		}

		return c.Next()
	}
}

// ProtectedPsiUser protege las rutas de autogestión de los psicólogos colegiados.
//
// Implementa el bloqueo estándar HTTP 401 (Unauthorized) cuando las credenciales
// son inválidas, faltantes o han expirado.
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

			// Inyección de Contexto en el ciclo de vida de Fiber
			c.Locals("psi_user", psi)
			if psi.Key == "" {
				return "", errors.New("session expired")
			}
			return psi.Key, nil
		})

		if err != nil || !token.Valid {
			return jwtError(c, fiber.StatusUnauthorized, "Sesión inválida o expirada.")
		}

		// ── Heartbeat (Telemetría Asíncrona) ──────────────────────────────────
		// Registra la actividad del psicólogo para los reportes de "Usuarios Activos".
		if psi, ok := c.Locals("psi_user").(*domain.PsiUserModel); ok && psi != nil {
			m.analytics.HeartbeatSession(psi.ID)
		}
		// ─────────────────────────────────────────────────────────────────────

		return c.Next()
	}
}

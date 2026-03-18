// api/internal/middleware/idempotency.go
package middleware

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// entry guarda la respuesta cacheada y cuándo expira
type entry struct {
	status  int
	body    []byte
	expires time.Time
}

// IdempotencyStore es un cache en memoria con TTL.
// Para múltiples instancias reemplazar por Redis.
type IdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]entry
}

func NewIdempotencyStore() *IdempotencyStore {
	s := &IdempotencyStore{
		entries: make(map[string]entry),
	}
	// Limpieza periódica de keys expiradas
	go s.cleanup()
	return s
}

func (s *IdempotencyStore) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, e := range s.entries {
			if now.After(e.expires) {
				delete(s.entries, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *IdempotencyStore) get(key string) (entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	if !ok || time.Now().After(e.expires) {
		return entry{}, false
	}
	return e, true
}

func (s *IdempotencyStore) set(key string, status int, body []byte, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = entry{
		status:  status,
		body:    body,
		expires: time.Now().Add(ttl),
	}
}

// ─────────────────────────────────────────────────────────────────────────────

// UserScopedIdempotency devuelve un middleware que:
//  1. Lee X-Idempotency-Key del header
//  2. Vincula la key al user ID del admin autenticado (evita reuso entre usuarios)
//  3. Si ya existe respuesta cacheada para esa key+usuario → la devuelve directamente
//  4. Si no → ejecuta el handler, cachea la respuesta y la devuelve
//
// TTL recomendado: 30 minutos (tiempo razonable para reintentos)
func UserScopedIdempotency(store *IdempotencyStore, ttl time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawKey := c.Get("X-Idempotency-Key")

		// Si no viene el header, pasar sin cachear (comportamiento normal)
		if rawKey == "" {
			return c.Next()
		}

		// Obtener el admin autenticado del contexto (seteado por ProtectedAdmin404)
		admin, ok := c.Locals("admin").(*domain.UserAdmin)
		if !ok || admin == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No autenticado.",
			})
		}

		// Construir una key compuesta: hash(userID + rawKey)
		// Así la misma rawKey de otro usuario no colisiona
		scopedKey := scopeKey(admin.ID.String(), rawKey)

		// ── Cache hit ────────────────────────────────────────────────────────
		if cached, found := store.get(scopedKey); found {
			c.Set("X-Idempotent-Replayed", "true")
			return c.Status(cached.status).Send(cached.body)
		}

		// ── Cache miss — ejecutar handler ────────────────────────────────────
		if err := c.Next(); err != nil {
			return err
		}

		// Solo cachear respuestas exitosas (2xx)
		if c.Response().StatusCode() >= 200 && c.Response().StatusCode() < 300 {
			store.set(scopedKey, c.Response().StatusCode(), c.Response().Body(), ttl)
		}

		return nil
	}
}

// scopeKey genera un hash determinista de userID+rawKey
func scopeKey(userID, rawKey string) string {
	h := sha256.Sum256([]byte(userID + ":" + rawKey))
	return fmt.Sprintf("%x", h)
}

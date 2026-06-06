// api/internal/middleware/idempotency.go

// Package middleware contiene los interceptores de la API.
//
// CONCEPTO DE IDEMPOTENCIA:
// En redes inestables (como conexiones móviles), un cliente puede enviar una petición POST
// (ej. "Crear Administrador"), no recibir la respuesta a tiempo por un micro-corte,
// y reintentar la misma petición. Sin idempotencia, el servidor crearía dos administradores.
// Este middleware garantiza que, si el cliente envía la misma cabecera 'X-Idempotency-Key',
// el servidor procesará la operación solo la primera vez, y en los reintentos simplemente
// devolverá la misma respuesta exacta que generó originalmente, sin volver a tocar la base de datos.
package middleware

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// entry es la estructura de almacenamiento interno para el caché.
// Guarda el estado exacto de la respuesta HTTP para poder "repetirla" (Replay) al cliente.
type entry struct {
	status  int       // Código HTTP resultante (ej. 201 Created)
	body    []byte    // El Payload JSON de respuesta ya procesado
	expires time.Time // Sello de tiempo matemático para recolección de basura (TTL)
}

// IdempotencyStore es un motor de almacenamiento en memoria clave-valor (In-Memory KV Store).
//
// Diseño de Concurrencia (Thread-Safety):
// Dado que los servidores web en Go manejan miles de peticiones simultáneas usando Goroutines,
// leer o escribir en un mapa (`map`) nativo sin protección causaría un "Data Race" y un Panic inmediato.
// Por ello, se utiliza `sync.RWMutex` (Read-Write Mutex) para bloquear el acceso seguro a la memoria.
//
// Nota de Escalabilidad: Este diseño es perfecto para un monolito (1 servidor).
// Si el sistema se escala horizontalmente (múltiples servidores detrás de un balanceador de carga),
// esta estructura debe ser reemplazada por una base de datos externa en RAM como Redis.
type IdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]entry
}

// NewIdempotencyStore inicializa el motor de caché y arranca su rutina de mantenimiento.
func NewIdempotencyStore() *IdempotencyStore {
	s := &IdempotencyStore{
		entries: make(map[string]entry),
	}
	// Limpieza periódica de keys expiradas ejecutada de forma asíncrona (Fire-and-Forget)
	go s.cleanup()
	return s
}

// cleanup es el recolector de basura (Garbage Collector) manual del Store.
// Prevención de Fugas de Memoria (OOM - Out of Memory):
// Como un mapa en Go no elimina sus llaves automáticamente, un servidor que reciba
// un millón de peticiones podría quedarse sin RAM. Esta Goroutine despierta cada 10 minutos,
// bloquea el mapa, destruye las respuestas expiradas y libera la memoria.
func (s *IdempotencyStore) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock() // Bloqueo total (Escritura) para modificar el mapa de forma segura
		now := time.Now()
		for k, e := range s.entries {
			if now.After(e.expires) {
				delete(s.entries, k)
			}
		}
		s.mu.Unlock() // Liberación del bloqueo
	}
}

// get intenta recuperar una respuesta cacheada previamente.
func (s *IdempotencyStore) get(key string) (entry, bool) {
	s.mu.RLock() // Bloqueo de solo lectura (permite a múltiples Goroutines leer al mismo tiempo)
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	// Doble validación: Si existe la llave pero su tiempo expiró, se trata como un "Cache Miss"
	if !ok || time.Now().After(e.expires) {
		return entry{}, false
	}
	return e, true
}

// set almacena la respuesta HTTP definitiva en el diccionario de memoria.
func (s *IdempotencyStore) set(key string, status int, body []byte, ttl time.Duration) {
	s.mu.Lock() // Bloqueo exclusivo para escribir
	defer s.mu.Unlock()
	s.entries[key] = entry{
		status:  status,
		body:    body,
		expires: time.Now().Add(ttl),
	}
}

// ─────────────────────────────────────────────────────────────────────────────

// UserScopedIdempotency devuelve el interceptor que aplica la lógica de idempotencia.
//
// Diseño de Seguridad Obligatorio (User Scoping):
// Si NO vinculáramos la clave al ID del usuario, ocurriría una fuga de datos crítica.
// Ejemplo del fallo: El Usuario A envía la clave "123" y obtiene su propio reporte financiero.
// Luego el Usuario B malicioso envía la misma clave "123". Si el caché fuera global,
// el servidor le entregaría el reporte del Usuario A al Usuario B.
// Este middleware exige que la respuesta cacheada solo pueda ser consumida por quien la generó.
//
// TTL recomendado: 30 minutos (tiempo razonable para reintentos del frontend)
func UserScopedIdempotency(store *IdempotencyStore, ttl time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawKey := c.Get("X-Idempotency-Key")

		// 1. Opt-In: La idempotencia es opcional por diseño.
		// Si el cliente (Frontend/App) no envía la cabecera, la petición sigue su curso
		// de manera normal sin ser interceptada ni cacheada.
		if rawKey == "" {
			return c.Next()
		}

		// 2. Extracción de Identidad Estricta:
		// Obtener el admin autenticado del contexto (seteado previamente por ProtectedAdmin404)
		admin, ok := c.Locals("admin").(*domain.UserAdmin)
		if !ok || admin == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No autenticado.",
			})
		}

		// 3. Aislamiento Criptográfico (Scoping):
		// Construir una key compuesta: hash(userID + rawKey)
		// Así, la misma rawKey ("123") usada por distintos usuarios resulta en
		// hashes de memoria completamente distintos, evitando colisiones cruzadas.
		scopedKey := scopeKey(admin.ID.String(), rawKey)

		// ── 4. Cache Hit (Respuesta Interceptada) ─────────────────────────────
		if cached, found := store.get(scopedKey); found {
			// Inyectamos un flag HTTP para que el cliente sepa que esta petición
			// NO ejecutó código nuevo, sino que es una "Repetición" (Replay) del caché.
			c.Set("X-Idempotent-Replayed", "true")
			return c.Status(cached.status).Send(cached.body)
		}

		// ── 5. Cache Miss (Ejecutar Petición) ────────────────────────────────
		// Si es la primera vez que vemos esta clave, pausamos el middleware y dejamos
		// que el flujo HTTP continúe hacia el Controlador y la Base de Datos.
		if err := c.Next(); err != nil {
			return err
		}

		// 6. Almacenamiento (Post-Procesamiento):
		// Una vez que el Controlador terminó, evaluamos el resultado.
		// Solo cacheamos respuestas que hayan sido EXITOSAS (HTTP 2xx).
		// Si el servidor dio un Error 500 o 400, no lo cacheamos, permitiendo
		// al cliente reintentar legítimamente hasta lograr el éxito.
		if c.Response().StatusCode() >= 200 && c.Response().StatusCode() < 300 {
			store.set(scopedKey, c.Response().StatusCode(), c.Response().Body(), ttl)
		}

		return nil
	}
}

// scopeKey genera un hash determinista de userID + rawKey.
//
// Normalización de Longitud y Seguridad:
// Si simplemente concatenáramos "UserID:Key", el diccionario de memoria tendría llaves
// de longitudes variables e impredecibles. Usar SHA-256 garantiza que todas las llaves
// del mapa (map[string]entry) tengan exactamente el mismo peso computacional y formato (Hexadecimal),
// haciéndolas resistentes a caracteres inválidos o inyecciones que el cliente pudiera
// enviar en la cabecera 'X-Idempotency-Key'.
func scopeKey(userID, rawKey string) string {
	h := sha256.Sum256([]byte(userID + ":" + rawKey))
	return fmt.Sprintf("%x", h)
}

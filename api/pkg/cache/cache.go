// Package cache provee un wrapper de caché con respaldo in-memory.
//
// Almacenamiento: si VALKEY_ADDR está configurado, los datos persisten en
// Valkey (compatible con Redis, despliegue multi-instancia). Si no, se usa un
// mapa en memoria (adecuado para desarrollo).
//
// Invalida por generación: cada namespace lleva un contador. Al invalidar, se
// incrementa el contador y las claves compuestas con la generación vieja quedan
// huérfanas (expiran por TTL). Esto permite "borrar todo un namespace" sin
// soporte de wildcards.
package cache

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/valkey"
	"github.com/rs/zerolog/log"
)

const generationPrefix = "cache:gen:"

type memEntry struct {
	value  []byte
	expiry time.Time
}

// Cache es un almacén de clave/valor con TTL y fallback in-memory.
type Cache struct {
	store fiber.Storage // nil = modo in-memory

	mu         sync.RWMutex
	mem        map[string]memEntry
	generation map[string]uint64
}

// New crea un Cache. Si addr está vacío usa un mapa en memoria.
func New(addr string) *Cache {
	c := &Cache{
		mem:        make(map[string]memEntry),
		generation: make(map[string]uint64),
	}

	if addr == "" {
		log.Info().Str("component", "cache").Msg("Caché: in-memory (sin VALKEY_ADDR configurado)")
		return c
	}

	// gofiber/storage/valkey panics si no puede conectar: degradar a in-memory.
	defer func() {
		if r := recover(); r != nil {
			log.Warn().Any("panic", r).Str("component", "cache").Msg("No se pudo conectar a Valkey. Usando caché in-memory.")
			c.store = nil
		}
	}()

	c.store = valkey.New(valkey.Config{
		InitAddress: []string{addr},
	})
	log.Info().Str("component", "cache").Str("addr", addr).Msg("Caché: Valkey — persistente, multi-instancia")
	return c
}

// generationFor devuelve la generación actual del namespace.
func (c *Cache) generationFor(ns string) uint64 {
	if c.store == nil {
		c.mu.RLock()
		g := c.generation[ns]
		c.mu.RUnlock()
		return g
	}

	key := generationPrefix + ns
	raw, err := c.store.Get(key)
	if err != nil || len(raw) == 0 {
		return 0
	}
	var g uint64
	for _, b := range raw {
		g = g<<8 | uint64(b)
	}
	return g
}

// Set guarda val bajo "ns:sub" con el TTL indicado.
func (c *Cache) Set(ns string, sub string, val []byte, ttl time.Duration) {
	if c.store == nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.mem[cacheKey(ns, sub)] = memEntry{value: val, expiry: time.Now().Add(ttl)}
		return
	}
	_ = c.store.Set(cacheKey(ns, sub), val, ttl)
}

// Get recupera un valor previamente guardado.
func (c *Cache) Get(ns string, sub string) ([]byte, bool) {
	key := cacheKey(ns, sub)

	if c.store == nil {
		c.mu.RLock()
		e, ok := c.mem[key]
		c.mu.RUnlock()
		if !ok {
			return nil, false
		}
		if time.Now().After(e.expiry) {
			c.mu.Lock()
			delete(c.mem, key)
			c.mu.Unlock()
			return nil, false
		}
		return e.value, true
	}

	raw, err := c.store.Get(key)
	if err != nil || len(raw) == 0 {
		return nil, false
	}
	return raw, true
}

// Invalidate invalida todas las entradas de un namespace subiendo su generación.
// Las entradas viejas quedan huérfanas y expiran por TTL.
func (c *Cache) Invalidate(ns string) {
	if c.store == nil {
		c.mu.Lock()
		c.generation[ns]++
		c.mu.Unlock()
		return
	}
	key := generationPrefix + ns
	raw, err := c.store.Get(key)
	if err != nil || len(raw) == 0 {
		_ = c.store.Set(key, []byte{0, 0, 0, 0, 0, 0, 0, 1}, 0)
		return
	}
	var g uint64
	for _, b := range raw {
		g = g<<8 | uint64(b)
	}
	g++
	enc := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		enc[i] = byte(g)
		g >>= 8
	}
	_ = c.store.Set(key, enc, 0)
}

// GenKey compone un "sub" que incluye la generación vigente del namespace.
// Se usa junto con Set/Get para que la invalidación haga efecto: al subir la
// generación, el sub cambia y las entradas viejas quedan huérfanas (TTL).
func (c *Cache) GenKey(ns string, sub string) string {
	return strconv.FormatUint(c.generationFor(ns), 10) + ":" + sub
}

func cacheKey(ns string, sub string) string {
	return "cache:" + ns + ":" + sub
}

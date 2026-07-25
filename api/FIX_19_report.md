# FIX-19: Eliminar Idempotencia Global

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-19 |
| **Hallazgo** | HIGH-12 |
| **Severidad** | ALTO |
| **Archivo** | `cmd/api/main.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

Existían dos capas de idempotencia en conflicto:

1. **Capa Global** (`main.go:169-172`):
   ```go
   app.Use(idempotency.New(idempotency.Config{
       Lifetime:  30 * time.Minute,
       KeyHeader: "X-Idempotency-Key",
   }))
   ```

2. **Capa Custom per-route** (`psi_router.go:43`, `post_router.go:48`):
   ```go
   middleware.UserScopedIdempotency(idempotencyStore, 30*time.Minute)
   ```

La capa global usaba el header `X-Idempotency-Key` y la capa custom usaba SHA-256 del body + userId. Esto causaba comportamiento indefinido: si un cliente enviaba un `X-Idempotency-Key`, la capa global procesaba primero y la custom nunca se ejecutaba. Si no lo enviaba, la custom procesaba correctamente.

---

## Solución

Eliminación de la capa global. La implementación custom (`UserScopedIdempotency`) es superior porque:

1. Es scoped por usuario (SHA-256 del body + userId)
2. Previene duplicados por usuario, no por key header
3. Usa store persistente (`sync.RWMutex + map`)

---

## Cambios Realizados

### `cmd/api/main.go`

| Antes | Después |
|-------|---------|
| Import `idempotency` presente | Import eliminado |
| `app.Use(idempotency.New(...))` en líneas 169-172 | Eliminado |

**Líneas eliminadas:** 4 (3 de configuración + 1 import)

---

## Verificación

- `go build ./...` — ✅ Compila sin errores
- `idempotency` import eliminado — ✅ Sin imports huérfanos
- `UserScopedIdempotency` se mantiene en `psi_router.go` y `post_router.go` — ✅

---

## Impacto

- **Riesgo:** BAJO — solo se eliminó middleware redundante
- **Dependencias:** Ninguna
- **Tests:** N/A (middleware global no tenía tests específicos)
- **Rollback:** Revertir eliminación de las 4 líneas

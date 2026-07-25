# FIX-11 Report — Casteo inseguro de c.Locals() sin nil-check

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-11 |
| **Hallazgo original** | HIGH-04 |
| **Archivos modificados** | `middleware/helpers.go` (nuevo), `handler/psi_handler.go`, `handler/psi_user_admin.go`, `handler/posts_handler.go`, `handler/specialty_handler.go`, `handler/admin_handler.go` |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

19 instancias en 5 handlers hacían type assertion directa sin verificar nil:

```go
// PANIC si c.Locals("psi_user") es nil:
psi := c.Locals("psi_user").(*domain.PsiUserModel)

// PANIC si c.Locals("admin") es nil:
admin := c.Locals("admin").(*domain.UserAdmin)
```

Si por cualquier bug el middleware no seteaba el valor, la API hacía **panic** y se caía.

---

## Corrección

### 1. `internal/middleware/helpers.go` — NUEVO ARCHIVO

```go
func GetAuthenticatedAdmin(c *fiber.Ctx) (*domain.UserAdmin, error) {
    admin, ok := c.Locals("admin").(*domain.UserAdmin)
    if !ok || admin == nil {
        return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Sesión administrativa inválida o expirada",
        })
    }
    return admin, nil
}

func GetAuthenticatedPsi(c *fiber.Ctx) (*domain.PsiUserModel, error) {
    psi, ok := c.Locals("psi_user").(*domain.PsiUserModel)
    if !ok || psi == nil {
        return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Sesión de psicólogo inválida o expirada",
        })
    }
    return psi, nil
}
```

### 2. Reemplazo en handlers (19 puntos)

```go
// ANTES (panic si nil):
admin := c.Locals("admin").(*domain.UserAdmin)

// DESPUÉS (retorna 401 si nil):
admin, err := middleware.GetAuthenticatedAdmin(c)
if err != nil {
    return err
}
```

### Detalle por handler

| Handler | Instancias | Tipo | Helper |
|---------|-----------|------|--------|
| `psi_handler.go` | 6 | `psi_user` | `GetAuthenticatedPsi` |
| `psi_user_admin.go` | 5 | `admin` | `GetAuthenticatedAdmin` |
| `posts_handler.go` | 2 | `admin` | `GetAuthenticatedAdmin` |
| `specialty_handler.go` | 3 | `admin` | `GetAuthenticatedAdmin` |
| `admin_handler.go` | 3 | `admin` | `GetAuthenticatedAdmin` |

---

## Testing

Se verificó compilación (`go vet`) y tests después de cada handler:

| Handler | `go vet` | `go test middleware` |
|---------|----------|---------------------|
| psi_handler | Pass | 15/15 |
| psi_user_admin | Pass | 15/15 |
| posts_handler | Pass | 15/15 |
| specialty_handler | Pass | 15/15 |
| admin_handler | Pass | 15/15 |

**Test final completo:**
- `go vet ./...` — Pass
- `go test ./internal/middleware/...` — 15/15 Pass
- `go test ./internal/utils/...` — 4/5 Pass (1 fallo pre-existente FIX-33)

---

## Nota sobre dependencia circular

No existe circular dependency entre `handler/` y `middleware/`. `handler/` importa `middleware/` para usar los helpers, pero `middleware/` NO importa `handler/`. `helpers.go` importa solo `domain/` y `fiber`.

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/middleware/helpers.go` | Helpers creados |
| `internal/handler/psi_handler.go` | 6 reemplazos |
| `internal/handler/psi_user_admin.go` | 5 reemplazos |
| `internal/handler/posts_handler.go` | 2 reemplazos |
| `internal/handler/specialty_handler.go` | 3 reemplazos |
| `internal/handler/admin_handler.go` | 3 reemplazos |
| `SECURITY_FIX_PLAN.md` (FIX-11) | Plan de seguridad referenciado |

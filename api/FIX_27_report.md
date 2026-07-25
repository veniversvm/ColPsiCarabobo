# FIX-27: Reemplazar status codes hardcoded por fiber.Status*

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-27 |
| **Hallazgo** | MED-08 |
| **Severidad** | MEDIO |
| **Archivos** | `psi_handler.go`, `psi_user_admin.go`, `analytics_handler.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

20 ocurrencias de status codes HTTP como enteros crudos (`400`, `403`, `500`, `201`) en 3 archivos. Inconsistencia con `admin_handler.go`, `posts_handler.go` y `specialty_handler.go` que ya usaban `fiber.Status*`.

---

## Cambios Realizados

| Reemplazo | Cantidad |
|-----------|----------|
| `400` → `fiber.StatusBadRequest` | 9 |
| `403` → `fiber.StatusForbidden` | 4 |
| `500` → `fiber.StatusInternalServerError` | 5 |
| `201` → `fiber.StatusCreated` | 2 |
| **Total** | **20** |

---

## Verificación

- `rg 'c\.Status\([0-9]' internal/handler/` → 0 resultados ✅
- `go build ./...` → sin errores ✅

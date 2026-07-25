# FIX-25: Documentar descarte intencional de errores FormFile

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-25 |
| **Hallazgo** | MED-06 |
| **Severidad** | MEDIO |
| **Archivos** | `psi_handler.go`, `posts_handler.go`, `psi_user_admin.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

13 llamadas a `c.FormFile()` descartaban el error con `_` sin documentación. Los archivos son genuinamente opcionales (profile pics, imágenes de post, etc.), por lo que el `_` es intencional — cuando no se envía archivo, `c.FormFile` retorna `fiber.ErrNotFound` y el código pasa `nil` al servicio.

---

## Solución

Agregados comments documentando el descarte intencional en cada grupo de FormFile calls:

| Archivo | Líneas | Cantidad |
|---------|--------|----------|
| `psi_handler.go` | 101-104 | 4 FormFile calls |
| `psi_handler.go` | 357-359 | 3 FormFile calls |
| `posts_handler.go` | 91 | 1 FormFile call |
| `posts_handler.go` | 172 | 1 FormFile call |
| `psi_user_admin.go` | 129-132 | 4 FormFile calls |

---

## Verificación

- `go build ./...` → sin errores ✅
- Sin cambios de comportamiento, solo documentación ✅

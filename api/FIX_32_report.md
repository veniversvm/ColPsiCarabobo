# FIX-32: Eliminar BoolFromForm() duplicada

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-32 |
| **Hallazgo** | MED-13 |
| **Severidad** | MEDIO |
| **Archivo** | `internal/request_structs/psi_user.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

Existían dos copias de `BoolFromForm()` con comportamiento diferente:

| Versión | Archivo | Normalización | "yes" | "True" | "" | Valor no reconocido |
|---------|---------|---------------|-------|--------|-----|---------------------|
| **Canonical** | `utils/geo_venezuela.go:144` | `ToLower + TrimSpace` | `*true` | `*true` | `nil` | `nil` |
| **Duplicada** | `request_structs/psi_user.go:117` | Ninguna | `*false` | `*false` | `nil` | `*false` |

La versión duplicada convirtiya valores no reconocidos a `false` en vez de `nil`, causando que campos de visibilidad se silenciaran accidentalmente. El flujo de admin (`UpdatePsiAdminRequest`) ya usaba correctamente `utils.BoolFromForm`. Solo el flujo de auto-edición (`PsiUserUpdateRequestSelf`) usaba la versión buggeada.

---

## Cambios Realizados

1. Eliminada función `BoolFromForm()` (líneas 115-123)
2. Agregado import `"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"`
3. Reemplazadas 16 llamadas `BoolFromForm(` → `utils.BoolFromForm(`

---

## Verificación

- `rg 'func BoolFromForm' request_structs/psi_user.go` → 0 resultados ✅
- `go build ./...` → sin errores ✅

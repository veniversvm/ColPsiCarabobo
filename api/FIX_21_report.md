# FIX-21: Consolidar mapDBErrorToHuman → MapDBError

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-21 |
| **Hallazgo** | MED-02 |
| **Severidad** | MEDIO |
| **Archivos** | `error_mapper.go`, `psi_service.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

Dos funciones de mapeo de errores de DB con solapamiento parcial:
- `mapDBErrorToHuman` (privada, psi_service.go) — manejaba length/UUID errors pero **filtraba errores crudos al cliente** (CWE-209)
- `MapDBError` (pública, error_mapper.go) — segura pero **no cubría** `_key` suffixes, length errors, ni UUID errors

---

## Solución

Consolidación en `MapDBError` unificando la cobertura de ambas:

| Patrón | Antes (privada) | Antes (pública) | Después (unificada) |
|--------|-----------------|------------------|---------------------|
| `idx_psi_users_ci` | ✅ | ✅ | ✅ |
| `uni_psi_users_ci` | ❌ | ✅ | ✅ |
| `psi_users_ci_key` | ✅ | ❌ | ✅ |
| `idx_psi_users_email` | ✅ | ❌ | ✅ |
| `uni_psi_users_email` | ❌ | ✅ | ✅ |
| `psi_users_email_key` | ✅ | ❌ | ✅ |
| `value too long` | ✅ | ❌ | ✅ |
| `invalid input syntax uuid` | ✅ | ❌ | ✅ |
| Fallback seguro (sin CWE-209) | ❌ (filtraba raw) | ✅ | ✅ |

---

## Cambios Realizados

1. `error_mapper.go` — Enriquecido `MapDBError` con patrones `_key`, length errors, UUID errors
2. `psi_service.go` — Eliminada `mapDBErrorToHuman` (33 líneas)
3. `psi_service.go` — Call site actualizado: `mapDBErrorToHuman(err)` → `MapDBError(err).Error()`

---

## Verificación

- `rg 'mapDBErrorToHuman' internal/service/` → 0 resultados ✅
- `go build ./...` → sin errores ✅
- Call sites existentes de `MapDBError` (`psi_user_admin_service.go:128`, `psi_service_xlsx.go:263`) sin cambios ✅

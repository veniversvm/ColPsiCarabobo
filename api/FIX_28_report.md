# FIX-28: Agregar Swagger + analytics a LoginLibrary

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-28 |
| **Hallazgo** | MED-09 |
| **Severidad** | MEDIO |
| **Archivos** | `psi_handler.go`, `psi_service.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

`LoginLibrary` no tenía:
1. Anotaciones Swagger (invisible en la documentación de la API)
2. Llamada a analytics (los logins de biblioteca no se registraban en las métricas)

---

## Solución

### Handler (`psi_handler.go`)
- Agregadas anotaciones Swagger completas (Summary, Description, Tags, Accept, Produce, Param, Success, Failure, Router)
- Agregada llamada `h.analytics.RecordLogin(...)` con source `"psi_library"` después de login exitoso

### Service (`psi_service.go`)
- Cambiada firma: `(string, error)` → `(string, *domain.PsiUserModel, error)`
- Retornos actualizados para incluir el modelo de usuario en éxito y `nil` en error

---

## Verificación

- `go build ./...` → sin errores ✅
- Firma consistente con `Login()` que también retorna `(string, *PsiUserModel, error)` ✅
- Analytics registra source `"psi_library"` para diferenciar de login normal (`"psi"`) ✅

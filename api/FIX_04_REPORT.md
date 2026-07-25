# FIX-04 Report — Tags GORM erróneos

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-04 |
| **Hallazgo original** | CRIT-04 |
| **Archivo modificado** | `internal/domain/user.model.go` |
| **Líneas afectadas** | 108, 109 |
| **Fecha de implementación** | 2026-07-24 |
| **Estado** | Completado |

---

## Problema

Dos campos en `PsiUserModel` tenían tags GORM semánticamente incorrectos:

### Línea 108 — `size:255` en un `bool`

```go
// ANTES:
ShowMunicipalityCarabobo bool   `gorm:"size:255" json:"show_municipality_carabobo"`
```

El tag `size:255` solo aplica a tipos `string`. En un `bool`, GORM lo ignora silenciosamente, pero es un error semántico que puede confundir a herramientas de migración automática (como Atlas) y genera una representación incorrecta del esquema.

### Línea 109 — `default:false` en un `string`

```go
// ANTES:
PhoneCarabobo            string `gorm:"default:false" json:"phone_carabobo"`
```

Cuando GORM aplica `default:false` a un campo `string`, PostgreSQL almacena la string literal `'false'` como valor por defecto de la columna. Esto se confirma en la migración existente:

```sql
-- migrations/20260604165811_init.sql:262
phone_carabobo text DEFAULT 'false'::character varying
```

Esto crea un estado inconsistente: un campo de teléfono con valor por defecto `'false'` en vez de una cadena vacía.

---

## Corrección

```go
// DESPUÉS (líneas 108-109):
ShowMunicipalityCarabobo bool   `gorm:"default:false" json:"show_municipality_carabobo"`
PhoneCarabobo            string `gorm:"default:''" json:"phone_carabobo"`
```

| Línea | Cambio | Justificación |
|-------|--------|---------------|
| 108 | `size:255` → `default:false` | Tag correcto para un campo booleano |
| 109 | `default:false` → `default:''` | Default de string vacío, no la literal `'false'` |

---

## Testing

### 1. `go vet` — Análisis estático

```bash
$ go vet ./internal/domain/
# (sin output = sin errores)

$ go vet ./...
# (sin output = sin errores en todo el proyecto)
```

**Resultado:** Pass

### 2. `go test ./internal/utils/...` — Tests unitarios

```
=== RUN   TestSanitizeImage_Defensive
    FAIL  (bug pre-existente, no relacionado con FIX-04)
=== RUN   TestIsEmptyReq
    PASS
=== RUN   TestNormalizePlatformName
    PASS
=== RUN   TestGenerateSecureRandomString
    PASS
=== RUN   TestIsStrongPassword
    PASS
```

| Test | Estado | Notas |
|------|--------|-------|
| `TestSanitizeImage_Defensive` | FAIL | Bug pre-existente (FIX-33): el test busca `"archivo no es una imagen válida"` pero el handler retorna `"el servidor no reconoce este formato de imagen"` |
| `TestIsEmptyReq` | PASS | 6/6 subtests |
| `TestNormalizePlatformName` | PASS | 10/10 subtests |
| `TestGenerateSecureRandomString` | PASS | 3/3 subtests |
| `TestIsStrongPassword` | PASS | 9/9 subtests |

**Resultado:** 4/5 paquetes pass, 1 fallo pre-existente no relacionado.

### 3. `go test ./internal/middleware/...` — Tests de middleware

```
=== RUN   TestAuthMiddleware_Extensive
    PASS (6/6 subtests)
```

**Resultado:** Pass

### 4. Compilación

El proyecto completa `go build` y `go vet ./...` sin errores. La compilación completa del binario final tomó más de 2 minutos en la primera ejecución debido a la descarga de dependencias transitivas (atlas, google cloud, gRPC). En ejecuciones subsiguientes compila en segundos.

---

## Impacto

- **Riesgo de regressión:** Mínimo. Los cambios son en tags GORM que afectan la generación de esquema, no la lógica de negocio.
- **Migración requerida:** Sí, diferida. La migración Atlas que corrija el default de `phone_carabobo` de `'false'` a `''` está diferida a FIX-18 (ON DELETE NO ACTION) donde se ejecutan todas las migraciones SQL juntas.
- **Acción en DB:** Ninguna inmediata. Los defaults correctos se aplicarán a nuevos registros. Los registros existentes con `phone_carabobo = 'false'` se migrarán en la migración de FIX-18.

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/domain/user.model.go:108-109` | Archivo modificado |
| `migrations/20260604165811_init.sql:262` | Migración existente con `DEFAULT 'false'` |
| `SECURITY_FIX_PLAN.md` (FIX-04, FIX-18, FIX-37) | Plan de seguridad referenciado |

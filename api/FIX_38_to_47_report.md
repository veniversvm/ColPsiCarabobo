# Reporte de Seguridad — Fase 4 (Bajos / Polish)

**Fecha:** 25 de Julio, 2026
**Branch:** docs
**Fixes aplicados:** FIX-38 a FIX-47 (10 fixes)

---

## Resumen Ejecutivo

La Fase 4 cerró los 10 hallazgos de severidad baja restantes. Se realizaron mejoras de polish, consistencia y deuda técnica que incluyen: corrección de typos en archivos y código, eliminación de código muerto, estandarización de tipos, unificación de modelos y migración a logs condicionales.

**Total de archivos modificados:** ~25 archivos Go + 1 migración SQL + 1 archivo de test nuevo

---

## FIX-38: Typos en nombres de archivos

| Hallazgo | Acción |
|----------|--------|
| `internal/utils/radom_string.go` | Renombrado a `random_string.go` |
| `internal/domain/post_respository.go` | Renombrado a `post_repository.go` |

**Riesgo:** Nulo. Los imports de Go son por paquete, no por nombre de archivo.

---

## FIX-39: Typos en código

| Hallazgo | Ubicación | Acción |
|----------|-----------|--------|
| `PsiUSerSolvency` | 32 ocurrencias en 8 archivos | Renombrado a `PsiUserSolvency` |
| `"emial invalido"` | `admin_service.go:379` | Corregido a `"email inválido"` |
| Log "BIO" en solvencies | `psi_user_admin.go:58` | Corregido a "SOLVENCIES" |
| Comentario "Empleado público" en Discapacity | `user.model.go:196` | Corregido a "Discapacidad" |

---

## FIX-40: Post sin TableName() explícito

Agregado `func (Post) TableName() string { return "posts" }` en `text.model.go`.

GORM ya infería `"posts"` correctamente, pero la explícita inconsistencia con el resto de modelos era un riesgo de mantenimiento.

---

## FIX-41: GraduationYear `string` → `int`

| Archivo | Cambio |
|---------|--------|
| `domain/user.model.go` | `GraduationYear string` → `int` |
| `request_structs/psi_user.go` | 3 structs: `string` → `int` / `*int` |
| `handler/psi_handler.go` | Agregado `strconv.Atoi` para parsear form value |
| `service/psi_service.go` | Sin cambios necesarios (tipos ya coinciden) |

**Migración:** `ALTER TABLE psi_user_post_grades ALTER COLUMN graduation_year TYPE bigint`

---

## FIX-42: UUID gen_random_uuid → uuidv7

| Hallazgo | Acción |
|----------|--------|
| `LoginEvent.ID` usaba `gen_random_uuid()` | Cambiado a `uuidv7()` |
| `ActiveSession.ID` usaba `gen_random_uuid()` | Cambiado a `uuidv7()` |
| Extensión `pg_uuidv7` no estaba declarada | Agregado `CREATE EXTENSION IF NOT EXISTS` |

**Migración:** `ALTER TABLE login_events ALTER COLUMN id SET DEFAULT uuidv7()` + `ALTER TABLE active_sessions ALTER COLUMN id SET DEFAULT uuidv7()`

---

## FIX-43: `context.TODO()` → `ctx`

Cambiado `ConnectS3()` → `ConnectS3(ctx context.Context)` en `pkg/s3/s3.go`.
Caller en `main.go` ahora pasa `context.Background()`.

---

## FIX-44: GetPresignedURL() comentado

Eliminadas 16 líneas de código muerto (`pkg/s3/upload.go:77-92`). La función no se usaba en ningún lado.

---

## FIX-45: Emojis en logs → prefijos textuales

32 reemplazos en 11 archivos. Mapeo:
- ✅ → `[OK]`
- ⚠️ → `[WARN]`
- ❌ → `[ERROR]`
- 🚀🌱🔄🕒 → `[INFO]`

---

## FIX-46: `[DEBUG AUTH]` en producción

3 `log.Println("[DEBUG AUTH]...")` en `middleware/auth.go` ahora son condicionales:
```go
if config.Envs != nil && config.Envs.Environment == "development" {
    log.Println("[DEBUG AUTH] ...")
}
```

Incluye nil-check para evitar panic en tests.

---

## FIX-47: Struct `Credentials` embebido

**Nuevo archivo:** `internal/domain/credentials.go`

Extrae 5 campos compartidos de `UserAdmin` y `PsiUserModel`:
- `Username`, `Email`, `Password`, `Key`, `IsActive`

**Beneficios:**
- Punto único de verdad para credenciales
- Tags unificados: `Email` → `size:255` (ambos), `IsActive` → `column:is_active;default:true`
- Previene drift futuro entre los dos structs

**Impacto:**
- ~10 archivos de producción ajustados (struct literals → `Credentials: domain.Credentials{...}`)
- ~9 archivos de test ajustados (mismo patrón)
- 4 tests nuevos creados (`credentials_test.go`)

**Migración:** `ALTER TABLE psi_users ALTER COLUMN email TYPE varchar(255)` (de varchar(50))

---

## Tests

| Paquete | Tests | Resultado |
|---------|-------|-----------|
| domain | 4/4 | ✅ PASS (nuevos) |
| middleware | 17/17 | ✅ PASS |
| utils | 5/5 | ✅ PASS |
| service | 20/22 | ⚠️ 2 pre-existente |
| **Total** | **46/48** | **0 regressions** |

Los 2 failures son pre-existentes desde antes de la Fase 1:
- `TestAdminService_All/CreateAdmin` — expectativa de error incorrecta
- `TestSpecialtyService_Update` — mock falta `GetByAdminID`

---

## Migración SQL

Archivo: `migrations/20260725010000_fix41_42_type_corrections.sql`

```sql
-- FIX-42
CREATE EXTENSION IF NOT EXISTS "pg_uuidv7";
ALTER TABLE login_events ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE active_sessions ALTER COLUMN id SET DEFAULT uuidv7();

-- FIX-41
ALTER TABLE psi_user_post_grades ALTER COLUMN graduation_year TYPE bigint USING graduation_year::bigint;

-- FIX-47
ALTER TABLE psi_users ALTER COLUMN email TYPE varchar(255);
```

---

*Fase 4 completa. Todos los hallazgos de la auditoría de seguridad (7 críticos + 12 altos + 18 medios + 15 bajos = 52 total) han sido cubiertos.*

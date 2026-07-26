# Reporte — Key Lifecycle Management

**Fecha:** 25 de Julio, 2026
**Branch:** docs
**Commit:** `b511d43`
**Fixes:** CRIT-03 parcial (key lifecycle completo)

---

## Tabla de Contenidos

- [Resumen Ejecutivo](#resumen-ejecutivo)
- [1. Auditoría — UUIDs preexistentes](#1-auditoría--uuids-preexistentes)
- [2. uuid.New() → uuid.NewV7()](#2-uuidnew--uuidnewv7)
- [3. Middleware — Empty Key Guard](#3-middleware--empty-key-guard)
- [4. Logout de Admin](#4-logout-de-admin)
- [5. Logout de PsiUser — Unificación](#5-logout-de-psiuser--unificación)
- [6. Cleanup Job — cmd/cleanup](#6-cleanup-job--cmdcleanup)
- [7. Tests nuevos](#7-tests-nuevos)
- [8. Archivos modificados](#8-archivos-modificados)
- [9. Migraciones SQL requeridas](#9-migraciones-sql-requeridas)
- [10. Estado de tests](#10-estado-de-tests)
- [11. Pendiente](#11-pendiente)

---

## Resumen Ejecutivo

Implementación completa del ciclo de vida de claves de sesión (key lifecycle). Todas las keys de sesión ahora son UUID v7 (timestamp embebido), el middleware rechaza keys vacías sin intentar crypto, ambos roles (Admin y PsiUser) tienen logout funcional, y un binario independiente limpia keys expiradas cada 30 minutos.

| Componente | Estado |
|------------|--------|
| UUID v7 en producción | ✅ 0 `uuid.New()` restantes |
| Empty key guard | ✅ Admin 404 + Psi 401 |
| Admin logout | ✅ `POST /admin/logout` |
| PsiUser logout | ✅ key → `""` (no rotación) |
| Cleanup job | ✅ `cmd/cleanup/` binario independiente |
| Tests nuevos | ✅ 4 tests en `pkg/job` |
| Regresiones | ✅ 0 |

---

## 1. Auditoría — UUIDs preexistentes

Se realizó un barrido completo de `uuid.New()` (v4) y `uuid.NewString()` en todo el código de producción:

| Archivo | Línea | Uso original | Riesgo |
|---------|-------|-------------|--------|
| `admin_service.go:81` | Login key rotation | **Alta** — key sin timestamp |
| `seed.go:34` | Seed admin ID | Baja — solo seed |
| `s3/upload.go:30` | S3 filename | Baja — nombre de archivo |
| `analytics.go:76` | Session cookie ID | Media — cookie sin timestamp |

**Resultado:** 4 archivos migrados, 0 `uuid.New()` restantes en producción.

---

## 2. uuid.New() → uuid.NewV7()

Todos los `uuid.New()` (v4 aleatorio) fueron reemplazados por `uuid.NewV7()` (v7 con timestamp).

| Archivo | Línea | Antes | Después |
|---------|-------|-------|---------|
| `internal/service/admin_service.go:81` | Login key rotation | `uuid.New().String()` | `uuid.Must(uuid.NewV7()).String()` |
| `pkg/database/seed.go:34` | Seed admin ID | `uuid.New()` | `uuid.Must(uuid.NewV7())` |
| `pkg/s3/upload.go:30` | S3 filename | `uuid.New().String()` | `uuid.Must(uuid.NewV7()).String()` |
| `internal/middleware/analytics.go:76` | Session cookie | `uuid.NewString()` | `uuid.Must(uuid.NewV7()).String()` |

**Por qué importa:**
- UUID v4: 128 bits aleatorios, sin información temporal
- UUID v7: primeros 48 bits = timestamp en milisegundos
- Permite al cleanup job determinar la edad de una key sin un campo `created_at` separado

---

## 3. Middleware — Empty Key Guard

**Archivo:** `internal/middleware/auth.go`

### ProtectedAdmin404() (línea 130-132)

```go
c.Locals("admin", admin)
if admin.Key == "" {
    return "", errors.New("session expired")
}
return admin.Key, nil
```

### ProtectedPsiUser() (línea 260-262)

```go
c.Locals("psi_user", psi)
if psi.Key == "" {
    return "", errors.New("session expired")
}
return psi.Key, nil
```

**Comportamiento:**

| Estado de key | Antes | Después |
|---------------|-------|---------|
| Key válida | JWT parse + verificación | JWT parse + verificación |
| Key vacía (logout/cleanup) | `jwt.Parse` con `[]byte("")` → error crypto | Retorno inmediato → 401/404 |
| Key inexistente en DB | Error de GORM → 401/404 | Error de GORM → 401/404 |

**Ventajas:**
- Sin operaciones crypto innecesarias cuando la sesión ya fue invalidada
- Error explícito de "session expired" en logs
- No se expone el internals de JWT a keys vacías

---

## 4. Logout de Admin

### Nuevo endpoint

| Método | Ruta | Middleware | Descripción |
|--------|------|-----------|-------------|
| `POST` | `/api/v1/admin/logout` | `ProtectedAdmin404()` | Invalida la sesión del administrador |

### Request

```http
POST /api/v1/admin/logout
Authorization: Bearer <jwt_token>
```

### Response

```json
// 200 OK
{
  "message": "Sesión cerrada correctamente"
}

// 401 Unauthorized
{
  "error": "No autenticado"
}

// 404 Not Found (token inválido/expirado)
{
  "message": "Cannot POST /admin/logout"
}
```

### Flujo

```
Client → POST /admin/logout
       → ProtectedAdmin404 middleware
         → getKeyFunc: SELECT key FROM user_admins WHERE id = ?
         → key != "" ? → valida JWT con key como HMAC secret
       → AdminHandler.Logout
         → middleware.GetAuthenticatedAdmin(c) → extrae admin de Locals
         → AdminService.Logout(ctx, admin)
           → admin.Key = ""
           → admin.UpdateBy = admin.Username
           → repo.UpdateKey(ctx, admin)
             → UPDATE user_admins SET key = '', updated_at = ? WHERE id = ?
       → 200 OK
```

### Impacto en sesiones activas

- **Todas las sesiones** del admin quedan invalidadas inmediatamente
- Cualquier request futuro con el JWT anterior fallará en el middleware (key vacía → error)
- No hay forma de "revertir" el logout sin un login nuevo

---

## 5. Logout de PsiUser — Unificación

**Antes (rotación):**

```go
// psi_service.go — ANTES
func (s *PsiService) Logout(ctx context.Context, psi *domain.PsiUserModel) error {
    psi.Key = uuid.Must(uuid.NewV7()).String()  // Rotaba a nueva key
    psi.UpdateBy = psi.Username
    psi.UpdateById = &psi.ID
    return s.repo.UpdateKey(ctx, psi)
}
```

**Después (eliminación):**

```go
// psi_service.go — DESPUÉS
func (s *PsiService) Logout(ctx context.Context, psi *domain.PsiUserModel) error {
    psi.Key = ""  // Elimina la key
    psi.UpdateBy = psi.Username
    psi.UpdateById = &psi.ID
    return s.repo.UpdateKey(ctx, psi)
}
```

**Por qué el cambio:**

| Aspecto | Rotación (antes) | Eliminación (después) |
|---------|-------------------|----------------------|
| Invalidación JWT | ✅ JWT viejo falla | ✅ JWT viejo falla |
| Cleanup job | Key "muerta" queda 24h | Key ya está vacía → ignorada |
| Consistencia | Diferente a Admin | Mismo patrón que Admin |
| Seguridad | Key anterior visible en DB hasta cleanup | Key eliminada inmediatamente |

---

## 6. Cleanup Job — cmd/cleanup

### Arquitectura

```
cmd/cleanup/main.go
├── config.InitConfig()          ← Mismas env vars que cmd/api
├── database.ConnectDB()         ← Misma conexión PostgreSQL
├── ticker cada 30 min
│   └── job.CleanExpiredKeys()
│       ├── SELECT key FROM user_admins WHERE key != '' AND deleted_at IS NULL
│       ├── Para cada key: uuid.Parse → extraer timestamp v7
│       │   └── Si timestamp > 24h → UPDATE key = ''
│       └── Repetir para psi_users
└── graceful shutdown (SIGINT/SIGTERM)
```

### pkg/job/key_cleanup.go

```go
// Función principal
func CleanExpiredKeys(ctx context.Context, db *gorm.DB, maxAge time.Duration) (KeyCleanupResult, error)

// Helper: parsea UUID v7, retorna true si timestamp < cutoff
func isKeyExpired(key string, cutoff time.Time) bool {
    parsed, err := uuid.Parse(key)
    if err != nil {
        return true  // Keys inválidas se consideran expiradas
    }
    sec, nsec := parsed.Time().UnixTime()
    ts := time.Unix(sec, nsec)
    return ts.Before(cutoff)
}
```

### Ejecución

```bash
# Desarrollo
APP_ENV=development go run cmd/cleanup/main.go

# Producción (binario compilado)
go build -o cleanup cmd/cleanup/main.go
./cleanup

# Docker (agregar a docker-compose.yml)
cleanup:
  build: .
  command: ["./cleanup"]
  env_file: .env
  depends_on:
    - postgres
```

### Comportamiento

| Instante | Acción |
|----------|--------|
| Arranque | Ejecuta limpieza inmediatamente |
| Cada 30 min | Ejecuta limpieza programada |
| SIGINT/SIGTERM | Cierra conexión DB y sale gracefully |

### Log de ejemplo

```
[CLEANUP] Iniciando servicio de limpieza de keys...
[CLEANUP] Tick cada 30m0s | Keys > 24h0m0s serán borradas
[CLEANUP] 3 keys expiradas borradas (admins: 1, psi: 2)
```

---

## 7. Tests nuevos

**Archivo:** `pkg/job/key_cleanup_test.go`

| Test | Descripción | Resultado |
|------|-------------|-----------|
| `TestIsKeyExpired/Key_v7_reciente_(no_expirada)` | UUID v7 reciente con cutoff anterior | ✅ PASS |
| `TestIsKeyExpired/Key_inválida_(expirada_por_defecto)` | String no-UUID → expirada | ✅ PASS |
| `TestIsKeyExpired/Key_vacía_(expirada_por_defecto)` | String vacío → expirada | ✅ PASS |
| `TestIsKeyExpired_OldKey` | UUID v7 con cutoff futuro → expirada | ✅ PASS |

**Mock actualizado:** `internal/service/admin_service_test.go`
- Agregado `UpdateKeyFunc` y método `UpdateKey()` al `mockAdminRepo`

---

## 8. Archivos modificados

### Nuevos (3)

| Archivo | Descripción |
|---------|-------------|
| `cmd/cleanup/main.go` | Binario de limpieza de keys |
| `pkg/job/key_cleanup.go` | Lógica de limpieza + isKeyExpired |
| `pkg/job/key_cleanup_test.go` | 4 tests unitarios |

### Modificados (11)

| Archivo | Cambio |
|---------|--------|
| `internal/domain/admin_repository.go` | Nuevo método `UpdateKey()` en interfaz |
| `internal/handler/admin_handler.go` | Nuevo handler `Logout()` |
| `internal/middleware/analytics.go:76` | `uuid.NewString()` → `uuid.Must(uuid.NewV7()).String()` |
| `internal/middleware/auth.go:130,260` | Empty key guard en ambos middlewares |
| `internal/repository/postgres/user_admin_repo.go` | Implementación de `UpdateKey()` |
| `internal/router/admin_router.go` | Nueva ruta `POST /admin/logout` |
| `internal/service/admin_service.go:81` | `uuid.New()` → `uuid.Must(uuid.NewV7())` |
| `internal/service/admin_service.go:113` | Nuevo método `Logout()` |
| `internal/service/admin_service_test.go` | Mock actualizado con `UpdateKey` |
| `internal/service/psi_service.go:1148` | Logout: key → `""` en vez de rotación |
| `pkg/database/seed.go:34` | `uuid.New()` → `uuid.Must(uuid.NewV7())` |
| `pkg/s3/upload.go:30` | `uuid.New().String()` → `uuid.Must(uuid.NewV7()).String()` |

---

## 9. Migraciones SQL requeridas

Ninguna. Todos los cambios son a nivel de código. Las columnas `key` ya existen como `varchar` y aceptan strings vacíos.

---

## 10. Estado de tests

| Paquete | Tests | Pass | Fail | Notas |
|---------|-------|------|------|-------|
| `internal/domain` | 2 | 2 | 0 | — |
| `internal/middleware` | 17 | 17 | 0 | — |
| `internal/utils` | 1 | 1 | 0 | — |
| `pkg/job` | 4 | 4 | 0 | **Nuevos** |
| `internal/service` | ~15 | 13 | 2 | Pre-existente |
| `internal/repository/postgres` | 4 | 0 | 4 | Sin DB local |

**Pre-existente (no causado por este cambio):**
- `TestAdminService_All/CreateAdmin:_Regla_'No_puedes_dar_lo_que_no_tienes'` — Test espera error de jerarquía, recibe "email inválido"
- `TestSpecialtyService_Update/Actualización_Parcial_(PATCH)` — Mock no implementa `GetByAdminID`

**0 regresiones introducidas.**

---

## 11. Pendiente

| Fix | Descripción | Estado |
|-----|-------------|--------|
| CRIT-03 | Migrar `admin_service.go:281` (CreateAdmin key) a `uuid.NewV7()` | Ya usa `uuid.Must(uuid.NewV7())` ✅ |
| CRIT-03 | Migrar `admin_service.go:406` (UpdateAdmin key) a `uuid.NewV7()` | Ya usa `uuid.Must(uuid.NewV7())` ✅ |
| — | docker-compose servicio cleanup | Pendiente |
| — | Tests de integración para logout endpoints | Pendiente |

---

## Commits relacionados

| Commit | Fix | Descripción |
|--------|-----|-------------|
| `8f6c49a` | FIX-04, 05, 06 | Fase 1 |
| `0c2cde4` | FIX-01, 02 | Fase 2a |
| `32a5e30` | FIX-11 | Fase 2b |
| `ea7f5bf` | FIX-09, 10, 13, 14, 15, 15b | Fase 3 early |
| `707ac67` | FIX-33 | Test fix |
| `f5d9042` | FIX-12 | Sentinel errors |
| `e86b7d1` | FIX-07 | SMTP + MailHog |
| `7a59584` | — | Fix reports |
| `c7f03e0` | FIX-18 | FK delete policy |
| `b94d62c` | FIX-19 | Remove global idempotency |
| `e716e1b` | FIX-22 | println→log.Printf |
| `5c52cc6` | FIX-24 | Remove debug fmt |
| `4f5b7b1` | FIX-26 | Remove PII from logs |
| `cb902df` | FIX-27 | Status codes→fiber constants |
| `d789381` | FIX-32 | Remove duplicate BoolFromForm |
| `1acef6b` | FIX-21 | Consolidate MapDBError |
| `8278df3` | FIX-25 | Document FormFile discard |
| `ef9c636` | FIX-28 | Swagger+analytics for LoginLibrary |
| `81749be` | FIX-30 | Document Save() contract |
| `f1b82b9` | FIX-29 | Centralize repo instantiation |
| `a307697` | FIX-23 | Decouple AnalyticsService from gorm.DB |
| `f4ceafe` | FIX-18, 19 | Pending files |
| `355c330` | FIX-38 to 47 | Fase 4 completa |
| **`b511d43`** | **CRIT-03** | **Key Lifecycle Management** |

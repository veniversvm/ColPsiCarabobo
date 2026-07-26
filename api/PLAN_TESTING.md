# Plan de Implementación de Testing — ColPsiCarabobo API

## Estado Actual

### Lo que ya existe
- **19 archivos de test** existentes (~3,400 líneas)
- **Cobertura por capa:**
  - Service layer: 10 archivos (mocks manuales, func override pattern)
  - Repository layer: 4 archivos (PostgreSQL real con transacciones)
  - Middleware: 1 archivo (auth_test.go con Fiber app.Test)
  - Utils: 1 archivo (utils_test.go con testify)
  - Domain: 1 archivo (credentials_test.go)
  - Job: 1 archivo (key_cleanup_test.go)
- **Librerías:** testify/require (repos), testing stdlib (services), SQLite in-memory (updates_safety)
- **Mocks:** Todos manuales (func override pattern), sin framework de mocks
- **DB de test:** `TEST_DB_DSN` env var → PostgreSQL real, creación de `colpsi_test`, rollback por transacción

### ✅ Fase 1 COMPLETADA — Foundation Tests (147 subtests, 4 paquetes)
- `internal/request_structs/directory_filter_sanitizer_test.go` — ~35 subtests
- `internal/service/error_mapper_test.go` — 19 tests + nil panic test
- `internal/utils/utils_test.go` — Expandido con NormalizeMunicipioCarabobo, NormalizeEstadoVenezuela, BoolFromForm, CleanAlphaNumeric
- `internal/domain/credentials_test.go` — Expandido con sentinel errors, PostGradeType.IsValid, table name tests, JSON tag validation

### ✅ Fase 2 COMPLETADA — Service Layer Tests (91 subtests nuevos en service pkg)
- `internal/service/analytics_service_test.go` — 15 test functions + concurrent safety test (mockAnalyticsRepo)
- `internal/service/psi_service_xlsx_test.go` — 31 subtests: getValorSeguro (9) + ImportFromXLSX (22)
- `internal/service/admin_service_test.go` — Expandido: 5 functions (18 subtests) edge cases (login inactive, wrong password, sudo bypass, pagination, delete/update RBAC)
- `internal/service/psi_service_auth_test.go` — Expandido: 4 functions (7 subtests) key rotation, UpdateKey error, email notification, audit trail
- **Total suite:** 91 PASS, 2 FAIL (pre-existing en admin_service y specialty_service)

### Lo que falta (gaps críticos)
1. **0 tests de handlers/controllers** — No hay ningún test HTTP end-to-end
2. **0 tests de integración completos** — No hay tests que ejerciten la ruta completa request→handler→service→repo→DB
3. **Repo sin tests:** analytics, solvency, social networks (standalone), observations, deontologia
4. **Service sin tests:** error_mapper, rate_limiter, idempotency middleware, analytics middleware, helpers middleware, directory filter sanitizer
5. **Sin Makefile** de tests — No hay forma estandarizada de ejecutar tests
6. **Sin CI/CD** — No hay pipeline automatizado
7. **Sin test fixtures** — Datos de test duplicados inline en cada archivo
8. **Sin test helpers compartidos** — Cada archivo de test repo duplica el setup de DB

---

## Fase 0: Infraestructura Base de Testing

**Objetivo:** Crear la base sobre la cual construir todos los tests.

### 0.1 — Test Helpers Compartidos

**Archivo nuevo:** `internal/testutil/testutil.go`

```go
// Funciones a implementar:
- SetupTestDB(t *testing.T) *gorm.DB          // PostgreSQL real con cleanup
- SetupSQLiteDB(t *testing.T) *gorm.DB         // SQLite in-memory para tests rápidos
- SeedAdmin(t *testing.T, db *gorm.DB) *domain.UserAdmin
- SeedPsiUser(t *testing.T, db *gorm.DB) *domain.PsiUserModel
- SeedSpecialty(t *testing.T, db *gorm.DB) *domain.PsiSpecialtyModel
- SeedPost(t *testing.T, db *gorm.DB) *domain.Post
- CleanupTables(t *testing.T, db *gorm.DB, tables ...string)
- AssertHTTPStatus(t *testing.T, got, want int)
- AssertJSONError(t *testing.T, body []byte, wantMsg string)
- CreateTestFiberApp() *fiber.App               // App Fiber configurada para tests
- GenerateTestJWT(t *testing.T, userID uuid.UUID, role string) string
```

**Archivo nuevo:** `internal/testutil/mocks.go`

```go
// Mocks reutilizables (ya no inline en cada test):
- MockAdminRepository     // Implementa domain.UserAdminRepository
- MockPsiRepository       // Implementa domain.PsiUserRepository
- MockPostRepository      // Implementa domain.PostRepository
- MockSpecialtyRepository // Implementa domain.SpecialtyRepository
- MockAnalyticsRepository // Implementa domain.AnalyticsRepository
- MockS3Client            // Interfaz para operaciones S3
- MockMailService         // Interfaz para envío de emails
```

Cada mock tendrá:
- Campos `func` para cada método de la interfaz
- Campo `Called int` para contar invocaciones (spy pattern existente)
- Campo `LastArgs` para inspeccionar argumentos

### 0.2 — Makefile de Tests

**Archivo nuevo:** `Makefile`

```makefile
test-unit:          # Tests sin DB (services, utils, middleware puros)
test-repo:          # Tests de repository (requiere PostgreSQL)
test-handler:       # Tests de handlers (requiere PostgreSQL)
test-integration:   # Tests end-to-end (requiere PostgreSQL + S3)
test-all:           # Todos los tests
test-race:          # Con detector de carreras
test-coverage:      # Con reporte de cobertura HTML
test-bench:         # Benchmarks
```

### 0.3 — Environment de Test

**Archivo nuevo:** `.env.test`

```
TEST_DB_DSN=postgres://postgres:postgres@localhost:5432/colpsi_test?sslmode=disable
APP_ENV=test
JWT_LIBRARY_SECRET=test-secret-key-for-testing-only
ABS_ADMIN_TOKEN=test-token
S3_ENDPOINT=http://localhost:9000
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin
AWS_S3_BUCKET=colpsi-test-bucket
```

---

## Fase 1: Tests Unitarios — Utils y Domain (Sin dependencias externas)

**Objetivo:** Cubrir código puro sin mocks ni DB.

### 1.1 — `internal/request_structs/directory_filter_sanitizer_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestSanitizeDirectoryFilter/SQL_injection_en_q` | `' OR 1=1--` → limpiado |
| `TestSanitizeDirectoryFilter/SQL_injection_en_location` | `'; DROP TABLE--` → limpiado |
| `TestSanitizeDirectoryFilter/caracteres_internacionales` | `José María` preservado |
| `TestSanitizeDirectoryFilter/gender_valido_M` | `"M"` → `"M"` |
| `TestSanitizeDirectoryFilter/gender_valido_F` | `"F"` → `"F"` |
| `TestSanitizeDirectoryFilter/gender_invalido` | `"X"`, `"male"`, `""` → `""` |
| `TestSanitizeDirectoryFilter/page_default` | `0` → `1`, `-5` → `1` |
| `TestSanitizeDirectoryFilter/limit_capping` | `200` → `100`, `0` → `10` |
| `TestSanitizeDirectoryFilter/multiples_espacios` | `"a  b   c"` → `"a b c"` |
| `TestSanitizeDirectoryFilter/truncation_unicode` | String de 150 runes con ñ → truncado a 100 |

### 1.2 — `internal/service/error_mapper_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestMapDBError/duplicate_ci` | Mensaje contiene `psi_users_ci` → `ErrUniqueViolation` con "Cedula" |
| `TestMapDBError/duplicate_fpv` | Mensaje contiene `psi_users_fpv` → "Numero FPV" |
| `TestMapDBError/duplicate_email` | Mensaje contiene `psi_users_email` → "Email" |
| `TestMapDBError/duplicate_username` | Mensaje contiene `psi_users_username` → "Username" |
| `TestMapDBError/variante_idx` | `idx_psi_users_ci` funciona |
| `TestMapDBError/variante_key` | `psi_users_ci_key` funciona |
| `TestMapDBError/variante_unique` | `uni_psi_users_ci` funciona |
| `TestMapDBError/varchar_25_overflow` | `value too long for type character varying(25)` → mensaje específico |
| `TestMapDBError/generic_value_too_long` | `value too long for type varchar` → mensaje genérico |
| `TestMapDBError/uuid_invalid` | `invalid input syntax for type uuid` → mensaje UUID |
| `TestMapDBError/error_desconocido` | Error no mapeado → error original passthrough |
| `TestMapDBError/nil_error` | `nil` → `nil` |

### 1.3 — Tests existentes:扩充 Coverage

**Archivos existentes a mejorar:**
- `internal/utils/utils_test.go` — Agregar tests para `NormalizeMunicipioCarabobo`, `NormalizeEstadoVenezuela`, `ValidateEmail`
- `internal/domain/credentials_test.go` — Agregar tests para todos los sentinel errors (`errors.Is`)

---

## Fase 2: Tests Unitarios — Service Layer (Con mocks)

**Objetivo:** Cubrir lógica de negocio no cubierta con mocks existentes/patrones.

### 2.1 — `internal/service/analytics_service_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestRecordLogin/lanza_goroutine` | Verificar que llama a repo.CreateLoginEvent y repo.UpsertActiveSession |
| `TestRecordLogout/lanza_goroutine` | Verificar que llama a repo.DeleteActiveSession |
| `TestHeartbeatSession` | Verificar que llama a repo.UpdateSessionHeartbeat con times correctos |
| `TestRecordSearch` | Verificar que llama a repo.CreateSearchEvent con datos correctos |
| `TestRecordProfileView` | Verificar que llama a repo.CreateProfileView |
| `TestRecordPageView` | Verificar que llama a repo.CreatePageView |
| `TestCountRecentPageViews` | Retorna count del repo |
| `TestGetDashboardStats` | Delega al repo, retorna stats |
| `TestPurgeOldData` | Llama a delete de page_views, search_events, profile_views pero NO login_events |
| `TestCleanExpiredSessions` | Llama a repo.DeleteExpiredSessions |

### 2.2 — `internal/service/psi_service_self_management_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestUpdateProfileSelf/password_incorrecta` | Retorna `ErrPasswordIncorrect` |
| `TestUpdateProfileSelf/password_new_mismatch` | `NewPassword1 != NewPassword2` → error |
| `TestUpdateProfileSelf/password_debil` | Password débil → error de validación |
| `TestUpdateProfileSelf/password_change_exitoso` | Actualiza password + regenera key |
| `TestUpdateProfileSelf/username_duplicado` | `ValidateUniqueCredentials` retorna error |
| `TestUpdateProfileSelf/email_duplicado` | `ValidateUniqueCredentials` retorna error |
| `TestUpdateProfileSelf/municipio_invalido` | `NormalizeMunicipioCarabobo` retorna error |
| `TestUpdateProfileSelf/estado_invalido` | `NormalizeEstadoVenezuela` retorna error |
| `TestUpdateProfileSelf/minibio_truncation` | String de 300 runes → truncado a 250 |
| `TestUpdateProfileSelf/fullbio_sanitization` | HTML con `<script>` → limpiado |
| `TestUpdateProfileSelf/coldata_lazy_load` | Sin cambios en coldata → no llama GetPsiUserColData |
| `TestUpdateProfileSelf/coldata_update` | Con cambios → llama GetPsiUserColData |
| `TestUpdateProfileSelf/db_failure_s3_rollback` | Error en DB → borra archivos S3 subidos |
| `TestUpdateProfileSelf/sin_cambios_opcionales` | Archivos nil → no procesa imágenes |

### 2.3 — `internal/service/psi_service_xlsx_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestGetValorSeguro/index_in_bounds` | Retorna valor correcto |
| `TestGetValorSeguro/index_out_of_bounds` | Retorna `""` sin panic |
| `TestGetValorSeguro/empty_row` | Row nil → `""` |
| `TestImportFromXLSX/archivo_corrupto` | Reader basura → error al abrir |
| `TestImportFromXLSX/hoja_no_existe` | Falta hoja "BD ColPsiCarabobo 2026" → 0 successes |
| `TestImportFromXLSX/fila_vacia_saltada` | Row con FPV y CI vacíos → skipped |
| `TestImportFromXLSX/municipio_invalido_en_fila` | Municipio inválido → fila en failedRecords |
| `TestImportFromXLSX/username_con_email` | Email presente → username = localPart + FPV |
| `TestImportFromXLSX/username_sin_email` | Sin email → username = "psi" + FPV + CI |
| `TestImportFromXLSX/proof_of_life_fallecido` | Valor "fallecido" → ProofOfLife = false |
| `TestImportFromXLSX/email_valido_envia_welcome` | Email válido → llama mailService.SendEmail |
| `TestImportFromXLSX/email_vacio_no_envia` | Email vacío → no llama SendEmail |
| `TestImportFromXLSX/error_email_no_fatal` | SendEmail falla → warning, no aborta |

### 2.4 — Tests existentes:扩充 Coverage

**Archivos existentes a mejorar:**
- `internal/service/admin_service_test.go` — Agregar: login con usuario inactivo, login con usuario no encontrado, update con permisos insuficientes
- `internal/service/psi_service_test.go` — Agregar: login con credenciales inválidas (ya cubierto en auth_test), GetPublicProfile con perfil oculto
- `internal/service/post_service_test.go` — Agregar: UpdatePost con XSS en contenido, GetPostsList con paginación inválida

---

## Fase 3: Tests de Middleware (Con Fiber app.Test)

**Objetivo:** Cubrir toda la lógica de middleware que no tiene tests.
**Estado:** Pendiente (0/4 archivos)
**Dependencias:** Solo stdlib + Fiber + crypto/sha256. Sin DB. Sin S3.

### Orden de implementación (menor a mayor complejidad)

#### 3.0 — Pre-requisito: Config de test

`analytics.go` accede a `config.Envs.Environment` sin nil-check → paniquea si `config.Envs == nil`.
**Solución:** En cada test file que necesite config, inicializar:
```go
config.Envs = &config.Config{Environment: "test"}
```

### 3.1 — `internal/middleware/helpers_test.go` (NUEVO) — TRIVIAL

**Objetivo:** Testear extracción de usuario autenticado desde `c.Locals`.
**Complejidad:** MUY BAJA. Sin dependencias externas. Solo `fiber.Ctx`.

| Test | Escenario | Assert |
|------|-----------|--------|
| `TestGetAuthenticatedAdmin/admin_valido` | `c.Locals("admin", &domain.UserAdmin{...})` | Retorna admin, err == nil |
| `TestGetAuthenticatedAdmin/admin_missing` | Sin locals seteados | Retorna nil, err != nil, status 401 |
| `TestGetAuthenticatedAdmin/admin_nil` | `c.Locals("admin", nil)` | Retorna nil, err != nil, status 401 |
| `TestGetAuthenticatedAdmin/tipo_incorrecto` | `c.Locals("admin", "string")` | Retorna nil, err != nil, status 401 |
| `TestGetAuthenticatedPsi/psi_valido` | `c.Locals("psi_user", &domain.PsiUserModel{...})` | Retorna psi, err == nil |
| `TestGetAuthenticatedPsi/psi_missing` | Sin locals seteados | Retorna nil, err != nil, status 401 |
| `TestGetAuthenticatedPsi/psi_nil` | `c.Locals("psi_user", nil)` | Retorna nil, err != nil, status 401 |

**Patrón de test:**
```go
app := fiber.New()
app.Get("/test", func(c *fiber.Ctx) error {
    admin, err := GetAuthenticatedAdmin(c)
    if err != nil { return err }
    return c.JSON(fiber.Map{"id": admin.ID.String()})
})
// Usar app.Test(httptest.NewRequest(...)) para cada caso
```

**Estimación:** ~80 líneas, 7 tests, ~15 min

---

### 3.2 — `internal/middleware/idempotency_test.go` (NUEVO) — AUTOCONTENIDO

**Objetivo:** Testear cache de respuestas idempotentes por usuario.
**Complejidad:** MEDIA. Estructura autocontenida (in-memory map + mutex). Sin dependencias externas.
**Clave:** `IdempotencyStore` es un struct exportado con constructor → se puede testear directamente.

| Test | Escenario | Setup | Assert |
|------|-----------|-------|--------|
| `TestUserScopedIdempotency/sin_header` | Request sin `X-Idempotency-Key` | Handler retorna 200 | 200, handler ejecutado, sin header `X-Idempotent-Replayed` |
| `TestUserScopedIdempotency/cache_miss` | Header nuevo | Handler retorna 200 | 200, handler ejecutado 1 vez |
| `TestUserScopedIdempotency/cache_hit` | 2da request misma key + mismo user | Handler retorna 200 | 200, handler ejecutado SOLO 1 vez, `X-Idempotent-Replayed: true` |
| `TestUserScopedIdempotency/respuesta_error_no_cacheada` | Handler retorna 500 | Misma key | 500, 2da request ejecuta handler de nuevo (no cached) |
| `TestUserScopedIdempotency/usuarios_diferentes` | Misma key, users distintos | Handler retorna 200 | 200, handler ejecutado 2 veces (scope different) |
| `TestUserScopedIdempotency/ttl_expiry` | TTL = 1ms, esperar 10ms | Handler retorna 200 | 2da request ejecuta handler de nuevo (expired) |
| `TestUserScopedIdempotency/concurrente` | 100 goroutines, misma key | Handler con atomic counter | Counter == 1 (solo 1 ejecuta) |
| `TestScopeKey/consistencia` | Mismo input → mismo hash | N/A | SHA-256 determinístico |
| `TestScopeKey/usuarios_diferentes` | User A + User B, misma key | N/A | Hashes diferentes |
| `TestNewIdempotencyStore/state_inicial` | Constructor | N/A | Store no nil, 0 entries |
| `TestIdempotencyStore_Cleanup/ttl_expiry` | entries expiradas | Set con TTL 1ms, esperar 10ms, trigger cleanup | entries eliminadas |

**Patrón de test:**
```go
store := NewIdempotencyStore()
app := fiber.New()
app.Post("/test", UserScopedIdempency(store, 5*time.Minute), handler)
// Para cache_hit: enviar 2 requests con mismo X-Idempotency-Key
// Para usuarios_diferentes: variar el c.Locals("admin") ID
```

**Mock para user context:**
```go
// Helper para inyectar admin ID en el contexto
func injectAdminID(c *fiber.Ctx, id uuid.UUID) error {
    c.Locals("admin", &domain.UserAdmin{ID: id})
    return c.Next()
}
```

**Estimación:** ~200 líneas, 11 tests, ~45 min

---

### 3.3 — `internal/middleware/analytics_test.go` (NUEVO) — MEDIUM

**Objetivo:** Testear tracking de page views y debouncing.
**Complejidad:** MEDIA. Goroutines + config global + cookies.
**Dependencias:** `*service.AnalyticsService` (con mocks), `config.Envs`.

**Pre-requisito:**
```go
func TestMain(m *testing.M) {
    config.Envs = &config.Config{Environment: "test"}
    os.Exit(m.Run())
}
```

**Mock necesario:**
```go
type mockAnalyticsRepo struct {
    domain.AnalyticsRepository
    CountRecentPageViewsFunc func(sessionID uuid.UUID, since time.Time) (int64, error)
    CreatePageViewFunc       func(view domain.PageView) error
    // ... otros métodos con nil-check default
}
```

| Test | Escenario | Setup | Assert |
|------|-----------|-------|--------|
| `TestShouldSkip/health` | path = "/health" | N/A | true |
| `TestShouldSkip/static` | path = "/static/app.js" | N/A | true |
| `TestShouldSkip/metrics` | path = "/metrics" | N/A | true |
| `TestShouldSkip/normal` | path = "/api/users" | N/A | false |
| `TestShouldSkip/favicon` | path = "/favicon.ico" | N/A | true |
| `TestAnalyticsMiddleware/metodo_no_get_saltado` | POST request | Handler retorna 200 | CreatePageView NO llamado |
| `TestAnalyticsMiddleware/path_skipped` | GET /health | Handler retorna 200 | CreatePageView NO llamado |
| `TestAnalyticsMiddleware/admin_saltado` | GET / con `c.Locals("admin")` | Handler retorna 200 | CreatePageView NO llamado |
| `TestAnalyticsMiddleware/primera_visita_genera_cookie` | GET / sin `_sid` cookie | Handler retorna 200 | Response tiene Set-Cookie `_sid` |
| `TestAnalyticsMiddleware/cookie_existente_reutilizada` | GET / con `_sid` cookie | Handler retorna 200 | Misma cookie value en response |
| `TestAnalyticsMiddleware/registra_page_view` | GET /api/users | CreatePageView retorna nil | CreatePageView llamado con path correcto |
| `TestAnalyticsMiddleware/debouncing_dentro_ventana` | 2 GETs en <30min | CountRecentPageViews retorna 1 | CreatePageView NO llamado en 2da request |
| `TestAnalyticsMiddleware/debouncing_fuera_ventana` | 2 GETs >30min aparte | CountRecentPageViews retorna 0 | CreatePageView llamado en ambas |
| `TestAnalyticsMiddleware/usuario_logueado` | `c.Locals("userID")` seteado | Handler retorna 200 | PageView.UserID no nil |

**Nota sobre goroutines:** El `go func()` en analytics.go hace fire-and-forget. Usar `time.Sleep(50ms)` después de `app.Test()` para asegurar que el goroutine complete.

**Estimación:** ~280 líneas, 14 tests, ~60 min

---

### 3.4 — `internal/middleware/rate_limiter_test.go` (NUEVO) — COMPLICADO POR sync.Once

**Objetivo:** Testear limitación de tasa por IP.
**Complejidad:** ALTA debida a `sync.Once` (el storage se inicializa una sola vez).
**Estrategia:** Testear solo el comportamiento observable (HTTP responses) sin resetear el storage.

**Limitación conocida:** No se puede testear `newRateLimiterStorage()` directamente. Los tests deben correr en orden o aceptar que el rate limiter state persiste entre sub-tests.

| Test | Escenario | Setup | Assert |
|------|-----------|-------|--------|
| `TestAuthRateLimiter/10_requests_ok` | 10 POSTs desde misma IP | Limpiar storage (si es in-memory) | Todas retornan 200 |
| `TestAuthRateLimiter/11_request_bloqueado` | 11 POSTs | Misma IP | 11ª retorna 429 |
| `TestAuthRateLimiter/get_no_cuenta` | GET request | N/A | 200 (GET no incrementa) |
| `TestAuthRateLimiter/options_bypass` | OPTIONS request | N/A | 200 (CORS preflight) |
| `TestAdminAuthRateLimiter/5_requests_ok` | 5 POSTs | Limpiar storage | Todas 200 |
| `TestAdminAuthRateLimiter/6_request_bloqueado` | 6 POSTs | Misma IP | 6ª retorna 429 |
| `TestAdminAuthRateLimiter/mensaje_espanol` | 6 POSTs | Leer body del 429 | Body contiene "demasiadas peticiones" |

**Desafío técnico:**
- `sync.Once` impide reinicializar el storage entre tests
- **Solución:** Usar `t.Cleanup` +Aceptar que los tests de rate_limiter deben correr en un solo `TestRateLimiterSuite` secuencial, o crear un test helper que resetee el state interno (si el storage es in-memory `limiter.New` de Fiber, esto no es trivial)
- **Alternativa pragmática:** Testear con una sola suite que acumule requests, o usar `httptest` con IPs diferentes para simular clientes distintos

**Estimación:** ~150 líneas, 7 tests, ~40 min

---

### Resumen Fase 3

| Archivo | Tests | Líneas estimadas | Tiempo |
|---------|-------|-----------------|--------|
| `helpers_test.go` | 7 | ~80 | 15 min |
| `idempotency_test.go` | 11 | ~200 | 45 min |
| `analytics_test.go` | 14 | ~280 | 60 min |
| `rate_limiter_test.go` | 7 | ~150 | 40 min |
| **Total** | **39** | **~710** | **~2.5 hrs** |

### Riesgos conocidos

1. **`config.Envs` nil panic** en `analytics.go:95` — Se resuelve con `TestMain` o setup por test
2. **`sync.Once` en rate_limiter.go** — No resettable. Tests secuenciales o IPs diferentes
3. **Goroutines fire-and-forget** — Requieren `time.Sleep` para sincronización
4. **Cookie `_sid` en analytics** — Requiere inspeccionar `Response.Header` de Fiber

---

## Fase 4: Tests de Handlers (HTTP end-to-end con mocks)

**Objetivo:** Tests de capa HTTP con `fiber.App.Test()`, mocks en services.

### 4.1 — `internal/handler/admin_handler_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestLogin/password_correcta` | 200 + JWT |
| `TestLogin/password_incorrecta` | 401 |
| `TestLogin/cuerpo_invalido` | 400 (JSON malformado) |
| `TestLogin/rate_limit` | 6º intento → 429 |
| `TestCreateAdmin/por_sudo` | 201 |
| `TestCreateAdmin/por_admin_sin_permiso` | 403 |
| `TestCreateAdmin/jerarquia_no_puede_dar_permisos_propios` | 403 |
| `TestGetAdmins/con_filtros` | 200 + paginación |
| `TestGetAdmins/sin_auth` | 404 (obscurity) |
| `TestUpdateAdmin/exitoso` | 200 |
| `TestUpdateAdmin/self_update` | 200 |
| `TestDeleteAdmin/exitoso` | 200 |
| `TestDeleteAdmin/no_puede_borrar_sudo` | 403 |
| `TestDeleteAdmin/no_puede_borrar_self` | 403 |
| `TestLogout/exitoso` | 200 |
| `TestLogout/sin_auth` | 401 |

### 4.2 — `internal/handler/psi_handler_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestLogin/credenciales_validas` | 200 + JWT |
| `TestLogin/usuario_inactivo` | 401 |
| `TestSearchDirectory/parametros_validos` | 200 + resultados |
| `TestSearchDirectory/sanitizacion_inyeccion` | SQL injection en q → sanitizado |
| `TestGetPublicProfile/exitoso` | 200 + perfil público |
| `TestGetPublicProfile/no_existe` | 404 |
| `TestGetMe/exitoso` | 200 + perfil propio |
| `TestGetMe/sin_auth` | 401 |
| `TestUpdateOwnProfile/password_requerida` | Sin password → 401 |
| `TestUpdateOwnProfile/password_incorrecta` | Password incorrecta → 401 |
| `TestUpdateOwnProfile/exitoso` | 200 |
| `TestAddPostGrade/exitoso` | 201 |
| `TestAddSocialNetwork/cuota_maxima` | 11 redes → error |
| `TestDeleteSocialNetwork/exitoso` | 200 |
| `TestDeleteSocialNetwork/otro_usuario` | 403 |
| `TestGetSitemapData` | 200 |

### 4.3 — `internal/handler/posts_handler_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestListPosts/publico` | Sin auth → solo publicados |
| `TestListPosts/admin` | Con admin auth → ve drafts |
| `TestListPosts/psi` | Con psi auth → ve publicados + propios drafts |
| `TestListPosts/paginacion` | page/limit válidos → 200 |
| `TestListPosts/limit_excesivo` | limit > 100 → capped a 100 |
| `TestCreatePost/exitoso` | Admin con CanPublish → 201 |
| `TestCreatePost/sin_permiso` | Admin sin CanPublish → 403 |
| `TestGetPost/encontrado` | 200 |
| `TestGetPost/no_encontrado` | 404 |
| `TestUpdatePost/exitoso` | 200 |
| `TestUpdatePost/scheduled_publishing` | publish_at futuro → status=scheduled |

### 4.4 — `internal/handler/specialty_handler_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestGetSpecialties/activas` | Sin auth → solo activas |
| `TestGetSpecialtyByID/exitoso` | 200 |
| `TestGetSpecialtyByID/no_existe` | 404 |
| `TestCreateSpecialty/con_permiso` | CanCreateTags → 201 |
| `TestCreateSpecialty/sin_permiso` | Sin CanCreateTags → 403 |
| `TestUpdateSpecialty/exitoso` | 200 |
| `TestDeleteSpecialty/exitoso` | 200 |
| `TestCountSpecialties/con_admin` | Admin con permisos → count total |
| `TestCountSpecialties/sin_admin` | Público → solo activas |

### 4.5 — `internal/handler/analytics_handler_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestGetDashboardStats/exitoso` | Admin autenticado → 200 + stats |
| `TestGetDashboardStats/sin_auth` | Sin auth → 404 |

---

## Fase 5: Tests de Repository (PostgreSQL real)

**Objetivo:** Cubrir repos sin tests existentes.

### 5.1 — `internal/repository/postgres/analytics_repo_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestCreateLoginEvent` | Inserta y verifica |
| `TestUpsertActiveSession/primera_vez` | Create cuando no existe |
| `TestUpsertActiveSession/update` | Update cuando ya existe |
| `TestDeleteActiveSession` | Borra y verifica que no existe |
| `TestUpdateSessionHeartbeat` | Actualiza last_seen y expires_at |
| `TestCreateSearchEvent` | Inserta y verifica |
| `TestCreateProfileView` | Inserta y verifica |
| `TestCreatePageView` | Inserta y verifica |
| `TestCountRecentPageViews/dentro_ventana` | Count correcto |
| `TestCountRecentPageViews/fuera_ventana` | Count = 0 |
| `TestGetDashboardStats` | Verificar estructura (sin validar valores exactos) |
| `TestDeletePageViewsOlderThan` | Borra solo los antiguos |
| `TestDeleteSearchEventsOlderThan` | Borra solo los antiguos |
| `TestDeleteProfileViewsOlderThan` | Borra solo los antiguos |
| `TestDeleteExpiredSessions` | Borra solo expiradas |

### 5.2 — Tests existentes:扩充 Coverage

**Archivos existentes a mejorar:**
- `psi_repo_test.go` — Agregar: CreateSocialNetwork, UpdateSocialNetwork, CreatePostGrade, UpdatePostGrade, GetSolvency, ValidateUniqueCredentials
- `post_repo_test.go` — Agregar: Create, Update, GetByID, GetSitemapData, PublishScheduled
- `user_admin_repo_test.go` — Agregar: Create con duplicado (unique violation), GetSudoExists

---

## Fase 6: Tests de Integración End-to-End

**Objetivo:** Flujo completo request → handler → service → repo → DB.

### 6.1 — `internal/integration/admin_flow_test.go` (NUEVO)

```
Flujo 1: Login → Crear Admin → Listar → Actualizar → Eliminar
Flujo 2: Login → Crear Admin → Intentar borrar Sudo → 403
Flujo 3: Login → Crear Admin sin permisos → Intentar crear psi → 403
```

### 6.2 — `internal/integration/psi_flow_test.go` (NUEVO)

```
Flujo 1: Admin crea psi → Admin lista → Admin edita → Admin borra
Flujo 2: Admin crea psi → Psi login → Psi ve perfil → Psi actualiza → Psi logout
Flujo 3: Admin crea psi → Psi login → Psi agrega postgrado → Psi agrega red social
Flujo 4: Admin crea psi → Psi login → Psi cambia password → Token viejo invalidado
```

### 6.3 — `internal/integration/post_flow_test.go` (NUEVO)

```
Flujo 1: Admin crea post → Público ve post → Admin actualiza → Admin archiva
Flujo 2: Admin crea post scheduled → PublishScheduled lo publica
Flujo 3: Público intenta ver draft → No lo ve
```

### 6.4 — `internal/integration/directory_flow_test.go` (NUEVO)

```
Flujo 1: Admin crea psi público → Directorio lo muestra → Búsqueda funciona
Flujo 2: Admin crea psi oculto → Directorio no lo muestra
Flujo 3: Admin crea psi → Cambia visibilidad → Verificar filtrado
Flujo 4: Sitemap data incluye solo activos + solventes
```

### 6.5 — `internal/integration/specialty_flow_test.go` (NUEVO)

```
Flujo 1: Admin crea especialidad → Asigna a psi → Psi aparece con especialidad
Flujo 2: Admin desactiva especialidad → Directorio público no la muestra
Flujo 3: Admin crea especialidad duplicada → Unique violation
```

---

## Fase 7: Tests de Seguridad

**Objetivo:** Verificar que los mecanismos de seguridad funcionan correctamente.

### 7.1 — `internal/security/auth_security_test.go` (NUEVO)

| Test | Escenario |
|------|-----------|
| `TestJWT_none_algorithm_attack` | Token con alg:"none" → rechazado |
| `TestJWT_firmado_con_otra_key` | Token firmado con key de otro usuario → rechazado |
| `TestJWT_token_expirado` | Token exp → 401 |
| `TestKeyRotation_logout_invalida_tokens` | Login → Logout → Token viejo → rechazado |
| `TestKeyRotation_doble_login` | Login 1 → Login 2 → Token 1 → rechazado |
| `TestRBAC/no_puede_dar_permisos_no_poseidos` | Admin sin CanDeleteAdmin → intentar crear admin con CanDeleteAdmin → 403 |
| `TestRBAC/sudo_irrevocable` | Intentar borrar Sudo → 403 |
| `TestRBAC/hierarchy_escalation` | Admin A intenta dar permiso que Admin B no tiene → 403 |
| `TestAdmin404/obscurity` | Token inválido en admin endpoint → 404 (no 401) |
| `TestPsiAuth/token_invalido` | Token inválido en psi endpoint → 401 |
| `TestPassword/weak_rechazado` | Password débil → error de validación |

---

## Fase 8: Cobertura y Benchmarks

### 8.1 — Cobertura

**Target:** ≥70% cobertura global, ≥80% en service layer.

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 8.2 — Benchmarks

**Archivo nuevo:** `internal/service/benchmarks_test.go`

```go
func BenchmarkMapDBError(b *testing.B)          // 100+ variants
func BenchmarkSanitizeDirectoryFilter(b *testing.B)
func BenchmarkIsStrongPassword(b *testing.B)
func BenchmarkGenerateSecureRandomString(b *testing.B)
```

### 8.3 — Race Detector

```bash
go test ./... -race -count=5
```

---

## Orden de Ejecución Recomendado

| Fase | Dependencias | Tiempo estimado | Prioridad |
|------|-------------|-----------------|-----------|
| **Fase 0** | Ninguna | 1 día | CRÍTICO |
| **Fase 1** | Fase 0 | 0.5 días | ALTA |
| **Fase 2** | Fase 0 | 1.5 días | ALTA |
| **Fase 3** | Fase 0 | 1 día | ALTA |
| **Fase 4** | Fase 0 | 2 días | ALTA |
| **Fase 5** | Fase 0 | 1.5 días | MEDIA |
| **Fase 6** | Fases 0-5 | 2 días | MEDIA |
| **Fase 7** | Fases 0-4 | 1 día | ALTA |
| **Fase 8** | Todas | 0.5 días | BAJA |

**Total estimado: ~11 días**

---

## Criterios de Aceptación

1. `make test-unit` pasa sin dependencias externas
2. `make test-repo` pasa con PostgreSQL corriendo en localhost:5432
3. `make test-all` pasa con el stack completo (Docker Compose)
4. `make test-race` sin data races detectados
5. Cobertura global ≥70%
6. Todos los tests existentes siguen pasando (regression)
7. Ningún test nuevo depende de datos de otros tests (aislamiento)
8. Tiempo total de tests < 5 minutos (unit + repo)

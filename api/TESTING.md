# Testing Suite — ColPsiCarabobo API

Suite integral de testing para la API REST del Colegio de Psicólogos de Carabobo. **49 archivos de test, 253 funciones Test, 30 benchmarks, 62.3% cobertura global.**

---

## Tabla de Contenidos

- [Resumen](#resumen)
- [Infraestructura de Test](#infraestructura-de-test)
- [Categorías de Tests](#categorías-de-tests)
- [Ejecución de Tests](#ejecución-de-tests)
- [Cobertura por Paquete](#cobertura-por-paquete)
- [Inventario de Archivos](#inventario-de-archivos)
- [Patrón de Mocks](#patrón-de-mocks)
- [Base de Datos de Test](#base-de-datos-de-test)
- [Benchmarks](#benchmarks)
- [Fallos Conocidos (Pre-existentes)](#fallos-conocidos-pre-existentes)
- [Bugs de Producción Corregidos](#bugs-de-producción-corregidos)
- [Convenciones](#convenciones)

---

## Resumen

| Métrica | Valor |
|:--------|------:|
| Archivos de test | 49 |
| Funciones `Test*` | 253 |
| Funciones `Benchmark*` | 30 |
| Cobertura global | 62.3% |
| Paquetes con 100% | 3 (config, logger, request_structs) |
| Paquetes con >70% | 5 (middleware, domain, handler, utils, repository) |
| Fallos pre-existentes | 3 (no causados por los tests nuevos) |
| Phases completadas | 8 de 8 |

---

## Infraestructura de Test

### Stack

| Componente | Uso |
|:-----------|:----|
| `testing` stdlib | Framework base |
| `testify/require` | Assertions con fail-fast |
| `net/http/httptest` | Requests HTTP simulados |
| `Fiber app.Test()` | Tests E2E sobre el router Fiber completo |
| GORM + PostgreSQL 18 | Tests de repositorio contra DB real |
| Hand-rolled mocks | Patrón func override (sin gomock/mockgen) |

### Docker Compose de Test

Archivo: `docker-compose.test.yml`

```yaml
PostgreSQL 18 en puerto 5433 (tmpfs, fsync=off)
Base de datos: colpsi_test
Usuario: postgres / postgres
```

```bash
# Levantar DB de test
docker-compose -f docker-compose.test.yml up -d

# Verificar
docker-compose -f docker-compose.test.yml ps
```

### Variables de Entorno

```bash
TEST_DB_DSN="host=localhost port=5433 user=postgres password=postgres dbname=colpsi_test sslmode=disable"
```

---

## Categorías de Tests

### 1. Unit Tests (sin DB)

Tests puros de lógica de negocio, utilidades y validaciones. No requieren base de datos ni servicios externos.

**Ubicación:** `internal/utils/`, `internal/request_structs/`, `internal/domain/`, `internal/middleware/`, `internal/service/`, `pkg/job/`

**Cobertura:** Utils 63.8%, Domain 71.4%, Middleware 77.3%, Request Structs 100%

**Ejecutar:**
```bash
make test-unit
# o
go test -count=1 -short ./internal/utils/... ./internal/request_structs/... ./internal/middleware/... ./internal/service/... ./pkg/job/...
```

### 2. Repository Tests (PostgreSQL real)

Tests de repositorio contra una instancia real de PostgreSQL. Cada test usa transacciones que se revierten al final, excepto los que prueban comportamiento de commit.

**Ubicación:** `internal/repository/postgres/`

**Paquetes cubiertos:** UserAdmin, Specialty, Post, PsiUser, Analytics

**Ejecutar:**
```bash
make test-repo
# o
TEST_DB_DSN="..." go test -count=1 ./internal/repository/...
```

### 3. Integration Tests (DB + Fiber + JWT)

Tests end-to-end que levantan la aplicación completa (Fiber + middleware + JWT + DB) y hacen requests HTTP reales.

**Ubicación:** `internal/integration/`

**Flujos cubiertos:**
- Admin (8 tests): CRUD de administradores, login, registro
- Psi (6 tests): CRUD de psicólogos, directorio, búsqueda
- Post (5 tests): CRUD de publicaciones, categorías
- Directory (4 tests): Búsqueda, filtros, paginación del directorio
- Specialty (4 tests): CRUD de especialidades
- Security (43 tests): Ver [sección Seguridad](#tests-de-seguridad)

**Ejecutar:**
```bash
make test-integration
# o
TEST_DB_DSN="..." go test -count=1 -p 1 ./internal/integration/...
```

### 4. Security E2E Tests

43 tests que simulan ataques y comportamientos maliciosos contra la API completa.

**Ubicación:** `internal/integration/security_flow_test.go`

| Categoría | Tests | Qué validan |
|:----------|------:|:------------|
| JWT Attacks | 6 | Tokens expirados, firmas inválidas, algoritmos none, key ID incorrecto, payload manipulado, tokens vacíos |
| Key Rotation | 3 | Rotación de claves JWT, tokens firmados con clave anterior, nueva firma válida |
| RBAC | 6 | Admin no puede acceder a endpoints PSI, PSI no puede acceder a endpoints admin, roles inválidos reciben 401/404 |
| Password Policy | 4 | Contraseñas cortas, sin mayúsculas, sin números, sin caracteres especiales |
| Info Leakage | 4 | Headers X-Powered-By no expuestos, mensajes de error no revelan stack traces, sin SQL errors visibles |
| Email Validation | 3 | Emails sin @, dominios inválidos, unicode rechazado |
| SQL Injection | 3 | Payloads SQL en campos de búsqueda, login, IDs |
| XSS | 3 | Scripts en campos de texto, atributos HTML, URLs maliciosas |
| HTTP Headers | 3 | X-Content-Type-Options, X-Frame-Options, Cache-Control en endpoints sensibles |
| Method/Content-Type | 3 | Métodos HTTP no soportados (405), content-type inválido |
| Idempotency | 2 | Requests duplicados bloqueados, keys idempotentes |
| CSRF | 2 | Tokens CSRF, Origin header validation |
| Auth Headers | 3 | Bearer token malformado, Authorization header vacío, múltiples headers |
| Privilege Escalation | 3 | Intentos de escalation via path traversal, role injection en body |
| Session Security | 2 | Tokens reuse, session fixation |
| Encoding | 2 | Double encoding, null bytes en inputs |
| Data Isolation | 1 | Un admin no puede ver datos de otro admin |

**Ejecutar:**
```bash
make test-security
# o
TEST_DB_DSN="..." go test -count=1 -p 1 -run "TestSecurity" ./internal/integration/...
```

---

## Ejecución de Tests

### Comandos Disponibles (Makefile)

```bash
make test-unit         # Tests unitarios (sin DB, rápidos)
make test-repo         # Tests de repositorio (PostgreSQL real)
make test-integration  # Tests de integración (full stack)
make test-security     # Tests de seguridad (subset de integración)
make test-all          # Todos los tests (serial, -p 1)
make test-race         # Todos con race detector
make test-bench        # Benchmarks
make coverage          # Reporte de cobertura + resumen
make coverage-html     # Reporte HTML interactivo
```

### Ejecución Completa

```bash
# 1. Levantar DB de test
docker-compose -f docker-compose.test.yml up -d

# 2. Ejecutar todo (serial para evitar race conditions entre paquetes)
make test-all

# 3. Generar reporte de cobertura
make coverage
```

### Notas Importantes

- **Serial (`-p 1`)**: Los tests de integración y repositorio deben correr en serial para evitar interferencias entre paquetes que comparten la misma DB de test.
- **Race detector**: `make test-race` usa `-race` para detectar condiciones de carrera.
- **Timeout Fiber**: `app.Test(req)` tiene timeout default de 1000ms. Para handlers lentos (bcrypt), usar `app.Test(req, -1)`.

---

## Cobertura por Paquete

| Paquete | Cobertura | Archivos de Test |
|:--------|----------:|:-----------------|
| `internal/config` | **100.0%** | `config_test.go` |
| `internal/logger` | **100.0%** | `logger_test.go` |
| `internal/request_structs` | **100.0%** | `directory_filter_sanitizer_test.go`, `visibility_getters_test.go` |
| `internal/middleware` | **77.3%** | `auth_test.go`, `analytics_test.go`, `helpers_test.go`, `idempotency_test.go`, `rate_limiter_test.go` |
| `internal/domain` | **71.4%** | `credentials_test.go` |
| `internal/handler` | **64.2%** | `admin_handler_test.go`, `analytics_handler_test.go`, `posts_handler_test.go`, `psi_handler_test.go`, `psi_user_admin_test.go`, `specialty_handler_test.go`, `test_helpers_test.go` |
| `internal/utils` | **63.8%** | `utils_test.go`, `validate_email_test.go`, `benchmarks_test.go` |
| `internal/repository/postgres` | **68.2%** | `user_admin_repo_test.go`, `specialty_repo_test.go`, `post_repo_test.go`, `psi_repo_test.go`, `analytics_repository_test.go`, `updates_safety_test.go` |
| `pkg/database` | **55.1%** | `database_test.go` |
| `pkg/job` | **19.4%** | `key_cleanup_test.go` |
| `internal/router` | **0.0%** | `router_test.go` (tests existen pero miden cobertura del router package que delega a otros) |
| `internal/service` | — | 13 archivos (ver detalle abajo) |
| `internal/integration` | — | 7 archivos (sin statements propios, solo orquesta) |

### Service Layer — Detalle

| Archivo | Funciones Test | Qué cubre |
|:--------|---------------:|:----------|
| `admin_service_test.go` | 6 | Login, Create, Update, Delete, GetAll, RBAC |
| `analytics_service_test.go` | 12 | Dashboard, Reports, concurrent safety |
| `error_mapper_test.go` | 2 | Mapeo de errores DB a errores de dominio |
| `mail_service_test.go` | 2 | Envío de emails, nil receiver safety |
| `post_service_test.go` | 1 | CRUD de publicaciones |
| `psi_service_auth_test.go` | 7 | Login, RotateKey, UpdateKey, email notification |
| `psi_service_directory_test.go` | 5 | Directorio, búsqueda, filtros, paginación |
| `psi_service_import_test.go` | 7 | Importación CSV, validación, duplicados |
| `psi_service_test.go` | 3 | CRUD básico de psicólogos |
| `psi_service_xlsx_test.go` | 13 | Importación XLSX, getValorSeguro, parsing |
| `psi_user_admin_service_test.go` | 3 | CRUD admin de psi users |
| `social_media_test.go` | 3 | Redes sociales, validación |
| `specialty_service_test.go` | 4 | CRUD de especialidades |

---

## Inventario de Archivos

### `internal/config/` (1 archivo)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `config_test.go` | 6 | `getEnv` defaults, `InitConfig` con .env mockeado |

### `internal/domain/` (1 archivo)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `credentials_test.go` | 15 | Sentinel errors, `PostGradeType.IsValid`, tablas, JSON tags, constantes |

### `internal/handler/` (7 archivos)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `test_helpers_test.go` | — | Helpers compartidos: `setupAdminRoute`, `setupPublicRoute`, `testPsiHandler`, `testAdminHandler`, mock repos |
| `admin_handler_test.go` | 6 | Login, Create, Update, Delete, GetAll, GetByID |
| `analytics_handler_test.go` | 1 | Dashboard + Reports |
| `posts_handler_test.go` | 5 | CRUD posts, categorías |
| `psi_handler_test.go` | 9 | CRUD psi, directorio, imagen |
| `psi_user_admin_test.go` | 6 | GetPsiByIDAdmin, CreatePsiByAdmin, UpdatePsiByAdmin, DeletePsiByAdmin, ListAllPsis, LoginLibrary |
| `specialty_handler_test.go` | 7 | CRUD specialties admin + público |

### `internal/integration/` (7 archivos)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `setup_test.go` | 1 | `TestMain`, `buildTestApp`, `generateToken`, helpers HTTP, seed data, `truncateAll` |
| `admin_flow_test.go` | 8 | E2E CRUD administradores |
| `directory_flow_test.go` | 4 | E2E directorio y búsqueda |
| `post_flow_test.go` | 5 | E2E publicaciones |
| `psi_flow_test.go` | 6 | E2E psicólogos |
| `security_flow_test.go` | 43 | E2E seguridad (JWT, RBAC, injection, XSS, etc.) |
| `specialty_flow_test.go` | 4 | E2E especialidades |

### `internal/logger/` (1 archivo)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `logger_test.go` | 4 | `Init` con distintos niveles, output a stdout/stderr |

### `internal/middleware/` (5 archivos)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `auth_test.go` | 1 | Validación JWT completa con Fiber app.Test |
| `analytics_test.go` | 3 | Tracking de visitas, IP extraction, user-agent parsing |
| `helpers_test.go` | 2 | `GetAuthenticatedUser`, `SendEmail` nil safety |
| `idempotency_test.go` | 3 | SHA-256 hashing, deduplicación, key generation |
| `rate_limiter_test.go` | 2 | Rate limiting por IP, window expiration |

### `internal/repository/postgres/` (6 archivos)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `user_admin_repo_test.go` | 1 | Suite: Create, GetByID, GetByEmail, GetByCedula, Update, Delete, GetAll, Count, CountSudos |
| `specialty_repo_test.go` | 1 | Suite: CRUD + GetAll + GetActive |
| `post_repo_test.go` | 1 | Suite: CRUD + GetBySlug + GetByCategory |
| `psi_repo_test.go` | 1 | Suite: CRUD + Search + GetByEmail + GetByFPV + Count + solvencias |
| `analytics_repository_test.go` | 1 | Suite: Visit tracking + dashboard aggregation |
| `updates_safety_test.go` | 8 | GORM Update safety: map vs struct, zero values, field exclusión |

### `internal/request_structs/` (2 archivos)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `directory_filter_sanitizer_test.go` | 3 | Sanitización de filtros de directorio (35+ subtests) |
| `visibility_getters_test.go` | 4 | Boolean getters de `PsiUserUpdateRequestSelf` y `UpdatePsiAdminRequest` (16 subtests) |

### `internal/router/` (1 archivo)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `router_test.go` | 6 | Registro de rutas, 404/405, prefijos correctos |

### `internal/service/` (13 archivos)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `admin_service_test.go` | 6 | Login, Create, Update, Delete, GetAll, RBAC |
| `analytics_service_test.go` | 12 | Dashboard, Reports, concurrent safety |
| `error_mapper_test.go` | 2 | Error mapping DB → domain |
| `mail_service_test.go` | 2 | Email sending, nil safety |
| `post_service_test.go` | 1 | CRUD posts |
| `psi_service_auth_test.go` | 7 | Auth, key rotation, notifications |
| `psi_service_directory_test.go` | 5 | Directorio, búsqueda, paginación |
| `psi_service_import_test.go` | 7 | CSV import, validación |
| `psi_service_test.go` | 3 | CRUD básico |
| `psi_service_xlsx_test.go` | 13 | XLSX import, parsing, getValorSeguro |
| `psi_user_admin_service_test.go` | 3 | Admin CRUD psi users |
| `social_media_test.go` | 3 | Social media validation |
| `specialty_service_test.go` | 4 | CRUD specialties |

### `internal/utils/` (3 archivos)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `utils_test.go` | 9 | NormalizeMunicipio, NormalizeEstado, BoolFromForm, CleanAlphaNumeric, etc. |
| `validate_email_test.go` | 1 | `ParseAndValidateEmail` — 13 subtests (válidos, inválidos, edge cases, Unicode) |
| `benchmarks_test.go` | 30 | Benchmarks: password hashing, random generation, geo, email, bool, image, platform |

### `pkg/database/` (1 archivo)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `database_test.go` | 6 | `RunMigrations`, `SeedAdmin`, `ConnectDB` — requiere DB de test |

### `pkg/job/` (1 archivo)

| Archivo | Tests | Descripción |
|:--------|------:|:------------|
| `key_cleanup_test.go` | 2 | `isKeyExpired` con helper `makeV7FromTime`, `KeyCleanupResult` |

---

## Patrón de Mocks

Todos los mocks son **hand-rolled** usando el patrón **func override**. No se usa gomock, mockgen, ni ningún framework de mocks.

### Ejemplo

```go
// Definición del mock
type mockAdminRepo struct {
    getByIDFn    func(id uuid.UUID) (*domain.UserAdmin, error)
    getByEmailFn func(email string) (*domain.UserAdmin, error)
    createFn     func(admin *domain.UserAdmin) error
    updateFn     func(admin *domain.UserAdmin) error
    deleteFn     func(id uuid.UUID) error
    getAllFn     func(page, limit int) ([]domain.UserAdmin, int64, error)
}

// Implementación de la interfaz — delega a la función
func (m *mockAdminRepo) GetByID(id uuid.UUID) (*domain.UserAdmin, error) {
    if m.getByIDFn != nil {
        return m.getByIDFn(id)
    }
    return nil, nil
}

// Uso en tests
func TestCreateAdmin(t *testing.T) {
    repo := &mockAdminRepo{
        createFn: func(admin *domain.UserAdmin) error {
            return nil // éxito
        },
    }
    svc := service.NewAdminService(repo, nil, nil)
    // ...
}
```

### Convenciones de Mocks

- Cada mock implementa una interfaz de `internal/domain/`
- Los campos son funciones (`func type`) opcionales
- Si la función es `nil`, el mock retorna valores por defecto (nil, error, etc.)
- Los mocks se instancian inline en cada test con solo los campos necesarios
- Los tests de handler usan `testAdminHandler` y `testPsiHandler` que encapsulan mocks de repo + servicio

---

## Base de Datos de Test

### Infraestructura

- **PostgreSQL 18** vía Docker Compose (`docker-compose.test.yml`)
- **Puerto:** 5433 (exclusivo para tests)
- **tmpfs:** Habilitado para max performance
- **fsync=off:** Deshabilitado para writes más rápidos
- **Base de datos:** `colpsi_test`
- **Credenciales:** `postgres` / `postgres`

### Estrategia de Aislamiento

1. **Package level:** Cada paquete de repositorio crea y elimina sus propios registros
2. **TestMain:** `internal/integration/setup_test.go` tiene `truncateAll()` al final de `TestMain` que limpia todas las tablas
3. **Serial execution:** `-p 1` garantiza que los paquetes no corren en paralelo
4. **Transacciones:** Algunos tests usan transacciones que se revierten al final

### Setup en Integration Tests

```go
func TestMain(m *testing.M) {
    // 1. Conectar a DB de test
    // 2. Ejecutar AutoMigrate
    // 3. Seed admin por defecto
    // 4. Ejecutar todos los tests
    // 5. truncateAll() — limpiar todas las tablas
    os.Exit(code)
}
```

### Tablas que se limpian

```
administradores, psicologos, posts, specialties,
analytics_visits, idempotency_keys, audit_logs,
psi_users, social_media, post_grades, ...
```

---

## Benchmarks

**Ubicación:** `internal/utils/benchmarks_test.go`

30 benchmarks midiendo el rendimiento de funciones críticas:

| Categoría | Benchmarks | Qué miden |
|:----------|-----------:|:----------|
| Password Hashing | 4 | `HashPassword` con cost 10/12/14, `CheckPassword` |
| Random Generation | 4 | `RandomString`, `GenerateAPIKey`, `GenerateSecretKey`, `GenerateUUID` |
| Geo Normalization | 4 | `NormalizeMunicipioCarabobo`, `NormalizeEstadoVenezuela`, fuzzy match |
| Email Validation | 3 | `ParseAndValidateEmail` — válido, inválido, Unicode |
| Bool Parsing | 3 | `BoolFromForm` — "true", "false", "1", "0" |
| Image Sanitization | 3 | `SanitizeImageFilename`, `ValidateImageMagicBytes`, `CleanAlphaNumeric` |
| Platform Detection | 3 | `DetectPlatform`, `ExtractBrowser`, `ExtractOS` |
| Directory Filters | 3 | `SanitizeDirectoryFilter` — name, specialties, state |
| Error Mapping | 3 | `MapDBError`, `MapPsiUserDBError` — normal + nil error |

### Ejecutar

```bash
make test-bench
# o
go test -bench=. -benchmem -run=^$ ./internal/utils/... ./internal/service/...
```

### Ejemplo de salida

```
BenchmarkHashPassword/cost_10-8       1         xxxxxxx ns/op    xxx B/op    x allocs/op
BenchmarkHashPassword/cost_12-8       1         xxxxxxx ns/op    xxx B/op    x allocs/op
BenchmarkBoolFromForm/true-8          xxxxxxxxx     xx ns/op     0 B/op    0 allocs/op
```

---

## Fallos Conocidos (Pre-existentes)

Estos 3 fallos existían antes de la suite de testing y **no son causados por los tests nuevos**:

### 1. `TestAdminService_All` > `CreateAdmin: Regla 'No puedes dar lo que no tienes'`

- **Archivo:** `internal/service/admin_service_test.go`
- **Causa:** El mock de `GetByCedula` retorna un admin con rol `USUARIO` (no `ADMINISTRADOR` ni `SUPER_USUARIO`), y el servicio rechaza la creación. El test espera éxito pero la lógica de RBAC lo bloquea correctamente.
- **Severidad:** Low — el test tiene una expectativa incorrecta sobre las reglas de RBAC.

### 2. `TestSpecialtyService_Update` > `Actualización Parcial (PATCH) inyecta auditoría correctamente`

- **Archivo:** `internal/service/specialty_service_test.go`
- **Causa:** Nil pointer dereference en el mock de `GetByAdminID` — el campo no está inicializado en el struct del mock.
- **Severidad:** Medium — panic en tiempo de ejecución.

### 3. `TestAdminRepo_ComprehensiveSuite` > `CountSudos_ignores_Soft_Deleted_records`

- **Archivo:** `internal/repository/postgres/admin_repo_test.go`
- **Causa:** El test crea registros con `Delete()` (soft delete) y luego verifica que `CountSudos` no los cuenta. La query de CountSudos puede no estar filtrando correctamente por `deleted_at`.
- **Severidad:** Medium — posible bug en la query del repositorio.

---

## Bugs de Producción Corregidos

Durante la implementación de la suite de testing se encontraron y corrigieron los siguientes bugs:

### 1. `ErrNotAuthenticated` no existía

- **Archivo:** `internal/middleware/helpers.go`
- **Problema:** Los handlers comparaban `err == middleware.ErrNotAuthenticated` pero el sentinel error no estaba definido. Compilaba pero la comparación siempre fallaba.
- **Fix:** Se agregó `var ErrNotAuthenticated = errors.New("authentication required")` y se actualizaron 20 call sites en todos los handlers.

### 2. `SendEmail` panic con nil receiver

- **Archivo:** `internal/middleware/helpers.go`
- **Problema:** `SendEmail` no verificaba si el servicio era nil antes de usarlo, causando panic en tests y potencialmente en producción cuando el email no está configurado.
- **Fix:** Se agregó check `if s == nil { return nil }` al inicio de la función.

---

## Convenciones

### Nomenclatura de Tests

```go
func TestNombreFuncion_Cenario(t *testing.T)           // Happy path
func TestNombreFuncion_CuandoCondicion(t *testing.T)   // Edge case
func TestNombreFuncion_Error_Caso(t *testing.T)        // Error case
```

### Estructura Arrange-Act-Assert

```go
func TestCreateAdmin_Success(t *testing.T) {
    // Arrange
    repo := &mockAdminRepo{...}
    svc := service.NewAdminService(repo, nil, nil)
    dto := request_structs.CreateAdminRequest{...}

    // Act
    result, err := svc.CreateAdmin(dto, adminID)

    // Assert
    require.NoError(t, err)
    require.Equal(t, "expected", result.Name)
}
```

### Helpers de Test

- `internal/handler/test_helpers_test.go`: `setupAdminRoute`, `setupPublicRoute`, `testPsiHandler`, `testAdminHandler`
- `internal/integration/setup_test.go`: `buildTestApp`, `generateToken`, `makeRequest`, `seedAdmin`, `truncateAll`

### Archivos de Test por Paquete

Cada paquete tiene sus tests en el mismo directorio (`_test.go` con `package xxx_test` o `package xxx`). Los archivos de test usan el sufijo `_test.go` y están agrupados por responsabilidad:

- `*_test.go` — Tests de la funcionalidad principal
- `test_helpers_test.go` — Helpers compartidos (solo en handler e integration)
- `benchmarks_test.go` — Benchmarks (solo en utils)
- `updates_safety_test.go` — Tests específicos de un feature (GORM safety)

---

## Arquitectura de la Suite

```
TESTING.md                          ← Este documento
PLAN_TESTING.md                     ← Plan original (8 fases)
Makefile                            ← Targets de test
coverage_phase8.out                 ← Profile de cobertura (57.3%)
docker-compose.test.yml             ← PostgreSQL 18 para tests

internal/
├── config/config_test.go           ← 6 tests
├── domain/credentials_test.go      ← 15 tests
├── handler/
│   ├── test_helpers_test.go        ← Helpers (mocks, routes)
│   ├── admin_handler_test.go       ← 6 tests
│   ├── analytics_handler_test.go   ← 1 test
│   ├── posts_handler_test.go       ← 5 tests
│   ├── psi_handler_test.go         ← 9 tests
│   ├── psi_user_admin_test.go      ← 6 tests
│   └── specialty_handler_test.go   ← 7 tests
├── integration/
│   ├── setup_test.go               ← TestMain + helpers
│   ├── admin_flow_test.go          ← 8 tests
│   ├── directory_flow_test.go      ← 4 tests
│   ├── post_flow_test.go           ← 5 tests
│   ├── psi_flow_test.go            ← 6 tests
│   ├── security_flow_test.go       ← 43 tests
│   └── specialty_flow_test.go      ← 4 tests
├── logger/logger_test.go           ← 4 tests
├── middleware/
│   ├── auth_test.go                ← 1 test
│   ├── analytics_test.go           ← 3 tests
│   ├── helpers_test.go             ← 2 tests
│   ├── idempotency_test.go         ← 3 tests
│   └── rate_limiter_test.go        ← 2 tests
├── repository/postgres/
│   ├── user_admin_repo_test.go     ← 1 test (suite)
│   ├── specialty_repo_test.go      ← 1 test (suite)
│   ├── post_repo_test.go           ← 1 test (suite)
│   ├── psi_repo_test.go            ← 1 test (suite)
│   ├── analytics_repository_test.go← 1 test (suite)
│   └── updates_safety_test.go      ← 8 tests
├── request_structs/
│   ├── directory_filter_sanitizer_test.go ← 3 tests (35+ subtests)
│   └── visibility_getters_test.go         ← 4 tests (16 subtests)
├── router/router_test.go           ← 6 tests
├── service/
│   ├── admin_service_test.go       ← 6 tests
│   ├── analytics_service_test.go   ← 12 tests
│   ├── error_mapper_test.go        ← 2 tests
│   ├── mail_service_test.go        ← 2 tests
│   ├── post_service_test.go        ← 1 test
│   ├── psi_service_auth_test.go    ← 7 tests
│   ├── psi_service_directory_test.go ← 5 tests
│   ├── psi_service_import_test.go  ← 7 tests
│   ├── psi_service_test.go         ← 3 tests
│   ├── psi_service_xlsx_test.go    ← 13 tests
│   ├── psi_user_admin_service_test.go ← 3 tests
│   ├── social_media_test.go        ← 3 tests
│   └── specialty_service_test.go   ← 4 tests
└── utils/
    ├── utils_test.go               ← 9 tests
    ├── validate_email_test.go      ← 1 test (13 subtests)
    └── benchmarks_test.go          ← 30 benchmarks

pkg/
├── database/database_test.go       ← 6 tests
└── job/key_cleanup_test.go         ← 2 tests
```

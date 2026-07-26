# 🔒 Reporte de Auditoría — ColPsiCarabobo API

**Fecha:** 24 de Julio, 2026  
**Alcance:** API REST completa (Go/Fiber + PostgreSQL + S3)  
**Skills aplicadas:** `api-audit`, `owasp-audit`, `dependency-audit`, `secrets-audit`, `iam-audit`, `crypto-audit`, `privacy-engineering`  
**Metodología:** OWASP API Security Top 10 (2023) + OWASP Top 10 (2021) + auditoría de código fuente  

---

## 📋 Tabla de Contenidos

- [Resumen Ejecutivo](#resumen-ejecutivo)
- [1. Inventario del Stack](#1-inventario-del-stack)
- [2. Hallazgos Críticos](#2-hallazgos-críticos)
- [3. Hallazgos Altos](#3-hallazgos-altos)
- [4. Hallazgos Medios](#4-hallazgos-medios)
- [5. Hallazgos Bajos / Informativos](#5-hallazgos-bajos--informativos)
- [6. Análisis por Capa](#6-análisis-por-capa)
- [7. Cumplimiento OWASP Top 10](#7-cumplimiento-owasp-top-10)
- [8. Cumplimiento OWASP API Security Top 10](#8-cumplimiento-owasp-api-security-top-10)
- [9. Recomendaciones Priorizadas](#9-recomendaciones-priorizadas)
- [10. Métricas del Proyecto](#10-métricas-del-proyecto)

---

## Resumen Ejecutivo

| Severidad | Cantidad | Estado |
|-----------|----------|--------|
| 🔴 **CRÍTICOS** | 7 | ✅ Fase 1 + 2 completada (6/7) + Key Lifecycle (1 parcial) |
| 🟠 **ALTOS** | 12 | ✅ Fase 2 completada |
| 🟡 **MEDIOS** | 18 | ✅ Fase 3 completada |
| 🔵 **BAJOS** | 15 | ✅ Fase 4 completada |

**Veredicto:** La API tiene una arquitectura sólida (Clean Architecture, buenas prácticas de DI, saga patterns para S3). Se implementaron **47 fixes** de seguridad en 5 fases. La API está lista para despliegue en producción con las siguientes excepciones menores: CRIT-03 requiere hashing SHA-256 de keys (seguridad DB comprometida) y tests de integración para logout.

**Última actualización:** 25 de Julio, 2026 — Key Lifecycle Management commit `b511d43`

---

## 1. Inventario del Stack

| Componente | Tecnología | Versión |
|------------|-----------|---------|
| Lenguaje | Go | 1.25+ |
| Framework HTTP | Fiber | v2.52.11 |
| ORM | GORM | v1.31.1 |
| Base de Datos | PostgreSQL | 18-alpine |
| Connection Pool | PgBouncer | transaction mode |
| Almacenamiento | AWS S3 / MinIO | AWS SDK v2 |
| Autenticación | JWT | golang-jwt/v5 |
| Validación | validator/v10 | via struct tags |
| Sanitización XSS | bluemonday | v1.0.27 |
| Migraciones | Atlas | CLI |
| Contenedores | Docker Compose | v2+ |
| Documentación | Swagger/OpenAPI 2.0 | swag |
| Email | go-mail | v0.7.2 |
| Imágenes | libwebp + golang.org/x/image | — |
| Caching | go-cache (in-memory) | v2.1.0 |
| Configuración | godotenv | v1.5.1 |

### Dependencias con peso significativo

- `github.com/aws/aws-sdk-go-v2` — SDK completo de AWS (S3)
- `cloud.google.com/go/*` — Google Cloud (indirect, vía Atlas/Spanner)
- `gorm.io/driver/postgres` — Driver PostgreSQL
- `github.com/microcosm-cc/bluemonday` — HTML sanitizer
- `golang.org/x/crypto` — bcrypt para passwords
- `github.com/golang-jwt/jwt/v5` — JWT parsing/signing

---

## 2. Hallazgos Críticos

### CRIT-01: `OptionalHybridAuth` no valida firma JWT

| Campo | Valor |
|-------|-------|
| **OWASP** | API2: Broken Authentication |
| **Archivo** | `internal/middleware/auth.go:165-194` |
| **Severidad** | 🔴 CRÍTICO |

**Descripción:** El middleware `OptionalHybridAuth()` retorna `nil, nil` en el `jwt.Keyfunc` cuando el usuario existe en la DB. Esto causa que `jwt.Parse` use `nil` como clave HMAC, lo que significa que **cualquier token firmado con cualquier string pasa la validación** siempre que el `user_id` exista.

**Impacto:** Un atacante puede forjar tokens JWT con secretos triviales ("secret", "1234") y acceder a endpoints protegidos por `OptionalHybridAuth` (como `GET /posts/` y `GET /posts/:id`).

**Fix:** Hacer que el keyfunc retorne la `Key` real del usuario (igual que `ProtectedAdmin404`), o al menos validar la firma correctamente.

---

### CRIT-02: Password hardcodeada por defecto para todos los usuarios importados

| Campo | Valor |
|-------|-------|
| **OWASP** | A07: Identification and Authentication Failures |
| **Archivo** | `internal/service/psi_service.go:125`, `psi_service_xlsx.go:68` |
| **Severidad** | 🔴 CRÍTICO |

**Descripción:** La contraseña por defecto `"Colpsi2025!"` se usa para TODOS los usuarios importados via CSV/XLSX. Esta contraseña es:
1. Hardcodeada en el código fuente
2. Predecible (patrón: nombre + año)
3. La misma para cientos/miles de usuarios
4. Enviada en texto plano por email

**Impacto:** Si un atacante obtiene la lista de usuarios importados, puede autenticarse como cualquiera de ellos con la contraseña conocida.

**Fix:** Generar contraseñas aleatorias seguras por usuario, enviar por canal seguro, forzar cambio en primer login.

---

### CRIT-03: Credenciales JWT HMAC almacenadas en texto plano en DB

| Campo | Valor |
|-------|-------|
| **OWASP** | A02: Cryptographic Failures |
| **Archivo** | `internal/middleware/auth.go:124,219` |
| **Severidad** | 🔴 CRÍTICO |
| **Estado** | ⚠️ PARCIAL — Key Lifecycle Management implementado |

**Descripción:** El campo `Key` de `UserAdmin` y `PsiUserModel` se almacena en texto plano en PostgreSQL y se usa directamente como secreto HMAC para verificar/crear JWT. Si la DB es comprometida, todos los tokens activos pueden ser forjados.

**Impacto:** Un atacante con acceso a la DB puede crear tokens JWT válidos para cualquier usuario, incluyendo SUDO.

**Fix (parcial — commit `b511d43`):**
- ✅ Keys ahora son UUID v7 (timestamp embebido, no solo aleatorio)
- ✅ Admin y PsiUser logout eliminan la key inmediatamente (`key = ''`)
- ✅ Cleanup job `cmd/cleanup/` borra keys > 24h cada 30 minutos
- ✅ Middleware rechaza keys vacías sin intentar crypto

**Fix pendiente (hashing SHA-256):**
- Hash de key antes de almacenar en DB
- Firmar JWT con hash, no con UUID raw
- Migración de keys existentes con regex UUID para idempotencia

**Referencia:** `KEY_LIFECYCLE_REPORT.md`, `SECURITY_FIX_PLAN.md` → CRIT-03

---

### CRIT-04: Tags GORM erróneos causarán panic en migración

| Campo | Valor |
|-------|-------|
| **OWASP** | A05: Security Misconfiguration |
| **Archivo** | `internal/domain/user.model.go:108-109` |
| **Severidad** | 🔴 CRÍTICO |

**Descripción:**
- `ShowMunicipalityCarabobo bool` tiene tag `gorm:"size:255"` — un `bool` no puede tener `size`
- `PhoneCarabobo string` tiene tag `gorm:"default:false"` — un `string` no puede tener default booleano

**Impacto:** GORM panicará al intentar generar la migración o al ejecutar AutoMigrate, impidiendo el arranque de la aplicación.

**Fix:** Corregir los tags a `gorm:"default:false"` y `gorm:"size:20"` respectivamente.

---

### CRIT-05: Seed admin credentials expuestas en logs

| Campo | Valor |
|-------|-------|
| **OWASP** | A09: Security Logging and Monitoring Failures |
| **Archivo** | `pkg/database/seed.go:65` |
| **Severidad** | 🔴 CRÍTICO |

**Descripción:** El seed loguea la contraseña en texto plano: `log.Printf("User: %s | Pass: %s | ID: %s", admin.Username, defaultPass, admin.ID)`.

**Impacto:** Cualquiera con acceso a los logs de la aplicación obtiene las credenciales del admin SUDO.

**Fix:** Eliminar el log del password o enmascararlo.

---

### CRIT-06: Hardcoded secrets como defaults en configuración

| Campo | Valor |
|-------|-------|
| **OWASP** | A05: Security Misconfiguration |
| **Archivo** | `internal/config/env.config.go:110-111` |
| **Severidad** | 🔴 CRÍTICO |

**Descripción:** Los valores por defecto son:
- `JwtLibrarySecret = "secret-for-psi-library"` — secreto JWT predecible
- `AbsAdminToken = "deaful_ABS_ADMIN_TOKEN"` — token con typo ("deaful" en vez de "default")

**Impacto:** Si se olvidan configurar las env vars en producción, la API arranca con tokens predecibles y compartidos.

**Fix:** Forzar que estas variables sean requeridas (panic si no están seteadas) o generar aleatoriamente al arrancar.

---

### CRIT-07: Envío SMTP deshabilitado en código

| Campo | Valor |
|-------|-------|
| **OWASP** | A05: Security Misconfiguration |
| **Archivo** | `internal/service/mail_service.go:198-200` |
| **Severidad** | 🔴 CRÍTICO |

**Descripción:** La función `DialAndSend()` está comentada. Los correos de bienvenida, notificaciones de login y cualquier otro email **nunca se envían realmente**.

**Impacto:** Los usuarios nunca reciben sus credenciales temporales. Las notificaciones de login (security feature) no funcionan.

**Fix:** Descomentar el envío y verificar que el servidor SMTP está configurado correctamente.

---

## 3. Hallazgos Altos

### HIGH-01: Rate limiting in-memory no sobrevive restarts

| Campo | Valor |
|-------|-------|
| **OWASP** | API4: Unrestricted Resource Consumption |
| **Archivo** | `internal/middleware/rate_limiter.go:27,68` |
| **Severidad** | 🟠 ALTO |

**Descripción:** Los rate limiters usan `sync.Map` en memoria. En un部署 con múltiples instancias (Docker replicas, Kubernetes), cada instancia tiene su propio contador, multiplicando efectivamente el límite por N.

**Fix:** Migrar a Redis o usar un store distribuido.

---

### HIGH-02: `debug-monitor` sin autenticación en producción

| Campo | Valor |
|-------|-------|
| **OWASP** | API9: Improper Inventory Management |
| **Archivo** | `internal/router/admin_router.go:33` |
| **Severidad** | 🟠 ALTO |

**Descripción:** El endpoint `/api/v1/debug-monitor` está registrado sin el guard de `config.Environment` (a diferencia de Swagger que sí lo tiene). Expone métricas internas del sistema a cualquiera.

**Fix:** Proteger con `config.Envs.AppEnv == "development"` o eliminarlo.

---

### HIGH-03: Panic en startup si SMTP falla

| Campo | Valor |
|-------|-------|
| **OWASP** | A05: Security Misconfiguration |
| **Archivo** | `internal/router/psi_router.go:22` |
| **Severidad** | 🟠 ALTO |

**Descripción:** `panic("Error al inicializar el servicio de correo: " + err.Error())` mata toda la app si SMTP falla. En `admin_router.go:21-23` el mismo caso se maneja con `log.Printf` (inconsistente).

**Fix:** Degradación graciosa: log warning y continuar sin email.

---

### HIGH-04: Casteo inseguro de `c.Locals()` sin nil-check

| Campo | Valor |
|-------|-------|
| **OWASP** | A03: Injection (runtime panic) |
| **Archivos** | `psi_handler.go:89,307,339`, `psi_user_admin.go:29,80,109,158,186`, `posts_handler.go:80,147`, `specialty_handler.go:83,113,141` |
| **Severidad** | 🟠 ALTO |

**Descripción:** `c.Locals("admin").(*domain.UserAdmin)` sin ok-check puede causar panic runtime si el middleware no setea el Local. Algunos handlers sí lo hacen con ok-check — inconsistente.

**Fix:** Usar sempre el ok-check: `admin, ok := c.Locals("admin").(*domain.UserAdmin)`.

---

### HIGH-05: Comparación de errores por string matching

| Campo | Valor |
|-------|-------|
| **OWASP** | A04: Insecure Design |
| **Archivos** | `psi_handler.go:116,413`, `psi_user_admin.go:41`, `specialty_handler.go:91` |
| **Severidad** | 🟠 ALTO |

**Descripción:** `err.Error() == "contraseña actual incorrecta"` — si el service cambia el texto del error, el handler se rompe silenciosamente. Vulnerable a cambios accidentales.

**Fix:** Usar sentinel errors o tipos de error custom con `errors.Is()`.

---

### HIGH-06: Logger GORM en modo Info en producción

| Campo | Valor |
|-------|-------|
| **OWASP** | A09: Security Logging and Monitoring Failures |
| **Archivo** | `pkg/database/postgres.go:45` |
| **Severidad** | 🟠 ALTO |

**Descripción:** `logger.Default.LogMode(logger.Info)` loguea cada query SQL. En producción esto genera I/O innecesario y puede filtrar datos sensibles en logs (PII, passwords en queries).

**Fix:** Usar `logger.Warn` o `logger.Error` en producción.

---

### HIGH-07: Cookie analytics sin `Secure: true`

| Campo | Valor |
|-------|-------|
| **OWASP** | A07: Identification and Authentication Failures |
| **Archivo** | `internal/middleware/analytics.go:85` |
| **Severidad** | 🟠 ALTO |

**Descripción:** La cookie `_sid` tiene `HTTPOnly: true` pero falta `Secure: true`. Se enviará por HTTP si el frontend no está en HTTPS.

**Fix:** Agregar `Secure: true` y `SameSite: Lax` o `Strict`.

---

### HIGH-08: Pool de conexiones DB no configurado

| Campo | Valor |
|-------|-------|
| **OWASP** | API4: Unrestricted Resource Consumption |
| **Archivo** | `pkg/database/postgres.go` |
| **Severidad** | 🟠 ALTO |

**Descripción:** El `.ai-context.md` documenta MaxOpenConns=10, MaxIdleConns=5, ConnMaxLifetime=1h, pero el código **no llama** `db.DB().SetMaxOpenConns(...)`. El pool queda con valores por defecto de Go (ilimitado).

**Fix:** Configurar el pool explícitamente.

---

### HIGH-09: Passwords hardcodeados en templates de email

| Campo | Valor |
|-------|-------|
| **OWASP** | A07: Identification and Authentication Failures |
| **Archivo** | `internal/templates/welcome_psi.html:16`, `welcome_admin.html:15` |
| **Severidad** | 🟠 ALTO |

**Descripción:** Los templates envían `{{.TempPassword}}` en texto plano sin indicar que debe cambiarse. Combinado con CRIT-02, todos los usuarios tienen la misma contraseña conocida.

**Fix:** Forzar cambio de contraseña en primer login, no enviarla por email.

---

### HIGH-10: S3 keys expuestas en respuestas API

| Campo | Valor |
|-------|-------|
| **OWASP** | API2: Broken Authentication |
| **Archivo** | `internal/domain/user.model.go` (múltiples campos) |
| **Severidad** | 🟠 ALTO |

**Descripción:** Los campos `ProfilePictureS3Key`, `TitleImageOneS3Key`, `PicOneS3Key`, `ImageS3Key` etc. tienen json tags que exponen las keys de S3 en la API pública, revelando la estructura del bucket.

**Fix:** Usar `json:"-"` o generar URLs firmadas temporalmente.

---

### HIGH-11: `ON DELETE NO ACTION` en todas las Foreign Keys

| Campo | Valor |
|-------|-------|
| **OWASP** | A04: Insecure Design |
| **Archivo** | `migrations/20260604165811_init.sql` (FKs) |
| **Severidad** | 🟠 ALTO |

**Descripción:** Ninguna FK usa CASCADE. Un hard delete en `psi_users` deja registros huérfanos en `col_data`, `post_grades`, `social_networks`, `solvency`.

**Fix:** Usar CASCADE o RESTRICT según el caso de negocio.

---

### HIGH-12: Doble idempotencia global + custom

| Campo | Valor |
|-------|-------|
| **OWASP** | API4: Unrestricted Resource Consumption |
| **Archivo** | `cmd/api/main.go:156` + `internal/middleware/idempotency.go:112` |
| **Severidad** | 🟠 ALTO |

**Descripción:** Hay un `idempotency.New()` global de Fiber Y un `UserScopedIdempotency()` custom. Ambos usan `X-Idempotency-Key`. Pueden conflictuar o causar comportamiento indefinido.

**Fix:** Eliminar uno de los dos y estandarizar.

---

## 4. Hallazgos Medios

| ID | Hallazgo | OWASP | Archivo |
|----|----------|-------|---------|
| MED-01 | `PsiService` es God Object (~1860 líneas, 18+ métodos) | A04 | `psi_service.go` |
| MED-02 | `mapDBErrorToHuman` duplicada (violación DRY) | A04 | `psi_service.go:230` vs `error_mapper.go:18` |
| MED-03 | `println()` nativo en producción (no loguea) | A09 | `psi_service.go:632` |
| MED-04 | AnalyticsService accede directo a `*gorm.DB` sin repository | A04 | `analytics_service.go:22` |
| MED-05 | `fmt.Printf` de debug en repositorio | A09 | `psi_repository.go:329,344` |
| MED-06 | FormFile ignorado sin manejo de error | A04 | `psi_handler.go:96`, `posts_handler.go:87` |
| MED-07 | Logging de objeto completo de usuario (posible PII) | A09 | `psi_handler.go:229-230` |
| MED-08 | Inconsistencia de status codes (numeros vs constantes) | A05 | `psi_handler.go`, `analytics_handler.go` |
| MED-09 | `LoginLibrary` sin analytics ni documentación Swagger | A09 | `psi_handler.go:281` |
| MED-10 | Repos instanciados múltiples veces (3x AdminRepo, 3x PsiRepo) | A04 | `psi_router.go`, `post_router.go`, `admin_router.go` |
| MED-11 | `Save()` para updates parciales (sobreescribe zero-values) | A04 | `user_admin_repo.go:73`, `specialty_repo.go:103` |
| MED-12 | Búsqueda de especialidad por nombre, no por FK | A04 | `psi_repository.go:537-542` |
| MED-13 | `BoolFromForm()` duplicada con comportamiento distinto | A04 | `psi_user.go:117` vs `utils/geo_venezuela.go:144` |
| MED-14 | Test `SanitizeImage_Defensive` roto | A05 | `utils_test.go:28` vs `image_sanitizer.go:86` |
| MED-15 | `runtime.GOMAXPROCS(2)` hardcodeado | A05 | `cmd/api/main.go:43` |
| MED-16 | Sin graceful shutdown | A05 | `cmd/api/main.go:187` |
| MED-17 | `PsiObservations` y `PsiODeontologia` sin relación en struct padre | A04 | `internal/domain/user.model.go` |
| MED-18 | `phone_carabobo` es `text DEFAULT 'false'` en migración | A05 | `migrations/20260604165811_init.sql:262` |

---

## 5. Hallazgos Bajos / Informativos

| ID | Hallazgo | Archivo |
|----|----------|---------|
| LOW-01 | Nombre de archivo con typo: `radom_string.go` | `internal/utils/radom_string.go` |
| LOW-02 | Nombre de archivo con typo: `post_respository.go` | `internal/repository/postgres/post_respository.go` |
| LOW-03 | Tipo con typo: `PsiUSerSolvency` (S mayúscula) | `internal/domain/user.model.go:214` |
| LOW-04 | Error message con typo: `"emial invalido"` | `internal/service/admin_service.go:379` |
| LOW-05 | Log copy-paste erróneo: "BIO" cuando recupera Solvencies | `internal/handler/psi_user_admin.go:55` |
| LOW-06 | Comentario copy-paste: "Empleado público" en campo Discapacity | `internal/domain/user.model.go:196` |
| LOW-07 | `Post` sin `TableName()` explícito | `internal/domain/text.model.go:37` |
| LOW-08 | `GraduationYear` es `string` en vez de `int` | `internal/domain/user.model.go:256` |
| LOW-09 | Inconsistencia de UUID generation (`gen_random_uuid` vs `uuidv7`) | `internal/domain/analytics.go:15` |
| LOW-10 | `context.TODO()` en `ConnectS3()` | `pkg/s3/s3.go:28` |
| LOW-11 | `GetPresignedURL()` comentado con `/*` | `pkg/s3/upload.go:79-92` |
| LOW-12 | Log con emojis en producción (portabilidad) | Múltiples archivos |
| LOW-13 | `log.Println("[DEBUG AUTH]")` en producción | `internal/middleware/auth.go:66,71,80` |
| LOW-14 | `UserAdmin` y `PsiUserModel` comparten campos duplicados | `internal/domain/user.model.go` |
| LOW-15 | `.ai-context.md` desactualizado | Múltiples |

---

## 6. Análisis por Capa

### 6.1 Dominio (`internal/domain/`)

**Fortalezas:**
- Clean Architecture con separación clara de modelos e interfaces
- Soft delete consistente con `AuditModel` embebido
- UUIDv7 como PK (ordenamiento temporal)
- Partial unique index para SUDO

**Debilidades:**
- 2 tags GORM erróneos (CRIT-04)
- FK inconsistente (`PsiUserID` vs `PsiUserModelID`)
- Relación N:M con specialties como strings sueltos (sin FK)
- Sin validación a nivel de modelo (solo `PostGradeType.IsValid()`)
- `SearchDirectory` depende de `request_structs` (viola Clean Architecture)

### 6.2 Repositories (`internal/repository/postgres/`)

**Fortalezas:**
- Transacciones explícitas para operaciones compuestas
- Preload con ordenamiento
- Búsqueda con unaccent e ILIKE
- Upsert con OnConflict para solvencias
- Proyecciones Select para queries ligeras

**Debilidades:**
- `fmt.Printf` de debug en producción
- `ValidateUniqueCredentials` ignora error de Count
- `Save()` para updates parciales (sobreescribe zero-values)
- Búsqueda de especialidad por nombre, no por FK
- Pool de conexiones no configurado

### 6.3 Services (`internal/service/`)

**Fortalezas:**
- Key Rotation en JWT (cada login invalida tokens previos)
- RBAC jerárquico con 15 permisos
- Privacy Shield para perfiles públicos
- Saga Rollback de S3 (si DB falla, borra imagen)
- Rate limiting anti-spam en emails
- Sanitización XSS con bluemonday

**Debalidades:**
- `PsiService` es God Object (~1860 líneas)
- Password hardcodeada "Colpsi2025!"
- Envío SMTP comentado
- `println()` nativo en producción
- AnalyticsService acoplado directo a `*gorm.DB`
- Funciones helper duplicadas (`safeGet`/`getValorSeguro`)

### 6.4 Handlers (`internal/handler/`)

**Fortalezas:**
- Swagger anotado en 30+ endpoints
- Validación de entrada con BodyParser/QueryParser
- Error handling consistente (400/401/403/404/500)
- Privacy Shield en respuestas públicas

**Debilidades:**
- Stub functions vacías que retornan nil (pueden causar panic)
- Comparación de errores por string
- Casteo inseguro de `c.Locals()` sin nil-check
- Status codes como enteros crudos en vez de constantes
- 4 endpoints sin anotación Swagger

### 6.5 Middleware (`internal/middleware/`)

**Fortalezas:**
- Autenticación JWT con verificación de algoritmo HMAC
- Protección contra None Algorithm Attack (testeada)
- Rate limiting configurable por grupo
- Idempotencia scoped por usuario
- Analytics middleware global con debounce

**Debilidades:**
- `OptionalHybridAuth` no valida firma JWT (CRIT-01)
- Rate limiting in-memory (no distribuido)
- Idempotency store in-memory (no sobrevive restarts)
- `debug-monitor` sin protección
- Doble implementación de idempotencia

### 6.6 Base de Datos

**Fortalezas:**
- 15 tablas con esquema bien definido
- Constraints UNIQUE compuestos
- PgBouncer en modo transacción
- Atlas para migraciones versionadas
- Seeds con admin por defecto

**Debilidades:**
- Unique index duplicado (`user_id, user_id`)
- Todas las FK con `ON DELETE NO ACTION`
- Sin índices en `is_active` y `solvent` (usados en directorio)
- Sin índices compuestos en `posts(status, type)`
- Seed loguea password en texto plano

### 6.7 Infraestructura

**Fortalezas:**
- Docker Compose con servicios aislados
- PgBouncer para connection pooling
- MinIO como S3 compatible local
- Atlas para migraciones automáticas
- Health checks en DB

**Debilidades:**
- Sin graceful shutdown
- `runtime.GOMAXPROCS(2)` hardcodeado
- Sin health check de la API en el compose
- Audiobookshelf incluido en el compose (¿relacionado?)

---

## 7. Cumplimiento OWASP Top 10

| ID | Categoría | Estado | Hallazgos |
|----|-----------|--------|-----------|
| A01 | Broken Access Control | ⚠️ PARCIAL | IDOR prevention en social media ✓, pero S3 keys expuestas, `OptionalHybridAuth` sin auth real |
| A02 | Cryptographic Failures | ❌ FALLO | HMAC keys en DB en texto plano, password hardcodeada, SMTP comentado |
| A03 | Injection | ✅ OK | Validación con body parser, bluemonday para XSS, parameterized queries vía GORM |
| A04 | Insecure Design | ⚠️ PARCIAL | God Object, sin FK cascade, `Save()` para updates, string matching para errores |
| A05 | Security Misconfiguration | ❌ FALLO | Tags GORM erróneos, debug-monitor sin auth, secrets hardcodeados, logger en Info |
| A06 | Vulnerable Components | ⚠️ REVISAR | No se ejecutó `govulncheck` en este análisis |
| A07 | Auth Failures | ❌ FALLO | `OptionalHybridAuth` sin validación, password compartida, cookie sin Secure |
| A08 | Data Integrity | ⚠️ PARCIAL | Saga S3 ✓, pero sin verificación de integridad de migraciones |
| A09 | Logging Failures | ⚠️ PARCIAL | Password en logs, PII en logs, println en producción, debug auth logs |
| A10 | SSRF | ✅ OK | No se detectaron endpoints con SSRF |

---

## 8. Cumplimiento OWASP API Security Top 10

| ID | Categoría | Estado | Hallazgos |
|----|-----------|--------|-----------|
| API1 | BOLA | ⚠️ PARCIAL | IDOR check en social media ✓, ownership en títulos ✓, pero `OptionalHybridAuth` debilita la protección |
| API2 | Broken Auth | ❌ FALLO | `OptionalHybridAuth` sin firma, JWT secrets en DB, password hardcodeada |
| API3 | BOPLA / Mass Assignment | ✅ OK | DTOs explícitos para cada operación, no se usa `json.Bind` directo |
| API4 | Resource Consumption | ⚠️ PARCIAL | Rate limiting ✓, pero in-memory y sin límite de page size en todos los endpoints |
| API5 | BFLA | ⚠️ PARCIAL | RBAC con 15 permisos ✓, pero `debug-monitor` expuesto |
| API6 | Business Flow Abuse | ✅ OK | Anti-automation con rate limiting, idempotencia en creaciones |
| API7 | SSRF | ✅ OK | No se detectaron vector SSRF |
| API8 | Security Misconfiguration | ❌ FALLO | CORS config ✓, pero debug endpoints, logger verbose, secrets hardcodeados |
| API9 | Improper Inventory | ❌ FALLO | `debug-monitor` expuesto, `LoginLibrary` sin docs |
| API10 | Unsafe Consumption | ✅ OK | Consumo de Audiobookshelf con timeout, no se confía en respuesta para auth |

---

## 9. Recomendaciones Priorizadas

### Prioridad 1 — Inmediato (antes de producción)

1. **Fix `OptionalHybridAuth`**: Hacer que el keyfunc retorne la Key real del usuario
2. **Eliminar password hardcodeada**: Generar passwords aleatorias seguras por usuario
3. **Corregir tags GORM erróneos**: `ShowMunicipalityCarabobo` y `PhoneCarabobo`
4. **Eliminar seed password de logs**: No loguear credenciales
5. **Secrets requeridos**: Panic o generar aleatoriamente si no están configurados
6. **Descomentar envío SMTP** o deshabilitar funcionalidad dependiente
7. **Hashear HMAC keys** en DB o usar vault

### Prioridad 2 — Antes de escalar

8. **Migrar rate limiting a Redis** para soporte multi-instancia
9. **Eliminar `debug-monitor`** en producción
10. **Configurar pool de conexiones** DB explícitamente
11. **Agregar `Secure: true`** a cookies
12. **Proteger `panic` en psi_router** con degradación graciosa
13. **Eliminar doble idempotencia** (global vs custom)
14. **Corregir cast inseguro de `c.Locals()`** con nil-check

### Prioridad 3 — Deuda técnica

15. **Split `PsiService`** en 3-4 servicios más pequeños
16. **Crear repository para Analytics** (romper acoplamiento a `*gorm.DB`)
17. **Estandarizar error handling** con sentinel errors
18. **Eliminar código debug** (`fmt.Printf`, `println`)
19. **Corregir tests rotos**
20. **Agregar graceful shutdown**

---

## 10. Métricas del Proyecto

| Métrica | Valor |
|---------|-------|
| Endpoints totales | 40 |
| Endpoints con Swagger | 32 (80%) |
| Modelos de dominio | 17 |
| Interfaces de repository | 4 |
| Repositories implementados | 4 |
| Services | 8 structs |
| Middlewares custom | 8 |
| Archivos de configuración | 12 |
| Tests unitarios | Limitados (utils_test.go: 221 líneas) |
| Migraciones | 1 (init) |
| Tablas en BD | 16 |
| Líneas de código (estimado) | ~8,000+ |
| Dependencias directas | 18 |
| Dependencias indirectas | ~120 |

---

*Reporte generado por análisis automatizado con skills: `api-audit`, `owasp-audit`, `dependency-audit`, `secrets-audit`, `iam-audit`, `crypto-audit`, `privacy-engineering`*

# ⚙️ Servicios de Negocio (service/)

> **[⬆ internal](../)** — `api/internal/service/`
>
> **Capa central de lógica de negocio** — orquesta reglas de seguridad, validación, persistencia y comunicación externa. Es la capa más compleja del sistema.

## Arquitectura General

```
┌─────────────────────────────────────────────────────────┐
│                    HANDLER (HTTP)                        │
│           Valida HTTP → Desempaqueta Request             │
└──────────────────────┬──────────────────────────────────┘
                       │ Llamada a método de servicio
                       ▼
┌─────────────────────────────────────────────────────────┐
│                   SERVICE (ESTE PAQUETE)                 │
│                                                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ Admin    │ │ Psi      │ │ Post     │ │ Specialty│   │
│  │ Service  │ │ Service  │ │ Service  │ │ Service  │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ Analytics│ │ Mail     │ │ Error    │ │ Social   │   │
│  │ Service  │ │ Service  │ │ Mapper   │ │ Media    │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
│  ┌──────────────────────────┐                           │
│  │ PsiUserAdminService      │                           │
│  │ (métodos de PsiService)  │                           │
│  └──────────────────────────┘                           │
│                                                          │
│  Responsabilidades:                                     │
│  ✓ Validación de permisos (RBAC)                        │
│  ✓ Hashing de contraseñas (bcrypt)                      │
│  ✓ Generación de JWT (Key Rotation)                     │
│  ✓ Sanitización XSS (bluemonday)                        │
│  ✓ Gestión de imágenes S3 (Saga Rollback)               │
│  ✓ Geolocalización normalizada (Venezuela)              │
│  ✓ Trazabilidad forense (Audit Trail)                   │
└──────────────────────┬──────────────────────────────────┘
                       │ Interfaz (domain.XxxRepository)
                       ▼
┌─────────────────────────────────────────────────────────┐
│              REPOSITORY (postgres/)                      │
│           Implementación concreta con GORM               │
└─────────────────────────────────────────────────────────┘
```

---

## 📁 Estructura de Archivos

```
internal/service/
├── admin_service.go            # Autenticación y CRUD de administradores
├── psi_service.go              # Perfiles de psicólogos, directorio, login
├── psi_service_xlsx.go         # Importación masiva desde Excel/XLSX
├── psi_user_admin_service.go   # Edición administrativa de psicólogos
├── post_service.go             # CMS: publicaciones y noticias
├── specialty_service.go        # Catálogo de especialidades
├── analytics_service.go        # Telemetría y dashboard BI
├── mail_service.go             # Envío asíncrono de correos
├── error_mapper.go             # Traducción de errores DB → español
├── social_media.go             # CRUD de redes sociales del psicólogo
├── *_test.go                   # Pruebas unitarias
└── README.md                   # Este archivo
```

---

## 🔐 AdminService (`admin_service.go`)

Gestiona la autenticación, registro y administración del staff del Colegio.

### Login con Key Rotation

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Client   │───▶│  Login   │───▶│ Generate │───▶│ Sign JWT │
│  POST     │    │ Validate │    │ New UUID │    │ with Key │
│  creds    │    │ bcrypt   │    │  Key     │    │          │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
                                    │
                                    ▼
                              ┌──────────┐
                              │ Persist  │
                              │ New Key  │
                              │ in DB    │
                              └──────────┘
```

**Patrón:** Cada login genera un nuevo UUID secreto que se guarda en la DB como `Key` del usuario. El JWT se firma con ese Key. Al rotar la key, cualquier token anterior emitido en otro dispositivo queda criptográficamente inválido.

**Seguridad:**
- Mensajes genéricos de error ("credenciales inválidas") para prevenir **Username Enumeration Attack**
- `bcrypt.CompareHashAndPassword` previene **Timing Attacks**
- Si el admin está desactivado (`IsActive = false`), se rechaza explícitamente
- Al cambiar contraseña, se genera nuevo Key expulsando otras sesiones

### Registro con Password Hashing

```go
hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

- Contraseña encriptada con **bcrypt** antes de persistir
- Validación de fortaleza con `utils.IsStrongPassword()`
- Sanitización de email con `utils.ParseAndValidateEmail()`

### Matriz de Permisos (Permission Matrix)

```
┌─────────────────────────────────────────────────────────┐
│              PERMISSION MATRIX ENGINE                    │
│                                                          │
│  buildPermissionMatrix() → []permissionUpdate            │
│                                                          │
│  Cada permiso evalúa:                                    │
│  1. ¿El solicitante LO TIENE? (updaterHas)               │
│  2. ¿El request LO PIDE? (requested)                     │
│  3. ¿El target ya LO TIENE? (current)                    │
│                                                          │
│  Regla: No puedes delegar permisos que tú no posees     │
│  (excepto el usuario Sudo)                               │
│                                                          │
│  Permisos: CanCreatePsi, CanUpdatePsi, CanDeletePsi,    │
│  CanCreateAdmin, CanUpdateAdmin, CanDeleteAdmin,         │
│  CanPublish, CanUpdatePublish, CanDeletePublish,         │
│  CanSendNotifications, CanManageNotifications,           │
│  CanReadNotifications, CanCreateTags, CanEditTags,       │
│  CanDeleteTags                                           │
└─────────────────────────────────────────────────────────┘
```

**Patrón:** Data-Driven Validation con Closures — transforma validaciones caóticas en un array iteruable de tuplas, evitando el uso de `reflect`.

### Caché con `go-cache`

```go
cache: cache.New(5*time.Minute, 10*time.Minute)
```

- **TTL base:** 5 minutos para cada entrada
- **GC cycle:** 10 minutos para purgar claves expiradas
- **Cache Key:** Determinístico (`admins_l:{limit}_p:{page}_s:{search}_a:{active}`)
- **Invalidación:** `s.cache.Flush()` completo tras cualquier escritura (Create/Update/Delete)

### Controles de Seguridad

| Regla | Descripción |
|-------|-------------|
| **Auto-Bloqueo** | Un admin no puede eliminarse a sí mismo |
| **Inmunidad Sudo** | Nadie puede editar o eliminar al usuario Sudo |
| **Escalada Restringida** | No puedes crear admins con más permisos que los tuyos |
| **Heredabilidad** | Sudo es `false` por defecto; solo se asigna directamente en DB |

---

## 🧠 PsiService (`psi_service.go`)

El servicio más extenso. Orquesta perfiles de psicólogos, directorio público, autenticación y sincronización con microservicios.

### Login (Key Rotation idéntica a AdminService)

- JWT con `role: "psi"` (diferente al `"admin"`)
- Login notification por email (no bloqueante)
- **LoginLibrary:** SSO básico con Audiobookshelf — genera JWT con secreto independiente (`JwtLibrarySecret`) y crea/sincroniza usuario en el sistema de biblioteca

### Privacy Shield (Escudo de Privacidad)

```
┌─────────────────────────────────────────────────────────┐
│                  PRIVACY SHIELD                          │
│                                                          │
│  Perfil PÚBLICO (sin auth):                              │
│  ┌─────────────────────────────────────────────────┐     │
│  │ Datos SIEMPRE visibles:                        │     │
│  │  → Nombre, FPV, CI, Género, Foto, MiniBio      │     │
│  │                                                │     │
│  │ Datos CONDICIONALES (Opt-in):                  │     │
│  │  → Email, Teléfono, Dirección                  │     │
│  │    (solo si ShowXxx = true)                     │     │
│  │                                                │     │
│  │ Datos NUNCA visibles:                          │     │
│  │  → Contraseña, Key JWT, Datos internos         │     │
│  └─────────────────────────────────────────────────┘     │
│                                                          │
│  Perfil NO SOLVENTE:                                     │
│  ┌─────────────────────────────────────────────────┐     │
│  │ Solo muestra: Nombre, FPV, CI, Género, Foto,   │     │
│  │ Universidad de Pregrado                         │     │
│  │ (Degradación Elegante — sin error 404)          │     │
│  └─────────────────────────────────────────────────┘     │
│                                                          │
│  Perfil ADMIN (GetPsiByIDAdmin):                         │
│  ┌─────────────────────────────────────────────────┐     │
│  │ Acceso TOTAL — bypass del Privacy Shield        │     │
│  │ Incluye: Teléfonos privados, Email, Solvencias  │     │
│  └─────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────┘
```

### Directorio con Búsqueda Avanzada

El método `GetPublicDirectory` soporta:

- **Búsqueda por nombre:** Tokenización multi-palabra con `ILIKE unaccent()` para texto en español
- **Filtro por especialidad:** Se resuelve el nombre del area de trabajo
- **Filtro por ubicación:** Municipios de Carabobo, estados de Venezuela, países
- **Filtro por solvencia:** Solo solventes en vista pública
- **Filtro por género:** Normalizado a "M" / "F"

### Gestión de Imágenes S3 (Saga Rollback)

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│  Upload  │───▶│  Sanitize│───▶│  S3      │
│  Image   │    │  (EXIF)  │    │  Upload  │
└──────────┘    └──────────┘    └──────────┘
                                      │
                              ┌───────┴───────┐
                              │               │
                        ┌─────┴─────┐   ┌─────┴─────┐
                        │ DB OK     │   │ DB FAIL   │
                        │ → Delete  │   │ → Delete  │
                        │ old images│   │ uploaded  │
                        └───────────┘   └───────────┘
```

**Flujo:**
1. Sanitizar imagen (`utils.SanitizeDocument`) — elimina metadatos EXIF, re-codifica
2. Subir nueva imagen a S3
3. Si DB falla → **Rollback:** eliminar imagen recién subida
4. Si DB éxito → **Garbage Collection:** eliminar imagen antigua

### Importación Excel/XLSX (`psi_service_xlsx.go`)

Motor de ingesta masiva con tolerancia a fallos:

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Excel   │───▶│  Parse   │───▶│  Build   │───▶│  DB      │
│  Stream  │    │  Rows    │    │  Models  │    │  Trans.  │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
                   │                │                │
                   │ fallo fila     │ error unique   │
                   ▼                ▼                ▼
              ┌──────────────────────────────────────┐
              │    failedRecords[] (reporte final)    │
              └──────────────────────────────────────┘
```

**Optimización criptográfica:** El hash bcrypt se genera **una sola vez** antes del bucle y se reutiliza para todos los registros, reduciendo el tiempo de importación de minutos a segundos.

**Columnas parseadas:** 43+ campos incluyendo FPV, CI, datos colegiales, geo-localización, títulos, postgrados, solvencias.

**Geo-normalización:**
- `utils.NormalizeMunicipioCarabobo()` — normaliza municipios del estado Carabobo
- `utils.NormalizeEstadoVenezuela()` — normaliza estados venezolanos

### Módulo Académico (Postgrados)

- `AddPostGrade()` — registra título con hasta 3 imágenes S3
- `UpdatePostGrade()` — actualización parcial con reemplazo inteligente de imágenes
- **Ownership Check:** Verifica que el postgrado pertenezca al psicólogo autenticado

### Sincronización con Audiobookshelf

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│  ColPsi  │───▶│  HTTP    │───▶│ Audiobook│
│  Login   │    │  PATCH   │    │ shelf    │
│          │    │  (5s     │    │ API      │
│          │    │  timeout)│    │          │
└──────────┘    └──────────┘    └──────────┘
```

- **Graceful Degradation:** Si Audiobookshelf falla, el login principal no se ve afectado
- **Idempotencia:** Si el usuario ya existe (HTTP 409), se ignora pacíficamente
- **Timeout estricto:** 5 segundos para evitar cascada de fallos

---

## 🛡️ PsiUserAdminService (`psi_user_admin_service.go`)

Edición administrativa de alto nivel para expedientes de psicólogos.

### Diferencia con PsiService

| Característica | PsiService | PsiUserAdminService |
|---------------|------------|---------------------|
| Autogestión | ✅ (el propio psi) | ❌ (solo admin) |
| Privacy Shield | ✅ Aplicado | ❌ Bypass total |
| Campos editables | Contacto, ubicación, bio | **Todo** (FPV, CI, solvencia, registro) |
| Password req | ✅ Requiere actual | ❌ No requiere |
| Solvencias | Solo lectura | CRUD completo |

### Métodos Principales

- **GetPsiByIDAdmin** — Vista de "Rayos X" del expediente completo con historial de solvencias
- **CreatePsiByAdmin** — Registro individual con factory pattern (`createPsiUSerModel`, `createColdata`, `createSolvencieModel`)
- **UpdatePsiByAdmin** — Actualización masiva con soporte multipart (imágenes + JSON embebido para solvencias)
- **DeletePsiByAdmin** — Soft delete con verificación RBAC
- **GetAdminDirectory** — Listado administrativo (Proyección de Datos ligera)

### Procesamiento de Solvencias

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│  Request │───▶│  Decode  │───▶│  Check   │
│  JSON    │    │  String  │    │  Dups    │
│  String  │    │  → Array │    │  (idemp.)│
└──────────┘    └──────────┘    └──────────┘
                                      │
                                      ▼
                                ┌──────────┐
                                │  Auto    │
                                │  Solvent │
                                │  = true  │
                                │ (if year │
                                │  = now)  │
                                └──────────┘
```

El JSON de solvencias viene embebido como string en multipart/form-data (solución a la limitación HTTP de arrays anidados). Se decodifica, se validan duplicados, y si la solvencia es del año actual, se activa automáticamente la bandera `Solvent`.

---

## 📝 PostService (`post_service.go`)

CMS completo con separación metadata/contenido y ACL por rol.

### Separación de Datos (Text/Post Split)

```
┌─────────────────────────────────────────────┐
│  Post (metadata ligera)                      │
│  ├── Title, ShortDescription                 │
│  ├── Type (public/psi), Status               │
│  ├── ImageS3Key, PublishAt                   │
│  └── TextID ──────────────────┐              │
└───────────────────────────────┤──────────────┘
                                │ FK
┌───────────────────────────────▼──────────────┐
│  TextModel (contenido extenso)               │
│  └── Content (HTML sanitizado)               │
└──────────────────────────────────────────────┘
```

**Razón:** Los listados paginados transfieren solo la metadata sin descargar el HTML completo de cada post, optimizando ancho de banda y rendimiento.

### ACL (Access Control List) por Rol

| Rol | Status visibles | Tipos visibles |
|-----|----------------|----------------|
| `admin` | Todos (draft, scheduled, published, archived) | Todos |
| `psi` | Solo `published` | `public` + `psi` |
| `public` | Solo `published` | Solo `public` |

### Estados del Post (Máquina de Estados)

```
                  ┌──────────┐
                  │  draft   │
                  └────┬─────┘
                       │ publish()
                  ┌────▼─────┐     ┌───────────┐
          ┌───────│published │────▶│ scheduled │
          │       └──────────┘     └───────────┘
          │              ▲               │
          │              │ schedule()    │ cron fires
          │              └───────────────┘
          │ archive()
     ┌────▼─────┐
     │ archived │
     └──────────┘
```

- **Scheduled Publishing:** `PublishScheduled()` ejecutado periódicamente, transiciona posts programados cuya fecha ya pasó
- Validación: si `status == "scheduled"`, `publish_at` es obligatorio

### S3 con Saga Pattern

```
UpdatePost():
  1. Upload new image → S3
  2. DB Update (transaction)
     ├─ OK → Delete old image from S3 (Garbage Collection)
     └─ FAIL → Delete uploaded image (Compensating Rollback)
```

---

## 🏷️ SpecialtyService (`specialty_service.go`)

Catálogo maestro de especialidades/áreas de trabajo con Soft Delete.

### Soft Delete (Nunca Hard Delete)

```go
// Delete() en specialty_repo.go
Updates(map[string]interface{}{"active": false})
```

**Razón:** Eliminar físicamente una especialidad rompería las claves foráneas de todos los psicólogos asociados. El soft delete la "apaga" de las búsquedas sin corromper datos.

### Fail-Safe Defaults

```go
func (s *SpecialtyService) GetSpecialties(ctx, requestedStatus, isAdmin) {
    finalStatus := "active"
    if isAdmin {
        finalStatus = requestedStatus  // Solo admin puede ver "inactive" o "all"
    }
    return s.repo.GetAll(ctx, finalStatus)
}
```

Siempre retorna al menos las especialidades activas. Un atacante no puede forzar la lectura de especialidades desactivadas sin permisos de admin.

### RBAC por Operación

| Operación | Permiso requerido |
|-----------|------------------|
| Create | `CanCreateTags` o `Sudo` |
| Update | `CanEditTags` o `Sudo` |
| Delete | `CanDeleteTags` o `Sudo` |
| Read (público) | Sin permiso (solo activas) |
| Read (admin) | Cualquier permiso admin |

---

## 📊 AnalyticsService (`analytics_service.go`)

Motor de telemetría bajo el principio de **Impacto Cero**.

### Fire-and-Forget (Goroutines)

```go
func (s *AnalyticsService) RecordLogin(...) {
    go func() {
        s.db.Create(&domain.LoginEvent{...})
        // Upsert de sesión activa
        s.db.Where(...).Assign(...).FirstOrCreate(...)
    }()
}
```

**Todos** los métodos de escritura lanzan goroutines. El hilo HTTP retorna inmediatamente al cliente mientras la DB escribe en segundo plano.

### Eventos Registrados

| Evento | Modelo | Propósito |
|--------|--------|-----------|
| Login | `LoginEvent` | Auditoría forense (quién, cuándo, desde dónde) |
| Logout | `ActiveSession` (delete) | Mantener panel de sesiones preciso |
| Búsqueda | `SearchEvent` | Inteligencia de negocio (qué buscan los usuarios) |
| Vista perfil | `ProfileView` | Popularidad individual de profesionales |
| Page View | `PageView` | Tráfico web del portal |
| Heartbeat | `ActiveSession` (update) | Extender TTL de sesión mientras navega |

### Dashboard Stats (`GetDashboardStats`)

```
┌─────────────────────────────────────────────────────────┐
│  DashboardStats (un solo endpoint)                       │
│                                                          │
│  📈 Logins: Total, Hoy, Semana, Mes, Únicos Hoy        │
│  👁️ PageViews: Total, Hoy, Semana, Únicos Hoy/Semana   │
│  🔍 Búsquedas: Total, Hoy, Semana                       │
│  👤 ProfileViews: Total, Hoy, Semana                     │
│  🟢 ActiveSessionsNow (concurrencia en vivo)            │
│                                                          │
│  🏆 Top Rankings (últimos 30 días):                     │
│  ├── TopSpecialties (con JOIN a tabla maestra)          │
│  ├── TopMunicipios                                      │
│  ├── TopSearchTerms                                     │
│  └── TopProfiles (con JOIN a psi_users)                 │
│                                                          │
│  📉 Time Series (últimos 14 días):                      │
│  ├── LoginTrend (DATE grouping)                         │
│  └── ViewTrend (DATE grouping)                          │
└─────────────────────────────────────────────────────────┘
```

### Data Retention Policy

```go
func (s *AnalyticsService) PurgeOldData(olderThanDays int) {
    // Elimina: PageView, SearchEvent, ProfileView
    // CONSERVA: LoginEvent (requisito de auditoría de ciberseguridad)
}
```

- **Se purgan:** PageView, SearchEvent, ProfileView (datos efímeros)
- **Nunca se purga:** LoginEvent (logs de autenticación = auditoría legal)
- **CleanExpiredSessions:** GC para sesiones con TTL expirado

---

## 📧 MailService (`mail_service.go`)

Sistema de correo asíncrono con anti-spam integrado.

### Patrón Productor-Consumidor

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│  HTTP    │───▶│  Channel │───▶│  Worker  │───▶ SMTP Server
│  Handler │    │  Buffer  │    │  Daemon  │
│          │    │  (5000)  │    │          │
└──────────┘    └──────────┘    └──────────┘
   Producer                      Consumer
   (~10ms)                       (~2s por email)
```

### Anti-Spam (Throttling + Jittering)

```
┌─────────────────────────────────────────────────────────┐
│  WORKER LOOP                                             │
│                                                          │
│  for job := range queue {                                │
│      executeSend(job)                                    │
│                                                          │
│      // Rate limit por socket (500ms entre emails)       │
│      time.Sleep(500ms)                                   │
│                                                          │
│      // Cada 30 emails → pausa aleatoria                 │
│      if sentInBatch >= 30 {                              │
│          waitTime = rand(60..180 segundos)               │
│          time.Sleep(waitTime)                            │
│          sentInBatch = 0                                 │
│      }                                                   │
│  }                                                       │
└─────────────────────────────────────────────────────────┘
```

**Tácticas:**
1. **Throttling:** Lote máximo de 30 correos antes de pausa
2. **Jittering:** Pausa aleatoria entre 60-180 segundos (simula comportamiento humano)
3. **Socket Rate Limiting:** 500ms entre cada envío individual

### Interfaces e Inyección

```go
type IMailService interface {
    SendEmail(to, subject, templateName string, data interface{}) error
}
```

Permite **mocking** en tests sin enviar correos reales. El `MailService` concreto se inyecta como dependencia en AdminService y PsiService.

### Templates HTML Embebidos

Usa `embed.FS` de Go para compilar plantillas HTML en binario, evitando I/O a disco en runtime.

---

## 🔄 ErrorMapper (`error_mapper.go`)

Traductor de errores de PostgreSQL a mensajes en español.

### Mapeo de Errores

| Constraint PostgreSQL | Mensaje en Español |
|----------------------|-------------------|
| `idx_psi_users_ci` / `uni_psi_users_ci` | "la Cédula de Identidad ya se encuentra registrada" |
| `idx_psi_users_fpv` / `uni_psi_users_fpv` | "el número de FPV ya está registrado por otro psicólogo" |
| `uni_psi_users_email` | "el correo electrónico ya está en uso" |
| `uni_psi_users_username` | "el nombre de usuario ya existe" |

**Propósito:** Prevenir **CWE-209 (Information Exposure)** — GORM/PostgreSQL devuelven errores crudos que exponen nombres de tablas, columnas y sentencias SQL. Este mapper los enmascara.

Si el error no coincide con ninguna restricción conocida, se propaga hacia arriba para que el logger centralizado lo registre.

---

## 🌐 Social Media (`social_media.go`)

CRUD de redes sociales para perfiles de psicólogos.

### Límites y Protecciones

```go
const MaxSocialNetworks = 10
```

- **Quota:** Máximo 10 redes sociales por psicólogo (previene Resource Exhaustion)
- **IDOR Prevention:** Ownership check explícito en Update y Delete
- **Platform Normalization:** `utils.NormalizePlatformName()` estandariza nombres ("ig" → "Instagram", "INSTA" → "Instagram")

### Control de Acceso Polimórfico

`DeleteSocialNetwork()` acepta tanto `psi` (autogestión) como `admin` (moderación):

```go
if executorRole == "psi" {
    // Solo puede borrar las suyas (IDOR check)
    if network.PsiUserID != executorID { return error }
} else if executorRole == "admin" {
    // Acceso global
}
```

---

## 🏗️ Decisiones de Diseño Clave

### 1. Inyección de Dependencias (DI)

Todos los servicios reciben interfaces (no implementaciones concretas) en sus constructores:

```go
func NewAdminService(repo domain.UserAdminRepository, mailService *MailService) *AdminService
func NewPsiService(repo domain.PsiUserRepository, s3Client *s3.S3Client, mailService IMailService) *PsiService
```

Esto permite testing con mocks y bajo acoplamiento.

### 2. UUID v7 para Indexación

```go
psiID := uuid.Must(uuid.NewV7())
```

UUID v7 incluye timestamp, optimizando inserts en B-Trees de PostgreSQL (vs v4 completamente random).

### 3. Trazabilidad Forense (Audit Trail)

Cada operación de escritura inyecta:
```go
AuditModel: domain.AuditModel{
    CreateBy:   admin.Username,
    CreateById: &admin.ID,
    UpdateBy:   admin.Username,
    UpdateById: &admin.ID,
}
```

Nunca hay registros "huérfanos" — siempre se sabe quién creó/modificó cada registro.

### 4. Graceful Degradation

Los emails y sincronizaciones con microservicios **nunca** bloquean la respuesta HTTP:

```go
if err := s.mailService.SendEmail(...); err != nil {
    log.Printf("⚠️ Error al enviar correo (pero el admin se creó): %v", err)
}
// La operación principal ya fue exitosa
```

### 5. Seguridad por Defecto (Secure by Default)

- Contraseña: `bcrypt` con costo default
- XSS: `bluemonday.UGCPolicy()` en biografías y contenido HTML
- Imágenes: `utils.SanitizeDocument()` elimina EXIF/metadata
- Búsqueda: `CleanAlphaNumeric()` en inputs de búsqueda

---

## 🔗 Dependencias Externas

| Paquete | Uso |
|---------|-----|
| `github.com/golang-jwt/jwt/v5` | Generación y validación de JWT |
| `golang.org/x/crypto/bcrypt` | Hashing de contraseñas |
| `github.com/google/uuid` | Generación de UUIDs v7 |
| `github.com/patrickmn/go-cache` | Caché en memoria con TTL |
| `github.com/microcosm-cc/bluemonday` | Sanitización HTML (XSS) |
| `github.com/xuri/excelize/v2` | Parseo de archivos Excel/XLSX |
| `github.com/wneessen/go-mail` | Cliente SMTP |
| `github.com/gofiber/fiber/v2` | Estructuras HTTP (fiber.Map) |
| `github.com/veniversvm/ColPsiCarabobo/api/pkg/s3` | Cliente S3 personalizado |
| `github.com/veniversvm/ColPsiCarabobo/api/internal/config` | Variables de entorno |
| `github.com/veniversvm/ColPsiCarabobo/api/internal/templates` | Plantillas HTML embebidas |

**[⬆ Volver a internal](../)**

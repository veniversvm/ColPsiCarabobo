# 📋 Modelo de Dominio (domain/)

> El **corazón del sistema**. Define qué sabe el sistema: entidades, modelos de datos y contratos de persistencia. Este paquete NO tiene dependencias de infraestructura — solo define acuerdos.

## Arquitectura

El paquete `domain/` implementa el patrón **Repository Interface** de Clean Architecture. Cada interfaz describe *qué* se puede hacer con los datos, sin importar *cómo* se hace (PostgreSQL, S3, etc.).

```
handler/ ──→ service/ ──→ domain/ (interfaces) ──→ infrastructure/ (implementaciones)
```

---

## 📦 Modelos

### AuditModel — Base de Auditoría

Estructura incrustada (embedded) en todos los modelos del dominio. Proporciona trazabilidad completa y borrado lógico.

| Campo | Tipo | Descripción |
|---|---|---|
| `CreatedAt` | `time.Time` | Fecha de creación (auto-managed por GORM) |
| `UpdatedAt` | `time.Time` | Última modificación (auto-managed por GORM) |
| `DeletedAt` | `gorm.DeletedAt` | Soft Delete — las consultas lo ignoran automáticamente |
| `CreateBy` | `string` | Nombre/texto del creador (auditoría humana) |
| `UpdateBy` | `string` | Nombre/texto del último editor |
| `CreateById` | `*uuid.UUID` | UUID del admin que creó el registro |
| `UpdateById` | `*uuid.UUID` | UUID del admin que modificó el registro |

---

### UserAdmin — Personal Administrativo

Implementa un sistema RBAC (Role-Based Access Control) con permisos granulares por módulo.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key (UUIDv7) |
| `Username` | `string` | Nombre de usuario (unique, max 25) |
| `Email` | `string` | Correo electrónico (unique) |
| `Password` | `string` | Hash bcrypt — **nunca expuesto en JSON** (`json:"-"`) |
| `Key` | `string` | Semilla de firma JWT —用于 rotación de sesiones (`json:"-"`) |
| `IsActive` | `bool` | Estado de la cuenta |
| `Sudo` | `bool` | Acceso total e irrevocable — solo asignable fuera de la API |

**Permisos granulares (booleans):**

- **Gestión de Colegiados:** `CanCreatePsi`, `CanUpdatePsi`, `CanDeletePsi`
- **Gestión de Personal:** `CanCreateAdmin`, `CanUpdateAdmin`, `CanDeleteAdmin`
- **Contenido:** `CanPublish`, `CanUpdatePublish`, `CanDeletePublish`
- **Notificaciones:** `CanSendNotifications`, `CanManageNotifications`, `CanReadNotifications`
- **Especialidades:** `CanCreateTags`, `CanEditTags`, `CanDeleteTags`

**Tabla:** `user_admins`

---

### PsiUserModel — Perfil de Psicólogo (Entidad Principal)

La entidad central del sistema. Agrupa identidad legal, contacto, ubicación geográfica, estado gremial y relaciones con otros módulos.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key (UUIDv7) |
| `Username` | `string` | Nombre de usuario (unique) |
| `Email` | `string` | Email institucional (unique) |
| `Password` | `string` | Hash bcrypt (`json:"-"`) |
| `Key` | `string` | Semilla JWT para rotación (`json:"-"`) |
| `IsActive` | `bool` | Estado de la cuenta |
| `AudioBookShellId` | `string` | ID del servicio de audiobooks |
| `FirstName` | `string` | Primer nombre |
| `SecondName` | `string` | Segundo nombre |
| `LastName` | `string` | Primer apellido |
| `SecondLastName` | `string` | Segundo apellido |
| `FPV` | `int` | Número de Federación de Psicólogos de Venezuela (unique index) |
| `CI` | `int` | Cédula de Identidad (unique index) |
| `Nationality` | `string` | `V` = venezolano, `E` = extranjero |
| `BornDate` | `time.Time` | Fecha de nacimiento |
| `Genre` | `string` | `M` = masculino, `F` = femenino |
| `Solvent` | `bool` | Solvencia con el Colegio |
| `ProofOfLife` | `bool` | Fe de vida presentada |
| `ProfilePictureS3Key` | `string` | S3 key de la foto de perfil |

**Campos de contacto (con flags de visibilidad `show_*`):**

- Contacto general: `ContactPhone`, `ContactCellPhone`, `ContactEmail`, `ServiceAddress`
- **Ubicación Carabobo:** `MunicipalityCarabobo`, `PhoneCarabobo`, `CelPhoneCarabobo` + flags show
- **Fuera de Carabobo (Venezuela):** `StateOutside`, `MunicipalityOutSideCarabobo`, teléfonos + flags show
- **Fuera de Venezuela:** `Country` (ISO 3166-1 alpha-2), teléfonos + flags show

**Especialidades:** `PrimaryWorkArea`, `SecondaryWorkArea`

**Biografía:** `MiniBio` (max 250 chars), `BioTextID` (FK → TextModel), `FullBio` (GORM preload)

**Tabla:** `psi_users`

#### Relaciones

```
PsiUserModel ──1:1──→ PsiUserColData
PsiUserModel ──1:N──→ PsiUserSocialNetwork
PsiUserModel ──1:N──→ PsiUserPostGrade
PsiUserModel ──1:N──→ PsiUserSolvency
PsiUserModel ──1:N──→ PsiObservations
PsiUserModel ──1:N──→ PsiODeontologia
PsiUserModel ──N:1──→ PsiSpecialtyModel (vía PrimaryWorkArea/SecondaryWorkArea)
PsiUserModel ──N:1──→ TextModel (vía BioTextID)
```

---

### PsiUserColData — Datos Colegiales

Información regulatoria y académica del Colegio. Relación **1:1** con `PsiUserModel`.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key |
| `PsiUserModelID` | `uuid.UUID` | FK unique → PsiUserModel |
| `GuildInscriptionDate` | `time.Time` | Fecha de inscripción gremial |
| `UniversityUndergraduate` | `string` | Universidad de pregrado |
| `GraduateDate` | `time.Time` | Fecha de graduación |
| `MentionUndergraduate` | `string` | Mención de pregrado |
| `TitleImageOneS3Key` | `string` | S3 key del título 1 |
| `TitleImageTwoS3Key` | `string` | S3 key del título 2 |
| `TitleImageThreeS3Key` | `string` | S3 key del título 3 |
| `RegisterTitleState` | `string` | Estado del registro del título |
| `RegisterTitleDate` | `time.Time` | Fecha de registro |
| `RegisterNumber` | `int` | Número de registro |
| `RegisterFolio` | `string` | Folio |
| `RegisterTome` | `string` | Tomo |
| `GuildDirector` | `bool` | Miembro de la Junta Directiva |
| `SixtyFiveOrPlus` | `bool` | Mayor de 65 años |
| `GuildCollaborator` | `bool` | Colaborador activo del Colegio |
| `PublicEmployee` | `bool` | Empleado público |
| `Discapacity` | `bool` | Persona con discapacidad |
| `UniversityProfessor` | `bool` | Docente universitario |
| `DateOfLastSolvency` | `time.Time` | Última fecha de pago de cuota |
| `DoubleGuild` | `bool` | Colegiado en más de un estado |
| `DoubleGuildLocation` | `string` | Ubicación de doble colegiación |
| `CPSM` | `bool` | Miembro del Colegio de Psicólogos de Miranda |

Todos los campos de contacto/ubicación tienen flags `Show*` para control de visibilidad pública.

**Tabla:** `psi_user_col_data`

---

### PsiUserSocialNetwork — Redes Sociales

Enlaces digitales del psicólogo. Relación **1:N** con `PsiUserModel`.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key |
| `PsiUserID` | `uuid.UUID` | FK → PsiUserModel |
| `Name` | `string` | Nombre de la plataforma (Instagram, LinkedIn, etc.) |
| `URL` | `string` | Enlace directo al perfil |
| `IsActive` | `bool` | Visibilidad (puede ocultarse sin borrar) |

**Tabla:** `psi_user_social_networks`

---

### PsiUserSolvency — Registro de Solvencias

Historial de pagos de cuotas del psicólogo. Relación **1:N** con `PsiUserModel`.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key |
| `PsiUserModelID` | `uuid.UUID` | FK → PsiUserModel (unique index compuesto) |
| `Date` | `time.Time` | Fecha del pago (unique index compuesto) |

**Tabla:** `psi_user_solvency`

---

### PsiUserPostGrade — Postgrados

Títulos académicos adicionales: Especializaciones, Maestrías, Doctorados, Diplomados. Relación **1:N** con `PsiUserModel`.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key |
| `PsiUserID` | `uuid.UUID` | FK → PsiUserModel |
| `Type` | `PostGradeType` | Enum: `diplomado`, `especializacion`, `maestria`, `doctorado` |
| `Title` | `string` | Título obtenido |
| `University` | `string` | Universidad |
| `GraduationYear` | `string` | Año de graduación |
| `Description` | `string` | Descripción opcional |
| `Active` | `bool` | Estado del registro |
| `PicOneS3Key` | `string` | S3 key del certificado 1 |
| `PicTwoS3Key` | `string` | S3 key del certificado 2 |
| `PicThreeS3Key` | `string` | S3 key del certificado 3 |

**Tabla:** `psi_user_post_grades`

---

### PsiObservaciones — Observaciones Internas

Notas internas del Colegio sobre un psicólogo. **ACCESO EXCLUSIVO** del personal autorizado — los psicólogos NUNCA pueden verlas.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key |
| `PsiUserID` | `uuid.UUID` | FK → PsiUserModel |
| `Content` | `string` | Contenido de la observación (text) |

**Tabla:** `psi_observations`

---

### PsiODeontologia — Expediente Deontológico

Registros disciplinarios o deontológicos abiertos por el Tribunal Disciplinario. **ACCESO EXCLUSIVO** del personal autorizado.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key |
| `PsiUserID` | `uuid.UUID` | FK → PsiUserModel |
| `Content` | `string` | Contenido del expediente (text) |

**Tabla:** `psi_deontologia`

---

### PsiSpecialtyModel — Catálogo de Especialidades

Catálogo de especialidades psicológicas. Usa `uint32` como PK (no UUID) para optimizar índices de búsqueda.

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uint32` | Primary Key (auto-increment) |
| `Name` | `string` | Nombre de la especialidad (unique index, max 100) |
| `Description` | `string` | Descripción (text) |
| `Active` | `bool` | Estado del catálogo (default: true) |
| `AuditModel` | embedded | Campos de auditoría |

**Tabla:** `psi_specialty_models`

---

### TextModel — Contenido Largo

Almacena contenido extenso de forma aislada (biografías, artículos). Separación por rendimiento y seguridad (anti-XSS).

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key |
| `Content` | `string` | Texto enriquecido (type text en Postgres) |
| `AuditModel` | embedded | Campos de auditoría |

---

### Post — Publicaciones (CMS)

Publicaciones, noticias y anuncios de la plataforma. Separa metadata (`Post`) de contenido extenso (`TextModel`).

| Campo | Tipo | Descripción |
|---|---|---|
| `ID` | `uuid.UUID` | Primary Key |
| `Title` | `string` | Título (max 100 chars) |
| `ShortDescription` | `string` | Resumen para el feed (max 250 chars) |
| `Type` | `string` | Visibilidad: `public`, `psi` |
| `TextID` | `uuid.UUID` | FK → TextModel |
| `ImageS3Key` | `string` | S3 key de la imagen de portada |
| `Status` | `PostStatus` | Enum: `draft`, `published`, `archived`, `scheduled` |
| `PublishAt` | `*time.Time` | Fecha de publicación (para status=scheduled) |

**Tabla:** `posts`

---

### Modelos de Analytics

Métricas de uso del sistema — no tienen `AuditModel`, solo campos de trazabilidad temporal.

| Modelo | PK | Descripción | Campos clave |
|---|---|---|---|
| `LoginEvent` | `uuid.UUID` | Cada inicio de sesión | UserID, Username, Role, IP, UserAgent |
| `PageView` | `uint` (auto) | Cada visita a una ruta | Path, Method, UserID, SessionID, Referer |
| `SearchEvent` | `uint` (auto) | Cada búsqueda en directorio | Query, Specialty, Municipality, ResultsCount |
| `ProfileView` | `uint` (auto) | Cada visita a perfil de psicólogo | PsiID, ViewerID, SessionID |
| `ActiveSession` | `uuid.UUID` | Sesiones activas (heartbeat) | UserID, Username, Role, IP, LastSeen, ExpiresAt |

**Método utilitario:** `ActiveSession.IsActive()` — retorna `true` si la sesión no ha expirado.

**Tablas:** `login_events`, `page_views`, `search_events`, `profile_views`, `active_sessions`

---

## 🔌 Interfaces de Repositorio

### PsiUserRepository

Interfaz más extensa del sistema. Gestión integral de psicólogos.

```go
type PsiUserRepository interface {
    // ── Perfil y Datos Core ──
    CreateWithColData(ctx, psi, colData, solvency, postgrades) error
    GetByID(ctx, id) (*PsiUserModel, error)
    GetByFPV(ctx, id) (PsiUserModel, error)
    GetByIdentifier(ctx, identifier) (*PsiUserModel, error)
    GetPsiUserColData(ctx, psiID) (*PsiUserColData, error)
    GetTextContentByID(ctx, id) (string, error)
    ValidateUniqueCredentials(ctx, username, email, excludeID) error
    GetSitemapData(ctx) ([]PsiUserModel, error)

    // ── Mutaciones ──
    Update(ctx, psi, colData, bioText, solvencies) error
    UpdatePublicProfile(ctx, psi, colData, bioText) error
    UpdateKey(ctx, psi) error
    Delete(ctx, id) error

    // ── Búsqueda y Directorio ──
    SearchDirectory(ctx, filter) ([]PsiUserModel, int64, error)
    Search(ctx, filters, page, pageSize) ([]PsiUserModel, int64, error)
    SearchAdmin(ctx, filter) ([]PsiUserModel, int64, error)

    // ── Postgrados ──
    CreatePostGrade(ctx, pg) error
    GetPostGradeByID(ctx, id) (*PsiUserPostGrade, error)
    UpdatePostGrade(ctx, pg) error

    // ── Solvencias ──
    CreateSolvency(ctx, pg) error
    GetSolvencies(ctx, id) ([]PsiUserSolvency, error)
    CreateOrUpdateSolvencies(ctx, solvencies) error

    // ── Redes Sociales ──
    CreateSocialNetwork(ctx, sn) error
    GetSocialNetworkByID(ctx, id) (*PsiUserSocialNetwork, error)
    UpdateSocialNetwork(ctx, sn) error
    DeleteSocialNetwork(ctx, id) error
    CountSocialNetworksByPsiID(ctx, psiID) (int64, error)
}
```

### PostRepository

Gestión de publicaciones CMS.

```go
type PostRepository interface {
    Create(ctx, post, content) error
    Update(ctx, post, content) error
    Delete(ctx, id) error
    GetByID(ctx, id) (*Post, error)
    List(ctx, filter, page, limit) ([]Post, int64, error)
    PublishScheduled(ctx) int64
    GetSitemapPosts(ctx) ([]Post, error)
}
```

### UserAdminRepository

Gestión de personal administrativo.

```go
type UserAdminRepository interface {
    GetByID(ctx, id) (*UserAdmin, error)
    List(ctx, active, search, page, limit) ([]UserAdmin, int64, error)
    GetByIdentifier(ctx, identifier) (*UserAdmin, error)
    Create(ctx, user) error
    Update(ctx, user) error
    Delete(ctx, id) error
    CountSudos(ctx) (int64, error)
}
```

### SpecialtyRepository

Catálogo de especialidades — usa **desactivación lógica** en lugar de hard delete.

```go
type SpecialtyRepository interface {
    Create(ctx, s) error
    GetAll(ctx, status) ([]PsiSpecialtyModel, error)
    GetByID(ctx, id, active) (*PsiSpecialtyModel, error)
    GetByAdminID(ctx, id) (*PsiSpecialtyModel, error)
    Update(ctx, s) error
    Delete(ctx, id) error    // soft-delete: marca active = false
    Count(ctx, actives) (int64, error)
    GetAllAdmin(ctx) ([]PsiSpecialtyModel, error)
}
```

---

## 🗂 Diagrama ER

```mermaid
erDiagram
    AuditModel {
        time CreatedAt
        time UpdatedAt
        gorm.DeletedAt DeletedAt
        string CreateBy
        string UpdateBy
        uuid CreateById
        uuid UpdateById
    }

    UserAdmin {
        uuid ID PK
        string Username UK
        string Email UK
        string Password
        string Key
        bool IsActive
        bool Sudo
        bool CanCreatePsi
        bool CanPublish
    }

    PsiUserModel {
        uuid ID PK
        string Username UK
        string Email UK
        string Password
        string FirstName
        string LastName
        int FPV UK
        int CI UK
        string Nationality
        date BornDate
        string Genre
        bool Solvent
        string ProfilePictureS3Key
        string PrimaryWorkArea
        string SecondaryWorkArea
        string MiniBio
        uuid BioTextID FK
    }

    PsiUserColData {
        uuid ID PK
        uuid PsiUserModelID FK UK
        date GuildInscriptionDate
        string UniversityUndergraduate
        date GraduateDate
        string RegisterTitleState
        bool GuildDirector
        bool UniversityProfessor
        date DateOfLastSolvency
        bool DoubleGuild
    }

    PsiUserSocialNetwork {
        uuid ID PK
        uuid PsiUserID FK
        string Name
        string URL
        bool IsActive
    }

    PsiUserSolvency {
        uuid ID PK
        uuid PsiUserModelID FK
        date Date
    }

    PsiUserPostGrade {
        uuid ID PK
        uuid PsiUserID FK
        PostGradeType Type
        string Title
        string University
        string GraduationYear
        bool Active
    }

    PsiObservations {
        uuid ID PK
        uuid PsiUserID FK
        string Content
    }

    PsiODeontologia {
        uuid ID PK
        uuid PsiUserID FK
        string Content
    }

    PsiSpecialtyModel {
        uint32 ID PK
        string Name UK
        string Description
        bool Active
    }

    TextModel {
        uuid ID PK
        string Content
    }

    Post {
        uuid ID PK
        string Title
        string ShortDescription
        string Type
        uuid TextID FK
        string ImageS3Key
        PostStatus Status
        time PublishAt
    }

    PageView {
        uint ID PK
        string Path
        string Method
        uuid UserID FK
        string SessionID
        string IP
    }

    ProfileView {
        uint ID PK
        uuid PsiID FK
        uuid ViewerID FK
        string SessionID
    }

    SearchEvent {
        uint ID PK
        string Query
        string Specialty
        string Municipality
        int ResultsCount
    }

    LoginEvent {
        uuid ID PK
        uuid UserID FK
        string Username
        string Role
        string IP
    }

    ActiveSession {
        uuid ID PK
        uuid UserID FK UK
        string Username
        string Role
        string IP
        time LastSeen
        time ExpiresAt
    }

    PsiUserModel ||--|| PsiUserColData : "1:1 col_data"
    PsiUserModel ||--o{ PsiUserSocialNetwork : "1:N social_networks"
    PsiUserModel ||--o{ PsiUserSolvency : "1:N solvencies"
    PsiUserModel ||--o{ PsiUserPostGrade : "1:N post_grades"
    PsiUserModel ||--o{ PsiObservations : "1:N observations"
    PsiUserModel ||--o{ PsiODeontologia : "1:N deontologia"
    PsiUserModel }o--|| TextModel : "N:1 full_bio"
    Post ||--|| TextModel : "1:1 text"
```

---

## 📊 Resumen de Tablas

| # | Modelo | Tabla | PK | Relaciones |
|---|---|---|---|---|
| 1 | UserAdmin | `user_admins` | UUID | — |
| 2 | PsiUserModel | `psi_users` | UUID | → TextModel |
| 3 | PsiUserColData | `psi_user_col_data` | UUID | → PsiUserModel (1:1) |
| 4 | PsiUserSocialNetwork | `psi_user_social_networks` | UUID | → PsiUserModel (1:N) |
| 5 | PsiUserSolvency | `psi_user_solvency` | UUID | → PsiUserModel (1:N) |
| 6 | PsiUserPostGrade | `psi_user_post_grades` | UUID | → PsiUserModel (1:N) |
| 7 | PsiObservations | `psi_observations` | UUID | → PsiUserModel (1:N) |
| 8 | PsiODeontologia | `psi_deontologia` | UUID | → PsiUserModel (1:N) |
| 9 | PsiSpecialtyModel | `psi_specialty_models` | uint32 | — |
| 10 | TextModel | `text_models` | UUID | — |
| 11 | Post | `posts` | UUID | → TextModel (1:1) |
| 12 | LoginEvent | `login_events` | UUID | — |
| 13 | PageView | `page_views` | uint | — |
| 14 | SearchEvent | `search_events` | uint | — |
| 15 | ProfileView | `profile_views` | uint | — |
| 16 | ActiveSession | `active_sessions` | UUID | — |

**Constraints UNIQUE:** `cedula (CI)`, `email`, `username`, `FPV`, `phone` (compuestos por tabla).

---

## 🏗 Convenciones

- **IDs:** UUIDv7 para entidades de negocio (ordenados por tiempo), `uint32` para catálogos de solo lectura
- **Soft Delete:** Todos los modelos usan `AuditModel.DeletedAt` — nunca se borran filas físicamente
- **Auditoría:** Cada mutación registra quién (`CreateById/UpdateById`) y cuándo (`CreatedAt/UpdatedAt`)
- **JSON:** Password y Key siempre `json:"-"` — jamás se exponen en respuestas API
- **Visibilidad:** Los campos sensibles de PsiUserModel tienen flags `Show*` que controlan el filtrado en el directorio público
- **S3:** Las imágenes se almacenan como S3 keys, nunca como URLs completas (flexibilidad de bucket)

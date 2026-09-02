# Plan: Sistema de Notificaciones — ColPsiCarabobo

> **Versión corregida** — Adaptado a la arquitectura real del proyecto
> (Clean Architecture: `router → handler → service → repository → DB`)

## 1. Objetivo

Implementar un sistema de notificaciones in-app para el Colegio de Psicólogos del Estado Carabobo que permita a los administradores enviar notificaciones de forma **inmediata o programada** a agremiados, con targeting por grupos (ciudad, estado, sexo, especialidad, solvencia) y soporte para adjuntos (imágenes/archivos vía S3). El envío por correo electrónico queda como opción habilitable usando el `IMailService` existente.

---

## 2. Requisitos funcionales

| # | Requisito | Detalle |
|---|---|---|
| RF-01 | Notificación global | Se envía a todos los agremiados elegibles |
| RF-02 | Notificación individual | Se envía a uno o más agremiados específicos por UUID |
| RF-03 | Notificación por grupos | Filtros combinables: ciudad (`municipality_carabobo`), estado (`state_outside`), sexo (`genre`), especialidad (`primary_specialty_id`/`secondary_specialty_id`), solvencia (`solvent`) |
| RF-04 | Preview de destinatarios | Antes de enviar, el admin puede ver cuántos y quiénes reciben la notificación |
| RF-05 | Envío inmediato | La notificación se envía al momento de crearla |
| RF-06 | Envío programado | El admin puede elegir fecha/hora futura para el envío |
| RF-07 | Adjuntos | Permitir adjuntar imágenes/archivos a la notificación (S3) |
| RF-08 | Email opcional | Opción de enviar copia por correo usando `IMailService` existente |
| RF-09 | Marcar como leída | El agremiado marca individualmente cada notificación como leída |
| RF-10 | Badge de no leídas | Endpoint para obtener el contador de notificaciones no leídas |
| RF-11 | Exclusión de fallecidos | No se envía a usuarios con `ProofOfLife = false` |
| RF-12 | Exclusión de inactivos | No se envía a usuarios con `IsActive = false` |
| RF-13 | Cancelación | El admin puede cancelar notificaciones programadas antes de que se envíen |

---

## 3. Requisitos no funcionales

| # | Requisito | Detalle |
|---|---|---|
| RNF-01 | Arquitectura | Seguir Clean Architecture: `router → handler → service → repository → DB` |
| RNF-02 | Storage | Adjuntos en S3/MinIO via `pkg/s3/` (consistente con posts) |
| RNF-03 | UUIDs | Todas las PKs son UUIDs v7 (`uuidv7()`), consistente con el resto |
| RNF-04 | Audit trail | Usar `AuditModel` existente (`CreatedAt`, `UpdatedAt`, `DeletedAt`, `CreateBy`, `UpdateBy`, `CreateById`, `UpdateById`) |
| RNF-05 | Consistencia | Mismos patrones de error: `fiber.Map{"error": "mensaje"}`, errores sentinela en `domain/errors.go` |
| RNF-06 | Seguridad | Reusar `ProtectedAdmin404()` y `ProtectedPsiUser()` existentes |
| RNF-07 | Rate limit emails | Reutilizar el worker async con throttling del `MailService` existente |

---

## 4. Campos existentes en `PsiUserModel` utilizados para targeting

| Campo | Tipo | Uso |
|---|---|---|
| `ID` | uuid | Identificador único del destinatario |
| `IsActive` | bool | Excluir inactivos (`false` = excluir) |
| `ProofOfLife` | bool | Excluir fallecidos (`false` = fallecido) |
| `Solvent` | bool | Filtro por estado de cuotas |
| `Genre` | string (1) | Filtro por sexo (`M`/`F`) |
| `MunicipalityCarabobo` | string | Filtro por municipio de Carabobo |
| `StateOutside` | string | Filtro por estado (fuera de Carabobo) |
| `PrimarySpecialtyID` | *uint32 | Filtro por especialidad primaria (FK al catálogo) |
| `SecondarySpecialtyID` | *uint32 | Filtro por especialidad secundaria (FK al catálogo) |
| `ContactEmail` | string | Destinatario del correo (si `send_email = true`) |

**No se requieren cambios en `PsiUserModel`.**

---

## 5. Modelos de datos

### 5.1 `Notification` — Tabla principal

**Archivo:** `internal/domain/notification.model.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationTargetType define el alcance de una notificación.
type NotificationTargetType string

const (
	NotificationTargetGlobal     NotificationTargetType = "global"
	NotificationTargetIndividual NotificationTargetType = "individual"
	NotificationTargetGroup      NotificationTargetType = "group"
)

// NotificationStatus define el ciclo de vida de una notificación.
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// Notification representa una notificación enviada por un administrador a agremiados.
type Notification struct {
	ID         uuid.UUID              `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel
	Title       string                 `gorm:"size:255;not null" json:"title"`
	Message     string                 `gorm:"type:text;not null" json:"message"`
	TargetType  NotificationTargetType `gorm:"size:20;not null" json:"target_type"`
	SenderID    uuid.UUID              `gorm:"type:uuid;not null" json:"sender_id"`
	SendEmail   bool                   `gorm:"default:false" json:"send_email"`
	ScheduledAt *time.Time             `gorm:"type:timestamptz" json:"scheduled_at,omitempty"`
	SentAt      *time.Time             `gorm:"type:timestamptz" json:"sent_at,omitempty"`
	Status      NotificationStatus     `gorm:"size:20;default:pending;not null" json:"status"`

	// Relaciones
	Targets  []NotificationTarget  `gorm:"foreignKey:NotificationID" json:"targets,omitempty"`
	Filters  []NotificationFilter  `gorm:"foreignKey:NotificationID" json:"filters,omitempty"`
	Attachs  []NotificationAttach  `gorm:"foreignKey:NotificationID" json:"attachments,omitempty"`
}

func (Notification) TableName() string { return "notifications" }
```

### 5.2 `NotificationTarget` — Destinatarios

**Archivo:** `internal/domain/notification.model.go` (mismo archivo)

```go
// NotificationTarget registra cada destinatario individual de una notificación.
type NotificationTarget struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	NotificationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"notification_id"`
	PsiUserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"psi_user_id"`
	IsRead         bool       `gorm:"default:false" json:"is_read"`
	ReadAt         *time.Time `gorm:"type:timestamptz" json:"read_at,omitempty"`
	EmailSent      bool       `gorm:"default:false" json:"email_sent"`
	EmailSentAt    *time.Time `gorm:"type:timestamptz" json:"email_sent_at,omitempty"`
}

func (NotificationTarget) TableName() string { return "notification_targets" }
```

### 5.3 `NotificationFilter` — Filtros de auditoría

**Archivo:** `internal/domain/notification.model.go` (mismo archivo)

```go
// NotificationFilter almacena los filtros aplicados para resolver destinatarios.
// Propósito: auditoría y reproducción de la query original.
type NotificationFilter struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	NotificationID uuid.UUID `gorm:"type:uuid;not null;index" json:"notification_id"`
	Municipality   string    `gorm:"size:255" json:"municipality,omitempty"`
	State          string    `gorm:"size:255" json:"state,omitempty"`
	Genre          string    `gorm:"size:1" json:"genre,omitempty"`
	SpecialtyID    *uint32   `json:"specialty_id,omitempty"`
	Solvent        *bool     `json:"solvent,omitempty"`
	TargetUserIDs  string    `gorm:"type:text" json:"target_user_ids,omitempty"` // JSON array de UUIDs para tipo "individual"
}

func (NotificationFilter) TableName() string { return "notification_filters" }
```

### 5.4 `NotificationAttach` — Adjuntos (S3)

**Archivo:** `internal/domain/notification.model.go` (mismo archivo)

```go
// NotificationAttach referencia un archivo adjunto almacenado en S3.
type NotificationAttach struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	NotificationID uuid.UUID `gorm:"type:uuid;not null;index" json:"notification_id"`
	Name           string    `gorm:"size:255" json:"name"`
	S3Key          string    `gorm:"size:512;not null" json:"s3_key"`
	ContentType    string    `gorm:"size:100" json:"content_type"`
	AuditModel
}

func (NotificationAttach) TableName() string { return "notification_attachments" }
```

> **Nota:** A diferencia del plan original (bytea en DB de imágenes), los adjuntos se
> almacenan en S3/MinIO via `pkg/s3/`, consistente con el patrón de posts e imágenes
> de perfil.

---

## 6. Contratos de persistencia (Repository Interface)

**Archivo:** `internal/domain/notification_repository.go`

```go
package domain

import (
	"context"

	"github.com/google/uuid"
)

// NotificationFilterParams parámetros para resolver destinatarios.
type NotificationFilterParams struct {
	Municipality string
	State        string
	Genre        string
	SpecialtyID  *uint32
	Solvent      *bool
}

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	CreateTargets(ctx context.Context, targets []NotificationTarget) error
	CreateFilters(ctx context.Context, filters []NotificationFilter) error
	CreateAttach(ctx context.Context, a *NotificationAttach) error

	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	ListBySender(ctx context.Context, senderID uuid.UUID, page, limit int) ([]Notification, int64, error)
	ListByUser(ctx context.Context, psiUserID uuid.UUID, page, limit int) ([]Notification, int64, error)
	GetTargets(ctx context.Context, notificationID uuid.UUID) ([]NotificationTarget, error)
	GetAttachs(ctx context.Context, notificationID uuid.UUID) ([]NotificationAttach, error)

	CountUnread(ctx context.Context, psiUserID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, notificationID, psiUserID uuid.UUID) error

	Cancel(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status NotificationStatus) error

	// ResolveRecipients retorna los UUIDs de psicólogos que cumplen los filtros.
	// Base SIEMPRE activa: is_active = true AND proof_of_life = true.
	ResolveRecipients(ctx context.Context, params NotificationFilterParams) ([]uuid.UUID, error)
	// ResolveAll retorna todos los UUIDs activos y con fe de vida.
	ResolveAll(ctx context.Context) ([]uuid.UUID, error)
	// ResolveByIDs valida que los UUIDs proporcionados existan y estén activos.
	ResolveByIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)

	// GetPsiUserInfo retorna nombre y email de un psicólogo (para preview).
	GetPsiUserInfo(ctx context.Context, ids []uuid.UUID) ([]PsiUserInfo, error)
}

// PsiUserInfo información mínima de un psicólogo para preview de destinatarios.
type PsiUserInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
```

---

## 7. Estructura de directorios (archivos a crear)

Siguiendo la arquitectura Clean Architecture del proyecto:

```
api/
├── internal/
│   ├── domain/
│   │   ├── notification.model.go           ← Modelos GORM (Notification, Target, Filter, Attach)
│   │   └── notification_repository.go      ← Interfaz de persistencia
│   ├── repository/postgres/
│   │   └── notification_repo.go            ← Implementación GORM
│   ├── service/
│   │   └── notification_service.go         ← Lógica de negocio + filter engine
│   ├── request_structs/
│   │   └── request_notifications.go        ← DTOs de entrada
│   ├── handler/
│   │   └── notification_handler.go         ← Adaptadores HTTP
│   └── router/
│       └── notification_router.go          ← Registro de rutas
├── migrations/
│   └── YYYYMMDDHHMMSS_add_notifications.sql  ← Migración Atlas
└── cmd/api/
    └── main.go                             ← (modificar) lanzar worker scheduler
```

---

## 8. Endpoints

### 8.1 Admin (protegidos con `ProtectedAdmin404()`)

| Método | Ruta | Handler | Descripción |
|---|---|---|---|
| `POST` | `/notifications/admin/preview` | `PreviewRecipients` | Retorna destinatarios potenciales sin crear notificación |
| `POST` | `/notifications/admin` | `CreateNotification` | Crear y (si es inmediata) enviar notificación |
| `POST` | `/notifications/admin/:id/attach` | `AttachFile` | Subir imagen/archivo adjunto (multipart) |
| `GET` | `/notifications/admin` | `GetMyNotifications` | Listar notificaciones creadas por el admin (paginado) |
| `GET` | `/notifications/admin/:id` | `GetNotificationDetail` | Detalle + estadísticas (leídos/total) |
| `DELETE` | `/notifications/admin/:id` | `CancelNotification` | Cancelar notificación programada (solo `status = pending`) |
| `GET` | `/notifications/admin/:id/targets` | `GetTargets` | Ver destinatarios con estado de lectura |

### 8.2 Agremiado (protegidos con `ProtectedPsiUser()`)

| Método | Ruta | Handler | Descripción |
|---|---|---|---|
| `GET` | `/notifications/psi-user` | `GetMyNotifications` | Notificaciones del usuario (paginado, DESC) |
| `GET` | `/notifications/psi-user/unread-count` | `GetUnreadCount` | Contador de no leídas (para badge) |
| `GET` | `/notifications/psi-user/:id` | `GetNotificationById` | Detalle (marca como leída automáticamente) |
| `GET` | `/notifications/psi-user/:id/attach/:attachId` | `DownloadAttachment` | Descargar adjunto (URL firmada de S3) |

---

## 9. Request Structs (DTOs de entrada)

**Archivo:** `internal/request_structs/request_notifications.go`

```go
package request_structs

import (
	"time"

	"github.com/google/uuid"
)

// ── Crear notificación ────────────────────────────────────────────────────

type CreateNotificationRequest struct {
	Title       string  `json:"title" validate:"required,max=255"`
	Message     string  `json:"message" validate:"required"`
	TargetType  string  `json:"target_type" validate:"required,oneof=global individual group"`
	SendEmail   bool    `json:"send_email"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	Filters     *NotificationFilterDTO `json:"filters,omitempty"`
	TargetUserIDs []uuid.UUID `json:"target_user_ids,omitempty"`
}

type NotificationFilterDTO struct {
	Municipality string  `json:"municipality,omitempty"`
	State        string  `json:"state,omitempty"`
	Genre        string  `json:"genre,omitempty"`
	SpecialtyID  *uint32 `json:"specialty_id,omitempty"`
	Solvent      *bool   `json:"solvent,omitempty"`
}

// ── Preview ──────────────────────────────────────────────────────────────

type PreviewNotificationRequest struct {
	TargetType   string                   `json:"target_type" validate:"required,oneof=global individual group"`
	Filters      *NotificationFilterDTO   `json:"filters,omitempty"`
	TargetUserIDs []uuid.UUID             `json:"target_user_ids,omitempty"`
}
```

---

## 10. Filtro engine (resolución de destinatarios)

Se implementa como método del repository (`ResolveRecipients`, `ResolveAll`, `ResolveByIDs`), NO como archivo separado. La lógica es una query GORM condicional:

### 10.1 Lógica por `TargetType`

| `TargetType` | Comportamiento |
|---|---|
| `global` | `ResolveAll()` — solo base: activos + con fe de vida |
| `individual` | `ResolveByIDs()` — validar que los UUIDs existan y estén activos |
| `group` | `ResolveRecipients(params)` — filtros combinables sobre la base |

### 10.2 Query base (siempre activa)

```sql
WHERE is_active = true AND proof_of_life = true
```

### 10.3 Filtros condicionales

```go
if params.Municipality != "" {
    query = query.Where("municipality_carabobo = ?", params.Municipality)
}
if params.State != "" {
    query = query.Where("state_outside = ?", params.State)
}
if params.Genre != "" {
    query = query.Where("genre = ?", params.Genre)
}
if params.SpecialtyID != nil {
    query = query.Where(
        "primary_specialty_id = ? OR secondary_specialty_id = ?",
        *params.SpecialtyID, *params.SpecialtyID,
    )
}
if params.Solvent != nil {
    query = query.Where("solvent = ?", *params.Solvent)
}
```

### 10.4 Preview

El endpoint de preview ejecuta la misma resolución pero **no crea** registros. Solo retorna:

```json
{
    "success": true,
    "total_recipients": 145,
    "recipients": [
        {"id": "uuid", "name": "Juan Pérez", "email": "juan@..."}
    ]
}
```

---

## 11. Service — Flujo de creación/envío

**Archivo:** `internal/service/notification_service.go`

```go
type NotificationService struct {
    repo      domain.NotificationRepository
    s3Client  *s3.S3Client
    mailSvc   service.IMailService  // reutiliza la interfaz existente
}
```

### 11.1 Crear notificación inmediata

```
1. Validar permisos (admin.CanSendNotifications || admin.Sudo)
2. Validar campos obligatorios (title, message, target_type)
3. Crear registro en notifications (status = "pending")
4. Resolver destinatarios según target_type
5. Crear notification_filters (auditoría)
6. Crear notification_targets (uno por destinatario, is_read = false)
7. Si send_email = true → encolar emails via mailSvc.SendEmail()
8. Actualizar notification: status = "sent", sent_at = NOW()
9. Retornar respuesta con estadísticas
```

### 11.2 Crear notificación programada

```
1. Validar permisos
2. Validar campos + scheduled_at (debe ser futuro)
3. Crear registro en notifications (status = "pending", scheduled_at = fecha)
4. Crear notification_filters
5. Retornar confirmación (NO se resuelven destinatarios aún)
```

### 11.3 Cancelar notificación

```
1. Verificar que la notificación pertenece al admin (o tiene CanManageNotifications)
2. Verificar que status = "pending"
3. Actualizar status = "cancelled"
4. Eliminar notification_targets si existen
```

### 11.4 Marcar como leída (agremiado)

```
1. Verificar que el target pertenece al usuario (psi_user_id del token)
2. Actualizar is_read = true, read_at = NOW()
```

### 11.5 Integración con email

- Reutiliza `IMailService` existente (`SendEmail(to, subject, templateName, data)`)
- El envío es **opcional**: el admin elige al crear (`send_email: true/false`)
- Se crean templates HTML embebidos en `internal/templates/` (ej. `notification.html`)
- Los emails se encolan (fire-and-forget) — el worker existente se encarga del throttling
- Cada envío se registra individualmente en `notification_targets.email_sent`

---

## 12. Worker de envío programado

### 12.1 Implementación

Se agrega como goroutine en `cmd/api/main.go`, siguiendo el patrón de `runABSSyncLoop` y el ticker de `PublishScheduled`:

```go
// En cmd/api/main.go, dentro del bloque de goroutines background:

// ── Worker de notificaciones programadas ────────────────────────────────
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-bgCtx.Done():
            return
        case <-ticker.C:
            notifCtx, notifCancel := context.WithTimeout(bgCtx, 60*time.Second)
            notificationSvc.ProcessScheduled(notifCtx)
            notifCancel()
        }
    }
}()
```

### 12.2 Lógica del worker

```go
func (s *NotificationService) ProcessScheduled(ctx context.Context) {
    // 1. Buscar notificaciones con status = "pending" AND scheduled_at <= now
    // 2. Para cada una:
    //    a. Resolver destinatarios (misma lógica que envío inmediato)
    //    b. Crear notification_targets
    //    c. Si send_email → encolar emails
    //    d. Actualizar status = "sent", sent_at = now
    // 3. Si falla → status = "failed"
}
```

---

## 13. Archivos a crear

| # | Archivo | Descripción |
|---|---|---|
| 1 | `internal/domain/notification.model.go` | Modelos: `Notification`, `NotificationTarget`, `NotificationFilter`, `NotificationAttach` + tipos/constantes |
| 2 | `internal/domain/notification_repository.go` | Interfaz `NotificationRepository` + struct `PsiUserInfo` |
| 3 | `internal/repository/postgres/notification_repo.go` | Implementación GORM de `NotificationRepository` |
| 4 | `internal/service/notification_service.go` | Lógica de negocio: crear, enviar, cancelar, marcar leída, process scheduled, preview |
| 5 | `internal/request_structs/request_notifications.go` | DTOs: `CreateNotificationRequest`, `PreviewNotificationRequest`, `NotificationFilterDTO` |
| 6 | `internal/handler/notification_handler.go` | Handlers HTTP para admin y agremiado |
| 7 | `internal/router/notification_router.go` | Registro de rutas `SetupNotificationRoutes()` |
| 8 | `internal/templates/notification.html` | Template HTML para el correo de notificación |
| 9 | `migrations/YYYYMMDDHHMMSS_add_notifications.sql` | Migración Atlas: tablas `notifications`, `notification_targets`, `notification_filters`, `notification_attachments` |

---

## 14. Archivos a modificar

| # | Archivo | Cambio |
|---|---|---|
| 1 | `internal/router/router.go` | Importar y llamar `SetupNotificationRoutes(api, adminRepo, psiRepo, s3Client, analyticsSvc, mailSvc)` |
| 2 | `internal/domain/errors.go` | Agregar sentinelas: `ErrNotificationNotFound`, `ErrNotificationCannotCancel`, `ErrNotificationAlreadySent`, `ErrNotificationPermDenied` |
| 3 | `cmd/api/main.go` | Instanciar `NotificationService` y lanzar goroutine del scheduler de notificaciones programadas |

---

## 15. Migración Atlas

**Archivo:** `migrations/YYYYMMDDHHMMSS_add_notifications.sql`

```sql
-- Notificaciones in-app
CREATE TABLE IF NOT EXISTS notifications (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    create_by varchar(255),
    update_by varchar(255),
    create_by_id uuid,
    update_by_id uuid,
    title varchar(255) NOT NULL,
    message text NOT NULL,
    target_type varchar(20) NOT NULL,
    sender_id uuid NOT NULL REFERENCES user_admins(id),
    send_email boolean DEFAULT false,
    scheduled_at timestamptz,
    sent_at timestamptz,
    status varchar(20) NOT NULL DEFAULT 'pending'
);

CREATE INDEX idx_notifications_sender ON notifications(sender_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_scheduled ON notifications(status, scheduled_at)
    WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS notification_targets (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    notification_id uuid NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    psi_user_id uuid NOT NULL REFERENCES psi_users(id),
    is_read boolean DEFAULT false,
    read_at timestamptz,
    email_sent boolean DEFAULT false,
    email_sent_at timestamptz
);

CREATE INDEX idx_notification_targets_notif ON notification_targets(notification_id);
CREATE INDEX idx_notification_targets_user ON notification_targets(psi_user_id);

CREATE TABLE IF NOT EXISTS notification_filters (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    notification_id uuid NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    municipality varchar(255),
    state varchar(255),
    genre varchar(1),
    specialty_id integer,
    solvent boolean,
    target_user_ids text
);

CREATE INDEX idx_notification_filters_notif ON notification_filters(notification_id);

CREATE TABLE IF NOT EXISTS notification_attachments (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    create_by_id uuid,
    notification_id uuid NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    name varchar(255),
    s3_key varchar(512) NOT NULL,
    content_type varchar(100)
);

CREATE INDEX idx_notification_attachs_notif ON notification_attachments(notification_id);
```

---

## 16. Orden de implementación

| Paso | Tarea | Archivos | Dependencias |
|---|---|---|---|
| 1 | Modelos GORM + tipos/constantes | `domain/notification.model.go` | Ninguna |
| 2 | Sentinela de errores | `domain/errors.go` (modificar) | Paso 1 |
| 3 | Interfaz de repositorio | `domain/notification_repository.go` | Paso 1 |
| 4 | Request structs (DTOs) | `request_structs/request_notifications.go` | Paso 1 |
| 5 | Migración Atlas | `migrations/..._add_notifications.sql` | Paso 1 |
| 6 | Implementación repositorio | `repository/postgres/notification_repo.go` | Pasos 1, 3 |
| 7 | Servicio + filter engine | `service/notification_service.go` | Pasos 3, 4, 6 |
| 8 | Template de correo | `templates/notification.html` | Ninguna |
| 9 | Handlers HTTP | `handler/notification_handler.go` | Pasos 4, 7 |
| 10 | Router | `router/notification_router.go` | Paso 9 |
| 11 | Registro en router principal | `router/router.go` (modificar) | Paso 10 |
| 12 | Worker scheduler | `cmd/api/main.go` (modificar) | Paso 7 |

---

## 17. Seguridad y permisos

| Operación | Permiso requerido | Middleware |
|---|---|---|
| Crear notificación | `CanSendNotifications = true` | `ProtectedAdmin404()` |
| Cancelar notificación | `CanManageNotifications = true` | `ProtectedAdmin404()` |
| Ver stats/targets | `CanReadNotifications = true` | `ProtectedAdmin404()` |
| Subir adjunto | `CanSendNotifications = true` | `ProtectedAdmin404()` |
| Ver mis notificaciones | Autenticado (agremiado) | `ProtectedPsiUser()` |
| Marcar como leída | Solo el destinatario (validar `psi_user_id` del token) | `ProtectedPsiUser()` |

**Reglas adicionales:**
- Un admin solo puede cancelar notificaciones que **él mismo creó** (a menos que tenga `CanManageNotifications`)
- Al marcar como leída, se verifica que el `psi_user_id` del token coincida con el `PsiUserID` del target
- No se exponen los UUIDs de destinatarios a admins que no crearon la notificación
- Adjuntos se sirven con URL firmada de S3 (no descarga directa)

---

## 18. Ejemplo de request

### Crear notificación por grupo (inmediata)

```json
POST /api/v1/notifications/admin
Authorization: Bearer <admin_token>

{
    "title": "Convocatoria a Asamblea General",
    "message": "Se convoca a todos los agremiados a la asamblea general el día 15 de octubre de 2026.",
    "target_type": "group",
    "send_email": true,
    "filters": {
        "municipality": "Valencia",
        "solvent": true
    }
}
```

### Crear notificación programada

```json
POST /api/v1/notifications/admin
Authorization: Bearer <admin_token>

{
    "title": "Recordatorio de cuota",
    "message": "Recuerde que la cuota trimestral vence el próximo viernes.",
    "target_type": "global",
    "send_email": false,
    "scheduled_at": "2026-09-15T08:00:00-04:00"
}
```

### Preview de destinatarios

```json
POST /api/v1/notifications/admin/preview
Authorization: Bearer <admin_token>

{
    "target_type": "group",
    "filters": {
        "state": "Aragua",
        "genre": "F",
        "specialty_id": 3
    }
}
```

### Respuesta del preview

```json
{
    "success": true,
    "total_recipients": 23,
    "recipients": [
        {"id": "uuid-1", "name": "María García", "email": "maria@..."},
        {"id": "uuid-2", "name": "Ana López", "email": "ana@..."}
    ]
}
```

---

## 19. Consideraciones técnicas

### 19.1 Goroutines y concurrencia

- El scheduler corre como goroutine con `context.WithCancel` (patrón de `main.go`)
- El envío de emails reutiliza el `MailService` existente (cola buffer 5000, throttling, jittering)
- Las operaciones de DB usan transacciones cuando involucran múltiples writes

### 19.2 Rendimiento

- `ResolveRecipients` usa `Pluck("id", ...)` para traer solo UUIDs
- El conteo de no leídas se puede cachear con polling en el frontend
- Los adjuntos se sirven via URL firmada de S3 (sin pasar por el backend)

### 19.3 Manejo de errores

- Si el envío de email falla para un destinatario → `email_sent = false`, notificación sigue como "sent"
- Si la resolución de destinatarios falla → notificación = "failed"
- Los errores de SMTP se loguean pero no detienen el envío de la notificación in-app

### 19.4 Frontend (futuro)

- El cliente SolidStart necesitará:
  - Página de administración de notificaciones (`/admin/notificaciones/`)
  - Formulario de creación con preview
  - Vista de agremiado con lista y detalle (`/psi/notificaciones/`)
  - Badge/header con contador de no leídas
- Implementación futura, fuera del alcance de este plan (solo backend)

### 19.5 Diferencias con el plan original

| Aspecto | Plan original | Este plan |
|---|---|---|
| Arquitectura | `router → presenter → controller → db → mapper` | `router → handler → service → repository → DB` |
| Directorios | `Api/src/notifications/` | `internal/{domain,repository,service,handler,router}/` |
| Modelos | `Api/src/models/notification.go` | `internal/domain/notification.model.go` |
| Adjuntos | Bytea en DB de imágenes | S3/MinIO via `pkg/s3/` |
| Auth middleware | `ProtectedAdminWithDynamicKey` (no existe) | `ProtectedAdmin404()` existente |
| Migraciones | AutoMigrate | Atlas SQL |
| Email | `src/messages/mail/mail.go` | `IMailService` existente |
| Worker | Goroutine simple | `context.WithCancel` + ticker (patrón main.go) |
| Filtro specialty | String legacy | `specialty_id` (FK al catálogo) |

---

**[⬆ Volver a raíz](./README.md)**

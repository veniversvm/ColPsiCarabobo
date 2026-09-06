// api/internal/domain/notification.model.go
// Package domain contiene las entidades de negocio y las interfaces del sistema.
// Este paquete es el núcleo de Clean Architecture y no debe tener dependencias externas.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationTargetType define el alcance de una notificación.
type NotificationTargetType string

const (
	// NotificationTargetGlobal envía a todos los agremiados elegibles.
	NotificationTargetGlobal NotificationTargetType = "global"
	// NotificationTargetIndividual envía a uno o más agremiados específicos por UUID.
	NotificationTargetIndividual NotificationTargetType = "individual"
	// NotificationTargetGroup envía a un grupo definido por filtros geográficos/demográficos.
	NotificationTargetGroup NotificationTargetType = "group"
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
	Title      string                 `gorm:"size:255;not null" json:"title"`
	Message    string                 `gorm:"type:text;not null" json:"message"`
	TargetType NotificationTargetType `gorm:"size:20;not null" json:"target_type"`
	SenderID   uuid.UUID              `gorm:"type:uuid;not null" json:"sender_id"`
	SendEmail  bool                   `gorm:"default:false" json:"send_email"`
	ScheduledAt *time.Time            `gorm:"type:timestamptz" json:"scheduled_at,omitempty"`
	SentAt     *time.Time             `gorm:"type:timestamptz" json:"sent_at,omitempty"`
	Status     NotificationStatus     `gorm:"size:20;default:pending;not null" json:"status"`

	Targets []NotificationTarget `gorm:"foreignKey:NotificationID" json:"targets,omitempty"`
	Filters []NotificationFilter `gorm:"foreignKey:NotificationID" json:"filters,omitempty"`
	Attachs []NotificationAttach `gorm:"foreignKey:NotificationID" json:"attachments,omitempty"`
}

func (Notification) TableName() string { return "notifications" }

// NotificationTarget registra cada destinatario individual de una notificación.
type NotificationTarget struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	NotificationID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_notification_targets_notification_psi,priority:1" json:"notification_id"`
	PsiUserID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_notification_targets_notification_psi,priority:2;index:idx_notification_targets_psi_read,priority:1" json:"psi_user_id"`
	IsRead         bool       `gorm:"default:false;index:idx_notification_targets_psi_read,priority:2" json:"is_read"`
	ReadAt         *time.Time `gorm:"type:timestamptz" json:"read_at,omitempty"`
	EmailSent      bool       `gorm:"default:false" json:"email_sent"`
	EmailSentAt    *time.Time `gorm:"type:timestamptz" json:"email_sent_at,omitempty"`

	PsiUser *PsiUserModel `gorm:"foreignKey:PsiUserID" json:"psi_user,omitempty"`
}

func (NotificationTarget) TableName() string { return "notification_targets" }

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
	TargetUserIDs  string    `gorm:"type:text" json:"target_user_ids,omitempty"`
}

func (NotificationFilter) TableName() string { return "notification_filters" }

// NotificationAttach referencia un archivo adjunto almacenado en S3/MinIO.
type NotificationAttach struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel
	NotificationID uuid.UUID `gorm:"type:uuid;not null;index" json:"notification_id"`
	Name           string    `gorm:"size:255" json:"name"`
	S3Key          string    `gorm:"size:512;not null" json:"s3_key"`
	ContentType    string    `gorm:"size:100" json:"content_type"`
}

func (NotificationAttach) TableName() string { return "notification_attachments" }

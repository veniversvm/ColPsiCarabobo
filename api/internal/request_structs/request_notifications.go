// api/internal/request_structs/request_notifications.go

// Package request_structs define los DTOs (Data Transfer Objects) que actúan como
// la primera línea de defensa y validación de los datos que entran a la API.
package request_structs

import (
	"time"

	"github.com/google/uuid"
)

// =========================================================================
// MÓDULO DE NOTIFICACIONES
// =========================================================================

// NotificationFilterDTO contiene los filtros demográficos/geográficos
// opcionales para resolver destinatarios. Todos los campos son opcionales.
type NotificationFilterDTO struct {
	Municipality string  `json:"municipality,omitempty"`
	State        string  `json:"state,omitempty"`
	Genre        string  `json:"genre,omitempty"`
	SpecialtyID  *uint32 `json:"specialty_id,omitempty"`
	Solvent      *bool   `json:"solvent,omitempty"`
}

// CreateNotificationRequest define la carga útil para crear una notificación.
type CreateNotificationRequest struct {
	Title       string                 `json:"title" validate:"required,max=255"`
	Message     string                 `json:"message" validate:"required"`
	TargetType  string                 `json:"target_type" validate:"required,oneof=global individual group"`
	SendEmail   bool                   `json:"send_email"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Filters     *NotificationFilterDTO `json:"filters,omitempty"`
	TargetUserIDs []uuid.UUID          `json:"target_user_ids,omitempty"`
}

// PreviewNotificationRequest define la carga útil para previsualizar destinatarios.
type PreviewNotificationRequest struct {
	TargetType   string                 `json:"target_type" validate:"required,oneof=global individual group"`
	Filters      *NotificationFilterDTO `json:"filters,omitempty"`
	TargetUserIDs []uuid.UUID           `json:"target_user_ids,omitempty"`
}

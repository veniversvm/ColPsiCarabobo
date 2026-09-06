// api/internal/domain/app_setting.model.go
// Almacenamiento clave-valor (KV) de configuración global del sitio.
// Los valores se guardan como JSON (jsonb) para admitir distintos shapes
// según la clave (p.ej. Recepción de tickets vs inscripciones).
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Claves de configuración de recepción global (interruptores).
const (
	SettingsKeyTicketsReception      = "tickets.reception_enabled"
	SettingsKeyInscriptionsReception = "inscriptions.reception_enabled"
)

// ReceptionSetting es el valor de los interruptores de recepción global.
type ReceptionSetting struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message,omitempty"` // Razón pública mostrada al usuario
}

// AppSetting es una entrada del KV global de configuración.
type AppSetting struct {
	Key       string         `gorm:"primaryKey;size:80" json:"key"`
	Value     datatypes.JSON `gorm:"type:jsonb" json:"value"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// SettingsAuditLog es el registro inmutable de cambios de configuración global.
type SettingsAuditLog struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ChangedByID       uuid.UUID `gorm:"type:uuid;index" json:"changed_by_id"`
	ChangedByUsername string    `gorm:"size:100" json:"changed_by_username"`
	Key               string    `gorm:"size:80;index" json:"key"`
	EnabledFrom       bool      `json:"enabled_from"`
	EnabledTo         bool      `json:"enabled_to"`
	MessageFrom       string    `gorm:"size:500" json:"message_from"`
	MessageTo         string    `gorm:"size:500" json:"message_to"`
	CreatedAt         time.Time `json:"created_at"`
}

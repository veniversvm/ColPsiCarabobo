// api/internal/domain/audit_log.model.go
// Registro inmutable de cambios a nivel de API (bitácora de auditoría).
// Cada fila describe UN suceso: qué entidad cambió, qué acción, quién la hizo
// y el diff por campo { from, to }. Es de escritura diferida (cola + worker)
// para no interferir con el flujo de la petición, y de solo lectura para el
// staff autorizado (gates por can_view_logs / can_export_logs).
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Entidades auditables (consistencia entre servicios, endpoints y UI).
const (
	AuditEntityPsi         = "psi"
	AuditEntityStaff       = "staff"
	AuditEntityAuth        = "auth"
	AuditEntityNotification = "notificacion"
	AuditEntityTicket      = "ticket"
	AuditEntityPost        = "post"
	AuditEntityProject     = "proyecto"
	AuditEntityArea        = "area"
	AuditEntityInscription = "ficha_inscripcion"
	AuditEntityConfig      = "config"
)

// Acciones auditables.
const (
	AuditActionCreate        = "create"
	AuditActionUpdate        = "update"
	AuditActionDelete        = "delete"
	AuditActionLogin         = "login"
	AuditActionLogout        = "logout"
	AuditActionResetPassword = "reset_password"
	AuditActionTransferSudo  = "transfer_sudo"
	AuditActionChangeState   = "estado"
	AuditActionSend          = "send"
	AuditActionMove          = "move"
)

// AuditChange describe el cambio de un campo: de qué valor a qué valor.
// `from` se omite en el JSON cuando es nil (campo nuevo en la entidad).
type AuditChange struct {
	From any `json:"from,omitempty"`
	To   any `json:"to,omitempty"`
}

// ApiChangeLog es el registro forense de un cambio a nivel de API.
//
// Índices compuestos pensados para las consultas de búsqueda:
//   - idx_audit_entity_created: (entity, entity_id, created_at) → historial de
//     una entidad (p. ej. todo lo que le pasó a un psicólogo).
//   - idx_audit_actor_created:  (actor_id, created_at)     → "lo que hizo X".
//   - idx_audit_entity_action:  (entity, action, created_at) → agregados/stats.
type ApiChangeLog struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:uuidv7()" json:"id"`
	Entity        string         `gorm:"size:50;not null;index:idx_audit_entity_created,priority:1;index:idx_audit_entity_action,priority:1" json:"entity"`
	EntityID      string         `gorm:"size:64;index:idx_audit_entity_created,priority:2" json:"entity_id"`
	EntityLabel   string         `gorm:"size:255" json:"entity_label"`
	Action        string         `gorm:"size:50;index:idx_audit_entity_action,priority:2;index" json:"action"`
	ActorID       uuid.UUID      `gorm:"type:uuid;index:idx_audit_actor_created,priority:1" json:"actor_id"`
	ActorRole     string         `gorm:"size:50" json:"actor_role"`
	ActorUsername string         `gorm:"size:100" json:"actor_username"`
	IP            string         `gorm:"size:64" json:"ip"`
	UserAgent     string         `gorm:"size:255" json:"user_agent"`
	Changes       datatypes.JSON `gorm:"type:jsonb" json:"changes"`
	Metadata      datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
	CreatedAt     time.Time      `gorm:"index:idx_audit_entity_created,priority:3;index:idx_audit_actor_created,priority:2;index:idx_audit_entity_action,priority:3;index" json:"created_at"`
}

func (ApiChangeLog) TableName() string { return "api_change_logs" }
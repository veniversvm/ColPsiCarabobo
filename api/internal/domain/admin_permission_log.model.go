// api/internal/domain/admin_permission_log.model.go
// Registro inmutable de auditoría: toda creación de personal o cambio de
// permisos/sudo deja huella (quién, a quién, qué cambió y cuándo). Tabla
// dedicada (admin_permission_logs) para que el historial no se mezcle con
// la tabla mutable user_admins.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdminPermissionLog es el registro forense de cambios de permisos del staff.
type AdminPermissionLog struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	TargetAdminID      uuid.UUID `gorm:"type:uuid;index" json:"target_admin_id"`
	TargetUsername     string    `gorm:"size:100" json:"target_username"`
	Action             string    `gorm:"size:50" json:"action"` // create | update_permissions | role_change | transfer_sudo
	ChangedByID        uuid.UUID `gorm:"type:uuid;index" json:"changed_by_id"`
	ChangedByUsername  string    `gorm:"size:100" json:"changed_by_username"`
	PermissionsChanged string    `gorm:"type:text" json:"permissions_changed"`
	RoleFrom           string    `gorm:"size:50" json:"role_from"`
	RoleTo             string    `gorm:"size:50" json:"role_to"`
	CreatedAt          time.Time `json:"created_at"`
}
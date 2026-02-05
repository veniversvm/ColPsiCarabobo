// api/internal/handler/requests.structs.go
package request_structs

import (
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

type CreateAdminRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	// Permisos que se le quieren asignar
	Permissions domain.UserAdmin `json:"permissions"`
}

type UpdateAdminRequest struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	Username *string   `json:"username"`
	Email    *string   `json:"email"`
	Password *string   `json:"password"`
	IsActive *bool     `json:"is_active"`
	// Permisos
	CanCreatePsi           *bool `json:"can_create_psi"`
	CanUpdatePsi           *bool `json:"can_update_psi"`
	CanDeletePsi           *bool `json:"can_delete_psi"`
	CanCreateAdmin         *bool `json:"can_create_admin"`
	CanUpdateAdmin         *bool `json:"can_update_admin"`
	CanDeleteAdmin         *bool `json:"can_delete_admin"`
	CanPublish             *bool `json:"can_publish"`
	CanUpdatePublish       *bool `json:"can_update_publish"`
	CanDeletePublish       *bool `json:"can_delete_publish"`
	CanSendNotifications   *bool `json:"can_send_notifications"`
	CanManageNotifications *bool `json:"can_manage_notifications"`
	CanReadNotifications   *bool `json:"can_read_notifications"`
	CanCreateTags          *bool `json:"can_create_tags"`
	CanEditTags            *bool `json:"can_edit_tags"`
	CanDeleteTags          *bool `json:"can_delete_tags"`
}

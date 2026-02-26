// api/internal/request_structs/requests_admin.structs.go

// Package request_structs define los contratos de entrada (DTOs) de la API.
package request_structs

import (
	"github.com/google/uuid"
)

// AdminPermissionsDTO centraliza los flags de permisos granulares.
// Al usar punteros (*bool), este DTO es 100% compatible con operaciones PATCH,
// permitiendo al "Motor de Permisos" saber exactamente qué valores se enviaron explícitamente
// y cuáles deben ignorarse (nil).
type AdminPermissionsDTO struct {
	CanCreatePsi           *bool `json:"can_create_psi" example:"false"`
	CanUpdatePsi           *bool `json:"can_update_psi" example:"false"`
	CanDeletePsi           *bool `json:"can_delete_psi" example:"false"`
	CanCreateAdmin         *bool `json:"can_create_admin" example:"false"`
	CanUpdateAdmin         *bool `json:"can_update_admin" example:"false"`
	CanDeleteAdmin         *bool `json:"can_delete_admin" example:"false"`
	CanPublish             *bool `json:"can_publish" example:"true"`
	CanUpdatePublish       *bool `json:"can_update_publish" example:"true"`
	CanDeletePublish       *bool `json:"can_delete_publish" example:"false"`
	CanSendNotifications   *bool `json:"can_send_notifications" example:"true"`
	CanManageNotifications *bool `json:"can_manage_notifications" example:"false"`
	CanReadNotifications   *bool `json:"can_read_notifications" example:"true"`
	CanCreateTags          *bool `json:"can_create_tags" example:"false"`
	CanEditTags            *bool `json:"can_edit_tags" example:"false"`
	CanDeleteTags          *bool `json:"can_delete_tags" example:"false"`
}

// CreateAdminRequest es la carga útil para registrar un nuevo miembro del staff.
// Valida estrictamente los campos obligatorios antes de tocar la base de datos.
type CreateAdminRequest struct {
	Username    string              `json:"username" validate:"required" example:"staff_admin"`
	Email       string              `json:"email" validate:"required,email" example:"staff@colpsicarabobo.com"`
	Password    string              `json:"password" validate:"required,min=8" example:"Segura123!"`
	Permissions AdminPermissionsDTO `json:"permissions"`
}

// UpdateAdminRequest define los campos permitidos para la edición de un administrador.
// Todos los campos de identidad son punteros para soportar actualizaciones parciales.
type UpdateAdminRequest struct {
	// ID es requerido para identificar al administrador a actualizar.
	// Arquitectura: Se puede usar como doble validación si también se envía por la URL (/admin/:id).
	ID uuid.UUID `json:"id" validate:"required"`

	Username *string `json:"username" example:"nuevo_username"`
	// omitempty permite que si el campo viene vacío no falle la validación de formato 'email'
	Email    *string `json:"email" validate:"omitempty,email" example:"nuevo@correo.com"`
	Password *string `json:"password" validate:"omitempty,min=8"`
	IsActive *bool   `json:"is_active" example:"true"`

	Permissions AdminPermissionsDTO `json:"permissions"`
}

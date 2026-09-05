// api/internal/request_structs/requests_admin.structs.go

// Package request_structs define los contratos de entrada (DTOs) de la API.
//
// Este archivo en particular gestiona la transferencia de datos para el módulo
// de Staff (Administradores), actuando como barrera de validación para
// operaciones altamente sensibles como el escalamiento de privilegios.
package request_structs

import (
	"github.com/google/uuid"
)

// AdminPermissionsDTO centraliza la matriz de control de acceso (ACL) mediante permisos granulares.
//
// Diseño de Seguridad e Inmutabilidad en PATCH:
// Al emplear punteros (*bool), este DTO resuelve un problema crítico en Go:
// El "Motor de Permisos" en la capa de Dominio puede distinguir perfectamente entre:
//   - nil: El cliente no envió este campo (el permiso actual NO debe ser alterado).
//   - *true / *false: El administrador maestro desea mutar explícitamente este permiso.
//
// Si se usaran booleanos primitivos (bool), un campo no enviado asumiría el valor "zero" (false),
// lo que resultaría en la revocación accidental de permisos durante una actualización parcial.
type AdminPermissionsDTO struct {
	// Gestión de Colegiados (CRUD de Psicólogos)
	CanCreatePsi *bool `json:"can_create_psi" example:"false"`
	CanUpdatePsi *bool `json:"can_update_psi" example:"false"`
	CanDeletePsi *bool `json:"can_delete_psi" example:"false"`

	// Gestión del Staff (Escalamiento y Delegación)
	CanCreateAdmin *bool `json:"can_create_admin" example:"false"`
	CanUpdateAdmin *bool `json:"can_update_admin" example:"false"`
	CanDeleteAdmin *bool `json:"can_delete_admin" example:"false"`

	// Sistema de Gestión de Contenidos (CMS)
	CanPublish       *bool `json:"can_publish" example:"true"`
	CanUpdatePublish *bool `json:"can_update_publish" example:"true"`
	CanDeletePublish *bool `json:"can_delete_publish" example:"false"`

	// Motor de Mensajería y Alertas
	CanSendNotifications   *bool `json:"can_send_notifications" example:"true"`
	CanManageNotifications *bool `json:"can_manage_notifications" example:"false"`
	CanReadNotifications   *bool `json:"can_read_notifications" example:"true"`

	// Taxonomías (Clasificación de contenido)
	CanCreateTags *bool `json:"can_create_tags" example:"false"`
	CanEditTags   *bool `json:"can_edit_tags" example:"false"`
	CanDeleteTags *bool `json:"can_delete_tags" example:"false"`

	// Proyectos (Kanban)
	CanManageProjects *bool `json:"can_manage_projects" example:"false"`
}

// CreateAdminRequest define la carga útil (Payload) para el aprovisionamiento
// de un nuevo miembro del Staff.
//
// Es un contrato estricto: exige credenciales válidas y seguras antes de permitir
// que la petición alcance la capa de infraestructura (hashing de contraseñas y base de datos).
// Por defecto, cualquier permiso no incluido en el JSON interno 'permissions' será
// inicializado de forma segura restringiendo el acceso (Principio de Menor Privilegio).
type CreateAdminRequest struct {
	Username    string              `json:"username" validate:"required" example:"staff_admin"`
	Email       string              `json:"email" validate:"required,email" example:"staff@colpsicarabobo.com"`
	Password    string              `json:"password" validate:"required,min=8" example:"Segura123!"`
	Permissions AdminPermissionsDTO `json:"permissions"`
}

// UpdateAdminRequest define el contrato para mutaciones de perfil de staff
// y ajustes en su matriz de privilegios.
//
// Validación Condicional Segura (Semántica PATCH):
// El uso del tag 'omitempty' acoplado a punteros le indica al validador (validator/v10)
// que reglas estrictas como 'email' o 'min=8' SOLO deben evaluarse si el campo fue
// efectivamente enviado en la petición (puntero != nil).
// Esto hace que la API sea verdaderamente RESTful, permitiendo actualizar únicamente
// el estado 'is_active' de un usuario sin obligar al cliente a reenviar su contraseña.
type UpdateAdminRequest struct {
	// ID es el identificador primario del administrador a afectar.
	// Nota Arquitectónica: En controladores robustos, este ID en el body debe
	// ser verificado cruzándolo con el ID de la URL (/api/admin/:id) para
	// evitar ataques de manipulación de Payload.
	ID uuid.UUID `json:"id" validate:"required"`

	Username *string `json:"username" example:"nuevo_username"`
	Email    *string `json:"email" validate:"omitempty,email" example:"nuevo@correo.com"`
	Password *string `json:"password" validate:"omitempty,min=8"`
	IsActive *bool   `json:"is_active" example:"true"`

	// Sub-estructura que será evaluada por el Motor de Permisos.
	Permissions AdminPermissionsDTO `json:"permissions"`
}

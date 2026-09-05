// api/internal/request_structs/kanban_requests.go

// Package request_structs define los contratos de entrada (DTOs) de la API.
//
// Este archivo gestiona los contratos para el módulo de Proyectos (Kanban)
// del panel administrativo.
package request_structs

import (
	"github.com/google/uuid"
)

// CreateProjectRequest define la carga útil para crear un nuevo proyecto.
type CreateProjectRequest struct {
	// Name es el título visible del proyecto.
	Name string `json:"name" example:"Convención Anual 2026" validate:"required,max=120"`
	// Description es opcional; contexto breve del proyecto.
	Description string `json:"description" example:"Organización del evento presencial" validate:"max=500"`
}

// UpdateProjectRequest facilita la mutación parcial de un proyecto.
type UpdateProjectRequest struct {
	Name        *string `json:"name" example:"Convención 2026"`
	Description *string `json:"description" example:"Nueva descripción"`
}

// CreateColumnRequest define la carga útil para crear una columna del tablero.
type CreateColumnRequest struct {
	Title string `json:"title" example:"Por hacer" validate:"required,max=120"`
}

// UpdateColumnRequest facilita la mutación parcial de una columna.
type UpdateColumnRequest struct {
	Title    *string `json:"title" example:"En revisión"`
	Position *int    `json:"position" example:"1"`
}

// CreateCardRequest define la carga útil para crear una tarjeta Kanban.
type CreateCardRequest struct {
	ColumnID    uuid.UUID `json:"column_id" validate:"required"`
	Title       string    `json:"title" example:"Diseñar afiche" validate:"required,max=200"`
	Description string    `json:"description" example:"Incluir logo y QR del formulario" validate:"max=2000"`
}

// UpdateCardRequest facilita la mutación parcial de una tarjeta.
// Si ColumnID cambia, la tarjeta se mueve de lista.
type UpdateCardRequest struct {
	ColumnID    *uuid.UUID `json:"column_id"`
	Title       *string    `json:"title" validate:"omitempty,max=200"`
	Description *string    `json:"description" validate:"omitempty,max=2000"`
	Position    *int       `json:"position" example:"0"`
}

// AddMemberRequest define la carga útil para añadir un miembro al proyecto.
type AddMemberRequest struct {
	UserAdminID uuid.UUID `json:"user_admin_id" validate:"required"`
	Role        string    `json:"role" example:"editor"`
}

// UpdateMemberRequest facilita la mutación del rol de un miembro.
type UpdateMemberRequest struct {
	Role *string `json:"role" example:"viewer"`
}

// CreateNoteRequest define la carga útil para añadir una nota a una tarjeta.
type CreateNoteRequest struct {
	Content string `json:"content" example:"Pendiente de aprobación de la junta" validate:"required,max=500"`
}

// api/internal/request_structs/ticket_requests.go

// Package request_structs define los contratos de entrada (DTOs) de la API.
//
// Este archivo gestiona los contratos del módulo de Tickets de Solicitudes
// (portal psi y panel administrativo).
package request_structs

// CreateTicketMotivoRequest define la carga útil para crear un motivo. Al
// crearlo, el sistema siembra automáticamente los estados por defecto
// (recibido → en proceso → cerrado). tickets_per_psi es el límite de tickets
// abiertos que un psicólogo puede tener para este motivo.
type CreateTicketMotivoRequest struct {
	Name          string `json:"name" example:"Solicitudes" validate:"required,max=120"`
	Description   string `json:"description" validate:"max=500"`
	TicketsPerPsi int    `json:"tickets_per_psi" example:"3" validate:"required,min=1"`
}

// UpdateTicketMotivoRequest facilita la mutación parcial de un motivo.
type UpdateTicketMotivoRequest struct {
	Name          *string `json:"name" example:"Solicitudes"`
	Description   *string `json:"description" example:"Nueva descripción"`
	TicketsPerPsi *int    `json:"tickets_per_psi" example:"5"`
}

// CreateTicketEstadoRequest define la carga útil para crear un estado de un motivo.
type CreateTicketEstadoRequest struct {
	MotivoID uint   `json:"motivo_id" validate:"required"`
	Name     string `json:"name" example:"En revisión" validate:"required,max=60"`
	Order    int    `json:"order" example:"2"`
	IsClosed bool   `json:"is_closed" example:"false"`
}

// UpdateTicketEstadoRequest facilita la mutación parcial de un estado.
type UpdateTicketEstadoRequest struct {
	Name     *string `json:"name" example:"En revisión"`
	Order    *int    `json:"order" example:"2"`
	IsClosed *bool   `json:"is_closed" example:"false"`
}

// CreateTicketRequest define la carga útil para que el psicólogo abra un ticket.
// Todo ticket requiere un motivo (solo motivo, sin área). El límite de tickets
// abiertos por psicólogo depende del motivo (tickets_per_psi). Los anexos
// iniciales opcionales viajan como multipart (campo "files").
type CreateTicketRequest struct {
	MotivoID    uint   `json:"motivo_id" validate:"required"`
	Title       string `json:"title" example:"Solicitud de constancia de solvencia" validate:"required,max=200"`
	Description string `json:"description" example:"Necesito la constancia para mi empleador" validate:"required,max=2000"`
}

// AddMensajeRequest define la carga útil para publicar un comentario en la
// conversación interna del ticket. El límite de caracteres se valida en el
// servicio según el autor (1000 para el psicólogo). Los anexos opcionales
// viajan como multipart (campo "files").
type AddMensajeRequest struct {
	Message string `json:"message" example:"Adjunto el recibo de pago" validate:"required"`
}

// UpdateTicketEstadoRequestAdmin define el contrato para cambiar el estado de
// un ticket desde el panel administrativo.
type UpdateTicketEstado struct {
	EstadoID uint   `json:"estado_id" validate:"required"`
	Reason   string `json:"reason" example:"El comprobante fue validado" validate:"max=500"`
}

// CloseTicketRequest define el contrato para cerrar un ticket. El motivo de
// cierre es obligatorio (tanto para el admin como para el psicólogo).
type CloseTicketRequest struct {
	CloseReason string `json:"close_reason" example:"Respondido y resuelto" validate:"required,max=500"`
}
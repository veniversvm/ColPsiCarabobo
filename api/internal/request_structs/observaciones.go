// api/internal/request_structs/observaciones.go

// Package request_structs contiene las definiciones de los objetos de transferencia de datos.
//
// Este archivo gestiona el contrato de entrada del submódulo de Observaciones Internas.
// Las notas son observaciones internas del Colegio sobre un psicólogo y son de acceso
// exclusivo al personal administrativo autorizado.
package request_structs

// CreateObservacionesRequest define la carga útil (Payload) necesaria para registrar
// una nueva observación interna sobre un psicólogo.
type CreateObservacionesRequest struct {
	// Content es el texto libre de la observación.
	// Es texto plano; la capa de servicio lo sanitiza con bluemonday antes de persistir.
	Content string `json:"content" validate:"required" example:"Nota interna de seguimiento."`
}

// UpdateObservacionesRequest define la carga útil (Payload) necesaria para editar una
// observación interna existente. Usa un puntero para soportar semántica PATCH:
// solo se actualiza el campo si viene presente en el cuerpo.
type UpdateObservacionesRequest struct {
	// Content es el nuevo texto de la observación (texto plano; se sanitiza en la API).
	Content *string `json:"content" example:"Nota interna actualizada."`
}

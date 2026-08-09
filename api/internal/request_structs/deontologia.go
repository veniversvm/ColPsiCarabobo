// api/internal/request_structs/deontologia.go

// Package request_structs contiene las definiciones de los objetos de transferencia de datos.
//
// Este archivo gestiona el contrato de entrada del submódulo de Expediente Deontológico.
// Las entradas son notas internas del Colegio sobre un psicólogo y son de acceso
// exclusivo al personal administrativo autorizado.
package request_structs

// CreateDeontologiaRequest define la carga útil (Payload) necesaria para registrar
// una nueva entrada deontológica sobre un psicólogo.
type CreateDeontologiaRequest struct {
	// Content es el texto libre de la entrada (ej: descripción del expediente disciplinario).
	// Es texto plano; la capa de servicio lo sanitiza con bluemonday antes de persistir.
	Content string `json:"content" validate:"required" example:"Expediente abierto por el Tribunal Disciplinario."`
}

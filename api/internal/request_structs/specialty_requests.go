// api/internal/request_structs/specialty_requests.go

// Package request_structs define los modelos de entrada para la API,
// desacoplando la lógica de transporte de los modelos de persistencia.
package request_structs

// CreateSpecialtyRequest define la carga útil necesaria para registrar una nueva
// especialidad en el catálogo maestro del Colegio.
type CreateSpecialtyRequest struct {
	// Name es el nombre único de la especialidad profesional (ej. Neuropsicología).
	// Se valida como requerido para garantizar la integridad del catálogo.
	Name string `json:"name" example:"Psicología Clínica" validate:"required"`

	// Description proporciona detalles adicionales sobre el alcance de la especialidad.
	Description string `json:"description" example:"Rama de la psicología que se encarga de la investigación de todos los factores, evaluación, diagnóstico, tratamiento y prevención..."`
}

// UpdateSpecialtyRequest facilita la modificación parcial de una especialidad existente.
// Arquitectura Senior: El uso de punteros permite implementar una semántica PATCH pura.
// Si un campo es 'nil', el servicio lo ignorará; si es un puntero a un string vacío,
// se interpretará como una instrucción de limpieza del campo.
type UpdateSpecialtyRequest struct {
	// Name es opcional. Si se proporciona, debe cumplir con la restricción de unicidad en la DB.
	Name *string `json:"name" example:"Psicología Organizacional"`

	// Description es opcional. Permite actualizar la definición de la especialidad.
	Description *string `json:"description" example:"Nueva descripción actualizada..."`

	// Active permite habilitar o deshabilitar la especialidad del directorio público.
	// Al ser *bool, evitamos que un valor por defecto (false) desactive el registro por error.
	Active *bool `json:"active" example:"true"`
}

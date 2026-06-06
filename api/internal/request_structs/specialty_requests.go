// api/internal/request_structs/specialty_requests.go

// Package request_structs define los modelos de entrada (DTOs) para la API,
// desacoplando la lógica de transporte HTTP de los modelos de persistencia.
//
// Este archivo gestiona los contratos para el Catálogo Maestro de Especialidades
// (Áreas de Desempeño Clínico), garantizando que las taxonomías del sistema
// se mantengan estandarizadas y controladas.
package request_structs

// CreateSpecialtyRequest define la carga útil (Payload) necesaria para registrar
// una nueva especialidad en el Catálogo Maestro del Colegio.
//
// Al centralizar la creación de especialidades mediante este DTO estructurado,
// se previene que los psicólogos introduzcan texto libre (evitando duplicados
// como "Psicologia clinica" y "Psicología Clínica"), manteniendo la integridad
// del motor de búsqueda del directorio.
type CreateSpecialtyRequest struct {
	// Name es el identificador único y formal de la especialidad profesional.
	// Se valida como requerido ('validate:"required"') para garantizar que
	// ningún registro huérfano o anónimo ingrese a la base de datos.
	Name string `json:"name" example:"Psicología Clínica" validate:"required"`

	// Description proporciona detalles adicionales sobre el alcance de la especialidad.
	// Útil para mostrar "Tooltips" o tarjetas informativas a los pacientes
	// que navegan por el directorio público buscando orientación.
	Description string `json:"description" example:"Rama de la psicología que se encarga de la investigación, evaluación, diagnóstico y tratamiento..."`
}

// UpdateSpecialtyRequest facilita la mutación parcial de una especialidad existente.
//
// Semántica PATCH Pura (Arquitectura Senior):
// El uso extensivo de punteros (*string, *bool) resuelve el problema de los "Zero Values" en Go.
// Permite al motor de actualización distinguir matemáticamente entre tres escenarios:
//  1. El campo no se envió en el JSON (puntero nil) -> El servicio ignora la columna.
//  2. El campo se envió con valor (ej. "Nueva Info") -> El servicio actualiza el texto.
//  3. El campo se envió explícitamente vacío ("") -> El servicio borra el contenido de la DB.
type UpdateSpecialtyRequest struct {
	// Name es opcional. Si se proporciona, el controlador/servicio debe validar
	// que el nuevo nombre no colisione con otra especialidad existente (Restricción UNIQUE).
	Name *string `json:"name" example:"Psicología Organizacional"`

	// Description es opcional. Permite redefinir o corregir la explicación del área clínica.
	Description *string `json:"description" example:"Nueva descripción actualizada..."`

	// Active funciona como un interruptor de visibilidad (Soft-Disable).
	// Permite ocultar una especialidad deprecada del formulario de registro público
	// sin borrarla, preservando la integridad referencial (Foreign Keys) de los
	// psicólogos antiguos que ya la tienen asignada.
	// Al ser *bool, evitamos que la ausencia del campo en el payload la desactive por accidente.
	Active *bool `json:"active" example:"true"`
}

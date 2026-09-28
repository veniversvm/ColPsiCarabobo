// api/internal/request_structs/emergency_contact.go

// Package request_structs contiene las definiciones de los objetos de transferencia de datos.
//
// Este archivo define los contratos del submódulo de Persona de Contacto para
// Emergencias. Son datos personales de un TERCERO (un familiar, pareja o
// colleague del agremiado) que el Colegio necesita para poder avisar cuando el
// profesional no puede ser localizado o sufre un accidente; por eso su
// validación es estricta: siempre se debe saber quién es y cómo contactarlo.
package request_structs

// CreateEmergencyContactRequest define la carga útil para registrar una persona de
// contacto para emergencias.
//
// Regla de negocio (obligatoria y verificada también en el service):
//   - Name y Relationship siempre;
//   - al menos UNO de Phone o Email.
//
// Se combinan `validate:"required"` (campos sin excepción) con una comprobación
// cruzada de canal de contacto en la capa de servicio, porque el estándar
// validator no puede expresar reglas entre campos ("A o B").
type CreateEmergencyContactRequest struct {
	// Name identifica a la persona a la que se avisaría (ej: "María Rodríguez").
	Name string `json:"name" validate:"required,min=2,max=255" example:"María Rodríguez"`

	// Relationship es el parentesco o vínculo con el agremiado. El frontend
	// ofrece un catálogo cerrado (padre, madre, conyuge, hermano, hijo, tio,
	// primo, amigo, colega) y la opción "otro" en texto libre; el backend solo
	// exige que no venga vacío.
	Relationship string `json:"relationship" validate:"required,min=2,max=100" example:"madre"`

	// Phone es el teléfono fijo o celular de la persona de contacto (opcional
	// si se informa Email).
	Phone string `json:"phone" validate:"omitempty,max=20" example:"+58 412 1234567"`

	// Email es el correo de la persona de contacto (opcional si se informa Phone).
	Email string `json:"email" validate:"omitempty,max=255" example:"maria.rodriguez@correo.com"`
}

// UpdateEmergencyContactRequest permite la edición parcial (PATCH) de un contacto.
//
// Semántica PATCH real: punteros para distinguir "no enviado" (nil -> se ignora)
// de "enviado vacío" (puntero no nil -> se borra el valor). El service vuelve a
// evaluar la regla de integridad sobre el resultado combinado, de modo que un
// PATCH no puede dejar el registro sin canal de contacto.
type UpdateEmergencyContactRequest struct {
	Name         *string `json:"name" example:"María Rodríguez"`
	Relationship *string `json:"relationship" example:"madre"`
	Phone        *string `json:"phone" example:"+58 412 1234567"`
	Email        *string `json:"email" example:"maria.rodriguez@correo.com"`
}

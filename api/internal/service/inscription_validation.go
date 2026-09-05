// api/internal/service/inscription_validation.go
//
// Validación de los campos obligatorios de la ficha de inscripción.
// Comparte una única regla entre el formulario público (submit) y la
// edición administrativa (PATCH): personales, académicos y al menos un
// bloque de ubicación completo.
package service

import "strings"

// ValidationError representa un error de validación de campos de la ficha.
// Se mapea a HTTP 400 en el handler.
type ValidationError struct {
	Msg string
}

// Error devuelve el mensaje legible de la validación.
func (e *ValidationError) Error() string { return e.Msg }

// FichaObligatoria agrupa los campos que deben venir completos en una ficha.
// Las fechas se indican por presencia (presente=true) porque cada request
// las transporta con un tipo distinto (*time.Time o *string).
// La mención, el número de registro, el tomo y el folio del título son
// opcionales: a veces no están indicados en el título.
type FichaObligatoria struct {
	SegundoApellido         string
	Genero                  string
	Telefono                string
	FechaNacimientoPresente bool
	TituloUniversidad       string
	FechaGraduacionPresente bool
	TituloRegistroEstado    string
	ServiceAddress          string
	MunicipalityCarabobo    string
	StateOutside            string
	MunicipalityOutside     string
	Country                 string
}

// locationBlockComplete indica si al menos un bloque de ubicación está
// completo: Carabobo (municipio + dirección), otro estado (estado + municipio)
// o exterior (país).
func locationBlockComplete(f FichaObligatoria) bool {
	carabobo := strings.TrimSpace(f.MunicipalityCarabobo) != "" && strings.TrimSpace(f.ServiceAddress) != ""
	otroEstado := strings.TrimSpace(f.StateOutside) != "" && strings.TrimSpace(f.MunicipalityOutside) != ""
	exterior := strings.TrimSpace(f.Country) != ""
	return carabobo || otroEstado || exterior
}

// ValidateFichaObligatoria devuelve un *ValidationError con el primer campo
// obligatorio que falta, o nil si la ficha está completa.
func ValidateFichaObligatoria(f FichaObligatoria) error {
	switch {
	case strings.TrimSpace(f.SegundoApellido) == "":
		return &ValidationError{Msg: "el segundo apellido es obligatorio"}
	case strings.TrimSpace(f.Genero) == "":
		return &ValidationError{Msg: "el género es obligatorio"}
	case strings.TrimSpace(f.Telefono) == "":
		return &ValidationError{Msg: "el teléfono de contacto es obligatorio"}
	case !f.FechaNacimientoPresente:
		return &ValidationError{Msg: "la fecha de nacimiento es obligatoria"}
	case strings.TrimSpace(f.TituloUniversidad) == "":
		return &ValidationError{Msg: "la universidad es obligatoria"}
	case !f.FechaGraduacionPresente:
		return &ValidationError{Msg: "la fecha de graduación es obligatoria"}
	case strings.TrimSpace(f.TituloRegistroEstado) == "":
		return &ValidationError{Msg: "el estado del registro es obligatorio"}
	case !locationBlockComplete(f):
		return &ValidationError{Msg: "debes completar al menos una ubicación completa (Carabobo, otro estado o exterior)"}
	}
	return nil
}

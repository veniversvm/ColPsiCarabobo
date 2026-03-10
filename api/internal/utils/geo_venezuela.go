// api/internal/utils/geo_venezuela.go
package utils

import "strings"

// ─────────────────────────────────────────────────────────────────────────────
// MUNICIPIOS DE CARABOBO
// ─────────────────────────────────────────────────────────────────────────────

// municipiosCarabobo es el catálogo oficial de los 14 municipios del estado Carabobo.
var municipiosCarabobo = []string{
	"Bejuma",
	"Carlos Arvelo",
	"Diego Ibarra",
	"Guacara",
	"Juan José Mora",
	"Libertador",
	"Los Guayos",
	"Miranda",
	"Montalbán",
	"Naguanagua",
	"Puerto Cabello",
	"San Diego",
	"San Joaquín",
	"Valencia",
}

// NormalizeMunicipioCarabobo valida que el municipio pertenezca al estado Carabobo.
// Es indiferente a mayúsculas/minúsculas y retorna el nombre normalizado del catálogo.
// Retorna ("", false) si no encuentra coincidencia.
func NormalizeMunicipioCarabobo(input string) (string, bool) {
	normalized := strings.TrimSpace(input)
	for _, m := range municipiosCarabobo {
		if strings.EqualFold(normalized, m) {
			return m, true
		}
	}
	return "", false
}

// ─────────────────────────────────────────────────────────────────────────────
// ESTADOS DE VENEZUELA (EXCLUYENDO CARABOBO)
// ─────────────────────────────────────────────────────────────────────────────

// estadosVenezuela es el catálogo de los 23 estados de Venezuela más el
// Distrito Capital, excluyendo Carabobo (jurisdicción propia del Colegio).
var estadosVenezuela = []string{
	"Amazonas",
	"Anzoátegui",
	"Apure",
	"Aragua",
	"Barinas",
	"Bolívar",
	"Cojedes",
	"Delta Amacuro",
	"Dependencias Federales",
	"Distrito Capital",
	"Falcón",
	"Guárico",
	"Lara",
	"La Guaira",
	"Mérida",
	"Miranda",
	"Monagas",
	"Nueva Esparta",
	"Portuguesa",
	"Sucre",
	"Táchira",
	"Trujillo",
	"Yaracuy",
	"Zulia",
}

// NormalizeEstadoVenezuela valida que el estado pertenezca a Venezuela y no sea Carabobo.
// Es indiferente a mayúsculas/minúsculas y retorna el nombre normalizado del catálogo.
// Retorna ("", false) si no encuentra coincidencia o si el input es "Carabobo".
func NormalizeEstadoVenezuela(input string) (string, bool) {
	normalized := strings.TrimSpace(input)
	for _, e := range estadosVenezuela {
		if strings.EqualFold(normalized, e) {
			return e, true
		}
	}
	return "", false
}

// BoolFromForm convierte un valor de campo multipart/form-data a *bool.
// Retorna nil si el string está vacío (campo no enviado).
// Acepta: "1", "true", "yes" → true | "0", "false", "no" → false
func BoolFromForm(s string) *bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		v := true
		return &v
	case "0", "false", "no":
		v := false
		return &v
	default:
		return nil
	}
}

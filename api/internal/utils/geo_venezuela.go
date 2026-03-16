// api/internal/utils/geo_venezuela.go
package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// removeDiacritics elimina tildes y diacríticos para comparación flexible.
func removeDiacritics(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// foldCompare compara dos strings ignorando mayúsculas/minúsculas y tildes.
func foldCompare(a, b string) bool {
	return strings.EqualFold(removeDiacritics(a), removeDiacritics(b))
}

// ─────────────────────────────────────────────────────────────────────────────
// MUNICIPIOS DE CARABOBO
// ─────────────────────────────────────────────────────────────────────────────

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

func NormalizeMunicipioCarabobo(input string) (string, bool) {
	normalized := strings.TrimSpace(input)
	for _, m := range municipiosCarabobo {
		if foldCompare(normalized, m) {
			return m, true
		}
	}
	return "", false
}

// ─────────────────────────────────────────────────────────────────────────────
// ESTADOS DE VENEZUELA (EXCLUYENDO CARABOBO)
// ─────────────────────────────────────────────────────────────────────────────

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

func NormalizeEstadoVenezuela(input string) (string, bool) {
	normalized := strings.TrimSpace(input)
	for _, e := range estadosVenezuela {
		if foldCompare(normalized, e) {
			return e, true
		}
	}
	return "", false
}

// ─────────────────────────────────────────────────────────────────────────────
// BOOL FROM FORM
// ─────────────────────────────────────────────────────────────────────────────

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

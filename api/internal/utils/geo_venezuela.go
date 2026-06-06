// api/internal/utils/geo_venezuela.go

// Package utils provee herramientas de propósito general para la aplicación.
//
// Este archivo concentra la lógica de Normalización de Datos y Master Data Management (MDM)
// para la geografía venezolana. Actúa como un embudo de calidad (Data Quality) que toma
// entradas de texto libre (caóticas) introducidas por los usuarios y las estandariza
// antes de que toquen la base de datos.
package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// removeDiacritics elimina tildes y marcas diacríticas de una cadena de texto.
//
// Ingeniería Unicode:
// Utiliza la Forma de Normalización D (NFD) para descomponer caracteres complejos
// (ej. 'á' se divide en 'a' + '´'). Luego, un filtro elimina todos los caracteres
// clasificados como `unicode.Mn` (Mark, nonspacing - las tildes visuales).
// Finalmente, la Forma C (NFC) vuelve a ensamblar el texto limpio.
// Esto es vital en sistemas en español para evitar duplicados en búsquedas.
func removeDiacritics(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// foldCompare evalúa la equivalencia semántica de dos cadenas de texto.
//
// Tolerancia a Errores Humanos (Fuzzy Matching Básico):
// Ignora tanto las diferencias de capitalización (Mayúsculas/Minúsculas) mediante EqualFold,
// como la ausencia o presencia de tildes mediante removeDiacritics.
// Ej: "San Joaquín", "SAN JOAQUIN", "San joaquin" y "san joaquín" se considerarán idénticos.
func foldCompare(a, b string) bool {
	return strings.EqualFold(removeDiacritics(a), removeDiacritics(b))
}

// ─────────────────────────────────────────────────────────────────────────────
// MUNICIPIOS DE CARABOBO
// ─────────────────────────────────────────────────────────────────────────────

// Diccionario en memoria (In-Memory Catalog) estricto.
// Previene que errores tipográficos corrompan la integridad de los datos
// (ej. evita tener "Naguanagua" y "Naganagua" como entidades separadas en la DB).
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

// NormalizeMunicipioCarabobo recibe texto libre y devuelve el nombre canónico del municipio.
// Retorna un booleano (ok) que indica si la validación fue exitosa, permitiendo a la capa
// superior (Servicio) abortar la operación si el usuario inyecta datos falsos.
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

// Diccionario Nacional.
// Regla de Negocio (Domain Logic): El Estado Carabobo está intencionalmente omitido
// de esta lista, ya que el sistema maneja la ubicación base (Carabobo) y la ubicación
// foránea (resto del país) en columnas y flujos lógicos separados.
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

// NormalizeEstadoVenezuela coteja el texto de entrada contra el catálogo nacional
// aplicando las mismas reglas de tolerancia a acentos y mayúsculas.
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
// BOOL FROM FORM (Parseo Tri-estatal)
// ─────────────────────────────────────────────────────────────────────────────

// BoolFromForm normaliza banderas booleanas que provienen de peticiones HTTP
// estructuradas como `multipart/form-data` o `application/x-www-form-urlencoded`.
//
// Diseño Tri-Valente (Semántica PATCH):
// Devuelve un puntero a bool (*bool) en lugar de un primitivo. Esto es fundamental:
//   - Si el string es "1", "true" o "yes" -> retorna puntero a TRUE.
//   - Si el string es "0", "false" o "no" -> retorna puntero a FALSE.
//   - Si el string no viene, está vacío o es irreconocible -> retorna NIL.
//
// Retornar NIL permite al motor de base de datos saber que este campo NO DEBE
// actualizarse (fue omitido por el cliente), previniendo que se guarden valores falsos por error.
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

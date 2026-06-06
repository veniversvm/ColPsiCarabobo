// api/internal/request_structs/directory_filter_sanitizer.go

// Package request_structs provee las estructuras de transferencia de datos (DTOs)
// y sus respectivos mecanismos de validación/sanitización.
//
// Actúa como la primera línea de defensa (Capa de Presentación/Controlador) antes
// de que los datos introducidos por el usuario alcancen la capa de Casos de Uso (Dominio),
// garantizando que las consultas a la base de datos sean predecibles y seguras.
package request_structs

import (
	"strings"
	"unicode"
)

// SanitizeDirectoryFilter actúa como middleware de sanitización para el DTO de búsqueda pública.
//
// Responsabilidades:
//  1. Limpieza de Texto: Filtra caracteres especiales en los términos de búsqueda para prevenir
//     comportamientos inesperados en las cláusulas ILIKE (ej. evitar que el usuario inserte '%' o '_').
//  2. Listas Blancas (Whitelisting): Fuerza que el campo Gender solo acepte valores estrictamente válidos.
//  3. Límites de Paginación (Hard Limits): Previene ataques de denegación de servicio (DoS)
//     a la base de datos ajustando las páginas y limitando la cantidad máxima de resultados por consulta.
func SanitizeDirectoryFilter(f PsiDirectoryFilterDTO) PsiDirectoryFilterDTO {
	// Sanitizamos los campos de entrada de texto libre aplicando un límite máximo
	// de 100 caracteres para evitar saturación de memoria.
	f.SearchTerm = cleanSearchString(f.SearchTerm, 100)
	f.Location = cleanSearchString(f.Location, 100)

	// Gender: Aplicamos un patrón de Lista Blanca (Allowlist) estricta.
	// Todo lo que no coincida exactamente con los casos esperados se limpia.
	switch f.Gender {
	case "M", "F":
		// Valor válido, se mantiene.
	default:
		f.Gender = ""
	}

	// Normalización matemática de la Paginación
	if f.Page < 1 {
		f.Page = 1 // La página 0 o negativa no es válida en SQL OFFSET.
	}
	if f.Limit < 1 {
		f.Limit = 10 // Límite por defecto si el cliente envía 0.
	}
	if f.Limit > 100 {
		f.Limit = 100 // Límite estricto para evitar bloqueos por consultas masivas (DoS).
	}

	return f
}

// cleanSearchString aplica una estricta política de lista blanca sobre una cadena de texto.
//
// Retiene EXCLUSIVAMENTE letras Unicode (incluyendo caracteres hispanos como tildes, ñ, diéresis)
// y dígitos. Cualquier otro símbolo o carácter de control (%, _, -, ', ", ;, =, <, >)
// es descartado silenciosamente.
//
// También actúa como formateador, colapsando múltiples espacios intermedios en uno solo.
func cleanSearchString(s string, maxLen int) string {
	// Optimizamos la asignación de memoria reservando de antemano el tamaño del string original.
	// Esto previene que strings.Builder reasigne memoria múltiples veces bajo el capó.
	var b strings.Builder
	b.Grow(len(s))

	prevSpace := false
	for _, r := range s {
		// unicode.IsLetter garantiza el soporte internacional (no solo ASCII [A-Za-z])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSpace = false
		} else if unicode.IsSpace(r) && !prevSpace && b.Len() > 0 {
			// Solo permitimos un espacio, y nos aseguramos de que no se inserte al inicio.
			b.WriteRune(' ')
			prevSpace = true
		}
		// Todo lo que caiga fuera de las condiciones anteriores se ignora (Saneamiento silencioso).
	}

	result := strings.TrimSpace(b.String())

	// Límite de Longitud Seguro (Runes vs Bytes):
	// En Go, los caracteres hispanos (como la 'ñ' o vocales acentuadas) ocupan más de 1 byte.
	// Si truncamos el string usando bytes [a:b], podríamos partir un carácter Unicode a la mitad,
	// generando el carácter de reemplazo (). Convertirlo a []rune garantiza un corte exacto.
	runes := []rune(result)
	if len(runes) > maxLen {
		result = string(runes[:maxLen])
	}

	return result
}

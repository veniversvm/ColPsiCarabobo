// api/internal/request_structs/directory_filter_sanitizer.go
package request_structs

import (
	"strings"
	"unicode"
)

// SanitizeDirectoryFilter limpia el DTO de búsqueda del directorio público.
// Solo permite letras (incluyendo tildes y ñ), números y espacios simples.
// Cualquier carácter especial, símbolo o control es eliminado silenciosamente.
func SanitizeDirectoryFilter(f PsiDirectoryFilterDTO) PsiDirectoryFilterDTO {
	f.SearchTerm = cleanSearchString(f.SearchTerm, 100)
	f.Location = cleanSearchString(f.Location, 100)

	// Gender: lista blanca estricta
	switch f.Gender {
	case "M", "F":
		// válido
	default:
		f.Gender = ""
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 10
	}
	if f.Limit > 100 {
		f.Limit = 100
	}

	return f
}

// cleanSearchString filtra el string dejando solo letras Unicode (incluye tildes,
// ñ, diéresis, etc.), dígitos y espacios. Colapsa espacios múltiples en uno solo
// y aplica un límite de longitud máxima.
func cleanSearchString(s string, maxLen int) string {
	var b strings.Builder
	b.Grow(len(s))

	prevSpace := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSpace = false
		} else if unicode.IsSpace(r) && !prevSpace && b.Len() > 0 {
			// Solo un espacio, nunca al inicio
			b.WriteRune(' ')
			prevSpace = true
		}
		// Todo lo demás (%, _, -, ', ", ;, =, <, >, etc.) se descarta silenciosamente
	}

	result := strings.TrimSpace(b.String())

	// Límite de longitud en runes (no bytes) para soportar Unicode correctamente
	runes := []rune(result)
	if len(runes) > maxLen {
		result = string(runes[:maxLen])
	}

	return result
}

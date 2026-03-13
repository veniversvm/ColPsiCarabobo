package utils // O el paquete donde la vayas a colocar

import (
	"strings"
	"unicode"
)

// CleanAlphaNumeric mantiene letras (incluyendo ñ y tildes) y números.
func CleanAlphaNumeric(s string) string {
	var builder strings.Builder
	// Pre-asignar memoria basándonos en la longitud original para mayor rendimiento
	builder.Grow(len(s))

	for _, r := range s {
		// IsLetter acepta caracteres Unicode (ñ, acentos)
		// IsDigit acepta números
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

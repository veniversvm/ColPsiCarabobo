// api/internal/utils/radom_string.go
// Package utils provee herramientas transversales de soporte para la aplicación.
package utils

import (
	"math/rand/v2"
	"strings"
)

// =========================================================================
// GENERACIÓN DE CRIPTOGRAFÍA Y LLAVES
// =========================================================================

// key_charset define el conjunto de caracteres permitidos para las llaves de sesión.
// Se incluye el guion bajo (_) y se omiten caracteres especiales conflictivos
// para garantizar total compatibilidad con serialización JSON y consultas SQL en GORM.
const key_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

// GenerateSecureRandomString genera una cadena aleatoria de longitud 'n'.
//
// LÓGICA DE IMPLEMENTACIÓN:
//  1. Eficiencia: Usa strings.Builder con sb.Grow(n) para evitar múltiples
//     asignaciones de memoria en el heap durante la concatenación.
//  2. Modernidad: Utiliza math/rand/v2, que implementa el algoritmo ChaCha8
//     para una generación de números aleatorios más robusta y veloz.
//  3. Seguridad: Ideal para generar 'Keys' de sesión dinámica, tokens de
//     verificación o identificadores de recursos temporales.
func GenerateSecureRandomString(n int) string {
	var sb strings.Builder
	// Pre-asignamos la memoria necesaria para el string final
	sb.Grow(n)

	charsetLen := len(key_charset)
	for i := 0; i < n; i++ {
		// rand.IntN es seguro para uso concurrente (múltiples goroutines)
		// y no requiere seed manual en la versión v2 de Go.
		randomIndex := rand.IntN(charsetLen)
		sb.WriteByte(key_charset[randomIndex])
	}

	return sb.String()
}

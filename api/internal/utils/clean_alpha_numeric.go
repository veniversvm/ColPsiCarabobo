// Package utils provee herramientas y funciones transversales (Cross-Cutting Concerns).
//
// Aloja utilidades puras (sin dependencias de estado o base de datos) utilizadas a lo largo
// de todo el sistema para tareas recurrentes como sanitización, validaciones criptográficas
// y formateo de datos.
package utils // O el paquete donde la vayas a colocar

import (
	"strings"
	"unicode"
)

// CleanAlphaNumeric purifica una cadena de texto extrayendo estrictamente caracteres seguros.
//
// Seguridad (Sanitización Defensiva):
// Elimina silenciosamente cualquier símbolo de puntuación, espacio o carácter de control
// (ej. <, >, ', ", ;, %, NULL). Es ideal para generar llaves de caché (Cache Keys) seguras,
// o limpiar parámetros de búsqueda que podrían ser utilizados para ataques de Inyección
// (SQLi / NoSQLi / XSS).
//
// Soporte i18n (Internacionalización):
// Al utilizar la librería nativa `unicode` en lugar de una Expresión Regular simple como `[a-zA-Z0-9]`,
// esta función respeta de forma nativa los alfabetos internacionales. Conserva perfectamente
// caracteres hispanos vitales como la 'ñ', 'Ñ' y las vocales acentuadas ('á', 'é', etc.).
func CleanAlphaNumeric(s string) string {
	var builder strings.Builder

	// Optimización de Memoria (Pre-asignación):
	// En Go, los strings son inmutables. Concatenar strings en un bucle (`str += "a"`)
	// provoca constantes re-asignaciones de memoria y sobrecarga el Garbage Collector (GC).
	// Al usar strings.Builder y pre-asignar el tamaño exacto del string original con Grow(),
	// garantizamos que la operación ocurra en un único bloque de memoria RAM (O(1) allocation).
	builder.Grow(len(s))

	// Iteración segura sobre runas (soporte UTF-8 multicapa)
	for _, r := range s {
		// IsLetter acepta caracteres Unicode completos (ñ, acentos, diéresis)
		// IsDigit acepta números estándar.
		// Todo lo demás (espacios, guiones, símbolos matemáticos) es ignorado.
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

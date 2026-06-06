// api/internal/utils/radom_string.go

// Package utils provee herramientas transversales de soporte para la aplicación.
//
// Contiene funciones de utilidad pura (sin estado) que no dependen de la infraestructura
// ni del dominio, diseñadas para ser consumidas a lo largo de todo el ciclo de vida del request.
package utils

import (
	"math/rand/v2"
	"strings"
)

// =========================================================================
// GENERACIÓN DE CRIPTOGRAFÍA Y LLAVES (TOKENS)
// =========================================================================

// key_charset define el conjunto de caracteres permitidos (Base63) para las llaves de sesión.
//
// Diseño Seguro por Defecto (Safe-Encoding):
// Se incluye el guion bajo (_) y se omiten deliberadamente caracteres especiales conflictivos
// (como comillas, diagonales, porcentajes o símbolos de escape). Esto garantiza que la llave
// generada sea "URL-Safe" y completamente robusta ante la serialización JSON o las
// consultas SQL en GORM, erradicando vectores pasivos de inyección o errores de codificación.
const key_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

// GenerateSecureRandomString genera una cadena pseudo-aleatoria de longitud 'n'.
//
// LÓGICA DE IMPLEMENTACIÓN Y ALTO RENDIMIENTO:
//
//  1. Optimización de Memoria (Zero-Allocation Pattern):
//     Usa `strings.Builder` con `sb.Grow(n)`. En lugar de crear un nuevo bloque de memoria
//     en el Heap de Go por cada carácter concatenado (lo cual saturaría el Garbage Collector
//     bajo alta carga), reserva la memoria exacta de antemano. La complejidad de asignación
//     pasa a ser de O(1) constante.
//
//  2. Modernidad y Concurrencia (Go 1.22+):
//     Utiliza `math/rand/v2`. A diferencia del paquete `rand` clásico (que sufría cuellos
//     de botella por bloqueos globales Mutex y requería inicializar semillas de tiempo manualmente),
//     la v2 implementa el potente algoritmo ChaCha8. Es veloz, no requiere "seeding" manual y
//     no se bloquea cuando miles de Goroutines piden tokens al mismo tiempo.
//
//  3. Aplicación de Seguridad:
//     Especialmente diseñado para la Rotación de Llaves de Sesión (Key Rotation),
//     generación de contraseñas temporales por correo o tokens de reseteo.
func GenerateSecureRandomString(n int) string {
	var sb strings.Builder
	// Pre-asignamos la memoria exacta necesaria para el string final, evitando realocaciones.
	sb.Grow(n)

	charsetLen := len(key_charset)
	for i := 0; i < n; i++ {
		// rand.IntN (v2) es Thread-Safe nativo para uso concurrente masivo
		// y delega al algoritmo ChaCha8 la aleatoriedad matemática.
		randomIndex := rand.IntN(charsetLen)
		sb.WriteByte(key_charset[randomIndex])
	}

	return sb.String()
}

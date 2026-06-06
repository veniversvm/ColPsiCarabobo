// api/internal/utils/secure_password.go

// Package utils provee herramientas transversales de soporte para la aplicación.
//
// Este archivo centraliza el Motor de Políticas de Contraseñas (Password Policy Engine).
// Garantiza que todas las credenciales ingresadas al sistema posean la entropía
// matemática necesaria para resistir ataques criptográficos, actuando como filtro
// previo antes de someter el string a la capa pesada de hashing (Bcrypt).
package utils

import (
	"unicode"
)

// =========================================================================
// VALIDACIÓN DE SEGURIDAD (POLÍTICAS DE ACCESO Y ENTROPÍA)
// =========================================================================

// IsStrongPassword implementa la lógica de auditoría de complejidad para credenciales.
//
// Diseño Criptográfico y Cumplimiento (Security Guidelines):
// Evalúa la cadena asegurando una entropía mínima (variabilidad y dispersión de caracteres)
// suficiente para proteger la identidad del psicólogo contra Vectores de Ataque comunes
// como Ataques de Diccionario (Dictionary Attacks) y Fuerza Bruta (Brute-Force).
//
// CRITERIOS ESTRICTOS DE EVALUACIÓN:
// 1. Longitud: Mínimo 8 caracteres (Aumenta exponencialmente el tiempo de crackeo).
// 2. Diversidad: Obliga la permutación de conjuntos de caracteres (Mayúsculas, Minúsculas y Números).
// 3. Robustez: Exige al menos un carácter especial (puntuación o símbolo).
func IsStrongPassword(password string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	// 1. Validación de Longitud Mínima (Filtro Base)
	if len(password) >= 8 {
		hasMinLen = true
	}

	// 2. Prevención de Inyección de Blancos (Short-Circuit Evaluation)
	// Los espacios en blanco suelen causar comportamientos indefinidos al codificar cabeceras
	// HTTP Basic Auth, o generar fricción UX cuando el usuario copia/pega la contraseña.
	// El uso de 'return false' aquí aplica un patrón Fail-Fast, abortando el proceso de
	// inmediato sin gastar CPU en el análisis de composición.
	for _, char := range password {
		if unicode.IsSpace(char) {
			return false
		}
	}

	// 3. Análisis de Composición Criptográfica (Iteración Segura por Runas)
	// En Go, iterar un string con `range` extrae Runas (caracteres Unicode reales de 32 bits)
	// en lugar de iterar byte por byte. Esto es crucial en ciberseguridad para asegurar
	// que si un usuario utiliza caracteres internacionales, el motor no lea la memoria de forma fragmentada.
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		// IsPunct e IsSymbol cubren matemáticamente todo el espectro de caracteres
		// especiales requeridos por las normativas corporativas (@, #, $, !, &, *, etc.).
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 4. Puerta Lógica AND (Validación Integral)
	// Retorna true SÓLO si la credencial superó satisfactoriamente los cinco pilares
	// de la política de seguridad del Colegio.
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

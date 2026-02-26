// api/internal/utils/secure_password.go
// Package utils provee herramientas transversales de soporte para la aplicación.
package utils

import (
	"unicode"
)

// =========================================================================
// VALIDACIÓN DE SEGURIDAD (POLÍTICAS DE ACCESO)
// =========================================================================

// IsStrongPassword implementa la lógica de auditoría de complejidad para contraseñas.
// Evalúa la cadena basándose en estándares modernos de seguridad (NIST), asegurando
// una entropía mínima suficiente para proteger la identidad del psicólogo.
//
// CRITERIOS DE EVALUACIÓN:
// 1. Longitud: Mínimo 8 caracteres.
// 2. Diversidad: Al menos una Mayúscula, una Minúscula y un Número.
// 3. Robustez: Al menos un carácter especial (puntuación o símbolo).
func IsStrongPassword(password string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	// 1. Validación de Longitud Mínima
	if len(password) >= 8 {
		hasMinLen = true
	}

	// 2. Validación de Espacios (No se permiten espacios en blanco)
	for _, char := range password {
		if unicode.IsSpace(char) {
			return false
		}
	}

	// 3. Análisis de Composición (Iteración por Runas)
	// Se itera sobre cada carácter (runa) para soportar caracteres UTF-8.
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		// IsPunct e IsSymbol cubren el espectro de caracteres especiales como @, #, $, !, etc.
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Retorna true solo si cumple con los cinco pilares de la política
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

package utils

import (
	"errors"
	"net/mail"
	"strings"
)

// ParseAndValidateEmail recibe un string, limpia los espacios y verifica si es un email válido.
// Retorna el email limpio en minúsculas o un error si el formato es incorrecto.
func ParseAndValidateEmail(email string) (string, error) {
	// 1. Limpiamos espacios en blanco a los lados
	email = strings.TrimSpace(email)

	if email == "" {
		return "", errors.New("el correo electrónico no puede estar vacío")
	}

	// 2. ParseAddress valida que cumpla con el estándar de emails
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", errors.New("formato de correo electrónico inválido")
	}

	// 3. (Opcional pero recomendado) Retornar siempre en minúsculas
	// addr.Address extrae solo el correo.
	// Si alguien pasa "Juan <juan@mail.com>", extrae solo "juan@mail.com"
	cleanEmail := strings.ToLower(addr.Address)

	return cleanEmail, nil
}

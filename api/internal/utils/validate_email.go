// api/internal/utils/parse_validate_email.go

// Package utils provee herramientas transversales de soporte para la aplicación.
//
// Este archivo actúa como una barrera de Calidad de Datos (Data Quality) y
// Normalización de Identidad. Garantiza que el principal vector de comunicación y
// autenticación del sistema (el correo electrónico) mantenga una integridad absoluta.
package utils

import (
	"errors"
	"net/mail"
	"strings"
)

// ParseAndValidateEmail orquesta el pipeline de saneamiento y validación de correos.
// Recibe una cadena de texto cruda, aplica reglas de limpieza y verifica el formato.
// Retorna la dirección canónica en minúsculas o un error descriptivo.
//
// Prevención de ReDoS (Regular Expression Denial of Service):
// Muchos desarrolladores utilizan Expresiones Regulares (Regex) complejas para
// validar correos, las cuales son propensas a bloqueos de CPU si un atacante envía
// un string maliciosamente diseñado. Al delegar esto a `net/mail`, nos apalancamos
// en un parser nativo optimizado y estrictamente apegado al estándar RFC 5322.
func ParseAndValidateEmail(email string) (string, error) {
	// 1. Sanitización de UX (Tolerancia a Errores Comunes)
	// Limpiamos los espacios en blanco invisibles (Leading/Trailing spaces).
	// Esto es crucial en mobile UX, ya que los teclados predictivos suelen inyectar
	// un espacio en blanco automático después de que el usuario autocompleta su correo,
	// lo cual arruinaría el inicio de sesión si no se recorta.
	email = strings.TrimSpace(email)

	// Failsafe de Negocio: Aborta rápidamente si el payload quedó vacío
	if email == "" {
		return "", errors.New("el correo electrónico no puede estar vacío")
	}

	// 2. Validación Estricta (RFC 5322 Compliance)
	// mail.ParseAddress no solo valida la estructura, sino que es capaz de destilar
	// direcciones formateadas. Extrae limpiamente la dirección enrutable ignorando
	// nombres de visualización o metadatos inyectados.
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", errors.New("formato de correo electrónico inválido")
	}

	// 3. Canonización de Identidad (Data Canonicalization)
	// La especificación dice que los dominios son Case-Insensitive, pero la parte local
	// podría no serlo. Sin embargo, en bases de datos modernas (como PostgreSQL),
	// "Juan@Mail.com" y "juan@mail.com" son strings distintos y romperían la restricción UNIQUE.
	// Forzar todo a minúsculas absolutas previene cuentas duplicadas y bloqueos
	// de login por variaciones en la capitalización del teclado.
	cleanEmail := strings.ToLower(addr.Address)

	return cleanEmail, nil
}

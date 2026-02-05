package utils

import (
	"math/rand/v2"
	"strings"
)

// Este charset funcionará en tu contexto específico de GORM + JSON
const key_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

// Charset con caracteres especiales "seguros" y comúnmente aceptados.
func GenerateSecureRandomString(n int) string {
	var sb strings.Builder
	sb.Grow(n) // Pre-asigna memoria para eficiencia

	charsetLen := len(key_charset)
	for i := 0; i < n; i++ {
		// rand.IntN es seguro para concurrencia y se siembra automáticamente.
		randomIndex := rand.IntN(charsetLen)
		sb.WriteByte(key_charset[randomIndex])
	}

	return sb.String()
}

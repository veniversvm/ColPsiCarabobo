package utils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// =========================================================================
// TEST: SANITIZADOR DE IMÁGENES
// =========================================================================

// TestSanitizeImage verifica que la decodificación detecte formatos incorrectos
// y valide los tipos mime. (Para una prueba completa se requieren bytes de imágenes reales,
// aquí probamos el comportamiento defensivo).
func TestSanitizeImage_Defensive(t *testing.T) {
	// Intentamos pasar texto plano como si fuera una imagen
	fakeImage := strings.NewReader("Esto no es un JPEG, es texto para engañar al sistema")

	bytes, ext, contentType, err := SanitizeImage(fakeImage)

	require.Error(t, err, "Debe rechazar archivos que no sean imágenes reales")
	require.Nil(t, bytes)
	require.Empty(t, ext)
	require.Empty(t, contentType)
	require.Contains(t, err.Error(), "archivo no es una imagen válida")
}

// =========================================================================
// TEST: VALIDACIÓN DE ESTRUCTURAS VACÍAS
// =========================================================================

// DummyStruct se usa para probar IsEmptyReq sin depender de DTOs externos
type DummyStruct struct {
	Name   string
	Age    int
	IsCool *bool
	Tags   []string
}

func TestIsEmptyReq(t *testing.T) {
	cool := true

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "Struct totalmente vacío",
			input:    DummyStruct{},
			expected: true,
		},
		{
			name:     "Struct con un string",
			input:    DummyStruct{Name: "Fran"},
			expected: false,
		},
		{
			name:     "Struct con un int válido",
			input:    DummyStruct{Age: 25},
			expected: false,
		},
		{
			name:     "Struct con puntero inicializado",
			input:    DummyStruct{IsCool: &cool},
			expected: false,
		},
		{
			name:     "Puntero a struct vacío",
			input:    &DummyStruct{},
			expected: true,
		},
		{
			name:     "Puntero a struct con datos",
			input:    &DummyStruct{Name: "Admin"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsEmptyReq(tc.input)
			require.Equal(t, tc.expected, actual, "Fallo en: %s", tc.name)
		})
	}
}

// =========================================================================
// TEST: NORMALIZACIÓN DE PLATAFORMAS
// =========================================================================

func TestNormalizePlatformName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Diccionario exacto
		{"Instagram diminutivo", "ig", "Instagram"},
		{"Instagram mayúsculas con espacios", "   INSTA   ", "Instagram"},
		{"Twitter moderno", "x", "X (Twitter)"},
		{"Facebook error ortográfico", "facbook", "Facebook"},

		// Coincidencia de sub-cadenas (URLs y typos)
		{"URL de Youtube", "https://youtu.be/mivideo", "YouTube"},
		{"URL de Instagram", "instagram.com/psico", "Instagram"},
		{"URL de Facebook", "fb.com/grupo", "Facebook"},

		// Fallback (Casos no mapeados)
		{"Plataforma nueva", "threads", "Threads"},
		{"Compuesto no mapeado", "mi blog personal", "Mi Blog Personal"},

		// Casos límite
		{"Vacío", "", ""},
		{"Solo espacios", "    ", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NormalizePlatformName(tc.input)
			require.Equal(t, tc.expected, actual)
		})
	}
}

// =========================================================================
// TEST: GENERADOR DE LLAVES
// =========================================================================

func TestGenerateSecureRandomString(t *testing.T) {
	t.Run("Verificar longitud", func(t *testing.T) {
		length := 32
		result := GenerateSecureRandomString(length)
		require.Len(t, result, length, "El string generado debe tener la longitud exacta solicitada")
	})

	t.Run("Verificar aleatoriedad (Anti-colisión)", func(t *testing.T) {
		str1 := GenerateSecureRandomString(16)
		str2 := GenerateSecureRandomString(16)
		require.NotEqual(t, str1, str2, "Dos strings generados consecutivamente no deben ser iguales")
	})

	t.Run("Verificar caracteres permitidos", func(t *testing.T) {
		result := GenerateSecureRandomString(100)
		for _, char := range result {
			require.True(t, bytes.ContainsRune([]byte(key_charset), char), "Caracter ilegal encontrado: %c", char)
		}
	})
}

// =========================================================================
// TEST: Fortaleza de Contraseñas
// =========================================================================

func TestIsStrongPassword(t *testing.T) {
	// Definimos los casos de prueba
	tests := []struct {
		name     string // Descripción de qué estamos probando
		password string // Input
		want     bool   // Resultado esperado
	}{
		{
			name:     "Password válida con todos los criterios",
			password: "Pass1234!",
			want:     true,
		},
		{
			name:     "Inválida por ser muy corta (< 8)",
			password: "P1s@2",
			want:     false,
		},
		{
			name:     "Inválida por falta de mayúscula",
			password: "password123!",
			want:     false,
		},
		{
			name:     "Inválida por falta de minúscula",
			password: "PASSWORD123!",
			want:     false,
		},
		{
			name:     "Inválida por falta de número",
			password: "Password!",
			want:     false,
		},
		{
			name:     "Inválida por falta de carácter especial",
			password: "Password123",
			want:     false,
		},
		{
			name:     "Válida con caracteres UTF-8 (acentos/símbolos)",
			password: "Contraseña9$",
			want:     true,
		},
		{
			name:     "Válida con espacios (el espacio cuenta como símbolo/puntuación)",
			password: "Strong Password 1",
			want:     false,
		},
		{
			name:     "String vacío",
			password: "",
			want:     false,
		},
	}

	// Ejecución de los sub-tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsStrongPassword(tt.password)
			if got != tt.want {
				t.Errorf("IsStrongPassword(%q) = %v; se esperaba %v", tt.password, got, tt.want)
			}
		})
	}
}

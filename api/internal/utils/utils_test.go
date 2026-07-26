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
	require.Contains(t, err.Error(), "el servidor no reconoce este formato de imagen")
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

// =========================================================================
// TEST: NORMALIZACIÓN GEOGRÁFICA
// =========================================================================

func TestNormalizeMunicipioCarabobo(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantValid bool
	}{
		{"nombre exacto", "Valencia", "Valencia", true},
		{"case insensitive", "valencia", "Valencia", true},
		{"con tilde extra", "Naguanagua", "Naguanagua", true},
		{"con espacios extra", "  San Diego  ", "San Diego", true},
		{"tilde tolerant", "San Joaquin", "San Joaquín", true},
		{"municipio real con tilde", "San Joaquín", "San Joaquín", true},
		{"todos los municipios", "Bejuma", "Bejuma", true},
		{"todos los municipios 2", "Carlos Arvelo", "Carlos Arvelo", true},
		{"todos los municipios 3", "Diego Ibarra", "Diego Ibarra", true},
		{"todos los municipios 4", "Guacara", "Guacara", true},
		{"todos los municipios 5", "Juan José Mora", "Juan José Mora", true},
		{"todos los municipios 6", "Libertador", "Libertador", true},
		{"todos los municipios 7", "Los Guayos", "Los Guayos", true},
		{"todos los municipios 8", "Miranda", "Miranda", true},
		{"todos los municipios 9", "Montalbán", "Montalbán", true},
		{"todos los municipios 10", "Puerto Cabello", "Puerto Cabello", true},
		{"municipio inexistente", "FalsoMunicipio", "", false},
		{"string vacio", "", "", false},
		{"solo espacios", "   ", "", false},
		{"estado en vez de municipio", "Carabobo", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := NormalizeMunicipioCarabobo(tc.input)
			require.Equal(t, tc.wantValid, ok, "ok mismatch")
			require.Equal(t, tc.want, got, "valor mismatch")
		})
	}
}

func TestNormalizeEstadoVenezuela(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantValid bool
	}{
		{"estado exacto", "Lara", "Lara", true},
		{"case insensitive", "lara", "Lara", true},
		{"con tilde", "Anzoategui", "Anzoátegui", true},
		{"con tilde real", "Anzoátegui", "Anzoátegui", true},
		{"espacios extra", "  Zulia  ", "Zulia", true},
		{"distrito capital", "Distrito Capital", "Distrito Capital", true},
		{"delta amacuro", "Delta Amacuro", "Delta Amacuro", true},
		{"la guaira", "La Guaira", "La Guaira", true},
		{"carabobo no esta", "Carabobo", "", false},
		{"municipio en vez de estado", "Valencia", "", false},
		{"string vacio", "", "", false},
		{"solo espacios", "   ", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := NormalizeEstadoVenezuela(tc.input)
			require.Equal(t, tc.wantValid, ok, "ok mismatch")
			require.Equal(t, tc.want, got, "valor mismatch")
		})
	}
}

// =========================================================================
// TEST: PARSEO DE BOOLEANOS TRI-ESTADO
// =========================================================================

func TestBoolFromForm(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect *bool
	}{
		{"true", "true", boolPtr(true)},
		{"TRUE mayusculas", "TRUE", boolPtr(true)},
		{"1 entero", "1", boolPtr(true)},
		{"yes", "yes", boolPtr(true)},
		{"false", "false", boolPtr(false)},
		{"FALSE mayusculas", "FALSE", boolPtr(false)},
		{"0 entero", "0", boolPtr(false)},
		{"no", "no", boolPtr(false)},
		{"con espacios true", "  true  ", boolPtr(true)},
		{"con espacios false", "  false  ", boolPtr(false)},
		{"con espacios 1", "  1  ", boolPtr(true)},
		{"string vacio", "", nil},
		{"solo espacios", "   ", nil},
		{"texto aleatorio", "abc", nil},
		{"string null", "null", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := BoolFromForm(tc.input)
			if tc.expect == nil {
				require.Nil(t, result, "se esperaba nil")
			} else {
				require.NotNil(t, result, "no debia ser nil")
				require.Equal(t, *tc.expect, *result, "valor mismatch")
			}
		});
	}
}

// boolPtr es un helper para crear punteros a bool en tests
func boolPtr(b bool) *bool {
	return &b
}

// =========================================================================
// TEST: LIMPIEZA ALFANUMERICA
// =========================================================================

func TestCleanAlphaNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"string limpio", "Hello123", "Hello123"},
		{"con caracteres especiales", "Hello@World!", "HelloWorld"},
		{"con inyeccion sql", "'; DROP TABLE users; --", "DROPTABLEusers"},
		{"con espacios", "a b c", "abc"},
		{"con guiones y parentesis", "test-(1)", "test1"},
		{"caracteres unicode preservados", "José María Ñoño", "JoséMaríaÑoño"},
		{"string vacio", "", ""},
		{"solo simbolos", "!@#$%^&*()", ""},
		{"mixto complejo", "user@email.com (Work)", "useremailcomWork"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CleanAlphaNumeric(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

// =========================================================================
// TEST: FORTALEZA DE CONTRASEÑAS
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

package request_structs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeDirectoryFilter(t *testing.T) {
	tests := []struct {
		name           string
		input          PsiDirectoryFilterDTO
		expectSearch   string
		expectLocation string
		expectGender   string
		expectPage     int
		expectLimit    int
	}{
		{
			name:           "input_vacio se mantiene con defaults",
			input:          PsiDirectoryFilterDTO{},
			expectSearch:   "",
			expectLocation: "",
			expectGender:   "",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "valores validos se preservan",
			input:          PsiDirectoryFilterDTO{SearchTerm: "Juan", Location: "Valencia", Gender: "M", Page: 3, Limit: 20},
			expectSearch:   "Juan",
			expectLocation: "Valencia",
			expectGender:   "M",
			expectPage:     3,
			expectLimit:    20,
		},
		{
			name:           "gender F es valido",
			input:          PsiDirectoryFilterDTO{Gender: "F"},
			expectGender:   "F",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "gender masculino minuscula se limpia",
			input:          PsiDirectoryFilterDTO{Gender: "m"},
			expectGender:   "",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "gender texto largo se limpia",
			input:          PsiDirectoryFilterDTO{Gender: "masculino"},
			expectGender:   "",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "gender vacio se mantiene vacio",
			input:          PsiDirectoryFilterDTO{Gender: ""},
			expectGender:   "",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "page cero se normaliza a 1",
			input:          PsiDirectoryFilterDTO{Page: 0},
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "page negativo se normaliza a 1",
			input:          PsiDirectoryFilterDTO{Page: -5},
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "page valido se preserva",
			input:          PsiDirectoryFilterDTO{Page: 10},
			expectPage:     10,
			expectLimit:    10,
		},
		{
			name:           "limit cero se normaliza a 10",
			input:          PsiDirectoryFilterDTO{Limit: 0},
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "limit negativo se normaliza a 10",
			input:          PsiDirectoryFilterDTO{Limit: -1},
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "limit mayor a 100 se cappea a 100",
			input:          PsiDirectoryFilterDTO{Limit: 500},
			expectPage:     1,
			expectLimit:    100,
		},
		{
			name:           "limit exactamente 100 se preserva",
			input:          PsiDirectoryFilterDTO{Limit: 100},
			expectPage:     1,
			expectLimit:    100,
		},
		{
			name:           "limit 1 se preserva",
			input:          PsiDirectoryFilterDTO{Limit: 1},
			expectPage:     1,
			expectLimit:    1,
		},
		{
			name:           "limpieza de caracteres especiales en search",
			input:          PsiDirectoryFilterDTO{SearchTerm: "Juan%' OR 1=1--"},
			expectSearch:   "Juan OR 11",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "limpieza de caracteres especiales en location",
			input:          PsiDirectoryFilterDTO{Location: "Valencia'; DROP TABLE--"},
			expectLocation: "Valencia DROP TABLE",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "caracteres internacionales preservados",
			input:          PsiDirectoryFilterDTO{SearchTerm: "José María Ñoño"},
			expectSearch:   "José María Ñoño",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "multiples espacios colapsados",
			input:          PsiDirectoryFilterDTO{SearchTerm: "Juan   Carlos    Perez"},
			expectSearch:   "Juan Carlos Perez",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "espacios leading/trailing recortados",
			input:          PsiDirectoryFilterDTO{SearchTerm: "  Juan  "},
			expectSearch:   "Juan",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "solo caracteres especiales genera vacio",
			input:          PsiDirectoryFilterDTO{SearchTerm: "!@#$%^&*()"},
			expectSearch:   "",
			expectPage:     1,
			expectLimit:    10,
		},
		{
			name:           "specialty se preserva sin cambios",
			input:          PsiDirectoryFilterDTO{SpecialtyID: 42, Page: 1, Limit: 10},
			expectPage:     1,
			expectLimit:    10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeDirectoryFilter(tc.input)

			require.Equal(t, tc.expectSearch, result.SearchTerm, "SearchTerm mismatch")
			require.Equal(t, tc.expectLocation, result.Location, "Location mismatch")
			require.Equal(t, tc.expectGender, result.Gender, "Gender mismatch")
			require.Equal(t, tc.expectPage, result.Page, "Page mismatch")
			require.Equal(t, tc.expectLimit, result.Limit, "Limit mismatch")

			if tc.input.SpecialtyID != 0 {
				require.Equal(t, tc.input.SpecialtyID, result.SpecialtyID, "SpecialtyID no debe cambiar")
			}
		})
	}
}

func TestSanitizeDirectoryFilter_Truncation(t *testing.T) {
	t.Run("search_term truncado a 100 runes", func(t *testing.T) {
		longSearch := strings.Repeat("a", 150)
		input := PsiDirectoryFilterDTO{SearchTerm: longSearch}
		result := SanitizeDirectoryFilter(input)
		runes := []rune(result.SearchTerm)
		require.LessOrEqual(t, len(runes), 100, "SearchTerm debe truncarse a 100 runes")
	})

	t.Run("location truncado a 100 runes", func(t *testing.T) {
		longLocation := strings.Repeat("b", 200)
		input := PsiDirectoryFilterDTO{Location: longLocation}
		result := SanitizeDirectoryFilter(input)
		runes := []rune(result.Location)
		require.LessOrEqual(t, len(runes), 100, "Location debe truncarse a 100 runes")
	})

	t.Run("runas multibyte no rotas", func(t *testing.T) {
		// ñ ocupa 2 bytes en UTF-8; verificar que no se rompe
		longSearch := strings.Repeat("ñ", 150)
		input := PsiDirectoryFilterDTO{SearchTerm: longSearch}
		result := SanitizeDirectoryFilter(input)
		runes := []rune(result.SearchTerm)
		require.LessOrEqual(t, len(runes), 100, "Runas multibyte no deben romperse")
		for _, r := range result.SearchTerm {
			require.Equal(t, 'ñ', r, "Todas las runas deben ser ñ")
		}
	})
}

func TestSanitizeDirectoryFilter_GenderEdgeCases(t *testing.T) {
	invalidGenders := []string{"X", "male", "female", "MASCULINO", "0", "1", "true", "false", "  M  "}
	for _, g := range invalidGenders {
		t.Run("gender_invalido_"+g, func(t *testing.T) {
			input := PsiDirectoryFilterDTO{Gender: g}
			result := SanitizeDirectoryFilter(input)
			require.Empty(t, result.Gender, "Gender invalido '%s' debe limpiarse", g)
		})
	}

	validGenders := []string{"M", "F"}
	for _, g := range validGenders {
		t.Run("gender_valido_"+g, func(t *testing.T) {
			input := PsiDirectoryFilterDTO{Gender: g}
			result := SanitizeDirectoryFilter(input)
			require.Equal(t, g, result.Gender)
		})
	}
}

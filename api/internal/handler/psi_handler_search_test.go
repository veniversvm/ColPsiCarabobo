package handler

import (
	"testing"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// TestEsBusquedaReal fija qué cuenta el panel como "Búsquedas" y qué no.
//
// El endpoint GET /psi/directory sirve para tres cosas distintas y solo una es
// una búsqueda: el listado inicial (el onMount del frontend entra sin filtros),
// la paginación por scroll infinito (loadMore repite el endpoint con page=2,3…)
// y la búsqueda de verdad. Grabarlas las tres hacía que "Búsquedas" no fuera
// comparable con "Visitas": 564 búsquedas contra 14 visitas el mismo día, 40×
// más, alimentadas en parte por un monitor que pegaba al directorio cada 60 s.
func TestEsBusquedaReal(t *testing.T) {
	si := true

	tests := []struct {
		nombre string
		filter request_structs.PsiDirectoryFilterDTO
		espera bool
	}{
		{
			nombre: "listado inicial sin filtros no es búsqueda",
			filter: request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 12},
			espera: false,
		},
		{
			nombre: "texto sí es búsqueda",
			filter: request_structs.PsiDirectoryFilterDTO{SearchTerm: "depresion", Page: 1},
			espera: true,
		},
		{
			nombre: "filtro por área sí es búsqueda",
			filter: request_structs.PsiDirectoryFilterDTO{SpecialtyID: 3, Page: 1},
			espera: true,
		},
		{
			nombre: "filtro por ubicación sí es búsqueda",
			filter: request_structs.PsiDirectoryFilterDTO{Location: "valencia", Page: 1},
			espera: true,
		},
		{
			nombre: "filtro por género sí es búsqueda",
			filter: request_structs.PsiDirectoryFilterDTO{Gender: "F", Page: 1},
			espera: true,
		},
		{
			nombre: "filtro por solvencia sí es búsqueda aunque sea false",
			filter: request_structs.PsiDirectoryFilterDTO{Solvent: &si, Page: 1},
			espera: true,
		},
		{
			nombre: "filtro por estado también cuenta",
			filter: request_structs.PsiDirectoryFilterDTO{Active: &si, Page: 1},
			espera: true,
		},
		{
			nombre: "página 2 sin filtros es paginación, no búsqueda",
			filter: request_structs.PsiDirectoryFilterDTO{Page: 2, Limit: 12},
			espera: false,
		},
		{
			nombre: "página 2 CON texto sigue siendo la misma búsqueda",
			filter: request_structs.PsiDirectoryFilterDTO{SearchTerm: "depresion", Page: 2},
			espera: false,
		},
		{
			nombre: "página 5 con área no se re-cuenta",
			filter: request_structs.PsiDirectoryFilterDTO{SpecialtyID: 3, Page: 5},
			espera: false,
		},
		{
			nombre: "espacios en blanco no convierten el listado en búsqueda",
			filter: request_structs.PsiDirectoryFilterDTO{SearchTerm: "   ", Location: "  ", Page: 1},
			espera: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.nombre, func(t *testing.T) {
			if got := esBusquedaReal(tc.filter); got != tc.espera {
				t.Errorf("esBusquedaReal(%+v) = %v, se esperaba %v", tc.filter, got, tc.espera)
			}
		})
	}
}

// TestEsBusquedaReal_ConElFiltroYaSanitizado comprueba el contrato con
// SanitizeDirectoryFilter, que es el que realmente se lee en el handler: la
// decisión se toma sobre el filtro normalizado, no sobre la query string cruda,
// para que Gender solo pueda valer ""/M/F y Page nunca sea 0.
func TestEsBusquedaReal_ConElFiltroYaSanitizado(t *testing.T) {
	crudo := request_structs.PsiDirectoryFilterDTO{
		SearchTerm: "  <script>alert(1)</script>  ",
		Gender:     "X",
		Page:       0,
		Limit:      0,
	}
	sanitizado := request_structs.SanitizeDirectoryFilter(crudo)

	if sanitizado.Page < 1 {
		t.Fatalf("Page debe quedar >= 1 tras sanitizar, quedó %d", sanitizado.Page)
	}
	if sanitizado.Gender != "" {
		t.Fatalf("un gender fuera de la allowlist debe quedar vacío, quedó %q", sanitizado.Gender)
	}
	if !esBusquedaReal(sanitizado) {
		t.Error("una búsqueda con texto debe seguir siéndolo tras sanitizar")
	}
}

// api/internal/request_structs/response_request.go

// Package request_structs define los objetos de transferencia de datos (DTOs).
package request_structs

// SpecialtyStats representa un resumen cuantitativo del catálogo de especialidades.
// Se utiliza principalmente en paneles de métricas y dashboards administrativos
// para proporcionar una visión rápida del estado del sistema sin exponer datos individuales.
type SpecialtyStats struct {
	// Total es la sumatoria absoluta de todos los registros en la tabla,
	// incluyendo activos e inactivos.
	Total int64 `json:"total" example:"45"`

	// Active representa la cantidad de especialidades visibles para el público general.
	Active int64 `json:"active" example:"40"`

	// Inactive representa la cantidad de especialidades que han sido dadas de baja
	// o que aún no han sido publicadas.
	Inactive int64 `json:"inactive" example:"5"`
}

// api/internal/request_structs/response_request.go

// Package request_structs define los objetos de transferencia de datos (DTOs)
// y View Models que estructuran los contratos de entrada y salida de la API.
package request_structs

// SpecialtyStats es un View Model de respuesta diseñado para la agregación de métricas (Data Aggregation).
//
// Propósito Arquitectónico y Rendimiento:
// Alimenta los indicadores clave de rendimiento (KPIs) y gráficos del dashboard administrativo.
// Delegar el conteo a la base de datos (mediante consultas COUNT optimizadas) y enviar
// únicamente esta estructura ligera (Stats) evita que el Frontend tenga que descargar
// arreglos masivos de objetos completos a través de la red solo para contarlos en memoria.
type SpecialtyStats struct {
	// Total representa el volumen absoluto del catálogo histórico.
	// Métricamente útil para auditar el tamaño de la base de datos y medir el crecimiento
	// de la taxonomía del sistema a lo largo del tiempo.
	Total int64 `json:"total" example:"45"`

	// Active cuantifica las áreas de desempeño clínico actualmente operativas.
	// Representa las especialidades que están indexadas y visibles para el público
	// en los motores de búsqueda del directorio de psicólogos.
	Active int64 `json:"active" example:"40"`

	// Inactive expone la cantidad de registros deshabilitados, deprecados o en estado de borrador.
	// Monitorear este valor (BI) permite a los administradores saber si existe
	// "deuda técnica" en el catálogo que requiera una depuración o limpieza profunda.
	Inactive int64 `json:"inactive" example:"5"`
}

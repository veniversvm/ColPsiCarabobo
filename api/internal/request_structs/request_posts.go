// api/internal/request_structs/request_posts.go

// Package request_structs define los DTOs (Data Transfer Objects) que actúan como
// la primera línea de defensa y validación de los datos que entran a la API.
package request_structs

// =========================================================================
// MÓDULO DE PUBLICACIONES Y NOTICIAS (CMS)
// =========================================================================

// CreatePostRequest define la carga útil exacta requerida para publicar una noticia.
// Tip Senior: Al combinar los tags 'form' y 'json', permitimos que el endpoint
// sea consumido mediante application/json (si no hay imagen) o multipart/form-data
// (si incluye un archivo adjunto) con exactamente el mismo código en el Handler.
type CreatePostRequest struct {
	Title            string `form:"title" json:"title" validate:"required,max=100" example:"Nuevo Reglamento de Honorarios"`
	ShortDescription string `form:"short_description" json:"short_description" validate:"max=250" example:"Resumen de los ajustes arancelarios para el nuevo trimestre."`
	Content          string `form:"content" json:"content" validate:"required" example:"<p>Estimados colegas, adjunto el contenido completo...</p>"`
	// Type restringe la visibilidad de la publicación a nivel de base de datos.
	Type     string `form:"type" json:"type" validate:"required,oneof=public psi" example:"public"`
	IsActive bool   `form:"is_active" json:"is_active" example:"true"`
}

// UpdatePostRequest es el DTO para operaciones de actualización parcial (PATCH).
// Todos los campos son punteros para permitir que el cliente envíe solo los campos
// que desea modificar.
type UpdatePostRequest struct {
	// 'omitempty' es crucial aquí: le dice al validador "si el puntero es nil, no apliques la regla max=100".
	Title            *string `form:"title" json:"title" validate:"omitempty,max=100"`
	ShortDescription *string `form:"short_description" json:"short_description" validate:"omitempty,max=250"`
	Content          *string `form:"content" json:"content"`
	Type             *string `form:"type" json:"type" validate:"omitempty,oneof=public psi"`
	IsActive         *bool   `form:"is_active" json:"is_active"`

	// Nota de Arquitectura: La imagen (file) se maneja intencionalmente fuera de este DTO.
	// Se extrae directamente en el Handler con c.FormFile("image") porque los archivos
	// no pueden ser mapeados limpiamente a structs genéricos mediante JSON.
}

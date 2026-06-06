// api/internal/request_structs/request_posts.go

// Package request_structs define los DTOs (Data Transfer Objects) que actúan como
// la primera línea de defensa y validación de los datos que entran a la API.
//
// Implementan el patrón de Capa Anticorrupción: aseguran que solo datos
// bien formados, sanitizados y validados alcancen la lógica de negocio,
// aislando al dominio de los protocolos de transporte HTTP.
package request_structs

import "time"

// =========================================================================
// MÓDULO DE PUBLICACIONES Y NOTICIAS (CMS)
// =========================================================================

// CreatePostRequest define la carga útil exacta requerida para publicar una noticia
// o comunicado institucional en el Content Management System (CMS).
//
// Flexibilidad de Transporte (Decisión de Diseño):
// La combinación de tags 'form' y 'json' permite que el mismo endpoint consuma
// peticiones application/json (cargas puramente textuales) o multipart/form-data
// (cuando se adjunta una imagen de portada), reutilizando el mismo validador y handler.
type CreatePostRequest struct {
	Title            string `form:"title" json:"title" validate:"required,max=100" example:"Nuevo Reglamento de Honorarios"`
	ShortDescription string `form:"short_description" json:"short_description" validate:"max=250" example:"Resumen de los ajustes arancelarios para el nuevo trimestre."`
	Content          string `form:"content" json:"content" validate:"required" example:"<p>Estimados colegas, adjunto el contenido completo...</p>"`

	// Type implementa el control de acceso basado en roles (RBAC) a nivel de contenido.
	// 'public' (visible para todo internet) o 'psi' (solo para colegiados autenticados).
	Type string `form:"type" json:"type" validate:"required,oneof=public psi" example:"public"`

	// Máquina de estados del ciclo de vida de la publicación.
	Status    string     `form:"status" json:"status" validate:"required,oneof=draft published archived scheduled"`
	PublishAt *time.Time `form:"publish_at" json:"publish_at,omitempty"`
}

// UpdatePostRequest representa el DTO para operaciones de actualización parcial (PATCH).
//
// Semántica PATCH Estricta:
// Todos los campos son punteros (*). Esto permite al framework de validación y a la base
// de datos diferenciar entre un valor explícitamente omitido (nil) y un valor
// intencionalmente enviado como vacío ("").
//
// Validación Condicional:
// El tag 'omitempty' en el validador es crucial; le indica a la librería validator/v10
// que si el puntero es nil, la regla subsecuente (ej. max=100) debe ignorarse.
// Esto evita falsos negativos en actualizaciones parciales.
type UpdatePostRequest struct {
	Title            *string    `form:"title" json:"title" validate:"omitempty,max=100"`
	ShortDescription *string    `form:"short_description" json:"short_description" validate:"omitempty,max=250"`
	Content          *string    `form:"content" json:"content"`
	Type             *string    `form:"type" json:"type" validate:"omitempty,oneof=public psi"`
	Status           *string    `form:"status" json:"status" validate:"omitempty,oneof=draft published archived scheduled"`
	PublishAt        *time.Time `form:"publish_at" json:"publish_at,omitempty"`

	// Nota de Arquitectura sobre Multimedia:
	// El archivo físico de la portada (ej. cover_image) se maneja intencionalmente
	// fuera de este DTO. Se extrae directamente en la capa del Handler (c.FormFile)
	// ya que los flujos binarios multipart no se mapean eficientemente a structs
	// estáticos bidireccionales en Go.
}

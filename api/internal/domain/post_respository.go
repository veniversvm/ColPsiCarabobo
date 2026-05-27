// api/internal/domain/post_respository.go
// Package domain define las entidades de negocio y los contratos de abstracción.
package domain

import (
	"context"

	"github.com/google/uuid"
)

// PostFilter define los criterios de búsqueda y segmentación para las publicaciones.
// Se utiliza para filtrar contenido en el directorio público, el portal de agremiados
// y el panel administrativo.
type PostFilter struct {
	// Type indica la visibilidad del post: "public", "psi" o "" para ignorar el filtro.
	Type string

	// IsActive es un puntero a booleano para permitir lógica tri-estatal:
	// - true: solo activos (publicados).
	// - false: solo inactivos (borradores).
	// - nil: todos los estados (usado por administradores).
	Status []PostStatus // reemplaza IsActive *bool

	// Search permite realizar búsquedas de texto parcial (fuzzy search) sobre el título.
	Search string
}

// PostRepository define el contrato de persistencia para el módulo de publicaciones.
// Implementa la estrategia de separación de contenido: la metadata reside en la tabla 'posts'
// y el contenido extenso en 'text_models', optimizando el rendimiento de listados masivos.
type PostRepository interface {
	// Create registra una nueva publicación y su contenido asociado en una operación atómica.
	// Requiere tanto el modelo de metadatos (post) como el de contenido (content).
	Create(ctx context.Context, post *Post, content *TextModel) error

	// Update modifica una publicación existente. Permite actualizar los metadatos
	// y, opcionalmente, el contenido extenso si 'content' no es nil.
	Update(ctx context.Context, post *Post, content *TextModel) error

	// Delete realiza un borrado lógico del post basado en su identificador único.
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID recupera una publicación completa, realizando un Eager Loading (Preload)
	// del contenido almacenado en TextModel.
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)

	// List retorna una colección paginada de publicaciones filtradas.
	// Optimización Senior: Este método normalmente no carga el 'TextModel' asociado
	// para ahorrar ancho de banda y memoria en la base de datos durante el scroll.
	List(ctx context.Context, filter PostFilter, page, limit int) ([]Post, int64, error)

	// PublishScheduled actualiza a 'published' todos los posts programados cuya fecha ya pasó.
	// Retorna el número de filas afectadas.
	PublishScheduled(ctx context.Context) int64

	// solo para uso de Google SEO
	GetSitemapPosts(ctx context.Context) ([]Post, error)
}

// Package postgres proporciona la implementación de persistencia para el dominio de publicaciones
// utilizando el motor PostgreSQL a través de GORM.
//
// # Arquitectura de datos
//
// Este paquete separa intencionalmente los datos en dos modelos complementarios:
//   - [domain.Post]: almacena la metadata ligera de cada publicación (título, tipo, estado, etc.).
//   - [domain.TextModel]: almacena el contenido extenso (cuerpo HTML/Markdown).
//
// Esta separación optimiza los listados masivos, ya que el frontend puede paginar y filtrar
// publicaciones sin transferir el cuerpo completo de cada una.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// postRepo implementa la interfaz [domain.PostRepository] sobre PostgreSQL mediante GORM.
// No debe instanciarse directamente; usar [NewPostRepository].
type postRepo struct {
	db *gorm.DB
}

// NewPostRepository construye e inicializa un repositorio de publicaciones listo para usarse.
// Recibe una conexión GORM activa y retorna la interfaz [domain.PostRepository],
// ocultando los detalles de implementación al resto de la aplicación.
func NewPostRepository(db *gorm.DB) domain.PostRepository {
	return &postRepo{db: db}
}

// Create registra una nueva publicación junto con su contenido de forma atómica.
//
// El proceso se ejecuta en una única transacción de base de datos con el siguiente orden:
//  1. Persiste el [domain.TextModel] (contenido extenso) y obtiene su ID generado.
//  2. Asigna ese ID al campo TextID de [domain.Post] y lo persiste.
//
// Si cualquiera de los dos pasos falla, la transacción se revierte completamente,
// garantizando que nunca exista un Post sin su TextModel correspondiente.
func (r *postRepo) Create(ctx context.Context, post *domain.Post, text *domain.TextModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Paso 1: persistir el contenido extenso para obtener su ID autogenerado.
		if err := tx.Create(text).Error; err != nil {
			return err
		}

		// Paso 2: vincular el ID del texto a la metadata y persistir el post.
		post.TextID = text.ID
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetByID recupera una publicación por su identificador único, incluyendo su contenido completo.
//
// Utiliza Eager Loading mediante GORM Preload para cargar el [domain.TextModel] asociado
// en una única consulta SQL eficiente, evitando el problema N+1.
// Retorna un error wrapeado de GORM si el registro no existe (gorm.ErrRecordNotFound).
func (r *postRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	var post domain.Post
	err := r.db.WithContext(ctx).Preload("Text").First(&post, "id = ?", id).Error
	return &post, err
}

// List busca y pagina publicaciones aplicando los filtros indicados en [domain.PostFilter].
//
// Comportamiento de los filtros (todos son opcionales y acumulables):
//   - Status: filtra por uno o más estados (ej. "published", "draft").
//   - Type "all_visible": retorna publicaciones de tipo "public" y "psi" (para psicólogos autenticados).
//   - Type <otro valor>: filtra por ese tipo exacto.
//   - Search: búsqueda case-insensitive por coincidencia parcial en el título (ILIKE).
//
// Nota de rendimiento: este método deliberadamente NO carga el [domain.TextModel],
// reduciendo el volumen de datos transferidos durante el renderizado de listados en el frontend.
//
// Retorna el slice de publicaciones de la página solicitada, el total de registros
// que coinciden con el filtro (para calcular la paginación en el cliente) y un error si ocurre.
func (r *postRepo) List(ctx context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Post{})

	// Filtro 1: uno o más estados permitidos.
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	// Filtro 2: visibilidad basada en el rol del solicitante (RBAC de contenido).
	if filter.Type != "" {
		if filter.Type == "all_visible" {
			// Psicólogos autenticados pueden acceder a contenido público y gremial.
			query = query.Where("type IN ?", []string{"public", "psi"})
		} else {
			query = query.Where("type = ?", filter.Type)
		}
	}

	// Filtro 3: búsqueda parcial e insensible a mayúsculas sobre el título.
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("title ILIKE ?", s)
	}

	// Conteo total de resultados para que el frontend pueda construir la paginación.
	query.Count(&total)

	// Consulta paginada ordenada del más reciente al más antiguo.
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&posts).Error

	return posts, total, err
}

// Update modifica de forma transaccional los datos de una publicación existente.
//
// Utiliza Updates() con mapas explícitos para proteger strings contra zero-value overwrites.
// Admite actualizaciones parciales:
//   - Siempre actualiza la metadata del [domain.Post] (título, estado, imagen, etc.).
//   - Solo actualiza el [domain.TextModel] si el parámetro text no es nil.
//
// La operación completa se revierte si cualquiera de las actualizaciones falla.
func (r *postRepo) Update(ctx context.Context, post *domain.Post, text *domain.TextModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(post).Updates(map[string]interface{}{
			"title":             post.Title,
			"short_description": post.ShortDescription,
			"type":              post.Type,
			"image_s3_key":      post.ImageS3Key,
			"status":            post.Status,
			"publish_at":        post.PublishAt,
			"update_by":         post.UpdateBy,
			"update_by_id":      post.UpdateById,
		}).Error; err != nil {
			return err
		}

		if text != nil {
			if err := tx.Model(text).Updates(map[string]interface{}{
				"content": text.Content,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete aplica un borrado lógico (soft delete) sobre la publicación indicada.
//
// Gracias al campo DeletedAt heredado de AuditFields, GORM marca el registro
// con la fecha de eliminación en lugar de borrarlo físicamente. El registro
// deja de aparecer en todas las consultas estándar pero puede recuperarse si es necesario.
func (r *postRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Post{}, "id = ?", id).Error
}

// PublishScheduled publica automáticamente todas las publicaciones programadas cuya
// fecha de publicación (publish_at) ya ha llegado o pasado.
//
// Este método está diseñado para ser invocado periódicamente por un job o cron interno.
// Transiciona el estado de [domain.PostStatusScheduled] a [domain.PostStatusPublished]
// y limpia el campo publish_at para reflejar que ya no hay una programación pendiente.
//
// Retorna la cantidad de filas afectadas, útil para registrar la actividad del job.
func (r *postRepo) PublishScheduled(ctx context.Context) int64 {
	result := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("status = ? AND publish_at <= ?", domain.PostStatusScheduled, time.Now()).
		Updates(map[string]interface{}{
			"status":     domain.PostStatusPublished,
			"publish_at": nil,
			"updated_at": time.Now(),
		})
	return result.RowsAffected
}

// GetSitemapPosts retorna el subconjunto mínimo de publicaciones necesario para generar
// el sitemap XML del sitio.
//
// Criterios de inclusión:
//   - Estado: únicamente publicaciones con status "published".
//   - Tipo: únicamente publicaciones de tipo "public" (se excluye el contenido gremial "psi").
//
// Proyección: solo se seleccionan los campos id, title y updated_at para minimizar
// el volumen de datos transferidos desde la base de datos.
//
// Retorna un error descriptivo wrapeado con fmt.Errorf si la consulta falla.
func (r *postRepo) GetSitemapPosts(ctx context.Context) ([]domain.Post, error) {
	var posts []domain.Post

	err := r.db.WithContext(ctx).
		Select("id, title, updated_at").
		Where("status = ?", "published").
		Where("type = ?", "public").
		Order("created_at DESC").
		Find(&posts).Error

	if err != nil {
		return nil, fmt.Errorf("error al obtener posts para sitemap: %w", err)
	}

	return posts, nil
}

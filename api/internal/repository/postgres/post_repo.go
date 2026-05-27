// api/internal/repository/postgres/post_repo.go

// Package postgres proporciona la implementación de persistencia para el dominio de publicaciones
// utilizando el motor PostgreSQL a través de GORM.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// postRepo implementa la interfaz domain.PostRepository.
// Utiliza una arquitectura de base de datos relacional para separar la metadata ligera (Post)
// del contenido pesado (TextModel), mejorando la velocidad de los listados masivos.
type postRepo struct {
	db *gorm.DB
}

// NewPostRepository crea una nueva instancia del repositorio de publicaciones.
func NewPostRepository(db *gorm.DB) domain.PostRepository {
	return &postRepo{db: db}
}

// Create registra una nueva publicación y su contenido extenso de forma atómica.
// Utiliza una transacción para asegurar que no se cree un Post sin su TextModel correspondiente,
// manteniendo la integridad referencial del sistema.
func (r *postRepo) Create(ctx context.Context, post *domain.Post, text *domain.TextModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Persistir el contenido extenso primero
		if err := tx.Create(text).Error; err != nil {
			return err
		}

		// 2. Vincular el ID del texto generado a la metadata del post
		post.TextID = text.ID
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetByID recupera una publicación específica junto con su contenido relacionado.
// Utiliza 'Preload' (Eager Loading) para obtener el TextModel en una sola operación eficiente.
func (r *postRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	var post domain.Post
	// Preload carga automáticamente la struct relacionada definida en el modelo.
	err := r.db.WithContext(ctx).Preload("Text").First(&post, "id = ?", id).Error
	return &post, err
}

// List implementa un buscador de publicaciones altamente optimizado con paginación y filtros.
// Nota de Arquitectura: Este método deliberadamente NO carga el 'TextModel' (el contenido largo)
// para minimizar el consumo de memoria y ancho de banda durante el renderizado de listados en el frontend.
func (r *postRepo) List(ctx context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64

	// Iniciamos la consulta sobre el modelo de metadatos
	query := r.db.WithContext(ctx).Model(&domain.Post{})

	// 1. Filtro de estado (Publicado / Borrador)
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	// 2. Filtro de visibilidad (RBAC de contenido)
	if filter.Type != "" {
		if filter.Type == "all_visible" {
			// Los psicólogos autenticados pueden ver tanto noticias públicas como gremiales.
			query = query.Where("type IN ?", []string{"public", "psi"})
		} else {
			query = query.Where("type = ?", filter.Type)
		}
	}

	// 3. Búsqueda por título (Fuzzy Search)
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("title ILIKE ?", s)
	}

	// 4. Conteo total para gestión de paginación en UI
	query.Count(&total)

	// 5. Ejecución de consulta paginada ordenada por fecha de creación
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&posts).Error

	return posts, total, err
}

// Update modifica los datos de una publicación existente de forma transaccional.
// Permite actualizaciones parciales: si el objeto 'text' es nil, solo se actualiza la metadata.
func (r *postRepo) Update(ctx context.Context, post *domain.Post, text *domain.TextModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Actualizar metadatos (título, descripción, imagen, etc.)
		if err := tx.Save(post).Error; err != nil {
			return err
		}

		// Actualizar el contenido extenso solo si se proporciona un nuevo modelo
		if text != nil {
			if err := tx.Save(text).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete realiza un borrado lógico del post.
// Nota: Debido a la herencia de AuditFields, GORM aplicará Soft-Delete automáticamente.
func (r *postRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Post{}, "id = ?", id).Error
}

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

// api/internal/repository/postgres/post_repo.go

func (r *postRepo) GetSitemapPosts(ctx context.Context) ([]domain.Post, error) {
	var posts []domain.Post

	err := r.db.WithContext(ctx).
		Select("id, title, updated_at").  // Campos para el sitemap
		Where("status = ?", "published"). // Solo los publicados (no borradores)
		Where("type = ?", "public").      // 👈 AÑADIDO: Solo los de tipo 'public'
		Order("created_at DESC").
		Find(&posts).Error

	if err != nil {
		// Retornar un error más descriptivo si es necesario
		return nil, fmt.Errorf("error al obtener posts para sitemap: %w", err)
	}

	return posts, nil
}

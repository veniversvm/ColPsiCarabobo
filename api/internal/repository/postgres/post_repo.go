package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

type postRepo struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) domain.PostRepository {
	return &postRepo{db: db}
}

func (r *postRepo) Create(ctx context.Context, post *domain.Post, text *domain.TextModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Crear el contenido pesado primero
		if err := tx.Create(text).Error; err != nil {
			return err
		}
		// 2. Vincular y crear el post
		post.TextID = text.ID
		if err := tx.Create(post).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *postRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	var post domain.Post
	// Preload("Text") carga el contenido pesado solo cuando pedimos el detalle
	err := r.db.WithContext(ctx).Preload("Text").First(&post, "id = ?", id).Error
	return &post, err
}

func (r *postRepo) List(ctx context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64

	// Optimizamos: NO hacemos Preload("Text") en el listado para que sea ligero
	query := r.db.WithContext(ctx).Model(&domain.Post{})

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	// Lógica de tipos: "public", "psi" o ambos
	if filter.Type != "" {
		if filter.Type == "all_visible" {
			// Caso especial para psicólogos: ven public y psi
			query = query.Where("type IN ?", []string{"public", "psi"})
		} else {
			query = query.Where("type = ?", filter.Type)
		}
	}

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("title ILIKE ?", s)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&posts).Error

	return posts, total, err
}

// ... (Update y Delete siguen la lógica estándar de GORM con Transaction)
func (r *postRepo) Update(ctx context.Context, post *domain.Post, text *domain.TextModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(post).Error; err != nil {
			return err
		}
		if text != nil {
			if err := tx.Save(text).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *postRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Post{}, "id = ?", id).Error
}

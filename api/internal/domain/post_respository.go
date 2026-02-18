package domain

import (
	"context"

	"github.com/google/uuid"
)

type PostFilter struct {
	Type     string // "public", "psi", "" (todos)
	IsActive *bool  // true, false, nil (todos)
	Search   string // Búsqueda por título
}

type PostRepository interface {
	Create(ctx context.Context, post *Post, content *TextModel) error
	Update(ctx context.Context, post *Post, content *TextModel) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
	List(ctx context.Context, filter PostFilter, page, limit int) ([]Post, int64, error)
}

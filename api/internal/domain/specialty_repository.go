package domain

import (
	"context"
)

type SpecialtyRepository interface {
	Create(ctx context.Context, s *PsiSpecialtyModel) error
	GetAll(ctx context.Context, status string) ([]PsiSpecialtyModel, error)
	GetByID(ctx context.Context, id uint32) (*PsiSpecialtyModel, error)
	Update(ctx context.Context, s *PsiSpecialtyModel) error
	Delete(ctx context.Context, id uint32) error
	Count(ctx context.Context, actives *bool) (int64, error)      // Método adicional para contar especialidades activas
	GetAllAdmin(ctx context.Context) ([]PsiSpecialtyModel, error) // Método para obtener todas las especialidades sin filtrar por estado
}

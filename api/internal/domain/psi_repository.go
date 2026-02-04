package domain

import (
	"context"

	"github.com/google/uuid"
)

type PsiUserRepository interface {
	// Atomic transaction to create User + ColData
	CreateWithColData(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData) error
	GetByID(ctx context.Context, id uuid.UUID) (*PsiUserModel, error)
	Search(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]PsiUserModel, int64, error)
	Update(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData) error
}

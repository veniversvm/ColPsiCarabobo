package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

type PsiUserRepository interface {
	// Atomic transaction to create User + ColData
	CreateWithColData(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData) error
	GetByID(ctx context.Context, id uuid.UUID) (*PsiUserModel, error)
	Search(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]PsiUserModel, int64, error)
	Update(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData) error
	UpdatePublicProfile(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData) error
	GetPsiUserColData(ctx context.Context, psiID uuid.UUID) (*PsiUserColData, error)
	SearchDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]PsiUserModel, int64, error)
	GetByIdentifier(ctx context.Context, identifier string) (*PsiUserModel, error)
	UpdateKey(ctx context.Context, psi *PsiUserModel) error
	Delete(ctx context.Context, id uuid.UUID) error
	// Postgrado CRUD
	CreatePostGrade(ctx context.Context, pg *PsiUserPostGrade) error
	GetPostGradeByID(ctx context.Context, id uuid.UUID) (*PsiUserPostGrade, error)
	UpdatePostGrade(ctx context.Context, pg *PsiUserPostGrade) error
	// SOCIAL MEDIA
	CreateSocialNetwork(ctx context.Context, sn *PsiUserSocialNetwork) error
	GetSocialNetworkByID(ctx context.Context, id uuid.UUID) (*PsiUserSocialNetwork, error)
	UpdateSocialNetwork(ctx context.Context, sn *PsiUserSocialNetwork) error
	DeleteSocialNetwork(ctx context.Context, id uuid.UUID) error
	CountSocialNetworksByPsiID(ctx context.Context, psiID uuid.UUID) (int64, error)
	SearchAdmin(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]PsiUserModel, int64, error)
}

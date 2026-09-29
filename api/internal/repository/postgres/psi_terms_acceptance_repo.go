// api/internal/repository/postgres/psi_terms_acceptance_repo.go
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// psiTermsAcceptanceRepo implementa domain.PsiTermsAcceptanceRepository.
//
// Deliberadamente NO expone Update ni Delete: la tabla es append-only. Cualquier
// modificación del histórico de aceptaciones tendría que venir de una migración
// escrita a mano, y por eso el código de aplicación no ofrece la vía.
type psiTermsAcceptanceRepo struct {
	db *gorm.DB
}

// NewPsiTermsAcceptanceRepository crea el repositorio de aceptaciones de términos.
func NewPsiTermsAcceptanceRepository(db *gorm.DB) domain.PsiTermsAcceptanceRepository {
	return &psiTermsAcceptanceRepo{db: db}
}

// Create inserta una aceptación.
func (r *psiTermsAcceptanceRepo) Create(ctx context.Context, acceptance *domain.PsiTermsAcceptance) error {
	return r.db.WithContext(ctx).Create(acceptance).Error
}

// GetByUserAndVersion devuelve la aceptación de ese agremiado para esa versión
// exacta; (nil, nil) si no existe.
func (r *psiTermsAcceptanceRepo) GetByUserAndVersion(
	ctx context.Context,
	psiUserID uuid.UUID,
	version string,
) (*domain.PsiTermsAcceptance, error) {
	var acceptance domain.PsiTermsAcceptance
	err := r.db.WithContext(ctx).
		Where("psi_user_id = ? AND version = ?", psiUserID, version).
		First(&acceptance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &acceptance, nil
}

// ListByUser devuelve el historial completo del agremiado, más reciente primero.
func (r *psiTermsAcceptanceRepo) ListByUser(ctx context.Context, psiUserID uuid.UUID) ([]domain.PsiTermsAcceptance, error) {
	var acceptances []domain.PsiTermsAcceptance
	err := r.db.WithContext(ctx).
		Where("psi_user_id = ?", psiUserID).
		Order("created_at DESC").
		Find(&acceptances).Error
	return acceptances, err
}

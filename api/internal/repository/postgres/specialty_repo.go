package postgres

import (
	"context"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

type specialtyRepo struct {
	db *gorm.DB
}

func NewSpecialtyRepository(db *gorm.DB) domain.SpecialtyRepository {
	return &specialtyRepo{db: db}
}

func (r *specialtyRepo) Create(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *specialtyRepo) GetAll(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
	var list []domain.PsiSpecialtyModel
	query := r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{}).Order("name asc")

	switch status {
	case "active":
		query = query.Where("active = ?", true)
	case "inactive":
		query = query.Where("active = ?", false)
		// case "all": no aplicamos filtro
	}

	err := query.Find(&list).Error
	return list, err
}

func (r *specialtyRepo) GetByID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	var s domain.PsiSpecialtyModel
	err := r.db.WithContext(ctx).First(&s, "id = ? AND active = ?", id, true).Error
	return &s, err
}

func (r *specialtyRepo) Update(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *specialtyRepo) Delete(ctx context.Context, id uint32) error {
	// Soft delete manual usando el flag Active además del deleted_at de AuditModel
	return r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{}).Where("id = ?", id).
		Updates(map[string]interface{}{"active": false}).Error
}

func (r *specialtyRepo) Count(ctx context.Context, active *bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{})

	// Si active no es nil, aplicamos el filtro.
	// Si es nil, GORM ignorará esta línea y contará TODO.
	if active != nil {
		query = query.Where("active = ?", *active)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *specialtyRepo) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {
	var list []domain.PsiSpecialtyModel
	// Simplemente ordenamos por nombre, sin filtrar por el campo 'active'
	err := r.db.WithContext(ctx).Order("name asc").Find(&list).Error
	return list, err
}

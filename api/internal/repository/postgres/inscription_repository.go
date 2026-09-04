// api/internal/repository/postgres/inscription_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"gorm.io/gorm"
)

// inscriptionRepo es la implementación GORM del repositorio de inscripciones.
type inscriptionRepo struct {
	db *gorm.DB
}

// NewInscriptionRepository crea una nueva instancia del repositorio de inscripciones.
func NewInscriptionRepository(db *gorm.DB) domain.InscriptionRepository {
	return &inscriptionRepo{db: db}
}

// Create inserta una nueva solicitud de pre-inscripción.
func (r *inscriptionRepo) Create(ctx context.Context, req *domain.PsiInscriptionRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

// GetByID recupera una solicitud por su UUID.
func (r *inscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiInscriptionRequest, error) {
	var req domain.PsiInscriptionRequest
	err := r.db.WithContext(ctx).First(&req, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("solicitud de inscripción no encontrada: %w", err)
		}
		return nil, err
	}
	return &req, nil
}

// Search lista solicitudes con filtros y paginación.
func (r *inscriptionRepo) Search(ctx context.Context, filter request_structs.InscriptionListFilter) ([]domain.PsiInscriptionRequest, int64, error) {
	var items []domain.PsiInscriptionRequest
	var total int64

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	q := r.db.WithContext(ctx).Model(&domain.PsiInscriptionRequest{})

	if filter.Status != "" && filter.Status != "all" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Q != "" {
		like := "%" + filter.Q + "%"
		q = q.Where("nombres ILIKE ? OR apellidos ILIKE ? OR CAST(cedula AS TEXT) ILIKE ?", like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ExistsPendingCI retorna true si existe una solicitud pendiente con esa cédula.
func (r *inscriptionRepo) ExistsPendingCI(ctx context.Context, ci int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PsiInscriptionRequest{}).
		Where("cedula = ? AND status = ?", ci, domain.InscriptionPending).
		Count(&count).Error
	return count > 0, err
}

// ExistsPendingFPV retorna true si existe una solicitud pendiente con ese FPV.
func (r *inscriptionRepo) ExistsPendingFPV(ctx context.Context, fpv int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PsiInscriptionRequest{}).
		Where("fpv = ? AND status = ? AND fpv IS NOT NULL", fpv, domain.InscriptionPending).
		Count(&count).Error
	return count > 0, err
}

// CIInPsiUsers retorna si la cédula ya está registrada en psi_users.
func (r *inscriptionRepo) CIInPsiUsers(ctx context.Context, ci int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("psi_users").
		Where("ci = ?", ci).
		Count(&count).Error
	return count > 0, err
}

// FPVInPsiUsers retorna si el FPV ya está registrado en psi_users.
func (r *inscriptionRepo) FPVInPsiUsers(ctx context.Context, fpv int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("psi_users").
		Where("fpv = ?", fpv).
		Count(&count).Error
	return count > 0, err
}

// NextControlNumber calcula el siguiente número de control secuencial
// basado en el MAX(control_number numérico) de psi_users + 1.
func (r *inscriptionRepo) NextControlNumber(ctx context.Context) (int, error) {
	var max int64
	err := r.db.WithContext(ctx).Table("psi_users").
		Select("COALESCE(MAX(CAST(control_number AS INTEGER)), 0)").
		Where("control_number ~ '^[0-9]+$'").
		Scan(&max).Error
	if err != nil {
		return 0, err
	}
	return int(max) + 1, nil
}

// Update actualiza el estado de una solicitud (aprobación/rechazo).
func (r *inscriptionRepo) Update(ctx context.Context, req *domain.PsiInscriptionRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

// Delete elimina físicamente una solicitud.
func (r *inscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PsiInscriptionRequest{}, "id = ?", id).Error
}

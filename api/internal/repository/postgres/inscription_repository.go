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

// ExistsPendingCIExcluding retorna true si existe OTRA solicitud pendiente con esa cédula.
func (r *inscriptionRepo) ExistsPendingCIExcluding(ctx context.Context, ci int, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PsiInscriptionRequest{}).
		Where("id <> ? AND cedula = ? AND status = ?", excludeID, ci, domain.InscriptionPending).
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

// ExistsPendingFPVExcluding retorna true si existe OTRA solicitud pendiente con ese FPV.
func (r *inscriptionRepo) ExistsPendingFPVExcluding(ctx context.Context, fpv int, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PsiInscriptionRequest{}).
		Where("id <> ? AND fpv = ? AND status = ? AND fpv IS NOT NULL", excludeID, fpv, domain.InscriptionPending).
		Count(&count).Error
	return count > 0, err
}

// ExistsPendingEmail retorna true si existe una solicitud pendiente con ese correo.
func (r *inscriptionRepo) ExistsPendingEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PsiInscriptionRequest{}).
		Where("LOWER(correo) = LOWER(?) AND status = ?", email, domain.InscriptionPending).
		Count(&count).Error
	return count > 0, err
}

// ExistsPendingEmailExcluding retorna true si existe OTRA solicitud pendiente con ese correo.
func (r *inscriptionRepo) ExistsPendingEmailExcluding(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PsiInscriptionRequest{}).
		Where("id <> ? AND LOWER(correo) = LOWER(?) AND status = ?", excludeID, email, domain.InscriptionPending).
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

// EmailInPsiUsers retorna si el correo ya está registrado en psi_users.
// No filtra soft-deleted para respetar el constraint único uni_psi_users_email.
func (r *inscriptionRepo) EmailInPsiUsers(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("psi_users").
		Where("LOWER(email) = LOWER(?)", email).
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

// UpdateNotes actualiza solo las notas administrativas de una solicitud.
func (r *inscriptionRepo) UpdateNotes(ctx context.Context, id uuid.UUID, notes string) error {
	return r.db.WithContext(ctx).Model(&domain.PsiInscriptionRequest{}).
		Where("id = ?", id).
		Update("notes", notes).Error
}

// Delete elimina físicamente una solicitud.
func (r *inscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PsiInscriptionRequest{}, "id = ?", id).Error
}

// =========================================================================
// DOCUMENTOS DE LA FICHA
// =========================================================================

// CreateDocuments persiste las fotos de documentos de la solicitud.
func (r *inscriptionRepo) CreateDocuments(ctx context.Context, docs []domain.PsiInscriptionDocument) error {
	if len(docs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&docs).Error
}

// ListDocumentsByRequestID recupera las fotos de documentos de la solicitud,
// ordenadas de la más reciente a la más antigua.
func (r *inscriptionRepo) ListDocumentsByRequestID(ctx context.Context, requestID uuid.UUID) ([]domain.PsiInscriptionDocument, error) {
	var docs []domain.PsiInscriptionDocument
	err := r.db.WithContext(ctx).
		Where("inscription_request_id = ?", requestID).
		Order("created_at DESC").
		Find(&docs).Error
	return docs, err
}

// GetInscriptionDocumentByID busca una foto de documento por su UUID.
func (r *inscriptionRepo) GetInscriptionDocumentByID(ctx context.Context, id uuid.UUID) (*domain.PsiInscriptionDocument, error) {
	var doc domain.PsiInscriptionDocument
	err := r.db.WithContext(ctx).First(&doc, "id = ?", id).Error
	return &doc, err
}

// UpdateInscriptionDocument actualiza una foto de documento de la ficha.
func (r *inscriptionRepo) UpdateInscriptionDocument(ctx context.Context, doc *domain.PsiInscriptionDocument) error {
	return r.db.WithContext(ctx).Model(doc).Select("s3_key", "title", "notes", "original_filename").Updates(doc).Error
}

// DeleteInscriptionDocument elimina físicamente la foto de un documento.
func (r *inscriptionRepo) DeleteInscriptionDocument(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PsiInscriptionDocument{}, "id = ?", id).Error
}

// DeleteInscriptionDocumentsByRequestID elimina las fotos de documentos de una solicitud.
func (r *inscriptionRepo) DeleteInscriptionDocumentsByRequestID(ctx context.Context, requestID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&domain.PsiInscriptionDocument{}, "inscription_request_id = ?", requestID).Error
}

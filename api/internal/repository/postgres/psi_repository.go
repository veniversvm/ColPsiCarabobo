package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

type psiRepo struct {
	db *gorm.DB
}

// NewPsiRepository inyecta la conexión de GORM y devuelve la interfaz del dominio
func NewPsiRepository(db *gorm.DB) domain.PsiUserRepository {
	return &psiRepo{db: db}
}

// CreateWithColData realiza una inserción atómica de usuario y sus datos colegiales
func (r *psiRepo) CreateWithColData(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Crear el modelo principal del Psicólogo
		if err := tx.Create(psi).Error; err != nil {
			return fmt.Errorf("error creating psi user: %w", err)
		}

		// 2. Vincular el ID del usuario recién creado a los datos colegiales
		colData.PsiUserModelID = psi.ID
		if err := tx.Create(colData).Error; err != nil {
			return fmt.Errorf("error creating col data: %w", err)
		}

		// 3. Actualizar la referencia circular en el modelo principal
		return tx.Model(psi).Update("psi_user_col_data_id", colData.ID).Error
	})
}

// GetByID recupera un psicólogo incluyendo sus relaciones (Eager Loading)
func (r *psiRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	var psi domain.PsiUserModel
	// Preload carga automáticamente las structs relacionadas en una sola consulta eficiente
	err := r.db.WithContext(ctx).
		Preload("ColData").
		Preload("PostGrades").
		First(&psi, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("psicólogo no encontrado: %w", err)
		}
		return nil, err
	}

	return &psi, nil
}

// Search implementa búsqueda filtrada y paginación profesional
func (r *psiRepo) Search(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]domain.PsiUserModel, int64, error) {
	var psis []domain.PsiUserModel
	var total int64

	// Iniciamos la consulta sobre el modelo principal
	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{})

	// Aplicamos filtros dinámicos (basado en tu lógica anterior)
	if ci, ok := filters["ci"]; ok && ci != "" {
		query = query.Where("ci = ?", ci)
	}
	if fpv, ok := filters["fpv"]; ok && fpv != "" {
		query = query.Where("fpv = ?", fpv)
	}
	if name, ok := filters["name"]; ok && name != "" {
		search := "%" + name.(string) + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ?", search, search)
	}

	// 1. Contar el total de registros para la paginación del frontend
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2. Recuperar la página de datos con sus relaciones básicas
	offset := (page - 1) * pageSize
	err := query.Offset(offset).
		Limit(pageSize).
		Preload("ColData").
		Order("created_at DESC").
		Find(&psis).Error

	return psis, total, err
}

// Update actualiza ambos modelos dentro de una transacción
func (r *psiRepo) Update(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Save actualiza todos los campos del modelo (incluyendo los de auditoría)
		if err := tx.Save(psi).Error; err != nil {
			return err
		}

		if colData != nil {
			if err := tx.Save(colData).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// api/internal/repository/postgres/psi_repository.go

// Package postgres provee la implementación concreta de los repositorios usando PostgreSQL y GORM.
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

// psiRepo implementa la interfaz domain.PsiUserRepository.
// Centraliza todas las interacciones con la base de datos relacionadas con los psicólogos,
// incluyendo su perfil, historial académico, solvencia y redes sociales.
type psiRepo struct {
	db *gorm.DB
}

// NewPsiRepository inyecta la conexión de GORM y devuelve la interfaz del dominio.
func NewPsiRepository(db *gorm.DB) domain.PsiUserRepository {
	return &psiRepo{db: db}
}

// =========================================================================
// GESTIÓN CORE DEL PSICÓLOGO
// =========================================================================

// CreateWithColData realiza una inserción atómica de usuario y sus datos colegiales.
// Utiliza una transacción para asegurar que no se cree un usuario sin sus datos base asociados.
func (r *psiRepo) CreateWithColData(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Crear el Psicólogo (Genera el ID y maneja auditoría básica)
		if err := tx.Create(psi).Error; err != nil {
			return fmt.Errorf("error creating psi user: %w", err)
		}

		// 2. Vincular los datos colegiales a la llave foránea del psicólogo recién creado
		colData.PsiUserModelID = psi.ID
		if err := tx.Create(colData).Error; err != nil {
			return fmt.Errorf("error creating col data: %w", err)
		}

		return nil
	})
}

// GetByID recupera un psicólogo incluyendo sus relaciones mediante Eager Loading.
// Optimiza la carga de postgrados ordenándolos cronológicamente de forma descendente.
func (r *psiRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	var psi domain.PsiUserModel

	err := r.db.WithContext(ctx).
		Preload("ColData").
		Preload("PostGrades", func(db *gorm.DB) *gorm.DB {
			return db.Order("graduation_year DESC")
		}).
		Preload("SocialNetworks").
		First(&psi, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("psicólogo no encontrado: %w", err)
		}
		return nil, err
	}

	return &psi, nil
}

// GetByIdentifier busca un psicólogo por su nombre de usuario o correo electrónico.
// Es una función crítica para los procesos de autenticación (Login).
func (r *psiRepo) GetByIdentifier(ctx context.Context, identifier string) (*domain.PsiUserModel, error) {
	var psi domain.PsiUserModel
	err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", identifier, identifier).
		First(&psi).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, err
	}
	return &psi, nil
}

// Delete realiza un borrado lógico (Soft Delete) del psicólogo.
// Al existir el campo DeletedAt en el modelo, GORM ejecuta un UPDATE en lugar de un DELETE físico.
func (r *psiRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PsiUserModel{}, "id = ?", id).Error
}

// =========================================================================
// ACTUALIZACIONES (MUTACIONES)
// =========================================================================

// Update actualiza tanto el perfil como los datos colegiales dentro de una transacción.
// Generalmente utilizado por administradores para ediciones globales.
func (r *psiRepo) Update(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

// UpdatePublicProfile actualiza los datos permitidos para edición por parte del usuario.
// Usa tx.Omit("ColData") para prevenir que GORM intente actualizar asociaciones no deseadas.
func (r *psiRepo) UpdatePublicProfile(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// FIX: Solo actualizar campos públicos permitidos (lista blanca)
		// Si usas Save(psi) te sobrescribe todo.
		if err := tx.Model(psi).Omit("CI", "FPV", "IsActive", "Solvent", "Password").Updates(psi).Error; err != nil {
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

// UpdateKey actualiza únicamente los campos relacionados con la sesión y auditoría.
// Mejora el rendimiento al usar Select() para limitar las columnas en la sentencia SQL.
func (r *psiRepo) UpdateKey(ctx context.Context, psi *domain.PsiUserModel) error {
	return r.db.WithContext(ctx).Model(psi).
		Select("Key", "UpdatedAt", "UpdateBy", "UpdateById").
		Updates(psi).Error
}

// GetPsiUserColData recupera exclusivamente la información colegial de un psicólogo.
// Útil para vistas ligeras donde no se requiere el perfil completo del usuario.
func (r *psiRepo) GetPsiUserColData(ctx context.Context, psiID uuid.UUID) (*domain.PsiUserColData, error) {
	var colData domain.PsiUserColData
	err := r.db.WithContext(ctx).First(&colData, "psi_user_model_id = ?", psiID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("datos colegiales no encontrados para el psicólogo: %w", err)
		}
		return nil, err
	}
	return &colData, nil
}

// =========================================================================
// MOTORES DE BÚSQUEDA Y ESTADÍSTICAS
// =========================================================================

// SearchDirectory implementa la lógica de búsqueda para el directorio público.
// Diferencia entre búsqueda por "Identidad" (exacta por CI/FPV) y "Navegación" (filtro por solvencia).
func (r *psiRepo) SearchDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	var users []domain.PsiUserModel
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
		Select("id, first_name, last_name, ci, fpv, profile_picture_s3_key, mini_bio, solvent, primary_specialty, secondary_specialty").
		Where("is_active = ?", true)

	if filter.SearchTerm != "" {
		// CASO IDENTIDAD: Búsqueda amplia por texto en campos clave.
		term := "%" + filter.SearchTerm + "%"
		query = query.Where(
			r.db.Where("first_name ILIKE ?", term).
				Or("last_name ILIKE ?", term).
				Or("CAST(ci AS TEXT) LIKE ?", term).
				Or("CAST(fpv AS TEXT) LIKE ?", term),
		)
	} else {
		// CASO NAVEGACIÓN: Restringe a usuarios solventes y aplica filtros de catálogo.
		query = query.Where("solvent = ?", true)

		if filter.SpecialtyID > 0 {
			var specName string
			r.db.Model(&domain.PsiSpecialtyModel{}).Select("name").Where("id = ?", filter.SpecialtyID).Scan(&specName)

			if specName != "" {
				query = query.Where("primary_specialty = ? OR secondary_specialty = ?", specName, specName)
			}
		}

		if filter.Location != "" {
			loc := "%" + filter.Location + "%"
			query = query.Where(
				r.db.Where("municipality_carabobo ILIKE ?", loc).
					Or("state_outside ILIKE ?", loc).
					// FIX: Corregido el nombre exacto de la columna generada por GORM
					Or("municipality_out_side_carabobo ILIKE ?", loc),
			)
		}

		if filter.Gender != "" {
			query = query.Where("genre = ?", filter.Gender)
		}
	}

	query.Count(&total)

	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("solvent DESC, last_name ASC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// SearchAdmin provee una búsqueda sin restricciones para el panel administrativo.
// Ignora estados de solvencia o privacidad para permitir la gestión total.
func (r *psiRepo) SearchAdmin(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	var users []domain.PsiUserModel
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
		Select("id, first_name, last_name, ci, fpv, email, solvent, is_active, primary_specialty, secondary_specialty")

	if filter.SearchTerm != "" {
		term := "%" + filter.SearchTerm + "%"
		query = query.Where(
			r.db.Where("first_name ILIKE ?", term).
				Or("last_name ILIKE ?", term).
				Or("CAST(ci AS TEXT) LIKE ?", term).
				Or("CAST(fpv AS TEXT) LIKE ?", term),
		)
	}

	if filter.SpecialtyID > 0 {
		var specName string
		r.db.Model(&domain.PsiSpecialtyModel{}).Select("name").Where("id = ?", filter.SpecialtyID).Scan(&specName)
		if specName != "" {
			query = query.Where("primary_specialty = ? OR secondary_specialty = ?", specName, specName)
		}
	}

	if filter.Location != "" {
		loc := "%" + filter.Location + "%"
		query = query.Where(
			r.db.Where("municipality_carabobo ILIKE ?", loc).
				Or("state_outside ILIKE ?", loc).
				// FIX: Corregido el nombre exacto de la columna generada por GORM
				Or("municipality_out_side_carabobo ILIKE ?", loc),
		)
	}

	if filter.Gender != "" {
		query = query.Where("genre = ?", filter.Gender)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// Count devuelve el número total de psicólogos, permitiendo filtrar por estado activo.
func (r *psiRepo) Count(ctx context.Context, active *bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{})

	if active != nil {
		query = query.Where("active = ?", *active)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("repo.Count: %w", err)
	}

	return count, nil
}

// Search implementa una búsqueda genérica mediante un mapa de filtros dinámicos.
func (r *psiRepo) Search(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]domain.PsiUserModel, int64, error) {
	var psis []domain.PsiUserModel
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{})

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
	if active, ok := filters["active"]; ok && active != nil {
		query = query.Where("active = ?", active)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).
		Limit(pageSize).
		Preload("ColData").
		Order("created_at DESC").
		Find(&psis).Error

	return psis, total, err
}

// =========================================================================
// MÓDULO ACADÉMICO (POSTGRADOS)
// =========================================================================

// CreatePostGrade registra un nuevo título o estudio de postgrado.
func (r *psiRepo) CreatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return r.db.WithContext(ctx).Create(pg).Error
}

// GetPostGradeByID recupera la información de un postgrado específico.
func (r *psiRepo) GetPostGradeByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserPostGrade, error) {
	var pg domain.PsiUserPostGrade
	err := r.db.WithContext(ctx).First(&pg, "id = ?", id).Error
	return &pg, err
}

// UpdatePostGrade actualiza los datos de un registro académico existente.
func (r *psiRepo) UpdatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return r.db.WithContext(ctx).Save(pg).Error
}

// =========================================================================
// PRESENCIA DIGITAL (REDES SOCIALES)
// =========================================================================

// CreateSocialNetwork vincula una nueva red social al perfil del psicólogo.
func (r *psiRepo) CreateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	return r.db.WithContext(ctx).Create(sn).Error
}

// GetSocialNetworkByID busca un registro de red social por su ID único.
func (r *psiRepo) GetSocialNetworkByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
	var sn domain.PsiUserSocialNetwork
	err := r.db.WithContext(ctx).First(&sn, "id = ?", id).Error
	return &sn, err
}

// UpdateSocialNetwork modifica el enlace o tipo de una red social existente.
func (r *psiRepo) UpdateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	return r.db.WithContext(ctx).Save(sn).Error
}

// DeleteSocialNetwork elimina una red social (aplica Soft Delete si el modelo lo soporta).
func (r *psiRepo) DeleteSocialNetwork(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PsiUserSocialNetwork{}, "id = ?", id).Error
}

// CountSocialNetworksByPsiID devuelve la cantidad de redes activas que tiene un usuario.
func (r *psiRepo) CountSocialNetworksByPsiID(ctx context.Context, psiID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.PsiUserSocialNetwork{}).
		Where("psi_user_id = ?", psiID).
		Count(&count).Error

	return count, err
}

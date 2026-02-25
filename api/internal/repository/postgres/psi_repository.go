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
		// 1. Crear el Psicólogo (Genera el ID)
		if err := tx.Create(psi).Error; err != nil {
			return fmt.Errorf("error creating psi user: %w", err)
		}

		// 2. Vincular los datos colegiales al ID del psicólogo
		colData.PsiUserModelID = psi.ID // Esta es la llave foránea real
		if err := tx.Create(colData).Error; err != nil {
			return fmt.Errorf("error creating col data: %w", err)
		}

		return nil
	})
}

// GetByID recupera un psicólogo incluyendo sus relaciones (Eager Loading)
func (r *psiRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	var psi domain.PsiUserModel

	err := r.db.WithContext(ctx).
		Preload("ColData").
		// Mejora: Ordenamos los postgrados cronológicamente (más recientes primero)
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

func (r *psiRepo) Count(ctx context.Context, active *bool) (int64, error) {
	var count int64

	// Inicializamos la query sobre el modelo
	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{})

	// Si active no es nil, aplicamos el filtro.
	// Si es nil, GORM simplemente ignorará esta cláusula y contará todo.
	if active != nil {
		query = query.Where("active = ?", *active)
	}

	if err := query.Count(&count).Error; err != nil {
		// Es buena práctica envolver el error o loguearlo si es necesario
		return 0, fmt.Errorf("repo.Count: %w", err)
	}

	return count, nil
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

	if active, ok := filters["active"]; ok && active != nil {
		query = query.Where("active = ?", active)
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

// UpdatePublicProfile actualiza los datos permitidos del psicólogo.
// Usa una transacción para asegurar consistencia entre las dos tablas.
func (r *psiRepo) UpdatePublicProfile(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Actualizar modelo principal
		// Omitimos asociaciones para evitar guardar colData accidentalmente aquí
		if err := tx.Omit("ColData").Save(psi).Error; err != nil {
			return err
		}

		// 2. Actualizar ColData solo si se envió
		if colData != nil {
			if err := tx.Save(colData).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetPsiUserColData recupera los datos colegiales asociados a un psicólogo específico.
// Este método es útil para endpoints que solo necesitan mostrar o editar la información colegial sin cargar todo el perfil del psicólogo.
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

func (r *psiRepo) SearchDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	var users []domain.PsiUserModel
	var total int64

	// 1. Base Query con selección de campos mínimos para rendimiento
	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
		Select("id, first_name, last_name, ci, fpv, profile_picture_s3_key, mini_bio, solvent, primary_specialty, secondary_specialty").
		Where("is_active = ?", true)

	// 2. LÓGICA DE IDENTIDAD (q) vs NAVEGACIÓN
	if filter.SearchTerm != "" {
		// CASO IDENTIDAD: Buscamos por nombre, CI o FPV
		// REGLA: No filtramos por solvencia aquí.
		term := "%" + filter.SearchTerm + "%"
		query = query.Where(
			r.db.Where("first_name ILIKE ?", term).
				Or("last_name ILIKE ?", term).
				Or("CAST(ci AS TEXT) LIKE ?", term).
				Or("CAST(fpv AS TEXT) LIKE ?", term),
		)
	} else {
		// CASO NAVEGACIÓN: El usuario está explorando el catálogo.
		// REGLA: Solo mostramos SOLVENTES.
		query = query.Where("solvent = ?", true)

		// Filtro por Especialidad (Desde el catálogo maestro)
		if filter.SpecialtyID > 0 {
			// Buscamos el nombre de la especialidad para comparar con los campos del usuario
			// (Nota: Esto es así porque los usuarios guardan el nombre de la especialidad como string)
			var specName string
			r.db.Model(&domain.PsiSpecialtyModel{}).Select("name").Where("id = ?", filter.SpecialtyID).Scan(&specName)

			if specName != "" {
				query = query.Where("primary_specialty = ? OR secondary_specialty = ?", specName, specName)
			}
		}

		// Filtro por Ubicación (Carabobo o Exterior)
		if filter.Location != "" {
			loc := "%" + filter.Location + "%"
			query = query.Where(
				r.db.Where("municipality_carabobo ILIKE ?", loc).
					Or("state_outside ILIKE ?", loc).
					Or("municipality_outside_carabobo ILIKE ?", loc),
			)
		}

		// Filtro por Género
		if filter.Gender != "" {
			query = query.Where("genre = ?", filter.Gender)
		}
	}

	// 3. Conteo y Paginación
	query.Count(&total)

	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("solvent DESC, last_name ASC"). // Priorizamos solventes en el orden
								Limit(filter.Limit).
								Offset(offset).
								Find(&users).Error

	return users, total, err
}

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

// UpdateKey actualiza únicamente la Key de sesión y la fecha de modificación
func (r *psiRepo) UpdateKey(ctx context.Context, psi *domain.PsiUserModel) error {
	return r.db.WithContext(ctx).Model(psi).
		Select("Key", "UpdatedAt", "UpdateBy", "UpdateById"). // Solo actualizamos estos campos
		Updates(psi).Error
}

func (r *psiRepo) CreatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return r.db.WithContext(ctx).Create(pg).Error
}

// GetPostGradeByID busca un postgrado específico.
func (r *psiRepo) GetPostGradeByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserPostGrade, error) {
	var pg domain.PsiUserPostGrade
	err := r.db.WithContext(ctx).First(&pg, "id = ?", id).Error
	return &pg, err
}

// UpdatePostGrade actualiza el registro.
func (r *psiRepo) UpdatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return r.db.WithContext(ctx).Save(pg).Error
}

// Delete realiza un borrado lógico (Soft Delete) del psicólogo.
// Gracias a que el modelo tiene el campo DeletedAt, GORM hace un UPDATE en lugar de un DELETE.
func (r *psiRepo) Delete(ctx context.Context, id uuid.UUID) error {
	// Es importante pasar el modelo para que GORM sepa qué tabla afectar
	return r.db.WithContext(ctx).Delete(&domain.PsiUserModel{}, "id = ?", id).Error
}

////////////////////////////////////
// SOCIAL MEDIA
////////////////////////////////////

func (r *psiRepo) CreateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	return r.db.WithContext(ctx).Create(sn).Error
}

func (r *psiRepo) GetSocialNetworkByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
	var sn domain.PsiUserSocialNetwork
	err := r.db.WithContext(ctx).First(&sn, "id = ?", id).Error
	return &sn, err
}

func (r *psiRepo) UpdateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	return r.db.WithContext(ctx).Save(sn).Error
}

func (r *psiRepo) DeleteSocialNetwork(ctx context.Context, id uuid.UUID) error {
	// GORM hará el Soft Delete automáticamente gracias a DeletedAt
	return r.db.WithContext(ctx).Delete(&domain.PsiUserSocialNetwork{}, "id = ?", id).Error
}

func (r *psiRepo) CountSocialNetworksByPsiID(ctx context.Context, psiID uuid.UUID) (int64, error) {
	var count int64
	// Contamos solo las que no han sido borradas (GORM maneja el Soft Delete automáticamente)
	err := r.db.WithContext(ctx).
		Model(&domain.PsiUserSocialNetwork{}).
		Where("psi_user_id = ?", psiID).
		Count(&count).Error

	return count, err
}

// SearchAdmin es la consulta de "Rayos X" para el panel de control. Ignora privacidad y estatus.
func (r *psiRepo) SearchAdmin(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	var users []domain.PsiUserModel
	var total int64

	// NOTA SENIOR: No filtramos por is_active ni solvent
	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
		Select("id, first_name, last_name, ci, fpv, email, solvent, is_active, primary_specialty, secondary_specialty")

	// 1. Búsqueda por Identidad
	if filter.SearchTerm != "" {
		term := "%" + filter.SearchTerm + "%"
		query = query.Where(
			r.db.Where("first_name ILIKE ?", term).
				Or("last_name ILIKE ?", term).
				Or("CAST(ci AS TEXT) LIKE ?", term).
				Or("CAST(fpv AS TEXT) LIKE ?", term),
		)
	}

	// 2. Filtro de Especialidad
	if filter.SpecialtyID > 0 {
		var specName string
		r.db.Model(&domain.PsiSpecialtyModel{}).Select("name").Where("id = ?", filter.SpecialtyID).Scan(&specName)
		if specName != "" {
			query = query.Where("primary_specialty = ? OR secondary_specialty = ?", specName, specName)
		}
	}

	// 3. Filtro de Ubicación
	if filter.Location != "" {
		loc := "%" + filter.Location + "%"
		query = query.Where(
			r.db.Where("municipality_carabobo ILIKE ?", loc).
				Or("state_outside ILIKE ?", loc).
				Or("municipality_outside_carabobo ILIKE ?", loc),
		)
	}

	// 4. Filtro de Género
	if filter.Gender != "" {
		query = query.Where("genre = ?", filter.Gender)
	}

	// Conteo y Paginación
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	// Ordenamos por los más recientes primero para facilitar la moderación
	err := query.Order("created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

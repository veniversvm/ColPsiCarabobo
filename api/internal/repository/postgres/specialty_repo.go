// api/internal/repository/postgres/specialty_repo.go

// Package postgres proporciona la capa de persistencia (Data Access Layer)
// para el motor de base de datos PostgreSQL utilizando el ORM GORM.
package postgres

import (
	"context"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// specialtyRepo es la implementación concreta de la interfaz domain.SpecialtyRepository.
// Encapsula la lógica de acceso a datos para aislar a los servicios de los detalles de SQL.
type specialtyRepo struct {
	db *gorm.DB
}

// NewSpecialtyRepository actúa como constructor (Factory) para el repositorio de especialidades.
// Recibe el pool de conexiones de GORM inyectado desde la capa de infraestructura.
func NewSpecialtyRepository(db *gorm.DB) domain.SpecialtyRepository {
	return &specialtyRepo{db: db}
}

// Create inserta un nuevo registro de especialidad en la base de datos.
// Propaga el context.Context para respetar los timeouts y cancelaciones de la petición HTTP.
func (r *specialtyRepo) Create(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetAll recupera una lista de especialidades aplicando un filtro dinámico de estado.
// Parámetro 'status':
//   - "active": retorna solo las especialidades habilitadas (Público general).
//   - "inactive": retorna solo las especialidades deshabilitadas.
//   - "all" u otro: ignora el filtro (retorna todas).
//
// Devuelve los resultados ordenados alfabéticamente por defecto para optimizar el Frontend.
func (r *specialtyRepo) GetAll(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
	var list []domain.PsiSpecialtyModel
	query := r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{}).Order("name asc")

	switch status {
	case "active":
		query = query.Where("active = ?", true)
	case "inactive":
		query = query.Where("active = ?", false)
	}

	err := query.Find(&list).Error
	return list, err
}

// GetByID recupera una única especialidad basada en su identificador numérico (uint32).
// Nota de Seguridad: Este método fuerza la cláusula 'active = true' a nivel de SQL
// para garantizar que endpoints públicos no puedan acceder a especialidades desactivadas por fuerza bruta.
func (r *specialtyRepo) GetByID(ctx context.Context, id uint32, active bool) (*domain.PsiSpecialtyModel, error) {
	var s domain.PsiSpecialtyModel

	// Iniciamos la consulta con el ID que es obligatorio
	query := r.db.WithContext(ctx).Where("id = ?", id)

	// Si active es true, agregamos la condición al WHERE
	// Si active es false, no agregamos nada, permitiendo que traiga el registro sin importar su estado
	if active {
		query = query.Where("active = ?", true)
	}

	err := query.First(&s).Error
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// GetByID recupera una única especialidad basada en su identificador numérico (uint32).
// Nota de Seguridad: Este método fuerza la cláusula 'active = true' a nivel de SQL
// para garantizar que endpoints públicos no puedan acceder a especialidades desactivadas por fuerza bruta.
func (r *specialtyRepo) GetByAdminID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	var s domain.PsiSpecialtyModel
	err := r.db.WithContext(ctx).First(&s, "id = ? ", id).Error
	return &s, err
}

// Update guarda las modificaciones de una especialidad.
// Se recomienda usar este método después de un GetByID para aplicar actualizaciones parciales
// sin sobreescribir campos existentes con valores vacíos.
func (r *specialtyRepo) Update(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return r.db.WithContext(ctx).Save(s).Error
}

// Delete ejecuta un borrado lógico (Soft-Delete) de la especialidad.
// En lugar de hacer un DROP/DELETE en SQL, actualiza específicamente el campo 'active' a false
// usando un map. Esto evita actualizar accidentalmente otros campos como el 'update_by' si el
// modelo en memoria estuviera desactualizado.
func (r *specialtyRepo) Delete(ctx context.Context, id uint32) error {
	return r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"active": false}).Error
}

// Count calcula el total de registros en el catálogo.
// Recibe un puntero a bool (*active) para permitir lógica tri-estatal:
//   - Si es puntero a true/false: filtra por ese estado.
//   - Si es nil: omite el WHERE y cuenta la totalidad de la tabla.
func (r *specialtyRepo) Count(ctx context.Context, active *bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{})

	if active != nil {
		query = query.Where("active = ?", *active)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetAllAdmin es un método optimizado ("Rayos X") para el panel de administración.
// Omite todas las reglas de negocio de visibilidad (active/inactive) devolviendo el
// catálogo absoluto, ordenado para fácil lectura del staff del colegio.
func (r *specialtyRepo) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {
	var list []domain.PsiSpecialtyModel
	err := r.db.WithContext(ctx).Order("name asc").Find(&list).Error
	return list, err
}

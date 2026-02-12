// Package postgres proporciona la implementación de persistencia para el motor de base de datos PostgreSQL
// utilizando el ORM GORM.
package postgres

import (
	"context"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// specialtyRepo implementa la interfaz domain.SpecialtyRepository.
// Utiliza una instancia de gorm.DB para realizar operaciones CRUD sobre la tabla 'psi_specialty_models'.
type specialtyRepo struct {
	db *gorm.DB
}

// NewSpecialtyRepository crea e inicializa una nueva instancia del repositorio de especialidades.
// Recibe el pool de conexiones de GORM inyectado desde la infraestructura.
func NewSpecialtyRepository(db *gorm.DB) domain.SpecialtyRepository {
	return &specialtyRepo{db: db}
}

// Create persiste una nueva especialidad en la base de datos.
// Se utiliza context.Context para permitir la cancelación de la consulta en caso de timeout.
func (r *specialtyRepo) Create(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetAll recupera una lista de especialidades basada en el estado solicitado.
// 'status' puede ser: "active" (solo activas), "inactive" (solo inactivas) o "all" (sin filtro).
// Los resultados se devuelven ordenados alfabéticamente por nombre.
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

// GetByID busca una especialidad por su identificador único numérico.
// Implementa un filtro de seguridad adicional para asegurar que el registro buscado esté activo.
func (r *specialtyRepo) GetByID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	var s domain.PsiSpecialtyModel
	// Se incluye la validación de 'active' para garantizar integridad en consultas públicas
	err := r.db.WithContext(ctx).First(&s, "id = ? AND active = ?", id, true).Error
	return &s, err
}

// Update actualiza todos los campos de una especialidad existente.
// Se recomienda cargar previamente el objeto desde la DB para evitar sobrescrituras accidentales.
func (r *specialtyRepo) Update(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return r.db.WithContext(ctx).Save(s).Error
}

// Delete realiza una desactivación lógica (soft-delete) del registro.
// En lugar de eliminar la fila físicamente, establece el flag 'active' en false.
// Esto permite preservar las relaciones históricas con los perfiles de psicólogos.
func (r *specialtyRepo) Delete(ctx context.Context, id uint32) error {
	return r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{}).Where("id = ?", id).
		Updates(map[string]interface{}{"active": false}).Error
}

// Count retorna el número total de especialidades según el filtro de estado proporcionado.
// Recibe un puntero a bool (*active) para permitir tres estados:
// - true: cuenta activas
// - false: cuenta inactivas
// - nil: cuenta todas sin discriminación
func (r *specialtyRepo) Count(ctx context.Context, active *bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{})

	if active != nil {
		query = query.Where("active = ?", *active)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetAllAdmin recupera el catálogo completo de especialidades sin aplicar filtros de estado.
// Diseñado para su uso exclusivo en paneles de administración.
func (r *specialtyRepo) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {
	var list []domain.PsiSpecialtyModel
	err := r.db.WithContext(ctx).Order("name asc").Find(&list).Error
	return list, err
}

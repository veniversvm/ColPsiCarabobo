// Package postgres proporciona la capa de persistencia (Data Access Layer)
// para el motor de base de datos PostgreSQL utilizando el ORM GORM.
//
// Este archivo encapsula específicamente la gestión del catálogo de Especialidades
// o Áreas de Desempeño Clínico (PsiSpecialtyModel).
package postgres

import (
	"context"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// specialtyRepo es la implementación concreta de la interfaz [domain.SpecialtyRepository].
// Centraliza la lógica de acceso a datos para aislar a los servicios de los detalles
// de infraestructura y SQL. No debe instanciarse directamente; usar [NewSpecialtyRepository].
type specialtyRepo struct {
	db *gorm.DB
}

// NewSpecialtyRepository actúa como constructor (Factory) para el repositorio de especialidades.
// Recibe el pool de conexiones de GORM inyectado desde la capa de infraestructura y retorna
// la abstracción del dominio, garantizando los principios de Clean Architecture.
func NewSpecialtyRepository(db *gorm.DB) domain.SpecialtyRepository {
	return &specialtyRepo{db: db}
}

// Create inserta un nuevo registro de especialidad en el catálogo.
//
// Propaga el context.Context para respetar los timeouts y cancelaciones
// provenientes de la petición HTTP, previniendo bloqueos en la base de datos.
func (r *specialtyRepo) Create(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetAll recupera una lista de especialidades aplicando un filtro dinámico.
// Diseñado principalmente para alimentar menús desplegables (dropdowns) en el frontend.
//
// Comportamiento del parámetro 'status':
//   - "active": retorna solo las especialidades habilitadas (para vistas públicas).
//   - "inactive": retorna solo las especialidades deshabilitadas (auditoría).
//   - "all" u otro: ignora el filtro y retorna la tabla completa.
//
// Los resultados se ordenan alfabéticamente (ASC) por defecto para optimizar la UX.
func (r *specialtyRepo) GetAll(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
	var list []domain.PsiSpecialtyModel

	// Iniciamos la consulta base con ordenamiento predefinido.
	query := r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{}).Order("name asc")

	// Aplicamos los filtros condicionales
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
//
// Regla de Seguridad (Escudo Público):
// Si el parámetro 'active' es true, fuerza la cláusula 'active = true' a nivel SQL.
// Esto bloquea intentos de fuerza bruta desde endpoints públicos que busquen acceder
// a especialidades ocultas o deprecadas adivinando el ID secuencial.
func (r *specialtyRepo) GetByID(ctx context.Context, id uint32, active bool) (*domain.PsiSpecialtyModel, error) {
	var s domain.PsiSpecialtyModel

	query := r.db.WithContext(ctx).Where("id = ?", id)

	if active {
		query = query.Where("active = ?", true)
	}

	err := query.First(&s).Error
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// GetByAdminID es un método de acceso privilegiado para el panel de administración.
//
// A diferencia de GetByID, omite deliberadamente cualquier regla de negocio relacionada
// con la visibilidad (active). Permite al staff consultar la metadata de una
// especialidad sin importar su estado actual en el sistema.
func (r *specialtyRepo) GetByAdminID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	var s domain.PsiSpecialtyModel
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

// Update persiste las modificaciones de una especialidad usando Updates() con mapa explícito.
// El campo Active usa gorm.Expr para forzar la escritura del valor real (true/false),
// evitando que GORM lo omita por ser un zero-value.
func (r *specialtyRepo) Update(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return r.db.WithContext(ctx).Model(s).Updates(map[string]interface{}{
		"name":        s.Name,
		"description": s.Description,
		"active":      gorm.Expr("?", s.Active),
		"update_by":   s.UpdateBy,
		"update_by_id": s.UpdateById,
	}).Error
}

// Delete ejecuta un borrado lógico (Desactivación) por regla de negocio.
//
// En lugar de hacer un DROP/DELETE físico en SQL (lo cual rompería las claves foráneas
// de los psicólogos asociados a esta especialidad), actualiza únicamente el flag 'active'
// a false usando un mapa estricto.
// Esto asegura que la operación sea atómica y no sobreescriba accidentalmente
// sellos de auditoría en memoria.
func (r *specialtyRepo) Delete(ctx context.Context, id uint32) error {
	return r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"active": false}).Error
}

// Count calcula el volumen total de registros en el catálogo.
//
// Implementa un patrón de filtrado tri-estatal mediante el puntero *active:
//   - Si apunta a true: cuenta las activas.
//   - Si apunta a false: cuenta las inactivas.
//   - Si es nil: omite la cláusula WHERE y devuelve el tamaño absoluto de la tabla.
func (r *specialtyRepo) Count(ctx context.Context, active *bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.PsiSpecialtyModel{})

	if active != nil {
		query = query.Where("active = ?", *active)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetAllAdmin es una proyección ("Rayos X") reservada para el panel de administración.
//
// Ignora todas las reglas de negocio de visibilidad y devuelve el catálogo absoluto,
// asegurando que las herramientas de gestión interna (como listados de edición)
// siempre tengan la foto completa de la base de datos.
func (r *specialtyRepo) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {
	var list []domain.PsiSpecialtyModel
	err := r.db.WithContext(ctx).Order("name asc").Find(&list).Error
	return list, err
}

// api/internal/repository/postgres/user_admin_repo.go

// Package postgres provee la implementación concreta de los repositorios usando PostgreSQL y GORM.
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// adminRepo implementa la interfaz domain.UserAdminRepository.
// Maneja la persistencia de los usuarios con privilegios administrativos del sistema,
// permitiendo la gestión de accesos y la supervisión de superusuarios (SUDOs).
type adminRepo struct {
	db *gorm.DB
}

// NewAdminRepository crea una nueva instancia del repositorio de administradores.
func NewAdminRepository(db *gorm.DB) domain.UserAdminRepository {
	return &adminRepo{db: db}
}

// =========================================================================
// GESTIÓN CORE DEL ADMINISTRADOR
// =========================================================================

// GetByIdentifier busca un administrador por su username O por su email.
// Es la pieza clave para un login flexible, permitiendo al usuario elegir su credencial de acceso.
func (r *adminRepo) GetByIdentifier(ctx context.Context, identifier string) (*domain.UserAdmin, error) {
	var admin domain.UserAdmin

	// Buscamos coincidencia en cualquiera de las dos columnas únicas (Email o Username)
	err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", identifier, identifier).
		First(&admin).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("administrador no encontrado")
		}
		return nil, err
	}

	return &admin, nil
}

// GetByID busca un administrador por su UUID único.
func (r *adminRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
	var admin domain.UserAdmin
	err := r.db.WithContext(ctx).First(&admin, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// Create inserta un nuevo administrador en la base de datos.
func (r *adminRepo) Create(ctx context.Context, user *domain.UserAdmin) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// =========================================================================
// ACTUALIZACIONES Y ESTADO
// =========================================================================

// Update actualiza todos los campos de un administrador existente.
// Utiliza Save(), lo cual incluye los campos de auditoría automáticos de GORM.
//
// Advertencia Técnica: Save() sobreescribe TODOS los campos incluyendo zero-values.
// El caller DEBE obtener el modelo completo vía GetByID antes de modificar campos.
// Nunca pasar un modelo parcialmente construido a este método.
func (r *adminRepo) Update(ctx context.Context, user *domain.UserAdmin) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete realiza un borrado lógico (Soft Delete).
// Gracias a la integración con AuditModel/DeletedAt, GORM ejecutará un UPDATE
// seteando la fecha actual en lugar de eliminar el registro físicamente.
func (r *adminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserAdmin{}, "id = ?", id).Error
}

// CountSudos cuenta cuántos administradores tienen privilegios de Superusuario activos.
// Es vital para validaciones de seguridad (ej. evitar que el sistema se quede sin administradores).
func (r *adminRepo) CountSudos(ctx context.Context) (int64, error) {
	var count int64
	// Solo contamos los que NO están borrados (Soft delete) y son SUDO
	err := r.db.WithContext(ctx).
		Model(&domain.UserAdmin{}).
		Where("sudo = ? AND deleted_at IS NULL", true).
		Count(&count).Error

	return count, err
}

// =========================================================================
// MOTORES DE BÚSQUEDA Y LISTADO
// =========================================================================

// List recupera una lista paginada de administradores con filtros opcionales.
// Soporta búsqueda parcial por texto y filtrado por estado de actividad.
func (r *adminRepo) List(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
	var admins []domain.UserAdmin
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.UserAdmin{})

	// Filtro por estado activo (se usa puntero para distinguir entre 'false' y 'no enviado')
	if active != nil {
		query = query.Where("is_active = ?", *active)
	}

	// Filtro por búsqueda parcial (Email o Username) usando ILIKE para PostgreSQL (Case-Insensitive)
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ?", s, s)
	}

	// 1. Contar el total de registros que coinciden con los filtros antes de paginar
	query.Count(&total)

	// 2. Aplicar paginación y ordenamiento (más recientes primero)
	offset := (page - 1) * limit
	err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&admins).Error

	return admins, total, err
}

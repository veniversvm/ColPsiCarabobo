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
// Utiliza Updates() con un mapa explícito para evitar que GORM sobreescriba
// booleanos con zero-values (false) — un error común con Save().
// Los campos booleanos usan gorm.Expr para forzar la escritura del valor real.
func (r *adminRepo) Update(ctx context.Context, user *domain.UserAdmin) error {
	return r.db.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		// ── Credenciales ──────────────────────────────────────────────
		"username": user.Username,
		"email":    user.Email,
		"password": user.Password,
		"key":      user.Key,

		// ── Estatus y Contraseña ─────────────────────────────────────
		"is_active":            gorm.Expr("?", user.IsActive),
		"must_change_password": gorm.Expr("?", user.MustChangePassword),
		"sudo":                 gorm.Expr("?", user.Sudo),

		// ── Permisos: Colegiados ─────────────────────────────────────
		"can_read_psi":   gorm.Expr("?", user.CanReadPsi),
		"can_create_psi": gorm.Expr("?", user.CanCreatePsi),
		"can_update_psi": gorm.Expr("?", user.CanUpdatePsi),
		"can_delete_psi": gorm.Expr("?", user.CanDeletePsi),

		// ── Permisos: Personal Administrativo ────────────────────────
		"can_create_admin": gorm.Expr("?", user.CanCreateAdmin),
		"can_update_admin": gorm.Expr("?", user.CanUpdateAdmin),
		"can_delete_admin": gorm.Expr("?", user.CanDeleteAdmin),

		// ── Permisos: Contenido y Comunicación ───────────────────────
		"can_publish":              gorm.Expr("?", user.CanPublish),
		"can_update_publish":       gorm.Expr("?", user.CanUpdatePublish),
		"can_delete_publish":       gorm.Expr("?", user.CanDeletePublish),
		"can_send_notifications":   gorm.Expr("?", user.CanSendNotifications),
		"can_manage_notifications": gorm.Expr("?", user.CanManageNotifications),
		"can_read_notifications":   gorm.Expr("?", user.CanReadNotifications),

		// ── Permisos: Catálogo de Especialidades ─────────────────────
		"can_create_tags": gorm.Expr("?", user.CanCreateTags),
		"can_edit_tags":   gorm.Expr("?", user.CanEditTags),
		"can_delete_tags": gorm.Expr("?", user.CanDeleteTags),

		// ── Permisos: Proyectos (Kanban) ──────────────────────────────
		"can_manage_projects": gorm.Expr("?", user.CanManageProjects),

		// ── Permisos: Tickets de Solicitudes ──────────────────────────
		"can_manage_tickets": gorm.Expr("?", user.CanManageTickets),

		// ── Rótulo del preset aplicado (solo metadato) ───────────────
		"role": user.Role,

		// ── Auditoría ────────────────────────────────────────────────
		"update_by":    user.UpdateBy,
		"update_by_id": user.UpdateById,
	}).Error
}

// Delete realiza un borrado lógico (Soft Delete).
// Gracias a la integración con AuditModel/DeletedAt, GORM ejecutará un UPDATE
// seteando la fecha actual en lugar de eliminar el registro físicamente.
func (r *adminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserAdmin{}, "id = ?", id).Error
}

// UpdateKey actualiza exclusivamente la semilla de firma (Key) y auditoría.
// Optimizado para logout/invalidación de sesiones — evita sobreescribir todo el modelo.
func (r *adminRepo) UpdateKey(ctx context.Context, user *domain.UserAdmin) error {
	return r.db.WithContext(ctx).Model(user).
		Select("Key", "UpdatedAt", "UpdateBy", "UpdateById").
		Updates(user).Error
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

// TransferSudo intercambia el estado de Sudo de forma atómica en una
// transacción: primero se revoca al actual (sudo=false), luego se otorga al
// sucesor (sudo=true). Respeta el índice parcial único sobre sudo=true.
func (r *adminRepo) TransferSudo(ctx context.Context, fromID, toID uuid.UUID) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.UserAdmin{}).Where("id = ?", fromID).Update("sudo", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.UserAdmin{}).Where("id = ?", toID).Update("sudo", true).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

// CreatePermissionLog inserta una entrada de auditoría de cambios de permisos.
func (r *adminRepo) CreatePermissionLog(ctx context.Context, log *domain.AdminPermissionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

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

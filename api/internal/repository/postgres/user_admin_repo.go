// api/internal/repository/postgres/user_admin_repo.go
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// adminRepo implementa la interfaz domain.UserAdminRepository
type adminRepo struct {
	db *gorm.DB
}

// NewAdminRepository crea una nueva instancia del repositorio de administradores
func NewAdminRepository(db *gorm.DB) domain.UserAdminRepository {
	return &adminRepo{db: db}
}

// GetByIdentifier busca un administrador por su username O por su email.
// Es la pieza clave para un login flexible.
func (r *adminRepo) GetByIdentifier(ctx context.Context, identifier string) (*domain.UserAdmin, error) {
	var admin domain.UserAdmin

	// Buscamos coincidencia en cualquiera de las dos columnas únicas
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

// GetByID busca un administrador por su UUID único
func (r *adminRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
	var admin domain.UserAdmin
	err := r.db.WithContext(ctx).First(&admin, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// Create inserta un nuevo administrador en la base de datos
func (r *adminRepo) Create(ctx context.Context, user *domain.UserAdmin) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update actualiza los datos de un administrador existente
func (r *adminRepo) Update(ctx context.Context, user *domain.UserAdmin) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete realiza un borrado lógico (Soft Delete) gracias a GORM y AuditModel
func (r *adminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserAdmin{}, "id = ?", id).Error
}

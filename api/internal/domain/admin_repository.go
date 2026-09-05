// api/internal/domain/admin_repository.go
package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserAdminRepository define la interfaz de abstracción para el acceso a datos
// de los administradores. Sigue el principio de Inversión de Dependencias (DIP).
type UserAdminRepository interface {
	// GetByID recupera un administrador por su identificador único (UUID).
	// Es fundamental para la validación de tokens JWT en el middleware.
	GetByID(ctx context.Context, id uuid.UUID) (*UserAdmin, error)

	List(ctx context.Context, active *bool, search string, page, limit int) ([]UserAdmin, int64, error)

	// GetByIdentifier busca un administrador por su nombre de usuario O correo electrónico.
	// Se utiliza principalmente en el proceso de inicio de sesión (Login).
	GetByIdentifier(ctx context.Context, identifier string) (*UserAdmin, error)

	// Create inserta un nuevo registro de administrador en la base de datos.
	Create(ctx context.Context, user *UserAdmin) error

	// Update guarda los cambios realizados en un administrador existente.
	// Incluye la rotación de la clave (Key) para el manejo de sesiones únicas.
	Update(ctx context.Context, user *UserAdmin) error

	// Delete realiza un borrado lógico del administrador.
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateKey actualiza exclusivamente la semilla de firma (Key) y auditoría.
	// Optimizado para logout/invalidación de sesiones.
	UpdateKey(ctx context.Context, user *UserAdmin) error

	CountSudos(ctx context.Context) (int64, error)

	// TransferSudo intercambia el estado de Sudo entre dos administradores en
	// una sola transacción (atómico): el que lo cede pasa a false y el sucesor
	// pasa a true, respetando el índice único parcial sobre sudo.
	TransferSudo(ctx context.Context, fromID, toID uuid.UUID) error

	// CreatePermissionLog persiste una entrada de auditoría de permisos.
	CreatePermissionLog(ctx context.Context, log *AdminPermissionLog) error
}

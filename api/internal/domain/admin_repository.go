package domain

import "context"

// internal/domain/user_admin_repository.go
type UserAdminRepository interface {
	// ... otros métodos ...
	GetByIdentifier(ctx context.Context, identifier string) (*UserAdmin, error)
}

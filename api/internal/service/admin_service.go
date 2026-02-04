// api/internal/service/admin_service.go

package service

import (
	"context"
	"errors"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	repo domain.UserAdminRepository
}

func NewAdminService(repo domain.UserAdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) Login(ctx context.Context, identifier, password string) (*domain.UserAdmin, error) {
	admin, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		// Por seguridad, no especificamos si el usuario no existe o si la clave es errónea
		return nil, errors.New("credenciales inválidas")
	}

	if !admin.IsActive {
		return nil, errors.New("cuenta desactivada, contacte al administrador superior")
	}

	// Comparar contraseñas
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	return admin, nil
}

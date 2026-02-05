// api/internal/service/admin_service.go

package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5" // Usar v5 para consistencia
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	repo domain.UserAdminRepository
}

func NewAdminService(repo domain.UserAdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

// Login valida credenciales y retorna un JWT firmado con una clave única rotativa.
func (s *AdminService) Login(ctx context.Context, identifier, password string) (string, error) {
	// 1. Buscar administrador
	admin, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		// Senior tip: Ocultamos errores específicos de DB para evitar enumeración de usuarios
		return "", errors.New("credenciales inválidas")
	}

	// 2. Verificar estado de la cuenta
	if !admin.IsActive {
		return "", errors.New("la cuenta está desactivada")
	}

	// 3. Verificar password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// 4. ROTACIÓN DE KEY (Single Session Support)
	// Generamos un nuevo UUID que servirá de "Secret" solo para este token.
	// Esto invalida automáticamente cualquier JWT emitido anteriormente.
	newKey := uuid.New().String()
	admin.Key = newKey

	// Persistimos la nueva Key en la DB
	if err := s.repo.Update(ctx, admin); err != nil {
		return "", errors.New("error al procesar inicio de sesión")
	}

	// 5. GENERACIÓN DEL TOKEN
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// Sincronizado con el middleware: usamos "user_id" como clave genérica
		"user_id": admin.ID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expira en 24h
		"iat":     time.Now().Unix(),
		"role":    "admin", // Útil para lógica rápida en el frontend
	})

	// 6. FIRMA DINÁMICA
	// Firmamos con la clave única del usuario que acabamos de guardar en la DB
	return token.SignedString([]byte(newKey))
}

// GetRepo permite al middleware acceder al repositorio si es necesario
func (s *AdminService) GetRepo() domain.UserAdminRepository {
	return s.repo
}

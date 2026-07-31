package service

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// Login authenticates a psychologist by identifier and password, rotating the session key and returning a signed JWT.
func (s *PsiService) Login(ctx context.Context, identifier, password string) (string, *domain.PsiUserModel, error) {
	psi, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		return "", nil, errors.New("credenciales inválidas")
	}

	if !psi.IsActive {
		return "", nil, errors.New("cuenta inactiva o suspendida")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(psi.Password), []byte(password)); err != nil {
		return "", nil, errors.New("credenciales inválidas")
	}

	newKey := uuid.Must(uuid.NewV7()).String()
	psi.Key = newKey

	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID

	if err := s.repo.UpdateKey(ctx, psi); err != nil {
		return "", nil, errors.New("error de sistema al iniciar sesión")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": psi.ID.String(),
		"role":    "psi",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	mailData := map[string]interface{}{
		"Name":      psi.Username,
		"Email":     psi.Email,
		"LoginTime": time.Now().Format(time.RFC1123),
	}

	if s.mailService != nil {
		if err := s.mailService.SendEmail(psi.Email, "Colegio de Psicólogos de Carabobo - Inicio de sesión en la plataforma.", "login_psi", mailData); err != nil {
			log.Warn().Err(err).Str("component", "psi_service_auth").Msg("Error al preparar el correo (pero el psicólogo se logueó)")
		}
	}

	signed, err := token.SignedString([]byte(newKey))
	return signed, psi, err
}

// AudiobookshelfUserResponse represents the API response when creating or fetching an Audiobookshelf user.
type AudiobookshelfUserResponse struct {
	User struct {
		ID string `json:"id"`
	} `json:"user"`
}

// LoginLibrary authenticates a psychologist and syncs the account with Audiobookshelf, returning a library-specific JWT.
func (s *PsiService) LoginLibrary(ctx context.Context, identifier, password string) (string, *domain.PsiUserModel, error) {
	psi, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		return "", nil, errors.New("credenciales inválidas")
	}

	if !psi.IsActive {
		return "", nil, errors.New("cuenta inactiva o suspendida")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(psi.Password), []byte(password)); err != nil {
		return "", nil, errors.New("credenciales inválidas")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": psi.ID.String(),
		"role":    "psi",
		"exp":     time.Now().Add(24 * 30 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	jwtLibrarySecret := config.Envs.JwtLibrarySecret
	signed, err := token.SignedString([]byte(jwtLibrarySecret))
	if err != nil {
		return "", nil, err
	}

	absID, absErr := s.sincronizarConAudiobookshelf(ctx, psi.Username, password, psi.Email)
	if absErr != nil {
		log.Warn().Err(absErr).Str("component", "psi_service_auth").Msg("Error sincronizando con Audiobookshelf")
	} else if absID != "" {
		psi.AudioBookShellId = absID
		s.repo.Update(ctx, psi, nil, nil, nil)
		log.Info().Str("component", "psi_service_auth").Str("abs_id", absID).Msg("Usuario creado en Audiobookshelf")
	} else {
		log.Info().Str("component", "psi_service_auth").Msg("El usuario ya existía en Audiobookshelf, no se generó un nuevo ID")
	}

	return signed, psi, nil
}

// Logout clears the session key for the given psychologist, effectively invalidating their JWT.
func (s *PsiService) Logout(ctx context.Context, psi *domain.PsiUserModel) error {
	psi.Key = ""
	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID
	return s.repo.UpdateKey(ctx, psi)
}

package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

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

	if err := s.mailService.SendEmail(psi.Email, "Colegio de Psicólogos de Carabobo - Inicio de sesión en la plataforma.", "login_psi", mailData); err != nil {
		log.Printf("[WARN] Error al preparar el correo (pero el psicólogo se logueó): %v", err)
	}

	signed, err := token.SignedString([]byte(newKey))
	return signed, psi, err
}

type AudiobookshelfUserResponse struct {
	User struct {
		ID string `json:"id"`
	} `json:"user"`
}

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
		log.Printf("WARN: Error sincronizando con Audiobookshelf: %v", absErr)
	} else if absID != "" {
		psi.AudioBookShellId = absID
		s.repo.Update(ctx, psi, nil, nil, nil)
		log.Printf("INFO: Usuario creado en Audiobookshelf con ID: %s", absID)
	} else {
		log.Printf("INFO: El usuario ya existía en Audiobookshelf, no se generó un nuevo ID.")
	}

	return signed, psi, nil
}

func (s *PsiService) Logout(ctx context.Context, psi *domain.PsiUserModel) error {
	psi.Key = ""
	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID
	return s.repo.UpdateKey(ctx, psi)
}

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

const resetTokenTTL = 1 * time.Hour

// hashToken calcula el hash SHA-256 (hex, 64 chars) de un token de
// recuperación. Nunca almacenamos el token en claro; solo el hash.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// siteURLBuilder devuelve la URL pública de la aplicación desde donde se
// renderizan los enlaces de recuperación. Prioriza SITE_URL sobre el valor
// por defecto (producción) para permitir pruebas locales.
func siteURLBuilder() string {
	if config.Envs != nil && config.Envs.AppURL != "" && config.Envs.AppURL != "http://localhost:3000" {
		return strings.TrimRight(config.Envs.AppURL, "/")
	}
	return "https://franhsabt-testing-ground.lat"
}

// RequestPasswordReset solicita el envío de un enlace de recuperación de
// contraseña al correo registrado. La respuesta es genérica por diseño:
// no revela si el correo existe (anti-enumeración).
func (s *PsiService) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// Se busca por identifier (Username o Email). Si no existe o la cuenta
	// está inactiva, igualmente devolvemos éxito para no revelar información.
	psi, err := s.repo.GetByIdentifier(ctx, email)
	if err != nil || psi == nil || !psi.IsActive {
		log.Warn().Str("component", "psi_service").Str("email", email).Msg("Solicitud de reset: email no encontrado o inactivo")
		return nil // Respuesta genérica — no revelar
	}

	// Generar token y hash
	token := utils.GenerateSecureRandomString(32)
	tokenHash := hashToken(token)

	resetToken := &domain.PsiPasswordResetToken{
		ID:        uuid.Must(uuid.NewV7()),
		PsiID:     psi.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(resetTokenTTL),
	}
	if err := s.repo.CreateResetToken(ctx, resetToken); err != nil {
		log.Warn().Err(err).Str("component", "psi_service").Msg("Error al crear token de reset")
		return nil // No fallar ante el usuario
	}

	// Enviar correo con el enlace (fire-and-forget)
	if s.mailService != nil {
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", siteURLBuilder(), token)
		mailData := map[string]interface{}{
			"Name":     psi.FirstName,
			"Email":    psi.Email,
			"ResetURL": resetURL,
		}
		if err := s.mailService.SendEmail(psi.Email, "Recupera tu contraseña", "reset_password_token", mailData); err != nil {
			log.Warn().Err(err).Str("component", "psi_service").Msg("Error al encolar correo de reset")
		}
	}

	return nil
}

// ResetPasswordWithToken valida el token de un solo uso, fija la nueva
// contraseña del psicólogo y rota la Key de sesión (invalidando JWTs).
func (s *PsiService) ResetPasswordWithToken(ctx context.Context, tokenStr, newPassword string) error {
	if len(newPassword) < 8 {
		return domain.ErrInvalidRequest
	}

	tokenHash := hashToken(tokenStr)

	resetToken, err := s.repo.GetResetTokenByHash(ctx, tokenHash)
	if err != nil {
		log.Warn().Err(err).Str("component", "psi_service").Msg("Error buscando token de reset")
		return domain.ErrInvalidRequest
	}
	if resetToken == nil {
		return domain.ErrInvalidRequest
	}

	// Validaciones: token no expirado, no usado
	if resetToken.UsedAt != nil {
		log.Warn().Str("component", "psi_service").Msg("Intento de reusar token de reset")
		return domain.ErrInvalidRequest
	}
	if time.Now().After(resetToken.ExpiresAt) {
		log.Warn().Str("component", "psi_service").Msg("Token de reset expirado")
		return domain.ErrInvalidRequest
	}

	// Hashear la nueva contraseña
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error al procesar la contraseña")
	}

	// Cargar el psicólogo y actualizar credenciales
	psi := &resetToken.Psi
	psi.Password = string(hashed)
	psi.Key = uuid.Must(uuid.NewV7()).String()
	psi.MustChangePassword = false
	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID

	if err := s.repo.ResetPassword(ctx, psi); err != nil {
		return fmt.Errorf("error al actualizar la contraseña: %w", err)
	}

	// Marcar token como consumido (single-use)
	if err := s.repo.MarkResetTokenUsed(ctx, resetToken.ID); err != nil {
		log.Warn().Err(err).Str("component", "psi_service").Msg("Error al marcar token como usado")
	}

	return nil
}
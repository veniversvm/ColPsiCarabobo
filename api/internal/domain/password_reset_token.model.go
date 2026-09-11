// api/internal/domain/password_reset_token.model.go
// Token de recuperación de contraseña de un solo uso para psicólogos.
// El token en claro nunca se persiste: solo el hash SHA-256 (hex, 64 chars).
package domain

import (
	"time"

	"github.com/google/uuid"
)

// PsiPasswordResetToken es el token de recuperación de credenciales del psicólogo.
// Nace con una ventana de validez (ExpiresAt, por defecto 1 hora) y se consume
// una sola vez (UsedAt) al fijar la nueva contraseña.
type PsiPasswordResetToken struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	PsiID     uuid.UUID    `gorm:"type:uuid;index;not null" json:"psi_id"`
	TokenHash string       `gorm:"size:64;uniqueIndex:idx_reset_token_hash;not null" json:"-"`
	ExpiresAt time.Time    `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time   `json:"used_at"`
	CreatedAt time.Time    `json:"created_at"`
	Psi       PsiUserModel `gorm:"foreignKey:PsiID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

// TableName devuelve el nombre físico de la tabla en PostgreSQL.
func (PsiPasswordResetToken) TableName() string { return "psi_password_reset_tokens" }
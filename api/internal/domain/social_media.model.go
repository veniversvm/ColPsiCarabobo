package domain

import "github.com/google/uuid"

// --- REDES SOCIALES ---

// PsiUserSocialNetwork representa los enlaces digitales del psicólogo.
type PsiUserSocialNetwork struct {
	ID uuid.UUID `gorm:"type:uuid;index;not null" json:"id"`

	AuditModel

	PsiUserID uuid.UUID `gorm:"type:uuid;index;not null" json:"psi_user_id"`

	// Name identifica la plataforma (ej: "Instagram", "LinkedIn", "Web Personal")
	Name string `gorm:"size:50;not null" json:"name"`

	// URL es el enlace directo al perfil o página
	URL string `gorm:"size:512;not null" json:"url"`

	// IsActive permite al usuario ocultar una red sin borrarla físicamente
	IsActive bool `gorm:"default:true;not null" json:"is_active"`
}

func (PsiUserSocialNetwork) TableName() string {
	return "psi_user_social_networks"
}

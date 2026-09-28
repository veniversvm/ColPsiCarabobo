// api/internal/domain/emergency_contact.model.go
package domain

import "github.com/google/uuid"

// --- CONTACTO DE EMERGENCIA ---

// PsiUserEmergencyContact es la persona de contacto que el agremiado autoriza
// para ser avisada en caso de emergencia: cuando no puede ser localizado o
// cuando ocurre un accidente durante el ejercicio profesional.
//
// IMPORTANTE: son datos personales de un TERCERO. NUNCA se exponen en endpoints
// públicos (el directorio y la ficha pública usan DTOs propios que no incluyen
// esta relación). Solo los consulta el propio agremiado (auto-gestión,
// `GET /psi/me`) y el personal autorizado del Colegio a través de la ficha
// administrativa (`GET /admin/psi/:id`).
//
// Regla de integridad (verificada en el service y reforzada con CHECK en la
// migración): nombre y parentesco siempre; y al menos UNO de Phone o Email.
// Relación N-a-1 con PsiUserModel.
type PsiUserEmergencyContact struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`

	AuditModel

	PsiUserID uuid.UUID `gorm:"type:uuid;index;not null" json:"psi_user_id"`

	// Nombre de la persona a la que se avisaría (obligatorio).
	Name string `gorm:"size:255;not null" json:"name"`

	// Parentesco o relación con el agremiado (obligatorio). Catálogo corto
	// definido en el frontend (padre, madre, conyuge, hermano, hijo, tio, primo,
	// amigo, colega) con opción "otro" en texto libre.
	Relationship string `gorm:"size:100;not null" json:"relationship"`

	// Canal de contacto: al menos uno de los dos.
	Phone string `gorm:"size:20" json:"phone"`
	Email string `gorm:"size:255" json:"email"`
}

func (PsiUserEmergencyContact) TableName() string { return "psi_user_emergency_contacts" }

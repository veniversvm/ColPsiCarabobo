// api/internal/domain/inscription_document.model.go
package domain

import (
	"github.com/google/uuid"
)

// PsiInscriptionDocument registra la foto de un documento requerido durante la
// pre-inscripción (cedula, titulo, rif, otro). Pertenece a la solicitud y, al
// aprobarse, migra a psi_user_documents con el expediente del psicólogo.
// Se permite un único archivo por categoría dentro de la misma solicitud
// (uniqueIndex compuesto inscription_request_id + document_type).
type PsiInscriptionDocument struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel

	InscriptionRequestID uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_inscription_document_type" json:"inscription_request_id"`
	DocumentType         DocumentType `gorm:"type:varchar(50);not null;uniqueIndex:idx_inscription_document_type" json:"document_type"`
	S3Key                string       `gorm:"size:512;not null" json:"s3_key"`
	Title                string       `gorm:"size:255" json:"title"`
	Notes                string       `gorm:"type:text" json:"notes"`
	OriginalFilename     string       `gorm:"size:255" json:"original_filename"`
}

// TableName devuelve el nombre de la tabla en la base de datos.
func (PsiInscriptionDocument) TableName() string { return "psi_inscription_documents" }
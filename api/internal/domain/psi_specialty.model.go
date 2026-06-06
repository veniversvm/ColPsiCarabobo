// api/internal/domain/psi_specialty.model.go
package domain

// PsiSpecialtyModel representa el catálogo de especialidades.
// Usamos este nombre para ser consistentes con tu archivo main.go
type PsiSpecialtyModel struct {
	ID          uint32 `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Active      bool   `gorm:"default:true;not null" json:"active"`

	// Campos de auditoría manuales para no chocar con el ID uint32
	AuditModel
}

// TableName define el nombre exacto de la tabla para evitar confusiones con GORM
func (PsiSpecialtyModel) TableName() string {
	return "psi_specialty_models"
}

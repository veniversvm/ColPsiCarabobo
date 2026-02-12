// Package domain contiene las entidades de negocio y las interfaces del sistema.
// Este paquete es el núcleo de Clean Architecture y no debe tener dependencias externas.
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditModel es una estructura base diseñada para ser incrustada (embedded) en otros modelos.
// Proporciona campos estandarizados para la identificación única, trazabilidad de auditoría
// y soporte para borrado lógico (Soft Delete).
type AuditModel struct {

	// CreatedAt registra la fecha y hora exacta en que se creó el registro.
	// GORM gestiona este campo automáticamente durante la inserción.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt registra la última fecha y hora en que se modificó el registro.
	// GORM actualiza este valor automáticamente en cada operación de guardado.
	UpdatedAt time.Time `json:"updated_at"`

	// DeletedAt permite el "Soft Delete" (borrado lógico).
	// En lugar de eliminar físicamente la fila, se marca con una fecha de eliminación.
	// Las consultas estándar de GORM ignorarán automáticamente los registros con este campo lleno.
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Trazabilidad de Usuario (Auditoría Humana)
	// Estos campos permiten saber quién realizó las acciones, incluso si el usuario original
	// cambia su nombre o perfil.

	// CreateBy almacena el nombre o identificador textual del creador.
	CreateBy string `gorm:"size:255" json:"create_by"`

	// UpdateBy almacena el nombre o identificador textual de la última persona en modificarlo.
	UpdateBy string `gorm:"size:255" json:"update_by"`

	// CreateById es el UUID del usuario/administrador que creó el registro.
	// Es un puntero para permitir valores nulos si la creación es automática por el sistema.
	CreateById *uuid.UUID `gorm:"type:uuid" json:"create_by_id"`

	// UpdateById es el UUID del usuario/administrador que realizó la última actualización.
	UpdateById *uuid.UUID `gorm:"type:uuid" json:"update_by_id"`
}

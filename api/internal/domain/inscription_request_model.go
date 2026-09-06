// api/internal/domain/inscription_request_model.go
// Package domain define las entidades y los contratos de persistencia del sistema.
//
// Este archivo define el modelo de datos de las solicitudes de pre-inscripción
// de nuevos profesionales (psi_inscription_requests), que alimenta el flujo de
// ingreso digital antes de la creación del expediente en psi_users.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// InscriptionStatus representa el estado de una solicitud de pre-inscripción.
type InscriptionStatus string

// Estados permitidos de una solicitud de pre-inscripción.
const (
	InscriptionPending  InscriptionStatus = "pending"
	InscriptionApproved InscriptionStatus = "approved"
	InscriptionRejected InscriptionStatus = "rejected"
)

// PsiInscriptionRequest es la entidad que representa una solicitud de
// pre-inscripción de un nuevo profesional del Colegio.
//
// Se crea de forma pública (sin login) a través del formulario /inscripcion y
// permanece en estado "pending" hasta que la administración la aprueba
// (creando el expediente en psi_users) o la rechaza (eliminando registro y
// archivos S3).
type PsiInscriptionRequest struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel

	// ── Identidad ─────────────────────────────────────────────────────────
	// uniqueIndex parcial: una sola solicitud "pending" por cédula (integridad
	// de la ficha digital; los aprobados/rechazados no bloquean nuevas fichas).
	Cedula       int    `gorm:"not null;uniqueIndex:idx_inscription_requests_cedula_pending,where:status = 'pending'" json:"cedula"`
	Nacionalidad string `gorm:"size:1;not null" json:"nacionalidad"` // V / E
	Nombres     string `gorm:"size:255;not null" json:"nombres"`
	Apellidos   string `gorm:"size:255;not null" json:"apellidos"`
	SegundoNombre   string `gorm:"size:255" json:"segundo_nombre"`
	SegundoApellido string `gorm:"size:255" json:"segundo_apellido"`
	Genero          string `gorm:"size:1" json:"genero"` // M / F (opcional)
	FPV         int    `gorm:"" json:"fpv"` // Nullable (puede no tener FPV aún)
	Telefono    string `gorm:"size:50" json:"telefono"`
	Correo      string `gorm:"size:255;not null" json:"correo"`
	FechaNacimiento *time.Time `gorm:"type:date" json:"fecha_nacimiento"`

	// ── Datos académicos ──────────────────────────────────────────────────
	TituloUniversidad     string     `gorm:"size:255" json:"titulo_universidad"`
	TituloFechaGraduacion *time.Time `gorm:"type:date" json:"titulo_fecha_graduacion"`
	TituloMencion         string     `gorm:"size:255" json:"titulo_mencion"`
	TituloRegistroNumero  string     `gorm:"size:100" json:"titulo_registro_numero"`
	TituloRegistroEstado  string     `gorm:"size:100" json:"titulo_registro_estado"`
	TituloRegistroTomo    string     `gorm:"size:100" json:"titulo_registro_tomo"`
	TituloRegistroFolio   string     `gorm:"size:100" json:"titulo_registro_folio"`
	RIF                   string     `gorm:"size:50" json:"rif"`

	// ── Ubicación (espejo de la ficha interna) ────────────────────────────
	// Según la ubicación declarada por el solicitante: Carabobo, otro estado
	// venezolano o el exterior. Al aprobar se traspasan a psi_users.
	ServiceAddress              string `gorm:"size:255" json:"service_address"`
	MunicipalityCarabobo        string `gorm:"size:255" json:"municipality_carabobo"`
	StateOutside                string `gorm:"size:255" json:"state_outside"`
	MunicipalityOutSideCarabobo string `gorm:"size:255" json:"municipality_outside_carabobo"`
	Country                     string `gorm:"size:255" json:"country"`

	// ── Modalidad de servicio (espejo de la ficha interna) ───────────────
	ServiceModalityPresencial bool `gorm:"default:false" json:"service_modality_presencial"`
	ServiceModalityDistance   bool `gorm:"default:false" json:"service_modality_distance"`
	ServiceModalityTelephone  bool `gorm:"default:false" json:"service_modality_telephone"`

	// ── Áreas de trabajo (FKs al catálogo de especialidades) ──────────────
	PrimarySpecialtyID   *uint32 `gorm:"column:primary_specialty_id" json:"primary_specialty_id,omitempty"`
	SecondarySpecialtyID *uint32 `gorm:"column:secondary_specialty_id" json:"secondary_specialty_id,omitempty"`

	// ── Archivos (S3 keys) ────────────────────────────────────────────────
	FotoS3Key        string `gorm:"size:512" json:"foto_s3_key"`
	ComprobanteS3Key string `gorm:"size:512" json:"comprobante_s3_key"`

	// ── Fotos de documentos requeridos (cedula, titulo, rif, otro) ─────────
	// Un documento por categoría (uniqueIndex compuesto). Al aprobar migran a
	// psi_user_documents. La foto y el comprobante son columnas propias arriba.
	Documents []PsiInscriptionDocument `gorm:"foreignKey:InscriptionRequestID" json:"-"`

	// ── Estado administrativo ─────────────────────────────────────────────
	Status        InscriptionStatus `gorm:"size:20;default:pending;index:idx_inscription_requests_status" json:"status"`
	// uniqueIndex parcial: control_number no vacío y único (los pending aún no
	// tienen número; el comité lo asigna al aprobar).
	ControlNumber string            `gorm:"size:50;uniqueIndex:idx_inscription_requests_control_number,where:control_number <> '' AND control_number IS NOT NULL" json:"control_number"`
	Notes         string            `gorm:"type:text" json:"notes"`

	// ── Vínculo ───────────────────────────────────────────────────────────
	// PsiUserID se rellena al aprobar: id del expediente creado en psi_users.
	PsiUserID *uuid.UUID `gorm:"type:uuid;index:idx_inscription_requests_psi_user_id" json:"psi_user_id,omitempty"`
}

// TableName devuelve el nombre de la tabla en la base de datos.
func (PsiInscriptionRequest) TableName() string { return "psi_inscription_requests" }

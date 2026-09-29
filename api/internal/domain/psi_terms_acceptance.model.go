// api/internal/domain/psi_terms_acceptance.model.go
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// =========================================================================
// ACEPTACIÓN DE TÉRMINOS Y CONDICIONES (PARTE II — AGREMIADOS)
// =========================================================================

// ErrTermsVersionInvalid se devuelve cuando el cliente envía una versión de los
// términos que no existe en el backend. El servidor es la única fuente de la
// versión vigente (constante TermsVersion): aceptar "lo que el navegador tiene"
// permitiría saltarse cambios del documento.
var ErrTermsVersionInvalid = errors.New("la versión de los términos indicada no es la vigente")

// PsiTermsAcceptance registra que un agremiado leyó y aceptó la Parte II de los
// Términos y Condiciones.
//
// DISEÑO: tabla APPEND-ONLY. Cada aceptación inserta una fila nueva y nunca se
// actualiza ni se borra (no hay UPDATE ni DELETE en el repositorio). Eso hace
// que la fila sea evidencia por sí misma: no se puede reescribir el pasado para
// fingir que alguien aceptó una versión que no vio. Un agremiado que acepta de
// nuevo (porque el texto cambió) genera una segunda fila con otra versión.
//
// El bloqueo del portal ante una versión no aceptada es responsibility del
// frontend (modal cliente); el backend nunca bloquea el acceso.
type PsiTermsAcceptance struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`

	AuditModel

	// PsiUserID es el agremiado que aceptó. ON DELETE NO ACTION en la FK: borrar
	// un psicólogo no debe arrastrar su historial de aceptación (y de hecho la
	// baja de agremiados es lógica, ver psi_users.deleted_at).
	PsiUserID uuid.UUID `gorm:"type:uuid;index;not null" json:"psi_user_id"`

	// Version es la versión del documento aceptada, p. ej. "2026-09-29".
	// Debe coincidir con la constante TermsVersion del backend.
	Version string `gorm:"size:32;not null" json:"version"`

	// AcceptedAt es el instante de la aceptación. Es redundante con CreatedAt
	// (AuditModel) a propósito: CreatedAt lo escribe GORM, AcceptedAt lo
	// escribe el servicio; separarlos permite detectar cualquier manipulación
	// futura de la columna de auditoría.
	AcceptedAt *time.Time `json:"accepted_at"`

	// IP y UserAgent se guardan como evidencia del contexto de la aceptación.
	// Mismo tratamiento que el resto de la bitácora: en claro y sin retención
	// (ver la Parte I §10.1, que declara esta limitación de forma expresa).
	IP        string `gorm:"size:45" json:"ip"`
	UserAgent string `json:"user_agent"`
}

func (PsiTermsAcceptance) TableName() string { return "psi_terms_acceptance" }

// TermsVersion es la versión vigente de los Términos y Condiciones.
//
// La declara el BACKEND y no el frontend a propósito: si cada lado tuviera su
// propia constante, un despliegue de la web sin el de la API (o al revés)
// dejaría al agremiado viendo un texto distinto del que el servidor cree
// vigente, y el control del modal dejaría de funcionar sin que nadie lo notara.
// El frontend la recibe por GET /psi/me/terms y la usa solo para mostrarla.
const TermsVersion = "2026-09-29"

// PsiTermsAcceptanceRepository abstrae el registro de aceptaciones de la Parte II.
type PsiTermsAcceptanceRepository interface {
	// Create inserta una aceptación. No hay Update ni Delete: la tabla es
	// append-only (ver PsiTermsAcceptance).
	Create(ctx context.Context, acceptance *PsiTermsAcceptance) error

	// GetByUserAndVersion devuelve la aceptación de ese agremiado para esa versión
	// exacta; (nil, nil) si no la hay. Es la consulta que decide si el modal
	// debe mostrarse.
	GetByUserAndVersion(ctx context.Context, psiUserID uuid.UUID, version string) (*PsiTermsAcceptance, error)

	// ListByUser devuelve todo el historial de aceptaciones del agremiado,
	// de la más reciente a la más antigua.
	ListByUser(ctx context.Context, psiUserID uuid.UUID) ([]PsiTermsAcceptance, error)
}

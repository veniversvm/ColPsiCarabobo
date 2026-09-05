// api/internal/request_structs/inscription_requests.go

// Package request_structs define los Data Transfer Objects (DTOs) utilizados para
// estructurar las peticiones (Requests) y respuestas (Responses) de la API.
//
// Este archivo contiene los DTOs para el módulo de pre-inscripción de profesionales.
package request_structs

import (
	"time"

	"github.com/google/uuid"
)

// =========================================================================
// CHECK DE UNICIDAD (PÚBLICO)
// =========================================================================

// CheckCIRequest define los query params del endpoint público de validación de cédula.
type CheckCIRequest struct {
	CI int `query:"ci" validate:"required"`
}

// CheckFPVRequest define los query params del endpoint público de validación de FPV.
type CheckFPVRequest struct {
	FPV int `query:"fpv" validate:"required"`
}

// CheckEmailRequest define los query params del endpoint público de validación de correo.
type CheckEmailRequest struct {
	Correo string `query:"correo" validate:"required,email"`
}

// UniquenessCheckResponse es la respuesta de los endpoints check-ci / check-fpv / check-email.
type UniquenessCheckResponse struct {
	Exists  bool   `json:"exists"`
	Message string `json:"message,omitempty"`
}

// =========================================================================
// LISTADO ADMIN
// =========================================================================

// InscriptionListFilter define los filtros de listado de solicitudes para el admin.
type InscriptionListFilter struct {
	Status string `query:"status"` // pending | approved | rejected
	Q      string `query:"q"`      // búsqueda por nombre o cédula
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

// InscriptionListDTO es el View Model para el listado administrativo de solicitudes.
type InscriptionListDTO struct {
	ID            uuid.UUID `json:"id"`
	Cedula        int       `json:"cedula"`
	Nombres       string    `json:"nombres"`
	Apellidos     string    `json:"apellidos"`
	FPV           int       `json:"fpv"`
	Correo        string    `json:"correo"`
	Status        string    `json:"status"`
	ControlNumber string    `json:"control_number"`
	CreatedAt     time.Time `json:"created_at"`
}

// InscriptionListResponse es la respuesta paginada de listado.
type InscriptionListResponse struct {
	Items      []InscriptionListDTO `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

// =========================================================================
// DETALLE ADMIN
// =========================================================================

// InscriptionDocumentDTO es el View Model de una foto de documento de la ficha.
type InscriptionDocumentDTO struct {
	ID               uuid.UUID `json:"id"`
	DocumentType     string    `json:"document_type"`
	URL              string    `json:"url"`
	Title            string    `json:"title,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	OriginalFilename string    `json:"original_filename,omitempty"`
}

// InscriptionDetailDTO es el View Model del detalle de una solicitud.
// Contiene las URLs públicas resueltas de los archivos S3.
type InscriptionDetailDTO struct {
	ID            uuid.UUID             `json:"id"`
	Cedula        int                   `json:"cedula"`
	Nacionalidad  string                `json:"nacionalidad"`
	Nombres       string                `json:"nombres"`
	Apellidos     string                `json:"apellidos"`
	SegundoNombre   string              `json:"segundo_nombre"`
	SegundoApellido string              `json:"segundo_apellido"`
	Genero          string              `json:"genero"`
	FPV           int                   `json:"fpv"`
	Telefono      string                `json:"telefono"`
	Correo        string                `json:"correo"`
	FechaNacimiento *time.Time          `json:"fecha_nacimiento"`
	TituloUniversidad     string        `json:"titulo_universidad"`
	TituloFechaGraduacion *time.Time    `json:"titulo_fecha_graduacion"`
	TituloMencion         string        `json:"titulo_mencion"`
	TituloRegistroNumero  string        `json:"titulo_registro_numero"`
	TituloRegistroEstado  string        `json:"titulo_registro_estado"`
	TituloRegistroTomo    string        `json:"titulo_registro_tomo"`
	TituloRegistroFolio   string        `json:"titulo_registro_folio"`
	RIF                   string        `json:"rif"`
	ServiceAddress              string `json:"service_address"`
	MunicipalityCarabobo        string `json:"municipality_carabobo"`
	StateOutside                string `json:"state_outside"`
	MunicipalityOutSideCarabobo string `json:"municipality_outside_carabobo"`
	Country                     string `json:"country"`
	ServiceModalityPresencial   bool   `json:"service_modality_presencial"`
	ServiceModalityDistance     bool   `json:"service_modality_distance"`
	ServiceModalityTelephone    bool   `json:"service_modality_telephone"`
	PrimarySpecialtyID          *uint32 `json:"primary_specialty_id,omitempty"`
	SecondarySpecialtyID        *uint32 `json:"secondary_specialty_id,omitempty"`
	FotoURL               string        `json:"foto_url"`
	ComprobanteURL        string        `json:"comprobante_url"`
	Documents             []InscriptionDocumentDTO `json:"documents"`
	Status                string        `json:"status"`
	ControlNumber         string        `json:"control_number"`
	Notes                 string        `json:"notes"`
	PsiUserID             *uuid.UUID    `json:"psi_user_id,omitempty"`
	SolvencyCount         int           `json:"solvency_count"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

// =========================================================================
// EDICIÓN ADMIN DE LA FICHA
// =========================================================================

// UpdateInscriptionRequest es el cuerpo JSON del PATCH /admin/inscripciones/:id.
// Semántica de reemplazo: el formulario admin envía todos los campos visibles.
type UpdateInscriptionRequest struct {
	Cedula         int    `json:"cedula"`
	Nacionalidad   string `json:"nacionalidad"`
	Nombres        string `json:"nombres"`
	Apellidos      string `json:"apellidos"`
	SegundoNombre  string `json:"segundo_nombre"`
	SegundoApellido string `json:"segundo_apellido"`
	Genero         string `json:"genero"`
	FPV            int    `json:"fpv"`
	Telefono       string `json:"telefono"`
	Correo         string `json:"correo"`
	FechaNacimiento *string `json:"fecha_nacimiento,omitempty"`

	TituloUniversidad     string `json:"titulo_universidad"`
	TituloFechaGraduacion *string `json:"titulo_fecha_graduacion,omitempty"`
	TituloMencion         string `json:"titulo_mencion"`
	TituloRegistroNumero  string `json:"titulo_registro_numero"`
	TituloRegistroEstado  string `json:"titulo_registro_estado"`
	TituloRegistroTomo    string `json:"titulo_registro_tomo"`
	TituloRegistroFolio   string `json:"titulo_registro_folio"`
	RIF                   string `json:"rif"`

	ServiceAddress              string `json:"service_address"`
	MunicipalityCarabobo        string `json:"municipality_carabobo"`
	StateOutside                string `json:"state_outside"`
	MunicipalityOutSideCarabobo string `json:"municipality_outside_carabobo"`
	Country                     string `json:"country"`

	ServiceModalityPresencial bool `json:"service_modality_presencial"`
	ServiceModalityDistance   bool `json:"service_modality_distance"`
	ServiceModalityTelephone  bool `json:"service_modality_telephone"`

	PrimarySpecialtyID   *uint32 `json:"primary_specialty_id,omitempty"`
	SecondarySpecialtyID *uint32 `json:"secondary_specialty_id,omitempty"`
}

// UpdateInscriptionPhotoRequest indica qué foto de la ficha se reemplaza.
type UpdateInscriptionPhotoRequest struct {
	Kind string // "foto" | "comprobante"
}

// AddInscriptionDocumentRequest indica la categoría del documento de la ficha.
type AddInscriptionDocumentRequest struct {
	DocumentType string // cedula | titulo | rif | otro
}

// =========================================================================
// APROBACIÓN / RECHAZO
// =========================================================================

// ApproveInscriptionResponse es la respuesta de aprobación de una solicitud.
type ApproveInscriptionResponse struct {
	Message       string    `json:"message"`
	PsiUserID     uuid.UUID `json:"psi_user_id"`
	ControlNumber string    `json:"control_number"`
	EmailSent     bool      `json:"email_sent"`
}

// =========================================================================
// NOTAS Y CONTACTO (ADMIN)
// =========================================================================

// UpdateNotesRequest contiene las notas administrativas de una solicitud.
type UpdateNotesRequest struct {
	Notes string `json:"notes"`
}

// SendEmailToApplicantRequest define el correo que el admin envía al solicitante.
type SendEmailToApplicantRequest struct {
	Subject string `json:"subject" validate:"required"`
	Message string `json:"message" validate:"required"`
}

// SendEmailToApplicantResponse es la respuesta del envío de correo al solicitante.
type SendEmailToApplicantResponse struct {
	EmailSent bool `json:"email_sent"`
}

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

// UniquenessCheckResponse es la respuesta de los endpoints check-ci / check-fpv.
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

// InscriptionDetailDTO es el View Model del detalle de una solicitud.
// Contiene las URLs públicas resueltas de los archivos S3.
type InscriptionDetailDTO struct {
	ID            uuid.UUID             `json:"id"`
	Cedula        int                   `json:"cedula"`
	Nacionalidad  string                `json:"nacionalidad"`
	Nombres       string                `json:"nombres"`
	Apellidos     string                `json:"apellidos"`
	FPV           int                   `json:"fpv"`
	Telefono      string                `json:"telefono"`
	Correo        string                `json:"correo"`
	FechaNacimiento *time.Time          `json:"fecha_nacimiento"`
	TituloUniversidad     string        `json:"titulo_universidad"`
	TituloFechaGraduacion *time.Time    `json:"titulo_fecha_graduacion"`
	TituloMencion         string        `json:"titulo_mencion"`
	TituloRegistroNumero  string        `json:"titulo_registro_numero"`
	TituloRegistroEstado  string        `json:"titulo_registro_estado"`
	RIF                   string        `json:"rif"`
	FotoURL               string        `json:"foto_url"`
	ComprobanteURL        string        `json:"comprobante_url"`
	Status                string        `json:"status"`
	ControlNumber         string        `json:"control_number"`
	Notes                 string        `json:"notes"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
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

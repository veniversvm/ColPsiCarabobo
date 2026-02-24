package request_structs

import "github.com/google/uuid"

// ==========================================
// PSI SELF: EDICIÓN RESTRINGIDA
// ==========================================

type PsiUserUpdateRequestSelf struct {
	// Datos de Contacto y Privacidad (Permitido)
	ContactEmail             *string `json:"contact_email"`
	ShowContactEmail         *bool   `json:"show_contact_email"`
	PublicPhone              *string `json:"public_phone"`
	ShowPublicPhone          *bool   `json:"show_public_phone"`
	ServiceAddress           *string `json:"service_address"`
	ShowPublicServiceAddress *bool   `json:"show_public_service_address"`

	// Ubicación (Permitido)
	MunicipalityCarabobo        *string `json:"municipality_carabobo"`
	PhoneCarabobo               *string `json:"phone_carabobo"`
	CelPhoneCarabobo            *string `json:"cel_phone_carabobo"`
	StateOutside                *string `json:"state_outside"`
	MunicipalityOutSideCarabobo *string `json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo        *string `json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo     *string `json:"cel_phone_outside_carabobo"`

	// Especialidades (Permitido, si no está en la tabla de catálogo)
	PrimarySpecialty   *string `json:"primary_specialty"`
	SecondarySpecialty *string `json:"secondary_specialty"`

	// Biografía (Permitido)
	MiniBio *string `json:"mini_bio"`
	// Nota: BioTextID debe manejarse por un endpoint separado si el texto es muy grande

	// Campos de ColData
	ShowUniversityUndergraduate *bool `json:"show_university_undergraduate"`
	ShowGraduateDate            *bool `json:"show_graduate_date"`
	ShowMentionUndergraduate    *bool `json:"show_mention_undergraduate"`
}

// ==========================================//
// POSTGRADOS (CREADOS POR EL USUARIO)		 //
// ==========================================//

type CreatePostGradeRequest struct {
	Title          string `form:"title" json:"title" validate:"required"`
	University     string `form:"university" json:"university" validate:"required"`
	GraduationYear string `form:"graduation_year" json:"graduation_year" validate:"required"`
	Description    string `form:"description" json:"description"`
}

// PsiDirectoryFilterDTO captura los parámetros de la URL
type PsiDirectoryFilterDTO struct {
	SearchTerm string // q: Nombre, Apellido, CI, FPV
	Specialty  string // filtro secundario
	Page       int
	Limit      int
}

// PsiMiniProfileDTO es la respuesta optimizada para el listado público.
type PsiMiniProfileDTO struct {
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	CI             int       `json:"ci"`
	FPV            int       `json:"fpv"`
	ProfilePicture string    `json:"profile_picture"` // S3 Key o URL
	MiniBio        string    `json:"mini_bio"`
	Solvent        bool      `json:"solvent"` // Útil para que el frontend muestre un badge
}

// PsiFullProfileDTO representa la ficha pública completa del psicólogo.
type PsiFullProfileDTO struct {
	// Identidad Pública
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	FPV            int       `json:"fpv"`
	Gender         string    `json:"gender"`
	ProfilePicture string    `json:"profile_picture"`
	Solvent        bool      `json:"solvent"`

	// Contacto (Condicional)
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`

	// Ubicación
	Location struct {
		State        string `json:"state"`
		Municipality string `json:"municipality"`
		FullAddress  string `json:"full_address,omitempty"`
	} `json:"location"`

	// Profesional y Académico
	Specialties []string `json:"specialties"`
	MiniBio     string   `json:"mini_bio"`

	// Datos Universitarios (Condicional)
	Undergraduate struct {
		University string `json:"university,omitempty"`
		Date       string `json:"date,omitempty"`
		Mention    string `json:"mention,omitempty"`
	} `json:"undergraduate"`

	// Postgrados (Siempre visibles si existen)
	PostGrades     []PostGradeDTO     `json:"post_grades,omitempty"`
	SocialNetworks []SocialNetworkDTO `json:"social_networks,omitempty"`
}

type PostGradeDTO struct {
	Title      string `json:"title"`
	University string `json:"university"`
	Year       string `json:"year"`
}

type PsiLoginRequest struct {
	Identifier string `json:"identifier" validate:"required" example:"psicologo@email.com"`
	Password   string `json:"password" validate:"required" example:"12345678"`
}

type UpdatePostGradeRequest struct {
	Title          *string `form:"title"`
	University     *string `form:"university"`
	GraduationYear *string `form:"graduation_year"`
	Description    *string `form:"description"`
	// Las imágenes se manejan directo en el Handler
}

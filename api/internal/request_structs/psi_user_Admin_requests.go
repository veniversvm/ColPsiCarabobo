package request_structs

import "github.com/google/uuid"

// CreatePsiAdminRequest para creación individual por parte del personal administrativo
type CreatePsiAdminRequest struct {
	// Auth & Identidad
	Username       string `json:"username" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	FirstName      string `json:"first_name" validate:"required"`
	SecondName     string `json:"second_name"`
	LastName       string `json:"last_name" validate:"required"`
	SecondLastName string `json:"second_last_name" validate:"required"`

	// Datos filiatorios
	CI          int    `json:"ci" validate:"required"`
	BornDate    string `json:"born_date" example:"1990-01-01" validate:"required,datetime=2006-01-02"`
	Genre       string `json:"genre" validate:"required,oneof=M F"`
	Nationality string `json:"nationality" validate:"required,oneof=V E"`

	// FPV es el Número de Federación de Psicólogos de Venezuela, obligatorio para registro administrativo
	FPV int `json:"fpv" validate:"required"`

	// Estatus Administrativo (Solo Admin)
	Solvent     bool `json:"solvent"`
	ProofOfLife bool `json:"proof_of_life"`
	IsActive    bool `json:"is_active"`

	// Datos de contacto publico
	ContactEmail   string `json:"contact_email"`
	PublicPhone    string `json:"public_phone"`
	ServiceAddress string `json:"service_address"`

	// Si el psicologo esta fuera del estado Carabobo
	StateOutside                string `json:"state_outside"`
	MunicipalityOutSideCarabobo string `json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo        string `json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo     string `json:"cel_phone_outside_carabobo"`

	// Especialidades
	PrimarySpecialty   string `json:"primary_specialty"`
	SecondarySpecialty string `json:"secondary_specialty"`

	// Dejamos las bio como responsabilidad del psicologo, no es necesario que el admin la llene

	// --- DATOS COLEGIALES --- //
	// Undergraduate Data
	UniversityUndergraduate string `json:"university_undergraduate"`
	GraduateDate            string `json:"graduate_date"`
	MentionUndergraduate    string `json:"mention_undergraduate"`

	// Register Title: Datos de registro legal del título en el estado.
	RegisterNumber     int    `json:"register_number"`
	RegisterTitleState string `json:"register_title_state"`
	RegisterTitleDate  string `json:"register_title_date"`
	RegisterFolio      string `json:"register_folio"`
	RegisterTome       string `json:"register_tome"`

	// Professional Flags: Roles y estatus especiales dentro del gremio.
	GuildDirector       bool `json:"guild_director"`
	SixtyFiveOrPlus     bool `json:"sixty_five_or_plus"`
	GuildCollaborator   bool `json:"guild_collaborator"`
	PublicEmployee      bool `json:"public_employee"`
	UniversityProfessor bool `json:"university_professor"`

	// Histórico de solvencia y membresías dobles.
	DateOfLastSolvency string `json:"date_of_last_solvency"`
	DoubleGuild        bool   `json:"double_guild"`
	CPSM               bool   `json:"cpsm"`
}

// UpdatePsiAdminRequest permite al admin modificar CUALQUIER campo
type UpdatePsiAdminRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`

	// Identidad (Solo Admin)
	FirstName      *string `json:"first_name"`
	SecondName     *string `json:"second_name"`
	LastName       *string `json:"last_name"`
	SecondLastName *string `json:"second_last_name"`
	Email          *string `json:"email"`

	// FPV es el Número de Federación de Psicólogos de Venezuela, obligatorio para registro administrativo
	FPV *int `json:"fpv"`

	// datos filiatorios (Solo Admin)
	CI          *int    `json:"ci"`
	BornDate    *string `json:"born_date" example:"1990-01-01"`
	Genre       *string `json:"genre"`
	Nationality *string `json:"nationality"`

	// Estatus Administrativo (Solo Admin)
	Solvent     *bool `json:"solvent"`
	ProofOfLife *bool `json:"proof_of_life"`
	IsActive    *bool `json:"is_active"`

	// Datos de contacto y ubicación
	ContactEmail       *string `json:"contact_email"`
	ShowContactEmail   *bool   `json:"show_contact_email"`
	PublicPhone        *string `json:"public_phone"`
	ShowPublicPhone    *bool   `json:"show_public_phone"`
	ServiceAddress     *string `json:"service_address"`
	ShowServiceAddress *bool   `json:"show_service_address"`

	// Si el psicologo esta fuera del estado Carabobo
	StateOutside                *string `json:"state_outside"`
	MunicipalityOutSideCarabobo *string `json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo        *string `json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo     *string `json:"cel_phone_outside_carabobo"`

	// Especialidades
	PrimarySpecialty   *string `json:"primary_specialty"`
	SecondarySpecialty *string `json:"secondary_specialty"`

	// En caso de que el admin quiera modificar las bios, aunque lo ideal es que el psicologo sea el encargado de esto
	MiniBio *string `json:"mini_bio"`
	FullBio *string `json:"full_bio"`

	// Datos Colegiales (ColData - Solo Admin puede alterar la info)
	UniversityUndergraduate *string `json:"university_undergraduate"`
	GraduateDate            *string `json:"graduate_date"`
	MentionUndergraduate    *string `json:"mention_undergraduate"`
	RegisterNumber          *int    `json:"register_number"`
	RegisterTitleState      *string `json:"register_title_state"`
	RegisterTitleDate       *string `json:"register_title_date"`
	RegisterFolio           *string `json:"register_folio"`
	RegisterTome            *string `json:"register_tome"`

	// Flags Profesionales (Solo Admin)
	GuildDirector       *bool `json:"guild_director"`
	SixtyFiveOrPlus     *bool `json:"sixty_five_or_plus"`
	GuildCollaborator   *bool `json:"guild_collaborator"`
	PublicEmployee      *bool `json:"public_employee"`
	UniversityProfessor *bool `json:"university_professor"`

	// Histórico de solvencia y membresías dobles.
	DateOfLastSolvency *string `json:"date_of_last_solvency"`
	DoubleGuild        *bool   `json:"double_guild"`
	CPSM               *bool   `json:"cpsm"`
}

type PsiAdminListDTO struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CI        int       `json:"ci"`
	FPV       int       `json:"fpv"`
	Email     string    `json:"email"`
	Solvent   bool      `json:"solvent"`   // Para mostrar estado de deuda
	IsActive  bool      `json:"is_active"` // Para mostrar si está suspendido/inactivo
}

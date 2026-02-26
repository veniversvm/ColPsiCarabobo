// api/internal/request_structs/psi_user_Admin_requests.go

// Package request_structs define los Data Transfer Objects (DTOs) utilizados para
// estructurar las peticiones (Requests) y respuestas (Responses) de la API.
// Aísla los contratos de la API de los modelos de base de datos internos.
package request_structs

import "github.com/google/uuid"

// =========================================================================
// OPERACIONES ADMINISTRATIVAS (ADMIN ONLY)
// =========================================================================

// CreatePsiAdminRequest define el payload esperado cuando un Administrador registra
// manualmente a un nuevo psicólogo en el sistema.
// Nota de Privacidad: Los campos booleanos de "Show..." (visibilidad pública) se omiten
// deliberadamente aquí para que el sistema los inicialice en 'false' por defecto,
// forzando al usuario a optar por mostrarlos (Opt-In).
type CreatePsiAdminRequest struct {
	// --- Auth & Identidad ---
	Username       string `json:"username" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	FirstName      string `json:"first_name" validate:"required"`
	SecondName     string `json:"second_name"`
	LastName       string `json:"last_name" validate:"required"`
	SecondLastName string `json:"second_last_name" validate:"required"`

	// --- Datos Filiatorios y Legales ---
	CI          int    `json:"ci" validate:"required"`
	BornDate    string `json:"born_date" example:"1990-01-01" validate:"required,datetime=2006-01-02"`
	Genre       string `json:"genre" validate:"required,oneof=M F"`
	Nationality string `json:"nationality" validate:"required,oneof=V E"`

	// FPV es el Número de Federación de Psicólogos de Venezuela, obligatorio para registro administrativo.
	FPV int `json:"fpv" validate:"required"`

	// --- Estatus Administrativo (Exclusivo de Admins) ---
	Solvent     bool `json:"solvent"`
	ProofOfLife bool `json:"proof_of_life"`
	IsActive    bool `json:"is_active"`

	// --- Datos de Contacto ---
	ContactEmail   string `json:"contact_email"`
	PublicPhone    string `json:"public_phone"`
	ServiceAddress string `json:"service_address"`

	// --- Ubicación (Fuera del estado Carabobo) ---
	StateOutside                string `json:"state_outside"`
	MunicipalityOutSideCarabobo string `json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo        string `json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo     string `json:"cel_phone_outside_carabobo"`

	// --- Perfil Profesional ---
	PrimarySpecialty   string `json:"primary_specialty"`
	SecondarySpecialty string `json:"secondary_specialty"`
	// Nota: La biografía (MiniBio/FullBio) se delega al psicólogo en su autogestión.

	// --- DATOS COLEGIALES (PsiUserColData) ---

	// Información de Pregrado
	UniversityUndergraduate string `json:"university_undergraduate"`
	GraduateDate            string `json:"graduate_date"`
	MentionUndergraduate    string `json:"mention_undergraduate"`

	// Registro de Título Estatal
	RegisterNumber     int    `json:"register_number"`
	RegisterTitleState string `json:"register_title_state"`
	RegisterTitleDate  string `json:"register_title_date"`
	RegisterFolio      string `json:"register_folio"`
	RegisterTome       string `json:"register_tome"`

	// Banderas Profesionales / Gremiales
	GuildDirector       bool `json:"guild_director"`
	SixtyFiveOrPlus     bool `json:"sixty_five_or_plus"`
	GuildCollaborator   bool `json:"guild_collaborator"`
	PublicEmployee      bool `json:"public_employee"`
	UniversityProfessor bool `json:"university_professor"`

	// Historial de Gremio
	DateOfLastSolvency string `json:"date_of_last_solvency"`
	DoubleGuild        bool   `json:"double_guild"`
	CPSM               bool   `json:"cpsm"`
}

// UpdatePsiAdminRequest permite al administrador modificar CUALQUIER campo del psicólogo.
// Arquitectura Senior: Todos los campos (excepto el ID) son PUNTEROS (*).
// Esto implementa la semántica PATCH real: si un campo no se envía en el JSON (es nil),
// el servicio sabe que NO debe actualizarlo en la base de datos.
type UpdatePsiAdminRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`

	// --- Identidad y Filiación (Solo Admin puede corregir esto) ---
	FirstName      *string `json:"first_name"`
	SecondName     *string `json:"second_name"`
	LastName       *string `json:"last_name"`
	SecondLastName *string `json:"second_last_name"`
	Email          *string `json:"email"`
	FPV            *int    `json:"fpv"`
	CI             *int    `json:"ci"`
	BornDate       *string `json:"born_date" example:"1990-01-01"`
	Genre          *string `json:"genre"`
	Nationality    *string `json:"nationality"`

	// --- Estatus Administrativo ---
	Solvent     *bool `json:"solvent"`
	ProofOfLife *bool `json:"proof_of_life"`
	IsActive    *bool `json:"is_active"`

	// --- Datos de Contacto y Visibilidad ---
	ContactEmail       *string `json:"contact_email"`
	ShowContactEmail   *bool   `json:"show_contact_email"`
	PublicPhone        *string `json:"public_phone"`
	ShowPublicPhone    *bool   `json:"show_public_phone"`
	ServiceAddress     *string `json:"service_address"`
	ShowServiceAddress *bool   `json:"show_service_address"`

	// --- Ubicación ---
	StateOutside                *string `json:"state_outside"`
	MunicipalityOutSideCarabobo *string `json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo        *string `json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo     *string `json:"cel_phone_outside_carabobo"`

	// --- Perfil Profesional ---
	PrimarySpecialty   *string `json:"primary_specialty"`
	SecondarySpecialty *string `json:"secondary_specialty"`
	MiniBio            *string `json:"mini_bio"`
	FullBio            *string `json:"full_bio"` // Por si el admin necesita moderar el texto

	// --- DATOS COLEGIALES (ColData) ---
	UniversityUndergraduate *string `json:"university_undergraduate"`
	GraduateDate            *string `json:"graduate_date"`
	MentionUndergraduate    *string `json:"mention_undergraduate"`
	RegisterNumber          *int    `json:"register_number"`
	RegisterTitleState      *string `json:"register_title_state"`
	RegisterTitleDate       *string `json:"register_title_date"`
	RegisterFolio           *string `json:"register_folio"`
	RegisterTome            *string `json:"register_tome"`

	// --- Banderas Profesionales ---
	GuildDirector       *bool   `json:"guild_director"`
	SixtyFiveOrPlus     *bool   `json:"sixty_five_or_plus"`
	GuildCollaborator   *bool   `json:"guild_collaborator"`
	PublicEmployee      *bool   `json:"public_employee"`
	UniversityProfessor *bool   `json:"university_professor"`
	DateOfLastSolvency  *string `json:"date_of_last_solvency"`
	DoubleGuild         *bool   `json:"double_guild"`
	CPSM                *bool   `json:"cpsm"`
}

// =========================================================================
// RESPUESTAS Y LISTADOS (VIEW MODELS)
// =========================================================================

// PsiAdminListDTO es el modelo de respuesta ligero para el Dashboard Administrativo (DataGrid).
// Proyección de Datos: Solo envía lo necesario para renderizar la tabla y aplicar colores
// (badges) basados en la solvencia y actividad, ahorrando ancho de banda.
type PsiAdminListDTO struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CI        int       `json:"ci"`
	FPV       int       `json:"fpv"`
	Email     string    `json:"email"`

	// Flags críticos para la interfaz de usuario administrativa
	Solvent  bool `json:"solvent"`   // Permite renderizar etiqueta Verde/Roja (Deudas)
	IsActive bool `json:"is_active"` // Permite renderizar opacidad si el usuario está suspendido
}

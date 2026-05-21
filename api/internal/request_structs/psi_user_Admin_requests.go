// api/internal/request_structs/psi_user_Admin_requests.go

// Package request_structs define los Data Transfer Objects (DTOs) utilizados para
// estructurar las peticiones (Requests) y respuestas (Responses) de la API.
// Aísla los contratos de la API de los modelos de base de datos internos.
package request_structs

import (
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// =========================================================================
// OPERACIONES ADMINISTRATIVAS (ADMIN ONLY)
// =========================================================================

// CreatePsiAdminRequest define el payload cuando un Administrador registra
// manualmente a un nuevo psicólogo en el sistema.
//
// Nota de Privacidad: Los campos booleanos Show* (visibilidad pública) se omiten
// deliberadamente para que el sistema los inicialice en 'false' por defecto,
// forzando al psicólogo a optar activamente por exponerlos (Opt-In Privacy).
type CreatePsiAdminRequest struct {

	// ── Auth & Credenciales ───────────────────────────────────────────────
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`

	// ── Identidad Legal ───────────────────────────────────────────────────
	FirstName      string `json:"first_name" validate:"required"`
	SecondName     string `json:"second_name"`
	LastName       string `json:"last_name" validate:"required"`
	SecondLastName string `json:"second_last_name"`
	CI             int    `json:"ci" validate:"required"`
	FPV            int    `json:"fpv" validate:"required"` // Número de Federación de Psicólogos de Venezuela
	Nationality    string `json:"nationality" validate:"required,oneof=V E"`
	BornDate       string `json:"born_date" example:"1990-01-01" validate:"required,datetime=2006-01-02"`
	Genre          string `json:"genre" validate:"required,oneof=M F"`

	// ── Estatus Administrativo ────────────────────────────────────────────
	// Campos exclusivos del admin — el psicólogo no puede auto-asignarse estos valores.
	Solvent     bool `json:"solvent"`
	ProofOfLife bool `json:"proof_of_life"`
	IsActive    bool `json:"is_active"`

	// ── Contacto Gremial y Público ────────────────────────────────────────
	ContactEmail     string `json:"contact_email"`
	ContactPhone     string `json:"contact_phone"`      // Reemplaza a public_phone
	ContactCellPhone string `json:"contact_cell_phone"` // Nuevo
	ServiceAddress   string `json:"service_address"`

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	// Para miembros con consulta o residencia dentro del estado.
	// MunicipalityCarabobo debe validarse contra el catálogo de municipios de Carabobo.
	MunicipalityCarabobo string `json:"municipality_carabobo"`
	PhoneCarabobo        string `json:"phone_carabobo"`
	CelPhoneCarabobo     string `json:"cel_phone_carabobo"`

	// ── Ubicación: Fuera de Carabobo (Venezuela) ─────────────────────────
	// Para miembros en otros estados venezolanos.
	// StateOutside debe validarse contra el catálogo de estados, excluyendo Carabobo.
	StateOutside                  string `json:"state_outside"`
	MunicipalityOutSideCarabobo   string `json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo          string `json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo       string `json:"cel_phone_outside_carabobo"`
	ServiceAddressOutSideCarabobo string `json:"service_address_outside_carabobo"`

	// ── Ubicación: Fuera de Venezuela ─────────────────────────────────────
	// Para miembros en el exterior. Country debe usar código ISO 3166-1 alpha-2.
	Country                        string `json:"country"`
	PhoneOutSideVenezuela          string `json:"phone_outside_venezuela"`
	CellPhoneOutSideVenezuela      string `json:"cell_phone_outside_venezuela"` // Nuevo
	ServiceAddressOutSideVenezuela string `json:"service_address_outside_venezuela"`

	// ── Perfil Profesional ────────────────────────────────────────────────
	// Las áreas de trabajo deben corresponder a entradas activas en el catálogo (WorkArea).
	// La biografía (MiniBio/FullBio) se delega al psicólogo para su autogestión.
	PrimaryWorkArea   string `json:"primary_work_area"`   // Reemplaza a primary_specialty
	SecondaryWorkArea string `json:"secondary_work_area"` // Reemplaza a secondary_specialty

	// ── Datos Colegiales: Pregrado e Inscripción ──────────────────────────
	GuildInscriptionDate    string `json:"guild_inscription_date"` // Nuevo
	UniversityUndergraduate string `json:"university_undergraduate"`
	GraduateDate            string `json:"graduate_date"`
	MentionUndergraduate    string `json:"mention_undergraduate"`

	// ── Datos Colegiales: Registro Legal del Título ───────────────────────
	RegisterTitleState string `json:"register_title_state"`
	RegisterTitleDate  string `json:"register_title_date"`
	RegisterNumber     int    `json:"register_number"`
	RegisterFolio      string `json:"register_folio"`
	RegisterTome       string `json:"register_tome"`

	// ── Datos Colegiales: Flags Gremiales ────────────────────────────────
	GuildDirector       bool `json:"guild_director"`       // Miembro de la Junta Directiva
	SixtyFiveOrPlus     bool `json:"sixty_five_or_plus"`   // Mayor de 65 años (tarifa diferenciada)
	GuildCollaborator   bool `json:"guild_collaborator"`   // Colaborador activo del Colegio
	PublicEmployee      bool `json:"public_employee"`      // Empleado público
	Discapacity         bool `json:"discapacity"`          // Nuevo: Posee alguna discapacidad
	UniversityProfessor bool `json:"university_professor"` // Docente universitario

	// ── Datos Colegiales: Historial Gremial ──────────────────────────────
	DateOfLastSolvency  string `json:"date_of_last_solvency"` // Última cuota saldada
	DoubleGuild         bool   `json:"double_guild"`          // Colegiado en más de un estado
	DoubleGuildLocation string `json:"double_guild_location"` // Nuevo: Dónde tiene la doble colegiatura
	CPSM                bool   `json:"cpsm"`                  // Miembro del Colegio de Psicólogos de Miranda
}

// UpdatePsiAdminRequest permite al administrador modificar CUALQUIER campo del psicólogo.
// Arquitectura Senior: Todos los campos (excepto el ID) son PUNTEROS (*).
// Esto implementa la semántica PATCH real: si un campo no se envía en el JSON (es nil),
// el servicio sabe que NO debe actualizarlo en la base de datos.
type UpdatePsiAdminRequest struct {
	ID uuid.UUID `json:"id" form:"id" validate:"required"`

	// ── Credenciales de acceso ────────────────────────────────────────────
	Username *string `json:"username" form:"username"`
	Email    *string `json:"email" form:"email"`

	// ── Identidad legal ───────────────────────────────────────────────────
	FirstName      *string `json:"first_name" form:"first_name"`
	SecondName     *string `json:"second_name" form:"second_name"`
	LastName       *string `json:"last_name" form:"last_name"`
	SecondLastName *string `json:"second_last_name" form:"second_last_name"`
	FPV            *int    `json:"fpv" form:"fpv"`
	CI             *int    `json:"ci" form:"ci"`
	BornDate       *string `json:"born_date" form:"born_date"`
	Genre          *string `json:"genre" form:"genre"`
	Nationality    *string `json:"nationality" form:"nationality"`

	// ── Estado gremial y multimedia ───────────────────────────────────────
	Solvent       *bool  `json:"solvent" form:"solvent"`
	ProofOfLife   *bool  `json:"proof_of_life" form:"proof_of_life"`
	IsActive      *bool  `json:"is_active" form:"is_active"`
	SolvenciesRaw string `json:"solvencies" form:"solvencies"`

	// ── Contacto interno del gremio ────────────────────────────────────
	ContactPhone     *string `json:"contact_phone" form:"contact_phone"`
	ContactCellPhone *string `json:"contact_cell_phone" form:"contact_cell_phone"`

	// ── Contacto público y privacidad ────────────────────────────────────
	ContactEmail                *string `json:"contact_email" form:"contact_email"`
	ShowContactEmailRaw         string  `json:"show_contact_email" form:"show_contact_email"`
	ServiceAddress              *string `json:"service_address" form:"service_address"`
	ShowPublicServiceAddressRaw string  `json:"show_public_service_address" form:"show_public_service_address"`

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	MunicipalityCarabobo        *string `json:"municipality_carabobo" form:"municipality_carabobo"`
	ShowMunicipalityCaraboboRaw string  `json:"show_municipality_carabobo" form:"show_municipality_carabobo"`
	PhoneCarabobo               *string `json:"phone_carabobo" form:"phone_carabobo"`
	ShowPhoneCaraboboRaw        string  `json:"show_phone_carabobo" `
	CellPhoneCarabobo           *string `json:"cell_phone_carabobo" form:"cell_phone_carabobo"`
	ShowCelPhoneCaraboboRaw     string  `json:"show_cel_phone_carabobo" form:"show_cel_phone_carabobo"`

	// ── Ubicación: Fuera de Carabobo (Venezuela) ─────────────────────────
	StateOutside                               *string `json:"state_outside" form:"state_outside"`
	ShowStateOutsideRaw                        string  `json:"show_state_outside" form:"show_state_outside"`
	MunicipalityOutSideCarabobo                *string `json:"municipality_outside_carabobo" form:"municipality_outside_carabobo"`
	ShowMunicipalityOutSideCaraboboRaw         string  `json:"show_municipality_outside_carabobo" form:"show_municipality_outside_carabobo"`
	PhoneOutSideCarabobo                       *string `json:"phone_outside_carabobo" form:"phone_outside_carabobo"`
	ShowPhoneOutSideCaraboboRaw                string  `json:"show_phone_outside_carabobo" form:"show_phone_outside_carabobo"`
	CelPhoneOutSideCarabobo                    *string `json:"cel_phone_outside_carabobo" form:"cel_phone_outside_carabobo"`
	ShowCellPhoneOutSideCaraboboRaw            string  `json:"show_cel_phone_outside_carabobo" form:"show_cel_phone_outside_carabobo"`
	ServiceAddressOutSideCarabobo              *string `json:"service_address_outside_carabobo" form:"service_address_outside_carabobo"`
	ShowPublicServiceAddressOutSideCaraboboRaw string  `json:"show_public_service_address_outside_carabobo" form:"show_public_service_address_outside_carabobo"`

	// ── Ubicación: Fuera de Venezuela ─────────────────────────────────────
	Country                                     *string `json:"country" form:"country"`
	PhoneOutSideVenezuela                       *string `json:"phone_outside_venezuela" form:"phone_outside_venezuela"`
	ShowPhoneOutSideVenezuelaRaw                string  `json:"show_phone_outside_venezuela" form:"show_phone_outside_venezuela"`
	CellPhoneOutSideVenezuela                   *string `json:"cell_phone_outside_venezuela" form:"cell_phone_outside_venezuela"`
	ShowCellPhoneOutSideVenezuelaRaw            string  `json:"show_cell_phone_outside_venezuela" form:"show_cell_phone_outside_venezuela"`
	ServiceAddressOutSideVenezuela              *string `json:"service_address_outside_venezuela" form:"service_address_outside_venezuela"`
	ShowPublicServiceAddressOutSideVenezuelaRaw string  `json:"show_public_service_address_outside_venezuela" form:"show_public_service_address_outside_venezuela"`

	// --- Perfil Profesional ---
	PrimaryWorkArea   *string `json:"primary_work_area" form:"primary_work_area"`
	SecondaryWorkArea *string `json:"secondary_work_area" form:"secondary_work_area"`
	MiniBio           *string `json:"mini_bio" form:"mini_bio"`
	FullBio           *string `json:"full_bio" form:"full_bio"`

	// ─────────────── Datos Colegiales ─────────────── //
	// ── Registro legal del título ─────────────────────────────────────────
	GuildInscriptionDate *string `json:"guild_inscription_date" form:"guild_inscription_date"`
	RegisterNumber       *int    `json:"register_number" form:"register_number"`
	RegisterTitleState   *string `json:"register_title_state" form:"register_title_state"`
	RegisterTitleDate    *string `json:"register_title_date" form:"register_title_date"`
	RegisterFolio        *string `json:"register_folio" form:"register_folio"`
	RegisterTome         *string `json:"register_tome" form:"register_tome"`

	// ── Solvencia y membresías ────────────────────────────────────────────
	DateOfLastSolvency  *string `json:"date_of_last_solvency" form:"date_of_last_solvency"`
	DoubleGuildLocation *string `json:"double_guild_location" form:"double_guild_location"`

	// ── Pregrado ──────────────────────────────────────────────────────────
	UniversityUndergraduate        *string `json:"university_undergraduate" form:"university_undergraduate"`
	ShowUniversityUndergraduateRaw string  `json:"show_university_undergraduate" form:"show_university_undergraduate"`
	GraduateDate                   *string `json:"graduate_date"`
	ShowGraduateDateRaw            string  `json:"show_graduate_date" form:"show_graduate_date"`
	MentionUndergraduate           *string `json:"mention_undergraduate"`
	ShowMentionUndergraduateRaw    string  `json:"show_mention_undergraduate" form:"show_mention_undergraduate"`

	// ── Flags gremiales ───────────────────────────────────────────────────
	// Roles y estatus especiales dentro de la estructura del Colegio.y
	GuildDirector       *bool `json:"guild_director" form:"guild_director"`
	SixtyFiveOrPlus     *bool `json:"sixty_five_or_plus" form:"sixty_five_or_plus"`
	GuildCollaborator   *bool `json:"guild_collaborator" form:"guild_collaborator"`
	PublicEmployee      *bool `json:"public_employee" form:"public_employee"`
	Discapacity         *bool `json:"discapacity" form:"discapacity"`
	UniversityProfessor *bool `json:"university_professor" form:"university_professor"`
	DoubleGuild         *bool `json:"double_guild" form:"double_guild"`
	CPSM                *bool `json:"cpsm" form:"cpsm"`

	// ── CAMPOS RAW (Para capturar los booleanos del Form/JSON como string) ──

}

// ── GETTERS DE VISIBILIDAD (Sincronizados con el struct) ──

func (r *UpdatePsiAdminRequest) ShowContactEmail() *bool {
	return utils.BoolFromForm(r.ShowContactEmailRaw)
}
func (r *UpdatePsiAdminRequest) ShowPublicServiceAddress() *bool {
	return utils.BoolFromForm(r.ShowPublicServiceAddressRaw)
}

func (r *UpdatePsiAdminRequest) ShowMunicipalityCarabobo() *bool {
	return utils.BoolFromForm(r.ShowMunicipalityCaraboboRaw)
}
func (r *UpdatePsiAdminRequest) ShowPhoneCarabobo() *bool {
	return utils.BoolFromForm(r.ShowPhoneCaraboboRaw)
}
func (r *UpdatePsiAdminRequest) ShowCelPhoneCarabobo() *bool {
	return utils.BoolFromForm(r.ShowCelPhoneCaraboboRaw)
}

func (r *UpdatePsiAdminRequest) ShowStateOutside() *bool {
	return utils.BoolFromForm(r.ShowStateOutsideRaw)
}
func (r *UpdatePsiAdminRequest) ShowMunicipalityOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowMunicipalityOutSideCaraboboRaw)
}
func (r *UpdatePsiAdminRequest) ShowPhoneOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowPhoneOutSideCaraboboRaw)
}
func (r *UpdatePsiAdminRequest) ShowCellPhoneOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowCellPhoneOutSideCaraboboRaw)
}
func (r *UpdatePsiAdminRequest) ShowPublicServiceAddressOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowPublicServiceAddressOutSideCaraboboRaw)
}

func (r *UpdatePsiAdminRequest) ShowPhoneOutSideVenezuela() *bool {
	return utils.BoolFromForm(r.ShowPhoneOutSideVenezuelaRaw)
}
func (r *UpdatePsiAdminRequest) ShowCellPhoneOutSideVenezuela() *bool {
	return utils.BoolFromForm(r.ShowCellPhoneOutSideVenezuelaRaw)
}
func (r *UpdatePsiAdminRequest) ShowPublicServiceAddressOutSideVenezuela() *bool {
	return utils.BoolFromForm(r.ShowPublicServiceAddressOutSideVenezuelaRaw)
}

func (r *UpdatePsiAdminRequest) ShowUniversityUndergraduate() *bool {
	return utils.BoolFromForm(r.ShowUniversityUndergraduateRaw)
}
func (r *UpdatePsiAdminRequest) ShowGraduateDate() *bool {
	return utils.BoolFromForm(r.ShowGraduateDateRaw)
}
func (r *UpdatePsiAdminRequest) ShowMentionUndergraduate() *bool {
	return utils.BoolFromForm(r.ShowMentionUndergraduateRaw)
}

type SolvenciesUpdate struct {
	ID   uuid.UUID `json:"id"`   // Opcional, por si envías solvencias existentes
	Date string    `json:"date"` // Formato "YYYY-MM-DD" o RFC3339
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

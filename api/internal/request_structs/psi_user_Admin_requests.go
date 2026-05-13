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

	// ── Contacto Público ──────────────────────────────────────────────────
	ContactEmail   string `json:"contact_email"`
	PublicPhone    string `json:"public_phone"`
	ServiceAddress string `json:"service_address"`

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
	ServiceAddressOutSideVenezuela string `json:"service_address_outside_venezuela"`

	// ── Perfil Profesional ────────────────────────────────────────────────
	// Las especialidades deben corresponder a entradas activas en PsiSpecialtyModel.
	// La biografía (MiniBio/FullBio) se delega al psicólogo para su autogestión.
	PrimarySpecialty   string `json:"primary_specialty"`
	SecondarySpecialty string `json:"secondary_specialty"`

	// ── Datos Colegiales: Pregrado ────────────────────────────────────────
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
	UniversityProfessor bool `json:"university_professor"` // Docente universitario

	// ── Datos Colegiales: Historial Gremial ──────────────────────────────
	DateOfLastSolvency string `json:"date_of_last_solvency"` // Última cuota saldada
	DoubleGuild        bool   `json:"double_guild"`          // Colegiado en más de un estado
	CPSM               bool   `json:"cpsm"`                  // Miembro del Colegio de Psicólogos de Miranda
}

// UpdatePsiAdminRequest permite al administrador modificar CUALQUIER campo del psicólogo.
// Arquitectura Senior: Todos los campos (excepto el ID) son PUNTEROS (*).
// Esto implementa la semántica PATCH real: si un campo no se envía en el JSON (es nil),
// el servicio sabe que NO debe actualizarlo en la base de datos.
type UpdatePsiAdminRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`

	// --- Identidad y Filiación ---
	FirstName      *string `json:"first_name"`
	SecondName     *string `json:"second_name"`
	LastName       *string `json:"last_name"`
	SecondLastName *string `json:"second_last_name"`
	Username       *string `json:"username"`
	Email          *string `json:"email"`
	FPV            *int    `json:"fpv"`
	CI             *int    `json:"ci"`
	BornDate       *string `json:"born_date" example:"1990-01-01"`
	Genre          *string `json:"genre"`
	Nationality    *string `json:"nationality"`

	// --- Estatus Administrativo ---
	Solvent     *bool               `json:"solvent"`
	Solvencies  *[]SolvenciesUpdate `json:"solvencies"`
	ProofOfLife *bool               `json:"proof_of_life"`
	IsActive    *bool               `json:"is_active"`

	// --- Datos de Contacto y Visibilidad ---
	ContactEmail   *string `json:"contact_email"`
	PublicPhone    *string `json:"public_phone"`
	ServiceAddress *string `json:"service_address"`

	ShowContactEmailRaw         string `json:"show_contact_email"`
	ShowPublicPhoneRaw          string `json:"show_public_phone"`
	ShowPublicServiceAddressRaw string `json:"show_public_service_address"` // ← fix: form → json

	// --- Ubicación: Carabobo ---
	MunicipalityCarabobo *string `json:"municipality_carabobo"`
	PhoneCarabobo        *string `json:"phone_carabobo"`
	CelPhoneCarabobo     *string `json:"cel_phone_carabobo"`

	// --- Ubicación: Fuera de Carabobo (Venezuela) ---
	StateOutside                  *string `json:"state_outside"`
	MunicipalityOutSideCarabobo   *string `json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo          *string `json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo       *string `json:"cel_phone_outside_carabobo"`
	ServiceAddressOutSideCarabobo *string `json:"service_address_outside_carabobo"`

	ShowPhoneOutSideCaraboboRaw          string `json:"show_phone_outside_carabobo"`
	ShowCellPhoneOutSideCaraboboRaw      string `json:"show_cel_phone_outside_carabobo"`
	ShowServiceAddressOutSideCaraboboRaw string `json:"show_public_service_address_outside_carabobo"`

	// --- Ubicación: Fuera de Venezuela ---
	Country                        *string `json:"country"`
	PhoneOutSideVenezuela          *string `json:"phone_outside_venezuela"`
	ServiceAddressOutSideVenezuela *string `json:"service_address_outside_venezuela"`

	ShowPhoneOutSideVenezuelaRaw          string `json:"show_phone_outside_venezuela"`
	ShowCellPhoneOutSideVenezuelaRaw      string `json:"show_cel_phone_outside_venezuela"`
	ShowServiceAddressOutSideVenezuelaRaw string `json:"show_public_service_address_outside_venezuela"`

	// --- Perfil Profesional ---
	PrimarySpecialty   *string `json:"primary_specialty"`
	SecondarySpecialty *string `json:"secondary_specialty"`
	MiniBio            *string `json:"mini_bio"`
	FullBio            *string `json:"full_bio"`

	// --- DATOS COLEGIALES (ColData) ---
	UniversityUndergraduate *string `json:"university_undergraduate"`
	GraduateDate            *string `json:"graduate_date"`
	MentionUndergraduate    *string `json:"mention_undergraduate"`
	RegisterNumber          *int    `json:"register_number"`
	RegisterTitleState      *string `json:"register_title_state"`
	RegisterTitleDate       *string `json:"register_title_date"`
	RegisterFolio           *string `json:"register_folio"`
	RegisterTome            *string `json:"register_tome"`

	ShowUniversityUndergraduateRaw string `json:"show_university_undergraduate"`
	ShowGraduateDateRaw            string `json:"show_graduate_date"`
	ShowMentionUndergraduateRaw    string `json:"show_mention_undergraduate"`

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

type SolvenciesUpdate struct {
	ID             uuid.UUID `json:"id" validate:"required"`
	PsiUserModelID uuid.UUID `json:"psi_user_model_id"`
	Date           string    `json:"date"`
}

// ── Getters: Contacto y Visibilidad ──────────────────────────────────────────

func (r *UpdatePsiAdminRequest) ShowContactEmail() *bool {
	return utils.BoolFromForm(r.ShowContactEmailRaw)
}
func (r *UpdatePsiAdminRequest) ShowPublicPhone() *bool {
	return utils.BoolFromForm(r.ShowPublicPhoneRaw)
}
func (r *UpdatePsiAdminRequest) ShowPublicServiceAddress() *bool {
	return utils.BoolFromForm(r.ShowPublicServiceAddressRaw)
}

// ── Getters: Fuera de Carabobo ────────────────────────────────────────────────

func (r *UpdatePsiAdminRequest) ShowPhoneOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowPhoneOutSideCaraboboRaw)
}
func (r *UpdatePsiAdminRequest) ShowCellPhoneOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowCellPhoneOutSideCaraboboRaw)
}
func (r *UpdatePsiAdminRequest) ShowPublicServiceAddressOutSideCarabobo() *bool { // ← fix: quitado Raw del nombre
	return utils.BoolFromForm(r.ShowServiceAddressOutSideCaraboboRaw)
}

// ── Getters: Fuera de Venezuela ───────────────────────────────────────────────

func (r *UpdatePsiAdminRequest) ShowPhoneOutSideVenezuela() *bool {
	return utils.BoolFromForm(r.ShowPhoneOutSideVenezuelaRaw)
}
func (r *UpdatePsiAdminRequest) ShowCellPhoneOutSideVenezuela() *bool {
	return utils.BoolFromForm(r.ShowCellPhoneOutSideVenezuelaRaw)
}
func (r *UpdatePsiAdminRequest) ShowPublicServiceAddressOutSideVenezuela() *bool { // ← fix: quitado Raw del nombre
	return utils.BoolFromForm(r.ShowServiceAddressOutSideVenezuelaRaw)
}

// ── Getters: Datos Colegiales ─────────────────────────────────────────────────

func (r *UpdatePsiAdminRequest) ShowUniversityUndergraduate() *bool {
	return utils.BoolFromForm(r.ShowUniversityUndergraduateRaw)
}
func (r *UpdatePsiAdminRequest) ShowGraduateDate() *bool {
	return utils.BoolFromForm(r.ShowGraduateDateRaw)
}
func (r *UpdatePsiAdminRequest) ShowMentionUndergraduate() *bool {
	return utils.BoolFromForm(r.ShowMentionUndergraduateRaw)
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

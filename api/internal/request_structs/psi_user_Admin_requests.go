// api/internal/request_structs/psi_user_Admin_requests.go

// Package request_structs define los Data Transfer Objects (DTOs) utilizados para
// estructurar las peticiones (Requests) y respuestas (Responses) de la API.
//
// Funciona como una Capa Anticorrupción: aísla los contratos públicos de la API
// de los modelos de dominio y base de datos internos, garantizando que cambios
// en la DB no rompan a los clientes, y que los clientes no puedan inyectar
// datos no permitidos en la DB.
package request_structs

import (
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// =========================================================================
// OPERACIONES ADMINISTRATIVAS (ADMIN ONLY)
// =========================================================================

// CreatePsiAdminRequest define el payload esperado cuando un Administrador registra
// manualmente a un nuevo psicólogo en el sistema (Bypass del registro público).
//
// Privacidad por Diseño (Opt-In Privacy):
// Los campos booleanos de visibilidad (Show*) se omiten deliberadamente de este DTO.
// Al no existir en la creación, la base de datos los inicializará en 'false',
// forzando al psicólogo a entrar a su perfil y elegir activamente qué datos
// desea hacer públicos, cumpliendo con principios de protección de datos.
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
	// Campos críticos: Solo el administrador puede asignar estos valores.
	Solvent     bool `json:"solvent"`
	ProofOfLife bool `json:"proof_of_life"`
	IsActive    bool `json:"is_active"`

	// ── Contacto Gremial y Público ────────────────────────────────────────
	ContactEmail     string `json:"contact_email"`
	ContactPhone     string `json:"contact_phone"`
	ContactCellPhone string `json:"contact_cell_phone"`
	ServiceAddress   string `json:"service_address"`

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	MunicipalityCarabobo string `json:"municipality_carabobo"`
	PhoneCarabobo        string `json:"phone_carabobo"`
	CelPhoneCarabobo     string `json:"cel_phone_carabobo"`

	// ── Ubicación: Fuera de Carabobo (Venezuela) ─────────────────────────
	StateOutside                  string `json:"state_outside"`
	MunicipalityOutSideCarabobo   string `json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo          string `json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo       string `json:"cel_phone_outside_carabobo"`
	ServiceAddressOutSideCarabobo string `json:"service_address_outside_carabobo"`

	// ── Ubicación: Fuera de Venezuela ─────────────────────────────────────
	Country                        string `json:"country"`
	PhoneOutSideVenezuela          string `json:"phone_outside_venezuela"`
	CellPhoneOutSideVenezuela      string `json:"cell_phone_outside_venezuela"`
	ServiceAddressOutSideVenezuela string `json:"service_address_outside_venezuela"`

	// ── Perfil Profesional ────────────────────────────────────────────────
	PrimaryWorkArea      string  `json:"primary_work_area"`
	SecondaryWorkArea    string  `json:"secondary_work_area"`
	PrimarySpecialtyID   *uint32 `json:"primary_specialty_id"`
	SecondarySpecialtyID *uint32 `json:"secondary_specialty_id"`

	// ── Datos Colegiales: Pregrado e Inscripción ──────────────────────────
	GuildInscriptionDate    string `json:"guild_inscription_date"`
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
	GuildDirector       bool `json:"guild_director"`
	SixtyFiveOrPlus     bool `json:"sixty_five_or_plus"`
	GuildCollaborator   bool `json:"guild_collaborator"`
	PublicEmployee      bool `json:"public_employee"`
	Discapacity         bool `json:"discapacity"`
	UniversityProfessor bool `json:"university_professor"`

	// ── Datos Colegiales: Historial Gremial ──────────────────────────────
	DateOfLastSolvency  string `json:"date_of_last_solvency"`
	DoubleGuild         bool   `json:"double_guild"`
	DoubleGuildLocation string `json:"double_guild_location"`
	CPSM                bool   `json:"cpsm"`
}

// UpdatePsiAdminRequest permite al administrador modificar datos de un psicólogo.
//
// Implementación de Semántica PATCH Real:
// A excepción del ID, todos los campos son PUNTEROS (*). Esto permite al servidor
// distinguir entre un valor que se envía explícitamente vacío ("") o falso (false),
// frente a un valor que simplemente no se envió en el payload (nil).
// Si el valor es nil, el repositorio sabe que NO debe actualizar esa columna.
//
// Patrón de Campos 'Raw':
// Para soportar peticiones `multipart/form-data` (usadas comúnmente al subir imágenes),
// los booleanos se reciben como strings (Raw) y se parsean mediante getters. Esto evita
// los problemas clásicos de los frameworks Go al mapear checkboxes HTML no marcados.
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
	SolvenciesRaw string `json:"solvencies" form:"solvencies"` // JSON serializado en string para multipart

	// ── Contacto interno del gremio ───────────────────────────────────────
	ContactPhone     *string `json:"contact_phone" form:"contact_phone"`
	ContactCellPhone *string `json:"contact_cell_phone" form:"contact_cell_phone"`

	// ── Contacto público y privacidad ─────────────────────────────────────
	ContactEmail                *string `json:"contact_email" form:"contact_email"`
	ShowContactEmailRaw         string  `json:"show_contact_email" form:"show_contact_email"`
	ServiceAddress              *string `json:"service_address" form:"service_address"`
	ShowPublicServiceAddressRaw string  `json:"show_public_service_address" form:"show_public_service_address"`

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	MunicipalityCarabobo        *string `json:"municipality_carabobo" form:"municipality_carabobo"`
	ShowMunicipalityCaraboboRaw string  `json:"show_municipality_carabobo" form:"show_municipality_carabobo"`
	PhoneCarabobo               *string `json:"phone_carabobo" form:"phone_carabobo"`
	ShowPhoneCaraboboRaw        string  `json:"show_phone_carabobo" form:"show_phone_carabobo"` // Añadido el tag form faltante
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

	// ── Perfil Profesional ────────────────────────────────────────────────
	PrimaryWorkArea      *string `json:"primary_work_area" form:"primary_work_area"`
	SecondaryWorkArea    *string `json:"secondary_work_area" form:"secondary_work_area"`
	PrimarySpecialtyID   *uint32 `json:"primary_specialty_id" form:"primary_specialty_id"`
	SecondarySpecialtyID *uint32 `json:"secondary_specialty_id" form:"secondary_specialty_id"`
	MiniBio              *string `json:"mini_bio" form:"mini_bio"`
	FullBio              *string `json:"full_bio" form:"full_bio"`

	// ── Modalidad de servicio ─────────────────────────────────────────────
	ServiceModalityPresencialRaw string `json:"service_modality_presencial" form:"service_modality_presencial"`
	ServiceModalityDistanceRaw   string `json:"service_modality_distance" form:"service_modality_distance"`
	ServiceModalityTelephoneRaw  string `json:"service_modality_telephone" form:"service_modality_telephone"`
	ShowServiceModalityRaw       string `json:"show_service_modality" form:"show_service_modality"`

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
	GraduateDate                   *string `json:"graduate_date" form:"graduate_date"`
	ShowGraduateDateRaw            string  `json:"show_graduate_date" form:"show_graduate_date"`
	MentionUndergraduate           *string `json:"mention_undergraduate" form:"mention_undergraduate"`
	ShowMentionUndergraduateRaw    string  `json:"show_mention_undergraduate" form:"show_mention_undergraduate"`

	// ── Flags gremiales ───────────────────────────────────────────────────
	GuildDirector       *bool `json:"guild_director" form:"guild_director"`
	SixtyFiveOrPlus     *bool `json:"sixty_five_or_plus" form:"sixty_five_or_plus"`
	GuildCollaborator   *bool `json:"guild_collaborator" form:"guild_collaborator"`
	PublicEmployee      *bool `json:"public_employee" form:"public_employee"`
	Discapacity         *bool `json:"discapacity" form:"discapacity"`
	UniversityProfessor *bool `json:"university_professor" form:"university_professor"`
	DoubleGuild         *bool `json:"double_guild" form:"double_guild"`
	CPSM                *bool `json:"cpsm" form:"cpsm"`

	// Requisito legal: la administración confirma la inscripción del psicólogo
	// en el Ministerio de Educación (Art. 5 Ley de Ejercicio de la Psicología).
	MinistryRegistrationConfirmed *bool `json:"ministry_registration_confirmed" form:"ministry_registration_confirmed"`
}

// ── GETTERS DE VISIBILIDAD (Sanitización Multipart) ────────────────────────
// Estos métodos convierten los campos booleanos recibidos como strings (ej. "true", "on", "1")
// a punteros booleanos reales usando el helper utils.BoolFromForm.
// Devuelven nil si el campo no fue enviado, respetando la arquitectura PATCH.

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

// ── Modalidad de servicio ──
func (r *UpdatePsiAdminRequest) ServiceModalityPresencial() *bool {
	return utils.BoolFromForm(r.ServiceModalityPresencialRaw)
}
func (r *UpdatePsiAdminRequest) ServiceModalityDistance() *bool {
	return utils.BoolFromForm(r.ServiceModalityDistanceRaw)
}
func (r *UpdatePsiAdminRequest) ServiceModalityTelephone() *bool {
	return utils.BoolFromForm(r.ServiceModalityTelephoneRaw)
}
func (r *UpdatePsiAdminRequest) ShowServiceModality() *bool {
	return utils.BoolFromForm(r.ShowServiceModalityRaw)
}

// SolvenciesUpdate define la estructura para actualizar el historial de solvencias
// a través de un array JSON embebido en las peticiones administrativas.
type SolvenciesUpdate struct {
	ID   uuid.UUID `json:"id"`   // Opcional, permite hacer Upsert si se envía un ID existente
	Date string    `json:"date"` // Formato esperado: "YYYY-MM-DD" o RFC3339
}

// =========================================================================
// RESPUESTAS Y LISTADOS (VIEW MODELS)
// =========================================================================

// PsiAdminListDTO es el View Model diseñado específicamente para el DataGrid del Dashboard Administrativo.
//
// Optimización de Transferencia (Proyección DTO):
// Transfiere exclusivamente la metadata necesaria para pintar la tabla principal,
// ahorrando ancho de banda al no transferir textos extensos (biografías) ni imágenes.
type PsiAdminListDTO struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CI        int       `json:"ci"`
	FPV       int       `json:"fpv"`
	Email     string    `json:"email"`

	// Nº de control interno (columna 'Nº' del Excel maestro); visible solo admin.
	ControlNumber string `json:"control_number"`

	// Edad calculada por el backend desde born_date (años cumplidos). Solo admin.
	Age int `json:"age"`

	// Banderas UI: Permiten al Frontend renderizar componentes visuales (Badges) sin hacer lógica pesada.
	Solvent  bool `json:"solvent"`   // Activa etiquetas visuales de Alerta Roja / Verde.
	IsActive bool `json:"is_active"` // Activa la opacidad de la fila si el usuario está suspendido o baneado.
}

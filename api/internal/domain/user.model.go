// api/internal/domain/user.model.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// ADMINISTRACIÓN
// =============================================================================

// UserAdmin representa al personal administrativo del Colegio de Psicólogos.
// Implementa un sistema de permisos granulares (RBAC) para segmentar capacidades
// de gestión según el cargo del operador (Secretaría, Tesorería, IT, etc.).
type UserAdmin struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel

	// Credenciales de acceso al sistema administrativo.
	Username string `gorm:"size:25;unique;not null" json:"username"`
	Email    string `gorm:"size:255;unique;not null" json:"email"`
	Password string `gorm:"size:512;not null" json:"-"` // bcrypt — nunca expuesto en JSON
	Key      string `gorm:"size:512;" json:"-"`         // usado para invalidar sesiones (key rotation)
	IsActive bool   `gorm:"default:true" json:"is_active"`

	// Sudo otorga acceso total e irrevocable. Solo asignable fuera de la API.
	Sudo bool `gorm:"default:false" json:"-"`

	// ── Permisos: Gestión de Colegiados ──────────────────────────────────
	CanCreatePsi bool `gorm:"default:false" json:"can_create_psi"`
	CanUpdatePsi bool `gorm:"default:false" json:"can_update_psi"`
	CanDeletePsi bool `gorm:"default:false" json:"can_delete_psi"`

	// ── Permisos: Gestión de Personal Administrativo ─────────────────────
	CanCreateAdmin bool `gorm:"default:false" json:"can_create_admin"`
	CanUpdateAdmin bool `gorm:"default:false" json:"can_update_admin"`
	CanDeleteAdmin bool `gorm:"default:false" json:"can_delete_admin"`

	// ── Permisos: Contenido y Comunicación ───────────────────────────────
	CanPublish             bool `gorm:"default:false" json:"can_publish"`
	CanUpdatePublish       bool `gorm:"default:false" json:"can_update_publish"`
	CanDeletePublish       bool `gorm:"default:false" json:"can_delete_publish"`
	CanSendNotifications   bool `gorm:"default:false" json:"can_send_notifications"`
	CanManageNotifications bool `gorm:"default:false" json:"can_manage_notifications"`
	CanReadNotifications   bool `gorm:"default:false" json:"can_read_notifications"`

	// ── Permisos: Catálogo de Especialidades (Tags) ───────────────────────
	CanCreateTags bool `gorm:"default:false" json:"can_create_tags"`
	CanEditTags   bool `gorm:"default:false" json:"can_edit_tags"`
	CanDeleteTags bool `gorm:"default:false" json:"can_delete_tags"`
}

func (UserAdmin) TableName() string { return "user_admins" }

// =============================================================================
// PSICÓLOGOS — PERFIL CORE
// =============================================================================

// PsiUserModel es la entidad principal de un Psicólogo colegiado.
// Agrupa identidad legal, contacto, ubicación geográfica, estado gremial
// y relaciones con los demás módulos del dominio.
type PsiUserModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel

	// ── Credenciales de acceso ────────────────────────────────────────────
	Username string `gorm:"size:25;unique;not null" json:"username"`
	Email    string `gorm:"size:255;unique;not null" json:"email"` // Email institucional del gremio (login)
	Password string `gorm:"size:512;not null" json:"-"`
	Key      string `gorm:"size:512;" json:"-"`
	IsActive bool   `gorm:"column:is_active" json:"is_active"`

	// ── Identidad legal ───────────────────────────────────────────────────
	FirstName      string    `gorm:"size:255;not null" json:"first_name"`
	SecondName     string    `gorm:"size:255" json:"second_name"`
	LastName       string    `gorm:"size:255;not null" json:"last_name"`
	SecondLastName string    `gorm:"size:255" json:"second_last_name"`
	FPV            int       `gorm:"not null;uniqueIndex" json:"fpv"`    // Número de Federación de Psicólogos de Venezuela
	CI             int       `gorm:"not null;uniqueIndex" json:"ci"`     // Cédula de Identidad
	Nationality    string    `gorm:"size:1;not null" json:"nationality"` // V = venezolano, E = extranjero
	BornDate       time.Time `gorm:"type:date;not null" json:"born_date"`
	Genre          string    `gorm:"size:1;not null" json:"genre"` // M = masculino, F = femenino

	// ── Estado gremial y multimedia ───────────────────────────────────────
	Solvent             bool   `gorm:"default:false" json:"solvent"`              // Solvencia con el Colegio
	ProofOfLife         bool   `gorm:"column:proof_of_life" json:"proof_of_life"` // Fe de vida presentada
	ProfilePictureS3Key string `gorm:"size:512" json:"profile_picture_url"`       // S3 key de la foto de perfil

	// ── Contacto interno del gremio ────────────────────────────────────
	ContactPhone     string `gorm:"size:255;not null" json:"contact_phone"`
	ContactCellPhone string `gorm:"size:255;not null" json:"contact_cell_phone"`

	// ── Contacto público y privacidad ────────────────────────────────────
	// Cada campo sensible tiene un flag show_* que controla su visibilidad
	// en el directorio público. El psicólogo gestiona estos desde su perfil.
	ContactEmail             string `gorm:"size:255;not null" json:"contact_email"`
	ShowContactEmail         bool   `gorm:"default:false" json:"show_contact_email"`
	ServiceAddress           string `gorm:"size:255" json:"service_address"`
	ShowPublicServiceAddress bool   `gorm:"default:false" json:"show_public_service_address"`

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	// Para miembros residentes o con consulta dentro del estado Carabobo.
	// MunicipalityCarabobo debe restringirse al catálogo de municipios del estado.
	MunicipalityCarabobo     string `gorm:"size:255" json:"municipality_carabobo"`
	ShowMunicipalityCarabobo bool   `gorm:"size:255" json:"show_municipality_carabobo"`
	PhoneCarabobo            string `gorm:"default:false" json:"phone_carabobo"`
	ShowPhoneCarabobo        bool   `gorm:"default:false" json:"show_phone_carabobo"`
	CelPhoneCarabobo         string `gorm:"size:20" json:"cel_phone_carabobo"`
	ShowCelPhoneCarabobo     bool   `gorm:"default:false" json:"show_cel_phone_carabobo"`

	// ── Ubicación: Fuera de Carabobo (Venezuela) ─────────────────────────
	// Para miembros en otros estados venezolanos.
	// StateOutside debe restringirse al catálogo de estados de Venezuela, excluyendo Carabobo.
	StateOutside                            string `gorm:"size:255" json:"state_outside"`
	ShowStateOutside                        bool   `gorm:"default:false" json:"show_state_outside"`
	MunicipalityOutSideCarabobo             string `gorm:"size:255" json:"municipality_outside_carabobo"`
	ShowMunicipalityOutSideCarabobo         bool   `gorm:"default:false" json:"show_municipality_outside_carabobo"`
	PhoneOutSideCarabobo                    string `gorm:"size:20" json:"phone_outside_carabobo"`
	ShowPhoneOutSideCarabobo                bool   `gorm:"default:false" json:"show_phone_outside_carabobo"`
	CelPhoneOutSideCarabobo                 string `gorm:"size:20" json:"cel_phone_outside_carabobo"`
	ShowCellPhoneOutSideCarabobo            bool   `gorm:"default:false" json:"show_cel_phone_outside_carabobo"`
	ServiceAddressOutSideCarabobo           string `gorm:"size:255" json:"service_address_outside_carabobo"`
	ShowPublicServiceAddressOutSideCarabobo bool   `gorm:"default:false" json:"show_public_service_address_outside_carabobo"`

	// ── Ubicación: Fuera de Venezuela ─────────────────────────────────────
	// Para miembros en el exterior. Country debe usar código ISO 3166-1 alpha-2.
	Country                                  string `gorm:"size:255" json:"country"`
	PhoneOutSideVenezuela                    string `gorm:"size:20" json:"phone_outside_venezuela"`
	ShowPhoneOutSideVenezuela                bool   `gorm:"default:false" json:"show_phone_outside_venezuela"`
	CellPhoneOutSideVenezuela                string `gorm:"size:20" json:"cell_phone_outside_venezuela"`
	ShowCellPhoneOutSideVenezuela            bool   `gorm:"default:false" json:"show_cel_phone_outside_venezuela"`
	ServiceAddressOutSideVenezuela           string `gorm:"size:255" json:"service_address_outside_venezuela"`
	ShowPublicServiceAddressOutSideVenezuela bool   `gorm:"default:false" json:"show_public_service_address_outside_venezuela"`

	// ── Especialidades profesionales ──────────────────────────────────────
	// Almacenadas como strings para búsqueda directa en el directorio.
	// Deben corresponder a entradas activas en el catálogo PsiSpecialtyModel.
	PrimaryWorkArea   string `gorm:"size:50" json:"primary_work_area"`
	SecondaryWorkArea string `gorm:"size:50" json:"secondary_work_area"`

	// ── Biografía profesional ─────────────────────────────────────────────
	MiniBio   string    `json:"mini_bio"`              // Resumen corto (max 250 chars) para el directorio
	BioTextID uuid.UUID `json:"bio_text_id,omitempty"` // FK hacia TextModel (contenido HTML sanitizado)
	FullBio   TextModel `gorm:"foreignKey:BioTextID" json:"full_bio,omitempty"`

	// ── Relaciones ────────────────────────────────────────────────────────
	ColData        PsiUserColData         `gorm:"foreignKey:PsiUserModelID" json:"col_data"`
	PostGrades     []PsiUserPostGrade     `gorm:"foreignKey:PsiUserID" json:"post_grades"`
	SocialNetworks []PsiUserSocialNetwork `gorm:"foreignKey:PsiUserID" json:"social_networks"`
	Solvencies     []PsiUSerSolvency      `gorm:"foreignKey:PsiUserModelID" json:"solvencies"`
}

func (PsiUserModel) TableName() string { return "psi_users" }

// =============================================================================
// DATOS COLEGIALES
// =============================================================================

// PsiUserColData almacena el historial académico y los datos regulatorios
// del Colegio. Es una relación 1-a-1 con PsiUserModel (uniqueIndex en PsiUserModelID).
type PsiUserColData struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel
	PsiUserModelID       uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"psi_user_model_id"`
	GuildInscriptionDate time.Time `gorm:"type:date" json:"guild_inscription_date"`

	// ── Pregrado ──────────────────────────────────────────────────────────
	UniversityUndergraduate     string    `gorm:"size:255" json:"university_undergraduate"`
	ShowUniversityUndergraduate bool      `gorm:"default:false" json:"show_university_undergraduate"`
	GraduateDate                time.Time `gorm:"type:date" json:"graduate_date"`
	ShowGraduateDate            bool      `gorm:"default:false" json:"show_graduate_date"`
	MentionUndergraduate        string    `gorm:"size:255" json:"mention_undergraduate"`
	ShowMentionUndergraduate    bool      `gorm:"default:false" json:"show_mention_undergraduate"`

	// S3 Keys para imágenes del título de pregrado (máx. 3 archivos).
	TitleImageOneS3Key   string `gorm:"size:512" json:"title_image_one_url"`
	TitleImageTwoS3Key   string `gorm:"size:512" json:"title_image_two_url"`
	TitleImageThreeS3Key string `gorm:"size:512" json:"title_image_three_url"`

	// ── Registro legal del título ─────────────────────────────────────────
	RegisterTitleState string    `gorm:"size:255" json:"register_title_state"`
	RegisterTitleDate  time.Time `gorm:"type:date" json:"register_title_date"`
	RegisterNumber     int       `json:"register_number"`
	RegisterFolio      string    `gorm:"size:255" json:"register_folio"`
	RegisterTome       string    `gorm:"size:255" json:"register_tome"`

	// ── Flags gremiales ───────────────────────────────────────────────────
	// Roles y estatus especiales dentro de la estructura del Colegio.
	GuildDirector       bool `gorm:"default:false" json:"guild_director"`       // Miembro de la Junta Directiva
	SixtyFiveOrPlus     bool `gorm:"default:false" json:"sixty_five_or_plus"`   // Mayor de 65 años (tarifa diferenciada)
	GuildCollaborator   bool `gorm:"default:false" json:"guild_collaborator"`   // Colaborador activo del Colegio
	PublicEmployee      bool `gorm:"default:false" json:"public_employee"`      // Empleado público
	Discapacity         bool `gorm:"default:false" json:"discapacity"`          // Empleado público
	UniversityProfessor bool `gorm:"default:false" json:"university_professor"` // Docente universitario

	// ── Solvencia y membresías ────────────────────────────────────────────
	DateOfLastSolvency  time.Time `gorm:"type:date" json:"date_of_last_solvency"` // Última fecha de pago de cuota
	DoubleGuild         bool      `gorm:"default:false" json:"double_guild"`      // Colegiado en más de un estado
	DoubleGuildLocation string    `gorm:"size:255" json:"double_guild_location"`  // Colegiado en más de un estado
	CPSM                bool      `gorm:"default:false" json:"cpsm"`              // Miembro del Colegio de Psicólogos de Miranda
}

func (PsiUserColData) TableName() string { return "psi_user_col_data" }

// =============================================================================
// SOLVENCIA
// =============================================================================

// PsiUSerSolvency es un registro de las solvencia que posee el psicologo.
// Relación N-a-1 con PsiUserModel.
type PsiUSerSolvency struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AuditModel
	// El nombre del índice debe ser el mismo en ambos campos para crear una clave compuesta
	PsiUserModelID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_psi_solvency_unique" json:"psi_user_model_id"`
	Date           time.Time `gorm:"type:date;not null;uniqueIndex:idx_psi_solvency_unique" json:"date"`
}

func (PsiUSerSolvency) TableName() string { return "psi_user_solvency" }

// =============================================================================
// POSTGRADOS
// =============================================================================

// PostGradeType define el tipo de estudio de postgrado
type PostGradeType string

// Constantes que representan los valores permitidos del enum
const (
	Diplomado       PostGradeType = "diplomado"
	Especializacion PostGradeType = "especializacion"
	Maestria        PostGradeType = "maestria"
	Doctorado       PostGradeType = "doctorado"
)

// PsiUserPostGrade representa títulos académicos adicionales del psicólogo:
// Especializaciones, Maestrías, Doctorados, Diplomados, etc.
// Relación N-a-1 con PsiUserModel.
type PsiUserPostGrade struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel
	PsiUserID uuid.UUID     `gorm:"type:uuid;index" json:"psi_user_id"`
	Type      PostGradeType `gorm:"type:varchar(50);not null" json:"post_grade_type"`

	Title          string `gorm:"size:255;not null" json:"post_grade_title"`
	University     string `gorm:"size:255" json:"post_grade_university"`
	GraduationYear string `gorm:"size:50" json:"post_grade_graduation_year"`
	Description    string `gorm:"type:text" json:"post_grade_description"`
	Active         bool   `gorm:"default:true" json:"is_active"`

	// S3 Keys para certificados del postgrado (máx. 3 archivos).
	PicOneS3Key   string `gorm:"size:512" json:"pic_one_url"`
	PicTwoS3Key   string `gorm:"size:512" json:"pic_two_url"`
	PicThreeS3Key string `gorm:"size:512" json:"pic_three_url"`
}

// IsValid es un método de ayuda (opcional) para validar que el string entrante sea correcto
// Es muy útil validarlo antes de guardarlo en la base de datos.
func (p PostGradeType) IsValid() bool {
	switch p {
	case Diplomado, Especializacion, Maestria, Doctorado:
		return true
	}
	return false
}

func (PsiUserPostGrade) TableName() string { return "psi_user_post_grades" }

// =============================================================================
// OBSERVACIONES INTERNAS
// =============================================================================

// PsiObservations almacena notas internas del Colegio sobre un psicólogo.
// IMPORTANTE: Los psicólogos NUNCA pueden ver ni acceder a sus propias observaciones.
// Solo el personal administrativo autorizado puede crearlas, editarlas y leerlas.
// Relación N-a-1 con PsiUserModel.
type PsiObservations struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel
	PsiUserID uuid.UUID `gorm:"type:uuid;index" json:"psi_user_id"`
	Content   string    `gorm:"type:text" json:"content"`
}

func (PsiObservations) TableName() string { return "psi_observations" }

// =============================================================================
// EXPEDIENTE DEONTOLÓGICO
// =============================================================================

// PsiODeontologia registra los expedientes disciplinarios o deontológicos
// abiertos contra un psicólogo por el Tribunal Disciplinario del Colegio.
// Al igual que las observaciones, es de acceso exclusivo al personal autorizado.
// Relación N-a-1 con PsiUserModel.
type PsiODeontologia struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel
	PsiUserID uuid.UUID `gorm:"type:uuid;index" json:"psi_user_id"`
	Content   string    `gorm:"type:text" json:"content"`
}

func (PsiODeontologia) TableName() string { return "psi_deontologia" }

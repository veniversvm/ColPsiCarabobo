package domain

import (
	"time"

	"github.com/google/uuid"
)

// --- ADMINISTRACIÓN ---

// UserAdmin representa a los usuarios del personal administrativo del Colegio.
// Implementa un sistema de permisos granulares para segmentar las capacidades
// de gestión según el cargo del operador (ej. Secretaría, Tesorería, IT).
type UserAdmin struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	// AuditModel proporciona identidad única y trazabilidad.
	AuditModel

	// Username es el identificador único para el inicio de sesión.
	Username string `gorm:"size:25;unique;not null" json:"username"`

	// Email es el correo institucional del administrador.
	Email string `gorm:"size:255;unique;not null" json:"email"`

	// Password almacena el hash (bcrypt) de la credencial. Oculto en JSON por seguridad.
	Password string `gorm:"size:512;not null" json:"-"`

	// IsActive permite revocar el acceso al sistema sin borrar al usuario.
	IsActive bool `gorm:"default:true" json:"is_active"`

	// Key se utiliza para procesos de recuperación de cuenta o MFA.
	Key string `gorm:"size:512;" json:"-"`

	// Sudo identifica a los administradores del sistema con acceso total e irrevocable.
	Sudo bool `gorm:"default:false" json:"-"`

	// --- Permisos Granulares de Gestión de Colegiados ---
	CanCreatePsi bool `gorm:"default:false" json:"can_create_psi"`
	CanUpdatePsi bool `gorm:"default:false" json:"can_update_psi"`
	CanDeletePsi bool `gorm:"default:false" json:"can_delete_psi"`

	// --- Permisos de Gestión de Personal ---
	CanCreateAdmin bool `gorm:"default:false" json:"can_create_admin"`
	CanUpdateAdmin bool `gorm:"default:false" json:"can_update_admin"`
	CanDeleteAdmin bool `gorm:"default:false" json:"can_delete_admin"`

	// --- Permisos de Contenido y Comunicación ---
	CanPublish             bool `gorm:"default:false" json:"can_publish"`
	CanUpdatePublish       bool `gorm:"default:false" json:"can_update_publish"`
	CanDeletePublish       bool `gorm:"default:false" json:"can_delete_publish"`
	CanSendNotifications   bool `gorm:"default:false" json:"can_send_notifications"`
	CanManageNotifications bool `gorm:"default:false" json:"can_manage_notifications"`
	CanReadNotifications   bool `gorm:"default:false" json:"can_read_notifications"`

	// --- Permisos de Clasificación (Tags) ---
	CanCreateTags bool `gorm:"default:false" json:"can_create_tags"`
	CanEditTags   bool `gorm:"default:false" json:"can_edit_tags"`
	CanDeleteTags bool `gorm:"default:false" json:"can_delete_tags"`
}

func (UserAdmin) TableName() string {
	return "user_admins"
}

// --- PSICÓLOGOS (PERFIL CORE) ---

// PsiUserModel es la entidad principal que representa a un Psicólogo colegiado.
// Contiene datos de identidad, contacto y estado gremial.
type PsiUserModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel
	Username string `gorm:"size:25;unique;not null" json:"username"`
	Email    string `gorm:"size:255;unique;not null" json:"email"`
	Password string `gorm:"size:512;not null" json:"-"`
	Key      string `gorm:"size:512;" json:"-"`
	IsActive bool   `gorm:"default:true" json:"is_active"`

	// Identity: Datos básicos de identificación legal.
	FirstName      string `gorm:"size:255;not null" json:"first_name"`
	SecondName     string `gorm:"size:255" json:"second_name"`
	LastName       string `gorm:"size:255;not null" json:"last_name"`
	SecondLastName string `gorm:"size:255" json:"second_last_name"`

	// FPV es el Número de Federación de Psicólogos de Venezuela.
	FPV int `gorm:"not null;uniqueIndex" json:"fpv"`

	// CI es la Cédula de Identidad del profesional.
	CI          int       `gorm:"not null;uniqueIndex" json:"ci"`
	Nationality string    `gorm:"size:1;not null" json:"nationality"`
	BornDate    time.Time `gorm:"type:date;not null" json:"born_date"`
	Genre       string    `gorm:"size:1;not null" json:"genre"`

	// Contact & Privacy: Control de visibilidad para el directorio público.
	ContactEmail             string `gorm:"size:255;not null" json:"contact_email"`
	ShowContactEmail         bool   `gorm:"default:false" json:"show_contact_email"`
	PublicPhone              string `gorm:"size:20" json:"public_phone"`
	ShowPublicPhone          bool   `gorm:"default:false" json:"show_public_phone"`
	ServiceAddress           string `gorm:"size:255" json:"service_address"`
	ShowPublicServiceAddress bool   `gorm:"default:false" json:"show_public_service_address"`

	// Status & Files: Estado de solvencia y archivos multimedia.
	Solvent             bool   `gorm:"default:false" json:"solvent"`
	ProofOfLife         bool   `gorm:"default:false" json:"proof_of_life"`
	ProfilePictureS3Key string `gorm:"size:512" json:"profile_picture_url"`

	// Location Carabobo: Ubicación dentro de la jurisdicción principal.
	MunicipalityCarabobo string `gorm:"size:255" json:"municipality_carabobo"`
	PhoneCarabobo        string `gorm:"size:20" json:"phone_carabobo"`
	CelPhoneCarabobo     string `gorm:"size:20" json:"cel_phone_carabobo"`

	// Location Outside: Ubicación para miembros en el exterior o fuera de Carabobo.
	StateOutside                string `gorm:"size:255" json:"state_outside"`
	MunicipalityOutSideCarabobo string `gorm:"size:255" json:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo        string `gorm:"size:20" json:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo     string `gorm:"size:20" json:"cel_phone_outside_carabobo"`

	// Especialidades: Categorización profesional.
	PrimarySpecialty   string `gorm:"size:50" json:"primary_specialty"`
	SecondarySpecialty string `gorm:"size:50" json:"secondary_specialty"`

	// Bio: Información profesional resumida y detallada.
	MiniBio   string    `json:"mini_bio"`
	BioTextID uuid.UUID `json:"bio_text_id,omitempty"`

	// Relaciones: Conexión con datos académicos y gremiales.
	ColData        PsiUserColData         `gorm:"foreignKey:PsiUserModelID" json:"col_data"`
	PostGrades     []PsiUserPostGrade     `gorm:"foreignKey:PsiUserID" json:"post_grades"`
	SocialNetworks []PsiUserSocialNetwork `gorm:"foreignKey:PsiUserID" json:"social_networks"`
}

func (PsiUserModel) TableName() string {
	return "psi_users"
}

// --- DATOS COLEGIALES ---

// PsiUserColData almacena información histórica y regulatoria de la carrera
// académica y gremial del psicólogo.
type PsiUserColData struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel
	PsiUserModelID uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"psi_user_model_id"`

	// Undergraduate Data: Información sobre el título de pregrado.
	UniversityUndergraduate     string    `gorm:"size:255" json:"university_undergraduate"`
	ShowUniversityUndergraduate bool      `gorm:"default:false" json:"show_university_undergraduate"`
	GraduateDate                time.Time `gorm:"type:date" json:"graduate_date"`
	ShowGraduateDate            bool      `gorm:"default:false" json:"show_graduate_date"`
	MentionUndergraduate        string    `gorm:"size:255" json:"mention_undergraduate"`
	ShowMentionUndergraduate    bool      `gorm:"default:false" json:"show_mention_undergraduate"`

	// Register Title: Datos de registro legal del título en el estado.
	RegisterTitleState string    `gorm:"size:255" json:"register_title_state"`
	RegisterTitleDate  time.Time `gorm:"type:date" json:"register_title_date"`
	RegisterNumber     int       `json:"register_number"`
	RegisterFolio      string    `gorm:"size:255" json:"register_folio"`
	RegisterTome       string    `gorm:"size:255" json:"register_tome"`

	// Professional Flags: Roles y estatus especiales dentro del gremio.
	GuildDirector       bool `gorm:"default:false" json:"guild_director"`
	SixtyFiveOrPlus     bool `gorm:"default:false" json:"sixty_five_or_plus"`
	GuildCollaborator   bool `gorm:"default:false" json:"guild_collaborator"`
	PublicEmployee      bool `gorm:"default:false" json:"public_employee"`
	UniversityProfessor bool `gorm:"default:false" json:"university_professor"`

	// Histórico de solvencia y membresías dobles.
	DateOfLastSolvency time.Time `gorm:"type:date" json:"date_of_last_solvency"`
	DoubleGuild        bool      `gorm:"default:false" json:"double_guild"`
	CPSM               bool      `gorm:"default:false" json:"cpsm"`
}

func (PsiUserColData) TableName() string {
	return "psi_user_col_data"
}

// --- POSTGRADOS ---

// PsiUserPostGrade representa los títulos académicos adicionales (Especializaciones,
// Maestrías, Doctorados) obtenidos por el profesional.
type PsiUserPostGrade struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuditModel
	PsiUserID uuid.UUID `gorm:"type:uuid;index" json:"psi_user_id"`

	Title          string `gorm:"size:255;not null" json:"post_grade_title"`
	University     string `gorm:"size:255;not null" json:"post_grade_university"`
	GraduationYear string `gorm:"size:50" json:"post_grade_graduation_year"`
	Description    string `gorm:"type:text" json:"post_grade_description"`
	Active         bool   `gorm:"default:true" json:"is_active"`

	// S3 Keys para certificados: Almacenan los comprobantes académicos en S3.
	PicOneS3Key   string `gorm:"size:512" json:"pic_one_url"`
	PicTwoS3Key   string `gorm:"size:512" json:"pic_two_url"`
	PicThreeS3Key string `gorm:"size:512" json:"pic_three_url"`
}

func (PsiUserPostGrade) TableName() string {
	return "psi_user_post_grades"
}

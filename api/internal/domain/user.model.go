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
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel
	Credentials

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

	// ── Permisos: Proyectos (Kanban) ──────────────────────────────────────
	// "Master": accede y administra CUALQUIER proyecto del panel, sin importar
	// quién sea el dueño. Los admins sin este flag solo ven sus proyectos y
	// los que les compartieron como viewer/editor.
	CanManageProjects bool `gorm:"default:false" json:"can_manage_projects"`
}

func (UserAdmin) TableName() string { return "user_admins" }

// =============================================================================
// PSICÓLOGOS — PERFIL CORE
// =============================================================================

// PsiUserModel es la entidad principal de un Psicólogo colegiado.
// Agrupa identidad legal, contacto, ubicación geográfica, estado gremial
// y relaciones con los demás módulos del dominio.
type PsiUserModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel
	Credentials

	// ── Campos exclusivos de Psi ───────────────────────────────────────────
	AudioBookShellId string `gorm:"size:50;unique;" json:"audio_book_shell_id"` // Id del servicio

	// ── Identidad legal ───────────────────────────────────────────────────
	FirstName      string    `gorm:"size:255;not null" json:"first_name"`
	SecondName     string    `gorm:"size:255" json:"second_name"`
	LastName       string    `gorm:"size:255;not null" json:"last_name"`
	SecondLastName string    `gorm:"size:255" json:"second_last_name"`
	FPV            int       `gorm:"not null;uniqueIndex" json:"fpv"`                                                                                          // Número de Federación de Psicólogos de Venezuela
	CI             int       `gorm:"not null;uniqueIndex" json:"ci"`                                                                                           // Cédula de Identidad
	Nationality    string    `gorm:"size:1;not null" json:"nationality"`                                                                                       // V = venezolano, E = extranjero
	ControlNumber  string    `gorm:"size:50;uniqueIndex:idx_psi_users_control_number,where:control_number <> '' AND deleted_at IS NULL" json:"control_number"` // Nº de control interno (columna 'Nº' del Excel, visible solo admin)
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

	// ── Modalidad de servicio ─────────────────────────────────────────────
	// Cómo atiende el psicólogo (puede ser combinación de varias). Si ninguna
	// está activa, significa que actualmente no presta servicio.
	ServiceModalityPresencial bool `gorm:"default:false" json:"service_modality_presencial"`
	ServiceModalityDistance   bool `gorm:"default:false" json:"service_modality_distance"`
	ServiceModalityTelephone  bool `gorm:"default:false" json:"service_modality_telephone"`
	// Opt-in del Privacy Shield: controla si la modalidad se muestra en el directorio público.
	ShowServiceModality bool `gorm:"default:false" json:"show_service_modality"`

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	// Para miembros residentes o con consulta dentro del estado Carabobo.
	// MunicipalityCarabobo debe restringirse al catálogo de municipios del estado.
	MunicipalityCarabobo     string `gorm:"size:255" json:"municipality_carabobo"`
	ShowMunicipalityCarabobo bool   `gorm:"default:false" json:"show_municipality_carabobo"`
	PhoneCarabobo            string `gorm:"size:20;default:''" json:"phone_carabobo"`
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

	// ── Áreas de desempeño (especialidades) ───────────────────────────────
	// PrimaryWorkArea/SecondaryWorkArea: strings legacy mantenidos por backwards
	// compatibility con el frontend. Se mantienen sincronizados con las FK.
	PrimaryWorkArea   string `gorm:"size:50" json:"primary_work_area"`
	SecondaryWorkArea string `gorm:"size:50" json:"secondary_work_area"`

	// FKs al catálogo de especialidades. Nullable porque se setean manualmente
	// por el usuario después de crear su perfil, no durante la importación.
	PrimarySpecialtyID   *uint32 `gorm:"column:primary_specialty_id" json:"primary_specialty_id,omitempty"`
	SecondarySpecialtyID *uint32 `gorm:"column:secondary_specialty_id" json:"secondary_specialty_id,omitempty"`

	// ── Biografía profesional ─────────────────────────────────────────────
	MiniBio   string    `json:"mini_bio"`              // Resumen corto (max 250 chars) para el directorio
	BioTextID uuid.UUID `json:"bio_text_id,omitempty"` // FK hacia TextModel (contenido HTML sanitizado)
	FullBio   TextModel `gorm:"foreignKey:BioTextID" json:"full_bio,omitempty"`

	// ── Relaciones ────────────────────────────────────────────────────────
	ColData        PsiUserColData         `gorm:"foreignKey:PsiUserModelID" json:"col_data"`
	PostGrades     []PsiUserPostGrade     `gorm:"foreignKey:PsiUserID" json:"post_grades"`
	SocialNetworks []PsiUserSocialNetwork `gorm:"foreignKey:PsiUserID" json:"social_networks"`
	Solvencies     []PsiUserSolvency      `gorm:"foreignKey:PsiUserModelID" json:"solvencies"`
	Documents      []PsiUserDocument      `gorm:"foreignKey:PsiUserID" json:"-"`
	Observations   []PsiObservations      `gorm:"foreignKey:PsiUserID" json:"-"`
	Deontologia    []PsiODeontologia      `gorm:"foreignKey:PsiUserID" json:"-"`
}

func (PsiUserModel) TableName() string { return "psi_users" }

// =============================================================================
// DATOS COLEGIALES
// =============================================================================

// PsiUserColData almacena el historial académico y los datos regulatorios
// del Colegio. Es una relación 1-a-1 con PsiUserModel (uniqueIndex en PsiUserModelID).
type PsiUserColData struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
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
	Discapacity         bool `gorm:"default:false" json:"discapacity"`          // Discapacidad
	UniversityProfessor bool `gorm:"default:false" json:"university_professor"` // Docente universitario

	// El psicólogo autoriza que el sistema avise a la administración en su cumpleaños.
	// Solo se usará si el miembro lo activa desde su portal de perfil (opt-in explícito).
	BirthdayNotification bool `gorm:"default:false" json:"birthday_notification"`

	// ministerio_confirmed: la administración confirma que el psicólogo está inscrito
	// en el Ministerio de Educación (requisito legal, Art. 5 de la Ley de Ejercicio).
	MinistryRegistrationConfirmed bool `gorm:"default:false" json:"ministry_registration_confirmed"`

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

// PsiUserSolvency es un registro de las solvencia que posee el psicologo.
// Relación N-a-1 con PsiUserModel.
type PsiUserSolvency struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel
	// El nombre del índice debe ser el mismo en ambos campos para crear una clave compuesta
	PsiUserModelID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_psi_solvency_unique" json:"psi_user_model_id"`
	Date           time.Time `gorm:"type:date;not null;uniqueIndex:idx_psi_solvency_unique" json:"date"`
}

func (PsiUserSolvency) TableName() string { return "psi_user_solvency" }

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
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel
	PsiUserID uuid.UUID     `gorm:"type:uuid;index" json:"psi_user_id"`
	Type      PostGradeType `gorm:"type:varchar(50);not null" json:"post_grade_type"`

	Title          string `gorm:"size:255;not null" json:"post_grade_title"`
	University     string `gorm:"size:255" json:"post_grade_university"`
	GraduationYear int    `json:"post_grade_graduation_year"`
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
// REGISTRO DIGITAL DE DOCUMENTOS
// =============================================================================

// DocumentType define la categoría de un documento digital del expediente.
type DocumentType string

// Constantes con las categorías soportadas para agrupar y clasificar documentos.
const (
	DocumentCedula      DocumentType = "cedula"
	DocumentTitulo      DocumentType = "titulo"
	DocumentRif         DocumentType = "rif"
	DocumentSolvencia   DocumentType = "solvencia"
	DocumentComprobante DocumentType = "comprobante"
	DocumentOtro        DocumentType = "otro"
)

// IsValid valida que el tipo de documento sea uno de los soportados.
func (d DocumentType) IsValid() bool {
	switch d {
	case DocumentCedula, DocumentTitulo, DocumentRif, DocumentSolvencia, DocumentComprobante, DocumentOtro:
		return true
	}
	return false
}

// PsiUserDocument registra un documento digital del expediente del psicólogo
// (CI, título, RIF, comprobantes de solvencia, etc.).
//
// IMPORTANTE: La gestión (carga, edición, eliminación) es EXCLUSIVA del personal
// administrativo autorizado. El psicólogo solo puede CONSULTAR sus propios
// documentos (nunca editarlos ni borrarlos).
// Relación N-a-1 con PsiUserModel.
type PsiUserDocument struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel
	PsiUserID uuid.UUID `gorm:"type:uuid;index" json:"psi_user_id"`

	// Categoría del documento (cedula, titulo, rif, solvencia, comprobante, otro).
	DocumentType DocumentType `gorm:"type:varchar(50);default:otro;not null" json:"document_type"`
	// Etiqueta libre descrita por la administración: "Cédula V-123456 anverso",
	// "Comprobante de solvencia 2025", "Título de pregrado", etc.
	Title string `gorm:"size:255;not null" json:"title"`
	// Notas u observaciones internas sobre el documento.
	Notes string `gorm:"type:text" json:"notes"`
	// Fecha opcional a la que corresponde el documento (útil para comprobantes por año).
	DocumentDate *time.Time `gorm:"type:date" json:"document_date"`

	// S3 Key del archivo almacenado. Se serializa como `document_url` una vez
	// resuelta por el service (mismo patrón que los postgrados).
	S3Key string `gorm:"size:512;not null" json:"document_url"`
	// Nombre original del archivo subido.
	Filename string `gorm:"size:255" json:"filename"`
	// Tipo MIME del archivo almacenado (image/webp o application/pdf).
	MimeType string `gorm:"size:100" json:"mime_type"`
	// Tamaño en bytes del archivo almacenado.
	SizeBytes int64 `json:"size_bytes"`
}

func (PsiUserDocument) TableName() string { return "psi_user_documents" }

// =============================================================================
// OBSERVACIONES INTERNAS
// =============================================================================

// PsiObservations almacena notas internas del Colegio sobre un psicólogo.
// IMPORTANTE: Los psicólogos NUNCA pueden ver ni acceder a sus propias observaciones.
// Solo el personal administrativo autorizado puede crearlas, editarlas y leerlas.
// Relación N-a-1 con PsiUserModel.
type PsiObservations struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
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
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel
	PsiUserID uuid.UUID `gorm:"type:uuid;index" json:"psi_user_id"`
	Content   string    `gorm:"type:text" json:"content"`
}

func (PsiODeontologia) TableName() string { return "psi_deontologia" }

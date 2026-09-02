// api/internal/request_structs/psi_user.go

// Package request_structs contiene las estructuras de datos (DTOs) para la entrada y salida de la API.
// Estas estructuras actúan como contratos entre el frontend y el backend, asegurando la integridad
// y la privacidad de la información según el rol del usuario.
package request_structs

import (
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// =========================================================================
// AUTENTICACIÓN
// =========================================================================

// PsiLoginRequest representa las credenciales necesarias para el inicio de sesión.
type PsiLoginRequest struct {
	// Identifier puede ser el correo electrónico o el nombre de usuario.
	Identifier string `json:"identifier" validate:"required" example:"psicologo@email.com"`
	Password   string `json:"password" validate:"required" example:"12345678"`
}

// =========================================================================
// AUTOGESTIÓN DEL PSICÓLOGO (SELF-MANAGEMENT)
// =========================================================================

// PsiUserUpdateRequestSelf define el contrato para que un psicólogo edite su propio perfil.
// Aplica el principio de "Menor Privilegio": omite campos sensibles (CI, FPV, Solvencia)
// para que el usuario no pueda alterarlos mediante inyección de JSON.
// Nota: Se usan punteros (*tipo) para soportar semántica PATCH (actualizar solo lo enviado).
// api/internal/delivery/http/request_structs/psi_update_self.go
//
// FIX: Los campos de privacidad (*bool) no se parsean correctamente desde
// multipart/form-data con el BodyParser de Fiber cuando llegan como "true"/"false".
//
// Solución: Cambiar los *bool a *string en el struct de request y hacer la
// conversión explícita en el handler o en un método helper.
// Esto es lo más robusto para multipart forms.
//
// BoolFromForm convierte los valores de formulario "1"/"0"/"true"/"false" a *bool.
// Retorna nil si el valor está vacío (campo no enviado → semántica PATCH).
// PsiUserUpdateRequestSelf define el contrato para que un psicólogo edite su propio perfil.
// Los campos opcionales usan *string para distinguir "no enviado" de "enviado vacío".
// Los booleanos de visibilidad usan string Raw + getter porque multipart/form-data
// no puede representar *bool directamente.
// Los campos opcionales usan *string para distinguir "no enviado" de "enviado vacío".
// Los booleanos de visibilidad usan string Raw + getter porque multipart/form-data
// no puede representar *bool directamente.
type PsiUserUpdateRequestSelf struct {

	// ── Credenciales ──────────────────────────────────────────────────────
	Username     *string `json:"username,omitempty" form:"username"`
	Email        *string `json:"email,omitempty" form:"email"`
	Password     string  `json:"password" form:"password" validate:"required"` // contraseña actual (siempre requerida)
	NewPassword1 *string `json:"new_password_1,omitempty" form:"new_password_1"`
	NewPassword2 *string `json:"new_password_2,omitempty" form:"new_password_2"`

	// ── Contacto Gremial y Privacidad ─────────────────────────────────────
	ContactEmail     *string `json:"contact_email" form:"contact_email"`
	ContactPhone     *string `json:"contact_phone" form:"contact_phone"`           // Reemplaza a public_phone
	ContactCellPhone *string `json:"contact_cell_phone" form:"contact_cell_phone"` // Nuevo
	ServiceAddress   *string `json:"service_address" form:"service_address"`

	// Raw bools: "0"/"1"/"true"/"false" — usar getters Show...()
	ShowContactEmailRaw         string `form:"show_contact_email"`
	ShowPublicServiceAddressRaw string `form:"show_public_service_address"`

	// ── Ubicación: Carabobo ───────────────────────────────────────────────
	MunicipalityCarabobo *string `json:"municipality_carabobo" form:"municipality_carabobo"`
	PhoneCarabobo        *string `json:"phone_carabobo" form:"phone_carabobo"`
	CelPhoneCarabobo     *string `json:"cel_phone_carabobo" form:"cel_phone_carabobo"`

	ShowMunicipalityCaraboboRaw string `form:"show_municipality_carabobo"`
	ShowPhoneCaraboboRaw        string `form:"show_phone_carabobo"`
	ShowCelPhoneCaraboboRaw     string `form:"show_cel_phone_carabobo"`

	// ── Ubicación: Fuera de Carabobo (Venezuela) ─────────────────────────
	StateOutside                  *string `json:"state_outside" form:"state_outside"`
	MunicipalityOutSideCarabobo   *string `json:"municipality_outside_carabobo" form:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo          *string `json:"phone_outside_carabobo" form:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo       *string `json:"cel_phone_outside_carabobo" form:"cel_phone_outside_carabobo"`
	ServiceAddressOutSideCarabobo *string `json:"service_address_outside_carabobo" form:"service_address_outside_carabobo"`

	ShowStateOutsideRaw                        string `form:"show_state_outside"`
	ShowMunicipalityOutSideCaraboboRaw         string `form:"show_municipality_outside_carabobo"`
	ShowPhoneOutSideCaraboboRaw                string `form:"show_phone_outside_carabobo"`
	ShowCellPhoneOutSideCaraboboRaw            string `form:"show_cel_phone_outside_carabobo"`
	ShowPublicServiceAddressOutSideCaraboboRaw string `form:"show_public_service_address_outside_carabobo"`

	// ── Ubicación: Fuera de Venezuela ─────────────────────────────────────
	Country                        *string `json:"country" form:"country"`
	PhoneOutSideVenezuela          *string `json:"phone_outside_venezuela" form:"phone_outside_venezuela"`
	CellPhoneOutSideVenezuela      *string `json:"cell_phone_outside_venezuela" form:"cell_phone_outside_venezuela"` // Añadido
	ServiceAddressOutSideVenezuela *string `json:"service_address_outside_venezuela" form:"service_address_outside_venezuela"`

	ShowPhoneOutSideVenezuelaRaw                string `form:"show_phone_outside_venezuela"`
	ShowCellPhoneOutSideVenezuelaRaw            string `form:"show_cell_phone_outside_venezuela"` // Añadido
	ShowPublicServiceAddressOutSideVenezuelaRaw string `form:"show_public_service_address_outside_venezuela"`

	// ── Perfil Profesional y Biografía ───────────────────────────────────
	PrimaryWorkArea      *string `json:"primary_work_area" form:"primary_work_area"`
	SecondaryWorkArea    *string `json:"secondary_work_area" form:"secondary_work_area"`
	PrimarySpecialtyID   *uint32 `json:"primary_specialty_id" form:"primary_specialty_id"`
	SecondarySpecialtyID *uint32 `json:"secondary_specialty_id" form:"secondary_specialty_id"`
	MiniBio              *string `json:"mini_bio" form:"mini_bio"`
	FullBio              *string `json:"full_bio" form:"full_bio"`

	// ── Modalidad de servicio (auto-gestión) ─────────────────────────────
	// "0"/"1"/"true"/"false" — usar getters ServiceModality...()
	ServiceModalityPresencialRaw string `form:"service_modality_presencial"`
	ServiceModalityDistanceRaw   string `form:"service_modality_distance"`
	ServiceModalityTelephoneRaw  string `form:"service_modality_telephone"`
	ShowServiceModalityRaw       string `form:"show_service_modality"`

	// ── Visibilidad de Datos Colegiales ──────────────────────────────────
	ShowUniversityUndergraduateRaw string `form:"show_university_undergraduate"`
	ShowGraduateDateRaw            string `form:"show_graduate_date"`
	ShowMentionUndergraduateRaw    string `form:"show_mention_undergraduate"`

	// ── Preferencias ─────────────────────────────────────────────────────
	// El psicólogo autoriza que se avise a la administración en su cumpleaños.
	BirthdayNotificationRaw string `form:"birthday_notification"`
}

// ── Getters de booleanos de privacidad ────────────────────────────────────────

// ── Contacto Público ──
func (r *PsiUserUpdateRequestSelf) ShowContactEmail() *bool {
	return utils.BoolFromForm(r.ShowContactEmailRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowPublicServiceAddress() *bool {
	return utils.BoolFromForm(r.ShowPublicServiceAddressRaw)
}

// ── Carabobo ──
func (r *PsiUserUpdateRequestSelf) ShowMunicipalityCarabobo() *bool {
	return utils.BoolFromForm(r.ShowMunicipalityCaraboboRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowPhoneCarabobo() *bool {
	return utils.BoolFromForm(r.ShowPhoneCaraboboRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowCelPhoneCarabobo() *bool {
	return utils.BoolFromForm(r.ShowCelPhoneCaraboboRaw)
}

// ── Fuera de Carabobo (Venezuela) ──
func (r *PsiUserUpdateRequestSelf) ShowStateOutside() *bool {
	return utils.BoolFromForm(r.ShowStateOutsideRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowMunicipalityOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowMunicipalityOutSideCaraboboRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowPhoneOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowPhoneOutSideCaraboboRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowCellPhoneOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowCellPhoneOutSideCaraboboRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowPublicServiceAddressOutSideCarabobo() *bool {
	return utils.BoolFromForm(r.ShowPublicServiceAddressOutSideCaraboboRaw)
}

// ── Fuera de Venezuela ──
func (r *PsiUserUpdateRequestSelf) ShowPhoneOutSideVenezuela() *bool {
	return utils.BoolFromForm(r.ShowPhoneOutSideVenezuelaRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowCellPhoneOutSideVenezuela() *bool {
	return utils.BoolFromForm(r.ShowCellPhoneOutSideVenezuelaRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowPublicServiceAddressOutSideVenezuela() *bool {
	return utils.BoolFromForm(r.ShowPublicServiceAddressOutSideVenezuelaRaw)
}

// ── Datos Colegiales ──
func (r *PsiUserUpdateRequestSelf) ShowUniversityUndergraduate() *bool {
	return utils.BoolFromForm(r.ShowUniversityUndergraduateRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowGraduateDate() *bool {
	return utils.BoolFromForm(r.ShowGraduateDateRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowMentionUndergraduate() *bool {
	return utils.BoolFromForm(r.ShowMentionUndergraduateRaw)
}

// ── Preferencias ──
func (r *PsiUserUpdateRequestSelf) BirthdayNotification() *bool {
	return utils.BoolFromForm(r.BirthdayNotificationRaw)
}

// ── Modalidad de servicio ──
func (r *PsiUserUpdateRequestSelf) ServiceModalityPresencial() *bool {
	return utils.BoolFromForm(r.ServiceModalityPresencialRaw)
}
func (r *PsiUserUpdateRequestSelf) ServiceModalityDistance() *bool {
	return utils.BoolFromForm(r.ServiceModalityDistanceRaw)
}
func (r *PsiUserUpdateRequestSelf) ServiceModalityTelephone() *bool {
	return utils.BoolFromForm(r.ServiceModalityTelephoneRaw)
}
func (r *PsiUserUpdateRequestSelf) ShowServiceModality() *bool {
	return utils.BoolFromForm(r.ShowServiceModalityRaw)
}

// =========================================================================
// DIRECTORIO PÚBLICO (BÚSQUEDA Y LECTURA)
// =========================================================================

// PsiDirectoryFilterDTO captura y tipa los parámetros de búsqueda desde la URL.
type PsiDirectoryFilterDTO struct {
	SearchTerm  string `query:"q"`         // Texto para buscar por Nombre, Apellido, CI o FPV
	SpecialtyID uint32 `query:"specialty"` // Filtrado por ID del catálogo de especialidades
	Location    string `query:"location"`  // Filtro por Municipio o Estado
	Gender      string `query:"gender"`    // Filtro por género (M / F)
	Solvent     *bool  `query:"solvent"`   // nil = todos; true = solventes; false = insolventes
	Active      *bool  `query:"active"`    // nil = todos; true = activos; false = inactivos
	Page        int    `query:"page"`      // Número de página para paginación
	Limit       int    `query:"limit"`     // Cantidad de registros por página
}

// PsiMiniProfileDTO es una respuesta optimizada para el listado masivo (Grid/Cards).
// Reduce el ancho de banda y protege datos de contacto sensibles en vistas generales.
type PsiMiniProfileDTO struct {
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	CI             int       `json:"ci"`
	FPV            int       `json:"fpv"`
	ProfilePicture string    `json:"profile_picture"` // Contiene el S3 Key o la URL de la imagen
	MiniBio        string    `json:"mini_bio"`
	Specialties    []string  `json:"specialties"` // Slice de especialidades principales
	// Modalidad de servicio (solo si el psicólogo autorizó su visibilidad).
	ServiceModality *ServiceModalityDTO `json:"service_modality,omitempty"`
}

type UndergraduateDTO struct {
	University         string `json:"university,omitempty"`
	Date               string `json:"date,omitempty"`
	Mention            string `json:"mention,omitempty"`
	TitleImageOneURL   string `json:"title_image_one_url,omitempty"`   // URL o S3 Key del título de pregrado
	TitleImageTwoURL   string `json:"title_image_two_url,omitempty"`   // URL o S3 Key de un segundo documento (opcional)
	TitleImageThreeURL string `json:"title_image_three_url,omitempty"` // URL o S3 Key de un tercer documento (opcional)
}

// PsiFullProfileDTO representa la ficha pública detallada de un psicólogo.
// Implementa el "Privacy Shield": los campos con omitempty solo se renderizan
// si el usuario autorizó su visibilidad o si contienen datos.
type PsiFullProfileDTO struct {
	// ── Identidad Pública (siempre visible) ──────────────────────────────
	FirstName      string `json:"first_name"`
	SecondName     string `json:"second_name,omitempty"`
	LastName       string `json:"last_name"`
	SecondLastName string `json:"second_last_name,omitempty"`
	FPV            int    `json:"fpv"`
	CI             int    `json:"ci"`
	Gender         string `json:"gender"`
	ProfilePicture string `json:"profile_picture"`
	Solvent        bool   `json:"solvent"`

	// ── Contacto (visibilidad condicional) ───────────────────────────────
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`

	// ── Ubicación Estructurada ────────────────────────────────────────────
	// Un psicólogo puede tener presencia simultánea en Carabobo, en otro
	// estado venezolano y/o en el exterior. Cada bloque es independiente
	// y solo se incluye si tiene datos.
	Location PsiLocationDTO `json:"location"`

	// ── Modalidad de servicio (solo si el psicólogo autorizó su visibilidad) ──
	ServiceModality *ServiceModalityDTO `json:"service_modality,omitempty"`

	// ── Perfil Profesional ────────────────────────────────────────────────
	WorkAreas            []string `json:"work_areas"`
	MiniBio              string   `json:"mini_bio,omitempty"`
	FullBioContent       string   `json:"full_bio_content,omitempty"`
	PrimaryWorkArea      string   `json:"primary_work_area"`
	SecondaryWorkArea    string   `json:"secondary_work_area"`
	PrimarySpecialtyID   *uint32  `json:"primary_specialty_id,omitempty"`
	SecondarySpecialtyID *uint32  `json:"secondary_specialty_id,omitempty"`

	// ── Datos de Pregrado (condicional por privacidad) ───────────────────
	Undergraduate UndergraduateDTO `json:"undergraduate"`

	// ── Relaciones ────────────────────────────────────────────────────────
	PostGrades     []PostGradeDTO     `json:"post_grades,omitempty"`
	SocialNetworks []SocialNetworkDTO `json:"social_networks,omitempty"`
}

// ServiceModalityDTO describe cómo atiende un psicólogo (puede ser combinación).
// Si los tres son false, el profesional no presta servicio actualmente.
type ServiceModalityDTO struct {
	Presencial bool `json:"presencial"`
	Distance   bool `json:"distance"`
	Telephone  bool `json:"telephone"`
}

// PsiLocationDTO agrupa las tres zonas geográficas posibles de un psicólogo.
// Cada bloque es opcional — solo se serializa si tiene al menos un campo no vacío.
type PsiLocationDTO struct {
	// Presencia en el estado Carabobo (jurisdicción principal del Colegio)
	Carabobo *PsiLocationCaraboboDTO `json:"carabobo,omitempty"`

	// Presencia en otro estado venezolano (excluye Carabobo)
	Venezuela *PsiLocationVenezuelaDTO `json:"venezuela,omitempty"`

	// Presencia en el exterior
	Exterior *PsiLocationExteriorDTO `json:"exterior,omitempty"`
}

type PsiLocationCaraboboDTO struct {
	Municipality string `json:"municipality"`
	Phone        string `json:"phone,omitempty"`
	CellPhone    string `json:"cell_phone,omitempty"`
	Address      string `json:"address,omitempty"`
}

type PsiLocationVenezuelaDTO struct {
	State        string `json:"state"`
	Municipality string `json:"municipality,omitempty"`
	Phone        string `json:"phone,omitempty"`
	CellPhone    string `json:"cell_phone,omitempty"`
	Address      string `json:"address,omitempty"`
}

type PsiLocationExteriorDTO struct {
	Country   string `json:"country"`
	Phone     string `json:"phone,omitempty"`
	CellPhone string `json:"cell_phone,omitempty"`
	Address   string `json:"address,omitempty"`
}

// =========================================================================
// MÓDULO ACADÉMICO (POSTGRADOS)
// =========================================================================

// CreatePostGradeRequest estructura la carga útil para registrar nuevos postgrados.
// Usa etiquetas 'form' para facilitar la carga de archivos (ej. diploma en PDF/JPG).
type CreatePostGradeRequest struct {
	Title          string `form:"title" json:"title" validate:"required"`
	University     string `form:"university" json:"university" validate:"required"`
	GraduationYear int    `form:"graduation_year" json:"graduation_year" validate:"required"`
	Description    string `form:"description" json:"description"`
}

// UpdatePostGradeRequest permite la edición parcial de un registro de postgrado.
type UpdatePostGradeRequest struct {
	Title          *string `form:"title"`
	University     *string `form:"university"`
	GraduationYear *int    `form:"graduation_year"`
	Description    *string `form:"description"`
}

// PostGradeDTO es la proyección pública de un postgrado para la ficha del psicólogo.
type PostGradeDTO struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	University  string `json:"university"`
	Year        int    `json:"year"`
	Description string `json:"description,omitempty"`
	PicOneURL   string `json:"pic_one_url,omitempty"`   // URL o S3 Key del diploma o certificado
	PicTwoURL   string `json:"pic_two_url,omitempty"`   // URL o S3 Key de un segundo documento (opcional)
	PicThreeURL string `json:"pic_three_url,omitempty"` // URL o S3 Key de un tercer documento (opcional)

}

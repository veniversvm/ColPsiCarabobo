// api/internal/request_structs/psi_user.go

// Package request_structs contiene las estructuras de datos (DTOs) para la entrada y salida de la API.
// Estas estructuras actúan como contratos entre el frontend y el backend, asegurando la integridad
// y la privacidad de la información según el rol del usuario.
package request_structs

import "github.com/google/uuid"

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
type PsiUserUpdateRequestSelf struct {
	// --- Datos de Usuario Básicos ---
	Username     *string `json:"username,omitempty" form:"username"`
	Email        *string `json:"email,omitempty" form:"email"`
	NewPassword1 *string `json:"new_password_1,omitempty" form:"new_password_1"`
	NewPassword2 *string `json:"new_password_2,omitempty" form:"new_password_2"`
	Password     string  `json:"password,omitempty" form:"password" validate:"required"`

	// --- Datos de Contacto y Privacidad ---
	ContactEmail             *string `json:"contact_email" form:"contact_email"`
	ShowContactEmail         *bool   `json:"show_contact_email" form:"show_contact_email"`
	PublicPhone              *string `json:"public_phone" form:"public_phone"`
	ShowPublicPhone          *bool   `json:"show_public_phone" form:"show_public_phone"`
	ServiceAddress           *string `json:"service_address" form:"service_address"`
	ShowPublicServiceAddress *bool   `json:"show_public_service_address" form:"show_public_service_address"`

	// --- Ubicación Geográfica ---
	MunicipalityCarabobo        *string `json:"municipality_carabobo" form:"municipality_carabobo"`
	PhoneCarabobo               *string `json:"phone_carabobo" form:"phone_carabobo"`
	CelPhoneCarabobo            *string `json:"cel_phone_carabobo" form:"cel_phone_carabobo"`
	StateOutside                *string `json:"state_outside" form:"state_outside"`
	MunicipalityOutSideCarabobo *string `json:"municipality_outside_carabobo" form:"municipality_outside_carabobo"`
	PhoneOutSideCarabobo        *string `json:"phone_outside_carabobo" form:"phone_outside_carabobo"`
	CelPhoneOutSideCarabobo     *string `json:"cel_phone_outside_carabobo" form:"cel_phone_outside_carabobo"`

	// --- Perfil Profesional y Biografía ---
	PrimarySpecialty   *string `json:"primary_specialty" form:"primary_specialty"`
	SecondarySpecialty *string `json:"secondary_specialty" form:"secondary_specialty"`
	MiniBio            *string `json:"mini_bio" form:"mini_bio"`

	// --- Visibilidad de Datos Colegiales (Pregrado) ---
	ShowUniversityUndergraduate *bool `json:"show_university_undergraduate" form:"show_university_undergraduate"`
	ShowGraduateDate            *bool `json:"show_graduate_date" form:"show_graduate_date"`
	ShowMentionUndergraduate    *bool `json:"show_mention_undergraduate" form:"show_mention_undergraduate"`
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
// Implementa el "Privacy Shield": los campos marcados con 'omitempty' solo se renderizan
// en el JSON si el usuario ha autorizado su visibilidad o si contienen datos.
type PsiFullProfileDTO struct {
	// Identidad Pública (Siempre visible)
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	SecondName     string    `json:"second_name,omitempty"`
	LastName       string    `json:"last_name"`
	SecondLastName string    `json:"second_last_name,omitempty"`
	FPV            int       `json:"fpv"`
	CI             int       `json:"ci"`
	Gender         string    `json:"gender"`
	ProfilePicture string    `json:"profile_picture"`
	Solvent        bool      `json:"solvent"`

	// Contacto (Visibilidad Condicional)
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`

	// Ubicación Estructurada
	Location struct {
		State        string `json:"state"`
		Municipality string `json:"municipality"`
		FullAddress  string `json:"full_address,omitempty"`
	} `json:"location"`

	// Historial Académico y Profesional
	Specialties []string `json:"specialties"`
	MiniBio     string   `json:"mini_bio"`

	// Datos de Pregrado (Condicional)
	Undergraduate UndergraduateDTO `json:"undergraduate"`

	// Relaciones (Colecciones)
	PostGrades     []PostGradeDTO     `json:"post_grades,omitempty"`
	SocialNetworks []SocialNetworkDTO `json:"social_networks,omitempty"`
}

// =========================================================================
// MÓDULO ACADÉMICO (POSTGRADOS)
// =========================================================================

// CreatePostGradeRequest estructura la carga útil para registrar nuevos postgrados.
// Usa etiquetas 'form' para facilitar la carga de archivos (ej. diploma en PDF/JPG).
type CreatePostGradeRequest struct {
	Title          string `form:"title" json:"title" validate:"required"`
	University     string `form:"university" json:"university" validate:"required"`
	GraduationYear string `form:"graduation_year" json:"graduation_year" validate:"required"`
	Description    string `form:"description" json:"description"`
}

// UpdatePostGradeRequest permite la edición parcial de un registro de postgrado.
type UpdatePostGradeRequest struct {
	Title          *string `form:"title"`
	University     *string `form:"university"`
	GraduationYear *string `form:"graduation_year"`
	Description    *string `form:"description"`
}

// PostGradeDTO es la proyección pública de un postgrado para la ficha del psicólogo.
type PostGradeDTO struct {
	Title       string `json:"title"`
	University  string `json:"university"`
	Year        string `json:"year"`
	Description string `json:"description,omitempty"`
	PicOneURL   string `json:"pic_one_url,omitempty"`   // URL o S3 Key del diploma o certificado
	PicTwoURL   string `json:"pic_two_url,omitempty"`   // URL o S3 Key de un segundo documento (opcional)
	PicThreeURL string `json:"pic_three_url,omitempty"` // URL o S3 Key de un tercer documento (opcional)

}

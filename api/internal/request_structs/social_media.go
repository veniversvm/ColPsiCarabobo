// api/internal/request_structs/social_media.go

// Package request_structs contiene las definiciones de los objetos de transferencia de datos.
package request_structs

// CreateSocialNetworkRequest define la carga útil necesaria para vincular una nueva
// red social al perfil de un psicólogo.
type CreateSocialNetworkRequest struct {
	// Name es el identificador de la red (ej: "ig", "facebook", "linkedin").
	// El servicio se encargará de normalizar este valor.
	Name string `json:"name" validate:"required" example:"ig"`

	// URL es el enlace absoluto al perfil del profesional.
	URL string `json:"url" validate:"required" example:"https://instagram.com/psicologo"`
}

// UpdateSocialNetworkRequest facilita la modificación parcial de un registro de red social.
// Arquitectura Senior: Se utilizan punteros para todos los campos para permitir la semántica PATCH.
// Esto permite distinguir entre un campo no enviado (nil) y un campo enviado con valor vacío o falso.
type UpdateSocialNetworkRequest struct {
	// Name opcional para corregir o cambiar el nombre de la plataforma.
	Name *string `json:"name" example:"Instagram"`

	// URL opcional para actualizar el enlace del perfil.
	URL *string `json:"url" example:"https://instagram.com/nuevo_perfil"`

	// IsActive permite habilitar o deshabilitar la visibilidad de la red social
	// en el perfil público sin eliminar el registro.
	IsActive *bool `json:"is_active" example:"true"`
}

// SocialNetworkDTO es el objeto de proyección utilizado en las respuestas de la API (Read Model).
// Proporciona una vista simplificada y limpia de las redes sociales para ser consumida
// directamente por el Frontend en SolidJS.
type SocialNetworkDTO struct {
	// Name contiene el nombre de la plataforma ya normalizado (ej: "Instagram").
	Name string `json:"name" example:"Instagram"`

	// URL contiene el enlace validado.
	URL string `json:"url" example:"https://instagram.com/psicologo"`
}

// api/internal/request_structs/social_media.go

// Package request_structs contiene las definiciones de los objetos de transferencia de datos.
//
// Este archivo gestiona los contratos de entrada y salida para el submódulo de Presencia Digital
// (Redes Sociales). Garantiza que las URLs introducidas por el psicólogo sean
// transportadas y validadas correctamente antes de persistirse.
package request_structs

// CreateSocialNetworkRequest define la carga útil (Payload) necesaria para vincular
// una nueva plataforma o red social al perfil público de un psicólogo.
type CreateSocialNetworkRequest struct {
	// Name es el identificador o alias de la plataforma (ej: "ig", "facebook", "linkedin").
	// Arquitectura: Se espera que la capa de Casos de Uso (Dominio) reciba este valor
	// en crudo y ejecute una lógica de normalización (ej. convertir "ig" a "Instagram")
	// antes de insertarlo en la base de datos para mantener la consistencia visual.
	Name string `json:"name" validate:"required" example:"ig"`

	// URL es el enlace web absoluto al perfil profesional.
	URL string `json:"url" validate:"required" example:"https://instagram.com/psicologo"`
}

// UpdateSocialNetworkRequest facilita la modificación parcial de un registro existente.
//
// Semántica PATCH Real (Arquitectura Senior):
// Se utilizan punteros (*string, *bool) para todos los campos mutables.
// Esto permite al servidor distinguir con precisión criptográfica entre un campo que
// no fue enviado en el payload (puntero nil -> se ignora en la DB) y un campo que fue
// enviado intencionalmente con un valor vacío o en falso.
type UpdateSocialNetworkRequest struct {
	// Name permite corregir un error tipográfico o cambiar el alias de la plataforma.
	Name *string `json:"name" example:"Instagram"`

	// URL permite actualizar el enlace en caso de que el profesional cambie su handle/arroba.
	URL *string `json:"url" example:"https://instagram.com/nuevo_perfil"`

	// IsActive funciona como un interruptor de visibilidad (Soft-Disable).
	// Regla de Negocio: Permite al usuario ocultar temporalmente la red social de
	// su perfil público (ej. si está rediseñando su Instagram) sin necesidad de
	// ejecutar un DELETE físico que destruiría el registro permanentemente.
	IsActive *bool `json:"is_active" example:"true"`
}

// SocialNetworkDTO es el objeto de proyección (Read Model) utilizado en las respuestas de la API.
//
// Optimización de UI:
// Proporciona una vista simplificada, aplanada y pre-procesada de la red social.
// Al entregar los datos en este formato puro, el framework de Frontend (SolidJS)
// puede inyectarlos directamente en su ciclo de renderizado (iterando componentes de UI)
// sin tener que parsear estados, aplicar lógicas condicionales o limpiar datos.
type SocialNetworkDTO struct {
	// Name contiene el identificador de la plataforma ya normalizado y formateado,
	// listo para mapear componentes de íconos (ej: renderizar el ícono de "Instagram").
	Name string `json:"name" example:"Instagram"`

	// URL contiene el enlace validado y listo para ser inyectado en un atributo href="...".
	URL string `json:"url" example:"https://instagram.com/psicologo"`
}

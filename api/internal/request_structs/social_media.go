package request_structs

type CreateSocialNetworkRequest struct {
	Name string `json:"name" validate:"required" example:"ig"`
	URL  string `json:"url" validate:"required" example:"https://instagram.com/psicologo"`
}

// UpdateSocialNetworkRequest para modificaciones parciales
type UpdateSocialNetworkRequest struct {
	Name     *string `json:"name" example:"Instagram"`
	URL      *string `json:"url" example:"https://instagram.com/nuevo_perfil"`
	IsActive *bool   `json:"is_active"`
}

type SocialNetworkDTO struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

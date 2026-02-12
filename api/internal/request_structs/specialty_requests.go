package request_structs

type CreateSpecialtyRequest struct {
	Name        string `json:"name" example:"Psicología Clínica" validate:"required"`
	Description string `json:"description" example:"Especialidad enfocada en..."`
}

type UpdateSpecialtyRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
}

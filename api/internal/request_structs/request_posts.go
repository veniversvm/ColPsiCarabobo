package request_structs

type CreatePostRequest struct {
	Title            string `form:"title" validate:"required,max=100"`
	ShortDescription string `form:"short_description" validate:"max=250"`
	Content          string `form:"content" validate:"required"` // HTML o Markdown
	Type             string `form:"type" validate:"required,oneof=public psi"`
	IsActive         bool   `form:"is_active"`
}

// UpdatePostRequest es el DTO para operaciones PATCH de solo texto/estado.
type UpdatePostRequest struct {
	Title            *string `form:"title"`
	ShortDescription *string `form:"short_description"`
	Content          *string `form:"content"`
	Type             *string `form:"type"`
	IsActive         *bool   `form:"is_active"`
	// Nota: La imagen se maneja por separado con c.FormFile("image")
}

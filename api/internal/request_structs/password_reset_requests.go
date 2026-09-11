// api/internal/request_structs/password_reset_requests.go

// DTOs de recuperación de contraseña para psicólogos. Flujo seguro de dos
// pasos: 1) solicitar enlace (solo email) y 2) fijar nueva contraseña con
// el token de un solo uso recibido por correo.
package request_structs

// RequestPasswordResetDTO es la carga útil para solicitar el enlace de
// recuperación. Solo requiere el correo registrado; la respuesta es genérica
// para no revelar si una cuenta existe (anti-enumeración).
type RequestPasswordResetDTO struct {
	Email string `json:"email" validate:"required,email" example:"psicologo@email.com"`
}

// ResetPasswordDTO fija la nueva contraseña con el token de un solo uso.
// La confirmación garantiza que el usuario no cometa errores de tipeo al
// introducir su nueva clave.
type ResetPasswordDTO struct {
	Token           string `json:"token" validate:"required" example:"a1b2c3..."`
	NewPassword     string `json:"new_password" validate:"required,min=8" example:"NuevaClaveSegura123!"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8" example:"NuevaClaveSegura123!"`
}
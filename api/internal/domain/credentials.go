// api/internal/domain/credentials.go
package domain

// Credentials es un struct embebido que encapsula los campos comunes de autenticación
// compartidos por UserAdmin y PsiUserModel. Centraliza credenciales, estado de cuenta
// y mecanismo de invalidación de sesiones (key rotation).
type Credentials struct {
	// Username es el identificador único de login del usuario.
	Username string `gorm:"size:25;unique;not null" json:"username"`

	// Email es la dirección de correo electrónico única del usuario.
	Email string `gorm:"size:255;unique;not null" json:"email"`

	// Password almacena el hash bcrypt de la contraseña. Nunca se expone en JSON.
	Password string `gorm:"size:512;not null" json:"-"`

	// Key es un secreto rotativo. Al cambiarlo, todos los JWT emitidos previamente
	// quedan inválidos automáticamente (key rotation).
	Key string `gorm:"size:512;" json:"-"`

	// IsActive controla si la cuenta puede iniciar sesión.
	// Default false: un psicólogo creado desde una inscripción aprobada nace inactivo
	// y solo puede activarse manualmente cuando la administración confirma los
	// requisitos legales (N° de FPV y solvencia; el FPV acredita la inscripción ministerial).
	IsActive bool `gorm:"column:is_active;default:false" json:"is_active"`

	// MustChangePassword indica que el usuario debe cambiar su contraseña en el próximo login.
	// Se activa automáticamente cuando se crea una cuenta con contraseña temporal (imports masivos).
	MustChangePassword bool `gorm:"column:must_change_password;default:false" json:"-"`
}

package domain

import "errors"

// Errores de Autenticación
var (
	ErrPasswordIncorrect  = errors.New("contraseña actual incorrecta")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrAccountInactive    = errors.New("cuenta inactiva o suspendida")
)

// Errores de Autorización
var (
	ErrPermissionDenied  = errors.New("no tienes permiso para editar este registro")
	ErrInsufficientPerms = errors.New("permisos insuficientes")
)

// Errores de Recursos
var (
	ErrPsiNotFound = errors.New("psicólogo no encontrado")
)

// Errores de Negocio
var (
	ErrMaxSocialNetworks = errors.New("límite máximo de redes sociales alcanzado")
	ErrSocialPermDenied  = errors.New("no tienes permiso para editar esta red social")
	ErrSocialOwnDenied   = errors.New("no puedes borrar una red social que no te pertenece")
	ErrPostPermDenied    = errors.New("no tienes permiso para publicar")
	ErrUniqueViolation   = errors.New("registro duplicado")
	ErrSudoExists        = errors.New("ya existe un usuario SUDO")
)

package service

import (
	"errors"
	"strings"
)

// MapDBError traduce errores de GORM/Postgres a mensajes de usuario amigables.
func MapDBError(err error) error {
	msg := err.Error()

	if strings.Contains(msg, "idx_psi_users_ci") || strings.Contains(msg, "uni_psi_users_ci") {
		return errors.New("la Cédula de Identidad ya se encuentra registrada")
	}
	if strings.Contains(msg, "idx_psi_users_fpv") || strings.Contains(msg, "uni_psi_users_fpv") {
		return errors.New("el número de FPV ya está registrado por otro psicólogo")
	}
	if strings.Contains(msg, "uni_psi_users_email") {
		return errors.New("el correo electrónico ya está en uso")
	}
	if strings.Contains(msg, "uni_psi_users_username") {
		return errors.New("el nombre de usuario ya existe")
	}

	return err // Si no es un error conocido, retornamos el original
}

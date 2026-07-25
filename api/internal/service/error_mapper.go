// Package service implementa la lógica de negocio y utilidades transversales.
package service

import (
	"errors"
	"strings"
)

// MapDBError actúa como un traductor y escudo de seguridad para los errores de base de datos.
//
// Propósito Arquitectónico y de Seguridad:
//  1. Prevención de Fugas de Información (CWE-209): GORM y PostgreSQL suelen devolver
//     errores crudos que exponen la estructura interna de la base de datos (nombres de
//     tablas, columnas, sentencias SQL). Este método intercepta esos errores y los enmascara.
//  2. Experiencia de Usuario (UX): Traduce violaciones de restricciones únicas (Unique Constraints)
//     criptográficas en mensajes de retroalimentación claros, accionables y en español
//     para que el Frontend pueda mostrarlos directamente al usuario final.
func MapDBError(err error) error {
	msg := err.Error()

	// ── Colisiones en Unique Keys (prefijos idx_ y uni_, sufijos _key) ──────

	if strings.Contains(msg, "idx_psi_users_ci") || strings.Contains(msg, "uni_psi_users_ci") || strings.Contains(msg, "psi_users_ci_key") {
		return errors.New("la Cédula de Identidad ya se encuentra registrada")
	}

	if strings.Contains(msg, "idx_psi_users_fpv") || strings.Contains(msg, "uni_psi_users_fpv") || strings.Contains(msg, "psi_users_fpv_key") {
		return errors.New("el número de FPV ya está registrado por otro psicólogo")
	}

	if strings.Contains(msg, "idx_psi_users_email") || strings.Contains(msg, "uni_psi_users_email") || strings.Contains(msg, "psi_users_email_key") {
		return errors.New("el correo electrónico ya está en uso")
	}

	if strings.Contains(msg, "idx_psi_users_username") || strings.Contains(msg, "uni_psi_users_username") || strings.Contains(msg, "psi_users_username_key") {
		return errors.New("el nombre de usuario ya existe")
	}

	// ── Errores de Longitud ─────────────────────────────────────────────────

	if strings.Contains(msg, "value too long for type character varying(25)") {
		return errors.New("el nombre de usuario generado excede el límite de 25 caracteres")
	}
	if strings.Contains(msg, "value too long") {
		return errors.New("un campo es demasiado largo para la base de datos")
	}

	// ── Errores de formato ──────────────────────────────────────────────────

	if strings.Contains(msg, "invalid input syntax for type uuid") {
		return errors.New("error interno: ID de sistema inválido")
	}

	// Fallback: no exponer el error crudo al cliente (CWE-209)
	return err
}

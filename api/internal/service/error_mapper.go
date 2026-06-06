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
	// Extraemos el string crudo del error emitido por la capa de infraestructura
	msg := err.Error()

	// Identificación de colisiones en Restricciones Únicas (Unique Keys).
	// Se evalúan tanto los prefijos estándar de índices (idx_) como los de unicidad (uni_)
	// para cubrir distintas versiones de convenciones de nombrado generadas por GORM/AutoMigrate.

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

	// Si el error no coincide con ninguna restricción de negocio conocida,
	// lo dejamos pasar hacia arriba en la pila de llamadas (Bubbling up)
	// para que sea registrado por el logger centralizado del servidor HTTP.
	return err
}

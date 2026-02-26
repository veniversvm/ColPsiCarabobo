// api/internal/utils/no_empty_req.go

// Package utils provee herramientas transversales de soporte para la aplicación.
package utils

import "reflect"

// =========================================================================
// VALIDACIÓN DE ESTRUCTURAS
// =========================================================================

// IsEmptyReq utiliza reflexión (reflection) para determinar si un struct de petición
// está completamente vacío (todos sus campos tienen el valor por defecto de Go).
//
// Esta función es especialmente útil para:
//  1. Validar cuerpos de peticiones PATCH antes de iniciar transacciones.
//  2. Evitar errores de base de datos al intentar actualizar registros sin cambios.
//  3. Retornar respuestas rápidas (400 Bad Request) si el JSON enviado es {}.
//
// Soporta tanto structs directos como punteros a structs.
func IsEmptyReq(s interface{}) bool {
	v := reflect.ValueOf(s)

	// 1. Manejo de Punteros: Si el valor es un puntero, obtenemos el elemento subyacente
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// 2. Inspección de Campos: Iteramos por cada campo del struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		// Si encontramos aunque sea UN campo que no sea el valor por defecto (Zero Value),
		// significa que la petición contiene información válida para procesar.
		// Nota: En structs con punteros (*string, *bool), un campo no vacío es aquel != nil.
		if !field.IsZero() {
			return false
		}
	}

	// Si el ciclo termina sin encontrar valores, el struct está vacío.
	return true
}

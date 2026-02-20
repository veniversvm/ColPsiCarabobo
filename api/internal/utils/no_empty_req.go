package utils

import "reflect"

func IsEmptyReq(s interface{}) bool {
	v := reflect.ValueOf(s)

	// Si es un puntero, obtenemos el valor al que apunta
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// Si encontramos aunque sea UN campo que no sea el valor por defecto, no está vacío
		if !field.IsZero() {
			return false
		}
	}
	return true
}

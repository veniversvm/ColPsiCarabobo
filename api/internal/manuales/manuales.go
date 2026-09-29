// api/internal/manuales/manuales.go
//
// Paquete manuales: incrusta los PDFs de los manuales (admin y psicólogo)
// dentro del binario de la API con `//go:embed` (mismo patrón que
// internal/templates). Los archivos NO viven en un directorio estático ni en
// web/public: solo son accesibles vía el endpoint autenticado
// GET /api/v1/manuales/:file, así que nunca se sirven a visitantes anónimos.
package manuales

import _ "embed"

//go:embed files/manual-admin.pdf
var adminPDF []byte

//go:embed files/manual-psiuser.pdf
var psiPDF []byte

// available es la whitelist de manuales servibles. El lookup por nombre usa
// este mapa (y NUNCA el filesystem) → path-traversal imposible por diseño; un
// nombre fuera de la lista simplemente no existe (404 en el handler).
var available = map[string][]byte{
	"manual-admin.pdf":   adminPDF,
	"manual-psiuser.pdf": psiPDF,
}

// Get devuelve los bytes del manual solicitado y true, o nil y false si el
// nombre no está en la whitelist.
func Get(name string) ([]byte, bool) {
	data, ok := available[name]
	return data, ok
}
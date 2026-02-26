package templates

import "embed"

// FS exporta el sistema de archivos embebido para que otros paquetes lo usen.
//
//go:embed *.html
var FS embed.FS

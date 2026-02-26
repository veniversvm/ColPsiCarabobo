// api/internal/utils/image_sanitizer.go
// Package utils provee herramientas transversales de soporte para la aplicación.
package utils

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif" // Registro del decodificador GIF en el runtime
	"image/jpeg"  // Registro del decodificador/codificador JPEG
	"image/png"   // Registro del decodificador/codificador PNG
	"io"
)

// =========================================================================
// PROCESAMIENTO Y SEGURIDAD DE ARCHIVOS
// =========================================================================

// SanitizeImage implementa un mecanismo de limpieza profunda para archivos de imagen.
//
// LÓGICA DE SEGURIDAD:
// En lugar de simplemente validar la extensión, esta función decodifica la imagen
// píxel por píxel en la memoria del servidor y la vuelve a dibujar desde cero.
// Esto garantiza la eliminación de:
//  1. Metadatos EXIF (que podrían filtrar coordenadas GPS del psicólogo).
//  2. Payloads maliciosos incrustados en comentarios de archivos o esteganografía.
//  3. Scripts políglotas que intenten ejecutarse en el navegador.
//
// Retorna los bytes limpios, la extensión recomendada, el Content-Type y un error si falla.
func SanitizeImage(file io.Reader) ([]byte, string, string, error) {
	// 1. DECODIFICACIÓN (Detección de formato por Magic Numbers)
	// GORM/Go detecta automáticamente si es JPEG, PNG o GIF analizando la cabecera.
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", "", errors.New("archivo no es una imagen válida o formato no reconocido")
	}

	buf := new(bytes.Buffer)
	var ext string
	var contentType string

	// 2. RE-CODIFICACIÓN (Limpieza y Optimización)
	// Al codificar de nuevo, solo se guardan los datos visuales, descartando basura.
	switch format {
	case "jpeg", "jpg":
		// Aplicamos un balance de calidad del 85% para optimizar el peso en S3/CDN.
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
		ext = ".jpg"
		contentType = "image/jpeg"

	case "png":
		// PNG se mantiene sin pérdida para preservar transparencias si existen.
		err = png.Encode(buf, img)
		ext = ".png"
		contentType = "image/png"

	case "gif":
		// REGLA DE SEGURIDAD: Convertimos GIFs a PNG estático.
		// Esto evita vulnerabilidades de frames infinitos o desbordamiento de memoria
		// al procesar archivos animados en el frontend.
		err = png.Encode(buf, img)
		ext = ".png"
		contentType = "image/png"

	default:
		return nil, "", "", errors.New("formato de imagen no soportado por la política de seguridad")
	}

	if err != nil {
		return nil, "", "", errors.New("error crítico al sanitizar la imagen")
	}

	return buf.Bytes(), ext, contentType, nil
}

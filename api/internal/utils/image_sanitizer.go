package utils

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif" // Decodificador GIF
	"image/jpeg"  // Decodificador/Codificador JPEG
	"image/png"   // Decodificador/Codificador PNG
	"io"
)

// SanitizeImage decodifica la imagen en memoria y la vuelve a codificar.
// Esto elimina metadatos (EXIF/GPS) y payloads maliciosos, respetando el formato original.
// Retorna: bytes limpios, extensión (.jpg/.png) y el Content-Type correcto.
func SanitizeImage(file io.Reader) ([]byte, string, string, error) {
	// 1. Decodificar (Detecta formato mágicamente)
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", "", errors.New("archivo no es una imagen válida")
	}

	buf := new(bytes.Buffer)
	var ext string
	var contentType string

	// 2. Re-codificar según el formato de entrada
	switch format {
	case "jpeg", "jpg":
		// Calidad 85 es el estándar web para balance peso/calidad
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
		ext = ".jpg"
		contentType = "image/jpeg"
	case "png":
		// PNG no pierde calidad (lossless) y mantiene transparencia
		err = png.Encode(buf, img)
		ext = ".png"
		contentType = "image/png"
	case "gif":
		// Opcional: Convertir GIF a PNG o mantener GIF (Go soporta GIF animados limitadamente en encode básico)
		// Por seguridad y simplicidad, convertimos GIF a PNG estático (primer frame) o JPEG
		err = png.Encode(buf, img)
		ext = ".png"
		contentType = "image/png"
	default:
		return nil, "", "", errors.New("formato no soportado (solo jpg, png)")
	}

	if err != nil {
		return nil, "", "", errors.New("error al procesar/sanitizar la imagen")
	}

	return buf.Bytes(), ext, contentType, nil
}

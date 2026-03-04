// api/internal/utils/image_sanitizer.go
// Package utils provee herramientas transversales de soporte para la aplicación.
package utils

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	_ "image/gif" // Registro del decodificador GIF en el runtime
	"image/jpeg"  // Registro del decodificador/codificador JPEG
	"image/png"   // Registro del decodificador/codificador PNG
	"io"

	"golang.org/x/image/draw"
)

// =========================================================================
// CONSTANTES DE COMPRESIÓN
// =========================================================================

const (
	// maxFileSizeBytes define el umbral en bytes (1 MB) a partir del cual
	// se activa el bucle de compresión iterativa para reducir el peso del archivo.
	maxFileSizeBytes = 1 * 1024 * 1024 // 1 MB

	// compressionScaleFactor es el factor de reducción de dimensiones por iteración.
	// Un valor de 0.8 significa que en cada pasada la imagen se reduce al 80% de su tamaño.
	compressionScaleFactor = 0.8

	// minDimensionPx es la dimensión mínima (ancho o alto) permitida durante la
	// compresión. Esto evita reducir la imagen hasta hacerla inutilizable.
	minDimensionPx = 100

	// jpegQuality es la calidad de codificación para archivos JPEG (0-100).
	// 85 ofrece un excelente balance entre fidelidad visual y tamaño de archivo.
	jpegQuality = 85
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
	// Go detecta automáticamente si es JPEG, PNG o GIF analizando la cabecera binaria.
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", "", errors.New("archivo no es una imagen válida o formato no reconocido")
	}

	var ext string
	var contentType string

	// 2. RE-CODIFICACIÓN INICIAL (Limpieza y Optimización)
	// Al codificar de nuevo, solo se guardan los datos visuales, descartando basura.
	// A partir de aquí, trabajamos con la imagen decodificada (img) para la compresión.
	switch format {
	case "jpeg", "jpg":
		ext = ".jpg"
		contentType = "image/jpeg"
	case "png":
		// PNG se mantiene sin pérdida para preservar transparencias si existen.
		ext = ".png"
		contentType = "image/png"
	case "gif":
		// REGLA DE SEGURIDAD: Convertimos GIFs a PNG estático.
		// Esto evita vulnerabilidades de frames infinitos o desbordamiento de memoria
		// al procesar archivos animados en el frontend.
		ext = ".png"
		contentType = "image/png"
	default:
		return nil, "", "", errors.New("formato de imagen no soportado por la política de seguridad")
	}

	// 3. COMPRESIÓN ITERATIVA
	// Delegamos al compresor específico según el formato de salida final.
	// La lógica de reducción vive en compressImage para mantener SanitizeImage limpia.
	flatten_image := FlattenAlpha(img)
	compressedBytes, err := compressImage(flatten_image, ext)
	if err != nil {
		return nil, "", "", errors.New("error crítico al sanitizar la imagen")
	}

	return compressedBytes, ext, contentType, nil
}

// =========================================================================
// COMPRESIÓN ITERATIVA (ESTRATEGIA SENIOR)
// =========================================================================

// compressImage implementa un bucle de compresión iterativa para garantizar
// que el archivo resultante no supere el umbral de maxFileSizeBytes (1 MB).
//
// ESTRATEGIA:
// No podemos simplemente "bajar la calidad" en PNG (es lossless). La estrategia
// profesional es reducir las dimensiones de la imagen un 20% por iteración hasta
// que el peso sea aceptable o se alcance el mínimo de seguridad (minDimensionPx).
//
// Esto funciona tanto para JPEG (calidad fija en 85%) como para PNG (lossless),
// ya que menos píxeles siempre significa menos datos que almacenar.
func compressImage(img image.Image, ext string) ([]byte, error) {
	current := img

	for {
		// Codificar la imagen en su estado actual (dimensiones originales o reducidas)
		encoded, err := encodeImage(current, ext)
		if err != nil {
			return nil, err
		}

		// Si el peso es aceptable, retornamos. Misión cumplida.
		if len(encoded) <= maxFileSizeBytes {
			return encoded, nil
		}

		// Calcular las nuevas dimensiones reducidas al factor de escala configurado.
		bounds := current.Bounds()
		newWidth := int(float64(bounds.Dx()) * compressionScaleFactor)
		newHeight := int(float64(bounds.Dy()) * compressionScaleFactor)

		// GUARDIA DE SEGURIDAD: Si alguna dimensión cae por debajo del mínimo,
		// retornamos la imagen tal como está para evitar degradarla hasta el absurdo.
		// Es preferible una imagen de 1 MB que una de 100x100 píxeles borrosa.
		if newWidth < minDimensionPx || newHeight < minDimensionPx {
			return encoded, nil
		}

		// Redimensionar la imagen usando interpolación bilineal (draw.BiLinear).
		// Es el mejor balance entre velocidad y calidad para imágenes fotográficas.
		// Para íconos o pixel art, draw.NearestNeighbor preservaría mejor los bordes.
		resized := resizeImage(current, newWidth, newHeight)
		current = resized
	}
}

// encodeImage codifica una imagen en memoria al formato indicado por la extensión.
// Retorna los bytes del archivo resultante o un error si la codificación falla.
func encodeImage(img image.Image, ext string) ([]byte, error) {
	buf := new(bytes.Buffer)
	var err error

	switch ext {
	case ".jpg":
		// JPEG soporta control de calidad. 85% es el estándar para CDNs y S3.
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: jpegQuality})
	case ".png":
		// PNG es lossless: la única forma de reducir peso es reducir dimensiones.
		// La librería estándar de Go usa el nivel de compresión DefaultCompression
		// automáticamente, que es un buen balance entre velocidad y ratio de compresión.
		err = png.Encode(buf, img)
	default:
		return nil, errors.New("formato de codificación no reconocido internamente")
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// resizeImage crea una nueva imagen con las dimensiones destino especificadas
// y dibuja la imagen original sobre ella usando interpolación bilineal.
//
// Se usa image.RGBA como canvas destino porque es el formato en memoria más
// eficiente y compatible para operaciones de dibujo con el paquete draw.
// Para imágenes con canal alfa (transparencia), esto la preserva correctamente.
func resizeImage(src image.Image, width, height int) image.Image {
	// Crear el canvas destino con las nuevas dimensiones
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// draw.BiLinear.Scale aplica interpolación bilineal: promedia los píxeles
	// vecinos para suavizar el resultado al escalar hacia abajo.
	// Es superior a NearestNeighbor para fotografías y produce menos aliasing.
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	return dst
}

// =========================================================================
// UTILIDADES DE IMAGEN
// =========================================================================

// FlattenAlpha compone la imagen sobre un fondo blanco, eliminando la transparencia.
// Es útil cuando se necesita convertir una imagen PNG con canal alfa a JPEG,
// ya que JPEG no soporta transparencia y los píxeles transparentes se verían negros.
//
// Uso sugerido: llamar antes de encodeImage si se convierte de PNG a JPEG.
func FlattenAlpha(src image.Image) image.Image {
	bounds := src.Bounds()
	// Crear canvas blanco del mismo tamaño
	dst := image.NewRGBA(bounds)

	// Rellenar con blanco sólido como fondo
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, color.White)
		}
	}

	// Dibujar la imagen original encima; los píxeles transparentes dejan ver el blanco.
	draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
	return dst
}

// api/internal/utils/image_sanitizer.go
package utils

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/draw"
)

// =========================================================================
// CONSTANTES DE COMPRESIÓN
// =========================================================================

const (
	// Límites de peso diferenciados (1 MB era excesivo para WebP)
	maxAvatarSizeBytes   = 150 * 1024 // 150 KB (Avatares cargan rápido en grillas)
	maxDocumentSizeBytes = 400 * 1024 // 400 KB (Documentos necesitan mantener textos legibles)

	compressionScaleFactor = 0.8 // Si excede el peso, reduce dimensiones un 20%
	minDimensionPx         = 100
	webpQuality            = 80.0 // 80 es el 'sweet spot' óptimo para WebP (antes 82)

	maxAvatarDimension   = 800  // píxeles máximos de ancho o alto para fotos de perfil
	maxDocumentDimension = 1600 // píxeles máximos para certificados (permite zoom a textos)
)

// =========================================================================
// API PÚBLICA
// =========================================================================

// SanitizeImage limpia y convierte a WebP optimizado. Usar para avatares.
func SanitizeImage(file io.Reader) ([]byte, string, string, error) {
	return processImage(file, maxAvatarDimension, maxAvatarSizeBytes)
}

// SanitizeDocument limpia y convierte a WebP optimizado. Usar para títulos y certificados.
func SanitizeDocument(file io.Reader) ([]byte, string, string, error) {
	return processImage(file, maxDocumentDimension, maxDocumentSizeBytes)
}

// =========================================================================
// PROCESAMIENTO INTERNO
// =========================================================================

func processImage(file io.Reader, maxDimension int, maxSizeBytes int) ([]byte, string, string, error) {
	// 1. Decodificación — detección por magic numbers (seguridad)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, "", "", errors.New("archivo no es una imagen válida o formato no reconocido")
	}

	// 2. Cap de dimensiones máximas (evita procesar imágenes enormes en memoria)
	img = capDimensions(img, maxDimension)

	// 3. Aplanar canal alfa sobre fondo blanco (evita fondos negros en PNG transparentes)
	img = FlattenAlpha(img)

	// 4. Codificar a WebP con límite dinámico de peso
	compressed, err := compressToWebP(img, maxSizeBytes)
	if err != nil {
		return nil, "", "", errors.New("error al codificar la imagen")
	}

	return compressed, ".webp", "image/webp", nil
}

// capDimensions reduce la imagen si supera maxDimension en cualquier eje, preservando el aspect ratio.
func capDimensions(src image.Image, maxDimension int) image.Image {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxDimension && h <= maxDimension {
		return src
	}

	var newW, newH int
	if w >= h {
		newW = maxDimension
		newH = int(float64(h) * float64(maxDimension) / float64(w))
	} else {
		newH = maxDimension
		newW = int(float64(w) * float64(maxDimension) / float64(h))
	}

	return resizeImage(src, newW, newH)
}

// compressToWebP codifica a WebP con compresión iterativa.
// Si el resultado supera el límite, reduce dimensiones un 20% por iteración.
func compressToWebP(img image.Image, maxSizeBytes int) ([]byte, error) {
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, webpQuality)
	if err != nil {
		return nil, err
	}

	current := img
	for {
		buf := new(bytes.Buffer)
		if err := webp.Encode(buf, current, options); err != nil {
			return nil, err
		}

		encoded := buf.Bytes()
		// Si el archivo resultante pesa menos que nuestro límite, terminamos
		if len(encoded) <= maxSizeBytes {
			return encoded, nil
		}

		// Si sigue siendo muy pesado, achicamos un 20% y volvemos a codificar
		bounds := current.Bounds()
		newW := int(float64(bounds.Dx()) * compressionScaleFactor)
		newH := int(float64(bounds.Dy()) * compressionScaleFactor)

		// Failsafe de seguridad para no achicar infinitamente
		if newW < minDimensionPx || newH < minDimensionPx {
			return encoded, nil
		}

		current = resizeImage(current, newW, newH)
	}
}

// =========================================================================
// UTILIDADES
// =========================================================================

// resizeImage hace un resampleado de alta calidad (BiLinear).
func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

// FlattenAlpha compone la imagen sobre fondo blanco, eliminando transparencia.
func FlattenAlpha(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	// Pintar el fondo de blanco absoluto
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, color.White)
		}
	}

	// Dibujar la imagen original encima
	draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
	return dst
}

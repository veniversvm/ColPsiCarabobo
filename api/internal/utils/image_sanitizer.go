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
	maxFileSizeBytes       = 1 * 1024 * 1024 // 1 MB
	compressionScaleFactor = 0.8
	minDimensionPx         = 100
	webpQuality            = 82.0 // float32 requerido por kolesa encoder

	maxAvatarDimension   = 800
	maxDocumentDimension = 1600
)

// =========================================================================
// API PÚBLICA
// =========================================================================

// SanitizeImage limpia y convierte a WebP optimizado. Usar para avatares (cap 800px).
func SanitizeImage(file io.Reader) ([]byte, string, string, error) {
	return processImage(file, maxAvatarDimension)
}

// SanitizeDocument limpia y convierte a WebP optimizado. Usar para títulos y certificados (cap 1600px).
func SanitizeDocument(file io.Reader) ([]byte, string, string, error) {
	return processImage(file, maxDocumentDimension)
}

// =========================================================================
// PROCESAMIENTO INTERNO
// =========================================================================

func processImage(file io.Reader, maxDimension int) ([]byte, string, string, error) {
	// 1. Decodificación — detección por magic numbers
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, "", "", errors.New("archivo no es una imagen válida o formato no reconocido")
	}

	// 2. Cap de dimensiones máximas (evita procesar imágenes enormes en memoria)
	img = capDimensions(img, maxDimension)

	// 3. Aplanar canal alfa sobre fondo blanco
	img = FlattenAlpha(img)

	// 4. Codificar a WebP con compresión iterativa
	compressed, err := compressToWebP(img)
	if err != nil {
		return nil, "", "", errors.New("error al codificar la imagen")
	}

	return compressed, ".webp", "image/webp", nil
}

// capDimensions reduce la imagen si supera maxDimension en cualquier eje,
// preservando la relación de aspecto.
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
// Si el resultado supera 1 MB, reduce dimensiones un 20% por iteración.
func compressToWebP(img image.Image) ([]byte, error) {
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
		if len(encoded) <= maxFileSizeBytes {
			return encoded, nil
		}

		bounds := current.Bounds()
		newW := int(float64(bounds.Dx()) * compressionScaleFactor)
		newH := int(float64(bounds.Dy()) * compressionScaleFactor)

		if newW < minDimensionPx || newH < minDimensionPx {
			return encoded, nil
		}

		current = resizeImage(current, newW, newH)
	}
}

// =========================================================================
// UTILIDADES
// =========================================================================

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

// FlattenAlpha compone la imagen sobre fondo blanco, eliminando transparencia.
func FlattenAlpha(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, color.White)
		}
	}

	draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
	return dst
}

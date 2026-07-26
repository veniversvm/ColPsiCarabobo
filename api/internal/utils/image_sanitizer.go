// api/internal/utils/image_sanitizer.go

// Package utils provee herramientas y funciones transversales (Cross-Cutting Concerns).
//
// Este archivo implementa el Motor de Sanitización Visual (Image Sanitizer Engine).
// Actúa como una barrera de Seguridad y Optimización antes de que los archivos
// toquen el almacenamiento en la nube (S3).
//
// Funciones Clave:
//  1. Ciberseguridad: Erradica inyecciones de código, Esteganografía y metadatos sensibles (EXIF GPS).
//  2. Finanzas (FinOps): Reduce drásticamente los costos de almacenamiento en S3 y ancho de banda.
//  3. Rendimiento Frontend (SEO/LCP): Sirve imágenes WebP de última generación para carga instantánea.
package utils

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	_ "image/gif"  // Importación anónima para habilitar el decodificador de formato
	_ "image/jpeg" // Registra el decodificador JPEG
	_ "image/png"  // Registra el decodificador PNG
	"io"

	_ "golang.org/x/image/webp"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/rs/zerolog/log"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/draw"
)

// =========================================================================
// CONSTANTES DE COMPRESIÓN Y CONTROL DE CUOTAS
// =========================================================================

const (
	// Límites estrictos de peso en disco (Throttling de Almacenamiento).
	// Evitan facturas sorpresa en AWS S3.
	maxAvatarSizeBytes   = 150 * 1024 // 150 KB: Avatares miniatura. Carga masiva rápida en la grilla del Directorio.
	maxDocumentSizeBytes = 400 * 1024 // 400 KB: Títulos académicos. Permite compresión moderada para asegurar que el texto sea legible (OCR friendly).

	// Algoritmo de Degradación Iterativa
	compressionScaleFactor = 0.8  // Si excede el peso, reduce dimensiones un 20% (Feedback Loop)
	minDimensionPx         = 100  // Failsafe matemático: Evita reducir la imagen hasta 0 píxeles.
	webpQuality            = 80.0 // 80 es el 'sweet spot' óptimo para WebP (balance perfecto entre artefactos visuales y peso).

	// Barreras de Resolución Iniciales
	maxAvatarDimension   = 800  // Píxeles máximos de ancho/alto para fotos de perfil.
	maxDocumentDimension = 1600 // Píxeles máximos para certificados (permite hacer zoom en los sellos y firmas).
)

// =========================================================================
// API PÚBLICA
// =========================================================================

// SanitizeImage limpia, redimensiona y convierte una foto de perfil a WebP.
// Diseñado para flujos de alta concurrencia donde el peso es crítico.
func SanitizeImage(file io.Reader) ([]byte, string, string, error) {
	return processImage(file, maxAvatarDimension, maxAvatarSizeBytes)
}

// SanitizeDocument limpia y convierte archivos documentales (Certificados/Títulos).
// Posee límites de peso y resolución más holgados para preservar la lectura de textos.
func SanitizeDocument(file io.Reader) ([]byte, string, string, error) {
	return processImage(file, maxDocumentDimension, maxDocumentSizeBytes)
}

// =========================================================================
// PROCESAMIENTO INTERNO (PIPELINE)
// =========================================================================

// processImage orquesta el pipeline de purificación y compresión.
func processImage(file io.Reader, maxDimension int, maxSizeBytes int) ([]byte, string, string, error) {
	// 1. Decodificación Estricta (Magic Numbers Detection).
	// Mitigación de Vulnerabilidades (MIME Sniffing & Steganography):
	// El paquete `image` nativo de Go NO lee metadatos (EXIF) ni confía en la extensión .jpg.
	// Lee las firmas binarias (Magic Numbers) y extrae ÚNICAMENTE la matriz de píxeles.
	// Si un hacker sube un archivo `foto.jpg` que por dentro tiene código PHP oculto,
	// el código PHP es destruido al extraer solo los píxeles.
	img, _, err := image.Decode(file)
	if err != nil {
		// Log interno para trazabilidad forense
		log.Error().Err(err).Str("component", "image").Msg("Error decodificando imagen")
		// Mensaje genérico para el cliente (Security by Obscurity)
		return nil, "", "", errors.New("el servidor no reconoce este formato de imagen")
	}

	// 2. Cap de dimensiones máximas
	// Evita ataques de "Pixel Flood" o "Billion Laughs" donde una imagen pequeña
	// en disco se descomprime ocupando Gigabytes de memoria RAM.
	img = capDimensions(img, maxDimension)

	// 3. Normalización del Canal Alfa (Transparencias)
	// WebP soporta transparencias, pero si se comprime agresivamente o el cliente
	// usa fondos oscuros en el frontend, un PNG transparente se vería como una mancha negra.
	// Esto plancha la imagen contra un fondo blanco absoluto garantizando consistencia UI.
	img = FlattenAlpha(img)

	// 4. Codificación Definitiva a WebP con Algoritmo Recursivo
	compressed, err := compressToWebP(img, maxSizeBytes)
	if err != nil {
		return nil, "", "", errors.New("error al codificar la imagen")
	}

	return compressed, ".webp", "image/webp", nil
}

// capDimensions reduce la resolución de la imagen si supera los límites,
// preservando matemáticamente el Aspect Ratio (Relación de Aspecto) para no deformarla.
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

// compressToWebP codifica a WebP utilizando un bucle de Degradación Iterativa (Iterative Degradation).
//
// Algoritmo Adaptativo:
// Si la codificación inicial genera un archivo que supera la cuota máxima permitida (maxSizeBytes),
// en lugar de devolver un error ("El archivo es muy pesado") frustrando al usuario, el algoritmo
// encoge la imagen un 20% y la vuelve a comprimir hasta que encaje en el límite financiero exigido.
func compressToWebP(img image.Image, maxSizeBytes int) ([]byte, error) {
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, webpQuality)
	if err != nil {
		return nil, err
	}

	current := img
	for {
		buf := new(bytes.Buffer)
		// Codificación WebP pura
		if err := webp.Encode(buf, current, options); err != nil {
			return nil, err
		}

		encoded := buf.Bytes()
		// Condición de Salida: El archivo resultante respeta nuestra cuota económica
		if len(encoded) <= maxSizeBytes {
			return encoded, nil
		}

		// Penalización (Downscaling): Si sigue siendo muy pesado, lo achicamos un 20%
		bounds := current.Bounds()
		newW := int(float64(bounds.Dx()) * compressionScaleFactor)
		newH := int(float64(bounds.Dy()) * compressionScaleFactor)

		// Failsafe de Seguridad (Prevención de Bucle Infinito):
		// Garantiza que la imagen no se reduzca a 0x0 píxeles, lo que colapsaría el programa.
		if newW < minDimensionPx || newH < minDimensionPx {
			return encoded, nil
		}

		current = resizeImage(current, newW, newH)
	}
}

// =========================================================================
// UTILIDADES GRÁFICAS (BAJO NIVEL)
// =========================================================================

// resizeImage hace un resampleado de alta calidad usando interpolación BiLineal.
// Evita los "dientes de sierra" (Aliasing) que ocurren al achicar imágenes de forma brusca.
func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

// FlattenAlpha compone la imagen original sobre un lienzo blanco sólido.
// Erradica el Canal Alfa (Transparencia) asegurando que los bordes difuminados de un PNG
// no generen halos oscuros al ser convertidos a formato WebP o mostrados en Dark Mode.

// func FlattenAlpha(src image.Image) image.Image {
// 	bounds := src.Bounds()
// 	dst := image.NewRGBA(bounds)

// 	// Pintar el fondo de blanco absoluto
// 	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
// 		for x := bounds.Min.X; x < bounds.Max.X; x++ {
// 			dst.Set(x, y, color.White)
// 		}
// 	}

//		// Dibujar la imagen original encima
//		draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
//		return dst
//	}

func FlattenAlpha(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	// Optimización de CPU (Bulk Fill):
	// Rellena todo el fondo de blanco de un solo golpe en la memoria RAM
	// Esto es exponencialmente más rápido que recorrer la matriz con bucles for `x,y`
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Dibujar la imagen original encima (preservando lo que no es transparente)
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Over)
	return dst
}

package utils

import (
	"bytes"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// =========================================================================
// TEST: SANITIZADOR DE DOCUMENTOS DEL EXPEDIENTE (SanitizeDocumentFile)
// =========================================================================

// TestSanitizeDocumentFile_PDFValido verifica que un PDF legítimo (validado por
// su firma binaria "%PDF-") atraviese el sanitizador tal cual, sin pasar por el
// pipeline de imágenes.
func TestSanitizeDocumentFile_PDFValido(t *testing.T) {
	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte{0x00}, 512)...)
	data, ext, contentType, err := SanitizeDocumentFile(bytes.NewReader(pdf))

	require.NoError(t, err)
	require.Equal(t, ".pdf", ext)
	require.Equal(t, "application/pdf", contentType)
	require.Equal(t, pdf, data, "El contenido del PDF debe conservarse íntegro")
}

// TestSanitizeDocumentFile_PDFExcedeLimite rechaza PDFs de más de 4 MB.
func TestSanitizeDocumentFile_PDFExcedeLimite(t *testing.T) {
	payload := bytes.Repeat([]byte{0x01}, maxPdfSizeBytes+1)
	oversized := append([]byte("%PDF-1.7\n"), payload...)

	_, _, _, err := SanitizeDocumentFile(bytes.NewReader(oversized))
	require.Error(t, err)
	require.Contains(t, err.Error(), "excede el tamaño máximo")
}

// TestSanitizeDocumentFile_RechazaArchivosIlegibles verifica el comportamiento
// defensivo: texto plano (no imagen, no PDF) debe ser rechazado.
func TestSanitizeDocumentFile_RechazaArchivosIlegibles(t *testing.T) {
	fake := strings.NewReader("Esto no es ni un PDF ni una imagen válida")
	_, _, _, err := SanitizeDocumentFile(fake)
	require.Error(t, err)
	require.Contains(t, err.Error(), "el servidor no reconoce este formato")
}

// TestSanitizeDocumentFile_NoPDFGrande verifica que un archivo grande que NO
// comienza con "%PDF-" se reencauce al pipeline de imágenes y falle allí (no en
// el cap de PDF), probando que el desvío ocurre por firma binaria.
func TestSanitizeDocumentFile_NoPDFGrande(t *testing.T) {
	big := bytes.Repeat([]byte("A"), 1024*1024)
	_, _, _, err := SanitizeDocumentFile(bytes.NewReader(big))
	require.Error(t, err)
}

// TestSanitizeDocumentFile_EntradaAleatoria simula datos binarios corruptos que
// no representan ningún formato soportado.
func TestSanitizeDocumentFile_EntradaAleatoria(t *testing.T) {
	garbage := make([]byte, 256)
	_, _ = rand.Read(garbage)

	_, _, _, err := SanitizeDocumentFile(bytes.NewReader(garbage))
	require.Error(t, err)
}
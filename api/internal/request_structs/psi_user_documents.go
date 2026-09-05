// api/internal/request_structs/psi_user_documents.go

// Package request_structs contiene los contratos de entrada/salida del módulo de
// Registro Digital de Documentos del expediente del psicólogo.
//
// La gestión de estos documentos es EXCLUSIVA del personal administrativo; el
// psicólogo solo puede consultarlos. Por eso los structs de escritura (Create/
// Update) solo se usan desde las rutas admin.
package request_structs

import (
	"errors"
	"strings"
	"time"
)

// Categorías soportadas de documentos (coinciden con domain.DocumentType).
// Se mantienen como strings locales para evitar un ciclo de importación
// domain → request_structs → domain.
const (
	DocTypeCedula     = "cedula"
	DocTypeTitulo     = "titulo"
	DocTypeRif        = "rif"
	DocTypeSolvencia  = "solvencia"
	DocTypeComprobante = "comprobante"
	DocTypeOtro       = "otro"
)

// ValidDocumentTypes devuelve las categorías aceptadas.
func ValidDocumentTypes() map[string]bool {
	return map[string]bool{
		DocTypeCedula: true, DocTypeTitulo: true, DocTypeRif: true,
		DocTypeSolvencia: true, DocTypeComprobante: true, DocTypeOtro: true,
	}
}

// Errores de validación del módulo de documentos.
var (
	// ErrDocumentInvalidRequest se devuelve cuando falta el archivo adjunto.
	ErrDocumentInvalidRequest = errors.New("el archivo del documento es obligatorio")
	// ErrDocumentInvalidTitle se devuelve cuando falta la etiqueta descriptiva del documento.
	ErrDocumentInvalidTitle = errors.New("el título del documento es obligatorio")
	// ErrDocumentInvalidType se devuelve cuando la categoría del documento no es válida.
	ErrDocumentInvalidType = errors.New("tipo de documento inválido")
)

// CreatePsiUserDocumentRequest define la carga útil para registrar un nuevo
// documento digital en el expediente de un psicólogo. Se construye en el handler
// a partir del form-data multipart (incluye el archivo adjunto).
type CreatePsiUserDocumentRequest struct {
	// Title es la etiqueta libre que describe el documento.
	// Ej: "Cédula V-123456 anverso", "Comprobante de solvencia 2025".
	Title string
	// DocumentType es la categoría del documento (cedula, titulo, rif, solvencia, comprobante, otro).
	DocumentType string
	// Notes son observaciones internas opcionales sobre el documento.
	Notes string
	// DocumentDate es la fecha opcional a la que corresponde el documento.
	DocumentDate *time.Time
}

// UpdatePsiUserDocumentRequest define la carga útil para editar los metadatos de
// un documento existente (y opcionalmente reemplazar su archivo). Los campos son
// punteros para soportar semántica PATCH: solo se actualiza lo que viene presente.
type UpdatePsiUserDocumentRequest struct {
	// Title es la nueva etiqueta descriptiva del documento.
	Title *string
	// DocumentType es la nueva categoría del documento.
	DocumentType *string
	// Notes son las nuevas observaciones internas.
	Notes *string
	// DocumentDate es la nueva fecha del documento (nil = no se toca).
	DocumentDate *time.Time
	// ClearDocumentDate permite vaciar explícitamente la fecha del documento.
	ClearDocumentDate bool
}

// Sanitize limpia los campos de texto libre del request de creación.
func (r *CreatePsiUserDocumentRequest) Sanitize() {
	r.Title = strings.TrimSpace(r.Title)
	r.Notes = strings.TrimSpace(r.Notes)
}

// Validate valida el request de creación antes de persistir.
func (r *CreatePsiUserDocumentRequest) Validate() error {
	switch {
	case r.Title == "":
		return ErrDocumentInvalidTitle
	case !ValidDocumentTypes()[r.DocumentType]:
		return ErrDocumentInvalidType
	}
	return nil
}

func (r *UpdatePsiUserDocumentRequest) Normalize() {
	if r.Title != nil {
		scrubbed := strings.TrimSpace(*r.Title)
		r.Title = &scrubbed
	}
	if r.Notes != nil {
		scrubbed := strings.TrimSpace(*r.Notes)
		r.Notes = &scrubbed
	}
}

// IsTypeValid valida una categoría de documento.
func IsTypeValid(dt string) bool {
	return ValidDocumentTypes()[dt]
}
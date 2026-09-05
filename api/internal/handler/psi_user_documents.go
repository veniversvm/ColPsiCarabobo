// api/internal/handler/psi_user_documents.go
package handler

import (
	"errors"
	"mime/multipart"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// parseDocumentDate convierte un string "2006-01-02" en un *time.Time (nil si vacío).
func parseDocumentDate(v string) *time.Time {
	if v == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil
	}
	return &t
}

// ListDocumentsByAdmin godoc
// @Summary      Listar documentos digitales (Admin)
// @Description  Retorna los documentos digitales del expediente de un psicólogo (CI, título, RIF, comprobantes, etc.). La gestión de estos documentos es exclusiva del personal administrativo autorizado.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Produce      json
// @Param        id   path      string  true  "UUID del Psicólogo"
// @Success      200  {array}   domain.PsiUserDocument
// @Failure      400  {object}  map[string]string "ID inválido"
// @Failure      403  {object}  map[string]string "Permisos insuficientes"
// @Failure      404  {object}  map[string]string "Psicólogo no encontrado"
// @Router       /admin/psi/{id}/documents [get]
func (h *PsiHandler) ListDocumentsByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	docs, err := h.service.ListDocumentsByAdmin(c.UserContext(), admin, targetID)
	if err != nil {
		if errors.Is(err, domain.ErrPsiNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(docs)
}

// AddDocumentByAdmin godoc
// @Summary      Registrar documento digital (Admin)
// @Description  Registra un nuevo documento digital en el expediente de un psicólogo. Acepta imágenes (convertidas a WebP) y PDF.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Accept       multipart/form-data
// @Produce      json
// @Param        id            path     string true  "UUID del Psicólogo"
// @Param        file          formData file   true  "Archivo del documento (imagen o PDF)"
// @Param        title         formData string true  "Etiqueta descriptiva (ej: Cédula anverso, Solvencia 2025)"
// @Param        document_type formData string false "Categoría: cedula | titulo | rif | solvencia | comprobante | otro (default: otro)"
// @Param        notes         formData string false "Observaciones internas"
// @Param        document_date formData string false "Fecha del documento (YYYY-MM-DD)"
// @Success      201 {object} domain.PsiUserDocument
// @Failure      400 {object} map[string]string "Validación fallida"
// @Failure      403 {object} map[string]string "Permisos insuficientes"
// @Failure      404 {object} map[string]string "Psicólogo no encontrado"
// @Router       /admin/psi/{id}/documents [post]
func (h *PsiHandler) AddDocumentByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID proporcionado no es un UUID válido"})
	}

	req := request_structs.CreatePsiUserDocumentRequest{
		Title:        c.FormValue("title"),
		DocumentType: c.FormValue("document_type"),
		Notes:        c.FormValue("notes"),
		DocumentDate: parseDocumentDate(c.FormValue("document_date")),
	}
	if req.DocumentType == "" {
		req.DocumentType = request_structs.DocTypeOtro
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": request_structs.ErrDocumentInvalidRequest.Error()})
	}

	doc, err := h.service.AddDocumentByAdmin(c.UserContext(), admin, targetID, req, file)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPsiNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, request_structs.ErrDocumentInvalidTitle),
			errors.Is(err, request_structs.ErrDocumentInvalidType),
			errors.Is(err, request_structs.ErrDocumentInvalidRequest):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(doc)
}

// UpdateDocumentByAdmin godoc
// @Summary      Editar documento digital (Admin)
// @Description  Edita los metadatos de un documento del expediente y, opcionalmente, reemplaza su archivo.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Accept       multipart/form-data
// @Produce      json
// @Param        id                 path     string false "UUID del Psicólogo"
// @Param        docId              path     string true  "UUID del documento"
// @Param        file               formData file   false "Nuevo archivo (imagen o PDF)"
// @Param        title              formData string false "Nueva etiqueta descriptiva"
// @Param        document_type      formData string false "Nueva categoría"
// @Param        notes              formData string false "Nuevas observaciones"
// @Param        document_date      formData string false "Nueva fecha del documento (YYYY-MM-DD)"
// @Param        clear_document_date formData string false "1 para vaciar la fecha del documento"
// @Success      200 {object} domain.PsiUserDocument
// @Failure      400 {object} map[string]string "Validación fallida"
// @Failure      403 {object} map[string]string "Permisos insuficientes"
// @Failure      404 {object} map[string]string "Documento no encontrado"
// @Router       /admin/psi/{id}/documents/{docId} [patch]
func (h *PsiHandler) UpdateDocumentByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	docID, err := uuid.Parse(c.Params("docId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID del documento no es un UUID válido"})
	}

	req := request_structs.UpdatePsiUserDocumentRequest{}
	if v := c.FormValue("title"); v != "" {
		req.Title = &v
	}
	if v := c.FormValue("document_type"); v != "" {
		req.DocumentType = &v
	}
	if v := c.FormValue("notes"); v != "" {
		req.Notes = &v
	}
	if v := parseDocumentDate(c.FormValue("document_date")); v != nil {
		req.DocumentDate = v
	}
	if c.FormValue("clear_document_date") == "1" {
		req.ClearDocumentDate = true
	}

	var file *multipart.FileHeader
	if f, err := c.FormFile("file"); err == nil {
		file = f
	}

	doc, err := h.service.UpdateDocumentByAdmin(c.UserContext(), admin, docID, req, file)
	if err != nil {
		switch {
		case errors.Is(err, request_structs.ErrDocumentInvalidTitle),
			errors.Is(err, request_structs.ErrDocumentInvalidType):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrDocumentNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(doc)
}

// DeleteDocumentByAdmin godoc
// @Summary      Eliminar documento digital (Admin)
// @Description  Elimina lógicamente un documento del expediente y borra su archivo del bucket.
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Produce      json
// @Param        id     path string true "UUID del Psicólogo"
// @Param        docId  path string true "UUID del documento"
// @Success      200 {object} map[string]string "message: Documento eliminado"
// @Failure      400 {object} map[string]string "ID inválido"
// @Failure      403 {object} map[string]string "Permisos insuficientes"
// @Failure      404 {object} map[string]string "Documento no encontrado"
// @Router       /admin/psi/{id}/documents/{docId} [delete]
func (h *PsiHandler) DeleteDocumentByAdmin(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	docID, err := uuid.Parse(c.Params("docId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ID del documento no es un UUID válido"})
	}

	if err := h.service.DeleteDocumentByAdmin(c.UserContext(), admin, docID); err != nil {
		if errors.Is(err, domain.ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Documento eliminado"})
}

// GetMyDocuments godoc
// @Summary      Mis documentos digitales (Psi)
// @Description  Retorna los documentos digitales del psicólogo autenticado (CI, título, RIF, comprobantes, etc.). Es SOLO LECTURA: el psicólogo nunca puede editar ni eliminar estos documentos; su gestión es exclusiva de la administración.
// @Security     BearerAuth
// @Tags         Psicólogos - Expediente
// @Produce      json
// @Success      200 {array} domain.PsiUserDocument
// @Failure      401 {object} map[string]string "No autenticado"
// @Router       /psi/me/documents [get]
func (h *PsiHandler) GetMyDocuments(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	docs, err := h.service.GetMyDocuments(c.UserContext(), psi)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudieron cargar los documentos"})
	}

	return c.JSON(docs)
}
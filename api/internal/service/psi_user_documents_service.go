// api/internal/service/psi_user_documents_service.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona el submódulo de Registro Digital de Documentos del
// expediente del psicólogo (CI, título, RIF, comprobantes de solvencia, etc.).
// Por diseño, la gestión (carga, edición, eliminación) es EXCLUSIVA del personal
// administrativo autorizado; el psicólogo SOLO puede consultar sus propios
// documentos (nunca editarlos ni borrarlos).
package service

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// resolveDocumentURL convierte la S3 key interna de un documento en su URL
// pública antes de serializarlo. No muta la base de datos.
func (s *PsiService) resolveDocumentURL(doc *domain.PsiUserDocument) {
	if doc == nil {
		return
	}
	doc.S3Key = s.publicURL(doc.S3Key)
}

// processDocumentFile sane el archivo (imagen→WebP o PDF validado) y lo sube a S3
// en la carpeta "documents". Devuelve la key, el MIME y el tamaño en bytes.
func (s *PsiService) processDocumentFile(ctx context.Context, psiID uuid.UUID, file *multipart.FileHeader) (key, mimeType string, sizeBytes int64, err error) {
	src, err := file.Open()
	if err != nil {
		return "", "", 0, domain.ErrInvalidRequest
	}
	defer src.Close()

	cleanBytes, ext, contentType, err := utils.SanitizeDocumentFile(src)
	if err != nil {
		return "", "", 0, err
	}

	shortUUID := uuid.Must(uuid.NewV7()).String()[:6]
	filename := fmt.Sprintf("psi_%s_doc_%s%s", psiID.String(), shortUUID, ext)
	newKey, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "documents", filename, contentType)
	if err != nil {
		return "", "", 0, err
	}

	return newKey, contentType, int64(len(cleanBytes)), nil
}

// AddDocumentByAdmin registra un nuevo documento digital en el expediente de un psicólogo.
func (s *PsiService) AddDocumentByAdmin(ctx context.Context, admin *domain.UserAdmin, psiID uuid.UUID, req request_structs.CreatePsiUserDocumentRequest, file *multipart.FileHeader) (*domain.PsiUserDocument, error) {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, domain.ErrInsufficientPerms
	}

	// 2. INTEGRIDAD REFERENCIAL
	if _, err := s.repo.GetByID(ctx, psiID); err != nil {
		return nil, domain.ErrPsiNotFound
	}

	// 3. VALIDACIÓN DE LA CARGA ÚTIL
	if file == nil {
		return nil, request_structs.ErrDocumentInvalidRequest
	}
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 4. PROCESAR Y SUBIR EL ARCHIVO
	key, mimeType, sizeBytes, err := s.processDocumentFile(ctx, psiID, file)
	if err != nil {
		return nil, err
	}

	// 5. PERSISTENCIA CON AUDITORÍA DEL EJECUTOR
	doc := &domain.PsiUserDocument{
		ID:           uuid.Must(uuid.NewV7()),
		AuditModel:   domain.AuditModel{CreateBy: admin.Username, CreateById: &admin.ID, UpdateBy: admin.Username, UpdateById: &admin.ID},
		PsiUserID:    psiID,
		DocumentType: domain.DocumentType(req.DocumentType),
		Title:        req.Title,
		Notes:        req.Notes,
		DocumentDate: req.DocumentDate,
		S3Key:        key,
		Filename:     file.Filename,
		MimeType:     mimeType,
		SizeBytes:    sizeBytes,
	}

	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		_ = s.s3Client.DeleteFile(context.Background(), key)
		return nil, err
	}

	s.resolveDocumentURL(doc)
	return doc, nil
}

// ListDocumentsByAdmin recupera los documentos digitales del expediente de un psicólogo.
func (s *PsiService) ListDocumentsByAdmin(ctx context.Context, admin *domain.UserAdmin, psiID uuid.UUID) ([]domain.PsiUserDocument, error) {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, domain.ErrInsufficientPerms
	}

	// 2. INTEGRIDAD REFERENCIAL
	if _, err := s.repo.GetByID(ctx, psiID); err != nil {
		return nil, domain.ErrPsiNotFound
	}

	// 3. CONSULTA Y RESOLUCIÓN DE URLs
	docs, err := s.repo.ListDocuments(ctx, psiID)
	if err != nil {
		return nil, err
	}
	for i := range docs {
		s.resolveDocumentURL(&docs[i])
	}
	return docs, nil
}

// GetMyDocuments recupera los documentos digitales del psicólogo autenticado.
// Es SOLO LECTURA: no existen rutas de escritura para el psicólogo sobre sus
// documentos, garantizando que nunca pueda editarlos ni borrarlos.
func (s *PsiService) GetMyDocuments(ctx context.Context, psi *domain.PsiUserModel) ([]domain.PsiUserDocument, error) {
	docs, err := s.repo.ListDocuments(ctx, psi.ID)
	if err != nil {
		return nil, err
	}
	for i := range docs {
		s.resolveDocumentURL(&docs[i])
	}
	return docs, nil
}

// UpdateDocumentByAdmin edita los metadatos de un documento existente y, de forma
// opcional, reemplaza su archivo en S3.
func (s *PsiService) UpdateDocumentByAdmin(ctx context.Context, admin *domain.UserAdmin, docID uuid.UUID, req request_structs.UpdatePsiUserDocumentRequest, file *multipart.FileHeader) (*domain.PsiUserDocument, error) {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi {
		return nil, domain.ErrInsufficientPerms
	}

	// 2. EXISTENCIA (404 si no está)
	doc, err := s.repo.GetDocument(ctx, docID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}

	// 3. NORMALIZACIÓN DE LA CARGA ÚTIL
	req.Normalize()
	if req.Title != nil {
		if *req.Title == "" {
			return nil, request_structs.ErrDocumentInvalidTitle
		}
		doc.Title = *req.Title
	}
	if req.DocumentType != nil {
		if !request_structs.IsTypeValid(*req.DocumentType) {
			return nil, request_structs.ErrDocumentInvalidType
		}
		doc.DocumentType = domain.DocumentType(*req.DocumentType)
	}
	if req.Notes != nil {
		doc.Notes = *req.Notes
	}
	if req.ClearDocumentDate {
		doc.DocumentDate = nil
	} else if req.DocumentDate != nil {
		doc.DocumentDate = req.DocumentDate
	}

	// 4. AUDITORÍA DEL EDITOR
	doc.UpdateBy = admin.Username
	doc.UpdateById = &admin.ID

	// 5. REEMPLAZO OPCIONAL DEL ARCHIVO
	var oldS3Key string
	if file != nil {
		newKey, mimeType, sizeBytes, err := s.processDocumentFile(ctx, doc.PsiUserID, file)
		if err != nil {
			return nil, err
		}
		oldS3Key = doc.S3Key
		doc.S3Key = newKey
		doc.Filename = file.Filename
		doc.MimeType = mimeType
		doc.SizeBytes = sizeBytes
	}

	// 6. PERSISTENCIA
	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return nil, err
	}

	// 7. ELIMINAR EL ARCHIVO ANTERIOR (solo tras persistir con éxito)
	if oldS3Key != "" {
		_ = s.s3Client.DeleteFile(context.Background(), oldS3Key)
	}

	s.resolveDocumentURL(doc)
	return doc, nil
}

// DeleteDocumentByAdmin elimina lógicamente un documento del expediente y borra
// su archivo del bucket.
func (s *PsiService) DeleteDocumentByAdmin(ctx context.Context, admin *domain.UserAdmin, docID uuid.UUID) error {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanDeletePsi {
		return domain.ErrInsufficientPerms
	}

	// 2. EXISTENCIA (404 si no está)
	doc, err := s.repo.GetDocument(ctx, docID)
	if err != nil {
		return domain.ErrDocumentNotFound
	}

	// 3. ELIMINACIÓN LÓGICA + ARCHIVO EN S3
	if err := s.repo.DeleteDocument(ctx, docID); err != nil {
		return err
	}
	if doc.S3Key != "" {
		_ = s.s3Client.DeleteFile(context.Background(), doc.S3Key)
	}
	return nil
}
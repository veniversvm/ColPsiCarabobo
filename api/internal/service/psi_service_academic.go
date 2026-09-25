package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// AddPostGrade creates a new academic post-grade record with optional certificate images uploaded to S3.
func (s *PsiService) AddPostGrade(ctx context.Context, psi *domain.PsiUserModel, req request_structs.CreatePostGradeRequest, files []*multipart.FileHeader) error {

	postGrade := &domain.PsiUserPostGrade{
		AuditModel: domain.AuditModel{
			CreateById: &psi.ID,
			CreateBy:   psi.Username,
			UpdateById: &psi.ID,
			UpdateBy:   psi.Username,
		},
		PsiUserID:      psi.ID,
		Title:          req.Title,
		University:     req.University,
		GraduationYear: req.GraduationYear,
		Description:    req.Description,
		Active:         true,
	}

	uploadHelper := func(fh *multipart.FileHeader) (string, error) {
		if fh == nil {
			return "", nil
		}

		src, err := fh.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return "", fmt.Errorf("error en imagen: %v", err)
		}

		filename := uuid.Must(uuid.NewV7()).String() + ext
		return s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "certificates", filename, contentType)
	}

	var err error
	if len(files) > 0 {
		postGrade.PicOneS3Key, err = uploadHelper(files[0])
	}
	if err != nil {
		return err
	}

	if len(files) > 1 {
		postGrade.PicTwoS3Key, err = uploadHelper(files[1])
	}
	if err != nil {
		return err
	}

	if len(files) > 2 {
		postGrade.PicThreeS3Key, err = uploadHelper(files[2])
	}
	if err != nil {
		return err
	}

	if err := s.repo.CreatePostGrade(ctx, postGrade); err != nil {
		return err
	}

	// Bitácora de cambios: alta de un título académico (auto-gestión).
	evt := auditPsiSelfEvent(psi, domain.AuditActionCreate)
	changes := map[string]domain.AuditChange{
		"post_grade_title": {To: postGrade.Title},
	}
	if postGrade.University != "" {
		changes["post_grade_university"] = domain.AuditChange{To: postGrade.University}
	}
	if postGrade.GraduationYear != 0 {
		changes["post_grade_graduation_year"] = domain.AuditChange{To: postGrade.GraduationYear}
	}
	if postGrade.Description != "" {
		changes["post_grade_description"] = domain.AuditChange{To: postGrade.Description}
	}
	evt.Changes = changes
	RecordAudit(ctx, evt)
	return nil
}

// UpdatePostGrade updates an existing academic post-grade record, replacing any provided certificate images.
func (s *PsiService) UpdatePostGrade(ctx context.Context, psi *domain.PsiUserModel, pgID uuid.UUID, req request_structs.UpdatePostGradeRequest, fileMap map[string]*multipart.FileHeader) error {

	pg, err := s.repo.GetPostGradeByID(ctx, pgID)
	if err != nil {
		return errors.New("título académico no encontrado")
	}

	if pg.PsiUserID != psi.ID {
		return domain.ErrPermissionDenied
	}

	// Snapshot previo para el diff por campo de la bitácora.
	beforeSnapshot := postGradeSnapshot(pg)
	imageReplaced := false

	pg.UpdateBy = psi.Username
	pg.UpdateById = &psi.ID
	pg.UpdatedAt = time.Now()

	if req.Title != nil {
		pg.Title = *req.Title
	}
	if req.University != nil {
		pg.University = *req.University
	}
	if req.GraduationYear != nil {
		pg.GraduationYear = *req.GraduationYear
	}
	if req.Description != nil {
		pg.Description = *req.Description
	}

	replaceImage := func(newFile *multipart.FileHeader, oldKey string) (string, error) {
		src, err := newFile.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return "", err
		}

		filename := uuid.Must(uuid.NewV7()).String() + ext
		newKey, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "certificates", filename, contentType)
		if err != nil {
			return "", err
		}

		if oldKey != "" {
			_ = s.s3Client.DeleteFile(ctx, oldKey)
		}
		return newKey, nil
	}

	if file, ok := fileMap["pic_one"]; ok {
		pg.PicOneS3Key, err = replaceImage(file, pg.PicOneS3Key)
		if err != nil {
			return err
		}
		imageReplaced = true
	}
	if file, ok := fileMap["pic_two"]; ok {
		pg.PicTwoS3Key, err = replaceImage(file, pg.PicTwoS3Key)
		if err != nil {
			return err
		}
		imageReplaced = true
	}
	if file, ok := fileMap["pic_three"]; ok {
		pg.PicThreeS3Key, err = replaceImage(file, pg.PicThreeS3Key)
		if err != nil {
			return err
		}
		imageReplaced = true
	}

	if err := s.repo.UpdatePostGrade(ctx, pg); err != nil {
		return err
	}

	// Bitácora de cambios: edición de título académico (auto-gestión).
	evt := auditPsiSelfEvent(psi, domain.AuditActionUpdate)
	evt.Changes = BuildDiff(beforeSnapshot, postGradeSnapshot(pg))
	if imageReplaced {
		evt.Metadata = map[string]any{"certificate_images_updated": true}
	}
	RecordAudit(ctx, evt)
	return nil
}

// DeletePostGrade elimina un título académico del propio expediente del
// psicólogo (auto-gestión) con chequeo de propiedad (IDOR), limpia los soportes
// (certificados) del bucket y registra la baja en la bitácora de cambios.
func (s *PsiService) DeletePostGrade(ctx context.Context, psi *domain.PsiUserModel, pgID uuid.UUID) error {
	pg, err := s.repo.GetPostGradeByID(ctx, pgID)
	if err != nil {
		return errors.New("título académico no encontrado")
	}
	if pg.PsiUserID != psi.ID {
		return domain.ErrPermissionDenied
	}

	if err := s.repo.DeletePostGrade(ctx, pgID); err != nil {
		return fmt.Errorf("error al eliminar el título académico: %w", err)
	}

	// Limpieza best-effort de los soportes (certificados) subidos al bucket.
	if s.s3Client != nil {
		for _, key := range []string{pg.PicOneS3Key, pg.PicTwoS3Key, pg.PicThreeS3Key} {
			if key != "" {
				_ = s.s3Client.DeleteFile(ctx, key)
			}
		}
	}

	// Bitácora de cambios: baja de título académico (auto-gestión).
	evt := auditPsiSelfEvent(psi, domain.AuditActionDelete)
	evt.Changes = postGradeRemovalChanges(pg)
	RecordAudit(ctx, evt)
	return nil
}

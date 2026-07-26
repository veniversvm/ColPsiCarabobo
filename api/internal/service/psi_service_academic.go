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

	return s.repo.CreatePostGrade(ctx, postGrade)
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
	}
	if file, ok := fileMap["pic_two"]; ok {
		pg.PicTwoS3Key, err = replaceImage(file, pg.PicTwoS3Key)
		if err != nil {
			return err
		}
	}
	if file, ok := fileMap["pic_three"]; ok {
		pg.PicThreeS3Key, err = replaceImage(file, pg.PicThreeS3Key)
		if err != nil {
			return err
		}
	}

	return s.repo.UpdatePostGrade(ctx, pg)
}

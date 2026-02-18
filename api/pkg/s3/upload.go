package s3

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// UploadFile gestiona la subida de archivos multipart a S3/MinIO.
// - Renombra el archivo con UUID para evitar colisiones.
// - Organiza los archivos en carpetas virtuales (ej: "posts/").
// - Retorna la "Key" (ruta relativa) para guardarla en la base de datos.
func (s *S3Client) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error) {
	// 1. Abrir el stream del archivo
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer src.Close()

	// 2. Generar nombre único (Seguridad y Unicidad)
	ext := filepath.Ext(fileHeader.Filename)
	// Ejemplo: "posts/550e8400-e29b-41d4-a716-446655440000.png"
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// 3. Subir a S3 usando Streaming (bajo consumo de RAM)
	_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(filename),
		Body:        src,
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
	})

	if err != nil {
		return "", fmt.Errorf("falló subida a S3: %w", err)
	}

	return filename, nil
}

// UploadStream sube bytes crudos a S3. Útil para imágenes procesadas en memoria.
func (s *S3Client) UploadStream(ctx context.Context, reader io.Reader, folder string, filename string, contentType string) (string, error) {
	key := fmt.Sprintf("%s/%s", folder, filename)

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("falló subida a S3: %w", err)
	}

	return key, nil
}

func (s *S3Client) DeleteFile(ctx context.Context, key string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("falló eliminación en S3: %w", err)
	}
	return nil
}

// GetPresignedURL genera una URL temporal para ver archivos privados (Opcional, futuro uso).
// Útil si decides que las imágenes no sean públicas por defecto.
/*
func (s *S3Client) GetPresignedURL(ctx context.Context, key string) (string, error) {
	presignClient := s3.NewPresignClient(s.Client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", err
	}
	return req.URL, nil
}
*/

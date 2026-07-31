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

// CacheControlFor devuelve el valor Cache-Control adecuado según la carpeta de destino.
//   - avatars/: la key es estable (avatars/{psiID}.webp) y se sobrescribe al re-subir,
//     así que el navegador debe revalidar con un max-age corto.
//   - posts/, titles/, certificates/: las keys llevan UUID (cambian por versión),
//     por lo que son inmutables y se pueden cachear por un año.
func CacheControlFor(folder string) string {
	if folder == "avatars" {
		return "public, max-age=86400"
	}
	return "public, max-age=31536000, immutable"
}

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
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.Must(uuid.NewV7()).String(), ext)

	// 3. Subir a S3 usando Streaming (bajo consumo de RAM)
	_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.Bucket),
		Key:          aws.String(filename),
		Body:         src,
		ContentType:  aws.String(fileHeader.Header.Get("Content-Type")),
		CacheControl: aws.String(CacheControlFor(folder)),
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
		Bucket:       aws.String(s.Bucket),
		Key:          aws.String(key),
		Body:         reader,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(CacheControlFor(folder)),
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

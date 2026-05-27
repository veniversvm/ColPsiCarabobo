// api/internal/service/post_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
// El PostService gestiona el ciclo de vida de las noticias y publicaciones,
// asegurando la integridad entre los metadatos y el contenido extenso.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// PostService encapsula las dependencias para la gestión de contenido.
// Utiliza bluemonday para prevenir ataques de Cross-Site Scripting (XSS).
type PostService struct {
	repo      domain.PostRepository
	s3Client  *s3.S3Client
	sanitizer *bluemonday.Policy
}

// NewPostService inicializa el servicio con una política de sanitización estándar (UGC).
func NewPostService(repo domain.PostRepository, s3 *s3.S3Client) *PostService {
	return &PostService{
		repo:      repo,
		s3Client:  s3,
		sanitizer: bluemonday.UGCPolicy(), // Política segura para contenido generado por usuario
	}
}

// =========================================================================
// CREACIÓN DE CONTENIDO
// =========================================================================

// CreatePost orquesta la creación de una noticia.
// Procesa imágenes, limpia el HTML y persiste metadatos y contenido en una sola transacción.
func (s *PostService) CreatePost(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreatePostRequest, file *multipart.FileHeader) error {
	if !admin.CanPublish && !admin.Sudo {
		return errors.New("no tienes permiso para publicar")
	}

	if req.Status == "scheduled" && req.PublishAt == nil {
		return errors.New("un post programado requiere publish_at")
	}

	// 1. Subir imagen a S3 (si existe + Sanitización)
	var s3Key string
	if file != nil {
		src, err := file.Open()
		if err != nil {
			return errors.New("error leyendo imagen")
		}
		defer src.Close()

		// RECIBIMOS 3 VALORES: bytes, extensión y mime-type
		// Re-codificación para eliminar scripts ocultos y metadatos sensibles (GPS/EXIF)
		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return fmt.Errorf("error de seguridad en imagen: %v", err)
		}

		// Generamos nombre único
		filename := uuid.Must(uuid.NewV7()).String() + ext

		// Pasamos el Content-Type correcto a S3 (antes estaba hardcodeado a image/jpeg)
		key, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "posts", filename, contentType)
		if err != nil {
			return err
		}
		s3Key = key
	}

	// 2. Sanitizar el HTML del contenido para evitar XSS
	cleanContent := s.sanitizer.Sanitize(req.Content)

	// 3. Preparar modelos
	textID := uuid.Must(uuid.NewV7())
	textModel := &domain.TextModel{
		AuditModel: domain.AuditModel{
			CreateBy: admin.Username, CreateById: &admin.ID,
			UpdateBy: admin.Username, UpdateById: &admin.ID,
		},
		Content: cleanContent,
	}

	postModel := &domain.Post{
		AuditModel: domain.AuditModel{
			CreateBy: admin.Username, CreateById: &admin.ID,
			UpdateBy: admin.Username, UpdateById: &admin.ID,
		},
		Title:            req.Title,
		ShortDescription: req.ShortDescription,
		Type:             req.Type,
		ImageS3Key:       s3Key,
		Status:           domain.PostStatus(req.Status),
		PublishAt:        req.PublishAt,
		TextID:           textID,
	}

	return s.repo.Create(ctx, postModel, textModel)
}

// =========================================================================
// CONSULTA Y VISIBILIDAD (ACL)
// =========================================================================

// GetPostsList implementa filtros de visibilidad dinámicos según el rol del usuario.
// - Admin: Ve todo (activos, inactivos, públicos y privados).
// - Psi: Ve activos de tipo 'public' y 'psi'.
// - Public: Solo ve activos de tipo 'public'.
func (s *PostService) GetPostsList(ctx context.Context, page, limit int, userRole string) (interface{}, error) {
	filter := domain.PostFilter{}

	switch userRole {
	case "admin":
		filter.Status = nil // ve todos los estados
	case "psi":
		filter.Status = []domain.PostStatus{domain.PostStatusPublished}
	default: // público
		filter.Status = []domain.PostStatus{domain.PostStatusPublished}
		filter.Type = "public"
	}

	posts, total, err := s.repo.List(ctx, filter, page, limit)
	if err != nil {
		return nil, err
	}

	// Opcional: Generar URLs firmadas de S3 para las imágenes aquí si son privadas

	return map[string]interface{}{
		"data":  posts,
		"total": total,
		"page":  page,
	}, nil
}

// GetPostByID recupera una noticia específica validando el nivel de acceso del solicitante.
func (s *PostService) GetPostByID(ctx context.Context, id uuid.UUID, userRole string) (*domain.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validación de lista de control de acceso (ACL)
	switch userRole {
	case "admin":
		// El admin puede ver todo (incluso borradores o privados)
		return post, nil

	case "psi":
		if post.Status != domain.PostStatusPublished {
			return nil, errors.New("post no disponible")
		}
		if post.Type != "public" && post.Type != "psi" {
			return nil, errors.New("acceso denegado")
		}

	default: // "public"
		// El público solo ve activos y de tipo "public"
		if post.Status != domain.PostStatusPublished || post.Type != "public" {
			return nil, errors.New("post no encontrado o privado")
		}
	}

	return post, nil
}

// =========================================================================
// ACTUALIZACIÓN Y LIMPIEZA
// =========================================================================

// UpdatePost permite la edición parcial (PATCH) y gestiona el reemplazo de archivos en S3.
// Implementa un mecanismo de "Rollback Manual" para S3: si la DB falla, borra la imagen nueva subida.
func (s *PostService) UpdatePost(ctx context.Context, admin *domain.UserAdmin, req request_structs.UpdatePostRequest, file *multipart.FileHeader, id uuid.UUID) error {
	// 1. Validar Permisos (Solo Admins pueden modificar)
	if !admin.CanUpdatePublish && !admin.Sudo {
		return errors.New("no tienes permiso para modificar publicaciones")
	}

	// 2. Obtener Post y su texto actual
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("publicación no encontrada")
	}

	// 3. Lógica de Auditoría
	post.UpdateBy = admin.Username
	post.UpdateById = &admin.ID

	// 4. PROCESAMIENTO DE IMAGEN (Si se envía una nueva)
	oldS3Key := post.ImageS3Key
	newS3Key := oldS3Key

	if file != nil {
		// A. Limpiar imagen (usando la utilidad que ya tenemos)
		src, err := file.Open()
		if err != nil {
			return errors.New("error leyendo imagen")
		}

		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return fmt.Errorf("error de seguridad en imagen: %v", err)
		}

		// B. Generar nombre único y subir
		filename := fmt.Sprintf("posts/updated_%s%s", uuid.Must(uuid.NewV7()).String(), ext)
		key, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "posts", filename, contentType)
		if err != nil {
			return err
		}

		newS3Key = key
	}

	// 5. APLICAR CAMBIOS PARCIALES (PATCH-LIKE)
	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.ShortDescription != nil {
		post.ShortDescription = *req.ShortDescription
	}
	if req.Type != nil {
		post.Type = *req.Type
	}
	if req.Status != nil {
		// Validar scheduled
		if *req.Status == "scheduled" && req.PublishAt == nil && post.PublishAt == nil {
			return errors.New("un post programado requiere publish_at")
		}
		post.Status = domain.PostStatus(*req.Status)
	}
	if req.PublishAt != nil {
		post.PublishAt = req.PublishAt
	}

	// Actualización de texto (si viene en el request)
	var textModel *domain.TextModel = nil
	if req.Content != nil {
		textModel = &domain.TextModel{
			ID:      post.TextID, // Se asume que siempre actualizamos el contenido existente
			Content: s.sanitizer.Sanitize(*req.Content),
			AuditModel: domain.AuditModel{
				UpdatedAt:  time.Now(),
				UpdateBy:   admin.Username,
				UpdateById: &admin.ID,
			},
		}
	}

	// 6. PERSISTENCIA Y LIMPIEZA DE S3
	if err := s.repo.Update(ctx, post, textModel); err != nil {
		// ROLLBACK S3: Si la base de datos falla, borramos la imagen que acabamos de subir
		if newS3Key != "" && newS3Key != oldS3Key {
			s.s3Client.DeleteFile(ctx, newS3Key)
		}
		return err
	}

	// 7. LIMPIEZA DE ARCHIVOS ANTIGUOS (Si se subió una nueva imagen)
	//Borramos la imagen antigua para liberar espacio
	if newS3Key != oldS3Key && oldS3Key != "" {
		s.s3Client.DeleteFile(ctx, oldS3Key)
	}

	return nil
}

// PublishScheduled busca posts programados cuya fecha ya llegó y los publica.
// Llamar desde un ticker en main.go cada minuto.
func (s *PostService) PublishScheduled(ctx context.Context) error {
	result := s.repo.PublishScheduled(ctx)
	if result > 0 {
		log.Printf("[CMS] %d post(s) programados publicados automáticamente", result)
	}
	return nil
}

func (s *PostService) GetSitemapData(ctx context.Context) (interface{}, error) {
	// Llamamos al repositorio de posts (noticias)
	return s.repo.GetSitemapPosts(ctx)
}

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
	"github.com/rs/zerolog/log"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// PostService encapsula las dependencias para la gestión del CMS.
// Coordina la base de datos (PostgreSQL), el almacenamiento de objetos (Amazon S3)
// y el motor de Sanitización HTML.
type PostService struct {
	repo      domain.PostRepository
	s3Client  *s3.S3Client
	sanitizer *bluemonday.Policy
}

// NewPostService inicializa el servicio con una política de sanitización estricta.
//
// Seguridad XSS: bluemonday.UGCPolicy() (User Generated Content) limpia el HTML,
// erradicando etiquetas <script>, eventos onMouseOver y estilos CSS peligrosos,
// garantizando que el contenido renderizado por el Frontend sea 100% seguro.
func NewPostService(repo domain.PostRepository, s3 *s3.S3Client) *PostService {
	return &PostService{
		repo:      repo,
		s3Client:  s3,
		sanitizer: bluemonday.UGCPolicy(),
	}
}

// =========================================================================
// CREACIÓN DE CONTENIDO
// =========================================================================

// CreatePost orquesta la publicación de una noticia.
// Procesa imágenes, limpia el HTML y persiste metadatos y contenido.
func (s *PostService) CreatePost(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreatePostRequest, file *multipart.FileHeader) error {
	// 1. Gatekeeping de Autorización
	if !admin.CanPublish && !admin.Sudo {
		return domain.ErrPostPermDenied
	}

	// 2. Validación de Estado (Máquina de Estados)
	if req.Status == "scheduled" && req.PublishAt == nil {
		return errors.New("un post programado requiere publish_at")
	}

	// 3. Subir imagen a S3 (Manejo Seguro de Archivos)
	var s3Key string
	if file != nil {
		src, err := file.Open()
		if err != nil {
			return errors.New("error leyendo imagen")
		}
		defer src.Close()

		// Defensa contra Esteganografía y Metadatos Sensibles:
		// La utilidad SanitizeDocument procesa los bytes crudos y elimina la metadata EXIF
		// (como ubicación GPS de la foto) que los psicólogos o admins puedan subir por error.
		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return fmt.Errorf("error de seguridad en imagen: %v", err)
		}

		// Prevenir Colisión de Nombres y Forzar Tipado MIME seguro
		filename := uuid.Must(uuid.NewV7()).String() + ext

		key, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "posts", filename, contentType)
		if err != nil {
			return err
		}
		s3Key = key
	}

	// 4. Sanitizar el HTML del contenido (Prevención Cross-Site Scripting)
	cleanContent := s.sanitizer.Sanitize(req.Content)

	// 5. Preparar modelos de Dominio y Auditoría
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

	// 6. Transacción Atómica
	return s.repo.Create(ctx, postModel, textModel)
}

// =========================================================================
// CONSULTA Y VISIBILIDAD (ACL - ACCESS CONTROL LIST)
// =========================================================================

// GetPostsList implementa filtros dinámicos de listado según la jerarquía del solicitante.
// En lugar de hacer IFs pesados en el código, muta el objeto `PostFilter` antes de enviarlo
// a la base de datos, delegando el filtrado al motor SQL para máximo rendimiento.
func (s *PostService) GetPostsList(ctx context.Context, page, limit int, userRole string) (interface{}, error) {
	filter := domain.PostFilter{}

	switch userRole {
	case "admin":
		// Visión de Rayos X: El admin ve borradores, programados y archivados.
		filter.Status = nil
	case "psi":
		// Acceso Gremial: El colegiado ve contenido publicado (tanto general como interno).
		filter.Status = []domain.PostStatus{domain.PostStatusPublished}
	default: // "public"
		// Acceso Restringido: El visitante anónimo solo ve contenido publicado y categorizado como público.
		filter.Status = []domain.PostStatus{domain.PostStatusPublished}
		filter.Type = "public"
	}

	posts, total, err := s.repo.List(ctx, filter, page, limit)
	if err != nil {
		return nil, err
	}

	// Convertir S3 keys a URLs públicas
	for i := range posts {
		s.resolvePostURLs(&posts[i])
	}

	return map[string]interface{}{
		"data":  posts,
		"total": total,
		"page":  page,
	}, nil
}

// GetPostByID recupera una noticia validando explícitamente el nivel de acceso (IDOR Protection).
// Evita que un atacante extraiga un borrador adivinando el UUID en la URL (/api/posts/:id).
func (s *PostService) GetPostByID(ctx context.Context, id uuid.UUID, userRole string) (*domain.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validación estricta de la Lista de Control de Acceso (ACL)
	switch userRole {
	case "admin":
		s.resolvePostURLs(post)
		return post, nil

	case "psi":
		if post.Status != domain.PostStatusPublished {
			return nil, errors.New("post no disponible")
		}
		if post.Type != "public" && post.Type != "psi" {
			return nil, errors.New("acceso denegado")
		}

	default: // "public"
		if post.Status != domain.PostStatusPublished || post.Type != "public" {
			return nil, errors.New("post no encontrado o privado")
		}
	}

	s.resolvePostURLs(post)
	return post, nil
}

// =========================================================================
// ACTUALIZACIÓN Y LIMPIEZA MANTENIDA (ORCHESTRATION)
// =========================================================================

// UpdatePost permite la edición parcial y gestiona el reemplazo de portadas.
//
// Arquitectura: Transacción Distribuida (Saga Pattern Lite).
// Actualizar la Base de Datos y un servicio de nube (S3) al mismo tiempo es peligroso
// porque no comparten la misma transacción. Si PostgreSQL falla pero S3 tuvo éxito,
// te quedas con archivos "zombie" huérfanos cobrando almacenamiento.
// Este método implementa un "Rollback de Compensación Manual" para S3.
func (s *PostService) UpdatePost(ctx context.Context, admin *domain.UserAdmin, req request_structs.UpdatePostRequest, file *multipart.FileHeader, id uuid.UUID) error {

	// 1. Validar Permisos
	if !admin.CanUpdatePublish && !admin.Sudo {
		return errors.New("no tienes permiso para modificar publicaciones")
	}

	// 2. Extraer estado original
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("publicación no encontrada")
	}

	// 3. Lógica de Auditoría Inmutable
	post.UpdateBy = admin.Username
	post.UpdateById = &admin.ID

	// 4. PROCESAMIENTO DE NUEVA IMAGEN (Si aplica)
	oldS3Key := post.ImageS3Key
	newS3Key := oldS3Key

	if file != nil {
		src, err := file.Open()
		if err != nil {
			return errors.New("error leyendo imagen")
		}

		cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)
		if err != nil {
			return fmt.Errorf("error de seguridad en imagen: %v", err)
		}

		// Generar nombre e insertar en S3
		filename := fmt.Sprintf("posts/updated_%s%s", uuid.Must(uuid.NewV7()).String(), ext)
		key, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "posts", filename, contentType)
		if err != nil {
			return err
		}

		newS3Key = key
	}

	// 5. APLICAR MUTACIONES (Semántica PATCH)
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
		if *req.Status == "scheduled" && req.PublishAt == nil && post.PublishAt == nil {
			return errors.New("un post programado requiere publish_at")
		}
		post.Status = domain.PostStatus(*req.Status)
	}
	if req.PublishAt != nil {
		post.PublishAt = req.PublishAt
	}

	// Si hay nuevo contenido de texto, se limpia (Sanitize) antes de preparar el modelo
	var textModel *domain.TextModel = nil
	if req.Content != nil {
		textModel = &domain.TextModel{
			ID:      post.TextID,
			Content: s.sanitizer.Sanitize(*req.Content),
			AuditModel: domain.AuditModel{
				UpdatedAt:  time.Now(),
				UpdateBy:   admin.Username,
				UpdateById: &admin.ID,
			},
		}
	}

	// 6. PERSISTENCIA Y COMPENSACIÓN (SAGA ROLLBACK)
	if err := s.repo.Update(ctx, post, textModel); err != nil {
		// ROLLBACK S3: La BD falló. Para evitar basura en el bucket, borramos
		// la imagen que acabamos de subir en el paso 4.
		if newS3Key != "" && newS3Key != oldS3Key {
			s.s3Client.DeleteFile(ctx, newS3Key)
		}
		return err
	}

	// 7. LIMPIEZA DE ESPACIO (Garbage Collection)
	// Si la actualización en la DB fue un éxito Y teníamos una imagen vieja,
	// la borramos del bucket para no pagar almacenamiento de portadas deprecadas.
	if newS3Key != oldS3Key && oldS3Key != "" {
		s.s3Client.DeleteFile(ctx, oldS3Key)
	}

	return nil
}

// PublishScheduled es un "Worker Job".
// Ejecutado periódicamente por un orquestador (cron/ticker), busca en la base de datos
// las publicaciones que estaban en sala de espera y las hace públicas si su fecha (`PublishAt`)
// ya se cumplió.
func (s *PostService) PublishScheduled(ctx context.Context) error {
	result := s.repo.PublishScheduled(ctx)
	if result > 0 {
		log.Info().Int64("count", result).Str("component", "post_service").Msg("posts programados publicados automáticamente")
	}
	return nil
}

// GetSitemapData proyecta el subconjunto de datos necesarios para la indexación SEO
// de los motores de búsqueda (Googlebot), abstrayendo la complejidad de la tabla Post.
func (s *PostService) GetSitemapData(ctx context.Context) (interface{}, error) {
	return s.repo.GetSitemapPosts(ctx)
}

// resolvePostURLs convierte la S3 key de un Post en URL pública completa.
func (s *PostService) resolvePostURLs(post *domain.Post) {
	if s.s3Client == nil || post == nil {
		return
	}
	post.ImageS3Key = s.s3Client.GetPublicURL(post.ImageS3Key)
}

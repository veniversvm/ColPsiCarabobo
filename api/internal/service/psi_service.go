package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// PsiService provides business operations for psychologist users including authentication, profile management, and academic data.
type PsiService struct {
	repo        domain.PsiUserRepository
	s3Client    *s3.S3Client
	mailService IMailService
	sanitizer   *bluemonday.Policy
	absSvc      *AudiobookshelfService
}

// NewPsiService creates a new PsiService with the given repository, S3 client, and mail service.
func NewPsiService(repo domain.PsiUserRepository, s3Client *s3.S3Client, mailService IMailService) *PsiService {
	policy := bluemonday.UGCPolicy()
	policy.AllowStyles("text-align").OnElements("p", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "li", "blockquote")

	return &PsiService{
		repo:        repo,
		s3Client:    s3Client,
		mailService: mailService,
		sanitizer:   policy,
	}
}

// ResolvePsiModelURLs convierte las S3 keys internas de un PsiUserModel en URLs públicas completas.
// Se invoca antes de serializar a JSON para que el frontend reciba URLs listas para usar.
func (s *PsiService) ResolvePsiModelURLs(psi *domain.PsiUserModel) {
	if s.s3Client == nil || psi == nil {
		return
	}
	psi.ProfilePictureS3Key = s.avatarURL(psi.ProfilePictureS3Key, psi.UpdatedAt)
	if psi.ColData.PsiUserModelID != uuid.Nil {
		psi.ColData.TitleImageOneS3Key = s.s3Client.GetPublicURL(psi.ColData.TitleImageOneS3Key)
		psi.ColData.TitleImageTwoS3Key = s.s3Client.GetPublicURL(psi.ColData.TitleImageTwoS3Key)
		psi.ColData.TitleImageThreeS3Key = s.s3Client.GetPublicURL(psi.ColData.TitleImageThreeS3Key)
	}
	for i := range psi.PostGrades {
		psi.PostGrades[i].PicOneS3Key = s.s3Client.GetPublicURL(psi.PostGrades[i].PicOneS3Key)
		psi.PostGrades[i].PicTwoS3Key = s.s3Client.GetPublicURL(psi.PostGrades[i].PicTwoS3Key)
		psi.PostGrades[i].PicThreeS3Key = s.s3Client.GetPublicURL(psi.PostGrades[i].PicThreeS3Key)
	}
}

// publicURL es un wrapper nil-safe para s3Client.GetPublicURL.
// Si el s3Client no está inicializado (tests, modo degradado), retorna el key original.
func (s *PsiService) publicURL(key string) string {
	if s.s3Client == nil {
		return key
	}
	return s.s3Client.GetPublicURL(key)
}

// avatarURL resuelve la URL pública del avatar con cache-busting (?v=updated_at).
// Al actualizarse la foto de perfil, updated_at cambia → la URL cambia → el navegador
// descarta la copia cacheada de la imagen anterior automáticamente.
func (s *PsiService) avatarURL(key string, updatedAt time.Time) string {
	if key == "" {
		return ""
	}
	url := s.publicURL(key)
	if !updatedAt.IsZero() {
		url = fmt.Sprintf("%s?v=%d", url, updatedAt.Unix())
	}
	return url
}

// SetAudiobookshelf inyecta el servicio de Audiobookshelf (nil-safe). Permite
// el aprovisionamiento de cuentas ABS sin cambiar la firma de NewPsiService.
func (s *PsiService) SetAudiobookshelf(absSvc *AudiobookshelfService) {
	s.absSvc = absSvc
}

// SetAudiobookshelfID persiste el id de la cuenta ABS del agremiado en el
// expediente (campo AudioBookShellId). Es una actualización ligera que no toca
// el resto del perfil. Si el id no cambió, no hace nada.
func (s *PsiService) SetAudiobookshelfID(ctx context.Context, psi *domain.PsiUserModel, absID string) error {
	if psi == nil || absID == "" || absID == psi.AudioBookShellId {
		return nil
	}
	psi.AudioBookShellId = absID
	return s.repo.UpdateAudioBookShellID(ctx, psi)
}

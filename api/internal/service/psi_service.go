package service

import (
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
	psi.ProfilePictureS3Key = s.s3Client.GetPublicURL(psi.ProfilePictureS3Key)
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

package service

import (
	"context"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// MOCKS (SIMULADORES)
// =========================================================================

type mockPostRepo struct {
	domain.PostRepository
	CreateFunc           func(ctx context.Context, p *domain.Post, t *domain.TextModel) error
	GetByIDFunc          func(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	ListFunc             func(ctx context.Context, f domain.PostFilter, page, limit int) ([]domain.Post, int64, error)
	UpdateFunc           func(ctx context.Context, p *domain.Post, t *domain.TextModel) error
	PublishScheduledFunc func(ctx context.Context) int64
}

func (m *mockPostRepo) Create(ctx context.Context, p *domain.Post, t *domain.TextModel) error {
	return m.CreateFunc(ctx, p, t)
}
func (m *mockPostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockPostRepo) List(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
	return m.ListFunc(ctx, f, p, l)
}
func (m *mockPostRepo) Update(ctx context.Context, p *domain.Post, t *domain.TextModel) error {
	return m.UpdateFunc(ctx, p, t)
}

type mockS3Client struct {
	DeleteCalledWith string
}

func (m *mockS3Client) UploadStream(ctx context.Context, r io.Reader, b, f, c string) (string, error) {
	return "posts/new_image.jpg", nil
}
func (m *mockS3Client) DeleteFile(ctx context.Context, key string) error {
	m.DeleteCalledWith = key
	return nil
}

// =========================================================================
// SUITE DE TESTS COMPLETA
// =========================================================================

func TestPostService_Extensive(t *testing.T) {
	repo := &mockPostRepo{}
	// NewPostService inicializa el sanitizer internamente
	svc := NewPostService(repo, nil)
	ctx := context.Background()

	// --- 1. TEST DE REGLAS DE ACCESO (ACL) ---
	t.Run("ACL_Visibility_Rules", func(t *testing.T) {
		postID := uuid.Must(uuid.NewV7())
		mockPost := &domain.Post{
			ID:     postID,
			Type:   "psi",
			Status: domain.PostStatus("draft"), // Post en borrador
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
			return mockPost, nil
		}

		// Caso Admin: Debe ver el post aunque sea draft (según línea 158 de tu svc)
		p, err := svc.GetPostByID(ctx, postID, "admin")
		if err != nil || p == nil {
			t.Error("Admin debería ver el post independientemente de su estado")
		}

		// Caso Public: No debe ver posts que no sean Status: published y Type: public
		_, err = svc.GetPostByID(ctx, postID, "public")
		if err == nil {
			t.Error("Público NO debería ver posts en borrador o privados")
		}
	})

	// --- 2. TEST DE SANITIZACIÓN (XSS) ---
	t.Run("XSS_Prevention", func(t *testing.T) {
		admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Username: "admin", CanPublish: true}

		// Inyectamos un script malicioso
		maliciousHTML := "<p>Hola</p><script>alert('hack')</script>"

		repo.CreateFunc = func(ctx context.Context, p *domain.Post, textModel *domain.TextModel) error {
			// El sanitizer del servicio debería haber eliminado el tag <script>
			if textModel.Content != "<p>Hola</p>" {
				t.Errorf("Sanitización falló, contenido: %s", textModel.Content)
			}
			return nil
		}

		req := request_structs.CreatePostRequest{
			Title: "Test", Content: maliciousHTML, Type: "public", Status: "published",
		}

		_ = svc.CreatePost(ctx, admin, req, nil)
	})

	// --- 3. TEST DE FILTROS DE LISTADO (ACL Logic) ---
	t.Run("List_Filters_By_Role", func(t *testing.T) {
		// Escenario Public: El servicio debe filtrar por Type="public" y Status="published"
		repo.ListFunc = func(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
			if f.Type != "public" {
				t.Errorf("Filtro incorrecto para público, se esperaba 'public' pero se obtuvo: %s", f.Type)
			}
			if len(f.Status) == 0 || f.Status[0] != domain.PostStatusPublished {
				t.Error("Filtro para público debería exigir estado 'published'")
			}
			return []domain.Post{}, 0, nil
		}
		_, _ = svc.GetPostsList(ctx, 1, 10, "public")

		// Escenario Psi: El servicio filtra por Status="published" pero NO filtra por Type (ve public y psi)
		repo.ListFunc = func(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
			if f.Type != "" {
				t.Errorf("Filtro de Type para PSI debería ser vacío para ver ambos tipos, se obtuvo: %s", f.Type)
			}
			if len(f.Status) == 0 || f.Status[0] != domain.PostStatusPublished {
				t.Error("Filtro para PSI debería exigir estado 'published'")
			}
			return []domain.Post{}, 0, nil
		}
		_, _ = svc.GetPostsList(ctx, 1, 10, "psi")

		// Escenario Admin: No aplica filtros de estado (Status = nil)
		repo.ListFunc = func(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
			if f.Status != nil {
				t.Error("Filtro para Admin debería ser nil en Status para ver borradores")
			}
			return []domain.Post{}, 0, nil
		}
		_, _ = svc.GetPostsList(ctx, 1, 10, "admin")
	})

	// --- 4. TEST DE PERMISOS ADMINISTRATIVOS ---
	t.Run("Admin_Permissions_Check", func(t *testing.T) {
		// Admin que NO tiene permiso de publicar (ni es Sudo)
		limitedAdmin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), CanPublish: false, Sudo: false}

		req := request_structs.CreatePostRequest{Title: "Intento Fallido"}
		err := svc.CreatePost(ctx, limitedAdmin, req, nil)

		if err == nil || err.Error() != "no tienes permiso para publicar" {
			t.Error("Se debió denegar la creación al admin sin permisos")
		}
	})
}

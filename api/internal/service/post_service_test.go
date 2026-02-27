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

// Mock del Repositorio usando Embedding para no implementar métodos innecesarios
type mockPostRepo struct {
	domain.PostRepository
	CreateFunc  func(ctx context.Context, p *domain.Post, t *domain.TextModel) error
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	ListFunc    func(ctx context.Context, f domain.PostFilter, page, limit int) ([]domain.Post, int64, error)
	UpdateFunc  func(ctx context.Context, p *domain.Post, t *domain.TextModel) error
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

// Mock de S3 (Simulamos el comportamiento de Amazon S3)
type mockS3Client struct {
	// Aquí no usamos embedding porque S3Client suele ser un struct,
	// pero simulamos las funciones que el Service llama.
	DeleteCalledWith string
	UploadStreamFunc func(ctx context.Context, reader io.Reader, folder, filename, contentType string) (string, error)
	DeleteFileFunc   func(ctx context.Context, key string) error
}

func (m *mockS3Client) UploadStream(ctx context.Context, r any, b, f, c string) (string, error) {
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
	// Nota: Si NewPostService pide *s3.S3Client (struct), podrías necesitar
	// pasar nil o cambiarlo a interfaz como te sugerí arriba.
	// Por ahora asumiremos que svc usa estas dependencias:
	svc := NewPostService(repo, nil)
	ctx := context.Background()

	// --- 1. TEST DE REGLAS DE ACCESO (ACL) ---
	t.Run("ACL_Visibility_Rules", func(t *testing.T) {
		postID := uuid.New()
		mockPost := &domain.Post{
			ID:       postID,
			Type:     "psi", // Solo para psicólogos
			IsActive: false, // Inactivo (borrador)
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
			return mockPost, nil
		}

		// Caso Admin: Debe ver TODO
		p, err := svc.GetPostByID(ctx, postID, "admin")
		if err != nil || p == nil {
			t.Error("Admin debería ver el post inactivo")
		}

		// Caso Public: No debe ver posts de tipo 'psi' o inactivos
		_, err = svc.GetPostByID(ctx, postID, "public")
		if err == nil {
			t.Error("Público NO debería ver posts privados/inactivos")
		}
	})

	// --- 2. TEST DE SANITIZACIÓN (XSS) ---
	t.Run("XSS_Prevention", func(t *testing.T) {
		admin := &domain.UserAdmin{ID: uuid.New(), Username: "admin", CanPublish: true}

		// Inyectamos un script malicioso
		maliciousHTML := "<p>Hola</p><script>alert('hack')</script>"

		repo.CreateFunc = func(ctx context.Context, p *domain.Post, textModel *domain.TextModel) error {
			// Ahora 't' se refiere correctamente a *testing.T del Test principal
			if textModel.Content != "<p>Hola</p>" {
				t.Errorf("Sanitización falló, contenido: %s", textModel.Content)
			}
			return nil
		}

		req := request_structs.CreatePostRequest{
			Title: "Test", Content: maliciousHTML, Type: "public",
		}

		_ = svc.CreatePost(ctx, admin, req, nil)
	})

	// --- 3. TEST DE FILTROS DE LISTADO ---
	t.Run("List_Filters_By_Role", func(t *testing.T) {
		// Probamos que para un usuario normal el filtro sea "public"
		repo.ListFunc = func(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
			if f.Type != "public" {
				t.Errorf("Filtro incorrecto para público, se obtuvo: %s", f.Type)
			}
			return []domain.Post{}, 0, nil
		}

		_, _ = svc.GetPostsList(ctx, 1, 10, "public")

		// Probamos que para un Psicólogo el filtro sea "all_visible"
		repo.ListFunc = func(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
			if f.Type != "all_visible" {
				t.Errorf("Filtro incorrecto para PSI, se obtuvo: %s", f.Type)
			}
			return []domain.Post{}, 0, nil
		}

		_, _ = svc.GetPostsList(ctx, 1, 10, "psi")
	})

	// --- 4. TEST DE PERMISOS ADMINISTRATIVOS ---
	t.Run("Admin_Permissions_Check", func(t *testing.T) {
		// Admin que NO tiene permiso de publicar
		limitedAdmin := &domain.UserAdmin{ID: uuid.New(), CanPublish: false, Sudo: false}

		req := request_structs.CreatePostRequest{Title: "Intento"}
		err := svc.CreatePost(ctx, limitedAdmin, req, nil)

		if err == nil || err.Error() != "no tienes permiso para publicar" {
			t.Error("Se debió bloquear la creación por falta de permisos")
		}
	})
}

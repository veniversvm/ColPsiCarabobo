package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// MOCKS (SIMULADORES DE INFRAESTRUCTURA)
// =========================================================================
// Patrón de Mocks Funcionales (Func Override):
// Permite aislar completamente la capa de Lógica de Negocio (Service) de la capa
// de Persistencia (PostgreSQL) y de la nube (AWS S3). Al utilizar funciones variables,
// cada sub-test puede programar su propio comportamiento (ej. simular un error de DB)
// sin requerir frameworks de mocking pesados ni contaminar el estado global.

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

// mockS3Client simula la interacción con Amazon S3 / MinIO.
// Evita conexiones de red reales durante la ejecución de los tests unitarios.
type mockS3Client struct {
	DeleteCalledWith string
}

func (m *mockS3Client) UploadStream(ctx context.Context, r io.Reader, b, f, c string) (string, error) {
	return "posts/new_image.jpg", nil // Simula una carga exitosa devolviendo una ruta ficticia
}
func (m *mockS3Client) DeleteFile(ctx context.Context, key string) error {
	m.DeleteCalledWith = key // Rastrea qué archivo se intentó borrar para aserciones posteriores
	return nil
}

// =========================================================================
// SUITE DE TESTS COMPLETA: SISTEMA DE GESTIÓN DE CONTENIDOS (CMS)
// =========================================================================

// TestPostService_Extensive somete a prueba las reglas de negocio del CMS.
// Evalúa críticamente el Control de Acceso (ACL), la higienización de entradas (Sanitización XSS)
// y los bloqueos jerárquicos de administración.
func TestPostService_Extensive(t *testing.T) {
	repo := &mockPostRepo{}
	// Inyección de Dependencias:
	// NewPostService inicializa internamente el motor de sanitización HTML (Bluemonday).
	// Usamos nil para el cliente S3 en los tests donde no se evalúa multimedia.
	svc := NewPostService(repo, nil)
	ctx := context.Background()

	// --- 1. TEST DE REGLAS DE ACCESO (ACL) ---
	// Vector Mitigado: Prevención de Fuga de Datos e IDOR.
	// Garantiza que, aunque un usuario público adivine el UUID de un post en estado "borrador"
	// o de uso exclusivo gremial ("psi"), la capa de servicios bloquee la entrega del payload.
	t.Run("ACL_Visibility_Rules", func(t *testing.T) {
		postID := uuid.Must(uuid.NewV7())
		mockPost := &domain.Post{
			ID:     postID,
			Type:   "psi",
			Status: domain.PostStatus("draft"), // Post en estado de revisión/borrador
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
			return mockPost, nil
		}

		// Caso Admin (Privilegio Absoluto): Debe ver el post aunque sea draft para poder editarlo.
		p, err := svc.GetPostByID(ctx, postID, "admin")
		if err != nil || p == nil {
			t.Error("Admin debería ver el post independientemente de su estado")
		}

		// Caso Public (Restricción Estricta): Se rechaza la lectura de contenido interno o no publicado.
		_, err = svc.GetPostByID(ctx, postID, "public")
		if err == nil {
			t.Error("Público NO debería ver posts en borrador o privados")
		}
	})

	// --- 2. TEST DE SANITIZACIÓN (XSS - CROSS-SITE SCRIPTING) ---
	// Vector Mitigado: Defensa en Profundidad contra inyección de código.
	// Dado que el CMS permite subir HTML enriquecido (Wysiwyg), un administrador comprometido
	// o malicioso podría intentar inyectar Javascript. El servicio debe limpiar (sanitizar)
	// el payload *antes* de que toque la base de datos.
	t.Run("XSS_Prevention", func(t *testing.T) {
		admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "admin"}, CanPublish: true}

		// Payload hostil: Incluye un tag script ejecutable
		maliciousHTML := "<p>Hola</p><script>alert('hack')</script>"

		repo.CreateFunc = func(ctx context.Context, p *domain.Post, textModel *domain.TextModel) error {
			// Aserción de Seguridad: El sanitizer interno (Bluemonday) debió erradicar el tag <script>
			// dejando únicamente el HTML seguro (<p>).
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

	// --- 3. TEST DE FILTROS DE LISTADO DINÁMICO (ACL Logic) ---
	// Evalúa el motor de mutación de consultas. El servicio debe forzar filtros seguros (WHERE clauses)
	// basados en el Rol del usuario (Contexto) que realiza la petición HTTP.
	t.Run("List_Filters_By_Role", func(t *testing.T) {

		// Escenario Public: Visitantes anónimos de internet.
		// Forzamiento estricto: Solo pueden ver posts Type="public" y Status="published".
		repo.ListFunc = func(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
			if f.Type != "public" {
				t.Errorf("Filtro incorrecto para público, se esperaba 'public' pero se obtuvo: %s", f.Type)
			}
			if len(f.Status) == 0 || f.Status[0] != domain.PostStatusPublished {
				t.Error("Filtro para público debería exigir estado 'published'")
			}
			return []domain.Post{}, 0, nil
		}
		_, _ = svc.GetPostsList(ctx, 1, 10, nil, "public")

		// Escenario Psi: Psicólogos Colegiados (Autenticados).
		// Forzamiento híbrido: Ven posts "published", pero el filtro de Tipo se vacía
		// para que la base de datos devuelva tanto noticias 'public' como avisos internos 'psi'.
		repo.ListFunc = func(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
			if f.Type != "" {
				t.Errorf("Filtro de Type para PSI debería ser vacío para ver ambos tipos, se obtuvo: %s", f.Type)
			}
			if len(f.Status) == 0 || f.Status[0] != domain.PostStatusPublished {
				t.Error("Filtro para PSI debería exigir estado 'published'")
			}
			return []domain.Post{}, 0, nil
		}
		_, _ = svc.GetPostsList(ctx, 1, 10, nil, "psi")

		// Escenario Admin: Personal de Staff.
		// Ausencia de filtros: Pueden ver todo el espectro de contenido (incluyendo 'draft' y 'archived').
		repo.ListFunc = func(ctx context.Context, f domain.PostFilter, p, l int) ([]domain.Post, int64, error) {
			if f.Status != nil {
				t.Error("Filtro para Admin debería ser nil en Status para ver borradores")
			}
			return []domain.Post{}, 0, nil
		}
		_, _ = svc.GetPostsList(ctx, 1, 10, nil, "admin")
	})

	// --- 4. TEST DE PERMISOS ADMINISTRATIVOS ---
	// Vector Mitigado: Control de Acceso Inadecuado.
	// Verifica el Principio de Menor Privilegio (PoLP). Que alguien sea administrador
	// no significa que tenga permiso para inyectar contenido en el portal público.
	t.Run("Admin_Permissions_Check", func(t *testing.T) {
		// Admin con bandera CanPublish deliberadamente apagada
		limitedAdmin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), CanPublish: false, Sudo: false}

		req := request_structs.CreatePostRequest{Title: "Intento Fallido"}
		err := svc.CreatePost(ctx, limitedAdmin, req, nil)

		// Aserción de Bloqueo Jerárquico
		if err == nil || !errors.Is(err, domain.ErrPostPermDenied) {
			t.Error("Se debió denegar la creación al admin sin permisos")
		}
	})

	// --- 5. TEST DE EDICIÓN DE UNA NOTICIA SIN FILA DE TEXTO ---
	// Las noticias publicadas antes de que existiera el TextModel quedaron con
	// posts.text_id = NULL, que GORM carga como uuid.Nil. El servicio debe pasar el
	// contenido con ese ID en cero (para que el repo cree y enlace la fila) y
	// rellenar la autoría de alta, en vez de intentar actualizar una fila inexistente.
	t.Run("Update_Post_Sin_TextID", func(t *testing.T) {
		admin := &domain.UserAdmin{
			ID:               uuid.Must(uuid.NewV7()),
			Credentials:      domain.Credentials{Username: "admin"},
			CanUpdatePublish: true,
		}
		orphanPost := &domain.Post{ID: uuid.Must(uuid.NewV7()), Title: "Noticia sin cuerpo", Status: domain.PostStatusPublished}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
			return orphanPost, nil
		}

		var gotText *domain.TextModel
		repo.UpdateFunc = func(ctx context.Context, p *domain.Post, textModel *domain.TextModel) error {
			gotText = textModel
			return nil
		}

		body := "<p>Cuerpo nuevo</p>"
		req := request_structs.UpdatePostRequest{Content: &body}
		if err := svc.UpdatePost(ctx, admin, req, nil, orphanPost.ID); err != nil {
			t.Fatalf("UpdatePost no debe fallar para un post sin text_id: %v", err)
		}

		if gotText == nil {
			t.Fatal("Se esperaba un TextModel con el nuevo contenido")
		}
		if gotText.ID != uuid.Nil {
			t.Errorf("El ID debe llegar en cero para que el repo cree la fila, llegó: %v", gotText.ID)
		}
		if gotText.Content != body {
			t.Errorf("El contenido no se persistió: %q", gotText.Content)
		}
		if gotText.CreateBy != "admin" || gotText.CreateById == nil || *gotText.CreateById != admin.ID {
			t.Error("La fila de texto creada debe quedar con la autoría de quien la generó")
		}
	})

	// --- 6. TEST DE EDICIÓN DE UNA NOTICIA QUE YA TIENE TEXTO ---
	// Con TextID real solo se cambia el contenido: no se toca la autoría de alta.
	t.Run("Update_Post_Con_TextID", func(t *testing.T) {
		admin := &domain.UserAdmin{
			ID:               uuid.Must(uuid.NewV7()),
			Credentials:      domain.Credentials{Username: "admin"},
			CanUpdatePublish: true,
		}
		textID := uuid.Must(uuid.NewV7())
		existing := &domain.Post{ID: uuid.Must(uuid.NewV7()), Title: "Noticia con cuerpo", TextID: textID, Status: domain.PostStatusPublished}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
			return existing, nil
		}

		var gotText *domain.TextModel
		repo.UpdateFunc = func(ctx context.Context, p *domain.Post, textModel *domain.TextModel) error {
			gotText = textModel
			return nil
		}

		body := "<p>Editado</p>"
		req := request_structs.UpdatePostRequest{Content: &body}
		if err := svc.UpdatePost(ctx, admin, req, nil, existing.ID); err != nil {
			t.Fatalf("UpdatePost falló: %v", err)
		}

		if gotText == nil || gotText.ID != textID {
			t.Fatal("Debe conservar el ID de la fila de texto existente")
		}
		if gotText.CreateBy != "" {
			t.Error("No se debe rellenar la autoría de alta al editar un texto existente")
		}
	})

	// --- 7. TEST DE PATCH PARCIAL (solo metadata, sin contenido) ---
	// Si el admin no manda content, no se toca la fila de texto en absoluto.
	t.Run("Update_Post_Solo_Metadata", func(t *testing.T) {
		admin := &domain.UserAdmin{
			ID:               uuid.Must(uuid.NewV7()),
			Credentials:      domain.Credentials{Username: "admin"},
			CanUpdatePublish: true,
		}
		existing := &domain.Post{ID: uuid.Must(uuid.NewV7()), Title: "Titulo viejo", TextID: uuid.Must(uuid.NewV7()), Status: domain.PostStatusPublished}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
			return existing, nil
		}

		repo.UpdateFunc = func(ctx context.Context, p *domain.Post, textModel *domain.TextModel) error {
			if textModel != nil {
				t.Error("Sin content en el request no debe tocarse el texto")
			}
			return nil
		}

		nuevo := "Titulo nuevo"
		req := request_structs.UpdatePostRequest{Title: &nuevo}
		if err := svc.UpdatePost(ctx, admin, req, nil, existing.ID); err != nil {
			t.Fatalf("UpdatePost falló: %v", err)
		}
	})
}

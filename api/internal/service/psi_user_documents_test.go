package service

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// mockPsiRepoDocumentos es un mock del repositorio acotado al submódulo de
// Registro Digital de Documentos. Usa el patrón "Func Override" del repo.
type mockPsiRepoDocumentos struct {
	domain.PsiUserRepository
	GetByIDFunc          func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	CreateDocumentFunc   func(ctx context.Context, doc *domain.PsiUserDocument) error
	ListDocumentsFunc    func(ctx context.Context, psiID uuid.UUID) ([]domain.PsiUserDocument, error)
	GetDocumentFunc      func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error)
	UpdateDocumentFunc   func(ctx context.Context, doc *domain.PsiUserDocument) error
	DeleteDocumentFunc   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPsiRepoDocumentos) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	if m.GetByIDFunc == nil {
		return nil, domain.ErrPsiNotFound
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *mockPsiRepoDocumentos) CreateDocument(ctx context.Context, doc *domain.PsiUserDocument) error {
	if m.CreateDocumentFunc == nil {
		return nil
	}
	return m.CreateDocumentFunc(ctx, doc)
}

func (m *mockPsiRepoDocumentos) ListDocuments(ctx context.Context, psiID uuid.UUID) ([]domain.PsiUserDocument, error) {
	if m.ListDocumentsFunc == nil {
		return nil, nil
	}
	return m.ListDocumentsFunc(ctx, psiID)
}

func (m *mockPsiRepoDocumentos) GetDocument(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
	if m.GetDocumentFunc == nil {
		return nil, domain.ErrDocumentNotFound
	}
	return m.GetDocumentFunc(ctx, id)
}

func (m *mockPsiRepoDocumentos) UpdateDocument(ctx context.Context, doc *domain.PsiUserDocument) error {
	if m.UpdateDocumentFunc == nil {
		return nil
	}
	return m.UpdateDocumentFunc(ctx, doc)
}

func (m *mockPsiRepoDocumentos) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	if m.DeleteDocumentFunc == nil {
		return nil
	}
	return m.DeleteDocumentFunc(ctx, id)
}

func TestPsiService_Documentos(t *testing.T) {
	ctx := context.Background()
	psiID := uuid.Must(uuid.NewV7())
	docID := uuid.Must(uuid.NewV7())

	adminSudo := &domain.UserAdmin{
		ID:          uuid.Must(uuid.NewV7()),
		Credentials: domain.Credentials{Username: "admin_docs"},
		Sudo:        true,
	}
	adminLector := &domain.UserAdmin{
		ID:          uuid.Must(uuid.NewV7()),
		Credentials: domain.Credentials{Username: "lector"},
	}

	psi := &domain.PsiUserModel{ID: psiID}

	// ── AddDocumentByAdmin ──────────────────────────────────────────────

	t.Run("AddDocumentByAdmin: permiso insuficiente", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		_, err := svc.AddDocumentByAdmin(ctx, adminLector, psiID, request_structs.CreatePsiUserDocumentRequest{}, nil)
		if !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, got %v", err)
		}
	})

	t.Run("AddDocumentByAdmin: psicólogo inexistente", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return nil, domain.ErrPsiNotFound
		}

		_, err := svc.AddDocumentByAdmin(ctx, adminSudo, psiID, request_structs.CreatePsiUserDocumentRequest{}, nil)
		if !errors.Is(err, domain.ErrPsiNotFound) {
			t.Errorf("Se esperaba ErrPsiNotFound, got %v", err)
		}
	})

	t.Run("AddDocumentByAdmin: archivo obligatorio", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		_, err := svc.AddDocumentByAdmin(ctx, adminSudo, psiID, request_structs.CreatePsiUserDocumentRequest{Title: "Cédula"}, nil)
		if !errors.Is(err, request_structs.ErrDocumentInvalidRequest) {
			t.Errorf("Se esperaba ErrDocumentInvalidRequest, got %v", err)
		}
	})

	t.Run("AddDocumentByAdmin: título obligatorio", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		_, err := svc.AddDocumentByAdmin(ctx, adminSudo, psiID, request_structs.CreatePsiUserDocumentRequest{DocumentType: request_structs.DocTypeCedula}, newTestFile())
		if !errors.Is(err, request_structs.ErrDocumentInvalidTitle) {
			t.Errorf("Se esperaba ErrDocumentInvalidTitle, got %v", err)
		}
	})

	t.Run("AddDocumentByAdmin: tipo inválido", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		_, err := svc.AddDocumentByAdmin(ctx, adminSudo, psiID, request_structs.CreatePsiUserDocumentRequest{Title: "Doc", DocumentType: "peligroso"}, newTestFile())
		if !errors.Is(err, request_structs.ErrDocumentInvalidType) {
			t.Errorf("Se esperaba ErrDocumentInvalidType, got %v", err)
		}
	})

	// ── ListDocumentsByAdmin ────────────────────────────────────────────

	t.Run("ListDocumentsByAdmin: éxito", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}
		repo.ListDocumentsFunc = func(ctx context.Context, id uuid.UUID) ([]domain.PsiUserDocument, error) {
			return []domain.PsiUserDocument{
				{ID: docID, PsiUserID: psiID, Title: "Título", S3Key: "documents/key.webp"},
			}, nil
		}

		docs, err := svc.ListDocumentsByAdmin(ctx, adminSudo, psiID)
		if err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
		if len(docs) != 1 || docs[0].ID != docID {
			t.Fatalf("docs = %+v, want 1 con ID=%v", docs, docID)
		}
		if docs[0].S3Key != "documents/key.webp" {
			t.Errorf("S3Key = %q, want passthrough (s3Client nil)", docs[0].S3Key)
		}
	})

	t.Run("ListDocumentsByAdmin: permiso insuficiente", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		_, err := svc.ListDocumentsByAdmin(ctx, adminLector, psiID)
		if !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, got %v", err)
		}
	})

	// ── GetMyDocuments (solo lectura del psicólogo) ─────────────────────

	t.Run("GetMyDocuments: éxito", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.ListDocumentsFunc = func(ctx context.Context, id uuid.UUID) ([]domain.PsiUserDocument, error) {
			return []domain.PsiUserDocument{
				{ID: docID, PsiUserID: psiID, Title: "CI", S3Key: "documents/ci.webp"},
			}, nil
		}

		docs, err := svc.GetMyDocuments(ctx, psi)
		if err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
		if len(docs) != 1 || docs[0].Title != "CI" {
			t.Errorf("docs = %+v, want 1 con Title=CI", docs)
		}
	})

	// ── UpdateDocumentByAdmin ───────────────────────────────────────────

	t.Run("UpdateDocumentByAdmin: éxito metadatos", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDocumentFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
			return &domain.PsiUserDocument{ID: docID, PsiUserID: psiID, Title: "Antes"}, nil
		}
		var saved *domain.PsiUserDocument
		repo.UpdateDocumentFunc = func(ctx context.Context, doc *domain.PsiUserDocument) error {
			saved = doc
			return nil
		}

		newTitle := "  Cédula V-123456  "
		req := request_structs.UpdatePsiUserDocumentRequest{Title: &newTitle}
		if _, err := svc.UpdateDocumentByAdmin(ctx, adminSudo, docID, req, nil); err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
		if saved.Title != "Cédula V-123456" {
			t.Errorf("Title = %q, want trim aplicado", saved.Title)
		}
		if saved.UpdateBy != adminSudo.Username {
			t.Errorf("UpdateBy = %q, want %q", saved.UpdateBy, adminSudo.Username)
		}
	})

	t.Run("UpdateDocumentByAdmin: vaciar fecha (ClearDocumentDate)", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDocumentFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
			date := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
			return &domain.PsiUserDocument{ID: docID, PsiUserID: psiID, Title: "Doc", DocumentDate: &date}, nil
		}
		var saved *domain.PsiUserDocument
		repo.UpdateDocumentFunc = func(ctx context.Context, doc *domain.PsiUserDocument) error {
			saved = doc
			return nil
		}

		req := request_structs.UpdatePsiUserDocumentRequest{ClearDocumentDate: true}
		if _, err := svc.UpdateDocumentByAdmin(ctx, adminSudo, docID, req, nil); err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
		if saved.DocumentDate != nil {
			t.Error("DocumentDate debería haberse vaciado con ClearDocumentDate")
		}
	})

	t.Run("UpdateDocumentByAdmin: documento inexistente", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDocumentFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
			return nil, domain.ErrDocumentNotFound
		}

		_, err := svc.UpdateDocumentByAdmin(ctx, adminSudo, docID, request_structs.UpdatePsiUserDocumentRequest{}, nil)
		if !errors.Is(err, domain.ErrDocumentNotFound) {
			t.Errorf("Se esperaba ErrDocumentNotFound, got %v", err)
		}
	})

	t.Run("UpdateDocumentByAdmin: sin permiso de edición", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDocumentFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
			return &domain.PsiUserDocument{ID: docID}, nil
		}

		_, err := svc.UpdateDocumentByAdmin(ctx, adminLector, docID, request_structs.UpdatePsiUserDocumentRequest{}, nil)
		if !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, got %v", err)
		}
	})

	t.Run("UpdateDocumentByAdmin: tipo inválido", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDocumentFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
			return &domain.PsiUserDocument{ID: docID}, nil
		}
		badType := "otro-tipo-invalido"
		req := request_structs.UpdatePsiUserDocumentRequest{DocumentType: &badType}

		_, err := svc.UpdateDocumentByAdmin(ctx, adminSudo, docID, req, nil)
		if !errors.Is(err, request_structs.ErrDocumentInvalidType) {
			t.Errorf("Se esperaba ErrDocumentInvalidType, got %v", err)
		}
	})

	// ── DeleteDocumentByAdmin ───────────────────────────────────────────

	t.Run("DeleteDocumentByAdmin: éxito sin archivo", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDocumentFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
			return &domain.PsiUserDocument{ID: docID, S3Key: ""}, nil
		}
		deleted := false
		repo.DeleteDocumentFunc = func(ctx context.Context, id uuid.UUID) error {
			deleted = true
			return nil
		}

		if err := svc.DeleteDocumentByAdmin(ctx, adminSudo, docID); err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
		if !deleted {
			t.Error("DeleteDocument no fue invocado")
		}
	})

	t.Run("DeleteDocumentByAdmin: documento inexistente", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDocumentFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
			return nil, domain.ErrDocumentNotFound
		}

		err := svc.DeleteDocumentByAdmin(ctx, adminSudo, docID)
		if !errors.Is(err, domain.ErrDocumentNotFound) {
			t.Errorf("Se esperaba ErrDocumentNotFound, got %v", err)
		}
	})

	t.Run("DeleteDocumentByAdmin: sin permiso de borrado", func(t *testing.T) {
		repo := &mockPsiRepoDocumentos{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDocumentFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserDocument, error) {
			return &domain.PsiUserDocument{ID: docID}, nil
		}

		err := svc.DeleteDocumentByAdmin(ctx, adminLector, docID)
		if !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, got %v", err)
		}
	})
}

// newTestFile devuelve un FileHeader "fantasma" sin datos físicos, suficiente
// para ejercitar las validaciones previas al upload en S3 (que fallan antes de
// intentar abrir el archivo).
func newTestFile() *multipart.FileHeader {
	return &multipart.FileHeader{Filename: "doc.png"}
}
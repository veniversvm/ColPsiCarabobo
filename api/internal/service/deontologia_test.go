package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// mockPsiRepoDeontologia es un mock funcional del repositorio, acotado a las
// operaciones que ejercitan el submódulo de expediente deontológico.
type mockPsiRepoDeontologia struct {
	domain.PsiUserRepository
	GetByIDFunc              func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	CreateDeontologiaFunc    func(ctx context.Context, entry *domain.PsiODeontologia) error
	ListDeontologiaByPsiIDFunc func(ctx context.Context, psiID uuid.UUID) ([]domain.PsiODeontologia, error)
	GetDeontologiaByIDFunc   func(ctx context.Context, id uuid.UUID) (*domain.PsiODeontologia, error)
	UpdateDeontologiaFunc    func(ctx context.Context, id uuid.UUID, content, updateBy string, updateById uuid.UUID) error
}

func (m *mockPsiRepoDeontologia) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *mockPsiRepoDeontologia) CreateDeontologia(ctx context.Context, entry *domain.PsiODeontologia) error {
	if m.CreateDeontologiaFunc != nil {
		return m.CreateDeontologiaFunc(ctx, entry)
	}
	return nil
}

func (m *mockPsiRepoDeontologia) ListDeontologiaByPsiID(ctx context.Context, psiID uuid.UUID) ([]domain.PsiODeontologia, error) {
	if m.ListDeontologiaByPsiIDFunc != nil {
		return m.ListDeontologiaByPsiIDFunc(ctx, psiID)
	}
	return nil, nil
}

func (m *mockPsiRepoDeontologia) GetDeontologiaByID(ctx context.Context, id uuid.UUID) (*domain.PsiODeontologia, error) {
	if m.GetDeontologiaByIDFunc != nil {
		return m.GetDeontologiaByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockPsiRepoDeontologia) UpdateDeontologia(ctx context.Context, id uuid.UUID, content, updateBy string, updateById uuid.UUID) error {
	if m.UpdateDeontologiaFunc != nil {
		return m.UpdateDeontologiaFunc(ctx, id, content, updateBy, updateById)
	}
	return nil
}

func TestPsiService_Deontologia(t *testing.T) {
	ctx := context.Background()
	psiID := uuid.Must(uuid.NewV7())
	entryID := uuid.Must(uuid.NewV7())

	adminSudo := &domain.UserAdmin{
		ID:          uuid.Must(uuid.NewV7()),
		Credentials: domain.Credentials{Username: "admin_tester"},
		Sudo:        true,
	}
	adminSinPermiso := &domain.UserAdmin{
		ID:          uuid.Must(uuid.NewV7()),
		Credentials: domain.Credentials{Username: "lector"},
	}

	psi := &domain.PsiUserModel{ID: psiID}

	t.Run("AddDeontologiaByAdmin: éxito con sanitización", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}
		var captured *domain.PsiODeontologia
		repo.CreateDeontologiaFunc = func(ctx context.Context, entry *domain.PsiODeontologia) error {
			captured = entry
			return nil
		}

		req := request_structs.CreateDeontologiaRequest{
			Content: "  <script>alert(1)</script>Expediente abierto.  ",
		}
		err := svc.AddDeontologiaByAdmin(ctx, adminSudo, psiID, req)
		if err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}

		if captured.PsiUserID != psiID {
			t.Errorf("PsiUserID = %v, want %v", captured.PsiUserID, psiID)
		}
		// El script debe eliminarse y el texto recortarse.
		if captured.Content != "Expediente abierto." {
			t.Errorf("Content = %q, want %q (XSS eliminado + trim)", captured.Content, "Expediente abierto.")
		}
		if captured.CreateBy != adminSudo.Username {
			t.Errorf("CreateBy = %q, want %q", captured.CreateBy, adminSudo.Username)
		}
		if captured.CreateById == nil || *captured.CreateById != adminSudo.ID {
			t.Error("CreateById debería ser el ID del admin ejecutor")
		}
	})

	t.Run("AddDeontologiaByAdmin: permiso insuficiente", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		err := svc.AddDeontologiaByAdmin(ctx, adminSinPermiso, psiID, request_structs.CreateDeontologiaRequest{Content: "x"})
		if !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, got %v", err)
		}
	})

	t.Run("AddDeontologiaByAdmin: psicólogo inexistente", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return nil, domain.ErrPsiNotFound
		}

		err := svc.AddDeontologiaByAdmin(ctx, adminSudo, psiID, request_structs.CreateDeontologiaRequest{Content: "x"})
		if !errors.Is(err, domain.ErrPsiNotFound) {
			t.Errorf("Se esperaba ErrPsiNotFound, got %v", err)
		}
	})

	t.Run("AddDeontologiaByAdmin: contenido vacío tras sanitizar", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		err := svc.AddDeontologiaByAdmin(ctx, adminSudo, psiID, request_structs.CreateDeontologiaRequest{Content: "<script>alert(1)</script>"})
		if !errors.Is(err, domain.ErrInvalidRequest) {
			t.Errorf("Se esperaba ErrInvalidRequest, got %v", err)
		}
	})

	t.Run("ListDeontologiaByPsiID: éxito", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}
		repo.ListDeontologiaByPsiIDFunc = func(ctx context.Context, id uuid.UUID) ([]domain.PsiODeontologia, error) {
			return []domain.PsiODeontologia{{ID: entryID, PsiUserID: psiID, Content: "A"}}, nil
		}

		entries, err := svc.ListDeontologiaByPsiID(ctx, adminSudo, psiID)
		if err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
		if len(entries) != 1 || entries[0].Content != "A" {
			t.Errorf("entries = %+v, want 1 con Content=A", entries)
		}
	})

	t.Run("ListDeontologiaByPsiID: permiso insuficiente", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		_, err := svc.ListDeontologiaByPsiID(ctx, adminSinPermiso, psiID)
		if !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, got %v", err)
		}
	})

	t.Run("UpdateDeontologiaByAdmin: éxito", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDeontologiaByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiODeontologia, error) {
			return &domain.PsiODeontologia{ID: entryID}, nil
		}
		updated := false
		repo.UpdateDeontologiaFunc = func(ctx context.Context, id uuid.UUID, content, updateBy string, updateById uuid.UUID) error {
			updated = true
			return nil
		}
		content := "Expediente corregido"
		req := request_structs.UpdateDeontologiaRequest{Content: &content}

		if err := svc.UpdateDeontologiaByAdmin(ctx, adminSudo, entryID, req); err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
		if !updated {
			t.Error("UpdateDeontologia no fue invocado")
		}
	})

	t.Run("UpdateDeontologiaByAdmin: entrada inexistente", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDeontologiaByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiODeontologia, error) {
			return nil, domain.ErrDeontologiaNotFound
		}
		content := "X"
		req := request_structs.UpdateDeontologiaRequest{Content: &content}

		err := svc.UpdateDeontologiaByAdmin(ctx, adminSudo, entryID, req)
		if !errors.Is(err, domain.ErrDeontologiaNotFound) {
			t.Errorf("Se esperaba ErrDeontologiaNotFound, got %v", err)
		}
	})

	t.Run("UpdateDeontologiaByAdmin: sin permisos", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDeontologiaByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiODeontologia, error) {
			return &domain.PsiODeontologia{ID: entryID}, nil
		}
		content := "X"
		req := request_structs.UpdateDeontologiaRequest{Content: &content}

		err := svc.UpdateDeontologiaByAdmin(ctx, adminSinPermiso, entryID, req)
		if !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, got %v", err)
		}
	})

	t.Run("UpdateDeontologiaByAdmin: contenido nulo", func(t *testing.T) {
		repo := &mockPsiRepoDeontologia{}
		svc := NewPsiService(repo, nil, &mockMailService{})
		repo.GetDeontologiaByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiODeontologia, error) {
			return &domain.PsiODeontologia{ID: entryID}, nil
		}
		req := request_structs.UpdateDeontologiaRequest{Content: nil}

		err := svc.UpdateDeontologiaByAdmin(ctx, adminSudo, entryID, req)
		if !errors.Is(err, domain.ErrInvalidRequest) {
			t.Errorf("Se esperaba ErrInvalidRequest, got %v", err)
		}
	})
}

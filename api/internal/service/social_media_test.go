package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// MOCK DEL REPOSITORIO (Usando Embedding)
// =========================================================================

type mockPsiRepoSocialMedia struct {
	domain.PsiUserRepository
	CountSocialNetworksFunc func(ctx context.Context, psiID uuid.UUID) (int64, error)
	CreateSocialNetworkFunc func(ctx context.Context, sn *domain.PsiUserSocialNetwork) error
	GetSocialNetworkFunc    func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error)
	UpdateSocialNetworkFunc func(ctx context.Context, sn *domain.PsiUserSocialNetwork) error
	DeleteSocialNetworkFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPsiRepoSocialMedia) CountSocialNetworksByPsiID(ctx context.Context, id uuid.UUID) (int64, error) {
	return m.CountSocialNetworksFunc(ctx, id)
}
func (m *mockPsiRepoSocialMedia) CreateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	return m.CreateSocialNetworkFunc(ctx, sn)
}
func (m *mockPsiRepoSocialMedia) GetSocialNetworkByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
	return m.GetSocialNetworkFunc(ctx, id)
}
func (m *mockPsiRepoSocialMedia) UpdateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	return m.UpdateSocialNetworkFunc(ctx, sn)
}
func (m *mockPsiRepoSocialMedia) DeleteSocialNetwork(ctx context.Context, id uuid.UUID) error {
	return m.DeleteSocialNetworkFunc(ctx, id)
}

// =========================================================================
// TESTS UNITARIOS
// =========================================================================

func TestPsiService_AddSocialNetwork(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo} // Asumiendo que tu struct se llama PsiService
	ctx := context.Background()
	psi := &domain.PsiUserModel{ID: uuid.New(), Username: "psicologo1"}

	t.Run("Éxito: Agrega red social dentro del límite", func(t *testing.T) {
		repo.CountSocialNetworksFunc = func(ctx context.Context, id uuid.UUID) (int64, error) {
			return 5, nil // Tiene 5, el límite es 10
		}
		repo.CreateSocialNetworkFunc = func(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
			return nil
		}

		req := request_structs.CreateSocialNetworkRequest{Name: "ig", URL: "https://instagram.com/test"}
		err := svc.AddSocialNetwork(ctx, psi, req)

		if err != nil {
			t.Errorf("No se esperaba error, se obtuvo: %v", err)
		}
	})

	t.Run("Error: Límite de cuota alcanzado", func(t *testing.T) {
		repo.CountSocialNetworksFunc = func(ctx context.Context, id uuid.UUID) (int64, error) {
			return 10, nil // Ya llegó al límite
		}

		req := request_structs.CreateSocialNetworkRequest{Name: "fb", URL: "url"}
		err := svc.AddSocialNetwork(ctx, psi, req)

		if err == nil || err.Error() != "límite de redes sociales alcanzado (10)" {
			t.Errorf("Se esperaba error de límite, se obtuvo: %v", err)
		}
	})
}

func TestPsiService_UpdateSocialNetwork_Ownership(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiA := &domain.PsiUserModel{ID: uuid.New(), Username: "psicologo_A"}
	psiB := &domain.PsiUserModel{ID: uuid.New(), Username: "psicologo_B"}
	netID := uuid.New()

	t.Run("Error: Intento de editar red ajena (ID Spoofing)", func(t *testing.T) {
		// La red pertenece al Psicólogo B
		repo.GetSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
			return &domain.PsiUserSocialNetwork{ID: netID, PsiUserID: psiB.ID}, nil
		}

		// El Psicólogo A intenta editarla
		req := request_structs.UpdateSocialNetworkRequest{}
		err := svc.UpdateSocialNetwork(ctx, psiA, netID, req)

		if err == nil || err.Error() != "no tienes permiso para editar esta red social" {
			t.Errorf("Se esperaba error de permiso, se obtuvo: %v", err)
		}
	})
}

func TestPsiService_DeleteSocialNetwork_Roles(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	ownerID := uuid.New()
	otherID := uuid.New()
	netID := uuid.New()

	repo.GetSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
		return &domain.PsiUserSocialNetwork{ID: netID, PsiUserID: ownerID}, nil
	}

	t.Run("Psi: Puede borrar su propia red", func(t *testing.T) {
		repo.DeleteSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		err := svc.DeleteSocialNetwork(ctx, "psi", ownerID, netID)
		if err != nil {
			t.Errorf("Error inesperado: %v", err)
		}
	})

	t.Run("Psi: No puede borrar red ajena", func(t *testing.T) {
		err := svc.DeleteSocialNetwork(ctx, "psi", otherID, netID)
		if err == nil || err.Error() != "no puedes borrar una red social que no te pertenece" {
			t.Error("Se debió denegar el borrado ajeno")
		}
	})

	t.Run("Admin: Puede borrar cualquier red", func(t *testing.T) {
		repo.DeleteSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		err := svc.DeleteSocialNetwork(ctx, "admin", uuid.New(), netID)
		if err != nil {
			t.Errorf("Admin debería poder borrar, error: %v", err)
		}
	})
}

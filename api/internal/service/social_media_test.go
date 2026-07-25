package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// MOCK DEL REPOSITORIO (Patrón Func Override usando Embedding)
// =========================================================================
// Arquitectura de Testing:
// Se integra (embeds) la interfaz original `domain.PsiUserRepository` para satisfacer
// el contrato implícitamente, pero se sobreescriben únicamente los métodos del
// módulo de Redes Sociales usando funciones dinámicas (Func fields). Esto permite
// aislar los escenarios de prueba en memoria sin requerir una base de datos real.

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
// TESTS UNITARIOS: PRESENCIA DIGITAL Y SEGURIDAD
// =========================================================================

// TestPsiService_AddSocialNetwork evalúa la Prevención de Agotamiento de Recursos.
// Garantiza que la lógica de "Cuotas" (Quotas) funcione correctamente, impidiendo
// que un usuario sature la base de datos creando redes sociales de forma infinita.
func TestPsiService_AddSocialNetwork(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo} // Asumiendo que tu struct se llama PsiService
	ctx := context.Background()
	psi := &domain.PsiUserModel{ID: uuid.Must(uuid.NewV7()), Username: "psicologo1"}

	t.Run("Éxito: Agrega red social dentro del límite", func(t *testing.T) {
		repo.CountSocialNetworksFunc = func(ctx context.Context, id uuid.UUID) (int64, error) {
			return 5, nil // El usuario tiene 5 redes, el límite del sistema es 10
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
			return 10, nil // Ya llegó al límite máximo permitido
		}

		req := request_structs.CreateSocialNetworkRequest{Name: "fb", URL: "url"}
		err := svc.AddSocialNetwork(ctx, psi, req)

		// Aserción Defensiva: El servicio debe rechazar tajantemente la inserción
		if err == nil || err.Error() != "límite de redes sociales alcanzado (10)" {
			t.Errorf("Se esperaba error de límite, se obtuvo: %v", err)
		}
	})
}

// TestPsiService_UpdateSocialNetwork_Ownership evalúa la vulnerabilidad IDOR
// (Insecure Direct Object Reference).
// Verifica el Principio de Confianza Cero (Zero Trust): el sistema no debe confiar
// ciegamente en el ID de la URL que envía el cliente, sino que debe validar en la BD
// que el usuario que ejecuta la acción es el dueño real del registro.
func TestPsiService_UpdateSocialNetwork_Ownership(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiA := &domain.PsiUserModel{ID: uuid.Must(uuid.NewV7()), Username: "psicologo_A"}
	psiB := &domain.PsiUserModel{ID: uuid.Must(uuid.NewV7()), Username: "psicologo_B"}
	netID := uuid.Must(uuid.NewV7())

	t.Run("Error: Intento de editar red ajena (ID Spoofing)", func(t *testing.T) {
		// Mock: La base de datos responde que la red pertenece al Psicólogo B
		repo.GetSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
			return &domain.PsiUserSocialNetwork{ID: netID, PsiUserID: psiB.ID}, nil
		}

		// Ataque Simulado: El Psicólogo A intenta editar la red del Psicólogo B
		req := request_structs.UpdateSocialNetworkRequest{}
		err := svc.UpdateSocialNetwork(ctx, psiA, netID, req)

		// Aserción de Seguridad: El ataque debe ser interceptado en la capa lógica
		if err == nil || !errors.Is(err, domain.ErrSocialPermDenied) {
			t.Errorf("Se esperaba error de permiso, se obtuvo: %v", err)
		}
	})
}

// TestPsiService_DeleteSocialNetwork_Roles evalúa el Control de Acceso Polimórfico.
// Asegura que un mismo método (`DeleteSocialNetwork`) enrute la lógica de autorización
// de forma distinta dependiendo de si quien la invoca es el dueño del registro (Autogestión)
// o el staff del colegio (Moderación Administrativa).
func TestPsiService_DeleteSocialNetwork_Roles(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	ownerID := uuid.Must(uuid.NewV7())
	otherID := uuid.Must(uuid.NewV7())
	netID := uuid.Must(uuid.NewV7())

	repo.GetSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
		return &domain.PsiUserSocialNetwork{ID: netID, PsiUserID: ownerID}, nil
	}

	// Escenario 1: Autogestión exitosa (Ownership Validado)
	t.Run("Psi: Puede borrar su propia red", func(t *testing.T) {
		repo.DeleteSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		err := svc.DeleteSocialNetwork(ctx, "psi", ownerID, netID)
		if err != nil {
			t.Errorf("Error inesperado: %v", err)
		}
	})

	// Escenario 2: Intento de Sabotaje bloqueado (IDOR)
	t.Run("Psi: No puede borrar red ajena", func(t *testing.T) {
		err := svc.DeleteSocialNetwork(ctx, "psi", otherID, netID)
		if err == nil || !errors.Is(err, domain.ErrSocialOwnDenied) {
			t.Error("Se debió denegar el borrado ajeno")
		}
	})

	// Escenario 3: Moderación Global Exitosa (RBAC Bypass por Rol)
	t.Run("Admin: Puede borrar cualquier red", func(t *testing.T) {
		repo.DeleteSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		err := svc.DeleteSocialNetwork(ctx, "admin", uuid.Must(uuid.NewV7()), netID)
		if err != nil {
			t.Errorf("Admin debería poder borrar, error: %v", err)
		}
	})
}

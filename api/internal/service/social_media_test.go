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
	GetByIDFunc             func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
}

func (m *mockPsiRepoSocialMedia) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
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
	psi := &domain.PsiUserModel{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "psicologo1"}}

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

	psiA := &domain.PsiUserModel{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "psicologo_A"}}
	psiB := &domain.PsiUserModel{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "psicologo_B"}}
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
	repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
		return &domain.PsiUserModel{ID: ownerID, FirstName: "Dueño", LastName: "Prueba", FPV: 99999, Credentials: domain.Credentials{Username: "dueno"}}, nil
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

// =========================================================================
// MODERACIÓN ADMINISTRATIVA
// =========================================================================

// TestPsiService_AddSocialNetworkByAdmin evalúa el gatekeeping del panel de
// moderación: solo staff con permisos de gestión de colegiados puede añadir
// una red social a una ficha ajena.
func TestPsiService_AddSocialNetworkByAdmin(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiID := uuid.Must(uuid.NewV7())
	adminSudo := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "sudo_admin"}, Sudo: true}
	adminSinPermisos := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "lector_admin"}}
	req := request_structs.CreateSocialNetworkRequest{Name: "ig", URL: "https://instagram.com/test"}

	repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
		return &domain.PsiUserModel{ID: psiID}, nil
	}
	repo.CountSocialNetworksFunc = func(ctx context.Context, id uuid.UUID) (int64, error) { return 2, nil }
	repo.CreateSocialNetworkFunc = func(ctx context.Context, sn *domain.PsiUserSocialNetwork) error { return nil }

	t.Run("Éxito: Sudo añade red a ficha ajena con auditoría admin", func(t *testing.T) {
		var captured *domain.PsiUserSocialNetwork
		repo.CreateSocialNetworkFunc = func(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
			captured = sn
			return nil
		}
		err := svc.AddSocialNetworkByAdmin(ctx, adminSudo, psiID, req)
		if err != nil {
			t.Fatalf("No se esperaba error, se obtuvo: %v", err)
		}
		if captured == nil || captured.PsiUserID != psiID {
			t.Error("La red no se persistió para la ficha indicada")
		}
		if captured == nil || captured.CreateBy != "sudo_admin" {
			t.Error("La auditoría debe registrar al operador administrativo")
		}
	})

	t.Run("Error: Permisos insuficientes", func(t *testing.T) {
		err := svc.AddSocialNetworkByAdmin(ctx, adminSinPermisos, psiID, req)
		if err == nil || !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, se obtuvo: %v", err)
		}
	})

	t.Run("Error: Cuota alcanzada", func(t *testing.T) {
		repo.CountSocialNetworksFunc = func(ctx context.Context, id uuid.UUID) (int64, error) { return 10, nil }
		err := svc.AddSocialNetworkByAdmin(ctx, adminSudo, psiID, req)
		if err == nil || !errors.Is(err, domain.ErrMaxSocialNetworks) {
			t.Errorf("Se esperaba ErrMaxSocialNetworks, se obtuvo: %v", err)
		}
		repo.CountSocialNetworksFunc = func(ctx context.Context, id uuid.UUID) (int64, error) { return 2, nil }
	})
}

// TestPsiService_UpdateSocialNetworkByAdmin evalúa el gatekeeping y la verificación
// de pertenencia (IDOR) en la edición administrativa de una red social.
func TestPsiService_UpdateSocialNetworkByAdmin(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiID := uuid.Must(uuid.NewV7())
	otraFicha := uuid.Must(uuid.NewV7())
	netID := uuid.Must(uuid.NewV7())
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "secretaria"}, CanUpdatePsi: true}

	repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
		return &domain.PsiUserModel{ID: psiID}, nil
	}

	t.Run("Éxito: actualiza red perteneciente a la ficha", func(t *testing.T) {
		repo.GetSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
			return &domain.PsiUserSocialNetwork{ID: netID, PsiUserID: psiID}, nil
		}
		repo.UpdateSocialNetworkFunc = func(ctx context.Context, sn *domain.PsiUserSocialNetwork) error { return nil }
		name := "Instagram"
		err := svc.UpdateSocialNetworkByAdmin(ctx, admin, psiID, netID, request_structs.UpdateSocialNetworkRequest{Name: &name})
		if err != nil {
			t.Errorf("No se esperaba error, se obtuvo: %v", err)
		}
	})

	t.Run("Error: IDOR — red de otra ficha", func(t *testing.T) {
		repo.GetSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
			return &domain.PsiUserSocialNetwork{ID: netID, PsiUserID: otraFicha}, nil
		}
		err := svc.UpdateSocialNetworkByAdmin(ctx, admin, psiID, netID, request_structs.UpdateSocialNetworkRequest{})
		if err == nil || !errors.Is(err, domain.ErrSocialPermDenied) {
			t.Errorf("Se esperaba ErrSocialPermDenied, se obtuvo: %v", err)
		}
	})

	t.Run("Error: Permisos insuficientes", func(t *testing.T) {
		lector := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "lector"}}
		err := svc.UpdateSocialNetworkByAdmin(ctx, lector, psiID, netID, request_structs.UpdateSocialNetworkRequest{})
		if err == nil || !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, se obtuvo: %v", err)
		}
	})
}

// TestPsiService_DeleteSocialNetworkByAdmin evalúa el gatekeeping y la verificación
// de pertenencia (IDOR) en el borrado administrativo de una red social.
func TestPsiService_DeleteSocialNetworkByAdmin(t *testing.T) {
	repo := &mockPsiRepoSocialMedia{}
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiID := uuid.Must(uuid.NewV7())
	otraFicha := uuid.Must(uuid.NewV7())
	netID := uuid.Must(uuid.NewV7())
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "secretaria"}, CanDeletePsi: true}

	repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
		return &domain.PsiUserModel{ID: psiID, FirstName: "Ficha", LastName: "Objetivo", FPV: 77777}, nil
	}

	t.Run("Éxito: borra red perteneciente a la ficha", func(t *testing.T) {
		repo.GetSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
			return &domain.PsiUserSocialNetwork{ID: netID, PsiUserID: psiID}, nil
		}
		repo.DeleteSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		err := svc.DeleteSocialNetworkByAdmin(ctx, admin, psiID, netID)
		if err != nil {
			t.Errorf("No se esperaba error, se obtuvo: %v", err)
		}
	})

	t.Run("Error: IDOR — red de otra ficha", func(t *testing.T) {
		repo.GetSocialNetworkFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
			return &domain.PsiUserSocialNetwork{ID: netID, PsiUserID: otraFicha}, nil
		}
		err := svc.DeleteSocialNetworkByAdmin(ctx, admin, psiID, netID)
		if err == nil || !errors.Is(err, domain.ErrSocialOwnDenied) {
			t.Errorf("Se esperaba ErrSocialOwnDenied, se obtuvo: %v", err)
		}
	})

	t.Run("Error: Permisos insuficientes", func(t *testing.T) {
		lector := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "lector"}}
		err := svc.DeleteSocialNetworkByAdmin(ctx, lector, psiID, netID)
		if err == nil || !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("Se esperaba ErrInsufficientPerms, se obtuvo: %v", err)
		}
	})
}

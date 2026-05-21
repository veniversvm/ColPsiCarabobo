package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// MOCKS PARA ADMIN REPOSITORY
// =========================================================================

type mockAdminRepo struct {
	domain.UserAdminRepository
	GetByIdentifierFunc func(ctx context.Context, identifier string) (*domain.UserAdmin, error)
	GetByIDFunc         func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error)
	UpdateFunc          func(ctx context.Context, admin *domain.UserAdmin) error
	CreateFunc          func(ctx context.Context, admin *domain.UserAdmin) error
	DeleteFunc          func(ctx context.Context, id uuid.UUID) error
	ListFunc            func(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error)
}

func (m *mockAdminRepo) GetByIdentifier(ctx context.Context, id string) (*domain.UserAdmin, error) {
	return m.GetByIdentifierFunc(ctx, id)
}
func (m *mockAdminRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockAdminRepo) Update(ctx context.Context, a *domain.UserAdmin) error {
	return m.UpdateFunc(ctx, a)
}
func (m *mockAdminRepo) Create(ctx context.Context, a *domain.UserAdmin) error {
	return m.CreateFunc(ctx, a)
}
func (m *mockAdminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *mockAdminRepo) List(ctx context.Context, a *bool, s string, p, l int) ([]domain.UserAdmin, int64, error) {
	return m.ListFunc(ctx, a, s, p, l)
}

// =========================================================================
// TEST SUITE
// =========================================================================

func TestAdminService_All(t *testing.T) {
	repo := &mockAdminRepo{}
	// Usamos nil para mailService por brevedad
	svc := NewAdminService(repo, nil)

	// --- 1. TEST DE LOGIN Y KEY ROTATION ---
	t.Run("Login: Éxito y Rotación de Key", func(t *testing.T) {
		pass := "Admin123!"
		hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		adminID := uuid.New()

		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{
				ID:       adminID,
				Password: string(hashed),
				IsActive: true,
				Username: "admin_test",
				Email:    "admin@gmail.com", // Fix: dominio real
			}, nil
		}

		repo.UpdateFunc = func(ctx context.Context, a *domain.UserAdmin) error {
			if a.Key == "" {
				t.Error("La Key de sesión no fue rotada")
			}
			return nil
		}

		token, err := svc.Login(context.Background(), "admin_test", pass)
		if err != nil || token == "" {
			t.Errorf("Error en login exitoso: %v", err)
		}
	})

	// --- 2. TEST DE CACHE-ASIDE PATTERN ---
	t.Run("GetAdmins: Verificación de Caché", func(t *testing.T) {
		callCount := 0
		repo.ListFunc = func(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
			callCount++
			return []domain.UserAdmin{{Username: "cached_user"}}, 1, nil
		}

		// Primera llamada (va a Repo)
		_, _ = svc.GetAdmins(context.Background(), nil, "", 1, 10)
		// Segunda llamada (debe venir de Caché)
		_, _ = svc.GetAdmins(context.Background(), nil, "", 1, 10)

		if callCount > 1 {
			t.Errorf("El sistema de caché no está funcionando, repo llamado %d veces", callCount)
		}
	})

	// --- 3. TEST DE JERARQUÍA (CREATE ADMIN) ---
	t.Run("CreateAdmin: Regla 'No puedes dar lo que no tienes'", func(t *testing.T) {
		creator := domain.UserAdmin{
			ID:             uuid.New(),
			Username:       "moderador",
			CanCreateAdmin: true,
			Sudo:           false, // No es Super Usuario
			CanPublish:     false, // <--- NO tiene permiso de publicar
		}

		trueVal := true
		req := request_structs.CreateAdminRequest{
			Username: "nuevo_admin",
			// FIX: Usamos @gmail.com para que pase la validación MX de utilidades
			Email:    "nuevo_admin_test@gmail.com",
			Password: "Password123!",
			Permissions: request_structs.AdminPermissionsDTO{
				CanPublish: &trueVal, // <--- Intenta dar un permiso que él no tiene
			},
		}

		err := svc.CreateAdmin(context.Background(), creator, req)
		if err == nil || err.Error() != "no puedes otorgar el permiso: Publish" {
			t.Errorf("Se esperaba bloqueo por jerarquía, se obtuvo: %v", err)
		}
	})

	// --- 4. TEST DE PROTECCIÓN SUDO (UPDATE) ---
	t.Run("UpdateAdmin: Proteger Super Usuario de ediciones externas", func(t *testing.T) {
		updater := domain.UserAdmin{ID: uuid.New(), Sudo: false}    // Admin normal
		targetSudo := &domain.UserAdmin{ID: uuid.New(), Sudo: true} // Destino Sudo

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return targetSudo, nil
		}

		req := request_structs.UpdateAdminRequest{ID: targetSudo.ID}
		err := svc.UpdateAdmin(context.Background(), updater, req)

		if err == nil || err.Error() != "no puedes editar a un Super Usuario" {
			t.Errorf("Un admin normal pudo editar a un Sudo: %v", err)
		}
	})

	// --- 5. TEST DE PREVENCIÓN DE AUTO-SUICIDIO (DELETE) ---
	t.Run("DeleteAdmin: Impedir auto-eliminación", func(t *testing.T) {
		adminID := uuid.New()
		updater := &domain.UserAdmin{ID: adminID}

		err := svc.DeleteAdmin(context.Background(), updater, adminID)
		if err == nil || err.Error() != "no puedes eliminar tu propia cuenta" {
			t.Errorf("Se esperaba error de auto-eliminación, se obtuvo: %v", err)
		}
	})
}

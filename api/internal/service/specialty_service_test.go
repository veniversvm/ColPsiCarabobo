package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// MOCK DEL REPOSITORIO
// =========================================================================

type mockSpecialtyRepo struct {
	domain.SpecialtyRepository // Embedding para cumplir la interfaz
	CreateFunc                 func(ctx context.Context, s *domain.PsiSpecialtyModel) error
	GetByIDFunc                func(ctx context.Context, id uint32, includeInactive bool) (*domain.PsiSpecialtyModel, error)
	UpdateFunc                 func(ctx context.Context, s *domain.PsiSpecialtyModel) error
	GetAllFunc                 func(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error)
	CountFunc                  func(ctx context.Context, active *bool) (int64, error)
}

func (m *mockSpecialtyRepo) Create(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return m.CreateFunc(ctx, s)
}
func (m *mockSpecialtyRepo) GetByID(ctx context.Context, id uint32, includeInactive bool) (*domain.PsiSpecialtyModel, error) {
	return m.GetByIDFunc(ctx, id, includeInactive)
}
func (m *mockSpecialtyRepo) Update(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	return m.UpdateFunc(ctx, s)
}
func (m *mockSpecialtyRepo) GetAll(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
	return m.GetAllFunc(ctx, status)
}
func (m *mockSpecialtyRepo) Count(ctx context.Context, active *bool) (int64, error) {
	return m.CountFunc(ctx, active)
}

// =========================================================================
// TESTS UNITARIOS
// =========================================================================

func TestSpecialtyService_Create(t *testing.T) {
	repo := &mockSpecialtyRepo{}
	svc := NewSpecialtyService(repo)
	ctx := context.Background()
	adminID := uuid.New()

	tests := []struct {
		name    string
		admin   *domain.UserAdmin
		req     request_structs.CreateSpecialtyRequest
		mockErr error
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Éxito: Admin con permiso CanCreateTags",
			admin:   &domain.UserAdmin{ID: adminID, Username: "admin1", CanCreateTags: true},
			req:     request_structs.CreateSpecialtyRequest{Name: "Psicología Clínica"},
			wantErr: false,
		},
		{
			name:    "Éxito: Superusuario (Sudo) sin permiso explícito",
			admin:   &domain.UserAdmin{ID: adminID, Username: "sudo_user", Sudo: true, CanCreateTags: false},
			req:     request_structs.CreateSpecialtyRequest{Name: "Neuropsicología"},
			wantErr: false,
		},
		{
			name:    "Error: Admin sin permisos",
			admin:   &domain.UserAdmin{ID: adminID, Username: "pobre_admin", CanCreateTags: false, Sudo: false},
			req:     request_structs.CreateSpecialtyRequest{Name: "Fake"},
			wantErr: true,
			errMsg:  "no tienes permiso para crear especialidades",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.CreateFunc = func(ctx context.Context, s *domain.PsiSpecialtyModel) error {
				return tt.mockErr
			}

			err := svc.Create(ctx, tt.admin, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Create() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestSpecialtyService_Update(t *testing.T) {
	repo := &mockSpecialtyRepo{}
	svc := NewSpecialtyService(repo)
	ctx := context.Background()

	nameUpdate := "Nuevo Nombre"

	t.Run("Actualización Parcial (PATCH) inyecta auditoría correctamente", func(t *testing.T) {
		admin := &domain.UserAdmin{ID: uuid.New(), Username: "editor", CanEditTags: true}
		existingSpec := &domain.PsiSpecialtyModel{ID: 10, Name: "Viejo Nombre"}

		repo.GetByIDFunc = func(ctx context.Context, id uint32, includeInactive bool) (*domain.PsiSpecialtyModel, error) {
			return existingSpec, nil
		}

		repo.UpdateFunc = func(ctx context.Context, s *domain.PsiSpecialtyModel) error {
			// Verificamos que los campos de auditoría se hayan actualizado
			if s.UpdateBy != admin.Username {
				t.Errorf("UpdateBy no se actualizó, got %s", s.UpdateBy)
			}
			if s.Name != nameUpdate {
				t.Errorf("Name no se actualizó, got %s", s.Name)
			}
			return nil
		}

		req := request_structs.UpdateSpecialtyRequest{Name: &nameUpdate}
		err := svc.Update(ctx, admin, 10, req)
		if err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
	})
}

func TestSpecialtyService_GetSpecialties_Visibility(t *testing.T) {
	repo := &mockSpecialtyRepo{}
	svc := NewSpecialtyService(repo)

	tests := []struct {
		name            string
		isAdmin         bool
		requestedStatus string
		expectedStatus  string
	}{
		{
			name:            "Usuario normal: Fuerza estado 'active' aunque pida 'all'",
			isAdmin:         false,
			requestedStatus: "all",
			expectedStatus:  "active",
		},
		{
			name:            "Admin: Puede pedir 'inactive'",
			isAdmin:         true,
			requestedStatus: "inactive",
			expectedStatus:  "inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.GetAllFunc = func(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
				if status != tt.expectedStatus {
					t.Errorf("Se esperaba status '%s', se recibió '%s'", tt.expectedStatus, status)
				}
				return []domain.PsiSpecialtyModel{}, nil
			}
			_, _ = svc.GetSpecialties(context.Background(), tt.requestedStatus, tt.isAdmin)
		})
	}
}

func TestSpecialtyService_Count_Rules(t *testing.T) {
	repo := &mockSpecialtyRepo{}
	svc := NewSpecialtyService(repo)

	t.Run("Usuario anónimo: fuerza conteo de solo activos", func(t *testing.T) {
		repo.CountFunc = func(ctx context.Context, active *bool) (int64, error) {
			if active == nil || !*active {
				t.Error("Se esperaba que 'active' fuera true para usuario anónimo")
			}
			return 10, nil
		}
		_, _ = svc.Count(context.Background(), nil, nil) // Admin nil
	})
}

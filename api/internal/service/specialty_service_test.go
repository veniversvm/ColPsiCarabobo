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
// Mediante el "Embedding" de `domain.SpecialtyRepository`, el mock satisface el
// contrato de la interfaz automáticamente. Sobrescribir los métodos con variables
// de tipo función (Func fields) permite inyectar comportamientos dinámicos
// ("Stubs" y "Spies") dentro del contexto cerrado de cada sub-test (t.Run).

type mockSpecialtyRepo struct {
	domain.SpecialtyRepository // Embedding para cumplir la interfaz implícitamente
	CreateFunc                 func(ctx context.Context, s *domain.PsiSpecialtyModel) error
	GetByIDFunc                func(ctx context.Context, id uint32, includeInactive bool) (*domain.PsiSpecialtyModel, error)
	GetByAdminIDFunc           func(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error)
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
func (m *mockSpecialtyRepo) GetByAdminID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	return m.GetByAdminIDFunc(ctx, id)
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
// TESTS UNITARIOS: MASTER DATA MANAGEMENT (MDM) Y SEGURIDAD
// =========================================================================

// TestSpecialtyService_Create evalúa el Control de Acceso Basado en Roles (RBAC).
//
// Patrón de Prueba (Table-Driven Testing - TDT):
// Se define una matriz de casos de prueba (`tests`) que evalúan exhaustivamente
// los límites de la matriz de permisos. Esto garantiza que la lógica de "Gatekeeping"
// sea determinística y no permita escaladas de privilegios accidentales.
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
		errIs   error
	}{
		{
			name:    "Éxito: Admin con permiso CanCreateTags",
			admin:   &domain.UserAdmin{ID: adminID, Credentials: domain.Credentials{Username: "admin1"}, CanCreateTags: true},
			req:     request_structs.CreateSpecialtyRequest{Name: "Psicología Clínica"},
			wantErr: false,
		},
		{
			name:    "Éxito: Superusuario (Sudo) sin permiso explícito",
			admin:   &domain.UserAdmin{ID: adminID, Credentials: domain.Credentials{Username: "sudo_user"}, Sudo: true, CanCreateTags: false},
			req:     request_structs.CreateSpecialtyRequest{Name: "Neuropsicología"},
			wantErr: false,
		},
		{
			name:    "Error: Admin sin permisos",
			admin:   &domain.UserAdmin{ID: adminID, Credentials: domain.Credentials{Username: "pobre_admin"}, CanCreateTags: false, Sudo: false},
			req:     request_structs.CreateSpecialtyRequest{Name: "Fake"},
			wantErr: true,
			errIs:   domain.ErrInsufficientPerms,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Stub de éxito silencioso para la base de datos
			repo.CreateFunc = func(ctx context.Context, s *domain.PsiSpecialtyModel) error {
				return tt.mockErr
			}

			err := svc.Create(ctx, tt.admin, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errIs != nil && !errors.Is(err, tt.errIs) {
				t.Errorf("Create() error = %v, want errors.Is %v", err, tt.errIs)
			}
		})
	}
}

// TestSpecialtyService_Update evalúa la Mutación de Datos y el Rastro de Auditoría.
//
// Verifica que la Semántica PATCH del DTO funcione, y garantiza el cumplimiento
// de "No Repudio" (Non-Repudiation): el servicio debe obligatoriamente sellar
// el registro con el Username de quien ejecutó la acción.
func TestSpecialtyService_Update(t *testing.T) {
	repo := &mockSpecialtyRepo{}
	svc := NewSpecialtyService(repo)
	ctx := context.Background()

	nameUpdate := "Nuevo Nombre"

	t.Run("Actualización Parcial (PATCH) inyecta auditoría correctamente", func(t *testing.T) {
		admin := &domain.UserAdmin{ID: uuid.New(), Credentials: domain.Credentials{Username: "editor"}, CanEditTags: true}
		existingSpec := &domain.PsiSpecialtyModel{ID: 10, Name: "Viejo Nombre"}

		// Simulamos la lectura previa (Read-Before-Write)
		repo.GetByAdminIDFunc = func(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
			return existingSpec, nil
		}

		// Espiamos el payload final (Spy Pattern) justo antes de que toque la BD.
		repo.UpdateFunc = func(ctx context.Context, s *domain.PsiSpecialtyModel) error {
			// Aserción de Trazabilidad Forense
			if s.UpdateBy != admin.Username {
				t.Errorf("UpdateBy no se actualizó, got %s", s.UpdateBy)
			}
			// Aserción de Mutación PATCH
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

// TestSpecialtyService_GetSpecialties_Visibility evalúa la Degradación de Visibilidad.
//
// Patrón de Diseño: Fail-Safe Defaults (Seguro por Defecto).
// Verifica que un usuario malintencionado no pueda aplicar "Fuzzing" de parámetros
// enviando `?status=all` para descubrir taxonomías internas. La lógica debe reescribir
// la variable forzando la seguridad.
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
			// Escenario Restringido: El Backend sobrescribe un parámetro hostil.
			name:            "Usuario normal: Fuerza estado 'active' aunque pida 'all'",
			isAdmin:         false,
			requestedStatus: "all",
			expectedStatus:  "active",
		},
		{
			// Escenario Confiable: Al administrador se le respeta su comando.
			name:            "Admin: Puede pedir 'inactive'",
			isAdmin:         true,
			requestedStatus: "inactive",
			expectedStatus:  "inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.GetAllFunc = func(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
				// Aserción de Filtrado Defensivo
				if status != tt.expectedStatus {
					t.Errorf("Se esperaba status '%s', se recibió '%s'", tt.expectedStatus, status)
				}
				return []domain.PsiSpecialtyModel{}, nil
			}
			_, _ = svc.GetSpecialties(context.Background(), tt.requestedStatus, tt.isAdmin)
		})
	}
}

// TestSpecialtyService_Count_Rules evalúa la Prevención de Fugas Telemétricas.
// Garantiza que los visitantes no puedan usar el endpoint de métricas para deducir
// el tamaño oculto de la base de datos institucional.
func TestSpecialtyService_Count_Rules(t *testing.T) {
	repo := &mockSpecialtyRepo{}
	svc := NewSpecialtyService(repo)

	t.Run("Usuario anónimo: fuerza conteo de solo activos", func(t *testing.T) {
		repo.CountFunc = func(ctx context.Context, active *bool) (int64, error) {
			// Aserción de Seguridad: El servicio debe inyectar el puntero en 'true'
			if active == nil || !*active {
				t.Error("Se esperaba que 'active' fuera true para usuario anónimo")
			}
			return 10, nil
		}
		// Invocación como visitante anónimo (Admin = nil)
		_, _ = svc.Count(context.Background(), nil, nil)
	})
}

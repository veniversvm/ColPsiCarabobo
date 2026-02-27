package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// MOCKS ESPECÍFICOS PARA ESTA SUITE
// =========================================================================

type mockMailService struct {
	SendEmailFunc func(to, subject, template string, data interface{}) error
}

func (m *mockMailService) SendEmail(to, subject, template string, data interface{}) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(to, subject, template, data)
	}
	return nil
}

// Reutilizamos el mock de repositorio previo pero añadiendo el método de Admin
type mockPsiRepoAdmin struct {
	domain.PsiUserRepository
	CreateWithColDataFunc func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData) error
	GetByIDFunc           func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	UpdateFunc            func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData) error
}

func (m *mockPsiRepoAdmin) CreateWithColData(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData) error {
	return m.CreateWithColDataFunc(ctx, p, c)
}
func (m *mockPsiRepoAdmin) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockPsiRepoAdmin) Update(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData) error {
	return m.UpdateFunc(ctx, p, c)
}

// =========================================================================
// TESTS DE CREACIÓN (RBAC & ATOMICIDAD)
// =========================================================================

func TestPsiService_CreateByAdmin(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	mail := &mockMailService{}
	svc := &PsiService{repo: repo, mailService: mail} // Inyección de dependencias
	ctx := context.Background()

	admin := &domain.UserAdmin{ID: uuid.New(), Username: "admin_tester", CanCreatePsi: true}

	t.Run("Éxito: Registro completo con ColData y Mail", func(t *testing.T) {
		repo.CreateWithColDataFunc = func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData) error {
			// Validamos que el ID se haya vinculado correctamente entre tablas
			if psi.ID != col.PsiUserModelID {
				t.Error("El ID del Psicólogo no coincide con el Foreign Key de ColData")
			}
			// Validamos auditoría
			if psi.CreateBy != "admin_tester" {
				t.Errorf("Auditoría falló: CreateBy = %s", psi.CreateBy)
			}
			return nil
		}

		req := request_structs.CreatePsiAdminRequest{
			Username: "lic_perez",
			Email:    "perez@test.com",
			Password: "secure_password",
			BornDate: "1990-05-20",
		}

		err := svc.CreatePsiByAdmin(ctx, admin, req)
		if err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
	})

	t.Run("Error: Admin sin permisos de creación", func(t *testing.T) {
		pobreAdmin := &domain.UserAdmin{CanCreatePsi: false, Sudo: false}
		err := svc.CreatePsiByAdmin(ctx, pobreAdmin, request_structs.CreatePsiAdminRequest{})

		if err == nil || err.Error() != "no tienes permiso para registrar psicólogos" {
			t.Error("Se debió denegar el acceso al admin sin permisos")
		}
	})
}

// =========================================================================
// TESTS DE ACTUALIZACIÓN (PATCH & PROPIEDAD)
// =========================================================================

func TestPsiService_UpdateByAdmin_Patch(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	svc := &PsiService{repo: repo}
	ctx := context.Background()
	targetID := uuid.New()
	admin := &domain.UserAdmin{ID: uuid.New(), Username: "super_admin", Sudo: true}

	t.Run("Actualización Parcial: Solo cambia el estatus de solvencia", func(t *testing.T) {
		// Mock: El psicólogo actual en DB
		currentPsi := &domain.PsiUserModel{
			ID:      targetID,
			Solvent: false,
			ColData: domain.PsiUserColData{RegisterNumber: 12345},
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return currentPsi, nil
		}

		repo.UpdateFunc = func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData) error {
			if !psi.Solvent {
				t.Error("El campo Solvent debería ser true tras el update")
			}
			if col.RegisterNumber != 12345 {
				t.Error("El RegisterNumber no debería haber cambiado (era nil en el request)")
			}
			return nil
		}

		newSolvency := true
		req := request_structs.UpdatePsiAdminRequest{
			Solvent: &newSolvency, // Solo enviamos solvencia
		}

		err := svc.UpdatePsiByAdmin(ctx, admin, targetID, req)
		if err != nil {
			t.Errorf("Error en Update: %v", err)
		}
	})
}

// =========================================================================
// TESTS DE VISIBILIDAD (ADMIN DIRECTORY)
// =========================================================================

func TestPsiService_GetAdminDirectory_Security(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	svc := &PsiService{repo: repo}

	t.Run("Bloqueo de seguridad: No admins no pueden listar", func(t *testing.T) {
		randomUser := &domain.UserAdmin{Sudo: false, CanUpdatePsi: false, CanCreatePsi: false}
		_, err := svc.GetAdminDirectory(context.Background(), randomUser, request_structs.PsiDirectoryFilterDTO{})

		if err == nil || err.Error() != "permisos insuficientes para listar agremiados" {
			t.Error("El sistema permitió el acceso al directorio a un usuario no autorizado")
		}
	})
}

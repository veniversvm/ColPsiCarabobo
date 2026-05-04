package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// MOCKS ESPECÍFICOS
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

type mockPsiRepoAdmin struct {
	domain.PsiUserRepository
	CreateWithColDataFunc func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData) error
	GetByIDFunc           func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	// Ajustado a 4 argumentos
	UpdateFunc func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, text *domain.TextModel) error
}

func (m *mockPsiRepoAdmin) CreateWithColData(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData) error {
	return m.CreateWithColDataFunc(ctx, p, c)
}
func (m *mockPsiRepoAdmin) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockPsiRepoAdmin) Update(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, t *domain.TextModel) error {
	return m.UpdateFunc(ctx, p, c, t)
}

// =========================================================================
// TESTS
// =========================================================================

func TestPsiService_CreateByAdmin(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	mail := &mockMailService{}

	// FIX: Usamos el constructor oficial para que inicialice el sanitizer internamente
	svc := NewPsiService(repo, nil, mail)
	ctx := context.Background()

	admin := &domain.UserAdmin{ID: uuid.New(), Username: "admin_tester", CanCreatePsi: true}

	t.Run("Éxito: Registro completo", func(t *testing.T) {
		repo.CreateWithColDataFunc = func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData) error {
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
}

func TestPsiService_UpdateByAdmin_Patch(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	mail := &mockMailService{}
	svc := NewPsiService(repo, nil, mail)

	ctx := context.Background()
	targetID := uuid.New()
	admin := &domain.UserAdmin{ID: uuid.New(), Username: "super_admin", Sudo: true}

	t.Run("Actualización Parcial: Solo cambia el estatus de solvencia", func(t *testing.T) {
		currentPsi := &domain.PsiUserModel{
			ID:      targetID,
			Solvent: false,
			ColData: domain.PsiUserColData{PsiUserModelID: targetID, RegisterNumber: 12345},
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return currentPsi, nil
		}

		repo.UpdateFunc = func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, text *domain.TextModel) error {
			return nil
		}

		// FIX: Inicializar punteros para evitar el pánico en los DEBUG Printf (líneas 245-246)
		newSolvency := true
		mun := "Valencia"
		state := "Aragua"

		req := request_structs.UpdatePsiAdminRequest{
			Solvent:              &newSolvency,
			MunicipalityCarabobo: &mun,   // Evita pánico en *req.MunicipalityCarabobo
			StateOutside:         &state, // Evita pánico en *req.StateOutside
		}

		// Los FileHeader pueden ser nil porque el servicio hace "if file != nil"
		err := svc.UpdatePsiByAdmin(ctx, admin, targetID, req, nil, nil, nil, nil)
		if err != nil {
			t.Errorf("Error en Update: %v", err)
		}
	})
}

func TestPsiService_GetAdminDirectory_Security(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	// Aquí da igual porque no llegamos a usar mail o sanitizer
	svc := NewPsiService(repo, nil, nil)

	t.Run("Bloqueo de seguridad: No admins", func(t *testing.T) {
		randomUser := &domain.UserAdmin{Sudo: false, CanUpdatePsi: false, CanCreatePsi: false}
		_, err := svc.GetAdminDirectory(context.Background(), randomUser, request_structs.PsiDirectoryFilterDTO{})

		if err == nil || err.Error() != "permisos insuficientes para listar agremiados" {
			t.Error("Se debió denegar el acceso")
		}
	})
}

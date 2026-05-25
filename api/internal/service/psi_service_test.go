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
// MOCK INTEGRAL DEL REPOSITORIO (Renombrado para evitar conflictos)
// =========================================================================

type mockPsiRepoSvc struct {
	domain.PsiUserRepository

	GetByIDFunc                   func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	GetByIdentifierFunc           func(ctx context.Context, id string) (*domain.PsiUserModel, error)
	GetByFPVFunc                  func(ctx context.Context, fpv int) (domain.PsiUserModel, error)
	UpdateKeyFunc                 func(ctx context.Context, psi *domain.PsiUserModel) error
	GetPsiUserColDataFunc         func(ctx context.Context, id uuid.UUID) (*domain.PsiUserColData, error)
	UpdatePublicProfileFunc       func(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, t *domain.TextModel) error
	GetPostGradeByIDFunc          func(ctx context.Context, id uuid.UUID) (*domain.PsiUserPostGrade, error)
	UpdatePostGradeFunc           func(ctx context.Context, pg *domain.PsiUserPostGrade) error
	CreatePostGradeFunc           func(ctx context.Context, pg *domain.PsiUserPostGrade) error
	GetTextContentByIDFunc        func(ctx context.Context, id uuid.UUID) (string, error)
	ValidateUniqueCredentialsFunc func(ctx context.Context, username, email string, excludeID uuid.UUID) error
}

func (m *mockPsiRepoSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockPsiRepoSvc) GetByIdentifier(ctx context.Context, id string) (*domain.PsiUserModel, error) {
	return m.GetByIdentifierFunc(ctx, id)
}
func (m *mockPsiRepoSvc) GetByFPV(ctx context.Context, fpv int) (domain.PsiUserModel, error) {
	return m.GetByFPVFunc(ctx, fpv)
}
func (m *mockPsiRepoSvc) UpdateKey(ctx context.Context, p *domain.PsiUserModel) error {
	return m.UpdateKeyFunc(ctx, p)
}
func (m *mockPsiRepoSvc) GetPsiUserColData(ctx context.Context, id uuid.UUID) (*domain.PsiUserColData, error) {
	return m.GetPsiUserColDataFunc(ctx, id)
}
func (m *mockPsiRepoSvc) UpdatePublicProfile(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, t *domain.TextModel) error {
	return m.UpdatePublicProfileFunc(ctx, p, c, t)
}
func (m *mockPsiRepoSvc) GetPostGradeByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserPostGrade, error) {
	return m.GetPostGradeByIDFunc(ctx, id)
}
func (m *mockPsiRepoSvc) UpdatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return m.UpdatePostGradeFunc(ctx, pg)
}
func (m *mockPsiRepoSvc) CreatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return m.CreatePostGradeFunc(ctx, pg)
}
func (m *mockPsiRepoSvc) GetTextContentByID(ctx context.Context, id uuid.UUID) (string, error) {
	if m.GetTextContentByIDFunc == nil {
		return "", nil
	}
	return m.GetTextContentByIDFunc(ctx, id)
}
func (m *mockPsiRepoSvc) ValidateUniqueCredentials(ctx context.Context, u, e string, ex uuid.UUID) error {
	return m.ValidateUniqueCredentialsFunc(ctx, u, e, ex)
}

type mockMailSvc struct {
	SendEmailFunc func(to, subject, template string, data any) error
}

func (m *mockMailSvc) SendEmail(to, subject, template string, data any) error {
	return m.SendEmailFunc(to, subject, template, data)
}

// =========================================================================
// TESTS
// =========================================================================

func TestPsiService_GetPublicProfile_Privacy(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)
	psiFPV := 12345

	t.Run("Restricción de Solvencia: No solvente no muestra Postgrados", func(t *testing.T) {
		repo.GetByFPVFunc = func(ctx context.Context, id int) (domain.PsiUserModel, error) {
			return domain.PsiUserModel{
				ID:       uuid.Must(uuid.NewV7()),
				FPV:      psiFPV,
				IsActive: true,
				Solvent:  false, // Gatilla el retorno temprano
				ColData: domain.PsiUserColData{
					UniversityUndergraduate: "UCV",
				},
			}, nil
		}

		profile, _, err := svc.GetPublicProfile(context.Background(), psiFPV)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}
		if len(profile.PostGrades) != 0 {
			t.Error("Un usuario NO solvente no debería mostrar sus postgrados")
		}
	})

	t.Run("Privacidad Personal: Ocultar email", func(t *testing.T) {
		bioID := uuid.Must(uuid.NewV7())
		repo.GetByFPVFunc = func(ctx context.Context, id int) (domain.PsiUserModel, error) {
			return domain.PsiUserModel{
				ID:               uuid.Must(uuid.NewV7()),
				FPV:              psiFPV,
				IsActive:         true,
				Solvent:          true, // Pasa al Escudo de Privacidad
				ContactEmail:     "privado@test.com",
				ShowContactEmail: false,
				BioTextID:        bioID,
				ColData:          domain.PsiUserColData{},
			}, nil
		}

		// FIX: El service llama a GetTextContentByID si el usuario es solvente
		repo.GetTextContentByIDFunc = func(ctx context.Context, id uuid.UUID) (string, error) {
			return "Bio content", nil
		}

		profile, _, err := svc.GetPublicProfile(context.Background(), psiFPV)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}
		if profile.Email != "" {
			t.Errorf("El email debería estar oculto, se obtuvo: %s", profile.Email)
		}
	})
}

func TestPsiService_Login(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	mailer := &mockMailSvc{SendEmailFunc: func(to, subject, template string, data any) error { return nil }}
	svc := NewPsiService(repo, nil, mailer)

	pass := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	psiID := uuid.Must(uuid.NewV7())

	t.Run("Login Exitoso", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				ID:       psiID,
				Password: string(hashed),
				IsActive: true,
				Username: "user_test",
			}, nil
		}
		repo.UpdateKeyFunc = func(ctx context.Context, psi *domain.PsiUserModel) error { return nil }

		token, _, err := svc.Login(context.Background(), "user_test", pass)
		if err != nil || token == "" {
			t.Errorf("Fallo login: %v", err)
		}
	})
}

func TestPsiService_UpdateProfileSelf_LazyLoading(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)
	psiID := uuid.Must(uuid.NewV7())
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)

	// Usuario con password para que bcrypt no falle
	psi := &domain.PsiUserModel{
		ID:       psiID,
		Username: "psico_1",
		Password: string(hashed),
	}

	t.Run("Lazy Loading: No llama a ColData si no hay cambios académicos", func(t *testing.T) {
		calledColData := false
		repo.GetPsiUserColDataFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserColData, error) {
			calledColData = true
			return &domain.PsiUserColData{}, nil
		}
		repo.ValidateUniqueCredentialsFunc = func(ctx context.Context, u, e string, ex uuid.UUID) error { return nil }
		repo.UpdatePublicProfileFunc = func(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, t *domain.TextModel) error {
			return nil
		}

		newBio := "Nueva biografía"
		req := request_structs.PsiUserUpdateRequestSelf{
			Password: "pass123", // Password correcta
			MiniBio:  &newBio,
		}

		_, err := svc.UpdateProfileSelf(context.Background(), psi, psi.ID, req, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("Error en UpdateProfileSelf: %v", err)
		}

		if calledColData {
			t.Error("Se llamó a ColData innecesariamente (Lazy Loading falló)")
		}
	})
}

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

func TestPsiService_Logout(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Logout limpia la Key", func(t *testing.T) {
		psiID := uuid.Must(uuid.NewV7())
		psi := &domain.PsiUserModel{
			ID: psiID,
			Credentials: domain.Credentials{
				Username: "psi_test",
				Key:      "old-session-key",
			},
		}

		repo.UpdateKeyFunc = func(ctx context.Context, p *domain.PsiUserModel) error {
			if p.Key != "" {
				t.Error("Key debería estar vacía tras Logout")
			}
			if p.UpdateBy != "psi_test" {
				t.Errorf("UpdateBy = %q, want psi_test", p.UpdateBy)
			}
			return nil
		}

		if err := svc.Logout(context.Background(), psi); err != nil {
			t.Fatalf("Logout error: %v", err)
		}
	})
}

func TestPsiService_Login_CredencialesInvalidas(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	mailer := &mockMailSvc{SendEmailFunc: func(to, subject, template string, data any) error { return nil }}
	svc := NewPsiService(repo, nil, mailer)

	t.Run("Usuario no encontrado", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return nil, domain.ErrPsiNotFound
		}

		_, _, err := svc.Login(context.Background(), "nonexistent", "pass")
		if err == nil || err.Error() != "credenciales inválidas" {
			t.Errorf("Se esperaba error de credenciales, got: %v", err)
		}
	})

	t.Run("Cuenta inactiva", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				Credentials: domain.Credentials{IsActive: false},
			}, nil
		}

		_, _, err := svc.Login(context.Background(), "inactive", "pass")
		if err == nil || err.Error() != "cuenta inactiva o suspendida" {
			t.Errorf("Se esperaba cuenta inactiva, got: %v", err)
		}
	})

	t.Run("Contraseña incorrecta", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				Credentials: domain.Credentials{
					IsActive: true,
					Password: string(hashed),
				},
			}, nil
		}

		_, _, err := svc.Login(context.Background(), "user", "wrong_password")
		if err == nil || err.Error() != "credenciales inválidas" {
			t.Errorf("Se esperaba credenciales inválidas, got: %v", err)
		}
	})
}

func TestPsiService_LoginLibrary_CredencialesInvalidas(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Usuario no encontrado en LoginLibrary", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return nil, domain.ErrPsiNotFound
		}

		_, _, err := svc.LoginLibrary(context.Background(), "ghost", "pass")
		if err == nil || err.Error() != "credenciales inválidas" {
			t.Errorf("Se esperaba credenciales inválidas, got: %v", err)
		}
	})

	t.Run("Cuenta inactiva en LoginLibrary", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				Credentials: domain.Credentials{IsActive: false},
			}, nil
		}

		_, _, err := svc.LoginLibrary(context.Background(), "banned", "pass")
		if err == nil || err.Error() != "cuenta inactiva o suspendida" {
			t.Errorf("Se esperaba cuenta inactiva, got: %v", err)
		}
	})
}

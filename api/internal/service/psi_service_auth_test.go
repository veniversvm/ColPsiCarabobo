package service

import (
	"context"
	"errors"
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

// =========================================================================
// TESTS EXPANDIDOS: CASOS EXTREMOS DE AUTH
// =========================================================================

func TestPsiService_Login_KeyRotation(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	mailer := &mockMailSvc{SendEmailFunc: func(to, subject, template string, data any) error { return nil }}
	svc := NewPsiService(repo, nil, mailer)

	pass := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	psiID := uuid.Must(uuid.NewV7())

	t.Run("Login exitoso rota Key", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				ID: psiID,
				Credentials: domain.Credentials{
					Password: string(hashed),
					IsActive: true,
					Username: "key_rotator",
				},
			}, nil
		}

		var savedKey string
		repo.UpdateKeyFunc = func(ctx context.Context, psi *domain.PsiUserModel) error {
			savedKey = psi.Key
			return nil
		}

		token, psi, err := svc.Login(context.Background(), "key_rotator", pass)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}
		if token == "" {
			t.Error("Token no debería estar vacío")
		}
		if savedKey == "" {
			t.Error("Key debería haberse rotado (no vacía)")
		}
		if psi.ID != psiID {
			t.Error("El usuario retornado debería coincidir")
		}
	})
}

func TestPsiService_Login_UpdateKeyError(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	mailer := &mockMailSvc{SendEmailFunc: func(to, subject, template string, data any) error { return nil }}
	svc := NewPsiService(repo, nil, mailer)

	pass := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	t.Run("Error en UpdateKey retorna error de sistema", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				ID: uuid.Must(uuid.NewV7()),
				Credentials: domain.Credentials{
					Password: string(hashed),
					IsActive: true,
					Username: "user1",
				},
			}, nil
		}

		repo.UpdateKeyFunc = func(ctx context.Context, psi *domain.PsiUserModel) error {
			return errors.New("database connection lost")
		}

		_, _, err := svc.Login(context.Background(), "user1", pass)
		if err == nil || err.Error() != "error de sistema al iniciar sesión" {
			t.Errorf("Se esperaba error de sistema, got: %v", err)
		}
	})
}

func TestPsiService_Login_EmailNotification(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	pass := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	t.Run("Login envía email de notificación", func(t *testing.T) {
		emailSent := false
		var sentTo, sentTemplate string
		mailer := &mockMailSvc{
			SendEmailFunc: func(to, subject, template string, data any) error {
				emailSent = true
				sentTo = to
				sentTemplate = template
				return nil
			},
		}
		svc := NewPsiService(repo, nil, mailer)

		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				ID: uuid.Must(uuid.NewV7()),
				Credentials: domain.Credentials{
					Password: string(hashed),
					IsActive: true,
					Username: "email_test_user",
					Email:    "emailtest@gmail.com",
				},
			}, nil
		}
		repo.UpdateKeyFunc = func(ctx context.Context, psi *domain.PsiUserModel) error { return nil }

		_, _, err := svc.Login(context.Background(), "email_test_user", pass)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}
		if !emailSent {
			t.Error("Se esperaba que se enviara un email de notificación")
		}
		if sentTo != "emailtest@gmail.com" {
			t.Errorf("Email enviado a %q, want emailtest@gmail.com", sentTo)
		}
		if sentTemplate != "login_psi" {
			t.Errorf("Template = %q, want login_psi", sentTemplate)
		}
	})

	t.Run("Error de email no bloquea login", func(t *testing.T) {
		mailer := &mockMailSvc{
			SendEmailFunc: func(to, subject, template string, data any) error {
				return errors.New("smtp timeout")
			},
		}
		svc := NewPsiService(repo, nil, mailer)

		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				ID: uuid.Must(uuid.NewV7()),
				Credentials: domain.Credentials{
					Password: string(hashed),
					IsActive: true,
					Username: "resilient_user",
				},
			}, nil
		}
		repo.UpdateKeyFunc = func(ctx context.Context, psi *domain.PsiUserModel) error { return nil }

		token, _, err := svc.Login(context.Background(), "resilient_user", pass)
		if err != nil {
			t.Errorf("Login debería continuar a pesar de error de email, got: %v", err)
		}
		if token == "" {
			t.Error("Token no debería estar vacío")
		}
	})
}

func TestPsiService_Logout_AuditTrail(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)

	t.Run("Logout establece audit trail correcto", func(t *testing.T) {
		psiID := uuid.Must(uuid.NewV7())
		psi := &domain.PsiUserModel{
			ID: psiID,
			Credentials: domain.Credentials{
				Username: "audit_user",
				Key:      "session-to-invalidate",
			},
		}

		var capturedUpdateBy string
		var capturedUpdateByID *uuid.UUID
		repo.UpdateKeyFunc = func(ctx context.Context, p *domain.PsiUserModel) error {
			capturedUpdateBy = p.UpdateBy
			capturedUpdateByID = p.UpdateById
			return nil
		}

		err := svc.Logout(context.Background(), psi)
		if err != nil {
			t.Fatalf("Logout error: %v", err)
		}
		if capturedUpdateBy != "audit_user" {
			t.Errorf("UpdateBy = %q, want audit_user", capturedUpdateBy)
		}
		if capturedUpdateByID == nil || *capturedUpdateByID != psiID {
			t.Error("UpdateById debería apuntar al ID del usuario")
		}
	})
}

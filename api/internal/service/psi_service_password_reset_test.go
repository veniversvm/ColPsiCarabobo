package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// RECUPERACIÓN DE CONTRASEÑA — TESTS UNITARIOS
// =========================================================================

func newResetMailer() *mockMailSvc {
	return &mockMailSvc{SendEmailFunc: func(to, subject, template string, data any) error { return nil }}
}

func activePsi() *domain.PsiUserModel {
	return &domain.PsiUserModel{
		ID: uuid.Must(uuid.NewV7()),
		Credentials: domain.Credentials{
			Username: "psico_test",
			Email:    "psico@test.com",
			IsActive: true,
		},
		FirstName: "Ana",
		LastName:  "Pérez",
	}
}

func TestPsiService_RequestPasswordReset(t *testing.T) {
	t.Run("genera token, lo guarda hasheado y envía correo", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		mailer := newResetMailer()
		svc := NewPsiService(repo, nil, mailer)

		psi := activePsi()
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return psi, nil
		}

		var savedToken *domain.PsiPasswordResetToken
		repo.CreateResetTokenFunc = func(ctx context.Context, token *domain.PsiPasswordResetToken) error {
			savedToken = token
			return nil
		}

		sentSubject := ""
		var sentData map[string]interface{}
		mailer.SendEmailFunc = func(to, subject, template string, data any) error {
			sentSubject = subject
			sentData = data.(map[string]interface{})
			return nil
		}

		if err := svc.RequestPasswordReset(context.Background(), psi.Email); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if savedToken == nil {
			t.Fatal("no se persistió ningún token")
		}

		// El token guardado es el HASH, no el token en claro.
		if savedToken.TokenHash == "" || len(savedToken.TokenHash) != 64 {
			t.Errorf("el token almacenado debe ser el hash sha256 (64 hex), se guardó %q", savedToken.TokenHash)
		}
		if savedToken.UsedAt != nil {
			t.Error("el token recién creado no puede estar usado")
		}
		if !savedToken.ExpiresAt.After(time.Now()) {
			t.Error("el token debe tener expiración futura (~1h)")
		}

		// El correo debe llevar el enlace con el token en claro (única vez que viaja).
		if sentSubject == "" {
			t.Error("no se envió correo")
		}
		resetURL, _ := sentData["ResetURL"].(string)
		if !strings.Contains(resetURL, "/reset-password?token=") {
			t.Errorf("el correo no contiene un enlace de recuperación válido: %q", resetURL)
		}
		if strings.Contains(resetURL, savedToken.TokenHash) {
			t.Error("el enlace expone el hash en lugar del token en claro")
		}
	})

	t.Run("no revela que el email no existe (respuesta genérica)", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, newResetMailer())

		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return nil, domain.ErrPsiNotFound
		}
		repo.CreateResetTokenFunc = func(ctx context.Context, token *domain.PsiPasswordResetToken) error {
			t.Fatal("no debe crear tokens para emails inexistentes")
			return nil
		}

		if err := svc.RequestPasswordReset(context.Background(), "nadie@test.com"); err != nil {
			t.Fatalf("debe responder éxito genérico, obtuvo: %v", err)
		}
	})

	t.Run("no genera token para cuentas inactivas", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, newResetMailer())

		inactive := activePsi()
		inactive.IsActive = false
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return inactive, nil
		}
		repo.CreateResetTokenFunc = func(ctx context.Context, token *domain.PsiPasswordResetToken) error {
			t.Fatal("no debe crear tokens para cuentas inactivas")
			return nil
		}

		if err := svc.RequestPasswordReset(context.Background(), inactive.Email); err != nil {
			t.Fatalf("debe responder éxito genérico, obtuvo: %v", err)
		}
	})
}

func TestPsiService_HashToken(t *testing.T) {
	token := "my-secret-token"
	h1 := hashToken(token)
	h2 := hashToken(token)

	if len(h1) != 64 {
		t.Errorf("hash debe tener 64 hex chars, tiene %d", len(h1))
	}
	if h1 != h2 {
		t.Error("hash debe ser determinístico")
	}
	if hashToken("other") == h1 {
		t.Error("tokens distintos no pueden colisionar")
	}
}

func TestPsiService_ResetPasswordWithToken(t *testing.T) {
	hashedOld, _ := bcrypt.GenerateFromPassword([]byte("vieja123"), bcrypt.DefaultCost)

	buildToken := func(used bool, expiresAt time.Time) *domain.PsiPasswordResetToken {
		psi := activePsi()
		psi.Password = string(hashedOld)
		psi.Key = "old-session-key"
		return &domain.PsiPasswordResetToken{
			ID:        uuid.Must(uuid.NewV7()),
			PsiID:     psi.ID,
			ExpiresAt: expiresAt,
			Psi:       *psi,
		}
	}

	t.Run("reset exitoso: rota key, nueva contraseña, must_change=false y marca usado", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, nil)

		token := buildToken(false, time.Now().Add(time.Hour))
		repo.GetResetTokenByHashFunc = func(ctx context.Context, h string) (*domain.PsiPasswordResetToken, error) {
			return token, nil
		}

		var updatedPSI *domain.PsiUserModel
		repo.ResetPasswordFunc = func(ctx context.Context, psi *domain.PsiUserModel) error {
			updatedPSI = psi
			return nil
		}

		var markedID uuid.UUID
		repo.MarkResetTokenUsedFunc = func(ctx context.Context, id uuid.UUID) error {
			markedID = id
			return nil
		}

		if err := svc.ResetPasswordWithToken(context.Background(), "raw-token", "NuevaClave123!"); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if updatedPSI == nil {
			t.Fatal("no se actualizó el psicólogo")
		}
		if updatedPSI.Password == string(hashedOld) {
			t.Error("la contraseña no fue re-hasheada")
		}
		if updatedPSI.Key == "old-session-key" || updatedPSI.Key == "" {
			t.Error("la Key de sesión debe rotarse (invalidar JWTs previos)")
		}
		if updatedPSI.MustChangePassword {
			t.Error("must_change_password debe apagarse tras fijar la nueva clave")
		}
		if bcrypt.CompareHashAndPassword([]byte(updatedPSI.Password), []byte("NuevaClave123!")) != nil {
			t.Error("el hash guardado no corresponde a la nueva contraseña")
		}
		if markedID != token.ID {
			t.Error("el token debe marcarse como usado (single-use)")
		}
	})

	t.Run("rechaza token inexistente", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, nil)
		repo.GetResetTokenByHashFunc = func(ctx context.Context, h string) (*domain.PsiPasswordResetToken, error) {
			return nil, nil
		}

		if err := svc.ResetPasswordWithToken(context.Background(), "raw-token", "NuevaClave123!"); err == nil {
			t.Error("debe rechazar un token desconocido")
		}
	})

	t.Run("rechaza token expirado", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, nil)
		token := buildToken(false, time.Now().Add(-time.Minute))
		repo.GetResetTokenByHashFunc = func(ctx context.Context, h string) (*domain.PsiPasswordResetToken, error) {
			return token, nil
		}

		if err := svc.ResetPasswordWithToken(context.Background(), "raw-token", "NuevaClave123!"); err == nil {
			t.Error("debe rechazar un token expirado")
		}
	})

	t.Run("rechaza token ya usado", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, nil)
		token := buildToken(false, time.Now().Add(time.Hour))
		now := time.Now()
		token.UsedAt = &now
		repo.GetResetTokenByHashFunc = func(ctx context.Context, h string) (*domain.PsiPasswordResetToken, error) {
			return token, nil
		}

		if err := svc.ResetPasswordWithToken(context.Background(), "raw-token", "NuevaClave123!"); err == nil {
			t.Error("debe rechazar un token ya consumido")
		}
	})

	t.Run("rechaza contraseña débil (menos de 8)", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, nil)

		if err := svc.ResetPasswordWithToken(context.Background(), "raw-token", "123"); err == nil {
			t.Error("debe rechazar contraseñas cortas")
		}
	})
}
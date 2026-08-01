package service

import (
	"testing"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

// TestNewResendMailService_RequiresAPIKey verifica el failsafe de arranque:
// sin RESEND_API_KEY el transporte no debe inicializarse.
func TestNewResendMailService_RequiresAPIKey(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	config.InitConfig()

	if _, err := NewResendMailService(); err == nil {
		t.Fatal("se esperaba un error sin RESEND_API_KEY")
	}
}

// TestResendMailService_InitialState verifica el estado inicial del servicio
// sin encolar nada (el worker queda bloqueado en la cola vacía: sin red).
func TestResendMailService_InitialState(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "re_testkey")
	config.InitConfig()

	ms, err := NewResendMailService()
	if err != nil {
		t.Fatalf("NewResendMailService() failed: %v", err)
	}
	defer ms.Close()

	if ms.client == nil {
		t.Error("client should not be nil")
	}
	if ms.from == "" {
		t.Error("from should not be empty")
	}
	if ms.queue == nil {
		t.Error("queue should not be nil")
	}
	if ms.cancel == nil {
		t.Error("cancel should not be nil")
	}

	// Tras Close, SendEmail debe fallar (canal cerrado).
	ms.Close()
	if err := ms.SendEmail("test@example.com", "test", "welcome", nil); err == nil {
		t.Error("SendEmail after Close() should have returned an error")
	}
}

// TestResendMailService_SkipsPlaceholderAndTestRecipients verifica que los
// destinatarios placeholder ("sincorreo") y los de prueba ("test") NO se
// encolen, para preservar la reputación del dominio en Resend.
func TestResendMailService_SkipsPlaceholderAndTestRecipients(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "re_testkey")
	config.InitConfig()

	// Construcción directa (sin worker) para inspeccionar la cola sin carreras.
	ms := &ResendMailService{queue: make(chan MailJob, resendQueueBuffer)}
	defer ms.Close()

	cases := []struct {
		name string
		to   string
	}{
		{"placeholder sincorreo", "juan.sincorreo@gmail.com"},
		{"sincorreo como dominio", "psicologo@sincorreo.com"},
		{"test en local part", "maria.test@dominio.com"},
		{"test en dominio", "maria@test.colpsi.local"},
		{"test mayúsculas", "MARIA@TEST.COLPSI.LOCAL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ms.SendEmail(tc.to, "subject", "welcome", nil); err != nil {
				t.Fatalf("SendEmail(%q) returned error: %v", tc.to, err)
			}
			select {
			case job := <-ms.queue:
				t.Fatalf("correo a %q no debió encolarse (encontrado en cola hacia %q)", tc.to, job.To)
			default:
				// Esperado: cola vacía.
			}
		})
	}

	// Control positivo: un correo real sí debe encolarse.
	if err := ms.SendEmail("maria.real@gmail.com", "subject", "welcome", nil); err != nil {
		t.Fatalf("SendEmail(real) returned error: %v", err)
	}
	select {
	case job := <-ms.queue:
		if job.To != "maria.real@gmail.com" {
			t.Fatalf("correo real encolado con destinatario inesperado: %q", job.To)
		}
	default:
		t.Fatal("correo real debió encolarse")
	}
}

// TestResendMailService_IsSkippableRecipient prueba la helper de filtrado.
func TestResendMailService_IsSkippableRecipient(t *testing.T) {
	cases := []struct {
		to   string
		want bool
	}{
		{"juan.sincorreo@gmail.com", true},
		{"psicologo@sincorreo.com", true},
		{"maria.test@dominio.com", true},
		{"maria@test.colpsi.local", true},
		{"MARIA@TEST.COLPSI.LOCAL", true},
		{"maria.real@gmail.com", false},
		{"psicologo@hotmail.com", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isSkippableRecipient(tc.to); got != tc.want {
			t.Errorf("isSkippableRecipient(%q) = %v, se esperaba %v", tc.to, got, tc.want)
		}
	}
}

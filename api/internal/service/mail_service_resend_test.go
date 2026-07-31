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

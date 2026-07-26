package service

import (
	"testing"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

func TestMailService_Close_ShutdownsWorker(t *testing.T) {
	config.InitConfig()

	ms, err := NewMailService()
	if err != nil {
		t.Fatalf("NewMailService() failed: %v", err)
	}

	// Verify the worker is running by enqueuing a job (it will fail to send
	// but the worker should process it without panicking).
	_ = ms.SendEmail("test@example.com", "test", "welcome", nil)

	// Give the worker time to pick up the job
	time.Sleep(200 * time.Millisecond)

	// Close — should cancel context and close channel without panic
	ms.Close()

	// After Close, SendEmail should fail because the channel is closed
	// Give a tiny window for close to propagate
	time.Sleep(10 * time.Millisecond)

	err = ms.SendEmail("test@example.com", "test", "welcome", nil)
	if err == nil {
		t.Error("SendEmail after Close() should have returned an error")
	}
}

func TestMailService_NewMailService_InitialState(t *testing.T) {
	config.InitConfig()

	ms, err := NewMailService()
	if err != nil {
		t.Fatalf("NewMailService() failed: %v", err)
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
}

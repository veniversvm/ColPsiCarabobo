package job

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIsKeyExpired(t *testing.T) {
	cutoff := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		key     string
		cutoff  time.Time
		want    bool
	}{
		{
			name:   "Key v7 reciente (no expirada)",
			key:    uuid.Must(uuid.NewV7()).String(),
			cutoff: cutoff,
			want:   false,
		},
		{
			name:    "Key inválida (expirada por defecto)",
			key:     "not-a-uuid",
			cutoff:  cutoff,
			want:    true,
		},
		{
			name:    "Key vacía (expirada por defecto)",
			key:     "",
			cutoff:  cutoff,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isKeyExpired(tt.key, tt.cutoff)
			if got != tt.want {
				t.Errorf("isKeyExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsKeyExpired_OldKey(t *testing.T) {
	key := uuid.Must(uuid.NewV7()).String()
	cutoff := time.Now().Add(1 * time.Hour) // cutoff en el futuro = key es "vieja"
	if !isKeyExpired(key, cutoff) {
		t.Errorf("Expected key to be expired with future cutoff, got false")
	}
}

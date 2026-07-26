package job

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// makeV7FromTime manually constructs a UUID v7 string with the given timestamp.
// This is needed because uuid.NewV7FromTime doesn't exist in uuid v1.6.0.
func makeV7FromTime(t time.Time) string {
	milli := t.UnixMilli()
	u := uuid.New()
	// Set version 7
	u[6] = 0x70 | (u[6] & 0x0F)
	u[7] = u[7] // random
	// Encode 48-bit milliseconds
	u[0] = byte(milli >> 40)
	u[1] = byte(milli >> 32)
	u[2] = byte(milli >> 24)
	u[3] = byte(milli >> 16)
	u[4] = byte(milli >> 8)
	u[5] = byte(milli)
	return u.String()
}

func TestIsKeyExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		key      string
		cutoff   time.Time
		expected bool
	}{
		{
			name:     "valid v7 key created now is NOT expired",
			key:      makeV7FromTime(now),
			cutoff:   now.Add(-1 * time.Hour),
			expected: false,
		},
		{
			name:     "valid v7 key created 2h ago IS expired with 1h cutoff",
			key:      makeV7FromTime(now.Add(-2 * time.Hour)),
			cutoff:   now.Add(-1 * time.Hour),
			expected: true,
		},
		{
			name:     "valid v7 key created 30m ago NOT expired with 1h cutoff",
			key:      makeV7FromTime(now.Add(-30 * time.Minute)),
			cutoff:   now.Add(-1 * time.Hour),
			expected: false,
		},
		{
			name:     "malformed key returns expired (true)",
			key:      "not-a-uuid",
			cutoff:   now.Add(-1 * time.Hour),
			expected: true,
		},
		{
			name:     "empty key returns expired (true)",
			key:      "",
			cutoff:   now.Add(-1 * time.Hour),
			expected: true,
		},
		{
			name:     "key created exactly at cutoff boundary",
			key:      makeV7FromTime(now.Add(-1 * time.Hour)),
			cutoff:   now.Add(-1 * time.Hour),
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isKeyExpired(tc.key, tc.cutoff)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestKeyCleanupResult_ZeroValues(t *testing.T) {
	result := KeyCleanupResult{}
	require.Equal(t, int64(0), result.AdminsCleaned)
	require.Equal(t, int64(0), result.PsiCleaned)
}

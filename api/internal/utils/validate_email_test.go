package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantEmail string
		wantErr   bool
	}{
		{
			name:      "valid email lowercase",
			input:     "user@test.com",
			wantEmail: "user@test.com",
			wantErr:   false,
		},
		{
			name:      "valid email uppercase canonicalized",
			input:     "USER@TEST.COM",
			wantEmail: "user@test.com",
			wantErr:   false,
		},
		{
			name:      "valid email with display name",
			input:     "Fran <fran@colpsi.com>",
			wantEmail: "fran@colpsi.com",
			wantErr:   false,
		},
		{
			name:      "valid email with surrounding whitespace",
			input:     "  user@test.com  ",
			wantEmail: "user@test.com",
			wantErr:   false,
		},
		{
			name:      "valid email with tabs",
			input:	"\tuser@test.com\t",
			wantEmail: "user@test.com",
			wantErr:   false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid format no at sign",
			input:   "user_test.com",
			wantErr: true,
		},
		{
			name:    "invalid format no domain",
			input:   "user@",
			wantErr: true,
		},
		{
			name:    "invalid format no local part",
			input:   "@test.com",
			wantErr: true,
		},
		{
			name:    "invalid format double dots",
			input:   "user@te..st.com",
			wantErr: true,
		},
		{
			name:    "sql injection attempt",
			input:   "'; DROP TABLE users; --",
			wantErr: true,
		},
		{
			name:    "xss attempt",
			input:   "<script>alert(1)</script>@test.com",
			wantErr: true,
		},
		{
			name:      "valid unicode local part",
			input:     "ñoño@tëst.com",
			wantEmail: "ñoño@tëst.com",
			wantErr:   false,
		},
		{
			name:    "very long local part",
			input:   string(make([]byte, 300)) + "@test.com",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseAndValidateEmail(tc.input)
			if tc.wantErr {
				require.Error(t, err, "expected error for input: %s", tc.input)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantEmail, got)
			}
		})
	}
}

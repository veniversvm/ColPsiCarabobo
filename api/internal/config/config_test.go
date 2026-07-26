package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		envValue string
		setEnv   bool
		want     string
	}{
		{
			name:     "env var set returns its value",
			key:      "TEST_GETENV_SET",
			fallback: "default_val",
			envValue: "custom_val",
			setEnv:   true,
			want:     "custom_val",
		},
		{
			name:     "env var not set returns fallback",
			key:      "TEST_GETENV_UNSET_XYZ_ABC",
			fallback: "default_val",
			setEnv:   false,
			want:     "default_val",
		},
		{
			name:     "empty env var returns empty string (not fallback)",
			key:      "TEST_GETENV_EMPTY",
			fallback: "default_val",
			envValue: "",
			setEnv:   true,
			want:     "",
		},
		{
			name:     "empty fallback with unset env returns empty",
			key:      "TEST_GETENV_UNSET_EMPTY_FALLBACK",
			fallback: "",
			setEnv:   false,
			want:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(tc.key, tc.envValue)
			}
			got := getEnv(tc.key, tc.fallback)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestGetEnv_OverwriteExisting(t *testing.T) {
	key := "TEST_GETENV_OVERWRITE_EXISTING"
	t.Setenv(key, "first")
	require.Equal(t, "first", getEnv(key, "fallback"))

	os.Setenv(key, "second")
	require.Equal(t, "second", getEnv(key, "fallback"))
}

func TestInitConfig_DefaultValues(t *testing.T) {
	InitConfig()

	require.NotNil(t, Envs)
	require.NotEmpty(t, Envs.Port, "Port should have a default")
	require.NotEmpty(t, Envs.DBHost, "DBHost should have a default")
	require.NotEmpty(t, Envs.DBPort, "DBPort should have a default")
	require.NotEmpty(t, Envs.DBUser, "DBUser should have a default")
	require.NotEmpty(t, Envs.DBName, "DBName should have a default")
	require.NotEmpty(t, Envs.SMTPHost, "SMTPHost should have a default")
	require.NotEmpty(t, Envs.S3Region, "S3Region should have a default")
	require.NotEmpty(t, Envs.S3Endpoint, "S3Endpoint should have a default")
	require.NotEmpty(t, Envs.AllowedOrigins, "AllowedOrigins should have a default")
	require.Greater(t, Envs.SMTPPort, 0, "SMTPPort should be a positive integer")
}

func TestInitConfig_EnvironmentOverride(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("APP_ENV", "staging")
	t.Setenv("DB_HOST", "remote-db.test")

	InitConfig()

	require.Equal(t, "9999", Envs.Port)
	require.Equal(t, "staging", Envs.Environment)
	require.Equal(t, "remote-db.test", Envs.DBHost)
}

func TestInitConfig_InvalidSMTPPort(t *testing.T) {
	t.Setenv("SMTP_PORT", "not_a_number")
	InitConfig()
	require.NotNil(t, Envs)
}

func TestInitConfig_SpecialValues(t *testing.T) {
	t.Setenv("VALKEY_ADDR", "valkey.internal:6379")
	t.Setenv("JWT_LIBRARY_SECRET", "super-secret-jwt-key")
	t.Setenv("ABS_ADMIN_TOKEN", "abs-admin-token-123")

	InitConfig()

	require.Equal(t, "valkey.internal:6379", Envs.ValkeyAddr)
	require.Equal(t, "super-secret-jwt-key", Envs.JwtLibrarySecret)
	require.Equal(t, "abs-admin-token-123", Envs.AbsAdminToken)
}

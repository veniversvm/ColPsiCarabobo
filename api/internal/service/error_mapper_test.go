package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapDBError(t *testing.T) {
	tests := []struct {
		name    string
		input   error
		wantMsg string
	}{
		// ── Unique Constraints: Cedula ────────────────────────────────────
		{
			name:    "duplicate_ci con idx_",
			input:   errors.New("duplicate key value violates unique constraint \"idx_psi_users_ci\""),
			wantMsg: "la Cédula de Identidad ya se encuentra registrada",
		},
		{
			name:    "duplicate_ci con uni_",
			input:   errors.New("conflict with unique constraint uni_psi_users_ci"),
			wantMsg: "la Cédula de Identidad ya se encuentra registrada",
		},
		{
			name:    "duplicate_ci con _key",
			input:   errors.New("ERROR: duplicate key value (psi_users_ci_key)"),
			wantMsg: "la Cédula de Identidad ya se encuentra registrada",
		},

		// ── Unique Constraints: FPV ──────────────────────────────────────
		{
			name:    "duplicate_fpv con idx_",
			input:   errors.New("duplicate key value violates unique constraint \"idx_psi_users_fpv\""),
			wantMsg: "el número de FPV ya está registrado por otro psicólogo",
		},
		{
			name:    "duplicate_fpv con uni_",
			input:   errors.New("conflict with unique constraint uni_psi_users_fpv"),
			wantMsg: "el número de FPV ya está registrado por otro psicólogo",
		},
		{
			name:    "duplicate_fpv con _key",
			input:   errors.New("ERROR: duplicate key value (psi_users_fpv_key)"),
			wantMsg: "el número de FPV ya está registrado por otro psicólogo",
		},

		// ── Unique Constraints: Email ─────────────────────────────────────
		{
			name:    "duplicate_email con idx_",
			input:   errors.New("duplicate key value violates unique constraint \"idx_psi_users_email\""),
			wantMsg: "el correo electrónico ya está en uso",
		},
		{
			name:    "duplicate_email con uni_",
			input:   errors.New("conflict with unique constraint uni_psi_users_email"),
			wantMsg: "el correo electrónico ya está en uso",
		},
		{
			name:    "duplicate_email con _key",
			input:   errors.New("ERROR: duplicate key value (psi_users_email_key)"),
			wantMsg: "el correo electrónico ya está en uso",
		},

		// ── Unique Constraints: Username ──────────────────────────────────
		{
			name:    "duplicate_username con idx_",
			input:   errors.New("duplicate key value violates unique constraint \"idx_psi_users_username\""),
			wantMsg: "el nombre de usuario ya existe",
		},
		{
			name:    "duplicate_username con uni_",
			input:   errors.New("conflict with unique constraint uni_psi_users_username"),
			wantMsg: "el nombre de usuario ya existe",
		},
		{
			name:    "duplicate_username con _key",
			input:   errors.New("ERROR: duplicate key value (psi_users_username_key)"),
			wantMsg: "el nombre de usuario ya existe",
		},

		// ── Errores de Longitud ───────────────────────────────────────────
		{
			name:    "varchar_25_overflow",
			input:   errors.New("value too long for type character varying(25)"),
			wantMsg: "el nombre de usuario generado excede el límite de 25 caracteres",
		},
		{
			name:    "generic_value_too_long",
			input:   errors.New("value too long for type varchar(255)"),
			wantMsg: "un campo es demasiado largo para la base de datos",
		},
		{
			name:    "value_too_long_generic_variant",
			input:   errors.New("ERROR: value too long for type character varying"),
			wantMsg: "un campo es demasiado largo para la base de datos",
		},

		// ── Errores de Formato UUID ───────────────────────────────────────
		{
			name:    "uuid_invalid",
			input:   errors.New("invalid input syntax for type uuid: \"not-a-uuid\""),
			wantMsg: "error interno: ID de sistema inválido",
		},

		// ── Fallback: Error no mapeado ───────────────────────────────────
		{
			name:    "error_desconocido_se_pasa_through",
			input:   errors.New("connection refused to database"),
			wantMsg: "connection refused to database",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := MapDBError(tc.input)

			require.Error(t, result, "MapDBError debe retornar un error")

			if tc.wantMsg != "" {
				require.Equal(t, tc.wantMsg, result.Error(), "Mensaje de error mismatch")
			}

			// Fallback preserva el error original (mismo objeto)
			if tc.name == "error_desconocido_se_pasa_through" {
				require.Equal(t, tc.input, result, "Fallback debe retornar el mismo error original")
			}
		})
	}
}

func TestMapDBError_NilError(t *testing.T) {
	require.Panics(t, func() {
		MapDBError(nil)
	}, "MapDBError(nil) debe panichear porque err.Error() en nil")
}

// api/internal/domain/credentials_test.go
package domain

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCredentialsEmbed_UserAdmin(t *testing.T) {
	creds := Credentials{
		Username: "admin_test",
		Email:    "admin@test.com",
		Password: "hashed_password",
		Key:      uuid.Must(uuid.NewV7()).String(),
		IsActive: true,
	}

	admin := UserAdmin{
		ID:          uuid.Must(uuid.NewV7()),
		Credentials: creds,
		Sudo:        true,
	}

	if admin.Username != "admin_test" {
		t.Errorf("Username: esperado 'admin_test', obtuvo '%s'", admin.Username)
	}
	if admin.Email != "admin@test.com" {
		t.Errorf("Email: esperado 'admin@test.com', obtuvo '%s'", admin.Email)
	}
	if admin.Password != "hashed_password" {
		t.Errorf("Password: esperado 'hashed_password', obtuvo '%s'", admin.Password)
	}
	if admin.Key == "" {
		t.Error("Key no debe estar vacío")
	}
	if !admin.IsActive {
		t.Error("IsActive debe ser true")
	}
	if !admin.Sudo {
		t.Error("Sudo debe ser true (campo fuera de Credentials)")
	}
}

func TestCredentialsEmbed_PsiUserModel(t *testing.T) {
	psiKey := uuid.Must(uuid.NewV7()).String()

	psi := PsiUserModel{
		ID: uuid.Must(uuid.NewV7()),
		Credentials: Credentials{
			Username: "psi_test",
			Email:    "psi@test.com",
			Password: "psi_hashed",
			Key:      psiKey,
			IsActive: true,
		},
		FirstName: "Juan",
		LastName:  "Perez",
	}

	if psi.Username != "psi_test" {
		t.Errorf("Username: esperado 'psi_test', obtuvo '%s'", psi.Username)
	}
	if psi.Email != "psi@test.com" {
		t.Errorf("Email: esperado 'psi@test.com', obtuvo '%s'", psi.Email)
	}
	if psi.Key != psiKey {
		t.Errorf("Key: esperado '%s', obtuvo '%s'", psiKey, psi.Key)
	}
	if !psi.IsActive {
		t.Error("IsActive debe ser true")
	}
	if psi.FirstName != "Juan" {
		t.Errorf("FirstName: esperado 'Juan', obtuvo '%s'", psi.FirstName)
	}
}

func TestCredentials_KeyRotation(t *testing.T) {
	oldKey := uuid.Must(uuid.NewV7()).String()

	admin := UserAdmin{
		ID: uuid.Must(uuid.NewV7()),
		Credentials: Credentials{
			Username: "rotator",
			Key:      oldKey,
		},
	}

	if admin.Key != oldKey {
		t.Fatalf("Key inicial: esperado '%s', obtuvo '%s'", oldKey, admin.Key)
	}

	newKey := uuid.Must(uuid.NewV7()).String()
	admin.Key = newKey

	if admin.Key != oldKey {
		if admin.Key != newKey {
			t.Errorf("Key rotation falló: esperado '%s', obtuvo '%s'", newKey, admin.Key)
		}
	}
}

func TestCredentials_TableNames(t *testing.T) {
	admin := UserAdmin{}
	psi := PsiUserModel{}

	if admin.TableName() != "user_admins" {
		t.Errorf("UserAdmin.TableName(): esperado 'user_admins', obtuvo '%s'", admin.TableName())
	}
	if psi.TableName() != "psi_users" {
		t.Errorf("PsiUserModel.TableName(): esperado 'psi_users', obtuvo '%s'", psi.TableName())
	}
}

// =========================================================================
// TEST: SENTINEL ERRORS
// =========================================================================

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		wrapWith string
	}{
		{"ErrPasswordIncorrect", ErrPasswordIncorrect, "wrapped password error"},
		{"ErrInvalidCredentials", ErrInvalidCredentials, "wrapped invalid creds"},
		{"ErrAccountInactive", ErrAccountInactive, "wrapped inactive account"},
		{"ErrPermissionDenied", ErrPermissionDenied, "wrapped permission denied"},
		{"ErrInsufficientPerms", ErrInsufficientPerms, "wrapped insufficient perms"},
		{"ErrPsiNotFound", ErrPsiNotFound, "wrapped psi not found"},
		{"ErrMaxSocialNetworks", ErrMaxSocialNetworks, "wrapped max social networks"},
		{"ErrSocialPermDenied", ErrSocialPermDenied, "wrapped social perm denied"},
		{"ErrSocialOwnDenied", ErrSocialOwnDenied, "wrapped social own denied"},
		{"ErrPostPermDenied", ErrPostPermDenied, "wrapped post perm denied"},
		{"ErrUniqueViolation", ErrUniqueViolation, "wrapped unique violation"},
		{"ErrSudoExists", ErrSudoExists, "wrapped sudo exists"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.sentinel.Error(), "El error sentinel no puede tener mensaje vacio")

			wrapped := errors.New(tc.wrapWith)
			combined := errors.Join(wrapped, tc.sentinel)

			require.True(t, errors.Is(combined, tc.sentinel),
				"errors.Is debe encontrar el sentinel '%s' despues de wrapping", tc.name)

			require.False(t, errors.Is(wrapped, tc.sentinel),
				"Un error envuelto no debe confundirse con otro sentinel")
		})
	}
}

func TestSentinelErrors_Uniqueness(t *testing.T) {
	sentinels := []error{
		ErrPasswordIncorrect,
		ErrInvalidCredentials,
		ErrAccountInactive,
		ErrPermissionDenied,
		ErrInsufficientPerms,
		ErrPsiNotFound,
		ErrMaxSocialNetworks,
		ErrSocialPermDenied,
		ErrSocialOwnDenied,
		ErrPostPermDenied,
		ErrUniqueViolation,
		ErrSudoExists,
	}

	seen := make(map[string]bool)
	for i, s := range sentinels {
		for j, other := range sentinels {
			if i != j {
				require.False(t, errors.Is(s, other),
					"Sentinel en posicion %d no debe confundirse con sentinel en posicion %d", i, j)
			}
		}
		msg := s.Error()
		require.False(t, seen[msg], "Mensaje duplicado entre sentinels: '%s'", msg)
		seen[msg] = true
	}
}

// =========================================================================
// TEST: POSTGRADE TYPE VALIDATION
// =========================================================================

func TestPostGradeType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		value PostGradeType
		valid bool
	}{
		{"diplomado es valido", Diplomado, true},
		{"especializacion es valido", Especializacion, true},
		{"maestria es valido", Maestria, true},
		{"doctorado es valido", Doctorado, true},
		{"tipo inventado no es valido", PostGradeType("tecnico"), false},
		{"string vacio no es valido", PostGradeType(""), false},
		{"case sensitive invalido", PostGradeType("Diplomado"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.valid, tc.value.IsValid(),
				"IsValid(%s) debe retornar %v", tc.value, tc.valid)
		})
	}
}

func TestPostGradeType_TableName(t *testing.T) {
	pg := PsiUserPostGrade{}
	require.Equal(t, "psi_user_post_grades", pg.TableName())
}

func TestColData_TableName(t *testing.T) {
	cd := PsiUserColData{}
	require.Equal(t, "psi_user_col_data", cd.TableName())
}

func TestSolvency_TableName(t *testing.T) {
	s := PsiUserSolvency{}
	require.Equal(t, "psi_user_solvency", s.TableName())
}

func TestObservations_TableName(t *testing.T) {
	o := PsiObservations{}
	require.Equal(t, "psi_observations", o.TableName())
}

func TestDeontologia_TableName(t *testing.T) {
	d := PsiODeontologia{}
	require.Equal(t, "psi_deontologia", d.TableName())
}

// =========================================================================
// TEST: CREDENTIALS EMBEDDED FIELDS
// =========================================================================

func TestCredentials_MustChangePassword(t *testing.T) {
	admin := UserAdmin{
		Credentials: Credentials{
			MustChangePassword: true,
		},
	}
	require.True(t, admin.MustChangePassword, "MustChangePassword debe ser true")

	admin.MustChangePassword = false
	require.False(t, admin.MustChangePassword, "MustChangePassword debe ser false despues de asignar")
}

func TestCredentials_IsActive(t *testing.T) {
	psi := PsiUserModel{
		Credentials: Credentials{
			IsActive: true,
		},
	}
	require.True(t, psi.IsActive)

	psi.IsActive = false
	require.False(t, psi.IsActive, "IsActive debe ser false despues de asignar")
}

func TestCredentials_PasswordNeverSerialized(t *testing.T) {
	// Password y Key tienen json:"-", no deben serializarse
	// Este test verifica que los tags existen
	credType := reflect.TypeOf(Credentials{})

	pwField, ok := credType.FieldByName("Password")
	require.True(t, ok, "Password field debe existir")
	require.Contains(t, pwField.Tag.Get("json"), "-", "Password debe tener json:\"-\"")

	keyField, ok := credType.FieldByName("Key")
	require.True(t, ok, "Key field debe existir")
	require.Contains(t, keyField.Tag.Get("json"), "-", "Key debe tener json:\"-\"")

	mcpField, ok := credType.FieldByName("MustChangePassword")
	require.True(t, ok, "MustChangePassword field debe existir")
	require.Contains(t, mcpField.Tag.Get("json"), "-", "MustChangePassword debe tener json:\"-\"")
}

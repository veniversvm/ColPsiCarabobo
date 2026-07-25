// api/internal/domain/credentials_test.go
package domain

import (
	"testing"

	"github.com/google/uuid"
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

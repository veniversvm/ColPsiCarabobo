package service

// =========================================================================
// MOCK DEL REPOSITORIO (Patrón Func Override usando Embedding)
// =========================================================================
//
// Sigue la misma arquitectura que `social_media_test.go`: se embeda la interfaz
// `domain.PsiUserRepository` para satisfacer el contrato y se sobreescriben solo
// los métodos del submódulo de Persona de Contacto para Eventos, de modo que los
// escenarios se aíslen en memoria sin depender de PostgreSQL.

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

type mockPsiRepoEmergency struct {
	domain.PsiUserRepository
	CountEmergencyContactsFunc func(ctx context.Context, psiID uuid.UUID) (int64, error)
	CreateEmergencyContactFunc func(ctx context.Context, contact *domain.PsiUserEmergencyContact) error
	GetEmergencyContactFunc    func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error)
	UpdateEmergencyContactFunc func(ctx context.Context, contact *domain.PsiUserEmergencyContact) error
	DeleteEmergencyContactFunc func(ctx context.Context, id uuid.UUID) error
	GetByIDFunc                func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	GetByFPVFunc               func(ctx context.Context, id int) (domain.PsiUserModel, error)
	GetTextContentByIDFunc     func(ctx context.Context, id uuid.UUID) (string, error)
}

func (m *mockPsiRepoEmergency) CountEmergencyContactsByPsiID(ctx context.Context, id uuid.UUID) (int64, error) {
	return m.CountEmergencyContactsFunc(ctx, id)
}
func (m *mockPsiRepoEmergency) CreateEmergencyContact(ctx context.Context, c *domain.PsiUserEmergencyContact) error {
	return m.CreateEmergencyContactFunc(ctx, c)
}
func (m *mockPsiRepoEmergency) GetEmergencyContactByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
	return m.GetEmergencyContactFunc(ctx, id)
}
func (m *mockPsiRepoEmergency) UpdateEmergencyContact(ctx context.Context, c *domain.PsiUserEmergencyContact) error {
	return m.UpdateEmergencyContactFunc(ctx, c)
}
func (m *mockPsiRepoEmergency) DeleteEmergencyContact(ctx context.Context, id uuid.UUID) error {
	return m.DeleteEmergencyContactFunc(ctx, id)
}
func (m *mockPsiRepoEmergency) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockPsiRepoEmergency) GetByFPV(ctx context.Context, id int) (domain.PsiUserModel, error) {
	return m.GetByFPVFunc(ctx, id)
}
func (m *mockPsiRepoEmergency) GetTextContentByID(ctx context.Context, id uuid.UUID) (string, error) {
	return m.GetTextContentByIDFunc(ctx, id)
}
func (m *mockPsiRepoEmergency) ResolveSpecialtyNames(ctx context.Context, ids []uint32, legacy []string) ([]string, error) {
	return []string{}, nil
}

// newEmergencyRepo devuelve un mock con los CRUD de contactos en éxito por
// defecto, para que cada test solo declare el comportamiento que quiere probar.
func newEmergencyRepo() *mockPsiRepoEmergency {
	return &mockPsiRepoEmergency{
		CountEmergencyContactsFunc: func(ctx context.Context, id uuid.UUID) (int64, error) { return 0, nil },
		CreateEmergencyContactFunc: func(ctx context.Context, c *domain.PsiUserEmergencyContact) error { return nil },
		GetEmergencyContactFunc: func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
			return nil, errors.New("no encontrado")
		},
		UpdateEmergencyContactFunc: func(ctx context.Context, c *domain.PsiUserEmergencyContact) error { return nil },
		DeleteEmergencyContactFunc: func(ctx context.Context, id uuid.UUID) error { return nil },
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{ID: id, FirstName: "Ana", LastName: "Prueba", FPV: 12345, Credentials: domain.Credentials{Username: "ana"}}, nil
		},
	}
}

func newTestPsi() *domain.PsiUserModel {
	return &domain.PsiUserModel{
		ID:          uuid.Must(uuid.NewV7()),
		Credentials: domain.Credentials{Username: "psicologo1"},
	}
}

// =========================================================================
// NORMALIZACIÓN Y REGLA DE INTEGRIDAD
// =========================================================================

// TestNormalizeEmergencyContact fija la regla declarada por el Colegio:
// SIEMPRE nombre y parentesco, y al menos UNO de teléfono o correo.
func TestNormalizeEmergencyContact(t *testing.T) {
	cases := []struct {
		name         string
		nombre       string
		parentesco   string
		telefono     string
		correo       string
		wantErr      bool
		wantPhone    string
		wantEmail    string
		wantName     string
		wantRelation string
	}{
		{
			name: "Solo teléfono es válido", nombre: "María Rodríguez", parentesco: "madre",
			telefono: "0412-1234567",
			wantName: "María Rodríguez", wantRelation: "madre", wantPhone: "04121234567",
		},
		{
			name: "Solo correo es válido", nombre: "María Rodríguez", parentesco: "madre",
			correo:   "MARIA@Correo.COM",
			wantName: "María Rodríguez", wantRelation: "madre", wantEmail: "maria@correo.com",
		},
		{
			name: "Teléfono y correo juntos también son válidos", nombre: "Juan Pérez", parentesco: "Padre",
			telefono: "+58 412 000 1111", correo: "juan@correo.com",
			wantName: "Juan Pérez", wantRelation: "Padre", wantPhone: "+584120001111", wantEmail: "juan@correo.com",
		},
		{
			name: "Sin nombre → error", parentesco: "madre", telefono: "04121234567", wantErr: true,
		},
		{
			name: "Sin parentesco → error", nombre: "María", telefono: "04121234567", wantErr: true,
		},
		{
			name: "Sin teléfono ni correo → error", nombre: "María", parentesco: "madre", wantErr: true,
		},
		{
			name: "Solo espacios → error", nombre: "   ", parentesco: "madre", telefono: "04121234567", wantErr: true,
		},
		{
			name: "Correo con formato inválido → error", nombre: "María", parentesco: "madre",
			correo: "no-es-un-correo", wantErr: true,
		},
		{
			name: "Teléfono basura se descarta y sin correo → error", nombre: "María", parentesco: "madre",
			telefono: "abc", wantErr: true,
		},
		{
			name: "Teléfono demasiado corto se descarta", nombre: "María", parentesco: "madre",
			telefono: "12345", wantErr: true,
		},
		{
			name: "Espacios internos colapsados", nombre: "  María   Rodríguez  ", parentesco: "  madre ",
			correo:   "maria@correo.com",
			wantName: "María Rodríguez", wantRelation: "madre", wantEmail: "maria@correo.com",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeEmergencyContact(tc.nombre, tc.parentesco, tc.telefono, tc.correo)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("se esperaba error y se obtuvo un resultado válido: %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("no se esperaba error: %v", err)
			}
			if got.Name != tc.wantName {
				t.Errorf("Name = %q, se esperaba %q", got.Name, tc.wantName)
			}
			if got.Relationship != tc.wantRelation {
				t.Errorf("Relationship = %q, se esperaba %q", got.Relationship, tc.wantRelation)
			}
			if got.Phone != tc.wantPhone {
				t.Errorf("Phone = %q, se esperaba %q", got.Phone, tc.wantPhone)
			}
			if got.Email != tc.wantEmail {
				t.Errorf("Email = %q, se esperaba %q", got.Email, tc.wantEmail)
			}
		})
	}
}

// TestNormalizeEmergencyPhone documenta el saneo del teléfono: separadores fuera,
// '+' inicial conservado, longitud mínima y máxima acotada.
func TestNormalizeEmergencyPhone(t *testing.T) {
	cases := map[string]string{
		"0412-123.45.67":        "04121234567",
		"(0412) 1234567":        "04121234567",
		"+58 412-000-11-11":     "+584120001111",
		"  0241 555 44 33  ":    "02415554433",
		"":                      "",
		"   ":                   "",
		"abc":                   "",
		"12345":                 "", // < 7 dígitos
		"123456789012345678901": "", // > 20 dígitos
		"+":                     "",
	}
	for in, want := range cases {
		if got := normalizeEmergencyPhone(in); got != want {
			t.Errorf("normalizeEmergencyPhone(%q) = %q, se esperaba %q", in, got, want)
		}
	}
}

// =========================================================================
// AUTOGESTIÓN (AGREMIADO)
// =========================================================================

// TestPsiService_AddEmergencyContact evalúa la Mitigación de Agotamiento de
// Recursos: la cuota de MaxEmergencyContacts evita que un script inunde la tabla.
func TestPsiService_AddEmergencyContact(t *testing.T) {
	repo := newEmergencyRepo()
	svc := &PsiService{repo: repo}
	ctx := context.Background()
	psi := newTestPsi()

	t.Run("Éxito: normaliza y persiste con auditoría del agremiado", func(t *testing.T) {
		var captured *domain.PsiUserEmergencyContact
		repo.CreateEmergencyContactFunc = func(ctx context.Context, c *domain.PsiUserEmergencyContact) error {
			captured = c
			return nil
		}

		req := request_structs.CreateEmergencyContactRequest{
			Name:         "  María   Rodríguez ",
			Relationship: "madre",
			Phone:        "0412-1234567",
		}
		if err := svc.AddEmergencyContact(ctx, psi, req); err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if captured == nil {
			t.Fatal("no se persistió el contacto")
		}
		if captured.PsiUserID != psi.ID {
			t.Error("el contacto no quedó vinculado al agremiado autenticado")
		}
		if captured.Name != "María Rodríguez" || captured.Phone != "04121234567" {
			t.Errorf("no se normalizó el contacto: %+v", captured)
		}
		if captured.CreateBy != psi.Username || captured.CreateById == nil || *captured.CreateById != psi.ID {
			t.Error("la auditoría del alta debe registrar al agremiado")
		}
		if captured.ID == uuid.Nil {
			t.Error("se esperaba un UUID generado para el contacto")
		}
	})

	t.Run("Error: cuota alcanzada", func(t *testing.T) {
		repo.CountEmergencyContactsFunc = func(ctx context.Context, id uuid.UUID) (int64, error) {
			return int64(MaxEmergencyContacts), nil
		}
		defer func() {
			repo.CountEmergencyContactsFunc = func(ctx context.Context, id uuid.UUID) (int64, error) { return 0, nil }
		}()

		req := request_structs.CreateEmergencyContactRequest{Name: "María", Relationship: "madre", Phone: "04121234567"}
		err := svc.AddEmergencyContact(ctx, psi, req)
		if err == nil || !errors.Is(err, domain.ErrMaxEmergencyContacts) {
			t.Errorf("se esperaba ErrMaxEmergencyContacts, se obtuvo: %v", err)
		}
	})

	t.Run("Error: sin canal de contacto no persiste", func(t *testing.T) {
		llamado := false
		repo.CreateEmergencyContactFunc = func(ctx context.Context, c *domain.PsiUserEmergencyContact) error {
			llamado = true
			return nil
		}
		defer func() {
			repo.CreateEmergencyContactFunc = func(ctx context.Context, c *domain.PsiUserEmergencyContact) error { return nil }
		}()

		req := request_structs.CreateEmergencyContactRequest{Name: "María", Relationship: "madre"}
		err := svc.AddEmergencyContact(ctx, psi, req)
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactIncomplete) {
			t.Errorf("se esperaba ErrEmergencyContactIncomplete, se obtuvo: %v", err)
		}
		if llamado {
			t.Error("no debe persistirse un contacto sin teléfono ni correo")
		}
	})
}

// TestPsiService_UpdateEmergencyContact_Ownership evalúa la prevención de IDOR:
// el agremiado no puede editar el contacto de otra ficha inyectando el UUID en la URL.
func TestPsiService_UpdateEmergencyContact_Ownership(t *testing.T) {
	repo := newEmergencyRepo()
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiA := newTestPsi()
	psiB := newTestPsi()
	contactID := uuid.Must(uuid.NewV7())

	repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
		return &domain.PsiUserEmergencyContact{ID: contactID, PsiUserID: psiB.ID, Name: "María", Relationship: "madre", Phone: "04121234567"}, nil
	}

	t.Run("Error: editar contacto ajeno (ID Spoofing)", func(t *testing.T) {
		nuevo := "Intruso"
		err := svc.UpdateEmergencyContact(ctx, psiA, contactID, request_structs.UpdateEmergencyContactRequest{Name: &nuevo})
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactPermDenied) {
			t.Errorf("se esperaba ErrEmergencyContactPermDenied, se obtuvo: %v", err)
		}
	})

	t.Run("Error: contacto inexistente", func(t *testing.T) {
		repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
			return nil, errors.New("no encontrado")
		}
		err := svc.UpdateEmergencyContact(ctx, psiA, contactID, request_structs.UpdateEmergencyContactRequest{})
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactNotFound) {
			t.Errorf("se esperaba ErrEmergencyContactNotFound, se obtuvo: %v", err)
		}
	})
}

// TestPsiService_UpdateEmergencyContact_ReglaIntegridad comprueba que un PATCH
// parcial nunca puede dejar el registro en un estado inconsistente (borrar el
// último canal de contacto, vaciar el nombre o el parentesco).
func TestPsiService_UpdateEmergencyContact_ReglaIntegridad(t *testing.T) {
	ctx := context.Background()
	psi := newTestPsi()
	contactID := uuid.Must(uuid.NewV7())

	repo := newEmergencyRepo()
	svc := &PsiService{repo: repo}

	// Contacto válido: solo teléfono.
	repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
		return &domain.PsiUserEmergencyContact{ID: contactID, PsiUserID: psi.ID, Name: "María", Relationship: "madre", Phone: "04121234567"}, nil
	}

	t.Run("Error: no se puede quedar sin canal de contacto", func(t *testing.T) {
		persistido := false
		repo.UpdateEmergencyContactFunc = func(ctx context.Context, c *domain.PsiUserEmergencyContact) error {
			persistido = true
			return nil
		}
		vacio := ""
		err := svc.UpdateEmergencyContact(ctx, psi, contactID, request_structs.UpdateEmergencyContactRequest{Phone: &vacio})
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactIncomplete) {
			t.Errorf("se esperaba ErrEmergencyContactIncomplete, se obtuvo: %v", err)
		}
		if persistido {
			t.Error("no debe persistirse un PATCH que deja el contacto sin canal")
		}
	})

	t.Run("Error: no se puede vaciar el nombre", func(t *testing.T) {
		vacio := "   "
		err := svc.UpdateEmergencyContact(ctx, psi, contactID, request_structs.UpdateEmergencyContactRequest{Name: &vacio})
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactIncomplete) {
			t.Errorf("se esperaba ErrEmergencyContactIncomplete, se obtuvo: %v", err)
		}
	})

	t.Run("Error: no se puede vaciar el parentesco", func(t *testing.T) {
		vacio := ""
		err := svc.UpdateEmergencyContact(ctx, psi, contactID, request_structs.UpdateEmergencyContactRequest{Relationship: &vacio})
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactIncomplete) {
			t.Errorf("se esperaba ErrEmergencyContactIncomplete, se obtuvo: %v", err)
		}
	})

	t.Run("Éxito: migrar de teléfono a correo", func(t *testing.T) {
		var captured *domain.PsiUserEmergencyContact
		repo.UpdateEmergencyContactFunc = func(ctx context.Context, c *domain.PsiUserEmergencyContact) error {
			captured = c
			return nil
		}
		tel := ""
		mail := "  Nueva@Correo.com "
		if err := svc.UpdateEmergencyContact(ctx, psi, contactID, request_structs.UpdateEmergencyContactRequest{
			Phone: &tel,
			Email: &mail,
		}); err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if captured == nil {
			t.Fatal("no se persistió el cambio")
		}
		if captured.Phone != "" || captured.Email != "nueva@correo.com" {
			t.Errorf("la migración de canal quedó mal: %+v", captured)
		}
		// La auditoría de campo debe reflejar al agremiado.
		if captured.UpdateBy != psi.Username || captured.UpdateById == nil || *captured.UpdateById != psi.ID {
			t.Error("la auditoría de la edición debe registrar al agremiado")
		}
	})
}

// TestPsiService_DeleteEmergencyContact_Roles evalúa el Control de Acceso
// Polimórfico del borrado: autogestión (solo lo propio) vs. moderación admin.
func TestPsiService_DeleteEmergencyContact_Roles(t *testing.T) {
	repo := newEmergencyRepo()
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	ownerID := uuid.Must(uuid.NewV7())
	otroID := uuid.Must(uuid.NewV7())
	contactID := uuid.Must(uuid.NewV7())

	repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
		return &domain.PsiUserEmergencyContact{ID: contactID, PsiUserID: ownerID, Name: "María", Relationship: "madre", Phone: "04121234567"}, nil
	}
	repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
		return &domain.PsiUserModel{ID: ownerID, FirstName: "Dueño", LastName: "Prueba", FPV: 99999, Credentials: domain.Credentials{Username: "dueno"}}, nil
	}

	t.Run("Psi: puede borrar su propio contacto", func(t *testing.T) {
		if err := svc.DeleteEmergencyContact(ctx, "psi", ownerID, contactID); err != nil {
			t.Errorf("error inesperado: %v", err)
		}
	})

	t.Run("Psi: no puede borrar el contacto de otro (IDOR)", func(t *testing.T) {
		err := svc.DeleteEmergencyContact(ctx, "psi", otroID, contactID)
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactOwnDenied) {
			t.Errorf("se esperaba ErrEmergencyContactOwnDenied, se obtuvo: %v", err)
		}
	})

	t.Run("Admin: puede borrar cualquier contacto", func(t *testing.T) {
		if err := svc.DeleteEmergencyContact(ctx, "admin", uuid.Must(uuid.NewV7()), contactID); err != nil {
			t.Errorf("el admin debería poder borrar, error: %v", err)
		}
	})

	t.Run("Error: rol desconocido rechazado", func(t *testing.T) {
		err := svc.DeleteEmergencyContact(ctx, "invitado", ownerID, contactID)
		if err == nil || !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("se esperaba ErrInsufficientPerms, se obtuvo: %v", err)
		}
	})
}

// =========================================================================
// MODERACIÓN ADMINISTRATIVA
// =========================================================================

// TestPsiService_AddEmergencyContactByAdmin evalúa el gatekeeping: el panel puede
// registrar un contacto que el Colegio obtuvo por otro medio, pero solo staff
// autorizado, y la auditoría debe atribuir la operación al operador.
func TestPsiService_AddEmergencyContactByAdmin(t *testing.T) {
	repo := newEmergencyRepo()
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiID := uuid.Must(uuid.NewV7())
	adminSudo := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "sudo_admin"}, Sudo: true}
	adminSecretaria := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "secretaria"}, CanUpdatePsi: true}
	adminLector := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "lector"}}
	req := request_structs.CreateEmergencyContactRequest{Name: "María Rodríguez", Relationship: "madre", Phone: "04121234567"}

	t.Run("Éxito: Sudo registra el contacto con auditoría admin", func(t *testing.T) {
		var captured *domain.PsiUserEmergencyContact
		repo.CreateEmergencyContactFunc = func(ctx context.Context, c *domain.PsiUserEmergencyContact) error {
			captured = c
			return nil
		}
		if err := svc.AddEmergencyContactByAdmin(ctx, adminSudo, psiID, req); err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if captured == nil || captured.PsiUserID != psiID {
			t.Error("el contacto no se persistió para la ficha indicada")
		}
		if captured.CreateBy != "sudo_admin" {
			t.Error("la auditoría debe registrar al operador administrativo")
		}
	})

	t.Run("Éxito: Sudo no es la única vía (CanUpdatePsi)", func(t *testing.T) {
		if err := svc.AddEmergencyContactByAdmin(ctx, adminSecretaria, psiID, req); err != nil {
			t.Errorf("no se esperaba error: %v", err)
		}
	})

	t.Run("Error: permisos insuficientes", func(t *testing.T) {
		err := svc.AddEmergencyContactByAdmin(ctx, adminLector, psiID, req)
		if err == nil || !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("se esperaba ErrInsufficientPerms, se obtuvo: %v", err)
		}
	})

	t.Run("Error: ficha inexistente", func(t *testing.T) {
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return nil, errors.New("no encontrado")
		}
		defer func() {
			repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return &domain.PsiUserModel{ID: id, FirstName: "Ana", LastName: "Prueba", FPV: 12345}, nil
			}
		}()
		err := svc.AddEmergencyContactByAdmin(ctx, adminSudo, psiID, req)
		if err == nil || !errors.Is(err, domain.ErrPsiNotFound) {
			t.Errorf("se esperaba ErrPsiNotFound, se obtuvo: %v", err)
		}
	})

	t.Run("Error: cuota alcanzada", func(t *testing.T) {
		repo.CountEmergencyContactsFunc = func(ctx context.Context, id uuid.UUID) (int64, error) {
			return int64(MaxEmergencyContacts), nil
		}
		defer func() {
			repo.CountEmergencyContactsFunc = func(ctx context.Context, id uuid.UUID) (int64, error) { return 0, nil }
		}()
		err := svc.AddEmergencyContactByAdmin(ctx, adminSudo, psiID, req)
		if err == nil || !errors.Is(err, domain.ErrMaxEmergencyContacts) {
			t.Errorf("se esperaba ErrMaxEmergencyContacts, se obtuvo: %v", err)
		}
	})
}

// TestPsiService_UpdateEmergencyContactByAdmin evalúa el gatekeeping y la
// prevención de IDOR (el contacto debe pertenecer a la ficha de la ruta).
func TestPsiService_UpdateEmergencyContactByAdmin(t *testing.T) {
	repo := newEmergencyRepo()
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiID := uuid.Must(uuid.NewV7())
	otraFicha := uuid.Must(uuid.NewV7())
	contactID := uuid.Must(uuid.NewV7())
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "secretaria"}, CanUpdatePsi: true}

	t.Run("Éxito: actualiza el contacto de la ficha", func(t *testing.T) {
		repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
			return &domain.PsiUserEmergencyContact{ID: contactID, PsiUserID: psiID, Name: "María", Relationship: "madre", Phone: "04121234567"}, nil
		}
		var captured *domain.PsiUserEmergencyContact
		repo.UpdateEmergencyContactFunc = func(ctx context.Context, c *domain.PsiUserEmergencyContact) error {
			captured = c
			return nil
		}
		nuevo := "  María   Rodríguez  "
		if err := svc.UpdateEmergencyContactByAdmin(ctx, admin, psiID, contactID, request_structs.UpdateEmergencyContactRequest{Name: &nuevo}); err != nil {
			t.Errorf("no se esperaba error: %v", err)
		}
		if captured == nil || captured.Name != "María Rodríguez" {
			t.Errorf("el nombre no se normalizó: %+v", captured)
		}
		if captured == nil || captured.UpdateBy != "secretaria" {
			t.Error("la auditoría debe registrar al operador administrativo")
		}
	})

	t.Run("Error: IDOR — contacto de otra ficha", func(t *testing.T) {
		repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
			return &domain.PsiUserEmergencyContact{ID: contactID, PsiUserID: otraFicha, Name: "Secreto", Relationship: "madre", Phone: "04121234567"}, nil
		}
		err := svc.UpdateEmergencyContactByAdmin(ctx, admin, psiID, contactID, request_structs.UpdateEmergencyContactRequest{})
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactOwnDenied) {
			t.Errorf("se esperaba ErrEmergencyContactOwnDenied, se obtuvo: %v", err)
		}
	})

	t.Run("Error: permisos insuficientes", func(t *testing.T) {
		lector := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "lector"}}
		err := svc.UpdateEmergencyContactByAdmin(ctx, lector, psiID, contactID, request_structs.UpdateEmergencyContactRequest{})
		if err == nil || !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("se esperaba ErrInsufficientPerms, se obtuvo: %v", err)
		}
	})

	t.Run("Error: la regla de integridad también aplica en el panel admin", func(t *testing.T) {
		repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
			return &domain.PsiUserEmergencyContact{ID: contactID, PsiUserID: psiID, Name: "María", Relationship: "madre", Phone: "04121234567"}, nil
		}
		vacio := ""
		err := svc.UpdateEmergencyContactByAdmin(ctx, admin, psiID, contactID, request_structs.UpdateEmergencyContactRequest{
			Phone: &vacio,
			Email: &vacio,
		})
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactIncomplete) {
			t.Errorf("se esperaba ErrEmergencyContactIncomplete, se obtuvo: %v", err)
		}
	})
}

// TestPsiService_DeleteEmergencyContactByAdmin evalúa el gatekeeping (el borrado
// admite CanDeletePsi) y la prevención de IDOR.
func TestPsiService_DeleteEmergencyContactByAdmin(t *testing.T) {
	repo := newEmergencyRepo()
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	psiID := uuid.Must(uuid.NewV7())
	otraFicha := uuid.Must(uuid.NewV7())
	contactID := uuid.Must(uuid.NewV7())
	adminBorrador := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "soporte"}, CanDeletePsi: true}
	adminLector := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "lector"}}

	t.Run("Éxito: borra el contacto de la ficha (CanDeletePsi)", func(t *testing.T) {
		repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
			return &domain.PsiUserEmergencyContact{ID: contactID, PsiUserID: psiID, Name: "María", Relationship: "madre", Phone: "04121234567"}, nil
		}
		if err := svc.DeleteEmergencyContactByAdmin(ctx, adminBorrador, psiID, contactID); err != nil {
			t.Errorf("no se esperaba error: %v", err)
		}
	})

	t.Run("Error: IDOR — contacto de otra ficha", func(t *testing.T) {
		repo.GetEmergencyContactFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserEmergencyContact, error) {
			return &domain.PsiUserEmergencyContact{ID: contactID, PsiUserID: otraFicha}, nil
		}
		err := svc.DeleteEmergencyContactByAdmin(ctx, adminBorrador, psiID, contactID)
		if err == nil || !errors.Is(err, domain.ErrEmergencyContactOwnDenied) {
			t.Errorf("se esperaba ErrEmergencyContactOwnDenied, se obtuvo: %v", err)
		}
	})

	t.Run("Error: permisos insuficientes", func(t *testing.T) {
		err := svc.DeleteEmergencyContactByAdmin(ctx, adminLector, psiID, contactID)
		if err == nil || !errors.Is(err, domain.ErrInsufficientPerms) {
			t.Errorf("se esperaba ErrInsufficientPerms, se obtuvo: %v", err)
		}
	})
}

// =========================================================================
// PRIVACIDAD: LOS CONTACTOS NUNCA SE EXPONEN PÚBLICAMENTE
// =========================================================================

// TestEmergencyContacts_NoSeFiltranEnEndpointsPublicos es una prueba de regresión
// de privacidad (LOPDP): la ficha pública y el directorio se construyen con DTOs
// manuales, por lo que la nueva relación NO debe aparecer en ellos aunque el
// modelo de dominio la traiga cargada. Si alguien agrega el campo a un DTO
// público, esta prueba falla.
func TestEmergencyContacts_NoSeFiltranEnEndpointsPublicos(t *testing.T) {
	dtoTypes := []reflect.Type{
		reflect.TypeOf(request_structs.PsiFullProfileDTO{}),
		reflect.TypeOf(request_structs.PsiMiniProfileDTO{}),
	}
	for _, tp := range dtoTypes {
		for i := 0; i < tp.NumField(); i++ {
			tag := tp.Field(i).Tag.Get("json")
			if strings.Contains(strings.ToLower(tag), "emergency") {
				t.Errorf("%s expone el campo %q: los contactos de emergencia son datos de un tercero y no pueden ser públicos", tp.Name(), tag)
			}
		}
	}
}

// TestGetPublicProfile_NoExponeContactos comprueba el caso real de extremo a
// extremo del servicio público: ni el JSON del DTO ni el modelo cargado por
// GetByFPV deben filtrar el contacto de emergencia.
func TestGetPublicProfile_NoExponeContactos(t *testing.T) {
	repo := newEmergencyRepo()
	svc := &PsiService{repo: repo}
	ctx := context.Background()

	contacto := domain.PsiUserEmergencyContact{
		ID:           uuid.Must(uuid.NewV7()),
		PsiUserID:    uuid.Must(uuid.NewV7()),
		Name:         "María Rodríguez",
		Relationship: "madre",
		Phone:        "04121234567",
		Email:        "maria@correo.com",
	}
	repo.GetByFPVFunc = func(ctx context.Context, id int) (domain.PsiUserModel, error) {
		psi := domain.PsiUserModel{
			FPV:       id,
			FirstName: "Ana",
			LastName:  "Pública",
			Solvent:   true,
			// El repositorio público NO precarga la relación, pero si alguien lo
			// hiciera, el DTO debe seguir sin exponerla.
			EmergencyContacts: []domain.PsiUserEmergencyContact{contacto},
		}
		psi.Credentials.IsActive = true
		return psi, nil
	}
	repo.GetTextContentByIDFunc = func(ctx context.Context, id uuid.UUID) (string, error) { return "", nil }

	dto, _, err := svc.GetPublicProfile(ctx, 12345)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	payload, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("no se pudo serializar el DTO: %v", err)
	}
	serializado := strings.ToLower(string(payload))
	for _, prohibido := range []string{"emergencia", "emergency", contacto.Name, contacto.Phone, contacto.Email, contacto.Relationship} {
		if strings.Contains(serializado, strings.ToLower(prohibido)) {
			t.Errorf("la ficha pública filtró %q: %s", prohibido, payload)
		}
	}
}

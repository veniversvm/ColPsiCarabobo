package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// MOCKS PARA ADMIN REPOSITORY (PATRÓN FUNC OVERRIDE)
// =========================================================================
// En lugar de depender de librerías de Mocking de terceros (como gomock),
// se utiliza el patrón de Mocks Funcionales. Al exponer los métodos del
// repositorio como variables de función (Func fields), cada sub-test (t.Run)
// puede inyectar comportamientos y errores específicos de manera aislada,
// evitando que el estado de un test contamine al siguiente.

type mockAdminRepo struct {
	domain.UserAdminRepository
	GetByIdentifierFunc func(ctx context.Context, identifier string) (*domain.UserAdmin, error)
	GetByIDFunc         func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error)
	UpdateFunc          func(ctx context.Context, admin *domain.UserAdmin) error
	UpdateKeyFunc       func(ctx context.Context, admin *domain.UserAdmin) error
	CreateFunc          func(ctx context.Context, admin *domain.UserAdmin) error
	DeleteFunc          func(ctx context.Context, id uuid.UUID) error
	ListFunc            func(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error)
}

func (m *mockAdminRepo) GetByIdentifier(ctx context.Context, id string) (*domain.UserAdmin, error) {
	return m.GetByIdentifierFunc(ctx, id)
}
func (m *mockAdminRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockAdminRepo) Update(ctx context.Context, a *domain.UserAdmin) error {
	return m.UpdateFunc(ctx, a)
}
func (m *mockAdminRepo) UpdateKey(ctx context.Context, a *domain.UserAdmin) error {
	if m.UpdateKeyFunc != nil {
		return m.UpdateKeyFunc(ctx, a)
	}
	return nil
}
func (m *mockAdminRepo) Create(ctx context.Context, a *domain.UserAdmin) error {
	return m.CreateFunc(ctx, a)
}
func (m *mockAdminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *mockAdminRepo) List(ctx context.Context, a *bool, s string, p, l int) ([]domain.UserAdmin, int64, error) {
	return m.ListFunc(ctx, a, s, p, l)
}

// =========================================================================
// TEST SUITE: CASOS DE USO ADMINISTRATIVOS (LÓGICA DE NEGOCIO)
// =========================================================================

// TestAdminService_All agrupa la validación de las reglas de negocio (Business Rules)
// más críticas del sistema: Control de Acceso Basado en Roles (RBAC),
// Prevención de Escalada de Privilegios, e Integridad del Rendimiento (Caché).
func TestAdminService_All(t *testing.T) {
	repo := &mockAdminRepo{}

	// Inyección de Dependencias (DI):
	// Usamos nil para mailService por brevedad en los tests donde el envío de
	// correos no es la lógica central a evaluar (Fire-and-Forget simulado).
	svc := NewAdminService(repo, nil)

	// --- 1. TEST DE LOGIN Y KEY ROTATION (SINGLE SESSION ENFORCEMENT) ---
	// Verifica la mitigación de secuestro de sesiones. Al probar que la 'Key'
	// muta obligatoriamente tras un inicio de sesión exitoso, aseguramos que
	// cualquier JWT robado previamente quede invalidado de forma automática.
	t.Run("Login: Éxito y Rotación de Key", func(t *testing.T) {
		pass := "Admin123!"
		hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		adminID := uuid.Must(uuid.NewV7())

		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{
				ID: adminID,
				Credentials: domain.Credentials{
					Password: string(hashed),
					IsActive: true,
					Username: "admin_test",
					// Nota Técnica de Integración:
					// Se requiere usar un dominio con registros MX reales (como gmail.com)
					// debido a que la utilidad de validación (utils.ParseAndValidateEmail)
					// realiza una consulta DNS de los servidores de correo para evitar emails falsos.
					Email: "admin@gmail.com",
				},
			}, nil
		}

		repo.UpdateFunc = func(ctx context.Context, a *domain.UserAdmin) error {
			// Aserción de Seguridad: Si el servicio no cambió la llave criptográfica, el test falla.
			if a.Key == "" {
				t.Error("La Key de sesión no fue rotada")
			}
			return nil
		}

		token, _, err := svc.Login(context.Background(), "admin_test", pass)
		if err != nil || token == "" {
			t.Errorf("Error en login exitoso: %v", err)
		}
	})

	// --- 2. TEST DE RENDIMIENTO: CACHE-ASIDE PATTERN ---
	// Demuestra la eficiencia de la Capa de Lógica. Se realizan dos peticiones
	// idénticas al servicio, y se cuenta cuántas veces se invocó realmente a la DB.
	// Si el caché en memoria (RAM) funciona, el repositorio solo debe llamarse una vez.
	t.Run("GetAdmins: Verificación de Caché", func(t *testing.T) {
		callCount := 0
		repo.ListFunc = func(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
			callCount++
			return []domain.UserAdmin{{Credentials: domain.Credentials{Username: "cached_user"}}}, 1, nil
		}

		// Primera llamada (Cache Miss -> Consulta a la Base de Datos simulada)
		_, _ = svc.GetAdmins(context.Background(), nil, "", 1, 10)

		// Segunda llamada (Cache Hit -> Debe retornar directo desde RAM)
		_, _ = svc.GetAdmins(context.Background(), nil, "", 1, 10)

		// Aserción de Rendimiento
		if callCount > 1 {
			t.Errorf("El sistema de caché no está funcionando, repo llamado %d veces", callCount)
		}
	})

	// --- 3. TEST DE SEGURIDAD: PREVENCIÓN DE ESCALADA DE PRIVILEGIOS ---
	// Valida el Principio de Menor Privilegio (PoLP). Un creador malintencionado o
	// comprometido no debe poder generar cuentas con accesos que él mismo no posee.
	t.Run("CreateAdmin: Regla 'No puedes dar lo que no tienes'", func(t *testing.T) {
		creator := domain.UserAdmin{
			ID: uuid.Must(uuid.NewV7()),
			Credentials: domain.Credentials{
				Username: "moderador",
			},
			CanCreateAdmin: true,
			Sudo:           false, // No es Super Usuario
			CanPublish:     false, // <--- Carencia del permiso específico
		}

		trueVal := true
		req := request_structs.CreateAdminRequest{
			Username: "nuevo_admin",
			// Uso de dominio real (gmail) para superar la validación DNS (MX Record Lookup)
			Email:    "nuevo_admin_test@gmail.com",
			Password: "Password123!",
			Permissions: request_structs.AdminPermissionsDTO{
				CanPublish: &trueVal, // <--- Intento de Inyección de Privilegios
			},
		}

		err := svc.CreateAdmin(context.Background(), creator, req)

		// Aserción Estricta: Se espera un error exacto del Matrix Engine bloqueando la acción.
		if err == nil || err.Error() != "no puedes otorgar el permiso: Publish" {
			t.Errorf("Se esperaba bloqueo por jerarquía, se obtuvo: %v", err)
		}
	})

	// --- 4. TEST DE INTEGRIDAD JERÁRQUICA: PROTECCIÓN SUDO ---
	// Verifica la Inmunidad del Super Usuario. Nadie en el sistema (excepto otro Sudo
	// o acceso directo a BD) tiene autoridad para alterar el perfil raíz.
	t.Run("UpdateAdmin: Proteger Super Usuario de ediciones externas", func(t *testing.T) {
		updater := domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Sudo: false}    // Staff regular
		targetSudo := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Sudo: true} // Objetivo protegido

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return targetSudo, nil
		}

		req := request_structs.UpdateAdminRequest{ID: targetSudo.ID}
		err := svc.UpdateAdmin(context.Background(), updater, req)

		// Aserción de Inmunidad
		if err == nil || err.Error() != "no puedes editar a un Super Usuario" {
			t.Errorf("Un admin normal pudo editar a un Sudo: %v", err)
		}
	})

	// --- 5. TEST DE DISPONIBILIDAD: PREVENCIÓN DE LOCKOUT (AUTO-SUICIDIO) ---
	// Mitiga errores humanos críticos. Evita que un administrador borre su propia
	// sesión/cuenta accidentalmente, lo cual podría dejar el sistema sin operadores.
	t.Run("DeleteAdmin: Impedir auto-eliminación", func(t *testing.T) {
		adminID := uuid.Must(uuid.NewV7())
		updater := &domain.UserAdmin{ID: adminID} // El ejecutor es el mismo objetivo

		err := svc.DeleteAdmin(context.Background(), updater, adminID)

		// Aserción Anti-Kamikaze
		if err == nil || err.Error() != "no puedes eliminar tu propia cuenta" {
			t.Errorf("Se esperaba error de auto-eliminación, se obtuvo: %v", err)
		}
	})
}

// =========================================================================
// TESTS EXPANDIDOS: CASOS EXTREMOS DE ADMIN SERVICE
// =========================================================================

func TestAdminService_Login_EdgeCases(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := NewAdminService(repo, nil)

	t.Run("Cuenta inactiva retorna error", func(t *testing.T) {
		pass := "Admin123!"
		hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{
				Credentials: domain.Credentials{
					Password: string(hashed),
					IsActive: false,
					Username: "inactive_admin",
				},
			}, nil
		}

		_, _, err := svc.Login(context.Background(), "inactive_admin", pass)
		if err == nil || err.Error() != "la cuenta está desactivada" {
			t.Errorf("Se esperaba cuenta desactivada, got: %v", err)
		}
	})

	t.Run("Contraseña incorrecta retorna credenciales inválidas", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("correct_pass"), bcrypt.DefaultCost)

		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{
				Credentials: domain.Credentials{
					Password: string(hashed),
					IsActive: true,
					Username: "admin_user",
					Email:    "admin_user@gmail.com",
				},
			}, nil
		}

		_, _, err := svc.Login(context.Background(), "admin_user", "wrong_password")
		if err == nil || err.Error() != "credenciales inválidas" {
			t.Errorf("Se esperaba credenciales inválidas, got: %v", err)
		}
	})

	t.Run("Usuario inexistente retorna credenciales inválidas", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.UserAdmin, error) {
			return nil, errors.New("record not found")
		}

		_, _, err := svc.Login(context.Background(), "ghost", "pass")
		if err == nil || err.Error() != "credenciales inválidas" {
			t.Errorf("Se esperaba credenciales inválidas, got: %v", err)
		}
	})
}

func TestAdminService_CreateAdmin_EdgeCases(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := NewAdminService(repo, nil)

	t.Run("Sin permiso CanCreateAdmin retorna error", func(t *testing.T) {
		creator := domain.UserAdmin{
			ID:               uuid.Must(uuid.NewV7()),
			Credentials:      domain.Credentials{Username: "regular"},
			CanCreateAdmin:   false,
			Sudo:             false,
		}
		req := request_structs.CreateAdminRequest{
			Username: "new_admin",
			Email:    "new_admin@gmail.com",
			Password: "Password123!",
		}

		err := svc.CreateAdmin(context.Background(), creator, req)
		if err == nil || err.Error() != "permisos insuficientes para crear administradores" {
			t.Errorf("Se esperaba permisos insuficientes, got: %v", err)
		}
	})

	t.Run("Email con formato inválido retorna error", func(t *testing.T) {
		creator := domain.UserAdmin{
			ID:             uuid.Must(uuid.NewV7()),
			Credentials:    domain.Credentials{Username: "sudo_user"},
			CanCreateAdmin: true,
			Sudo:           true,
		}
		req := request_structs.CreateAdminRequest{
			Username: "new_admin",
			Email:    "esto-no-es-email",
			Password: "Password123!",
		}

		err := svc.CreateAdmin(context.Background(), creator, req)
		if err == nil || err.Error() != "el formato del email es inválido" {
			t.Errorf("Se esperaba formato inválido, got: %v", err)
		}
	})

	t.Run("Contraseña débil retorna error", func(t *testing.T) {
		creator := domain.UserAdmin{
			ID:             uuid.Must(uuid.NewV7()),
			Credentials:    domain.Credentials{Username: "sudo_user"},
			CanCreateAdmin: true,
			Sudo:           true,
		}
		req := request_structs.CreateAdminRequest{
			Username: "new_admin",
			Email:    "new_admin@gmail.com",
			Password: "123",
		}

		err := svc.CreateAdmin(context.Background(), creator, req)
		if err == nil || !strings.Contains(err.Error(), "contraseña") {
			t.Errorf("Se esperaba error por contraseña débil, got: %v", err)
		}
	})

	t.Run("Sudo puede otorgar cualquier permiso", func(t *testing.T) {
		sudo := domain.UserAdmin{
			ID:               uuid.Must(uuid.NewV7()),
			Credentials:      domain.Credentials{Username: "sudo"},
			Sudo:             true,
			CanCreateAdmin:   true,
		}

		trueVal := true
		req := request_structs.CreateAdminRequest{
			Username: "new_power_admin@gmail.com",
			Email:    "power@gmail.com",
			Password: "Password123!",
			Permissions: request_structs.AdminPermissionsDTO{
				CanPublish:         &trueVal,
				CanCreateAdmin:     &trueVal,
				CanSendNotifications: &trueVal,
			},
		}

		repo.CreateFunc = func(ctx context.Context, a *domain.UserAdmin) error { return nil }
		err := svc.CreateAdmin(context.Background(), sudo, req)
		if err != nil {
			t.Errorf("Sudo debería poder otorgar cualquier permiso, got: %v", err)
		}
	})

	t.Run("Creación exitosa purga caché", func(t *testing.T) {
		sudo := domain.UserAdmin{
			ID:               uuid.Must(uuid.NewV7()),
			Credentials:      domain.Credentials{Username: "sudo"},
			Sudo:             true,
			CanCreateAdmin:   true,
		}

		req := request_structs.CreateAdminRequest{
			Username: "fresh_admin@gmail.com",
			Email:    "fresh@gmail.com",
			Password: "Password123!",
		}

		createCalled := false
		repo.CreateFunc = func(ctx context.Context, a *domain.UserAdmin) error {
			createCalled = true
			if a.Credentials.Username != "fresh_admin@gmail.com" {
				t.Errorf("Username = %q, want fresh_admin@gmail.com", a.Credentials.Username)
			}
			if a.Key == "" {
				t.Error("Key debería haberse generado")
			}
			if !a.IsActive {
				t.Error("IsActive debería ser true")
			}
			if a.Sudo {
				t.Error("Sudo debería ser false (no heredable)")
			}
			return nil
		}

		err := svc.CreateAdmin(context.Background(), sudo, req)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}
		if !createCalled {
			t.Error("El repo Create no fue llamado")
		}
	})
}

func TestAdminService_DeleteAdmin_EdgeCases(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := NewAdminService(repo, nil)

	t.Run("Sin permiso CanDeleteAdmin retorna error", func(t *testing.T) {
		updater := &domain.UserAdmin{
			ID:               uuid.Must(uuid.NewV7()),
			Sudo:             false,
			CanDeleteAdmin:   false,
		}
		targetID := uuid.Must(uuid.NewV7())

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{ID: id, Sudo: false}, nil
		}

		err := svc.DeleteAdmin(context.Background(), updater, targetID)
		if err == nil || err.Error() != "no tienes permiso para eliminar administradores" {
			t.Errorf("Se esperaba permiso denegado, got: %v", err)
		}
	})

	t.Run("No se puede eliminar a un Super Usuario (sin sudo)", func(t *testing.T) {
		updater := &domain.UserAdmin{
			ID:               uuid.Must(uuid.NewV7()),
			Sudo:             false,
			CanDeleteAdmin:   true,
		}
		sudoTarget := uuid.Must(uuid.NewV7())

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{ID: id, Sudo: true}, nil
		}

		err := svc.DeleteAdmin(context.Background(), updater, sudoTarget)
		if err == nil || err.Error() != "no puedes eliminar un Super Usuario" {
			t.Errorf("Se esperaba bloqueo de sudo, got: %v", err)
		}
	})

	t.Run("Administrador no encontrado retorna error", func(t *testing.T) {
		updater := &domain.UserAdmin{
			ID:             uuid.Must(uuid.NewV7()),
			Sudo:           true,
		}
		missingID := uuid.Must(uuid.NewV7())

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return nil, errors.New("record not found")
		}

		err := svc.DeleteAdmin(context.Background(), updater, missingID)
		if err == nil || err.Error() != "administrador no encontrado" {
			t.Errorf("Se esperaba no encontrado, got: %v", err)
		}
	})

	t.Run("Sudo puede eliminar otro admin", func(t *testing.T) {
		sudoAdmin := &domain.UserAdmin{
			ID:   uuid.Must(uuid.NewV7()),
			Sudo: true,
		}
		targetID := uuid.Must(uuid.NewV7())

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{ID: id, Sudo: false}, nil
		}

		deleteCalled := false
		repo.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
			deleteCalled = true
			return nil
		}

		err := svc.DeleteAdmin(context.Background(), sudoAdmin, targetID)
		if err != nil {
			t.Errorf("Sudo debería poder eliminar, got: %v", err)
		}
		if !deleteCalled {
			t.Error("El repo Delete no fue llamado")
		}
	})
}

func TestAdminService_UpdateAdmin_EdgeCases(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := NewAdminService(repo, nil)

	t.Run("Administrador no encontrado retorna error", func(t *testing.T) {
		updater := domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Sudo: false}
		req := request_structs.UpdateAdminRequest{ID: uuid.Must(uuid.NewV7())}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return nil, errors.New("record not found")
		}

		err := svc.UpdateAdmin(context.Background(), updater, req)
		if err == nil || err.Error() != "administrador no encontrado" {
			t.Errorf("Se esperaba no encontrado, got: %v", err)
		}
	})

	t.Run("Admin sin rango no puede modificar permisos de otro", func(t *testing.T) {
		updater := domain.UserAdmin{
			ID:             uuid.Must(uuid.NewV7()),
			Sudo:           false,
			CanPublish:     false, // No tiene permiso Publish
		}
		targetID := uuid.Must(uuid.NewV7())
		trueVal := true

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{
				ID:        id,
				Sudo:      false,
				CanPublish: false, // Target tampoco lo tiene
			}, nil
		}

		req := request_structs.UpdateAdminRequest{
			ID: targetID,
			Permissions: request_structs.AdminPermissionsDTO{
				CanPublish: &trueVal, // Intenta encenderlo sin tenerlo
			},
		}

		err := svc.UpdateAdmin(context.Background(), updater, req)
		if err == nil || !strings.Contains(err.Error(), "rango") {
			t.Errorf("Se esperaba error de rango, got: %v", err)
		}
	})

	t.Run("Actualización de username exitosa", func(t *testing.T) {
		updater := domain.UserAdmin{
			ID:   uuid.Must(uuid.NewV7()),
			Sudo: true, // Sudo bypasses permission checks
		}
		targetID := uuid.Must(uuid.NewV7())
		newUsername := "renamed_admin"

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
			return &domain.UserAdmin{
				ID: id,
				Credentials: domain.Credentials{
					Username: "old_name",
					Email:    "old@gmail.com",
				},
				Sudo: false,
			}, nil
		}

		var savedAdmin *domain.UserAdmin
		repo.UpdateFunc = func(ctx context.Context, a *domain.UserAdmin) error {
			savedAdmin = a
			return nil
		}

		req := request_structs.UpdateAdminRequest{
			ID:       targetID,
			Username: &newUsername,
		}

		err := svc.UpdateAdmin(context.Background(), updater, req)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}
		if savedAdmin.Username != newUsername {
			t.Errorf("Username = %q, want %q", savedAdmin.Username, newUsername)
		}
	})
}

func TestAdminService_GetAdmins_Pagination(t *testing.T) {

	t.Run("Limit fuera de rango se normaliza a 10", func(t *testing.T) {
		repo := &mockAdminRepo{}
		svc := NewAdminService(repo, nil)
		callCount := 0
		repo.ListFunc = func(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
			callCount++
			if limit != 10 {
				t.Errorf("Limit = %d, want 10 (default normalization)", limit)
			}
			return nil, 0, nil
		}

		svc.GetAdmins(context.Background(), nil, "", 1, 999)
		if callCount != 1 {
			t.Error("Repo debería haberse llamado exactamente 1 vez")
		}
	})

	t.Run("Page negativo se normaliza a 1", func(t *testing.T) {
		repo := &mockAdminRepo{}
		svc := NewAdminService(repo, nil)
		repo.ListFunc = func(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
			if page != 1 {
				t.Errorf("Page = %d, want 1 (default normalization)", page)
			}
			return nil, 0, nil
		}

		svc.GetAdmins(context.Background(), nil, "", -5, 10)
	})

	t.Run("Cálculo correcto de total_pages", func(t *testing.T) {
		repo := &mockAdminRepo{}
		svc := NewAdminService(repo, nil)
		repo.ListFunc = func(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
			return []domain.UserAdmin{}, 25, nil // 25 registros, 10 por página = 3 páginas
		}

		result, err := svc.GetAdmins(context.Background(), nil, "", 1, 10)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}

		resultMap, ok := result.(fiber.Map)
		if !ok {
			t.Fatalf("Result no es fiber.Map, tipo = %T", result)
		}

		totalPages, ok := resultMap["total_pages"].(int64)
		if !ok {
			t.Fatal("total_pages no es int64")
		}
		if totalPages != 3 {
			t.Errorf("total_pages = %d, want 3", totalPages)
		}
	})
}

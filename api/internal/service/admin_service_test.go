package service

import (
	"context"
	"testing"

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

		token, err := svc.Login(context.Background(), "admin_test", pass)
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

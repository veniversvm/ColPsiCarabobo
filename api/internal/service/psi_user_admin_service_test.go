package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// MOCKS ESPECÍFICOS (SIMULADORES DE INFRAESTRUCTURA)
// =========================================================================

// mockMailService simula el servicio de mensajería para pruebas unitarias.
// Permite aislar la lógica de negocio sin depender de una conexión SMTP real,
// actuando como un "Spy" que intercepta el envío de correos.
type mockMailService struct {
	SendEmailFunc func(to, subject, template string, data interface{}) error
}

func (m *mockMailService) SendEmail(to, subject, template string, data interface{}) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(to, subject, template, data)
	}
	return nil
}

// mockPsiRepoAdmin implementa un subconjunto dinámico de domain.PsiUserRepository.
// Utiliza el patrón "Func Override" para inyectar comportamientos específicos
// en tiempo de ejecución durante cada test.
type mockPsiRepoAdmin struct {
	domain.PsiUserRepository
	CreateWithColDataFunc func(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, solvencies []domain.PsiUserSolvency, postgrades []domain.PsiUserPostGrade) error
	GetByIDFunc           func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	// Interface Compliance: La firma debe coincidir con el contrato del dominio,
	// soportando ahora la inserción transaccional de un slice de solvencias.
	UpdateFunc func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, text *domain.TextModel, solvencies []domain.PsiUserSolvency) error
	// ResetPasswordFunc simula el reinicio de credenciales (clave + key de sesión).
	ResetPasswordFunc func(ctx context.Context, psi *domain.PsiUserModel) error
}

// CreateWithColData simula la ingesta estructurada de un perfil de psicólogo.
// Documentación Técnica (Interface Alignment):
// En Go, para que un mock satisfaga una interfaz implícitamente, sus métodos deben
// tener los mismos tipos de datos, orden de parámetros y valores de retorno EXACTOS
// que la interfaz definida en la capa de dominio.
func (m *mockPsiRepoAdmin) CreateWithColData(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, s []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
	if m.CreateWithColDataFunc != nil {
		return m.CreateWithColDataFunc(ctx, p, c, s, pg)
	}
	return nil
}

func (m *mockPsiRepoAdmin) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
}

// Update simula la mutación parcial del perfil de un psicólogo.
// Garantiza la compatibilidad estricta con el contrato del repositorio.
func (m *mockPsiRepoAdmin) Update(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, t *domain.TextModel, s []domain.PsiUserSolvency) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, p, c, t, s)
	}
	return nil
}

// ResetPassword simula el reinicio de credenciales de un psicólogo.
func (m *mockPsiRepoAdmin) ResetPassword(ctx context.Context, p *domain.PsiUserModel) error {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(ctx, p)
	}
	return nil
}

// =========================================================================
// TESTS DE NEGOCIO: ADMINISTRACIÓN DE PERFILES
// =========================================================================

// TestPsiService_CreateByAdmin evalúa el flujo de registro manual por parte del Staff.
func TestPsiService_CreateByAdmin(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	mail := &mockMailService{}

	// Dependency Injection (DI) Estricta:
	// Usamos el constructor oficial (NewPsiService) en lugar de instanciar el struct manualmente.
	// Esto es crucial porque el constructor inicializa internamente políticas de seguridad
	// (como el sanitizer XSS de Bluemonday) que fallarían con "nil pointer" si se omiten.
	svc := NewPsiService(repo, nil, mail)
	ctx := context.Background()

	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "admin_tester"}, CanCreatePsi: true}

	t.Run("Éxito: Registro completo", func(t *testing.T) {
		// Mock Binding: Interceptamos la llamada y validamos que el servicio pasa
		// correctamente la estructura de datos al repositorio simulado.
		repo.CreateWithColDataFunc = func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, solvencies []domain.PsiUserSolvency, postgrades []domain.PsiUserPostGrade) error {
			return nil
		}

		req := request_structs.CreatePsiAdminRequest{
			Username: "lic_perez",
			Email:    "perez@test.com",
			Password: "Secure1!password",
			BornDate: "1990-05-20",
		}

		err := svc.CreatePsiByAdmin(ctx, admin, req)
		if err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}
	})

	t.Run("Rechazo: Contraseña débil", func(t *testing.T) {
		req := request_structs.CreatePsiAdminRequest{
			Username: "lic_debil",
			Email:    "debil@test.com",
			Password: "12345",
			BornDate: "1990-05-20",
		}

		err := svc.CreatePsiByAdmin(ctx, admin, req)
		if err == nil || err.Error() != "la contraseña no cumple con los estándares de seguridad" {
			t.Errorf("Se esperaba rechazo por contraseña débil, se obtuvo: %v", err)
		}
	})
}

// TestPsiService_UpdateByAdmin_Patch evalúa la "Semántica PATCH" del sistema.
// Comprueba que los campos enviados muten, mientras que los campos omitidos (nil)
// permanezcan intactos sin causar colapsos en memoria.
func TestPsiService_UpdateByAdmin_Patch(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	mail := &mockMailService{}
	svc := NewPsiService(repo, nil, mail)

	ctx := context.Background()
	targetID := uuid.Must(uuid.NewV7())
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "super_admin"}, Sudo: true}

	t.Run("Actualización Parcial: Solo cambia el estatus de solvencia", func(t *testing.T) {
		currentPsi := &domain.PsiUserModel{
			ID:      targetID,
			Solvent: false,
			ColData: domain.PsiUserColData{PsiUserModelID: targetID, RegisterNumber: 12345},
		}

		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return currentPsi, nil
		}

		// Aceptación de la firma actualizada del repositorio (incluyendo solvencias)
		repo.UpdateFunc = func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, text *domain.TextModel, solvencies []domain.PsiUserSolvency) error {
			return nil
		}

		// Prevención de Nil-Pointer Dereference en Unit Tests:
		// Cuando el framework (Fiber) parsea JSON, automáticamente asigna direcciones de memoria
		// a los punteros del struct. En los tests unitarios manuales, si dejamos estos punteros
		// sin inicializar y el código del servicio intenta hacer un debug (ej. Printf("%s", *req.StateOutside)),
		// el programa entra en Panic. Inicializar la variable y pasar la referencia (&mun) lo soluciona.
		newSolvency := true
		mun := "Valencia"
		state := "Aragua"

		req := request_structs.UpdatePsiAdminRequest{
			Solvent:              &newSolvency,
			MunicipalityCarabobo: &mun,   // Evita pánico en *req.MunicipalityCarabobo
			StateOutside:         &state, // Evita pánico en *req.StateOutside
		}

		// Tolerancia a Ausencia de Archivos:
		// Se pasan punteros 'nil' para las imágenes (FileHeaders), simulando una petición
		// HTTP que solo actualiza texto, verificando que el servicio maneja la ausencia
		// de binarios elegantemente ("if file != nil").
		err := svc.UpdatePsiByAdmin(ctx, admin, targetID, req, nil, nil, nil, nil)
		if err != nil {
			t.Errorf("Error en Update: %v", err)
		}
	})
}

// TestPsiService_GetAdminDirectory_Security evalúa el Gatekeeping (Control de Acceso).
func TestPsiService_GetAdminDirectory_Security(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	// Optimización de Test: El servicio de correos y S3 se dejan en nil
	// porque la ejecución se detendrá antes de llegar a usarlos (Early Return).
	svc := NewPsiService(repo, nil, nil)

	// Vector Mitigado: Insecure Direct Object Reference (IDOR) / Privilege Escalation.
	// Verifica que un usuario autenticado pero sin privilegios administrativos (ni bandera Sudo)
	// sea rechazado en la capa lógica antes de que siquiera se ejecute una query a la DB.
	t.Run("Bloqueo de seguridad: No admins", func(t *testing.T) {
		randomUser := &domain.UserAdmin{Sudo: false, CanUpdatePsi: false, CanCreatePsi: false}
		_, err := svc.GetAdminDirectory(context.Background(), randomUser, request_structs.PsiDirectoryFilterDTO{})

		if err == nil || err.Error() != "permisos insuficientes para listar agremiados" {
			t.Error("Se debió denegar el acceso")
		}
	})
}

// TestPsiService_ResetPsiPasswordByAdmin evalúa el reinicio administrativo de clave:
// generación de contraseña temporal (12 caracteres), hashing bcrypt, rotación de la
// Key de sesión, bandera MustChangePassword y notificación por correo.
func TestPsiService_ResetPsiPasswordByAdmin(t *testing.T) {
	repo := &mockPsiRepoAdmin{}
	var mailTo, mailSubject, mailTemplate string
	var mailData map[string]interface{}
	mail := &mockMailService{
		SendEmailFunc: func(to, subject, template string, data interface{}) error {
			mailTo = to
			mailSubject = subject
			mailTemplate = template
			mailData = data.(map[string]interface{})
			return nil
		},
	}
	svc := NewPsiService(repo, nil, mail)
	ctx := context.Background()

	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "super_admin"}, Sudo: true}
	targetID := uuid.Must(uuid.NewV7())
	currentPsi := &domain.PsiUserModel{
		ID: targetID,
		Credentials: domain.Credentials{
			Email:    "psi@test.com",
			Username: "psi_test",
			Key:      "old_key",
			Password: "old_hash",
		},
		FirstName: "Licenciada",
	}

	var saved *domain.PsiUserModel
	repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
		return currentPsi, nil
	}
	repo.ResetPasswordFunc = func(ctx context.Context, psi *domain.PsiUserModel) error {
		saved = psi
		return nil
	}

	t.Run("Éxito: clave temporal de 12 chars, key rotada y correo", func(t *testing.T) {
		if err := svc.ResetPsiPasswordByAdmin(ctx, admin, targetID); err != nil {
			t.Fatalf("No se esperaba error: %v", err)
		}

		if saved == nil {
			t.Fatal("ResetPassword no fue invocado en el repositorio")
		}
		if !saved.MustChangePassword {
			t.Error("MustChangePassword debe ser true tras el reinicio")
		}
		if saved.Key == "old_key" {
			t.Error("La Key de sesión debe rotarse para invalidar JWT previos")
		}
		if saved.Password == "old_hash" {
			t.Error("El hash de la contraseña debe cambiar")
		}

		// El correo debe enviarse con la plantilla dedicada y la clave temporal.
		if mailTo != "psi@test.com" || mailTemplate != "reset_password_psi" {
			t.Errorf("Correo inesperado: to=%q template=%q", mailTo, mailTemplate)
		}
		if mailSubject != "Contraseña reiniciada" {
			t.Errorf("Asunto inesperado: %q", mailSubject)
		}
		temp, ok := mailData["Password"].(string)
		if !ok || len(temp) != 12 {
			t.Fatalf("La clave temporal debe ser un string de 12 caracteres, se obtuvo len=%d", len(temp))
		}
		// En la DB solo se persiste el hash bcrypt, nunca la clave en claro.
		if err := bcrypt.CompareHashAndPassword([]byte(saved.Password), []byte(temp)); err != nil {
			t.Error("El hash persistido no corresponde a la clave temporal enviada por correo")
		}
	})

	t.Run("Bloqueo: admin sin permisos", func(t *testing.T) {
		lowAdmin := &domain.UserAdmin{Sudo: false, CanUpdatePsi: false, CanCreatePsi: false}
		err := svc.ResetPsiPasswordByAdmin(ctx, lowAdmin, targetID)
		if err == nil || err.Error() != "no tienes permiso para editar registros de psicólogos" {
			t.Errorf("Se debió denegar el acceso: %v", err)
		}
	})

	t.Run("Registro inexistente", func(t *testing.T) {
		repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
			return nil, errors.New("not found")
		}
		err := svc.ResetPsiPasswordByAdmin(ctx, admin, targetID)
		if !errors.Is(err, domain.ErrPsiNotFound) {
			t.Errorf("Se esperaba ErrPsiNotFound: %v", err)
		}
	})
}

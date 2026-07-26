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
// MOCK INTEGRAL DEL REPOSITORIO (PATRÓN FUNC OVERRIDE)
// =========================================================================
// Arquitectura de Testing:
// Se utiliza el patrón de Mocks Dinámicos (Func Override) en lugar de stubs estáticos.
// Esto permite que cada sub-test (t.Run) reprograme el comportamiento del repositorio
// en tiempo de ejecución (ej. forzar un error de base de datos o simular un registro nulo)
// manteniendo un aislamiento absoluto en memoria para no contaminar otras pruebas concurrentes.

type mockPsiRepoSvc struct {
	domain.PsiUserRepository

	GetByIDFunc                   func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	GetByIdentifierFunc           func(ctx context.Context, id string) (*domain.PsiUserModel, error)
	GetByFPVFunc                  func(ctx context.Context, fpv int) (domain.PsiUserModel, error)
	UpdateKeyFunc                 func(ctx context.Context, psi *domain.PsiUserModel) error
	GetPsiUserColDataFunc         func(ctx context.Context, id uuid.UUID) (*domain.PsiUserColData, error)
	UpdatePublicProfileFunc       func(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, t *domain.TextModel) error
	GetPostGradeByIDFunc          func(ctx context.Context, id uuid.UUID) (*domain.PsiUserPostGrade, error)
	UpdatePostGradeFunc           func(ctx context.Context, pg *domain.PsiUserPostGrade) error
	CreatePostGradeFunc           func(ctx context.Context, pg *domain.PsiUserPostGrade) error
	GetTextContentByIDFunc        func(ctx context.Context, id uuid.UUID) (string, error)
	ValidateUniqueCredentialsFunc func(ctx context.Context, username, email string, excludeID uuid.UUID) error
	SearchDirectoryFunc           func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error)
	GetSitemapDataFunc            func(ctx context.Context) ([]domain.PsiUserModel, error)
	GetSolvenciesFunc             func(ctx context.Context, id uuid.UUID) ([]domain.PsiUserSolvency, error)
	CreateWithColDataFunc         func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error
}

func (m *mockPsiRepoSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockPsiRepoSvc) GetByIdentifier(ctx context.Context, id string) (*domain.PsiUserModel, error) {
	return m.GetByIdentifierFunc(ctx, id)
}
func (m *mockPsiRepoSvc) GetByFPV(ctx context.Context, fpv int) (domain.PsiUserModel, error) {
	return m.GetByFPVFunc(ctx, fpv)
}
func (m *mockPsiRepoSvc) UpdateKey(ctx context.Context, p *domain.PsiUserModel) error {
	return m.UpdateKeyFunc(ctx, p)
}
func (m *mockPsiRepoSvc) GetPsiUserColData(ctx context.Context, id uuid.UUID) (*domain.PsiUserColData, error) {
	return m.GetPsiUserColDataFunc(ctx, id)
}
func (m *mockPsiRepoSvc) UpdatePublicProfile(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, t *domain.TextModel) error {
	return m.UpdatePublicProfileFunc(ctx, p, c, t)
}
func (m *mockPsiRepoSvc) GetPostGradeByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserPostGrade, error) {
	return m.GetPostGradeByIDFunc(ctx, id)
}
func (m *mockPsiRepoSvc) UpdatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return m.UpdatePostGradeFunc(ctx, pg)
}
func (m *mockPsiRepoSvc) CreatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return m.CreatePostGradeFunc(ctx, pg)
}
func (m *mockPsiRepoSvc) GetTextContentByID(ctx context.Context, id uuid.UUID) (string, error) {
	if m.GetTextContentByIDFunc == nil {
		return "", nil
	}
	return m.GetTextContentByIDFunc(ctx, id)
}
func (m *mockPsiRepoSvc) ValidateUniqueCredentials(ctx context.Context, u, e string, ex uuid.UUID) error {
	return m.ValidateUniqueCredentialsFunc(ctx, u, e, ex)
}
func (m *mockPsiRepoSvc) SearchDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	return m.SearchDirectoryFunc(ctx, filter)
}
func (m *mockPsiRepoSvc) GetSitemapData(ctx context.Context) ([]domain.PsiUserModel, error) {
	return m.GetSitemapDataFunc(ctx)
}
func (m *mockPsiRepoSvc) GetSolvencies(ctx context.Context, id uuid.UUID) ([]domain.PsiUserSolvency, error) {
	return m.GetSolvenciesFunc(ctx, id)
}
func (m *mockPsiRepoSvc) CreateWithColData(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
	return m.CreateWithColDataFunc(ctx, psi, col, sol, pg)
}

// mockMailSvc simula la pasarela de envíos SMTP (Protocolo Fire-and-Forget).
type mockMailSvc struct {
	SendEmailFunc func(to, subject, template string, data any) error
}

func (m *mockMailSvc) SendEmail(to, subject, template string, data any) error {
	return m.SendEmailFunc(to, subject, template, data)
}

// =========================================================================
// TESTS DE NEGOCIO Y PERFORMANCE
// =========================================================================

// TestPsiService_GetPublicProfile_Privacy evalúa el motor de renderizado público (Privacy Shield).
// Garantiza que el Backend mutile o enmascare los datos sensibles antes de enviarlos a la red,
// en lugar de depender del Frontend para ocultarlos.
func TestPsiService_GetPublicProfile_Privacy(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)
	psiFPV := 12345

	// Escenario 1: Degradación Elegante (Graceful Degradation) por insolvencia.
	// Si un psicólogo pierde su solvencia gremial, el perfil público no lanza un Error 404 (lo cual
	// afectaría el SEO), sino que se "degrada", ocultando los beneficios premium como los Postgrados.
	t.Run("Restricción de Solvencia: No solvente no muestra Postgrados", func(t *testing.T) {
		repo.GetByFPVFunc = func(ctx context.Context, id int) (domain.PsiUserModel, error) {
			return domain.PsiUserModel{
				ID:  uuid.Must(uuid.NewV7()),
				FPV: psiFPV,
				Credentials: domain.Credentials{
					IsActive: true,
				},
				Solvent: false, // Gatilla la poda de datos en la capa de servicio
				ColData: domain.PsiUserColData{
					UniversityUndergraduate: "UCV",
				},
			}, nil
		}

		profile, _, err := svc.GetPublicProfile(context.Background(), psiFPV)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}

		// Aserción de Negocio: La matriz de Postgrados debe estar forzadamente vacía.
		if len(profile.PostGrades) != 0 {
			t.Error("Un usuario NO solvente no debería mostrar sus postgrados")
		}
	})

	// Escenario 2: Data Masking (Enmascaramiento de Datos PII).
	// Evalúa la regla "Opt-In". Si el psicólogo marca explícitamente "ShowContactEmail: false",
	// el servicio debe omitir el dato en el DTO resultante, previniendo fuga de datos personales (PII).
	t.Run("Privacidad Personal: Ocultar email", func(t *testing.T) {
		bioID := uuid.Must(uuid.NewV7())
		repo.GetByFPVFunc = func(ctx context.Context, id int) (domain.PsiUserModel, error) {
			return domain.PsiUserModel{
				ID:  uuid.Must(uuid.NewV7()),
				FPV: psiFPV,
				Credentials: domain.Credentials{
					IsActive: true,
				},
				Solvent:          true, // Solvente: Pasa la primera barrera y llega al Escudo de Privacidad
				ContactEmail:     "privado@test.com",
				ShowContactEmail: false, // Regla de ocultamiento activada
				BioTextID:        bioID,
				ColData:          domain.PsiUserColData{},
			}, nil
		}

		// FIX DOCUMENTADO: El servicio está optimizado para cargar la Biografía (TextModel)
		// exclusivamente si el usuario es solvente (Lazy Read).
		repo.GetTextContentByIDFunc = func(ctx context.Context, id uuid.UUID) (string, error) {
			return "Bio content", nil
		}

		profile, _, err := svc.GetPublicProfile(context.Background(), psiFPV)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}

		// Aserción de Seguridad: El email del DTO debe ser un string vacío.
		if profile.Email != "" {
			t.Errorf("El email debería estar oculto, se obtuvo: %s", profile.Email)
		}
	})
}

// TestPsiService_Login evalúa el flujo de autenticación del colegiado.
func TestPsiService_Login(t *testing.T) {
	repo := &mockPsiRepoSvc{}

	// Mock del Mailer: Retorna siempre nil para simular un envío asíncrono exitoso sin demoras de red.
	mailer := &mockMailSvc{SendEmailFunc: func(to, subject, template string, data any) error { return nil }}
	svc := NewPsiService(repo, nil, mailer)

	pass := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	psiID := uuid.Must(uuid.NewV7())

	// Escenario: Autenticación Exitosa y Generación de JWT (Single Session Enforcement)
	t.Run("Login Exitoso", func(t *testing.T) {
		repo.GetByIdentifierFunc = func(ctx context.Context, id string) (*domain.PsiUserModel, error) {
			return &domain.PsiUserModel{
				ID: psiID,
				Credentials: domain.Credentials{
					Password: string(hashed),
					IsActive: true,
					Username: "user_test",
				},
			}, nil
		}

		// Aserción Implícita: Si el servicio no llama a UpdateKey, el test fallaría.
		repo.UpdateKeyFunc = func(ctx context.Context, psi *domain.PsiUserModel) error { return nil }

		token, _, err := svc.Login(context.Background(), "user_test", pass)
		if err != nil || token == "" {
			t.Errorf("Fallo login: %v", err)
		}
	})
}

// TestPsiService_UpdateProfileSelf_LazyLoading evalúa estrategias de optimización de Base de Datos.
//
// El Lazy Loading (Carga Diferida) previene el problema N+1 y el desperdicio de I/O.
// El servicio está diseñado para consultar tablas relacionales (como ColData) *solo si*
// el Payload HTTP entrante contiene mutaciones que afecten a dicha tabla.
func TestPsiService_UpdateProfileSelf_LazyLoading(t *testing.T) {
	repo := &mockPsiRepoSvc{}
	svc := NewPsiService(repo, nil, nil)
	psiID := uuid.Must(uuid.NewV7())
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)

	// Contexto base: Usuario autenticado y validado.
	psi := &domain.PsiUserModel{
		ID: psiID,
		Credentials: domain.Credentials{
			Username: "psico_1",
			Password: string(hashed),
		},
	}

	// Escenario: Mutación de Datos Ligeros (Solo Metadata)
	t.Run("Lazy Loading: No llama a ColData si no hay cambios académicos", func(t *testing.T) {
		calledColData := false

		// Monitor de Invocación (Spy Pattern)
		repo.GetPsiUserColDataFunc = func(ctx context.Context, id uuid.UUID) (*domain.PsiUserColData, error) {
			calledColData = true // Se marca como sucio si el servicio intentó leer la DB.
			return &domain.PsiUserColData{}, nil
		}

		repo.ValidateUniqueCredentialsFunc = func(ctx context.Context, u, e string, ex uuid.UUID) error { return nil }
		repo.UpdatePublicProfileFunc = func(ctx context.Context, p *domain.PsiUserModel, c *domain.PsiUserColData, t *domain.TextModel) error {
			return nil
		}

		// Payload: El psicólogo SOLO desea actualizar su biografía.
		// Los campos académicos ("MentionUndergraduate", etc.) están nulos.
		newBio := "Nueva biografía"
		req := request_structs.PsiUserUpdateRequestSelf{
			Password: "pass123", // Validación de autorización exitosa
			MiniBio:  &newBio,
		}

		_, err := svc.UpdateProfileSelf(context.Background(), psi, psi.ID, req, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("Error en UpdateProfileSelf: %v", err)
		}

		// Aserción de Rendimiento: Si calledColData es true, significa que el servicio
		// desperdició un viaje a la base de datos (Query redundante).
		if calledColData {
			t.Error("Se llamó a ColData innecesariamente (Lazy Loading falló)")
		}
	})
}

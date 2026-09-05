package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// mockInscriptionRepo es un mock del repositorio de inscripciones acotado a la
// ficha (unicidad excluyente + CRUD de documentos). Usa el patrón "Func Override".
type mockInscriptionRepo struct {
	domain.InscriptionRepository
	CIInPsiUsersFunc              func(ctx context.Context, ci int) (bool, error)
	FPVInPsiUsersFunc             func(ctx context.Context, fpv int) (bool, error)
	EmailInPsiUsersFunc           func(ctx context.Context, email string) (bool, error)
	ExistsPendingCIFunc           func(ctx context.Context, ci int) (bool, error)
	ExistsPendingCIExcludingFunc  func(ctx context.Context, ci int, exclude uuid.UUID) (bool, error)
	ExistsPendingFPVExcludingFunc func(ctx context.Context, fpv int, exclude uuid.UUID) (bool, error)
	ExistsPendingEmailExcludingFunc func(ctx context.Context, email string, exclude uuid.UUID) (bool, error)
	ExistsPendingEmailFunc        func(ctx context.Context, email string) (bool, error)
	GetByIDFunc                   func(ctx context.Context, id uuid.UUID) (*domain.PsiInscriptionRequest, error)
	UpdateFunc                    func(ctx context.Context, req *domain.PsiInscriptionRequest) error
	NextControlNumberFunc         func(ctx context.Context) (int, error)
	ListDocumentsFunc             func(ctx context.Context, reqID uuid.UUID) ([]domain.PsiInscriptionDocument, error)
	DeleteDocsByRequestFunc       func(ctx context.Context, reqID uuid.UUID) error
	DeleteFunc                    func(ctx context.Context, id uuid.UUID) error
	SearchFunc                    func(ctx context.Context, filter request_structs.InscriptionListFilter) ([]domain.PsiInscriptionRequest, int64, error)
}

func (m *mockInscriptionRepo) CIInPsiUsers(ctx context.Context, ci int) (bool, error) {
	if m.CIInPsiUsersFunc == nil {
		return false, nil
	}
	return m.CIInPsiUsersFunc(ctx, ci)
}
func (m *mockInscriptionRepo) FPVInPsiUsers(ctx context.Context, fpv int) (bool, error) {
	if m.FPVInPsiUsersFunc == nil {
		return false, nil
	}
	return m.FPVInPsiUsersFunc(ctx, fpv)
}
func (m *mockInscriptionRepo) EmailInPsiUsers(ctx context.Context, email string) (bool, error) {
	if m.EmailInPsiUsersFunc == nil {
		return false, nil
	}
	return m.EmailInPsiUsersFunc(ctx, email)
}
func (m *mockInscriptionRepo) ExistsPendingCI(ctx context.Context, ci int) (bool, error) {
	if m.ExistsPendingCIFunc == nil {
		return false, nil
	}
	return m.ExistsPendingCIFunc(ctx, ci)
}
func (m *mockInscriptionRepo) ExistsPendingCIExcluding(ctx context.Context, ci int, exclude uuid.UUID) (bool, error) {
	if m.ExistsPendingCIExcludingFunc == nil {
		return false, nil
	}
	return m.ExistsPendingCIExcludingFunc(ctx, ci, exclude)
}
func (m *mockInscriptionRepo) ExistsPendingFPVExcluding(ctx context.Context, fpv int, exclude uuid.UUID) (bool, error) {
	if m.ExistsPendingFPVExcludingFunc == nil {
		return false, nil
	}
	return m.ExistsPendingFPVExcludingFunc(ctx, fpv, exclude)
}
func (m *mockInscriptionRepo) ExistsPendingEmail(ctx context.Context, email string) (bool, error) {
	if m.ExistsPendingEmailFunc == nil {
		return false, nil
	}
	return m.ExistsPendingEmailFunc(ctx, email)
}
func (m *mockInscriptionRepo) ExistsPendingEmailExcluding(ctx context.Context, email string, exclude uuid.UUID) (bool, error) {
	if m.ExistsPendingEmailExcludingFunc == nil {
		return false, nil
	}
	return m.ExistsPendingEmailExcludingFunc(ctx, email, exclude)
}
func (m *mockInscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiInscriptionRequest, error) {
	if m.GetByIDFunc == nil {
		return nil, ErrInscriptionNotFound
	}
	return m.GetByIDFunc(ctx, id)
}
func (m *mockInscriptionRepo) Update(ctx context.Context, req *domain.PsiInscriptionRequest) error {
	if m.UpdateFunc == nil {
		return nil
	}
	return m.UpdateFunc(ctx, req)
}
func (m *mockInscriptionRepo) NextControlNumber(ctx context.Context) (int, error) {
	if m.NextControlNumberFunc == nil {
		return 1000, nil
	}
	return m.NextControlNumberFunc(ctx)
}
func (m *mockInscriptionRepo) ListDocumentsByRequestID(ctx context.Context, reqID uuid.UUID) ([]domain.PsiInscriptionDocument, error) {
	if m.ListDocumentsFunc == nil {
		return nil, nil
	}
	return m.ListDocumentsFunc(ctx, reqID)
}
func (m *mockInscriptionRepo) DeleteInscriptionDocumentsByRequestID(ctx context.Context, reqID uuid.UUID) error {
	if m.DeleteDocsByRequestFunc == nil {
		return nil
	}
	return m.DeleteDocsByRequestFunc(ctx, reqID)
}
func (m *mockInscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc == nil {
		return nil
	}
	return m.DeleteFunc(ctx, id)
}
func (m *mockInscriptionRepo) Search(ctx context.Context, filter request_structs.InscriptionListFilter) ([]domain.PsiInscriptionRequest, int64, error) {
	if m.SearchFunc == nil {
		return nil, 0, nil
	}
	return m.SearchFunc(ctx, filter)
}

// mockPsiRepoInscripcion es un mock del repositorio de psicólogos acotado al
// flujo de aprobación de una inscripción.
type mockPsiRepoInscripcion struct {
	domain.PsiUserRepository
	CreateWithColDataFunc func(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, solvencies []domain.PsiUserSolvency, postgrades []domain.PsiUserPostGrade) error
	CreateDocumentFunc    func(ctx context.Context, doc *domain.PsiUserDocument) error
	GetSolvenciesFunc     func(ctx context.Context, psiID uuid.UUID) ([]domain.PsiUserSolvency, error)
}

func (m *mockPsiRepoInscripcion) CreateWithColData(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, solvencies []domain.PsiUserSolvency, postgrades []domain.PsiUserPostGrade) error {
	if m.CreateWithColDataFunc == nil {
		return nil
	}
	return m.CreateWithColDataFunc(ctx, psi, colData, solvencies, postgrades)
}
func (m *mockPsiRepoInscripcion) CreateDocument(ctx context.Context, doc *domain.PsiUserDocument) error {
	if m.CreateDocumentFunc == nil {
		return nil
	}
	return m.CreateDocumentFunc(ctx, doc)
}
func (m *mockPsiRepoInscripcion) GetSolvencies(ctx context.Context, psiID uuid.UUID) ([]domain.PsiUserSolvency, error) {
	if m.GetSolvenciesFunc == nil {
		return nil, nil
	}
	return m.GetSolvenciesFunc(ctx, psiID)
}

func TestInscriptionService_CheckEmail(t *testing.T) {
	ctx := context.Background()
	repo := &mockInscriptionRepo{}
	svc := NewInscriptionService(repo, nil, nil, &mockMailService{})

	t.Run("correo sin uso", func(t *testing.T) {
		res, err := svc.CheckEmail(ctx, "nuevo@test.com")
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if res.Exists {
			t.Fatal("se esperaba Exists=false")
		}
	})

	t.Run("correo ya registrado en psi_users", func(t *testing.T) {
		repo.EmailInPsiUsersFunc = func(ctx context.Context, email string) (bool, error) { return true, nil }
		res, err := svc.CheckEmail(ctx, "usado@test.com")
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if !res.Exists {
			t.Fatal("se esperaba Exists=true")
		}
	})

	t.Run("correo con solicitud pendiente", func(t *testing.T) {
		repo.EmailInPsiUsersFunc = func(ctx context.Context, email string) (bool, error) { return false, nil }
		repo.ExistsPendingEmailFunc = func(ctx context.Context, email string) (bool, error) { return true, nil }
		res, err := svc.CheckEmail(ctx, "pendiente@test.com")
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if !res.Exists {
			t.Fatal("se esperaba Exists=true")
		}
	})
}

func TestInscriptionService_Permisos(t *testing.T) {
	ctx := context.Background()
	id := uuid.Must(uuid.NewV7())

	adminSinPermisos := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "solo_lectura"}}
	adminSudo := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "root"}, Sudo: true}

	t.Run("List: sin permisos de gestión → ErrPermissionDenied", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		_, err := svc.List(ctx, adminSinPermisos, request_structs.InscriptionListFilter{})
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("Detail: sin permisos de gestión → ErrPermissionDenied", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		_, err := svc.Detail(ctx, adminSinPermisos, id)
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("List: con permiso de edición → OK e incluye ficha", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		editAdmin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "editor"}, CanUpdatePsi: true}
		repo.SearchFunc = func(ctx context.Context, filter request_structs.InscriptionListFilter) ([]domain.PsiInscriptionRequest, int64, error) {
			return []domain.PsiInscriptionRequest{{ID: id, Cedula: 1, Nombres: "A", Apellidos: "B", Status: domain.InscriptionPending}}, 1, nil
		}
		res, err := svc.List(ctx, editAdmin, request_structs.InscriptionListFilter{})
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if len(res.Items) != 1 {
			t.Fatalf("esperaba 1 solicitud, got %d", len(res.Items))
		}
	})

	t.Run("UpdateNotes: sin permiso → ErrPermissionDenied", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		err := svc.UpdateNotes(ctx, adminSinPermisos, id, "nota")
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("Approve: requiere CanCreatePsi", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		_, err := svc.Approve(ctx, adminSinPermisos, id)
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("Reject: requiere CanDeletePsi", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		err := svc.Reject(ctx, adminSinPermisos, id)
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("UpdateFicha: admin con permisos aprobado por Sudo", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		req := &domain.PsiInscriptionRequest{
			ID: id, Cedula: 10, Nacionalidad: "V", Nombres: "Ana",
			Apellidos: "Lopez", Correo: "ana@test.com", Status: domain.InscriptionPending,
		}
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		repo.UpdateFunc = func(ctx context.Context, r *domain.PsiInscriptionRequest) error {
			if r.Cedula != 20 {
				t.Fatalf("la cédula no se actualizó: %v", r.Cedula)
			}
			if r.ServiceAddress != "Av. Bolívar 1" {
				t.Fatalf("la dirección no se actualizó: %q", r.ServiceAddress)
			}
			if !r.ServiceModalityPresencial {
				t.Fatal("la modalidad presencial no se guardó")
			}
			return nil
		}
		body := &request_structs.UpdateInscriptionRequest{
			Cedula: 20, Nacionalidad: "V", Nombres: "Ana", Apellidos: "Lopez",
			Correo: "ana@test.com", ServiceAddress: "Av. Bolívar 1",
			ServiceModalityPresencial: true,
		}
		dto, err := svc.UpdateFicha(ctx, adminSudo, id, body)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if dto == nil {
			t.Fatal("esperaba DTO")
		}
	})
}

func TestInscriptionService_UpdateFicha_UnicidadExcluyente(t *testing.T) {
	ctx := context.Background()
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "editor"}, CanUpdatePsi: true}
	id := uuid.Must(uuid.NewV7())

	req := &domain.PsiInscriptionRequest{
		ID: id, Cedula: 10, Nacionalidad: "V", Nombres: "Ana",
		Apellidos: "Lopez", Correo: "ana@test.com", Status: domain.InscriptionPending,
	}

	t.Run("cédula duplicada en otra solicitud pendiente → ErrCIExists", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		repo.ExistsPendingCIExcludingFunc = func(ctx context.Context, ci int, exclude uuid.UUID) (bool, error) { return true, nil }
		_, err := svc.UpdateFicha(ctx, admin, id, &request_structs.UpdateInscriptionRequest{
			Cedula: 99, Correo: "ana@test.com",
		})
		if !errors.Is(err, ErrCIExists) {
			t.Fatalf("esperaba ErrCIExists, got %v", err)
		}
	})

	t.Run("correo duplicado en psi_users → ErrEmailExists", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		repo.EmailInPsiUsersFunc = func(ctx context.Context, email string) (bool, error) { return true, nil }
		_, err := svc.UpdateFicha(ctx, admin, id, &request_structs.UpdateInscriptionRequest{
			Cedula: 10, Correo: "otro@test.com",
		})
		if !errors.Is(err, ErrEmailExists) {
			t.Fatalf("esperaba ErrEmailExists, got %v", err)
		}
	})

	t.Run("mismo correo que la propia solicitud → OK sin re-validar", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		repo.UpdateFunc = func(ctx context.Context, r *domain.PsiInscriptionRequest) error { return nil }
		_, err := svc.UpdateFicha(ctx, admin, id, &request_structs.UpdateInscriptionRequest{
			Cedula: 10, Correo: "ANA@test.com",
		})
		if err != nil {
			t.Fatalf("error inesperado al editar con correo sin cambios: %v", err)
		}
	})
}

func TestInscriptionService_Approve_MapFichaYMigra(t *testing.T) {
	ctx := context.Background()
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "aprobador"}, CanCreatePsi: true}
	id := uuid.Must(uuid.NewV7())

	specID := uint32(7)
	req := &domain.PsiInscriptionRequest{
		ID: id, Cedula: 100, Nacionalidad: "V", Nombres: "María", Apellidos: "Rojas",
		Correo: "maria@test.com", FPV: 401000, Status: domain.InscriptionPending,
		ServiceAddress:          "Urb. La Viña",
		MunicipalityCarabobo:    "Naguanagua",
		ServiceModalityDistance: true,
		PrimarySpecialtyID:      &specID,
	}
	doc := domain.PsiInscriptionDocument{
		ID: uuid.Must(uuid.NewV7()), InscriptionRequestID: id,
		DocumentType: domain.DocumentCedula, S3Key: "inscripciones/documentos/cedula/x.png",
		OriginalFilename: "x.png",
	}

	repo := &mockInscriptionRepo{}
	repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
	repo.ListDocumentsFunc = func(ctx context.Context, reqID uuid.UUID) ([]domain.PsiInscriptionDocument, error) { return []domain.PsiInscriptionDocument{doc}, nil }

	var savedPSI *domain.PsiUserModel
	psiRepo := &mockPsiRepoInscripcion{}
	psiRepo.CreateWithColDataFunc = func(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, solvencies []domain.PsiUserSolvency, postgrades []domain.PsiUserPostGrade) error {
		savedPSI = psi
		return nil
	}
	migrated := false
	psiRepo.CreateDocumentFunc = func(ctx context.Context, d *domain.PsiUserDocument) error {
		if d.PsiUserID != savedPSI.ID {
			t.Fatalf("documento migrado a otro psi: %v", d.PsiUserID)
		}
		if d.DocumentType != domain.DocumentCedula {
			t.Fatalf("tipo de documento incorrecto: %v", d.DocumentType)
		}
		migrated = true
		return nil
	}
	deleted := false
	repo.DeleteDocsByRequestFunc = func(ctx context.Context, reqID uuid.UUID) error { deleted = true; return nil }

	svc := NewInscriptionService(repo, psiRepo, nil, &mockMailService{})
	if _, err := svc.Approve(ctx, admin, id); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if savedPSI == nil {
		t.Fatal("no se creó el psicólogo")
	}
	if savedPSI.ServiceAddress != req.ServiceAddress {
		t.Fatalf("dirección no mapeada: %q", savedPSI.ServiceAddress)
	}
	if savedPSI.MunicipalityCarabobo != req.MunicipalityCarabobo {
		t.Fatalf("municipio no mapeado: %q", savedPSI.MunicipalityCarabobo)
	}
	if !savedPSI.ServiceModalityDistance {
		t.Fatal("modalidad a distancia no mapeada")
	}
	if savedPSI.PrimarySpecialtyID == nil || *savedPSI.PrimarySpecialtyID != specID {
		t.Fatalf("área principal no mapeada: %v", savedPSI.PrimarySpecialtyID)
	}
	if !migrated {
		t.Fatal("el documento no se migró al expediente")
	}
	if !deleted {
		t.Fatal("las filas de la ficha no se limpiaron tras migrar")
	}
}
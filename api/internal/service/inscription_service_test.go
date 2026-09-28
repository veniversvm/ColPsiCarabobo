package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// mockInscriptionRepo es un mock del repositorio de inscripciones acotado a la
// ficha (unicidad excluyente + CRUD de documentos). Usa el patrón "Func Override".
type mockInscriptionRepo struct {
	domain.InscriptionRepository
	CIInPsiUsersFunc                func(ctx context.Context, ci int) (bool, error)
	FPVInPsiUsersFunc               func(ctx context.Context, fpv int) (bool, error)
	EmailInPsiUsersFunc             func(ctx context.Context, email string) (bool, error)
	UsernameInPsiUsersFunc          func(ctx context.Context, username string) (bool, error)
	ExistsPendingCIFunc             func(ctx context.Context, ci int) (bool, error)
	ExistsPendingCIExcludingFunc    func(ctx context.Context, ci int, exclude uuid.UUID) (bool, error)
	ExistsPendingFPVExcludingFunc   func(ctx context.Context, fpv int, exclude uuid.UUID) (bool, error)
	ExistsPendingEmailExcludingFunc func(ctx context.Context, email string, exclude uuid.UUID) (bool, error)
	ExistsPendingEmailFunc          func(ctx context.Context, email string) (bool, error)
	GetByIDFunc                     func(ctx context.Context, id uuid.UUID) (*domain.PsiInscriptionRequest, error)
	UpdateFunc                      func(ctx context.Context, req *domain.PsiInscriptionRequest) error
	NextControlNumberFunc           func(ctx context.Context) (int, error)
	ListDocumentsFunc               func(ctx context.Context, reqID uuid.UUID) ([]domain.PsiInscriptionDocument, error)
	DeleteDocsByRequestFunc         func(ctx context.Context, reqID uuid.UUID) error
	DeleteFunc                      func(ctx context.Context, id uuid.UUID) error
	SearchFunc                      func(ctx context.Context, filter request_structs.InscriptionListFilter) ([]domain.PsiInscriptionRequest, int64, error)
	UpdateNotesFunc                 func(ctx context.Context, id uuid.UUID, notes string) error
	AppendNotesHistoryFunc          func(ctx context.Context, note *domain.PsiInscriptionNote) error
	ListNotesHistoryFunc            func(ctx context.Context, requestID uuid.UUID) ([]domain.PsiInscriptionNote, error)
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
func (m *mockInscriptionRepo) UsernameInPsiUsers(ctx context.Context, username string) (bool, error) {
	if m.UsernameInPsiUsersFunc == nil {
		return false, nil
	}
	return m.UsernameInPsiUsersFunc(ctx, username)
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
func (m *mockInscriptionRepo) UpdateNotes(ctx context.Context, id uuid.UUID, notes string) error {
	if m.UpdateNotesFunc == nil {
		return nil
	}
	return m.UpdateNotesFunc(ctx, id, notes)
}
func (m *mockInscriptionRepo) AppendNotesHistory(ctx context.Context, note *domain.PsiInscriptionNote) error {
	if m.AppendNotesHistoryFunc == nil {
		return nil
	}
	return m.AppendNotesHistoryFunc(ctx, note)
}
func (m *mockInscriptionRepo) ListNotesHistory(ctx context.Context, requestID uuid.UUID) ([]domain.PsiInscriptionNote, error) {
	if m.ListNotesHistoryFunc == nil {
		return nil, nil
	}
	return m.ListNotesHistoryFunc(ctx, requestID)
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
	svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})

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
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		_, err := svc.List(ctx, adminSinPermisos, request_structs.InscriptionListFilter{})
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("Detail: sin permisos de gestión → ErrPermissionDenied", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		_, err := svc.Detail(ctx, adminSinPermisos, id)
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("List: con permiso de edición → OK e incluye ficha", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
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
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		err := svc.UpdateNotes(ctx, adminSinPermisos, id, "nota")
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("Approve: requiere CanCreatePsi", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		_, err := svc.Approve(ctx, adminSinPermisos, id)
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("Reject: requiere CanDeletePsi", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		err := svc.Reject(ctx, adminSinPermisos, id)
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("UpdateFicha: admin con permisos aprobado por Sudo", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
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
			SegundoApellido: "Perez", Genero: "F", Telefono: "04141234567",
			Correo: "ana@test.com", ServiceAddress: "Av. Bolívar 1",
			MunicipalityCarabobo: "Valencia",
			TituloUniversidad:    "UC", TituloMencion: "Clínica",
			TituloRegistroNumero: "123", TituloRegistroEstado: "Carabobo",
			TituloRegistroTomo: "1", TituloRegistroFolio: "1",
			ServiceModalityPresencial: true,
		}
		fn := "2000-01-01"
		fg := "2010-01-01"
		body.FechaNacimiento = &fn
		body.TituloFechaGraduacion = &fg
		dto, err := svc.UpdateFicha(ctx, adminSudo, id, body)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if dto == nil {
			t.Fatal("esperaba DTO")
		}
	})
}

func TestInscriptionService_NotasHistorico(t *testing.T) {
	ctx := context.Background()
	id := uuid.Must(uuid.NewV7())
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "jefe"}, Sudo: true}
	viewer := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "solo_lectura"}}

	t.Run("UpdateNotes: con cambio persiste nota y agrega versión al histórico", func(t *testing.T) {
		var updated, appended string
		var appendedBy string
		var appendedByID *uuid.UUID
		svc := NewInscriptionService(&mockInscriptionRepo{
			GetByIDFunc: func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) {
				return &domain.PsiInscriptionRequest{ID: i, Nombres: "Ana", Apellidos: "Perez", FPV: 123, Notes: "versión anterior"}, nil
			},
			UpdateNotesFunc: func(ctx context.Context, i uuid.UUID, notes string) error { updated = notes; return nil },
			AppendNotesHistoryFunc: func(ctx context.Context, note *domain.PsiInscriptionNote) error {
				appended = note.Notes
				appendedBy = note.CreateBy
				appendedByID = note.CreateById
				return nil
			},
		}, nil, nil, nil, &mockMailService{})

		if err := svc.UpdateNotes(ctx, admin, id, "nueva versión"); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if updated != "nueva versión" {
			t.Fatalf("nota persistida = %q, se esperaba %q", updated, "nueva versión")
		}
		if appended != "nueva versión" {
			t.Fatalf("versión histórica = %q, se esperaba %q", appended, "nueva versión")
		}
		if appendedBy != "jefe" || appendedByID == nil || *appendedByID != admin.ID {
			t.Fatalf("autor de la versión = %q/%v, se esperaba jefe/%v", appendedBy, appendedByID, admin.ID)
		}
	})

	t.Run("UpdateNotes: mismo texto no reescribe ni versiona", func(t *testing.T) {
		reewrote, versioned := false, false
		svc := NewInscriptionService(&mockInscriptionRepo{
			GetByIDFunc: func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) {
				return &domain.PsiInscriptionRequest{ID: i, Notes: "igual"}, nil
			},
			UpdateNotesFunc: func(ctx context.Context, i uuid.UUID, notes string) error { reewrote = true; return nil },
			AppendNotesHistoryFunc: func(ctx context.Context, note *domain.PsiInscriptionNote) error {
				versioned = true
				return nil
			},
		}, nil, nil, nil, &mockMailService{})

		if err := svc.UpdateNotes(ctx, admin, id, "igual"); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if reewrote || versioned {
			t.Fatalf("se reescribió (update=%v) o versionó (append=%v) sin cambio real", reewrote, versioned)
		}
	})

	t.Run("NotesHistory: sin permiso de gestión → ErrPermissionDenied", func(t *testing.T) {
		svc := NewInscriptionService(&mockInscriptionRepo{}, nil, nil, nil, &mockMailService{})
		if _, err := svc.NotesHistory(ctx, viewer, id); !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("esperaba ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("NotesHistory: solicitud inexistente → ErrInscriptionNotFound", func(t *testing.T) {
		svc := NewInscriptionService(&mockInscriptionRepo{}, nil, nil, nil, &mockMailService{})
		if _, err := svc.NotesHistory(ctx, admin, id); !errors.Is(err, ErrInscriptionNotFound) {
			t.Fatalf("esperaba ErrInscriptionNotFound, got %v", err)
		}
	})

	t.Run("NotesHistory: devuelve versiones de la más reciente a la más antigua", func(t *testing.T) {
		svc := NewInscriptionService(&mockInscriptionRepo{
			GetByIDFunc: func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) {
				return &domain.PsiInscriptionRequest{ID: i}, nil
			},
			ListNotesHistoryFunc: func(ctx context.Context, requestID uuid.UUID) ([]domain.PsiInscriptionNote, error) {
				return []domain.PsiInscriptionNote{
					{ID: uuid.Must(uuid.NewV7()), InscriptionRequestID: requestID, Notes: "nueva", AuditModel: domain.AuditModel{CreateBy: "jefe", CreatedAt: time.Now()}},
				}, nil
			},
		}, nil, nil, nil, &mockMailService{})

		items, err := svc.NotesHistory(ctx, admin, id)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if len(items) != 1 || items[0].Notes != "nueva" || items[0].CreateBy != "jefe" {
			t.Fatalf("histórico = %+v, se esperaba [nueva/jefe]", items)
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

	// validUpdateFicha construye un payload completo (pasa la validación de
	// campos obligatorios) para poder ejercitar la unicidad excluyente.
	valid := func(correo string) *request_structs.UpdateInscriptionRequest {
		fn := "2000-01-01"
		fg := "2010-01-01"
		return &request_structs.UpdateInscriptionRequest{
			Cedula: 10, Nacionalidad: "V", Nombres: "Ana", Apellidos: "Lopez",
			SegundoApellido: "Perez", Genero: "F", Telefono: "04141234567",
			Correo: correo, FechaNacimiento: &fn,
			TituloUniversidad: "UC", TituloFechaGraduacion: &fg, TituloMencion: "Clínica",
			TituloRegistroNumero: "123", TituloRegistroEstado: "Carabobo",
			TituloRegistroTomo: "1", TituloRegistroFolio: "1",
			ServiceAddress: "Av. Bolívar 1", MunicipalityCarabobo: "Valencia",
		}
	}

	t.Run("cédula duplicada en otra solicitud pendiente → ErrCIExists", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		repo.ExistsPendingCIExcludingFunc = func(ctx context.Context, ci int, exclude uuid.UUID) (bool, error) { return true, nil }
		body := valid("ana@test.com")
		body.Cedula = 99
		_, err := svc.UpdateFicha(ctx, admin, id, body)
		if !errors.Is(err, ErrCIExists) {
			t.Fatalf("esperaba ErrCIExists, got %v", err)
		}
	})

	t.Run("correo duplicado en psi_users → ErrEmailExists", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		repo.EmailInPsiUsersFunc = func(ctx context.Context, email string) (bool, error) { return true, nil }
		_, err := svc.UpdateFicha(ctx, admin, id, valid("otro@test.com"))
		if !errors.Is(err, ErrEmailExists) {
			t.Fatalf("esperaba ErrEmailExists, got %v", err)
		}
	})

	t.Run("mismo correo que la propia solicitud → OK sin re-validar", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		repo.UpdateFunc = func(ctx context.Context, r *domain.PsiInscriptionRequest) error { return nil }
		_, err := svc.UpdateFicha(ctx, admin, id, valid("ANA@test.com"))
		if err != nil {
			t.Fatalf("error inesperado al editar con correo sin cambios: %v", err)
		}
	})

	t.Run("cédula cero → rechazo inmediato (nunca una ficha con cédula 0)", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		body := valid("ana@test.com")
		body.Cedula = 0
		_, err := svc.UpdateFicha(ctx, admin, id, body)
		if err == nil || err.Error() != "la cédula debe ser un número positivo" {
			t.Fatalf("esperaba rechazo por cédula 0, got %v", err)
		}
	})
}

func TestInscriptionService_Approve_MapFichaYMigra(t *testing.T) {
	ctx := context.Background()
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "aprobador"}, CanCreatePsi: true}
	id := uuid.Must(uuid.NewV7())

	specID := uint32(7)
	nac := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	grad := time.Date(2015, 7, 1, 0, 0, 0, 0, time.UTC)
	req := &domain.PsiInscriptionRequest{
		ID: id, Cedula: 100, Nacionalidad: "V", Nombres: "María", Apellidos: "Rojas",
		SegundoNombre: "Luz", SegundoApellido: "Villalobos", Genero: "F",
		Telefono: "04141234567", Correo: "maria@test.com", FPV: 401000,
		FechaNacimiento: &nac, Status: domain.InscriptionPending,
		TituloUniversidad: "Universidad de Carabobo", TituloFechaGraduacion: &grad,
		TituloRegistroEstado: "Carabobo",
		ServiceAddress:          "Urb. La Viña",
		MunicipalityCarabobo:    "Naguanagua",
		ServiceModalityDistance: true,
		PrimarySpecialtyID:      &specID,
		ComprobanteS3Key:        "inscripciones/comprobantes/y.png",
	}
	doc := domain.PsiInscriptionDocument{
		ID: uuid.Must(uuid.NewV7()), InscriptionRequestID: id,
		DocumentType: domain.DocumentCedula, S3Key: "inscripciones/documentos/cedula/x.png",
		OriginalFilename: "x.png",
	}

	repo := &mockInscriptionRepo{}
	repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
	repo.ListDocumentsFunc = func(ctx context.Context, reqID uuid.UUID) ([]domain.PsiInscriptionDocument, error) {
		return []domain.PsiInscriptionDocument{doc}, nil
	}

	var savedPSI *domain.PsiUserModel
	psiRepo := &mockPsiRepoInscripcion{}
	psiRepo.CreateWithColDataFunc = func(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, solvencies []domain.PsiUserSolvency, postgrades []domain.PsiUserPostGrade) error {
		savedPSI = psi
		return nil
	}
	migrated := false
	comprobanteMigrated := false
	psiRepo.CreateDocumentFunc = func(ctx context.Context, d *domain.PsiUserDocument) error {
		if d.PsiUserID != savedPSI.ID {
			t.Fatalf("documento migrado a otro psi: %v", d.PsiUserID)
		}
		switch d.DocumentType {
		case domain.DocumentCedula:
			if d.S3Key != "inscripciones/documentos/cedula/x.png" || d.Title != "Cédula de identidad (copia)" {
				t.Fatalf("documento de cédula mal mapeado: %+v", d)
			}
			migrated = true
		case domain.DocumentComprobante:
			if d.S3Key != "inscripciones/comprobantes/y.png" || d.Notes != "N° de control 1000" {
				t.Fatalf("comprobante mal mapeado: %+v", d)
			}
			comprobanteMigrated = true
		default:
			t.Fatalf("tipo de documento inesperado: %v", d.DocumentType)
		}
		return nil
	}
	deleted := false
	repo.DeleteDocsByRequestFunc = func(ctx context.Context, reqID uuid.UUID) error { deleted = true; return nil }

	svc := NewInscriptionService(repo, psiRepo, nil, nil, &mockMailService{})
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
	if !savedPSI.Credentials.IsActive {
		t.Fatal("la cuenta debería nacer activa al aprobar la inscripción")
	}
	if !savedPSI.Solvent {
		t.Fatal("el psicólogo debería nacer solvente al aprobar la inscripción")
	}
	if !savedPSI.ProofOfLife {
		t.Fatal("el psicólogo debería nacer con fe de vida activa al aprobar la inscripción")
	}
	if !migrated {
		t.Fatal("el documento no se migró al expediente")
	}
	if !comprobanteMigrated {
		t.Fatal("el comprobante de pago no se migró al expediente")
	}
	if !deleted {
		t.Fatal("las filas de la ficha no se limpiaron tras migrar")
	}
}

// readyInscriptionRequest construye una ficha completa y válida (pasa la regla
// de campos obligatorios y la identidad legal) para ejercitar el gate de approve.
func readyInscriptionRequest(id uuid.UUID) *domain.PsiInscriptionRequest {
	nac := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	grad := time.Date(2015, 7, 1, 0, 0, 0, 0, time.UTC)
	return &domain.PsiInscriptionRequest{
		ID: id, Cedula: 100, Nacionalidad: "V", Nombres: "María", Apellidos: "Rojas",
		SegundoNombre: "Luz", SegundoApellido: "Villalobos", Genero: "F",
		Telefono: "04141234567", Correo: "maria@test.com", FPV: 401000,
		FechaNacimiento: &nac, Status: domain.InscriptionPending,
		TituloUniversidad: "Universidad de Carabobo", TituloFechaGraduacion: &grad,
		TituloRegistroEstado: "Carabobo",
		ServiceAddress: "Urb. La Viña", MunicipalityCarabobo: "Naguanagua",
	}
}

// TestInscriptionService_Approve_RechazaCeroYNegativos garantiza que ninguna
// ficha con cédula o FPV 0 pueda convertirse en psicólogo (dicho perfil es
// inalcanzable en el directorio público). El gate devuelve ErrInscriptionNotReady
// con el detalle, nunca un 500.
func TestInscriptionService_Approve_RechazaCeroYNegativos(t *testing.T) {
	ctx := context.Background()
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "aprobador"}, CanCreatePsi: true}
	id := uuid.Must(uuid.NewV7())

	t.Run("ficha con FPV cero → ErrInscriptionNotReady", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		bad := readyInscriptionRequest(id)
		bad.FPV = 0
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return bad, nil }
		_, err := svc.Approve(ctx, admin, id)
		if !errors.Is(err, ErrInscriptionNotReady) {
			t.Fatalf("esperaba ErrInscriptionNotReady al aprobar ficha con FPV 0, got %v", err)
		}
		if !strings.Contains(err.Error(), "N° FPV") {
			t.Fatalf("el mensaje debe mencionar el N° FPV, got %v", err)
		}
	})

	t.Run("ficha con cédula cero → ErrInscriptionNotReady", func(t *testing.T) {
		repo := &mockInscriptionRepo{}
		svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
		bad := readyInscriptionRequest(id)
		bad.Cedula = 0
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return bad, nil }
		_, err := svc.Approve(ctx, admin, id)
		if !errors.Is(err, ErrInscriptionNotReady) {
			t.Fatalf("esperaba ErrInscriptionNotReady al aprobar ficha con cédula 0, got %v", err)
		}
		if !strings.Contains(err.Error(), "cédula") {
			t.Fatalf("el mensaje debe mencionar la cédula, got %v", err)
		}
	})
}

// TestInscriptionService_Approve_GateIntegral cubre el gate de aprobación:
// la ficha no se aprueba si cualquier identificador único (CI, FPV, correo o
// username generado) ya pertenece a otro psicólogo, ni si faltan campos
// obligatorios. Todos los problemas se reportan juntos (nunca un 500).
func TestInscriptionService_Approve_GateIntegral(t *testing.T) {
	ctx := context.Background()
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "aprobador"}, CanCreatePsi: true}
	id := uuid.Must(uuid.NewV7())

	newSvc := func(mut func(*mockInscriptionRepo)) (*InscriptionService, *domain.PsiInscriptionRequest) {
		repo := &mockInscriptionRepo{}
		req := readyInscriptionRequest(id)
		repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) { return req, nil }
		if mut != nil {
			mut(repo)
		}
		return NewInscriptionService(repo, nil, nil, nil, &mockMailService{}), req
	}

	assertNotReady := func(t *testing.T, err error, fragment string) {
		t.Helper()
		if !errors.Is(err, ErrInscriptionNotReady) {
			t.Fatalf("esperaba ErrInscriptionNotReady, got %v", err)
		}
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("el mensaje debe contener %q, got %v", fragment, err)
		}
	}

	t.Run("cédula duplicada en psi_users", func(t *testing.T) {
		svc, _ := newSvc(func(r *mockInscriptionRepo) {
			r.CIInPsiUsersFunc = func(ctx context.Context, ci int) (bool, error) { return true, nil }
		})
		_, err := svc.Approve(ctx, admin, id)
		assertNotReady(t, err, "la cédula ya se encuentra registrada en el directorio")
	})

	t.Run("FPV duplicado en psi_users", func(t *testing.T) {
		svc, _ := newSvc(func(r *mockInscriptionRepo) {
			r.FPVInPsiUsersFunc = func(ctx context.Context, fpv int) (bool, error) { return true, nil }
		})
		_, err := svc.Approve(ctx, admin, id)
		assertNotReady(t, err, "el N° FPV ya se encuentra registrado en el directorio")
	})

	t.Run("correo duplicado en psi_users", func(t *testing.T) {
		svc, _ := newSvc(func(r *mockInscriptionRepo) {
			r.EmailInPsiUsersFunc = func(ctx context.Context, email string) (bool, error) { return true, nil }
		})
		_, err := svc.Approve(ctx, admin, id)
		assertNotReady(t, err, "el correo ya se encuentra registrado en el directorio")
	})

	t.Run("username generado duplicado", func(t *testing.T) {
		svc, _ := newSvc(func(r *mockInscriptionRepo) {
			r.UsernameInPsiUsersFunc = func(ctx context.Context, username string) (bool, error) { return true, nil }
		})
		_, err := svc.Approve(ctx, admin, id)
		assertNotReady(t, err, "el nombre de usuario generado ya existe")
	})

	t.Run("se reportan todos los problemas juntos", func(t *testing.T) {
		svc, req := newSvc(func(r *mockInscriptionRepo) {
			r.EmailInPsiUsersFunc = func(ctx context.Context, email string) (bool, error) { return true, nil }
		})
		req.FPV = 0
		req.SegundoApellido = ""
		_, err := svc.Approve(ctx, admin, id)
		if !errors.Is(err, ErrInscriptionNotReady) {
			t.Fatalf("esperaba ErrInscriptionNotReady, got %v", err)
		}
		var notReady *InscriptionNotReadyError
		if !errors.As(err, &notReady) {
			t.Fatalf("esperaba *InscriptionNotReadyError, got %T", err)
		}
		// [segundo apellido obligatorio, N° FPV positivo, correo registrado]
		if len(notReady.Issues) != 3 {
			t.Fatalf("esperaba 3 problemas, got %v", notReady.Issues)
		}
		for _, frag := range []string{"N° FPV", "segundo apellido", "correo ya se encuentra registrado"} {
			if !strings.Contains(err.Error(), frag) {
				t.Fatalf("el mensaje debe contener %q, got %v", frag, err)
			}
		}
	})

	t.Run("ficha completa sin conflictos → supera el gate", func(t *testing.T) {
		svc, _ := newSvc(nil)
		issues, err := svc.approvalReadinessIssues(ctx, readyInscriptionRequest(id))
		if err != nil {
			t.Fatalf("error inesperado en el gate: %v", err)
		}
		if len(issues) != 0 {
			t.Fatalf("la ficha completa no debería tener problemas de aprobación, got %v", issues)
		}
	})
}

// TestValidateFichaObligatoria cubre la regla única de campos obligatorios de
// la ficha (personales, académicos y bloques de ubicación).
func TestValidateFichaObligatoria(t *testing.T) {
	valid := FichaObligatoria{
		SegundoApellido:         "Perez",
		Genero:                  "F",
		Telefono:                "04141234567",
		FechaNacimientoPresente: true,
		TituloUniversidad:       "UC",
		FechaGraduacionPresente: true,
		TituloRegistroEstado:    "Carabobo",
		MunicipalityCarabobo:    "Valencia",
		ServiceAddress:          "Av. Bolívar 1",
	}

	t.Run("ficha completa → nil", func(t *testing.T) {
		if err := ValidateFichaObligatoria(valid); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
	})

	missing := []struct {
		name string
		mut  func(*FichaObligatoria)
		want string
	}{
		{"segundo apellido", func(f *FichaObligatoria) { f.SegundoApellido = "" }, "el segundo apellido es obligatorio"},
		{"género", func(f *FichaObligatoria) { f.Genero = " " }, "el género es obligatorio"},
		{"teléfono", func(f *FichaObligatoria) { f.Telefono = "" }, "el teléfono de contacto es obligatorio"},
		{"fecha nacimiento", func(f *FichaObligatoria) { f.FechaNacimientoPresente = false }, "la fecha de nacimiento es obligatoria"},
		{"universidad", func(f *FichaObligatoria) { f.TituloUniversidad = "" }, "la universidad es obligatoria"},
		{"fecha graduación", func(f *FichaObligatoria) { f.FechaGraduacionPresente = false }, "la fecha de graduación es obligatoria"},
		{"estado registro", func(f *FichaObligatoria) { f.TituloRegistroEstado = "" }, "el estado del registro es obligatorio"},
	}
	for _, tc := range missing {
		t.Run("falta "+tc.name, func(t *testing.T) {
			f := valid
			tc.mut(&f)
			err := ValidateFichaObligatoria(f)
			if err == nil {
				t.Fatal("esperaba error de validación")
			}
			if err.Error() != tc.want {
				t.Fatalf("esperaba %q, got %q", tc.want, err.Error())
			}
		})
	}
}

func TestValidateFichaObligatoria_Ubicacion(t *testing.T) {
	base := FichaObligatoria{
		SegundoApellido:         "Perez",
		Genero:                  "F",
		Telefono:                "04141234567",
		FechaNacimientoPresente: true,
		TituloUniversidad:       "UC",
		FechaGraduacionPresente: true,
		TituloRegistroEstado:    "Carabobo",
	}

	t.Run("sin ubicación → error", func(t *testing.T) {
		err := ValidateFichaObligatoria(base)
		if err == nil {
			t.Fatal("esperaba error de ubicación")
		}
		if err.Error() != "debes completar al menos una ubicación completa (Carabobo, otro estado o exterior)" {
			t.Fatalf("mensaje inesperado: %v", err)
		}
	})

	t.Run("Carabobo incompleto (solo municipio) → error", func(t *testing.T) {
		f := base
		f.MunicipalityCarabobo = "Valencia"
		if err := ValidateFichaObligatoria(f); err == nil {
			t.Fatal("esperaba error de ubicación")
		}
	})

	t.Run("Otro estado incompleto (solo estado) → error", func(t *testing.T) {
		f := base
		f.StateOutside = "Lara"
		if err := ValidateFichaObligatoria(f); err == nil {
			t.Fatal("esperaba error de ubicación")
		}
	})

	t.Run("Carabobo completo → ok", func(t *testing.T) {
		f := base
		f.MunicipalityCarabobo = "Valencia"
		f.ServiceAddress = "Av. Bolívar 1"
		if err := ValidateFichaObligatoria(f); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
	})

	t.Run("Otro estado completo → ok", func(t *testing.T) {
		f := base
		f.StateOutside = "Lara"
		f.MunicipalityOutside = "Barquisimeto"
		if err := ValidateFichaObligatoria(f); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
	})

	t.Run("Exterior completo → ok", func(t *testing.T) {
		f := base
		f.Country = "España"
		if err := ValidateFichaObligatoria(f); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
	})
}

// TestInscriptionService_UpdateFicha_CamposObligatorios verifica que el PATCH
// admin rechaza una ficha incompleta con ValidationError (HTTP 400).
func TestInscriptionService_UpdateFicha_CamposObligatorios(t *testing.T) {
	ctx := context.Background()
	admin := &domain.UserAdmin{ID: uuid.Must(uuid.NewV7()), Credentials: domain.Credentials{Username: "editor"}, CanUpdatePsi: true}
	id := uuid.Must(uuid.NewV7())

	repo := &mockInscriptionRepo{}
	svc := NewInscriptionService(repo, nil, nil, nil, &mockMailService{})
	repo.GetByIDFunc = func(ctx context.Context, i uuid.UUID) (*domain.PsiInscriptionRequest, error) {
		return &domain.PsiInscriptionRequest{ID: id, Cedula: 10, Correo: "ana@test.com"}, nil
	}

	var ve *ValidationError
	_, err := svc.UpdateFicha(ctx, admin, id, &request_structs.UpdateInscriptionRequest{
		Cedula: 10, Nacionalidad: "V", Nombres: "Ana", Apellidos: "Lopez", Correo: "ana@test.com",
	})
	if !errors.As(err, &ve) {
		t.Fatalf("esperaba ValidationError, got %v", err)
	}
	if ve.Msg != "el segundo apellido es obligatorio" {
		t.Fatalf("mensaje inesperado: %s", ve.Msg)
	}
}

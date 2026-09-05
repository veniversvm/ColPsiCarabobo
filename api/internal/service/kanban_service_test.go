package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// =========================================================================
// MOCK DEL REPOSITORIO KANBAN (Patrón Func Override usando Embedding)
// =========================================================================

type mockKanbanRepo struct {
	domain.KanbanRepository
	CreateProjectFunc  func(ctx context.Context, p *domain.KanbanProject) error
	GetProjectByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.KanbanProject, error)
	ListProjectsFunc   func(ctx context.Context, adminID uuid.UUID, isMaster bool) ([]domain.KanbanProject, error)
	UpdateProjectFunc  func(ctx context.Context, p *domain.KanbanProject) error
	DeleteProjectFunc  func(ctx context.Context, id uuid.UUID) error

	AddMemberFunc     func(ctx context.Context, m *domain.KanbanMember) error
	GetMemberFunc     func(ctx context.Context, projectID, userAdminID uuid.UUID) (*domain.KanbanMember, error)
	GetMemberByIDFunc func(ctx context.Context, memberID uuid.UUID) (*domain.KanbanMember, error)
	ListMembersFunc   func(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanMember, error)
	UpdateMemberFunc  func(ctx context.Context, m *domain.KanbanMember) error
	RemoveMemberFunc  func(ctx context.Context, memberID uuid.UUID) error

	CreateColumnFunc func(ctx context.Context, col *domain.KanbanColumn) error
	GetColumnFunc    func(ctx context.Context, id uuid.UUID) (*domain.KanbanColumn, error)
	GetColumnsFunc   func(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanColumn, error)
	UpdateColumnFunc func(ctx context.Context, col *domain.KanbanColumn) error
	DeleteColumnFunc func(ctx context.Context, id uuid.UUID) error

	GetBoardFunc   func(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanColumn, error)
	CreateCardFunc func(ctx context.Context, card *domain.KanbanCard) error
	GetCardFunc    func(ctx context.Context, id uuid.UUID) (*domain.KanbanCard, error)
	GetCardsFunc   func(ctx context.Context, columnID uuid.UUID) ([]domain.KanbanCard, error)
	UpdateCardFunc func(ctx context.Context, card *domain.KanbanCard) error
	DeleteCardFunc func(ctx context.Context, id uuid.UUID) error

	CreateNoteFunc func(ctx context.Context, n *domain.KanbanNote) error
	GetNoteFunc    func(ctx context.Context, id uuid.UUID) (*domain.KanbanNote, error)
	DeleteNoteFunc func(ctx context.Context, id uuid.UUID) error
	CountNotesFunc func(ctx context.Context, cardID uuid.UUID) (int64, error)

	CountMembersByProjectFunc func(ctx context.Context) (map[uuid.UUID]int64, error)
	CountCardsByProjectFunc   func(ctx context.Context) (map[uuid.UUID]int64, error)
}

func (m *mockKanbanRepo) CreateProject(ctx context.Context, p *domain.KanbanProject) error {
	return m.CreateProjectFunc(ctx, p)
}
func (m *mockKanbanRepo) GetProjectByID(ctx context.Context, id uuid.UUID) (*domain.KanbanProject, error) {
	return m.GetProjectByIDFunc(ctx, id)
}
func (m *mockKanbanRepo) ListProjects(ctx context.Context, adminID uuid.UUID, isMaster bool) ([]domain.KanbanProject, error) {
	return m.ListProjectsFunc(ctx, adminID, isMaster)
}
func (m *mockKanbanRepo) UpdateProject(ctx context.Context, p *domain.KanbanProject) error {
	return m.UpdateProjectFunc(ctx, p)
}
func (m *mockKanbanRepo) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return m.DeleteProjectFunc(ctx, id)
}
func (m *mockKanbanRepo) AddMember(ctx context.Context, mm *domain.KanbanMember) error {
	return m.AddMemberFunc(ctx, mm)
}
func (m *mockKanbanRepo) GetMember(ctx context.Context, projectID, userAdminID uuid.UUID) (*domain.KanbanMember, error) {
	return m.GetMemberFunc(ctx, projectID, userAdminID)
}
func (m *mockKanbanRepo) GetMemberByID(ctx context.Context, memberID uuid.UUID) (*domain.KanbanMember, error) {
	return m.GetMemberByIDFunc(ctx, memberID)
}
func (m *mockKanbanRepo) ListMembers(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanMember, error) {
	return m.ListMembersFunc(ctx, projectID)
}
func (m *mockKanbanRepo) UpdateMember(ctx context.Context, mm *domain.KanbanMember) error {
	return m.UpdateMemberFunc(ctx, mm)
}
func (m *mockKanbanRepo) RemoveMember(ctx context.Context, memberID uuid.UUID) error {
	return m.RemoveMemberFunc(ctx, memberID)
}
func (m *mockKanbanRepo) CreateColumn(ctx context.Context, col *domain.KanbanColumn) error {
	return m.CreateColumnFunc(ctx, col)
}
func (m *mockKanbanRepo) GetColumn(ctx context.Context, id uuid.UUID) (*domain.KanbanColumn, error) {
	return m.GetColumnFunc(ctx, id)
}
func (m *mockKanbanRepo) GetColumns(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanColumn, error) {
	return m.GetColumnsFunc(ctx, projectID)
}
func (m *mockKanbanRepo) UpdateColumn(ctx context.Context, col *domain.KanbanColumn) error {
	return m.UpdateColumnFunc(ctx, col)
}
func (m *mockKanbanRepo) DeleteColumn(ctx context.Context, id uuid.UUID) error {
	return m.DeleteColumnFunc(ctx, id)
}
func (m *mockKanbanRepo) GetBoard(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanColumn, error) {
	return m.GetBoardFunc(ctx, projectID)
}
func (m *mockKanbanRepo) CreateCard(ctx context.Context, card *domain.KanbanCard) error {
	return m.CreateCardFunc(ctx, card)
}
func (m *mockKanbanRepo) GetCard(ctx context.Context, id uuid.UUID) (*domain.KanbanCard, error) {
	return m.GetCardFunc(ctx, id)
}
func (m *mockKanbanRepo) GetCards(ctx context.Context, columnID uuid.UUID) ([]domain.KanbanCard, error) {
	return m.GetCardsFunc(ctx, columnID)
}
func (m *mockKanbanRepo) UpdateCard(ctx context.Context, card *domain.KanbanCard) error {
	return m.UpdateCardFunc(ctx, card)
}
func (m *mockKanbanRepo) DeleteCard(ctx context.Context, id uuid.UUID) error {
	return m.DeleteCardFunc(ctx, id)
}
func (m *mockKanbanRepo) CreateNote(ctx context.Context, n *domain.KanbanNote) error {
	return m.CreateNoteFunc(ctx, n)
}
func (m *mockKanbanRepo) GetNote(ctx context.Context, id uuid.UUID) (*domain.KanbanNote, error) {
	return m.GetNoteFunc(ctx, id)
}
func (m *mockKanbanRepo) DeleteNote(ctx context.Context, id uuid.UUID) error {
	return m.DeleteNoteFunc(ctx, id)
}
func (m *mockKanbanRepo) CountNotes(ctx context.Context, cardID uuid.UUID) (int64, error) {
	return m.CountNotesFunc(ctx, cardID)
}
func (m *mockKanbanRepo) CountMembersByProject(ctx context.Context) (map[uuid.UUID]int64, error) {
	return m.CountMembersByProjectFunc(ctx)
}
func (m *mockKanbanRepo) CountCardsByProject(ctx context.Context) (map[uuid.UUID]int64, error) {
	return m.CountCardsByProjectFunc(ctx)
}

// =========================================================================
// TESTS
// =========================================================================

func TestKanbanService_CreateProject_siembraColumnasPorDefecto(t *testing.T) {
	repo := &mockKanbanRepo{}
	createdColumns := 0
	repo.CreateProjectFunc = func(ctx context.Context, p *domain.KanbanProject) error {
		p.ID = uuid.New()
		return nil
	}
	repo.CreateColumnFunc = func(ctx context.Context, col *domain.KanbanColumn) error {
		createdColumns++
		return nil
	}
	repo.GetMemberFunc = func(ctx context.Context, projectID, userAdminID uuid.UUID) (*domain.KanbanMember, error) {
		return nil, errors.New("no row")
	}

	svc := NewKanbanService(repo, &mockAdminRepo{})
	admin := &domain.UserAdmin{ID: uuid.New(), Credentials: domain.Credentials{Username: "admin1"}}

	project, err := svc.CreateProject(context.Background(), admin, request_structs.CreateProjectRequest{Name: "  Convención 2026  "})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if project.OwnerID != admin.ID {
		t.Fatalf("owner incorrecto: esperado %v, obtuve %v", admin.ID, project.OwnerID)
	}
	if project.Name != "Convención 2026" {
		t.Fatalf("nombre sin sanitizar: %q", project.Name)
	}
	if createdColumns != 3 {
		t.Fatalf("se esperaban 3 columnas por defecto, se crearon %d", createdColumns)
	}
}

func TestKanbanService_CreateProject_nombreVacio(t *testing.T) {
	repo := &mockKanbanRepo{}
	svc := NewKanbanService(repo, &mockAdminRepo{})
	admin := &domain.UserAdmin{ID: uuid.New()}

	_, err := svc.CreateProject(context.Background(), admin, request_structs.CreateProjectRequest{Name: "   "})
	if err == nil {
		t.Fatal("se esperaba error por nombre vacío")
	}
}

func TestKanbanService_CreateNote_límiteDe10Notas(t *testing.T) {
	repo := &mockKanbanRepo{}
	cardID := uuid.New()
	projectID := uuid.New()
	ownerID := uuid.New()

	// El admin es el dueño del proyecto → puede editar.
	repo.GetProjectByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.KanbanProject, error) {
		return &domain.KanbanProject{ID: projectID, OwnerID: ownerID}, nil
	}
	repo.GetMemberFunc = func(ctx context.Context, projectID, userAdminID uuid.UUID) (*domain.KanbanMember, error) {
		return nil, errors.New("no row")
	}
	repo.GetCardFunc = func(ctx context.Context, id uuid.UUID) (*domain.KanbanCard, error) {
		return &domain.KanbanCard{ID: cardID, ProjectID: projectID}, nil
	}
	repo.CountNotesFunc = func(ctx context.Context, cardID uuid.UUID) (int64, error) {
		return MaxNotesPerCard, nil // ya hay 10 notas
	}

	svc := NewKanbanService(repo, &mockAdminRepo{})
	admin := &domain.UserAdmin{ID: ownerID, Credentials: domain.Credentials{Username: "owner"}}

	_, err := svc.CreateNote(context.Background(), admin, cardID, request_structs.CreateNoteRequest{Content: "nota 11"})
	if !errors.Is(err, domain.ErrNoteLimitReached) {
		t.Fatalf("se esperaba ErrNoteLimitReached, obtuve: %v", err)
	}
}

func TestKanbanService_CreateNote_excede500Caracteres(t *testing.T) {
	repo := &mockKanbanRepo{}
	cardID := uuid.New()
	projectID := uuid.New()
	ownerID := uuid.New()

	repo.GetProjectByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.KanbanProject, error) {
		return &domain.KanbanProject{ID: projectID, OwnerID: ownerID}, nil
	}
	repo.GetMemberFunc = func(ctx context.Context, projectID, userAdminID uuid.UUID) (*domain.KanbanMember, error) {
		return nil, errors.New("no row")
	}
	repo.GetCardFunc = func(ctx context.Context, id uuid.UUID) (*domain.KanbanCard, error) {
		return &domain.KanbanCard{ID: cardID, ProjectID: projectID}, nil
	}
	repo.CountNotesFunc = func(ctx context.Context, cardID uuid.UUID) (int64, error) {
		return 0, nil
	}

	svc := NewKanbanService(repo, &mockAdminRepo{})
	admin := &domain.UserAdmin{ID: ownerID, Credentials: domain.Credentials{Username: "owner"}}

	long := strings.Repeat("a", MaxNoteLengthChars+1)
	_, err := svc.CreateNote(context.Background(), admin, cardID, request_structs.CreateNoteRequest{Content: long})
	if !errors.Is(err, domain.ErrNoteTooLong) {
		t.Fatalf("se esperaba ErrNoteTooLong, obtuve: %v", err)
	}
}

func TestKanbanService_InvitarMiembro_soloDueñoOMaster(t *testing.T) {
	projectID := uuid.New()
	ownerID := uuid.New()
	targetID := uuid.New()

	tests := []struct {
		name    string
		admin   *domain.UserAdmin
		wantErr bool
		errIs   error
	}{
		{name: "dueño puede invitar", admin: &domain.UserAdmin{ID: ownerID}, wantErr: false},
		{name: "master (CanManageProjects) puede invitar", admin: &domain.UserAdmin{ID: uuid.New(), CanManageProjects: true}, wantErr: false},
		{name: "master (Sudo) puede invitar", admin: &domain.UserAdmin{ID: uuid.New(), Sudo: true}, wantErr: false},
		{name: "viewer no puede invitar", admin: &domain.UserAdmin{ID: uuid.New()}, wantErr: true, errIs: domain.ErrInsufficientPerms},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockKanbanRepo{}
			repo.GetProjectByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.KanbanProject, error) {
				return &domain.KanbanProject{ID: projectID, OwnerID: ownerID}, nil
			}
			repo.GetMemberFunc = func(ctx context.Context, pid, uid uuid.UUID) (*domain.KanbanMember, error) {
				if uid == tt.admin.ID {
					return &domain.KanbanMember{ProjectID: pid, UserAdminID: uid, Role: domain.MemberRoleViewer}, nil
				}
				return nil, errors.New("no row")
			}
			repo.AddMemberFunc = func(ctx context.Context, m *domain.KanbanMember) error { return nil }

			adminRepo := &mockAdminRepo{GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return &domain.UserAdmin{ID: id}, nil
			}}

			svc := NewKanbanService(repo, adminRepo)
			err := svc.AddMember(context.Background(), tt.admin, projectID, request_structs.AddMemberRequest{UserAdminID: targetID, Role: "editor"})
			if tt.wantErr {
				if err == nil || !errors.Is(err, tt.errIs) {
					t.Fatalf("se esperaba error %v, obtuve: %v", tt.errIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
		})
	}
}

func TestKanbanService_rolInvalido_alInvitar(t *testing.T) {
	projectID := uuid.New()
	ownerID := uuid.New()
	repo := &mockKanbanRepo{}
	repo.GetProjectByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.KanbanProject, error) {
		return &domain.KanbanProject{ID: projectID, OwnerID: ownerID}, nil
	}
	repo.GetMemberFunc = func(ctx context.Context, pid, uid uuid.UUID) (*domain.KanbanMember, error) {
		return nil, errors.New("no row")
	}

	svc := NewKanbanService(repo, &mockAdminRepo{})
	admin := &domain.UserAdmin{ID: ownerID}

	err := svc.AddMember(context.Background(), admin, projectID, request_structs.AddMemberRequest{
		UserAdminID: uuid.New(),
		Role:        "super-admin",
	})
	if !errors.Is(err, domain.ErrInvalidMemberRole) {
		t.Fatalf("se esperaba ErrInvalidMemberRole, obtuve: %v", err)
	}
}

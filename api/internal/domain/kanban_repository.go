// api/internal/domain/kanban_repository.go
package domain

import (
	"context"

	"github.com/google/uuid"
)

// KanbanRepository define el contrato de persistencia para el módulo de
// Proyectos (tableros Kanban) del panel administrativo.
// Sigue el principio de Dependency Inversion (DIP): la lógica de negocio
// depende de esta interfaz, no de la implementación concreta de DB.
type KanbanRepository interface {
	// ── Proyectos ─────────────────────────────────────────────────────────
	CreateProject(ctx context.Context, p *KanbanProject) error
	GetProjectByID(ctx context.Context, id uuid.UUID) (*KanbanProject, error)
	ListProjects(ctx context.Context, adminID uuid.UUID, isMaster bool) ([]KanbanProject, error)
	UpdateProject(ctx context.Context, p *KanbanProject) error
	DeleteProject(ctx context.Context, id uuid.UUID) error

	// ── Miembros ──────────────────────────────────────────────────────────
	AddMember(ctx context.Context, m *KanbanMember) error
	GetMember(ctx context.Context, projectID, userAdminID uuid.UUID) (*KanbanMember, error)
	GetMemberByID(ctx context.Context, memberID uuid.UUID) (*KanbanMember, error)
	ListMembers(ctx context.Context, projectID uuid.UUID) ([]KanbanMember, error)
	UpdateMember(ctx context.Context, m *KanbanMember) error
	RemoveMember(ctx context.Context, memberID uuid.UUID) error

	// ── Columnas ──────────────────────────────────────────────────────────
	CreateColumn(ctx context.Context, col *KanbanColumn) error
	GetColumn(ctx context.Context, id uuid.UUID) (*KanbanColumn, error)
	GetColumns(ctx context.Context, projectID uuid.UUID) ([]KanbanColumn, error)
	UpdateColumn(ctx context.Context, col *KanbanColumn) error
	DeleteColumn(ctx context.Context, id uuid.UUID) error

	// ── Tarjetas y tablero ────────────────────────────────────────────────
	GetBoard(ctx context.Context, projectID uuid.UUID) ([]KanbanColumn, error)
	CreateCard(ctx context.Context, card *KanbanCard) error
	GetCard(ctx context.Context, id uuid.UUID) (*KanbanCard, error)
	GetCards(ctx context.Context, columnID uuid.UUID) ([]KanbanCard, error)
	UpdateCard(ctx context.Context, card *KanbanCard) error
	DeleteCard(ctx context.Context, id uuid.UUID) error

	// ── Notas ─────────────────────────────────────────────────────────────
	CreateNote(ctx context.Context, n *KanbanNote) error
	GetNote(ctx context.Context, id uuid.UUID) (*KanbanNote, error)
	DeleteNote(ctx context.Context, id uuid.UUID) error
	CountNotes(ctx context.Context, cardID uuid.UUID) (int64, error)

	// ── Conteos agregados (listado de proyectos) ──────────────────────────
	CountMembersByProject(ctx context.Context) (map[uuid.UUID]int64, error)
	CountCardsByProject(ctx context.Context) (map[uuid.UUID]int64, error)
}

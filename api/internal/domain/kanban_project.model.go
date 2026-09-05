// api/internal/domain/kanban_project.model.go
package domain

import (
	"github.com/google/uuid"
)

// MemberRole define el nivel de acceso de un miembro dentro de un proyecto Kanban.
type MemberRole string

const (
	// MemberRoleViewer: solo lectura del tablero.
	MemberRoleViewer MemberRole = "viewer"
	// MemberRoleEditor: puede crear/editar/mover tarjetas y añadir notas.
	MemberRoleEditor MemberRole = "editor"
)

// IsValid valida que el rol sea uno de los soportados.
func (r MemberRole) IsValid() bool {
	return r == MemberRoleViewer || r == MemberRoleEditor
}

// KanbanProject es un proyecto dueño de un tablero Kanban.
// El creador (OwnerID) es el "dueño"; el acceso de otros admins se delega
// mediante KanbanMember. Un admin con Sudo o CanManageProjects actúa como
// master y administra cualquier proyecto.
type KanbanProject struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel

	Name        string    `gorm:"size:120;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	OwnerID     uuid.UUID `gorm:"type:uuid;index;not null" json:"owner_id"`

	// Relaciones
	Owner   *UserAdmin     `gorm:"foreignKey:OwnerID;references:ID" json:"owner,omitempty"`
	Members []KanbanMember `gorm:"foreignKey:ProjectID" json:"-"`
	Columns []KanbanColumn `gorm:"foreignKey:ProjectID" json:"-"`

	// Campos calculados (no persistidos)
	MemberCount int64      `gorm:"-" json:"member_count"`
	CardCount   int64      `gorm:"-" json:"card_count"`
	MyRole      MemberRole `gorm:"-" json:"my_role,omitempty"`
	IsMaster    bool       `gorm:"-" json:"is_master"`
	IsOwner     bool       `gorm:"-" json:"is_owner"`
}

func (KanbanProject) TableName() string { return "kanban_projects" }

// KanbanMember relaciona un administrador (user_admins) con un proyecto y su rol.
// La dupla (project_id, user_admin_id) es única: un admin solo puede tener un rol por proyecto.
type KanbanMember struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel

	ProjectID   uuid.UUID  `gorm:"type:uuid;index;uniqueIndex:idx_kanban_member_unique;not null" json:"project_id"`
	UserAdminID uuid.UUID  `gorm:"type:uuid;index;uniqueIndex:idx_kanban_member_unique;not null" json:"user_admin_id"`
	Role        MemberRole `gorm:"size:20;not null" json:"role"`

	User *UserAdmin `gorm:"foreignKey:UserAdminID;references:ID" json:"user,omitempty"`
}

func (KanbanMember) TableName() string { return "kanban_project_members" }

// KanbanColumn es una columna (lista) del tablero Kanban de un proyecto.
type KanbanColumn struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel

	ProjectID uuid.UUID `gorm:"type:uuid;index;not null" json:"project_id"`
	Title     string    `gorm:"size:120;not null" json:"title"`
	Position  int       `json:"position"`

	Cards []KanbanCard `gorm:"foreignKey:ColumnID" json:"cards,omitempty"`
}

func (KanbanColumn) TableName() string { return "kanban_columns" }

// KanbanCard es una tarjeta dentro de una columna del tablero.
type KanbanCard struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel

	ProjectID   uuid.UUID `gorm:"type:uuid;index;not null" json:"project_id"`
	ColumnID    uuid.UUID `gorm:"type:uuid;index;not null" json:"column_id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Description string    `gorm:"size:2000" json:"description"`
	Position    int       `json:"position"`

	Notes []KanbanNote `gorm:"foreignKey:CardID" json:"notes,omitempty"`
}

func (KanbanCard) TableName() string { return "kanban_cards" }

// KanbanNote es una nota dentro de una tarjeta.
// Reglas de negocio: máximo 10 notas por tarjeta y 500 caracteres por nota
// (se validan en la capa de servicio, nunca del lado del cliente).
type KanbanNote struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	AuditModel

	CardID  uuid.UUID `gorm:"type:uuid;index;not null" json:"card_id"`
	Content string    `gorm:"size:500;not null" json:"content"`
}

func (KanbanNote) TableName() string { return "kanban_card_notes" }

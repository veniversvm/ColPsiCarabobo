// api/internal/domain/ticket.model.go
// Módulo de Tickets de Solicitudes.
//
// A diferencia del resto del sistema (que usa UUIDs), todas las tablas de este
// módulo usan IDs `uint` auto-increment para acelerar el conteo y la consulta
// FIFO (orden de llegada) del panel administrativo. Las únicas FKs con tipo
// UUID son las que apuntan a psi_users / user_admins.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// AutorType identifica quién realizó una acción dentro de un ticket
// (mensaje en la conversación o cambio de estado).
type AutorType string

const (
	// AutorAdmin: el autor es un administrador del colegio.
	AutorAdmin AutorType = "admin"
	// AutorPsi: el autor es el psicólogo dueño del ticket.
	AutorPsi AutorType = "psi"
	// AutorSystem: la acción fue generada automáticamente por el sistema.
	AutorSystem AutorType = "system"
)

// TicketMotivo es un motivo de atención de un ticket definido por el colegio.
// Todos los tickets tienen un motivo y, según el motivo, un límite de tickets
// abiertos por psicólogo (TicketsPerPsi). Cada motivo tiene su propio conjunto
// de estados (TicketEstado), permitiendo que "el colegio defina los posibles
// estados para los tickets según su motivo". Al crear un motivo se siembran
// por defecto los estados: recibido → en proceso → cerrado.
type TicketMotivo struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	AuditModel

	Name          string `gorm:"size:120;not null" json:"name"`
	Description   string `gorm:"size:500" json:"description"`
	TicketsPerPsi int    `gorm:"not null;default:3" json:"tickets_per_psi"`

	Estados []TicketEstado `gorm:"foreignKey:MotivoID" json:"estados,omitempty"`
}

func (TicketMotivo) TableName() string { return "ticket_motivos" }

// TicketEstado es un estado posible que puede tomar un ticket. Los estados se
// definen por motivo (FK MotivoID). Los estados default al crear un motivo son:
// recibido(orden 1) → en proceso(orden 2) → cerrado(orden 3, es_cerrado=true).
type TicketEstado struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	AuditModel

	MotivoID  uint   `gorm:"index;uniqueIndex:idx_ticket_estados_motivo_nombre;not null" json:"motivo_id"`
	Name      string `gorm:"size:60;uniqueIndex:idx_ticket_estados_motivo_nombre;not null" json:"name"`
	Order     int    `gorm:"not null;default:0" json:"order"`
	IsClosed  bool   `gorm:"not null;default:false" json:"is_closed"`

	Motivo *TicketMotivo `gorm:"foreignKey:MotivoID;references:ID" json:"motivo,omitempty"`
}

func (TicketEstado) TableName() string { return "ticket_estados" }

// Ticket es la entidad principal del módulo. Solo el psicólogo crea tickets;
// el administrador los administra (respuesta, cambio de estado, cierre).
type Ticket struct {
	ID uint `gorm:"primaryKey;autoIncrement;index:idx_tickets_psi_user_id,priority:2;index:idx_tickets_estado_id,priority:2;index:idx_tickets_motivo_id,priority:2" json:"id"`
	AuditModel

	PsiUserID uuid.UUID `gorm:"type:uuid;index:idx_tickets_psi_user_id,priority:1;not null" json:"psi_user_id"`
	MotivoID  uint      `gorm:"index:idx_tickets_motivo_id,priority:1;not null" json:"motivo_id"`
	EstadoID  uint      `gorm:"index:idx_tickets_estado_id,priority:1;not null" json:"estado_id"`

	Title             string    `gorm:"size:200;not null" json:"title"`
	Description       string    `gorm:"type:text" json:"description"`        // descripción inicial del ticket (denormalizada para listados)
	CloseReason       string    `gorm:"size:500" json:"close_reason"`        // motivo de cierre (obligatorio al cerrar)
	ClosedByType      AutorType `gorm:"size:10" json:"closed_by_type"`       // admin | psi | ""
	ClosedByAdminID   *uuid.UUID `gorm:"type:uuid" json:"closed_by_admin_id,omitempty"`
	ClosedByPsiID     *uuid.UUID `gorm:"type:uuid" json:"closed_by_psi_id,omitempty"`
	ClosedAt          *time.Time `json:"closed_at,omitempty"`

	// Relaciones
	Psi     *PsiUserModel `gorm:"foreignKey:PsiUserID;references:ID" json:"-"`
	Motivo  *TicketMotivo `gorm:"foreignKey:MotivoID;references:ID" json:"motivo,omitempty"`
	Estado  *TicketEstado `gorm:"foreignKey:EstadoID;references:ID" json:"estado,omitempty"`

	Mensajes   []TicketMensaje    `gorm:"foreignKey:TicketID" json:"mensajes,omitempty"`
	StatusLogs []TicketStatusLog  `gorm:"foreignKey:TicketID" json:"status_logs,omitempty"`

	// Campos calculados (no persistidos) para el listado admin/psi.
	PsiFirstName string `gorm:"-" json:"psi_first_name,omitempty"`
	PsiLastName  string `gorm:"-" json:"psi_last_name,omitempty"`
	IsClosed     bool   `gorm:"-" json:"is_closed"`
}

func (Ticket) TableName() string { return "tickets" }

// TicketStatusLog registra cada cambio de estado de un ticket (auditoría).
// Es de solo-escritura: nunca se edita ni se borra.
type TicketStatusLog struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	TicketID        uint      `gorm:"index;not null" json:"ticket_id"`
	PreviousStateID *uint     `gorm:"index" json:"previous_state_id,omitempty"`
	NewStateID      uint      `gorm:"index;not null" json:"new_state_id"`
	ChangedByType   AutorType `gorm:"size:10;not null" json:"changed_by_type"`
	ChangedByAdminID *uuid.UUID `gorm:"type:uuid" json:"changed_by_admin_id,omitempty"`
	ChangedByPsiID  *uuid.UUID `gorm:"type:uuid" json:"changed_by_psi_id,omitempty"`
	Reason          string    `gorm:"size:500" json:"reason"` // comentario opcional del cambio (ej. motivo de cierre)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Ticket *Ticket      `gorm:"foreignKey:TicketID;references:ID" json:"-"`
	NewState *TicketEstado `gorm:"foreignKey:NewStateID;references:ID" json:"new_state,omitempty"`
	PreviousState *TicketEstado `gorm:"foreignKey:PreviousStateID;references:ID" json:"previous_state,omitempty"`
}

func (TicketStatusLog) TableName() string { return "ticket_status_logs" }

// TicketMensaje es un comentario de la conversación interna del ticket.
// Reglas de negocio:
//   - El psicólogo no puede publicar más de 3 mensajes seguidos en la conversación.
//   - El mensaje del psicólogo no puede superar los 1000 caracteres.
//   - La conversación solo admite mensajes mientras el ticket no esté cerrado.
//   - Los mensajes son inmutables (nunca se editan ni borran).
type TicketMensaje struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	TicketID     uint      `gorm:"index;not null" json:"ticket_id"`
	AuthorType   AutorType `gorm:"size:10;not null" json:"author_type"`
	AuthorAdminID *uuid.UUID `gorm:"type:uuid;index" json:"author_admin_id,omitempty"`
	AuthorPsiID  *uuid.UUID `gorm:"type:uuid;index" json:"author_psi_id,omitempty"`
	Message      string    `gorm:"type:text;not null" json:"message"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Ticket  *Ticket         `gorm:"foreignKey:TicketID;references:ID" json:"-"`
	Admin   *UserAdmin      `gorm:"foreignKey:AuthorAdminID;references:ID" json:"-"`
	Psi     *PsiUserModel   `gorm:"foreignKey:AuthorPsiID;references:ID" json:"-"`
	Adjuntos []TicketAdjunto `gorm:"foreignKey:MensajeID" json:"adjuntos,omitempty"`

	// Campos calculados (no persistidos).
	AuthorName string `gorm:"-" json:"author_name,omitempty"`
}

func (TicketMensaje) TableName() string { return "ticket_mensajes" }

// TicketAdjunto es un archivo adjunto de un mensaje de la conversación.
// Se almacena en S3 (carpeta "tickets"); en BD se guarda la key, nunca la URL.
type TicketAdjunto struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	MensajeID      uint   `gorm:"index;not null" json:"mensaje_id"`
	S3Key          string `gorm:"size:512;not null" json:"-"`
	OriginalName   string `gorm:"size:255" json:"original_name"`
	MimeType       string `gorm:"size:100" json:"mime_type"`
	SizeBytes      int64  `json:"size_bytes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Mensaje *TicketMensaje `gorm:"foreignKey:MensajeID;references:ID" json:"-"`

	// URL pública resuelta (no persistida).
	URL string `gorm:"-" json:"url,omitempty"`
}

func (TicketAdjunto) TableName() string { return "ticket_adjuntos" }
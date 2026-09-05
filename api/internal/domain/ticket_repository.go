// api/internal/domain/ticket_repository.go
package domain

import (
	"context"

	"github.com/google/uuid"
)

// TicketFilter define los criterios de búsqueda del panel administrativo.
// Por defecto (SoloAbiertos=true) se listan solo los tickets no cerrados,
// ordenados por orden de llegada (FIFO: created_at ASC).
type TicketFilter struct {
	SoloAbiertos bool       // si true, excluye tickets con estado es_cerrado=true
	MotivoID     *uint      // filtro por motivo
	EstadoID     *uint      // filtro por estado
	PsiUserID    *uuid.UUID // filtro por psicólogo
	Search       string     // búsqueda de texto en título/descripción
	Page         int
	Limit        int
}

// TicketConfigRepository define el contrato de persistencia de la configuración
// del módulo: motivos y estados (los define el colegio). Cada motivo tiene su
// propio límite de tickets abiertos por psicólogo (tickets_per_psi).
type TicketConfigRepository interface {
	// ── Motivos ────────────────────────────────────────────────────────────
	ListMotivos(ctx context.Context) ([]TicketMotivo, error)
	GetMotivo(ctx context.Context, id uint) (*TicketMotivo, error)
	CreateMotivo(ctx context.Context, motivo *TicketMotivo) error
	// CreateMotivoWithDefaults persiste un motivo junto con sus estados por
	// defecto (recibido → en proceso → cerrado) en una sola transacción.
	CreateMotivoWithDefaults(ctx context.Context, motivo *TicketMotivo, estados []TicketEstado) error
	UpdateMotivo(ctx context.Context, motivo *TicketMotivo) error
	DeleteMotivo(ctx context.Context, id uint) error
	CountTicketsByMotivo(ctx context.Context, motivoID uint) (int64, error)

	// ── Estados ────────────────────────────────────────────────────────────
	ListEstados(ctx context.Context, motivoID uint) ([]TicketEstado, error)
	GetEstado(ctx context.Context, id uint) (*TicketEstado, error)
	CreateEstado(ctx context.Context, estado *TicketEstado) error
	UpdateEstado(ctx context.Context, estado *TicketEstado) error
	DeleteEstado(ctx context.Context, id uint) error
	IsEstadoInUse(ctx context.Context, estadoID uint) (bool, error)
}

// TicketRepository define el contrato de persistencia de tickets, conversación
// y auditoría de cambios de estado.
type TicketRepository interface {
	// CreateTicket persiste un ticket y su primer log de estado en una sola
	// transacción atómica (evita tickets huérfanos sin log inicial).
	CreateTicket(ctx context.Context, ticket *Ticket, initialLog *TicketStatusLog) error

	// GetByID recupera un ticket con sus relaciones (Motivo, Estado, PsiUser).
	GetByID(ctx context.Context, id uint) (*Ticket, error)

	// ListMyTickets lista los tickets de un psicólogo (ordenados DESC).
	ListMyTickets(ctx context.Context, psiID uuid.UUID, page, limit int) ([]Ticket, int64, error)

	// ListTickets lista tickets para el panel admin. FIFO: created_at ASC.
	ListTickets(ctx context.Context, filter TicketFilter) ([]Ticket, int64, error)

	// CountActiveByPsiAndMotivo cuenta los tickets abiertos (estado no cerrado,
	// no soft-deleted) de un psicólogo en un motivo. Se usa para el límite
	// configurable por el colegio (tickets_per_psi del motivo).
	CountActiveByPsiAndMotivo(ctx context.Context, psiID uuid.UUID, motivoID uint) (int64, error)

	// CountPendientesAdmin cuenta los tickets no cerrados (para el badge admin).
	CountPendientesAdmin(ctx context.Context) (int64, error)

	// UpdateEstado actualiza el estado actual del ticket.
	UpdateEstado(ctx context.Context, ticket *Ticket) error

	// CreateStatusLog persiste un cambio de estado (auditoría inmutable).
	CreateStatusLog(ctx context.Context, log *TicketStatusLog) error

	// ListStatusLogs devuelve el historial de estados de un ticket (ASC).
	ListStatusLogs(ctx context.Context, ticketID uint) ([]TicketStatusLog, error)

	// CreateMensaje persiste un comentario de la conversación.
	CreateMensaje(ctx context.Context, msg *TicketMensaje) error

	// DeleteMensaje elimina físicamente un mensaje (solo para rollback de
	// transacciones fallidas; la conversación es inmutable).
	DeleteMensaje(ctx context.Context, id uint) error

	// CreateAdjuntos persiste los anexos de un mensaje.
	CreateAdjuntos(ctx context.Context, adjuntos []TicketAdjunto) error

	// ListMensajes devuelve la conversación completa de un ticket (ASC).
	ListMensajes(ctx context.Context, ticketID uint) ([]TicketMensaje, error)

	// ListLastMensajes devuelve los últimos n mensajes de un ticket (ASC).
	// Se usa para aplicar la regla de máximo 3 mensajes seguidos por el psi.
	ListLastMensajes(ctx context.Context, ticketID uint, n int) ([]TicketMensaje, error)
}
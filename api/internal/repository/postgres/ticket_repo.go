// api/internal/repository/postgres/ticket_repo.go
package postgres

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// ticketRepo implementa domain.TicketRepository sobre GORM/PostgreSQL.
type ticketRepo struct {
	db *gorm.DB
}

// NewTicketRepository construye el repositorio de tickets.
func NewTicketRepository(db *gorm.DB) domain.TicketRepository {
	return &ticketRepo{db: db}
}

func (r *ticketRepo) CreateTicket(ctx context.Context, ticket *domain.Ticket, initialLog *domain.TicketStatusLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ticket).Error; err != nil {
			return err
		}
		initialLog.TicketID = ticket.ID
		return tx.Create(initialLog).Error
	})
}

func (r *ticketRepo) GetByID(ctx context.Context, id uint) (*domain.Ticket, error) {
	var ticket domain.Ticket
	err := r.db.WithContext(ctx).
		Preload("Psi").
		Preload("Motivo").
		Preload("Estado").
		Preload("StatusLogs.NewState").
		Preload("StatusLogs.PreviousState").
		Preload("Mensajes.Adjuntos").
		Preload("Mensajes.Admin").
		Preload("Mensajes.Psi").
		First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepo) ListMyTickets(ctx context.Context, psiID uuid.UUID, page, limit int) ([]domain.Ticket, int64, error) {
	var tickets []domain.Ticket
	var total int64

	base := r.db.WithContext(ctx).Model(&domain.Ticket{}).Where("psi_user_id = ?", psiID)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := base.Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Preload("Motivo").
		Preload("Estado").
		Find(&tickets).Error; err != nil {
		return nil, 0, err
	}
	return tickets, total, nil
}

func (r *ticketRepo) ListTickets(ctx context.Context, filter domain.TicketFilter) ([]domain.Ticket, int64, error) {
	var tickets []domain.Ticket
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Ticket{})

	// FIFO administrativo: por defecto solo abiertos, salvo que se filtre por un
	// estado concreto (el estado elegido manda, p.ej. ver los cerrados).
	if filter.SoloAbiertos && filter.EstadoID == nil {
		q = q.Joins("JOIN ticket_estados te ON te.id = tickets.estado_id").
			Where("te.is_closed = FALSE")
	}
	if filter.MotivoID != nil {
		q = q.Where("tickets.motivo_id = ?", *filter.MotivoID)
	}
	if filter.EstadoID != nil {
		q = q.Where("tickets.estado_id = ?", *filter.EstadoID)
	}
	if filter.PsiUserID != nil {
		q = q.Where("tickets.psi_user_id = ?", *filter.PsiUserID)
	}
	if s := strings.TrimSpace(filter.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("(tickets.title ILIKE ? OR tickets.description ILIKE ?)", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Keyset pagination (FIFO): pide tickets posteriores al cursor (id autoincremental).
	if filter.Cursor != nil {
		err := q.Where("tickets.id > ?", *filter.Cursor).
			Order("tickets.id ASC").Limit(filter.Limit).
			Preload("Psi").
			Preload("Motivo").
			Preload("Estado").
			Find(&tickets).Error
		if err != nil {
			return nil, 0, err
		}
		return tickets, total, nil
	}

	err := q.Order("tickets.id ASC").
		Offset((filter.Page - 1) * filter.Limit).Limit(filter.Limit).
		Preload("Psi").
		Preload("Motivo").
		Preload("Estado").
		Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}
	return tickets, total, nil
}

func (r *ticketRepo) CountActiveByPsiAndMotivo(ctx context.Context, psiID uuid.UUID, motivoID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Ticket{}).
		Joins("JOIN ticket_estados te ON te.id = tickets.estado_id").
		Where("tickets.psi_user_id = ? AND tickets.motivo_id = ? AND te.is_closed = FALSE", psiID, motivoID).
		Count(&count).Error
	return count, err
}

func (r *ticketRepo) CountPendientesAdmin(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Ticket{}).
		Joins("JOIN ticket_estados te ON te.id = tickets.estado_id").
		Where("te.is_closed = FALSE").
		Count(&count).Error
	return count, err
}

func (r *ticketRepo) UpdateEstado(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Model(ticket).
		Select("estado_id", "close_reason", "closed_by_type", "closed_by_admin_id", "closed_by_psi_id", "closed_at", "updated_at").
		Updates(map[string]interface{}{
			"estado_id":          ticket.EstadoID,
			"close_reason":       ticket.CloseReason,
			"closed_by_type":     ticket.ClosedByType,
			"closed_by_admin_id": ticket.ClosedByAdminID,
			"closed_by_psi_id":   ticket.ClosedByPsiID,
			"closed_at":          ticket.ClosedAt,
			"updated_at":         ticket.UpdatedAt,
		}).Error
}

func (r *ticketRepo) CreateStatusLog(ctx context.Context, log *domain.TicketStatusLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *ticketRepo) ListStatusLogs(ctx context.Context, ticketID uint) ([]domain.TicketStatusLog, error) {
	var logs []domain.TicketStatusLog
	err := r.db.WithContext(ctx).
		Preload("NewState").
		Preload("PreviousState").
		Where("ticket_id = ?", ticketID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *ticketRepo) CreateMensaje(ctx context.Context, msg *domain.TicketMensaje) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *ticketRepo) DeleteMensaje(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.TicketMensaje{}, id).Error
}

func (r *ticketRepo) CreateAdjuntos(ctx context.Context, adjuntos []domain.TicketAdjunto) error {
	if len(adjuntos) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&adjuntos).Error
}

func (r *ticketRepo) ListMensajes(ctx context.Context, ticketID uint) ([]domain.TicketMensaje, error) {
	var msgs []domain.TicketMensaje
	err := r.db.WithContext(ctx).
		Preload("Adjuntos").
		Preload("Admin").
		Preload("Psi").
		Where("ticket_id = ?", ticketID).
		Order("created_at ASC").
		Find(&msgs).Error
	return msgs, err
}

func (r *ticketRepo) ListLastMensajes(ctx context.Context, ticketID uint, n int) ([]domain.TicketMensaje, error) {
	var msgs []domain.TicketMensaje
	err := r.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("created_at DESC").
		Limit(n).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	// Revertir para devolverlos en orden cronológico (ASC).
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
// api/internal/repository/postgres/ticket_config_repo.go
package postgres

import (
	"context"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// ticketConfigRepo implementa domain.TicketConfigRepository sobre GORM/PostgreSQL.
type ticketConfigRepo struct {
	db *gorm.DB
}

// NewTicketConfigRepository construye el repositorio de configuración de tickets.
func NewTicketConfigRepository(db *gorm.DB) domain.TicketConfigRepository {
	return &ticketConfigRepo{db: db}
}

// ── Motivos ──────────────────────────────────────────────────────────────

func (r *ticketConfigRepo) ListMotivos(ctx context.Context) ([]domain.TicketMotivo, error) {
	var motivos []domain.TicketMotivo
	err := r.db.WithContext(ctx).
		Preload("Estados", func(db *gorm.DB) *gorm.DB {
			return db.Order("ticket_estados.order ASC")
		}).
		Order("created_at ASC").
		Find(&motivos).Error
	return motivos, err
}

func (r *ticketConfigRepo) GetMotivo(ctx context.Context, id uint) (*domain.TicketMotivo, error) {
	var motivo domain.TicketMotivo
	err := r.db.WithContext(ctx).Preload("Estados").First(&motivo, id).Error
	if err != nil {
		return nil, err
	}
	return &motivo, nil
}

func (r *ticketConfigRepo) CreateMotivo(ctx context.Context, motivo *domain.TicketMotivo) error {
	return r.db.WithContext(ctx).Create(motivo).Error
}

func (r *ticketConfigRepo) CreateMotivoWithDefaults(ctx context.Context, motivo *domain.TicketMotivo, estados []domain.TicketEstado) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(motivo).Error; err != nil {
			return err
		}
		for i := range estados {
			estados[i].MotivoID = motivo.ID
			if err := tx.Create(&estados[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ticketConfigRepo) UpdateMotivo(ctx context.Context, motivo *domain.TicketMotivo) error {
	return r.db.WithContext(ctx).Model(motivo).
		Select("name", "description", "tickets_per_psi", "updated_at").
		Updates(map[string]interface{}{
			"name":            motivo.Name,
			"description":     motivo.Description,
			"tickets_per_psi": motivo.TicketsPerPsi,
			"updated_at":      motivo.UpdatedAt,
		}).Error
}

func (r *ticketConfigRepo) DeleteMotivo(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.TicketMotivo{}, id).Error
}

func (r *ticketConfigRepo) CountTicketsByMotivo(ctx context.Context, motivoID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Ticket{}).
		Where("motivo_id = ?", motivoID).
		Count(&count).Error
	return count, err
}

// ── Estados ──────────────────────────────────────────────────────────────

func (r *ticketConfigRepo) ListEstados(ctx context.Context, motivoID uint) ([]domain.TicketEstado, error) {
	var estados []domain.TicketEstado
	err := r.db.WithContext(ctx).
		Where("motivo_id = ?", motivoID).
		Order("ticket_estados.order ASC").
		Find(&estados).Error
	return estados, err
}

func (r *ticketConfigRepo) GetEstado(ctx context.Context, id uint) (*domain.TicketEstado, error) {
	var estado domain.TicketEstado
	err := r.db.WithContext(ctx).First(&estado, id).Error
	if err != nil {
		return nil, err
	}
	return &estado, nil
}

func (r *ticketConfigRepo) CreateEstado(ctx context.Context, estado *domain.TicketEstado) error {
	return r.db.WithContext(ctx).Create(estado).Error
}

func (r *ticketConfigRepo) UpdateEstado(ctx context.Context, estado *domain.TicketEstado) error {
	return r.db.WithContext(ctx).Model(estado).
		Select("name", "order", "is_closed", "updated_at").
		Updates(map[string]interface{}{
			"name":      estado.Name,
			"order":     estado.Order,
			"is_closed": estado.IsClosed,
			"updated_at": estado.UpdatedAt,
		}).Error
}

func (r *ticketConfigRepo) DeleteEstado(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.TicketEstado{}, id).Error
}

func (r *ticketConfigRepo) IsEstadoInUse(ctx context.Context, estadoID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Ticket{}).
		Where("estado_id = ?", estadoID).
		Count(&count).Error
	return count > 0, err
}
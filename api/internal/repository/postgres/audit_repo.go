// api/internal/repository/postgres/audit_repo.go
package postgres

import (
	"context"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// auditRepo implementa domain.AuditRepository sobre PostgreSQL.
type auditRepo struct {
	db *gorm.DB
}

// NewAuditRepository crea el repositorio de la bitácora de cambios.
func NewAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &auditRepo{db: db}
}

// CreateBatch inserta un lote de eventos de auditoría. Se invoca SOLO desde el
// worker diferido (nunca en el hot path de una petición).
func (r *auditRepo) CreateBatch(ctx context.Context, logs []domain.ApiChangeLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&logs).Error
}

// List busca sucesos con filtros combinables y paginación consistente
// (COUNT con los mismos WHERE, luego SELECT ordenado por fecha DESC).
func (r *auditRepo) List(ctx context.Context, f domain.AuditLogFilters) ([]domain.ApiChangeLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.ApiChangeLog{})
	if f.Entity != "" {
		q = q.Where("entity = ?", f.Entity)
	}
	if f.EntityID != "" {
		q = q.Where("entity_id = ?", f.EntityID)
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.ActorID != nil {
		q = q.Where("actor_id = ?", *f.ActorID)
	}
	if f.Q != "" {
		q = q.Where("entity_label ILIKE ?", "%"+f.Q+"%")
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	limit := f.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var logs []domain.ApiChangeLog
	err := q.
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// Stats agrupa los sucesos desde `since` por (entity, action), los más
// frecuentes primero. Sirve para el dashboard y la tabla de estadísticas.
func (r *auditRepo) Stats(ctx context.Context, since time.Time) ([]domain.AuditStat, error) {
	var stats []domain.AuditStat
	err := r.db.WithContext(ctx).
		Model(&domain.ApiChangeLog{}).
		Select("entity, action, COUNT(*) AS count").
		Where("created_at >= ?", since).
		Group("entity, action").
		Order("count DESC").
		Scan(&stats).Error
	return stats, err
}

// PurgeOlderThan borra lógicamente los sucesos anteriores a `before` y
// devuelve cuántos se eliminaron (política de retención).
func (r *auditRepo) PurgeOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&domain.ApiChangeLog{})
	return res.RowsAffected, res.Error
}
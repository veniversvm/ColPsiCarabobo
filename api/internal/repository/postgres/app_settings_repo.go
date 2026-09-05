// api/internal/repository/postgres/app_settings_repo.go
package postgres

import (
	"context"
	"errors"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// appSettingsRepo implementa domain.AppSettingsRepository sobre PostgreSQL.
type appSettingsRepo struct {
	db *gorm.DB
}

// NewAppSettingsRepository crea el repositorio de configuración global.
func NewAppSettingsRepository(db *gorm.DB) domain.AppSettingsRepository {
	return &appSettingsRepo{db: db}
}

// Get lee una clave del KV; retorna nil si no existe.
func (r *appSettingsRepo) Get(ctx context.Context, key string) (*domain.AppSetting, error) {
	var setting domain.AppSetting
	err := r.db.WithContext(ctx).First(&setting, "key = ?", key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

// Upsert inserta o actualiza la clave de configuración de forma idempotente.
func (r *appSettingsRepo) Upsert(ctx context.Context, setting *domain.AppSetting) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(setting).Error
}

// CreateAudit inserta una entrada de auditoría de cambios de configuración.
func (r *appSettingsRepo) CreateAudit(ctx context.Context, log *domain.SettingsAuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

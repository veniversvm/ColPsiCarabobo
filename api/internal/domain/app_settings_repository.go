// api/internal/domain/app_settings_repository.go
package domain

import "context"

// AppSettingsRepository abstrae el acceso al KV global de configuración
// (tabla app_settings) y a su auditoría (settings_audit_logs).
type AppSettingsRepository interface {
	// Get lee una clave; retorna (nil, nil) si no existe (el llamador aplica su default).
	Get(ctx context.Context, key string) (*AppSetting, error)

	// Upsert inserta o actualiza la clave (on conflict update value/updated_at).
	Upsert(ctx context.Context, setting *AppSetting) error

	// CreateAudit persiste una entrada de auditoría de cambios de configuración.
	CreateAudit(ctx context.Context, log *SettingsAuditLog) error
}

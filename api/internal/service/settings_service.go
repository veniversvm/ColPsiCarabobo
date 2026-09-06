// api/internal/service/settings_service.go

// Interruptor global de recepción (tickets e inscripciones): vive en el KV
// app_settings, se controla solo por el Sudo y se audita en settings_audit_logs.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// ErrInvalidSettingKey se retorna al intentar actualizar una clave desconocida.
var ErrInvalidSettingKey = errors.New("clave de configuración inválida")

// ReceptionDisabledError indica que la recepción está desactivada; el handler
// lo traduce a HTTP 409 con code "reception_disabled".
type ReceptionDisabledError struct {
	Reason string
}

func (e *ReceptionDisabledError) Error() string {
	if strings.TrimSpace(e.Reason) != "" {
		return e.Reason
	}
	return "la recepción se encuentra temporalmente desactivada. Intenta más tarde"
}

// GetReceptionSetting lee un interruptor de recepción aplicando el default
// (habilitado) si la clave aún no existe en el KV.
func GetReceptionSetting(ctx context.Context, repo domain.AppSettingsRepository, key string) (domain.ReceptionSetting, error) {
	def := domain.ReceptionSetting{Enabled: true}
	if repo == nil {
		return def, nil
	}
	setting, err := repo.Get(ctx, key)
	if err != nil || setting == nil {
		return def, err
	}
	var v domain.ReceptionSetting
	if err := json.Unmarshal(setting.Value, &v); err != nil {
		return def, err
	}
	return v, nil
}

// AssertReceptionEnabled retorna nil si la recepción está habilitada; en caso
// contrario, un ReceptionDisabledError con el mensaje público configurado.
func AssertReceptionEnabled(ctx context.Context, repo domain.AppSettingsRepository, key string) error {
	v, err := GetReceptionSetting(ctx, repo, key)
	if err != nil {
		return err
	}
	if !v.Enabled {
		return &ReceptionDisabledError{Reason: v.Message}
	}
	return nil
}

// SettingsService agrupa la lectura/escritura de la configuración global,
// usada por los endpoints de administración (solo Sudo para cambios).
type SettingsService struct {
	repo domain.AppSettingsRepository
}

// NewSettingsService crea un SettingsService.
func NewSettingsService(repo domain.AppSettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// ReceptionSwitchesSnapshot devuelve el estado de ambos interruptores de recepción.
type ReceptionSwitchesSnapshot struct {
	Tickets      domain.ReceptionSetting `json:"tickets"`
	Inscriptions domain.ReceptionSetting `json:"inscriptions"`
}

// GetReceptionSwitches devuelve el estado de los interruptores de recepción.
func (s *SettingsService) GetReceptionSwitches(ctx context.Context) (ReceptionSwitchesSnapshot, error) {
	tickets, err := GetReceptionSetting(ctx, s.repo, domain.SettingsKeyTicketsReception)
	if err != nil {
		return ReceptionSwitchesSnapshot{}, err
	}
	inscriptions, err := GetReceptionSetting(ctx, s.repo, domain.SettingsKeyInscriptionsReception)
	if err != nil {
		return ReceptionSwitchesSnapshot{}, err
	}
	return ReceptionSwitchesSnapshot{
		Tickets:      tickets,
		Inscriptions: inscriptions,
	}, nil
}

// UpdateReception actualiza un interruptor de recepción y audita el cambio.
func (s *SettingsService) UpdateReception(ctx context.Context, changedBy domain.UserAdmin, key string, enabled bool, message string) error {
	if key != domain.SettingsKeyTicketsReception && key != domain.SettingsKeyInscriptionsReception {
		return ErrInvalidSettingKey
	}
	message = strings.TrimSpace(message)
	if len([]rune(message)) > 500 {
		message = string([]rune(message)[:500])
	}

	before, err := GetReceptionSetting(ctx, s.repo, key)
	if err != nil {
		return err
	}

	value, err := json.Marshal(domain.ReceptionSetting{Enabled: enabled, Message: message})
	if err != nil {
		return err
	}

	now := time.Now()
	if err := s.repo.Upsert(ctx, &domain.AppSetting{
		Key:       key,
		Value:     value,
		UpdatedAt: now,
	}); err != nil {
		return err
	}

	if err := s.repo.CreateAudit(ctx, &domain.SettingsAuditLog{
		ID:                uuid.Must(uuid.NewV7()),
		ChangedByID:       changedBy.ID,
		ChangedByUsername: changedBy.Username,
		Key:               key,
		EnabledFrom:       before.Enabled,
		EnabledTo:         enabled,
		MessageFrom:       before.Message,
		MessageTo:         message,
		CreatedAt:         now,
	}); err != nil {
		return err
	}

	return nil
}

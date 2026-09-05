// api/internal/service/settings_service_test.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"gorm.io/datatypes"
)

// mockAppSettingsRepo es un mock del KV global con patrón "Func Override".
type mockAppSettingsRepo struct {
	domain.AppSettingsRepository
	GetFunc         func(ctx context.Context, key string) (*domain.AppSetting, error)
	UpsertFunc      func(ctx context.Context, setting *domain.AppSetting) error
	CreateAuditFunc func(ctx context.Context, log *domain.SettingsAuditLog) error
}

func (m *mockAppSettingsRepo) Get(ctx context.Context, key string) (*domain.AppSetting, error) {
	if m.GetFunc == nil {
		return nil, nil
	}
	return m.GetFunc(ctx, key)
}

func (m *mockAppSettingsRepo) Upsert(ctx context.Context, setting *domain.AppSetting) error {
	if m.UpsertFunc == nil {
		return nil
	}
	return m.UpsertFunc(ctx, setting)
}

func (m *mockAppSettingsRepo) CreateAudit(ctx context.Context, log *domain.SettingsAuditLog) error {
	if m.CreateAuditFunc == nil {
		return nil
	}
	return m.CreateAuditFunc(ctx, log)
}

func TestGetReceptionSetting_DefaultEnabled(t *testing.T) {
	// Clave ausente → default habilitado.
	repo := &mockAppSettingsRepo{}
	setting, err := GetReceptionSetting(context.Background(), repo, domain.SettingsKeyTicketsReception)
	require.NoError(t, err)
	require.True(t, setting.Enabled)
	require.Empty(t, setting.Message)
}

func TestGetReceptionSetting_ConNils(t *testing.T) {
	// Repo nil (tests de otros módulos) → default habilitado, sin pánico.
	setting, err := GetReceptionSetting(context.Background(), nil, domain.SettingsKeyTicketsReception)
	require.NoError(t, err)
	require.True(t, setting.Enabled)
}

func TestAssertReceptionEnabled_DisabledConMensaje(t *testing.T) {
	value, _ := json.Marshal(domain.ReceptionSetting{Enabled: false, Message: "Abrimos el 20 de septiembre"})
	repo := &mockAppSettingsRepo{
		GetFunc: func(ctx context.Context, key string) (*domain.AppSetting, error) {
			return &domain.AppSetting{Key: key, Value: datatypes.JSON(value)}, nil
		},
	}

	err := AssertReceptionEnabled(context.Background(), repo, domain.SettingsKeyTicketsReception)
	require.Error(t, err)
	var rdErr *ReceptionDisabledError
	require.True(t, errors.As(err, &rdErr))
	require.Equal(t, "Abrimos el 20 de septiembre", rdErr.Error())
}

func TestSettingsService_UpdateReception_ValidaYAudita(t *testing.T) {
	var upserted *domain.AppSetting
	var audited *domain.SettingsAuditLog

	repo := &mockAppSettingsRepo{
		GetFunc: func(ctx context.Context, key string) (*domain.AppSetting, error) {
			return nil, nil // no existe todavía → before = enabled=true
		},
		UpsertFunc: func(ctx context.Context, setting *domain.AppSetting) error {
			upserted = setting
			return nil
		},
		CreateAuditFunc: func(ctx context.Context, log *domain.SettingsAuditLog) error {
			audited = log
			return nil
		},
	}

	svc := NewSettingsService(repo)
	changedBy := domain.UserAdmin{ID: uuid.New(), Credentials: domain.Credentials{Username: "sudo_admin"}}

	err := svc.UpdateReception(context.Background(), changedBy, domain.SettingsKeyTicketsReception, false, "Cerrado por mantenimiento")
	require.NoError(t, err)

	require.NotNil(t, upserted)
	require.Equal(t, domain.SettingsKeyTicketsReception, upserted.Key)

	var stored domain.ReceptionSetting
	require.NoError(t, json.Unmarshal(upserted.Value, &stored))
	require.False(t, stored.Enabled)
	require.Equal(t, "Cerrado por mantenimiento", stored.Message)

	require.NotNil(t, audited)
	require.Equal(t, "sudo_admin", audited.ChangedByUsername)
	require.Equal(t, domain.SettingsKeyTicketsReception, audited.Key)
	require.True(t, audited.EnabledFrom)
	require.False(t, audited.EnabledTo)
	require.Empty(t, audited.MessageFrom)
	require.Equal(t, "Cerrado por mantenimiento", audited.MessageTo)
}

func TestSettingsService_UpdateRejection_ClaveInvalida(t *testing.T) {
	svc := NewSettingsService(&mockAppSettingsRepo{})
	err := svc.UpdateReception(context.Background(), domain.UserAdmin{}, "fake.key", true, "")
	require.ErrorIs(t, err, ErrInvalidSettingKey)
}

func TestInscriptionService_SubmitBloqueado(t *testing.T) {
	value, _ := json.Marshal(domain.ReceptionSetting{Enabled: false, Message: "Inscripciones cerradas"})
	settingsRepo := &mockAppSettingsRepo{
		GetFunc: func(ctx context.Context, key string) (*domain.AppSetting, error) {
			return &domain.AppSetting{Key: key, Value: datatypes.JSON(value)}, nil
		},
	}

	// Con recepción desactivada, Submit retorna ReceptionDisabledError
	// ANTES de consultar unicidades.
	calledCI := false
	inscRepo := &mockInscriptionRepo{
		CIInPsiUsersFunc: func(ctx context.Context, ci int) (bool, error) {
			calledCI = true
			return false, nil
		},
	}
	svc := NewInscriptionService(inscRepo, nil, settingsRepo, nil, &mockMailService{})
	_, err := svc.Submit(context.Background(), &SubmitInscriptionRequest{Cedula: 123})
	require.False(t, calledCI, "no debe consultar unicidades si la recepción está desactivada")
	require.Error(t, err)
	var rdErr *ReceptionDisabledError
	require.True(t, errors.As(err, &rdErr))
	require.Equal(t, "Inscripciones cerradas", rdErr.Error())
}

func TestTicketService_CreateTicketBloqueado(t *testing.T) {
	value, _ := json.Marshal(domain.ReceptionSetting{Enabled: false, Message: "No estamos abriendo tickets"})
	settingsRepo := &mockAppSettingsRepo{
		GetFunc: func(ctx context.Context, key string) (*domain.AppSetting, error) {
			return &domain.AppSetting{Key: key, Value: datatypes.JSON(value)}, nil
		},
	}

	// configRepo nil: si la recepción está desactivada, no debe alcanzarse.
	svc := NewTicketService(nil, nil, settingsRepo, nil, nil)
	_, err := svc.CreateTicket(context.Background(), &domain.PsiUserModel{},
		request_structs.CreateTicketRequest{Title: "t", Description: "d", MotivoID: 1}, nil)
	require.Error(t, err)
	var rdErr *ReceptionDisabledError
	require.True(t, errors.As(err, &rdErr))
	require.Equal(t, "No estamos abriendo tickets", rdErr.Error())
}

// api/internal/domain/notification_repository.go
// Package domain define las entidades de negocio y los contratos de persistencia del sistema.
package domain

import (
	"context"

	"github.com/google/uuid"
)

// NotificationFilterParams parámetros para resolver destinatarios en el filter engine.
// Todos los campos son opcionales y acumulables; solo se aplican si tienen valor.
type NotificationFilterParams struct {
	Municipality string
	State        string
	Genre        string
	SpecialtyID  *uint32
	Solvent      *bool
}

// PsiUserInfo información mínima de un psicólogo para preview de destinatarios.
type PsiUserInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

// NotificationRepository define el contrato de persistencia para el módulo de notificaciones.
type NotificationRepository interface {
	// =========================================================================
	// ESCRITURA
	// =========================================================================

	// Create persiste una notificación.
	Create(ctx context.Context, n *Notification) error

	// CreateTargets persiste de forma masiva los destinatarios de una notificación.
	CreateTargets(ctx context.Context, targets []NotificationTarget) error

	// CreateFilters persiste los filtros de auditoría de una notificación.
	CreateFilters(ctx context.Context, filters []NotificationFilter) error

	// CreateAttach persiste un adjunto de la notificación.
	CreateAttach(ctx context.Context, a *NotificationAttach) error

	// DeleteTargets elimina los destinatarios de una notificación (revertir envío).
	DeleteTargets(ctx context.Context, notificationID uuid.UUID) error

	// =========================================================================
	// LECTURA
	// =========================================================================

	// GetByID recupera una notificación completa con sus relaciones.
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)

	// ListBySender lista las notificaciones creadas por un administrador (paginado DESC).
	ListBySender(ctx context.Context, senderID uuid.UUID, page, limit int) ([]Notification, int64, error)

	// ListByUser lista las notificaciones del agremiado (paginado DESC).
	ListByUser(ctx context.Context, psiUserID uuid.UUID, page, limit int) ([]Notification, int64, error)

	// GetTargets recupera los destinatarios de una notificación con su psicólogo.
	GetTargets(ctx context.Context, notificationID uuid.UUID) ([]NotificationTarget, error)

	// GetAttachs recupera los adjuntos de una notificación.
	GetAttachs(ctx context.Context, notificationID uuid.UUID) ([]NotificationAttach, error)

	// GetTargetByUserAndNotification recupera un target específico de un usuario.
	GetTargetByUserAndNotification(ctx context.Context, notificationID, psiUserID uuid.UUID) (*NotificationTarget, error)

	// CountUnread cuenta las notificaciones no leídas de un agremiado.
	CountUnread(ctx context.Context, psiUserID uuid.UUID) (int64, error)

	// =========================================================================
	// MUTACIONES
	// =========================================================================

	// MarkAsRead marca como leída una notificación para el agremiado.
	MarkAsRead(ctx context.Context, notificationID, psiUserID uuid.UUID) error

	// Cancel actualiza el estado a "cancelled".
	Cancel(ctx context.Context, id uuid.UUID) error

	// UpdateStatus actualiza el estado y sent_at de una notificación.
	UpdateStatus(ctx context.Context, id uuid.UUID, status NotificationStatus) error

	// RecordEmailSent marca el email como enviado para un target.
	RecordEmailSent(ctx context.Context, targetID uuid.UUID) error

	// =========================================================================
	// FILTER ENGINE (resolución de destinatarios)
	// =========================================================================

	// ResolveRecipients retorna los UUIDs de psicólogos que cumplen los filtros.
	// Base SIEMPRE activa: is_active = true AND proof_of_life = true.
	ResolveRecipients(ctx context.Context, params NotificationFilterParams) ([]uuid.UUID, error)

	// ResolveAll retorna todos los UUIDs activos y con fe de vida.
	ResolveAll(ctx context.Context) ([]uuid.UUID, error)

	// ResolveByIDs valida que los UUIDs proporcionados existan y estén activos.
	ResolveByIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)

	// GetPsiUserInfo retorna nombre y email de los psicólogos (para preview).
	GetPsiUserInfo(ctx context.Context, ids []uuid.UUID) ([]PsiUserInfo, error)

	// ListPendingScheduled retorna las notificaciones programadas pendientes cuyo scheduled_at ya pasó.
	ListPendingScheduled(ctx context.Context) ([]Notification, error)
}

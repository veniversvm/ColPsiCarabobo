// api/internal/repository/postgres/notification_repo.go

// Package postgres provee la implementación concreta de los repositorios usando PostgreSQL y GORM.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// notificationRepo implementa la interfaz domain.NotificationRepository.
type notificationRepo struct {
	db *gorm.DB
}

// NewNotificationRepository construye e inicializa un repositorio de notificaciones.
func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepo{db: db}
}

// =========================================================================
// ESCRITURA
// =========================================================================

func (r *notificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepo) CreateTargets(ctx context.Context, targets []domain.NotificationTarget) error {
	if len(targets) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&targets).Error
}

func (r *notificationRepo) CreateFilters(ctx context.Context, filters []domain.NotificationFilter) error {
	if len(filters) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&filters).Error
}

func (r *notificationRepo) CreateAttach(ctx context.Context, a *domain.NotificationAttach) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *notificationRepo) DeleteTargets(ctx context.Context, notificationID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		Delete(&domain.NotificationTarget{}).Error
}

// =========================================================================
// LECTURA
// =========================================================================

func (r *notificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	var n domain.Notification
	err := r.db.WithContext(ctx).
		Preload("Targets").
		Preload("Filters").
		Preload("Attachs").
		First(&n, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *notificationRepo) ListBySender(ctx context.Context, senderID uuid.UUID, page, limit int) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("sender_id = ?", senderID)

	query.Count(&total)
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error

	return notifications, total, err
}

func (r *notificationRepo) ListByUser(ctx context.Context, psiUserID uuid.UUID, page, limit int) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var total int64

	// Subconsulta: IDs de notificaciones donde el usuario es target.
	sub := r.db.WithContext(ctx).Model(&domain.NotificationTarget{}).
		Select("notification_id").
		Where("psi_user_id = ?", psiUserID)

	query := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("id IN (?) AND status = ?", sub, domain.NotificationStatusSent)

	query.Count(&total)
	offset := (page - 1) * limit

	// Preload del target del agremiado actual para exponer is_read/read_at en
	// el listado (la UI usa targets[0] para distinguir leída de no leída).
	err := query.Preload("Targets", "psi_user_id = ?", psiUserID).
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error

	return notifications, total, err
}

func (r *notificationRepo) GetTargets(ctx context.Context, notificationID uuid.UUID) ([]domain.NotificationTarget, error) {
	var targets []domain.NotificationTarget
	err := r.db.WithContext(ctx).
		Preload("PsiUser").
		Where("notification_id = ?", notificationID).
		Find(&targets).Error
	return targets, err
}

func (r *notificationRepo) GetAttachs(ctx context.Context, notificationID uuid.UUID) ([]domain.NotificationAttach, error) {
	var attachs []domain.NotificationAttach
	err := r.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		Find(&attachs).Error
	return attachs, err
}

func (r *notificationRepo) GetTargetByUserAndNotification(ctx context.Context, notificationID, psiUserID uuid.UUID) (*domain.NotificationTarget, error) {
	var target domain.NotificationTarget
	err := r.db.WithContext(ctx).
		Where("notification_id = ? AND psi_user_id = ?", notificationID, psiUserID).
		First(&target).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *notificationRepo) CountUnread(ctx context.Context, psiUserID uuid.UUID) (int64, error) {
	var count int64
	// Unir con notifications para contar solo las enviadas.
	err := r.db.WithContext(ctx).
		Model(&domain.NotificationTarget{}).
		Joins("JOIN notifications ON notifications.id = notification_targets.notification_id").
		Where("notification_targets.psi_user_id = ?", psiUserID).
		Where("notification_targets.is_read = ?", false).
		Where("notifications.status = ?", domain.NotificationStatusSent).
		Count(&count).Error
	return count, err
}

// =========================================================================
// MUTACIONES
// =========================================================================

func (r *notificationRepo) MarkAsRead(ctx context.Context, notificationID, psiUserID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.NotificationTarget{}).
		Where("notification_id = ? AND psi_user_id = ?", notificationID, psiUserID).
		Updates(map[string]interface{}{
			"is_read":  true,
			"read_at":  time.Now(),
		}).Error
}

func (r *notificationRepo) Cancel(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", id).
		Update("status", domain.NotificationStatusCancelled).Error
}

func (r *notificationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  status,
			"sent_at": time.Now(),
		}).Error
}

func (r *notificationRepo) RecordEmailSent(ctx context.Context, targetID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.NotificationTarget{}).
		Where("id = ?", targetID).
		Updates(map[string]interface{}{
			"email_sent":     true,
			"email_sent_at":  time.Now(),
		}).Error
}

// =========================================================================
// FILTER ENGINE (resolución de destinatarios)
// =========================================================================

// baseRecipientsQuery filtra SIEMPRE agremiados activos y con fe de vida.
func (r *notificationRepo) baseRecipientsQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
		Where("is_active = ? AND proof_of_life = ?", true, true)
}

func (r *notificationRepo) ResolveRecipients(ctx context.Context, params domain.NotificationFilterParams) ([]uuid.UUID, error) {
	query := r.baseRecipientsQuery(ctx)

	if params.Municipality != "" {
		query = query.Where("municipality_carabobo = ?", params.Municipality)
	}
	if params.State != "" {
		query = query.Where("state_outside = ?", params.State)
	}
	if params.Genre != "" {
		query = query.Where("genre = ?", params.Genre)
	}
	if params.SpecialtyID != nil {
		query = query.Where(
			"(primary_specialty_id = ? OR secondary_specialty_id = ?)",
			*params.SpecialtyID, *params.SpecialtyID,
		)
	}
	if params.Solvent != nil {
		query = query.Where("solvent = ?", *params.Solvent)
	}

	var ids []uuid.UUID
	err := query.Pluck("id", &ids).Error
	return ids, err
}

func (r *notificationRepo) ResolveAll(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.baseRecipientsQuery(ctx).Pluck("id", &ids).Error
	return ids, err
}

func (r *notificationRepo) ResolveByIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	var resolved []uuid.UUID
	err := r.baseRecipientsQuery(ctx).
		Where("id IN ?", ids).
		Pluck("id", &resolved).Error
	return resolved, err
}

func (r *notificationRepo) GetPsiUserInfo(ctx context.Context, ids []uuid.UUID) ([]domain.PsiUserInfo, error) {
	if len(ids) == 0 {
		return []domain.PsiUserInfo{}, nil
	}
	var infos []domain.PsiUserInfo
	err := r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
		Select("id, CONCAT(first_name, ' ', last_name) AS name, contact_email AS email").
		Where("id IN ?", ids).
		Scan(&infos).Error
	return infos, err
}

// ListPendingScheduled retorna las notificaciones programadas pendientes cuyo scheduled_at ya llegó.
func (r *notificationRepo) ListPendingScheduled(ctx context.Context) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", domain.NotificationStatusPending, time.Now()).
		Find(&notifications).Error
	return notifications, err
}

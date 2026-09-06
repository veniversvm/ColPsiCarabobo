// api/internal/service/notification_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
package service

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// IMailService se reutiliza (definido en mail_service.go) para el envío opcional de correos.

// NotificationService orquesta la lógica de negocio del módulo de notificaciones.
type NotificationService struct {
	repo     domain.NotificationRepository
	s3Client *s3.S3Client
	mailSvc  IMailService
}

// NewNotificationService construye el servicio con sus dependencias inyectadas (DI).
func NewNotificationService(repo domain.NotificationRepository, s3Client *s3.S3Client, mailSvc IMailService) *NotificationService {
	return &NotificationService{repo: repo, s3Client: s3Client, mailSvc: mailSvc}
}

// mailTemplateName es el nombre del template embebido (sin extensión).
const mailTemplateName = "notification"

// =========================================================================
// RESOLUCIÓN DE DESTINATARIOS (filter engine)
// =========================================================================

// resolveRecipients resuelve los UUIDs destinatarios según el target type y filtros.
func (s *NotificationService) resolveRecipients(ctx context.Context, req request_structs.CreateNotificationRequest) ([]uuid.UUID, error) {
	switch domain.NotificationTargetType(req.TargetType) {
	case domain.NotificationTargetGlobal:
		return s.repo.ResolveAll(ctx)
	case domain.NotificationTargetIndividual:
		return s.repo.ResolveByIDs(ctx, req.TargetUserIDs)
	case domain.NotificationTargetGroup:
		if req.Filters == nil {
			return s.repo.ResolveAll(ctx)
		}
		params := domain.NotificationFilterParams{
			Municipality: req.Filters.Municipality,
			State:        req.Filters.State,
			Genre:        req.Filters.Genre,
			SpecialtyID:  req.Filters.SpecialtyID,
			Solvent:      req.Filters.Solvent,
		}
		return s.repo.ResolveRecipients(ctx, params)
	default:
		return nil, domain.ErrNotificationInvalidTargetType
	}
}

// PreviewRecipients retorna los destinatarios potenciales sin crear registros.
func (s *NotificationService) PreviewRecipients(ctx context.Context, admin *domain.UserAdmin, req request_structs.PreviewNotificationRequest) (map[string]interface{}, error) {
	if !admin.CanSendNotifications && !admin.Sudo {
		return nil, domain.ErrNotificationPermDenied
	}

	created := request_structs.CreateNotificationRequest{
		TargetType:   req.TargetType,
		Filters:      req.Filters,
		TargetUserIDs: req.TargetUserIDs,
	}

	ids, err := s.resolveRecipients(ctx, created)
	if err != nil {
		return nil, err
	}

	infos, err := s.repo.GetPsiUserInfo(ctx, ids)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_recipients": len(infos),
		"recipients":       infos,
	}, nil
}

// =========================================================================
// CREACIÓN Y ENVÍO
// =========================================================================

// toFilterModel serializa los filtros en modelos de auditoría (NotificationFilter).
func (s *NotificationService) toFilterModel(notificationID uuid.UUID, req request_structs.CreateNotificationRequest) []domain.NotificationFilter {
	var targetUserIDs string
	if len(req.TargetUserIDs) > 0 {
		if b, err := json.Marshal(req.TargetUserIDs); err == nil {
			targetUserIDs = string(b)
		}
	}

	filters := []domain.NotificationFilter{
		{
			NotificationID: notificationID,
			TargetUserIDs:  targetUserIDs,
		},
	}
	if req.Filters != nil {
		filters[0].Municipality = req.Filters.Municipality
		filters[0].State = req.Filters.State
		filters[0].Genre = req.Filters.Genre
		filters[0].SpecialtyID = req.Filters.SpecialtyID
		filters[0].Solvent = req.Filters.Solvent
	}
	return filters
}

// CreateNotification crea una notificación (inmediata o programada).
func (s *NotificationService) CreateNotification(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreateNotificationRequest) (map[string]interface{}, error) {
	if !admin.CanSendNotifications && !admin.Sudo {
		return nil, domain.ErrNotificationPermDenied
	}

	scheduled := req.ScheduledAt != nil
	if scheduled && req.ScheduledAt.Before(time.Now()) {
		return nil, domain.ErrNotificationInvalidSchedule
	}

	now := time.Now()
	notification := &domain.Notification{
		Title:       req.Title,
		Message:     req.Message,
		TargetType:  domain.NotificationTargetType(req.TargetType),
		SenderID:    admin.ID,
		SendEmail:   req.SendEmail,
		Status:      domain.NotificationStatusPending,
		AuditModel:  domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
	if scheduled {
		notification.ScheduledAt = req.ScheduledAt
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	// Registrar filtros de auditoría.
	if err := s.repo.CreateFilters(ctx, s.toFilterModel(notification.ID, req)); err != nil {
		return nil, err
	}

	// Si es programada, no se resuelven destinatarios aún.
	if scheduled {
		return map[string]interface{}{
			"id":             notification.ID,
			"status":         notification.Status,
			"scheduled_at":   notification.ScheduledAt,
			"message":        "Notificación programada creada",
		}, nil
	}

	// Envío inmediato.
	stats, err := s.send(ctx, notification, &req)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// notificationToReq reconstruye el request de destinatarios desde una notificación.
// Usado por el worker para las notificaciones programadas: reconstruye los filtros
// y destinatarios individuales a partir de los NotificationFilter persistidos.
func (s *NotificationService) notificationToReq(n *domain.Notification) (request_structs.CreateNotificationRequest, error) {
	req := request_structs.CreateNotificationRequest{
		TargetType: string(n.TargetType),
	}

	// Cargar filtros de auditoría persistidos.
	filters := n.Filters
	if len(filters) == 0 {
		all, err := s.repo.GetByID(context.TODO(), n.ID)
		if err == nil && all != nil {
			filters = all.Filters
		}
	}
	if len(filters) > 0 {
		f := filters[0]
		if f.Municipality != "" || f.State != "" || f.Genre != "" || f.SpecialtyID != nil || f.Solvent != nil {
			req.Filters = &request_structs.NotificationFilterDTO{
				Municipality: f.Municipality,
				State:        f.State,
				Genre:        f.Genre,
				SpecialtyID:  f.SpecialtyID,
				Solvent:      f.Solvent,
			}
		}
		if f.TargetUserIDs != "" {
			_ = json.Unmarshal([]byte(f.TargetUserIDs), &req.TargetUserIDs)
		}
	}

	return req, nil
}

// send materializa una notificación pendiente: resuelve destinatarios, crea targets
// y opcionalmente encola emails. Retorna estadísticas.
func (s *NotificationService) send(ctx context.Context, n *domain.Notification, recipientReq *request_structs.CreateNotificationRequest) (map[string]interface{}, error) {
	var req request_structs.CreateNotificationRequest
	if recipientReq != nil {
		req = *recipientReq
	} else {
		rebuilt, err := s.notificationToReq(n)
		if err != nil {
			return nil, err
		}
		req = rebuilt
	}

	ids, err := s.resolveRecipients(ctx, req)
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, n.ID, domain.NotificationStatusFailed)
		return nil, err
	}

	var targets []domain.NotificationTarget
	for _, id := range ids {
		targets = append(targets, domain.NotificationTarget{
			NotificationID: n.ID,
			PsiUserID:      id,
			IsRead:         false,
		})
	}

	if err := s.repo.CreateTargets(ctx, targets); err != nil {
		return nil, err
	}

	// Envío opcional de emails (fire-and-forget, el worker del MailService gestiona el throttling).
	if n.SendEmail && s.mailSvc != nil {
		infos, err := s.repo.GetPsiUserInfo(ctx, ids)
		if err == nil {
			for _, info := range infos {
				if info.Email == "" {
					continue
				}
				data := map[string]interface{}{
					"Title":   n.Title,
					"Message": n.Message,
				}
				if err := s.mailSvc.SendEmail(info.Email, n.Title, mailTemplateName, data); err != nil {
					log.Warn().Err(err).Str("component", "notification").Str("to", info.Email).Msg("Error encolando correo de notificación")
				}
			}
		}
	}

	if err := s.repo.UpdateStatus(ctx, n.ID, domain.NotificationStatusSent); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":            n.ID,
		"status":        domain.NotificationStatusSent,
		"total_sent":    len(targets),
		"total_emails":  len(targets),
		"message":       "Notificación enviada",
	}, nil
}

// ProcessScheduled envía las notificaciones programadas cuyo scheduled_at ya llegó.
// Es invocado periódicamente por el worker en cmd/api/main.go.
func (s *NotificationService) ProcessScheduled(ctx context.Context) {
	pending, err := s.repo.ListPendingScheduled(ctx)
	if err != nil {
		log.Error().Err(err).Str("component", "notification").Msg("Error consultando notificaciones programadas")
		return
	}
	for _, n := range pending {
		if _, err := s.send(ctx, &n, nil); err != nil {
			_ = s.repo.UpdateStatus(ctx, n.ID, domain.NotificationStatusFailed)
			log.Error().Err(err).Str("component", "notification").Str("id", n.ID.String()).Msg("Error enviando notificación programada")
		}
	}
}

// NotifyPSI crea una notificación inmediata dirigida a un agremiado concreto,
// marcada como "sent" desde el inicio (sin email ni programación). Se usa para
// avisos generados por submódulos internos — ej. el módulo de Tickets: cuando
// el admin cambia el estado de un ticket, responde o lo cierra. No requiere
// permisos de notificaciones ni tablero admin: el poller del portal psi
// (/notifications/psi-user) la recoge y reproduce el sonido de aviso.
func (s *NotificationService) NotifyPSI(ctx context.Context, senderID uuid.UUID, senderName string, psiUserID uuid.UUID, title, message string) error {
	now := time.Now()
	sentAt := now
	notification := &domain.Notification{
		Title:      title,
		Message:    message,
		TargetType: domain.NotificationTargetIndividual,
		SenderID:   senderID,
		SendEmail:  false,
		Status:     domain.NotificationStatusSent,
		SentAt:     &sentAt,
		AuditModel: domain.AuditModel{
			CreateBy:   senderName,
			CreateById: &senderID,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return err
	}
	return s.repo.CreateTargets(ctx, []domain.NotificationTarget{{
		NotificationID: notification.ID,
		PsiUserID:      psiUserID,
		IsRead:         false,
	}})
}

// ListMyNotifications lista las notificaciones creadas por un admin (paginado).
func (s *NotificationService) ListMyNotifications(ctx context.Context, admin *domain.UserAdmin, page, limit int) (map[string]interface{}, error) {
	list, total, err := s.repo.ListBySender(ctx, admin.ID, page, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data":  list,
		"total": total,
		"page":  page,
	}, nil
}

// GetNotificationDetail retorna el detalle + estadísticas de una notificación.
func (s *NotificationService) GetNotificationDetail(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID) (map[string]interface{}, error) {
	if !admin.CanReadNotifications && !admin.Sudo {
		return nil, domain.ErrNotificationPermDenied
	}
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrNotificationNotFound
	}
	read := 0
	for _, t := range n.Targets {
		if t.IsRead {
			read++
		}
	}
	// Resolver URLs públicas de adjuntos.
	for i := range n.Attachs {
		if s.s3Client != nil {
			n.Attachs[i].S3Key = s.s3Client.GetPublicURL(n.Attachs[i].S3Key)
		}
	}
	return map[string]interface{}{
		"notification":   n,
		"total_recipients": len(n.Targets),
		"total_read":       read,
		"total_unread":     len(n.Targets) - read,
	}, nil
}

// GetTargetsAdmin retorna los destinatarios de una notificación con estado de lectura.
func (s *NotificationService) GetTargetsAdmin(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID) ([]domain.NotificationTarget, error) {
	if !admin.CanReadNotifications && !admin.Sudo {
		return nil, domain.ErrNotificationPermDenied
	}
	return s.repo.GetTargets(ctx, id)
}

// CancelNotification cancela una notificación programada pendiente.
func (s *NotificationService) CancelNotification(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID) error {
	if !admin.CanManageNotifications && !admin.Sudo {
		return domain.ErrNotificationPermDenied
	}
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.ErrNotificationNotFound
	}
	// Solo el creador o quien tenga CanManageNotifications puede cancelar.
	if n.SenderID != admin.ID && !admin.CanManageNotifications && !admin.Sudo {
		return domain.ErrNotificationPermDenied
	}
	if n.Status != domain.NotificationStatusPending {
		return domain.ErrNotificationCannotCancel
	}
	if err := s.repo.Cancel(ctx, id); err != nil {
		return err
	}
	// Limpiar targets si existieran.
	return s.repo.DeleteTargets(ctx, id)
}

// AttachFile asocia un archivo subido a S3 con una notificación.
// Retorna la key de S3 generada si el archivo se subió y persistió.
func (s *NotificationService) AttachFile(ctx context.Context, admin *domain.UserAdmin, notificationID uuid.UUID, fileHeader *multipart.FileHeader) (string, error) {
	if !admin.CanSendNotifications && !admin.Sudo {
		return "", domain.ErrNotificationPermDenied
	}
	if _, err := s.repo.GetByID(ctx, notificationID); err != nil {
		return "", domain.ErrNotificationNotFound
	}
	if s.s3Client == nil {
		return "", errors.New("almacenamiento no disponible")
	}

	s3Key, err := s.s3Client.UploadFile(ctx, fileHeader, "notifications")
	if err != nil {
		return "", err
	}

	attach := &domain.NotificationAttach{
		NotificationID: notificationID,
		Name:           fileHeader.Filename,
		S3Key:          s3Key,
		ContentType:    fileHeader.Header.Get("Content-Type"),
		AuditModel: domain.AuditModel{
			CreateById: &admin.ID,
		},
	}
	if err := s.repo.CreateAttach(ctx, attach); err != nil {
		// Rollback S3 para no dejar archivos huérfanos.
		_ = s.s3Client.DeleteFile(ctx, s3Key)
		return "", err
	}
	return s.s3Client.GetPublicURL(s3Key), nil
}

// =========================================================================
// AGREMIDO — CONSULTA
// =========================================================================

// ListUserNotifications lista las notificaciones del agremiado (paginado DESC).
func (s *NotificationService) ListUserNotifications(ctx context.Context, psi *domain.PsiUserModel, page, limit int) (map[string]interface{}, error) {
	list, total, err := s.repo.ListByUser(ctx, psi.ID, page, limit)
	if err != nil {
		return nil, err
	}
	// Resolver adjuntos para cada notificación.
	for i := range list {
		if s.s3Client != nil {
			for j := range list[i].Attachs {
				list[i].Attachs[j].S3Key = s.s3Client.GetPublicURL(list[i].Attachs[j].S3Key)
			}
		}
	}
	return map[string]interface{}{
		"data":  list,
		"total": total,
		"page":  page,
	}, nil
}

// GetUnreadCount retorna el contador de notificaciones no leídas del agremiado.
func (s *NotificationService) GetUnreadCount(ctx context.Context, psi *domain.PsiUserModel) (int64, error) {
	return s.repo.CountUnread(ctx, psi.ID)
}

// GetUserNotificationById recupera una notificación para el agremiado y la marca como leída.
func (s *NotificationService) GetUserNotificationById(ctx context.Context, psi *domain.PsiUserModel, id uuid.UUID) (*domain.Notification, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrNotificationNotFound
	}
	// Verificar que el usuario sea destinatario.
	if _, err := s.repo.GetTargetByUserAndNotification(ctx, id, psi.ID); err != nil {
		return nil, domain.ErrNotificationTargetNotOwned
	}
	if err := s.repo.MarkAsRead(ctx, id, psi.ID); err != nil {
		return nil, err
	}
	// Resolver adjuntos.
	for i := range n.Attachs {
		if s.s3Client != nil {
			n.Attachs[i].S3Key = s.s3Client.GetPublicURL(n.Attachs[i].S3Key)
		}
	}
	return n, nil
}

// MarkAsRead marca como leída una notificación del agremiado (endpoint semántico).
// Se elimina la auto-marca al abrir: el psicólogo debe pulsar el botón "Marcar como leída".
func (s *NotificationService) MarkAsRead(ctx context.Context, psi *domain.PsiUserModel, id uuid.UUID) error {
	if _, err := s.repo.GetTargetByUserAndNotification(ctx, id, psi.ID); err != nil {
		return domain.ErrNotificationTargetNotOwned
	}
	if err := s.repo.MarkAsRead(ctx, id, psi.ID); err != nil {
		return err
	}
	return nil
}

// GetAttachmentURL retorna la URL pública de un adjunto si el agremiado es destinatario.
func (s *NotificationService) GetAttachmentURL(ctx context.Context, psi *domain.PsiUserModel, notificationID, attachID uuid.UUID) (string, error) {
	if _, err := s.repo.GetTargetByUserAndNotification(ctx, notificationID, psi.ID); err != nil {
		return "", domain.ErrNotificationTargetNotOwned
	}
	attachs, err := s.repo.GetAttachs(ctx, notificationID)
	if err != nil {
		return "", err
	}
	for _, a := range attachs {
		if a.ID == attachID {
			if s.s3Client == nil {
				return "", errors.New("almacenamiento no disponible")
			}
			return s.s3Client.GetPublicURL(a.S3Key), nil
		}
	}
	return "", domain.ErrAttachmentNotFound
}

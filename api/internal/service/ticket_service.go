// api/internal/service/ticket_service.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona el módulo de Tickets de Solicitudes:
//   - Creación de tickets (solo psicólogos) con límite configurable por motivo
//     (el colegio define cuántos tickets abiertos puede tener el psi por motivo).
//   - Conversación interna mientras el ticket no esté cerrado.
//   - Regla de "máximo 3 mensajes seguidos" del psicólogo y límite de
//     caracteres (1000 por comentario del psi).
//   - Cierre por parte de admin o psi con motivo de cierre obligatorio.
//   - Estados configurados por motivo (los define el colegio).
//   - FIFO administrativo: el panel listo los tickets abiertos más antiguos.
//   - Notificaciones al psicólogo cuando el admin cambia de estado, responde
//     o cierra su ticket (vía NotificationService.NotifyPSI).
package service

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// Límites de negocio del módulo de tickets.
const (
	// MaxPsiMensajeLengthChars: máximo de caracteres por comentario del psicólogo.
	MaxPsiMensajeLengthChars = 1000
	// MaxAdminMensajeLengthChars: el admin puede redactar comentarios más extensos.
	MaxAdminMensajeLengthChars = 4000
	// MaxConsecutivePsiComments: el psi no puede publicar más de 3 mensajes seguidos.
	MaxConsecutivePsiComments = 3
	// límites de los campos de configuración.
	maxMotivoNameLen   = 120
	maxMotivoDescLen   = 500
	maxEstadoNameLen   = 60
	maxTicketTitleLen  = 200
	maxTicketDescLen   = 2000
	maxCloseReasonLen  = 500
	maxChangeReasonLen = 500
)

// defaultEstados se siembran automáticamente al crear un motivo: el colegio
// puede luego editarlos/ampliarlos.
type defaultEstado struct {
	name     string
	isClosed bool
}

var defaultEstados = []defaultEstado{
	{name: "Recibido", isClosed: false},
	{name: "En proceso", isClosed: false},
	{name: "Cerrado", isClosed: true},
}

// ticketNotifier abstrae la creación de notificaciones hacia el psicólogo.
// Lo implementa (*service.NotificationService).NotifyPSI.
type ticketNotifier interface {
	NotifyPSI(ctx context.Context, senderID uuid.UUID, senderName string, psiUserID uuid.UUID, title, message string) error
}

// TicketService orquesta las reglas de negocio del módulo de tickets.
type TicketService struct {
	repo       domain.TicketRepository
	configRepo domain.TicketConfigRepository
	settings   domain.AppSettingsRepository
	s3Client   *s3.S3Client
	notifier   ticketNotifier
}

// NewTicketService construye el servicio con sus dependencias inyectadas (DI).
func NewTicketService(repo domain.TicketRepository, configRepo domain.TicketConfigRepository, settings domain.AppSettingsRepository, s3Client *s3.S3Client, notifier ticketNotifier) *TicketService {
	return &TicketService{
		repo:       repo,
		configRepo: configRepo,
		settings:   settings,
		s3Client:   s3Client,
		notifier:   notifier,
	}
}

// ReceptionStatus devuelve el estado del interruptor de tickets de solicitudes.
func (s *TicketService) ReceptionStatus(ctx context.Context) (domain.ReceptionSetting, error) {
	return GetReceptionSetting(ctx, s.settings, domain.SettingsKeyTicketsReception)
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPERS COMUNES
// ─────────────────────────────────────────────────────────────────────────────

// sanitizeTicketText limpia y recorta texto libre de usuario por runas.
func sanitizeTicketText(s string, maxRunes int) string {
	s = bluemonday.StrictPolicy().Sanitize(s)
	s = strings.TrimSpace(s)
	if maxRunes <= 0 {
		return s
	}
	if r := utf8.RuneCountInString(s); r > maxRunes {
		for i := range s {
			if utf8.RuneCountInString(s[:i]) >= maxRunes {
				return s[:i]
			}
		}
	}
	return s
}

// canManage valida el permiso administrativo del módulo (Sudo bypassa todo).
func canManageTickets(admin *domain.UserAdmin) bool {
	return admin != nil && (admin.Sudo || admin.CanManageTickets)
}

// findInitialEstado devuelve el primer estado no-cerrado (orden ASC) de un motivo.
func findInitialEstado(estados []domain.TicketEstado) *domain.TicketEstado {
	var best *domain.TicketEstado
	for i := range estados {
		if estados[i].IsClosed {
			continue
		}
		if best == nil || estados[i].Order < best.Order {
			best = &estados[i]
		}
	}
	return best
}

// findClosedEstado devuelve el primer estado cerrado (orden ASC) de un motivo.
func findClosedEstado(estados []domain.TicketEstado) *domain.TicketEstado {
	var best *domain.TicketEstado
	for i := range estados {
		if !estados[i].IsClosed {
			continue
		}
		if best == nil || estados[i].Order < best.Order {
			best = &estados[i]
		}
	}
	return best
}

func (s *TicketService) fillTicket(t *domain.Ticket) {
	if t == nil {
		return
	}
	if t.Psi != nil {
		t.PsiFirstName = t.Psi.FirstName
		t.PsiLastName = t.Psi.LastName
	}
	t.IsClosed = t.Estado != nil && t.Estado.IsClosed
	if t.ClosedAt != nil {
		t.IsClosed = true
	}
	for i := range t.Mensajes {
		s.fillMensaje(&t.Mensajes[i])
	}
}

func (s *TicketService) fillMensaje(m *domain.TicketMensaje) {
	if m == nil {
		return
	}
	switch m.AuthorType {
	case domain.AutorAdmin:
		if m.Admin != nil {
			m.AuthorName = m.Admin.Username
		}
	case domain.AutorPsi:
		if m.Psi != nil {
			m.AuthorName = strings.TrimSpace(m.Psi.FirstName + " " + m.Psi.LastName)
		}
	}
	for i := range m.Adjuntos {
		if s.s3Client != nil {
			m.Adjuntos[i].URL = s.s3Client.GetPublicURL(m.Adjuntos[i].S3Key)
		}
	}
}

// psiNameConstruye un nombre de referencia para la auditoría.
func psiAuditName(psi *domain.PsiUserModel) string {
	if psi == nil {
		return "psi"
	}
	return strings.TrimSpace(psi.FirstName + " " + psi.LastName)
}

// processTicketAttachments sanea (imagen→WebP o PDF validado) y sube a S3 los
// anexos de un mensaje de la conversación. Devuelve los modelos listos para
// persistir (sin MensajeID).
func (s *TicketService) processTicketAttachments(ctx context.Context, files []*multipart.FileHeader) ([]domain.TicketAdjunto, error) {
	if len(files) == 0 || s.s3Client == nil {
		return nil, nil
	}

	adjuntos := make([]domain.TicketAdjunto, 0, len(files))
	for _, file := range files {
		if file == nil {
			continue
		}
		src, err := file.Open()
		if err != nil {
			return nil, domain.ErrInvalidRequest
		}

		cleanBytes, ext, contentType, err := utils.SanitizeDocumentFile(src)
		_ = src.Close()
		if err != nil {
			return nil, err
		}

		shortUUID := uuid.Must(uuid.NewV7()).String()[:6]
		filename := fmt.Sprintf("ticket_m_%s%s", shortUUID, ext)
		key, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "tickets", filename, contentType)
		if err != nil {
			return nil, err
		}

		adjuntos = append(adjuntos, domain.TicketAdjunto{
			S3Key:        key,
			OriginalName: file.Filename,
			MimeType:     contentType,
			SizeBytes:    int64(len(cleanBytes)),
		})
	}
	return adjuntos, nil
}

// createMensajeWithAttachments persiste un comentario y sus anexos con rollback
// si algo falla (la conversación es inmutable, no se puede dejar a medias).
func (s *TicketService) createMensajeWithAttachments(ctx context.Context, msg *domain.TicketMensaje, files []*multipart.FileHeader) error {
	if err := s.repo.CreateMensaje(ctx, msg); err != nil {
		return err
	}

	adjuntos, err := s.processTicketAttachments(ctx, files)
	if err != nil {
		_ = s.repo.DeleteMensaje(ctx, msg.ID)
		return err
	}
	for i := range adjuntos {
		adjuntos[i].MensajeID = msg.ID
	}
	if len(adjuntos) == 0 {
		return nil
	}
	if err := s.repo.CreateAdjuntos(ctx, adjuntos); err != nil {
		for _, a := range adjuntos {
			_ = s.s3Client.DeleteFile(context.Background(), a.S3Key)
		}
		_ = s.repo.DeleteMensaje(ctx, msg.ID)
		return err
	}
	msg.Adjuntos = adjuntos
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CONFIGURACIÓN (panel admin)
// ─────────────────────────────────────────────────────────────────────────────

// ListMotivosConfig devuelve los motivos con sus estados (admin y psi). Es la
// única fuente de configuración: un ticket pertenece a un motivo, y cada motivo
// define su propio límite de tickets abiertos por psicólogo.
func (s *TicketService) ListMotivosConfig(ctx context.Context) ([]domain.TicketMotivo, error) {
	return s.configRepo.ListMotivos(ctx)
}

// buildDefaultEstados construye los 3 estados por defecto de un motivo.
func buildDefaultEstados(now time.Time, by string, byID uuid.UUID) []domain.TicketEstado {
	estados := make([]domain.TicketEstado, 0, len(defaultEstados))
	for i, e := range defaultEstados {
		estados = append(estados, domain.TicketEstado{
			AuditModel: domain.AuditModel{CreateBy: by, CreateById: &byID, UpdateBy: by, UpdateById: &byID, CreatedAt: now, UpdatedAt: now},
			MotivoID:   0, // se setea en la transacción
			Name:      e.name,
			Order:     i + 1,
			IsClosed:  e.isClosed,
		})
	}
	return estados
}

// ListMotivos devuelve los motivos disponibles.
func (s *TicketService) ListMotivos(ctx context.Context) ([]domain.TicketMotivo, error) {
	return s.configRepo.ListMotivos(ctx)
}

// CreateMotivo crea un motivo y siembra los estados por defecto. El límite de
// tickets abiertos por psicólogo para este motivo viene en tickets_per_psi.
func (s *TicketService) CreateMotivo(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreateTicketMotivoRequest) (*domain.TicketMotivo, error) {
	if !canManageTickets(admin) {
		return nil, domain.ErrInsufficientPerms
	}
	if req.TicketsPerPsi < 1 {
		return nil, domain.ErrMotivoLimitInvalid
	}
	name := sanitizeTicketText(req.Name, maxMotivoNameLen)
	if name == "" {
		return nil, domain.ErrInvalidRequest
	}

	now := time.Now()
	motivo := &domain.TicketMotivo{
		AuditModel: domain.AuditModel{CreateBy: admin.Username, CreateById: &admin.ID, UpdateBy: admin.Username, UpdateById: &admin.ID, CreatedAt: now, UpdatedAt: now},
		Name:        name,
		Description: sanitizeTicketText(req.Description, maxMotivoDescLen),
		TicketsPerPsi: req.TicketsPerPsi,
	}
	if err := s.configRepo.CreateMotivoWithDefaults(ctx, motivo, buildDefaultEstados(now, admin.Username, admin.ID)); err != nil {
		return nil, err
	}
	return s.configRepo.GetMotivo(ctx, motivo.ID)
}

// UpdateMotivo actualiza la metadata de un motivo.
func (s *TicketService) UpdateMotivo(ctx context.Context, admin *domain.UserAdmin, motivoID uint, req request_structs.UpdateTicketMotivoRequest) (*domain.TicketMotivo, error) {
	if !canManageTickets(admin) {
		return nil, domain.ErrInsufficientPerms
	}
	motivo, err := s.configRepo.GetMotivo(ctx, motivoID)
	if err != nil {
		return nil, domain.ErrMotivoNotFound
	}
	if req.Name != nil {
		name := sanitizeTicketText(*req.Name, maxMotivoNameLen)
		if name == "" {
			return nil, domain.ErrInvalidRequest
		}
		motivo.Name = name
	}
	if req.Description != nil {
		motivo.Description = sanitizeTicketText(*req.Description, maxMotivoDescLen)
	}
	if req.TicketsPerPsi != nil {
		if *req.TicketsPerPsi < 1 {
			return nil, domain.ErrMotivoLimitInvalid
		}
		motivo.TicketsPerPsi = *req.TicketsPerPsi
	}
	motivo.UpdateBy = admin.Username
	motivo.UpdateById = &admin.ID
	motivo.UpdatedAt = time.Now()
	if err := s.configRepo.UpdateMotivo(ctx, motivo); err != nil {
		return nil, err
	}
	return s.configRepo.GetMotivo(ctx, motivoID)
}

// DeleteMotivo elimina un motivo solo si no tiene tickets asociados.
func (s *TicketService) DeleteMotivo(ctx context.Context, admin *domain.UserAdmin, motivoID uint) error {
	if !canManageTickets(admin) {
		return domain.ErrInsufficientPerms
	}
	if _, err := s.configRepo.GetMotivo(ctx, motivoID); err != nil {
		return domain.ErrMotivoNotFound
	}
	count, err := s.configRepo.CountTicketsByMotivo(ctx, motivoID)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrMotivoInUse
	}
	return s.configRepo.DeleteMotivo(ctx, motivoID)
}

// ListEstadosConfig devuelve los estados de un motivo (orden ASC).
func (s *TicketService) ListEstadosConfig(ctx context.Context, motivoID uint) ([]domain.TicketEstado, error) {
	return s.configRepo.ListEstados(ctx, motivoID)
}

// CreateEstado agrega un estado a un motivo.
func (s *TicketService) CreateEstado(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreateTicketEstadoRequest) (*domain.TicketEstado, error) {
	if !canManageTickets(admin) {
		return nil, domain.ErrInsufficientPerms
	}
	if _, err := s.configRepo.GetMotivo(ctx, req.MotivoID); err != nil {
		return nil, domain.ErrMotivoNotFound
	}
	name := sanitizeTicketText(req.Name, maxEstadoNameLen)
	if name == "" {
		return nil, domain.ErrInvalidRequest
	}
	now := time.Now()
	estado := &domain.TicketEstado{
		AuditModel: domain.AuditModel{CreateBy: admin.Username, CreateById: &admin.ID, UpdateBy: admin.Username, UpdateById: &admin.ID, CreatedAt: now, UpdatedAt: now},
		MotivoID:  req.MotivoID,
		Name:      name,
		Order:     req.Order,
		IsClosed:  req.IsClosed,
	}
	if err := s.configRepo.CreateEstado(ctx, estado); err != nil {
		return nil, err
	}
	return s.configRepo.GetEstado(ctx, estado.ID)
}

// UpdateEstadoConfig actualiza un estado (reyordenar, renombrar, marcar cerrado).
func (s *TicketService) UpdateEstadoConfig(ctx context.Context, admin *domain.UserAdmin, estadoID uint, req request_structs.UpdateTicketEstadoRequest) (*domain.TicketEstado, error) {
	if !canManageTickets(admin) {
		return nil, domain.ErrInsufficientPerms
	}
	estado, err := s.configRepo.GetEstado(ctx, estadoID)
	if err != nil {
		return nil, domain.ErrEstadoNotFound
	}
	if req.Name != nil {
		name := sanitizeTicketText(*req.Name, maxEstadoNameLen)
		if name == "" {
			return nil, domain.ErrInvalidRequest
		}
		estado.Name = name
	}
	if req.Order != nil {
		estado.Order = *req.Order
	}
	if req.IsClosed != nil {
		estado.IsClosed = *req.IsClosed
	}
	estado.UpdateBy = admin.Username
	estado.UpdateById = &admin.ID
	estado.UpdatedAt = time.Now()
	if err := s.configRepo.UpdateEstado(ctx, estado); err != nil {
		return nil, err
	}
	return s.configRepo.GetEstado(ctx, estadoID)
}

// DeleteEstadoConfig elimina un estado siempre que ningún ticket lo use.
func (s *TicketService) DeleteEstadoConfig(ctx context.Context, admin *domain.UserAdmin, estadoID uint) error {
	if !canManageTickets(admin) {
		return domain.ErrInsufficientPerms
	}
	if _, err := s.configRepo.GetEstado(ctx, estadoID); err != nil {
		return domain.ErrEstadoNotFound
	}
	inUse, err := s.configRepo.IsEstadoInUse(ctx, estadoID)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrEstadoInUse
	}
	return s.configRepo.DeleteEstado(ctx, estadoID)
}

// ─────────────────────────────────────────────────────────────────────────────
// PORTAL PSI
// ─────────────────────────────────────────────────────────────────────────────

// ListMyTickets lista los tickets del psicólogo (más recientes primero).
func (s *TicketService) ListMyTickets(ctx context.Context, psi *domain.PsiUserModel, page, limit int) ([]domain.Ticket, int64, error) {
	tickets, total, err := s.repo.ListMyTickets(ctx, psi.ID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range tickets {
		s.fillTicket(&tickets[i])
	}
	return tickets, total, nil
}

// GetTicketAsPsi devuelve el detalle de un ticket validando la propiedad.
func (s *TicketService) GetTicketAsPsi(ctx context.Context, psi *domain.PsiUserModel, ticketID uint) (*domain.Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}
	if ticket.PsiUserID != psi.ID {
		return nil, domain.ErrTicketNotOwner
	}
	s.fillTicket(ticket)
	return ticket, nil
}

// CreateTicket abre un nuevo ticket del psicólogo. Valida:
//   - el motivo existe;
//   - límite de tickets abiertos del psi para ese motivo (tickets_per_psi);
//   - siembra una conversación con un primer mensaje = descripción inicial
//     (con los anexos opcionales de la creación).
func (s *TicketService) CreateTicket(ctx context.Context, psi *domain.PsiUserModel, req request_structs.CreateTicketRequest, files []*multipart.FileHeader) (*domain.Ticket, error) {
	// Recepción global: si el colegio desactivó la apertura de tickets, bloquear.
	if err := AssertReceptionEnabled(ctx, s.settings, domain.SettingsKeyTicketsReception); err != nil {
		return nil, err
	}

	motivo, err := s.configRepo.GetMotivo(ctx, req.MotivoID)
	if err != nil {
		return nil, domain.ErrMotivoNotFound
	}

	initial := findInitialEstado(motivo.Estados)
	if initial == nil {
		return nil, domain.ErrEstadoNotFound
	}

	// Límite por motivo: solo cuentan tickets abiertos (no cerrados).
	count, err := s.repo.CountActiveByPsiAndMotivo(ctx, psi.ID, motivo.ID)
	if err != nil {
		return nil, err
	}
	if count >= int64(motivo.TicketsPerPsi) {
		return nil, domain.ErrTicketLimitReached
	}

	title := sanitizeTicketText(req.Title, maxTicketTitleLen)
	description := sanitizeTicketText(req.Description, maxTicketDescLen)
	if title == "" || description == "" {
		return nil, domain.ErrInvalidRequest
	}

	now := time.Now()
	ticket := &domain.Ticket{
		AuditModel:  domain.AuditModel{CreateBy: psiAuditName(psi), CreateById: &psi.ID, UpdateBy: psiAuditName(psi), UpdateById: &psi.ID, CreatedAt: now, UpdatedAt: now},
		PsiUserID:   psi.ID,
		MotivoID:    motivo.ID,
		EstadoID:    initial.ID,
		Title:       title,
		Description: description,
	}

	initialLog := &domain.TicketStatusLog{
		TicketID:      0, // se setea en la transacción
		PreviousStateID: nil,
		NewStateID:     initial.ID,
		ChangedByType:  domain.AutorPsi,
		ChangedByPsiID: &psi.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateTicket(ctx, ticket, initialLog); err != nil {
		return nil, err
	}

	// Primer mensaje de la conversación = descripción inicial (+ anexos opcionales).
	firstMsg := &domain.TicketMensaje{
		TicketID:    ticket.ID,
		AuthorType:  domain.AutorPsi,
		AuthorPsiID: &psi.ID,
		Message:     description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.createMensajeWithAttachments(ctx, firstMsg, files); err != nil {
		// No crítico: la descripción queda en el ticket; solo perdemos la
		// copia inicial en la conversación.
		log.Warn().Err(err).Uint("ticket_id", ticket.ID).Str("component", "tickets").Msg("No se pudo sembrar el primer mensaje de la conversación")
	}

	ticket, err = s.repo.GetByID(ctx, ticket.ID)
	if err != nil {
		return nil, err
	}
	s.fillTicket(ticket)
	return ticket, nil
}

// AddMensajeAsPsi publica un comentario del psicólogo respetando:
//   - ticket abierto (no cerrado);
//   - máximo 1000 caracteres;
//   - no más de 3 mensajes seguidos del mismo psi.
func (s *TicketService) AddMensajeAsPsi(ctx context.Context, psi *domain.PsiUserModel, ticketID uint, message string, files []*multipart.FileHeader) (*domain.TicketMensaje, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}
	if ticket.PsiUserID != psi.ID {
		return nil, domain.ErrTicketNotOwner
	}
	if ticket.Estado != nil && ticket.Estado.IsClosed {
		return nil, domain.ErrTicketClosed
	}

	msg := sanitizeTicketText(message, MaxPsiMensajeLengthChars)
	if msg == "" {
		return nil, domain.ErrMensajeVacio
	}
	if utf8.RuneCountInString(msg) > MaxPsiMensajeLengthChars {
		return nil, domain.ErrMensajeTooLong
	}

	// Regla anti-spam: revisar los últimos 3 mensajes del ticket.
	ultimos, err := s.repo.ListLastMensajes(ctx, ticketID, MaxConsecutivePsiComments)
	if err != nil {
		return nil, err
	}
	if len(ultimos) >= MaxConsecutivePsiComments {
		allByPsi := true
		for i := range ultimos {
			u := &ultimos[i]
			if u.AuthorType != domain.AutorPsi || u.AuthorPsiID == nil || *u.AuthorPsiID != psi.ID {
				allByPsi = false
				break
			}
		}
		if allByPsi {
			return nil, domain.ErrMaxConsecutiveComments
		}
	}

	now := time.Now()
	mensaje := &domain.TicketMensaje{
		TicketID:    ticketID,
		AuthorType:  domain.AutorPsi,
		AuthorPsiID: &psi.ID,
		Message:     msg,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.createMensajeWithAttachments(ctx, mensaje, files); err != nil {
		return nil, err
	}
	// Cargar nombre del autor para la respuesta.
	mensaje.Psi = psi
	s.fillMensaje(mensaje)
	return mensaje, nil
}

// CloseTicketAsPsi cierra el ticket del psicólogo. El motivo de cierre es
// obligatorio. El estado pasa al primer estado cerrado del motivo.
func (s *TicketService) CloseTicketAsPsi(ctx context.Context, psi *domain.PsiUserModel, ticketID uint, req request_structs.CloseTicketRequest) (*domain.Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}
	if ticket.PsiUserID != psi.ID {
		return nil, domain.ErrTicketNotOwner
	}
	if ticket.Estado != nil && ticket.Estado.IsClosed {
		return nil, domain.ErrTicketClosed
	}

	closeReason := sanitizeTicketText(req.CloseReason, maxCloseReasonLen)
	if closeReason == "" {
		return nil, domain.ErrCloseReasonRequired
	}

	motivo, err := s.configRepo.GetMotivo(ctx, ticket.MotivoID)
	if err != nil {
		return nil, domain.ErrMotivoNotFound
	}
	closed := findClosedEstado(motivo.Estados)
	if closed == nil {
		return nil, domain.ErrEstadoNotFound
	}

	return s.applyClose(ctx, ticket, closed.ID, closeReason, domain.AutorPsi, &psi.ID, nil, psiAuditName(psi))
}

// ─────────────────────────────────────────────────────────────────────────────
// PANEL ADMINISTRATIVO
// ─────────────────────────────────────────────────────────────────────────────

// ListTicketsAdmin lista los tickets de la cola FIFO (por defecto solo abiertos).
func (s *TicketService) ListTicketsAdmin(ctx context.Context, admin *domain.UserAdmin, filter domain.TicketFilter) ([]domain.Ticket, int64, error) {
	if !canManageTickets(admin) {
		return nil, 0, domain.ErrInsufficientPerms
	}
	tickets, total, err := s.repo.ListTickets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	for i := range tickets {
		s.fillTicket(&tickets[i])
	}
	return tickets, total, nil
}

// GetTicketAdmin devuelve el detalle completo de un ticket (conversación e historial).
func (s *TicketService) GetTicketAdmin(ctx context.Context, admin *domain.UserAdmin, ticketID uint) (*domain.Ticket, error) {
	if !canManageTickets(admin) {
		return nil, domain.ErrInsufficientPerms
	}
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}
	s.fillTicket(ticket)
	return ticket, nil
}

// CountPendientesAdmin devuelve el número de tickets abiertos (badge del menú admin).
func (s *TicketService) CountPendientesAdmin(ctx context.Context, admin *domain.UserAdmin) (int64, error) {
	if !canManageTickets(admin) {
		return 0, domain.ErrInsufficientPerms
	}
	return s.repo.CountPendientesAdmin(ctx)
}

// AddMensajeAsAdmin responde en la conversación de un ticket. El admin no tiene
// límite de mensajes seguidos ni el límite de caracteres del psi; sí el ticket
// debe estar abierto.
func (s *TicketService) AddMensajeAsAdmin(ctx context.Context, admin *domain.UserAdmin, ticketID uint, message string, files []*multipart.FileHeader) (*domain.TicketMensaje, error) {
	if !canManageTickets(admin) {
		return nil, domain.ErrInsufficientPerms
	}
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}
	if ticket.Estado != nil && ticket.Estado.IsClosed {
		return nil, domain.ErrTicketClosed
	}

	msg := sanitizeTicketText(message, MaxAdminMensajeLengthChars)
	if msg == "" {
		return nil, domain.ErrMensajeVacio
	}
	if utf8.RuneCountInString(msg) > MaxAdminMensajeLengthChars {
		return nil, domain.ErrMensajeTooLong
	}

	now := time.Now()
	mensaje := &domain.TicketMensaje{
		TicketID:     ticketID,
		AuthorType:   domain.AutorAdmin,
		AuthorAdminID: &admin.ID,
		Message:      msg,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.createMensajeWithAttachments(ctx, mensaje, files); err != nil {
		return nil, err
	}
	mensaje.Admin = admin
	mensaje.AuthorName = admin.Username
	for i := range mensaje.Adjuntos {
		if s.s3Client != nil {
			mensaje.Adjuntos[i].URL = s.s3Client.GetPublicURL(mensaje.Adjuntos[i].S3Key)
		}
	}

	s.notifyPsi(ctx, admin.ID, admin.Username, ticket, fmt.Sprintf("El colegio respondió en tu ticket #%d (%s)", ticket.ID, ticket.Title))
	return mensaje, nil
}

// UpdateTicketEstado cambia el estado administrativo de un ticket. El nuevo
// estado debe pertenecer al motivo del ticket. Notifica al psicólogo.
// Si el estado es "cerrado", registra el cierre con auditoría.
func (s *TicketService) UpdateTicketEstado(ctx context.Context, admin *domain.UserAdmin, ticketID uint, req request_structs.UpdateTicketEstado) (*domain.Ticket, error) {
	if !canManageTickets(admin) {
		return nil, domain.ErrInsufficientPerms
	}
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}
	estado, err := s.configRepo.GetEstado(ctx, req.EstadoID)
	if err != nil {
		return nil, domain.ErrEstadoNotFound
	}
	if estado.MotivoID != ticket.MotivoID {
		return nil, domain.ErrEstadoNotInMotivo
	}

	prevStateID := ticket.EstadoID
	wasClosed := ticket.Estado != nil && ticket.Estado.IsClosed
	now := time.Now()
	reason := sanitizeTicketText(req.Reason, maxChangeReasonLen)

	ticket.EstadoID = estado.ID
	ticket.Estado = estado
	ticket.UpdateBy = admin.Username
	ticket.UpdateById = &admin.ID
	ticket.UpdatedAt = now

	if estado.IsClosed && !wasClosed {
		ticket.CloseReason = reason
		ticket.ClosedByType = domain.AutorAdmin
		ticket.ClosedByAdminID = &admin.ID
		ticket.ClosedAt = &now
	} else if !estado.IsClosed {
		ticket.CloseReason = ""
		ticket.ClosedByType = ""
		ticket.ClosedByAdminID = nil
		ticket.ClosedByPsiID = nil
		ticket.ClosedAt = nil
	}

	if err := s.repo.UpdateEstado(ctx, ticket); err != nil {
		return nil, err
	}
	log := &domain.TicketStatusLog{
		TicketID:        ticket.ID,
		PreviousStateID: &prevStateID,
		NewStateID:      estado.ID,
		ChangedByType:   domain.AutorAdmin,
		ChangedByAdminID: &admin.ID,
		Reason:          reason,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateStatusLog(ctx, log); err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("El estado de tu ticket #%d (%s) cambió a «%s»", ticket.ID, ticket.Title, estado.Name)
	if reason != "" {
		msg += ". " + reason
	}
	if estado.IsClosed && !wasClosed {
		msg = fmt.Sprintf("Tu ticket #%d (%s) fue cerrado: %s", ticket.ID, ticket.Title, reason)
	}
	s.notifyPsi(ctx, admin.ID, admin.Username, ticket, msg)

	s.fillTicket(ticket)
	return ticket, nil
}

// CloseTicketAsAdmin cierra un ticket desde el panel. Motivo obligatorio.
func (s *TicketService) CloseTicketAsAdmin(ctx context.Context, admin *domain.UserAdmin, ticketID uint, req request_structs.CloseTicketRequest) (*domain.Ticket, error) {
	if !canManageTickets(admin) {
		return nil, domain.ErrInsufficientPerms
	}
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}
	if ticket.Estado != nil && ticket.Estado.IsClosed {
		return nil, domain.ErrTicketClosed
	}

	closeReason := sanitizeTicketText(req.CloseReason, maxCloseReasonLen)
	if closeReason == "" {
		return nil, domain.ErrCloseReasonRequired
	}

	motivo, err := s.configRepo.GetMotivo(ctx, ticket.MotivoID)
	if err != nil {
		return nil, domain.ErrMotivoNotFound
	}
	closed := findClosedEstado(motivo.Estados)
	if closed == nil {
		return nil, domain.ErrEstadoNotFound
	}

	return s.applyClose(ctx, ticket, closed.ID, closeReason, domain.AutorAdmin, nil, &admin.ID, admin.Username)
}

// applyClose ejecuta el cierre común (empelado por psi y admin), crea el log de
// estado y, si el cierre fue del admin, notifica al psicólogo.
func (s *TicketService) applyClose(ctx context.Context, ticket *domain.Ticket, closedStateID uint, reason string, byType domain.AutorType, byPsiID, byAdminID *uuid.UUID, byName string) (*domain.Ticket, error) {
	prevStateID := ticket.EstadoID
	now := time.Now()

	ticket.EstadoID = closedStateID
	ticket.CloseReason = reason
	ticket.ClosedByType = byType
	ticket.ClosedByPsiID = byPsiID
	ticket.ClosedByAdminID = byAdminID
	ticket.ClosedAt = &now
	ticket.UpdateBy = byName
	ticket.UpdatedAt = now

	if err := s.repo.UpdateEstado(ctx, ticket); err != nil {
		return nil, err
	}
	log := &domain.TicketStatusLog{
		TicketID:        ticket.ID,
		PreviousStateID: &prevStateID,
		NewStateID:      closedStateID,
		ChangedByType:   byType,
		ChangedByPsiID:  byPsiID,
		ChangedByAdminID: byAdminID,
		Reason:          reason,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateStatusLog(ctx, log); err != nil {
		return nil, err
	}

	if byType == domain.AutorAdmin && byAdminID != nil {
		s.notifyPsi(ctx, *byAdminID, byName, ticket, fmt.Sprintf("Tu ticket #%d (%s) fue cerrado: %s", ticket.ID, ticket.Title, reason))
	}

	ticket, err := s.repo.GetByID(ctx, ticket.ID)
	if err != nil {
		return nil, err
	}
	s.fillTicket(ticket)
	return ticket, nil
}

// notifyPsi dispara la notificación al psicólogo dueño del ticket. Los errores
// de notificación no bloquean la operación principal (se registran en logs).
func (s *TicketService) notifyPsi(ctx context.Context, senderID uuid.UUID, senderName string, ticket *domain.Ticket, message string) {
	if s.notifier == nil || ticket == nil {
		return
	}
	title := fmt.Sprintf("Actualización de tu ticket #%d", ticket.ID)
	if err := s.notifier.NotifyPSI(ctx, senderID, senderName, ticket.PsiUserID, title, message); err != nil {
		log.Error().Err(err).Uint("ticket_id", ticket.ID).Str("component", "tickets").Msg("No se pudo notificar al psicólogo del ticket")
	}
}
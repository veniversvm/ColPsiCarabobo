// api/internal/service/audit_service.go

// AuditService difiere la escritura de los eventos de auditoría para que el log
// NUNCA interfiera con el flujo de la petición (mismo patrón productor-consumidor
// que el MailService): `Record` captura y encola en memoria; un worker en
// background desagua la cola en lotes hacia `api_change_logs`.
//
// Es best-effort: un fallo en el worker o una cola llena jamás revierten la
// operación principal; el evento se descarta con log.Warn().
package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/datatypes"
)

const (
	auditQueueBuffer   = 5000 // Eventos en memoria antes de descartar (drop)
	auditBatchSize     = 50   // Filas por INSERT lote en el worker
	auditFlushInterval = time.Second
	auditWriteTimeout  = 5 * time.Second
)

// AuditEvent describe un cambio a registrar en la bitácora. Los campos IP y
// UserAgent se completan desde el contexto de la petición si vienen vacíos.
type AuditEvent struct {
	Entity        string
	EntityID      string
	EntityLabel   string
	Action        string
	ActorID       uuid.UUID
	ActorRole     string
	ActorUsername string
	IP            string
	UserAgent     string
	Changes       map[string]domain.AuditChange
	Metadata      map[string]any
}

// AuditService gestiona la cola de eventos y el worker de persistencia.
type AuditService struct {
	repo      domain.AuditRepository
	queue     chan AuditEvent
	mu        sync.Mutex
	closed    bool
	batchSize int
}

// NewAuditService construye el servicio con su cola. El worker se arranca con
// Start(ctx) usando el contexto de background del proceso.
func NewAuditService(repo domain.AuditRepository) *AuditService {
	return &AuditService{
		repo:      repo,
		queue:     make(chan AuditEvent, auditQueueBuffer),
		batchSize: auditBatchSize,
	}
}

// Start lanza el worker de fondo.
func (s *AuditService) Start(ctx context.Context) {
	go s.startWorker(ctx)
}

// Close detiene la captura de nuevos eventos y cierra la cola (el worker drena
// lo pendiente antes de salir). Idempotente.
func (s *AuditService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.queue)
}

// Record encola un evento SIN bloquear el hilo de la petición. Si la cola está
// llena, el evento se descarta (el log es evidencia, no carga).
func (s *AuditService) Record(evt AuditEvent) {
	if s == nil {
		return
	}
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return
	}
	select {
	case s.queue <- evt:
	default:
		log.Warn().
			Str("component", "audit").
			Str("entity", evt.Entity).
			Str("action", evt.Action).
			Msg("Cola de audit logs llena; evento descartado")
	}
}

// List registra los filtros de búsqueda contra el repositorio.
func (s *AuditService) List(ctx context.Context, f domain.AuditLogFilters) ([]domain.ApiChangeLog, int64, error) {
	return s.repo.List(ctx, f)
}

// Stats devuelve los agregados por (entity, action) desde `since`.
func (s *AuditService) Stats(ctx context.Context, since time.Time) ([]domain.AuditStat, error) {
	return s.repo.Stats(ctx, since)
}

// PurgeOlderThan borra los sucesos anteriores a `before` (retención).
func (s *AuditService) PurgeOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return s.repo.PurgeOlderThan(ctx, before)
}

// startWorker es el consumidor perpetuo de la cola: acumula eventos y los
// persiste en lotes (o por tick). Al cancelarse el contexto hace un último
// drenaje best-effort antes de salir.
func (s *AuditService) startWorker(ctx context.Context) {
	log.Info().Str("component", "audit").Msg("Audit Log Worker iniciado y escuchando cola...")

	buf := make([]domain.ApiChangeLog, 0, s.batchSize)
	ticker := time.NewTicker(auditFlushInterval)
	defer ticker.Stop()

	flush := func(cancelCtx context.Context) {
		if len(buf) == 0 {
			return
		}
		rows := make([]domain.ApiChangeLog, len(buf))
		copy(rows, buf)
		buf = buf[:0]
		fctx, cancel := context.WithTimeout(cancelCtx, auditWriteTimeout)
		defer cancel()
		if err := s.repo.CreateBatch(fctx, rows); err != nil {
			log.Warn().
				Err(err).
				Int("count", len(rows)).
				Str("component", "audit").
				Msg("No se pudieron persistir eventos de auditoría")
		}
	}

	for {
		select {
		case <-ctx.Done():
			// Drenaje final no bloqueante: tomar todo lo que haya en la cola.
			for {
				select {
				case evt, ok := <-s.queue:
					if !ok {
						flush(context.Background())
						return
					}
					buf = append(buf, evt.toRow())
					if len(buf) >= s.batchSize {
						flush(context.Background())
					}
				default:
					flush(context.Background())
					return
				}
			}
		case evt, ok := <-s.queue:
			if !ok {
				flush(context.Background())
				return
			}
			buf = append(buf, evt.toRow())
			if len(buf) >= s.batchSize {
				flush(context.Background())
			}
		case <-ticker.C:
			flush(context.Background())
		}
	}
}

// toRow convierte un evento en el registro persistible.
func (e AuditEvent) toRow() domain.ApiChangeLog {
	var changes datatypes.JSON
	if len(e.Changes) > 0 {
		if b, err := json.Marshal(e.Changes); err == nil {
			changes = b
		}
	}
	var metadata datatypes.JSON
	if len(e.Metadata) > 0 {
		if b, err := json.Marshal(e.Metadata); err == nil {
			metadata = b
		}
	}
	return domain.ApiChangeLog{
		ID:            uuid.Must(uuid.NewV7()),
		Entity:        e.Entity,
		EntityID:      e.EntityID,
		EntityLabel:   e.EntityLabel,
		Action:        e.Action,
		ActorID:       e.ActorID,
		ActorRole:     e.ActorRole,
		ActorUsername: e.ActorUsername,
		IP:            e.IP,
		UserAgent:     e.UserAgent,
		Changes:       changes,
		Metadata:      metadata,
		CreatedAt:     time.Now().UTC(),
	}
}

// BuildDiff compara dos estados (`before`/`after`) y devuelve únicamente los
// campos que cambiaron, en forma {campo: {from, to}}. Los campos iguales y los
// pares (ausente, cero) se omiten para no ensuciar la bitácora.
func BuildDiff(before, after map[string]any) map[string]domain.AuditChange {
	diff := map[string]domain.AuditChange{}
	keys := make(map[string]struct{}, len(before)+len(after))
	for k := range before {
		keys[k] = struct{}{}
	}
	for k := range after {
		keys[k] = struct{}{}
	}
	for k := range keys {
		b, bOk := before[k]
		a, aOk := after[k]
		if !bOk || !aOk {
			diff[k] = domain.AuditChange{From: b, To: a}
			continue
		}
		if b != a {
			diff[k] = domain.AuditChange{From: b, To: a}
		}
	}
	return diff
}
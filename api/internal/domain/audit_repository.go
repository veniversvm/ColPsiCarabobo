// api/internal/domain/audit_repository.go
// Contrato de persistencia de la bitácora de cambios (api_change_logs).
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditLogFilters agrupa los filtros combinables de búsqueda de la bitácora.
type AuditLogFilters struct {
	Entity   string // Entidad (psi, staff, notificacion, ...)
	EntityID string // ID/identificador de la entidad dentro de la búsqueda
	Action   string // Suceso (create, update, delete, login, ...)
	ActorID  *uuid.UUID
	Q        string // Búsqueda libre sobre entity_label
	From     *time.Time
	To       *time.Time
	Page     int
	Limit    int
}

// AuditStat es un agregado de sucesos por entidad/acción (para el dashboard).
type AuditStat struct {
	Entity string `json:"entity"`
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// AuditRepository define el acceso a la bitácora de cambios.
type AuditRepository interface {
	// CreateBatch persiste un lote de eventos (usado por el worker diferido).
	CreateBatch(ctx context.Context, logs []ApiChangeLog) error
	// List devuelve los sucesos que coinciden con los filtros, paginados y
	// ordenados por fecha descendente, junto con el total real (para paginar).
	List(ctx context.Context, f AuditLogFilters) ([]ApiChangeLog, int64, error)
	// Stats agrupa los sucesos desde `since` por (entity, action).
	Stats(ctx context.Context, since time.Time) ([]AuditStat, error)
	// PurgeOlderThan borra los sucesos anteriores a `before` (retención).
	// Devuelve cuántas filas se eliminaron.
	PurgeOlderThan(ctx context.Context, before time.Time) (int64, error)
}
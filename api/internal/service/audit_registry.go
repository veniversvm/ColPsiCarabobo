// api/internal/service/audit_registry.go

// Registro global de emisión de auditoría (nil-safe).
//
// Se elige una instancia global (única por proceso) en lugar de inyectar el
// AuditService en los ~320 constructores de servicios/tests: todos los servicios
// del paquete pueden emitir eventos sin tocar sus firmas, y los tests quedan
// nil-safe (sin instancia, `RecordAudit` no hace nada).
//
// Inicialización en el arranque (cmd/api/main.go): InitAuditLogs(auditSvc).
package service

import (
	"context"
	"sync"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

var (
	globalAudit   *AuditService
	globalAuditMu sync.RWMutex
)

// InitAuditLogs registra la instancia global del AuditService. Debe llamarse
// una sola vez al arrancar la API, antes de servir tráfico.
func InitAuditLogs(svc *AuditService) {
	globalAuditMu.Lock()
	defer globalAuditMu.Unlock()
	globalAudit = svc
}

// auditLog devuelve la instancia global (puede ser nil en tests).
func auditLog() *AuditService {
	globalAuditMu.RLock()
	defer globalAuditMu.RUnlock()
	return globalAudit
}

// auditMetaKey y auditMeta se transportan en el context.Context de la petición
// para que los servicios etiqueten IP/User-Agent sin recibir el fiber.Ctx.
type auditMetaKey struct{}

type auditMeta struct {
	IP        string
	UserAgent string
}

// WithAuditMeta inyecta IP/User-Agent en el contexto de la petición. Lo usa el
// middleware AuditRequestMeta antes de que el handler llame al servicio.
func WithAuditMeta(ctx context.Context, ip, userAgent string) context.Context {
	return context.WithValue(ctx, auditMetaKey{}, auditMeta{IP: ip, UserAgent: userAgent})
}

// auditMetaFrom extrae IP/User-Agent del contexto (vacío si no viajó).
func auditMetaFrom(ctx context.Context) auditMeta {
	if ctx == nil {
		return auditMeta{}
	}
	if meta, ok := ctx.Value(auditMetaKey{}).(auditMeta); ok {
		return meta
	}
	return auditMeta{}
}

// RecordAudit emite un evento de auditoría de forma diferida (cola + worker).
// Completa IP/User-Agent desde el contexto de la petición si el evento no los
// trae. Nil-safe: sin instancia registrada no hace nada (no rompe el flujo).
func RecordAudit(ctx context.Context, evt AuditEvent) {
	svc := auditLog()
	if svc == nil {
		return
	}
	meta := auditMetaFrom(ctx)
	if evt.IP == "" {
		evt.IP = meta.IP
	}
	if evt.UserAgent == "" {
		evt.UserAgent = meta.UserAgent
	}
	svc.Record(evt)
}

// auditRoleOfAdmin resume el rol del actor admin para la bitácora:
// "sudo" si tiene acceso total, si no su etiqueta de preset
// (secretaria, comunicacion, ...), o "staff" si no tiene rotulado.
func auditRoleOfAdmin(admin *domain.UserAdmin) string {
	if admin == nil {
		return ""
	}
	if admin.Sudo {
		return "sudo"
	}
	if admin.Role != "" {
		return admin.Role
	}
	return "staff"
}
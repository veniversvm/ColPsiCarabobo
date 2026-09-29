// api/internal/service/psi_service_terms.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// =========================================================================
// TÉRMINOS Y CONDICIONES — PARTE II (AGREMIADOS)
// =========================================================================

// TermsStatus es lo que el frontend necesita para decidir si mostrar el modal.
// Deliberadamente mínimo: NO expone el texto (que vive en el frontend como
// módulo estático, patrón de `lib/documentos/`) ni datos del agremiado.
type TermsStatus struct {
	// CurrentVersion es la versión vigente que declara el backend.
	CurrentVersion string `json:"current_version"`
	// Accepted indica si el agremiado ya aceptó ESA versión exacta.
	// Si el documento cambia, vuelve a false y el modal reaparece.
	Accepted bool `json:"accepted"`
	// AcceptedAt es cuándo la aceptó (null si nunca).
	AcceptedAt *time.Time `json:"accepted_at"`
	// History son las aceptaciones previas, de la más reciente a la más antigua.
	History []TermsAcceptanceSummary `json:"history"`
}

// TermsAcceptanceSummary es una fila del historial.
type TermsAcceptanceSummary struct {
	Version    string    `json:"version"`
	AcceptedAt time.Time `json:"accepted_at"`
}

// TermsService expone la versión vigente de los términos y registra las
// aceptaciones de la Parte II.
//
// El control de aceptación es responsabilidad del CLIENTE: el backend no
// bloquea ninguna ruta si el agremiado no ha aceptado. Motivo: si el backend
// rechazara el acceso, un fallo de la tabla (o un despliegue a medias) dejaría
// a los agremiados fuera del portal. El aviso vive en el modal, y esta API solo
// informa. Ver `psiTermsAcceptanceRepo`: tampoco hay Update/Delete.
type TermsService struct {
	repo domain.PsiTermsAcceptanceRepository
}

// NewTermsService construye el servicio de términos.
func NewTermsService(repo domain.PsiTermsAcceptanceRepository) *TermsService {
	return &TermsService{repo: repo}
}

// CurrentVersion devuelve la versión vigente (constante del backend).
func (s *TermsService) CurrentVersion() string { return domain.TermsVersion }

// GetStatus devuelve el estado de aceptación del agremiado para la versión
// vigente, con su historial.
//
// Un error de repositorio NO se silencia: se propaga para que el handler
// responda 500. El frontend, en cambio, trata ese fallo como "no mostrar modal"
// (degradación segura), nunca como "bloquear el portal".
func (s *TermsService) GetStatus(ctx context.Context, psiUserID uuid.UUID) (*TermsStatus, error) {
	current, err := s.repo.GetByUserAndVersion(ctx, psiUserID, domain.TermsVersion)
	if err != nil {
		return nil, err
	}

	status := &TermsStatus{
		CurrentVersion: domain.TermsVersion,
		Accepted:       current != nil,
		History:        []TermsAcceptanceSummary{},
	}
	if current != nil {
		status.AcceptedAt = current.AcceptedAt
	}

	history, err := s.repo.ListByUser(ctx, psiUserID)
	if err != nil {
		return nil, err
	}
	for _, h := range history {
		summary := TermsAcceptanceSummary{Version: h.Version}
		if h.AcceptedAt != nil {
			summary.AcceptedAt = *h.AcceptedAt
		} else {
			summary.AcceptedAt = h.CreatedAt
		}
		status.History = append(status.History, summary)
	}

	return status, nil
}

// Accept registra la aceptación de la Parte II.
//
// IDEMPOTENTE por diseño: si el agremiado ya aceptó esa versión, no inserta una
// fila duplicada ni emite un segundo evento de bitácora, y devuelve el estado
// ya aceptado. El caso real es el doble clic impaciente o el reintento de una
// respuesta que se perdió.
//
// La versión se valida contra la constante del backend. Aceptar un texto distinto
// del vigente exigiría que el cliente afirmara algo que el servidor no reconoce,
// que es exactamente lo que el control de versión busca impedir.
func (s *TermsService) Accept(ctx context.Context, psiUserID uuid.UUID, version, ip, userAgent string) (*TermsStatus, error) {
	if version != domain.TermsVersion {
		return nil, domain.ErrTermsVersionInvalid
	}

	existing, err := s.repo.GetByUserAndVersion(ctx, psiUserID, version)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Ya aceptó esta versión: no se duplica la evidencia ni el evento.
		return s.GetStatus(ctx, psiUserID)
	}

	now := time.Now().UTC()
	acceptance := &domain.PsiTermsAcceptance{
		PsiUserID:  psiUserID,
		Version:    version,
		AcceptedAt: &now,
		IP:         ip,
		UserAgent:  userAgent,
	}
	if err := s.repo.Create(ctx, acceptance); err != nil {
		return nil, err
	}

	RecordAudit(ctx, AuditEvent{
		Entity:      "terminos",
		EntityID:    psiUserID.String(),
		EntityLabel: "Términos y Condiciones — Parte II (agremiados)",
		Action:      "accept",
		ActorID:     psiUserID,
		ActorRole:   "psi_user",
		IP:          ip,
		UserAgent:   userAgent,
		Metadata: map[string]any{
			"version": version,
		},
	})

	return s.GetStatus(ctx, psiUserID)
}

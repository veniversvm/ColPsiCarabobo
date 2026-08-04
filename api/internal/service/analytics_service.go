// api/internal/service/analytics_service.go

// Package service implementa la lógica de negocio central del sistema.
//
// Este archivo en particular contiene el motor de Telemetría y Business Intelligence (BI).
// Está diseñado bajo el principio de "Impacto Cero": la recolección de métricas
// nunca debe ralentizar ni interrumpir el flujo principal del usuario final.
//
// Protección contra conexiones colgadas:
//   - Todas las escrituras fire-and-forget corren bajo un contexto con timeout
//     (analyticsWriteTimeout) y un semáforo que acota la concurrencia máxima.
//     Si la BD se degrada o hay locks, una goroutine muere sola a los pocos
//     segundos en lugar de esperar una conexión del pool indefinidamente.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// analyticsWriteTimeout acota cada escritura analítica fire-and-forget.
const analyticsWriteTimeout = 5 * time.Second

// analyticsVisitWindow define el umbral de "Debouncing" para el rastreo de visitas.
// Si un mismo identificador de sesión (_sid) navega por múltiples rutas o recarga
// la página dentro de esta ventana de 30 minutos, se considera una única visita continua.
// Esto previene el inflado artificial de métricas y el agotamiento de la base de datos.
const analyticsVisitWindow = 30 * time.Minute

// analyticsSemCapacity acota cuántas operaciones analíticas pueden correr a la vez
// contra la BD. Evita que una avalancha de goroutines agote el pool de conexiones.
const analyticsSemCapacity = 50

// AnalyticsService encapsula la lógica de negocio de telemetría,
// delegando la persistencia a un AnalyticsRepository (Dependency Inversion).
type AnalyticsService struct {
	repo domain.AnalyticsRepository
	sem  chan struct{}
}

// NewAnalyticsService actúa como constructor para inyectar la dependencia del repositorio.
func NewAnalyticsService(repo domain.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		repo: repo,
		sem:  make(chan struct{}, analyticsSemCapacity),
	}
}

// acquire intenta reservar un slot del semáforo analítico.
// Espera hasta que el ctx se cancele (timeout). Si no consigue slot, devuelve
// false y el evento se descarta: es preferible perder una métrica a colgar la app.
func (s *AnalyticsService) acquire(ctx context.Context) bool {
	select {
	case s.sem <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// release libera un slot del semáforo analítico.
func (s *AnalyticsService) release() {
	<-s.sem
}

// ── ESCRITURA (FIRE-AND-FORGET) ──────────────────────────────────────────────

// RecordLogin registra una huella de auditoría cada vez que un usuario accede al sistema.
func (s *AnalyticsService) RecordLogin(userID uuid.UUID, username, role, ip, userAgent string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), analyticsWriteTimeout)
		defer cancel()
		if !s.acquire(ctx) {
			return
		}
		defer s.release()

		_ = s.repo.CreateLoginEvent(ctx, domain.LoginEvent{
			UserID:    userID,
			Username:  username,
			Role:      role,
			IP:        ip,
			UserAgent: userAgent,
		})

		now := time.Now()
		_ = s.repo.UpsertActiveSession(ctx, domain.ActiveSession{
			UserID:    userID,
			Username:  username,
			Role:      role,
			IP:        ip,
			LastSeen:  now,
			ExpiresAt: now.Add(8 * time.Hour),
		})
	}()
}

// RecordLogout destruye la estampa de sesión activa del usuario.
func (s *AnalyticsService) RecordLogout(userID uuid.UUID) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), analyticsWriteTimeout)
		defer cancel()
		if !s.acquire(ctx) {
			return
		}
		defer s.release()

		_ = s.repo.DeleteActiveSession(ctx, userID)
	}()
}

// HeartbeatSession actualiza la marca de tiempo (LastSeen) de una sesión.
func (s *AnalyticsService) HeartbeatSession(userID uuid.UUID) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), analyticsWriteTimeout)
		defer cancel()
		if !s.acquire(ctx) {
			return
		}
		defer s.release()

		now := time.Now()
		_ = s.repo.UpdateSessionHeartbeat(ctx, userID, now, now.Add(8*time.Hour))
	}()
}

// RecordSearch almacena los metadatos de una búsqueda realizada en el directorio.
func (s *AnalyticsService) RecordSearch(
	query, specialty, municipality, state string,
	resultsCount int,
	userID *uuid.UUID,
	sessionID, ip string,
) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), analyticsWriteTimeout)
		defer cancel()
		if !s.acquire(ctx) {
			return
		}
		defer s.release()

		_ = s.repo.CreateSearchEvent(ctx, domain.SearchEvent{
			Query:        query,
			Specialty:    specialty,
			Municipality: municipality,
			State:        state,
			ResultsCount: resultsCount,
			UserID:       userID,
			SessionID:    sessionID,
			IP:           ip,
		})
	}()
}

// RecordProfileView rastrea la popularidad individual de los profesionales.
func (s *AnalyticsService) RecordProfileView(psiID uuid.UUID, viewerID *uuid.UUID, sessionID, ip string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), analyticsWriteTimeout)
		defer cancel()
		if !s.acquire(ctx) {
			return
		}
		defer s.release()

		_ = s.repo.CreateProfileView(ctx, domain.ProfileView{
			PsiID:     psiID,
			ViewerID:  viewerID,
			SessionID: sessionID,
			IP:        ip,
		})
	}()
}

// RecordPageView persiste una visita a una página del portal.
func (s *AnalyticsService) RecordPageView(view domain.PageView) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), analyticsWriteTimeout)
		defer cancel()
		if !s.acquire(ctx) {
			return
		}
		defer s.release()

		_ = s.repo.CreatePageView(ctx, view)
	}()
}

// TrackPageView aplica debouncing (una visita por sesión en la ventana) y persiste
// la visita de forma SÍNCRONA bajo el semáforo. La invoca el middleware analytics
// dentro de su propia goroutine, evitando anidamiento de goroutines por request.
func (s *AnalyticsService) TrackPageView(ctx context.Context, view domain.PageView) {
	if !s.acquire(ctx) {
		return
	}
	defer s.release()

	count, _ := s.repo.CountRecentPageViews(ctx, view.SessionID, time.Now().Add(-analyticsVisitWindow))
	if count > 0 {
		return // Ya registrado en esta ventana
	}
	_ = s.repo.CreatePageView(ctx, view)
}

// CountRecentPageViews cuenta las visitas recientes de una sesión (debouncing).
func (s *AnalyticsService) CountRecentPageViews(ctx context.Context, sessionID string, since time.Time) (int64, error) {
	return s.repo.CountRecentPageViews(ctx, sessionID, since)
}

// ── LECTURA / DASHBOARD ──────────────────────────────────────────────────────

// DashboardStats es el Objeto de Proyección (Read Model) maestro.
type DashboardStats = domain.DashboardStats

// GetDashboardStats orquesta la recopilación de todas las métricas en tiempo real.
func (s *AnalyticsService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}

// ── LIMPIEZA Y MANTENIMIENTO ─────────────────────────────────────────────────

// PurgeOldData asegura que el crecimiento de la base de datos sea sostenible.
func (s *AnalyticsService) PurgeOldData(ctx context.Context, olderThanDays int) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	_ = s.repo.DeletePageViewsOlderThan(ctx, cutoff)
	_ = s.repo.DeleteSearchEventsOlderThan(ctx, cutoff)
	_ = s.repo.DeleteProfileViewsOlderThan(ctx, cutoff)
	// LoginEvent se conserva siempre (es auditoría)
}

// CleanExpiredSessions es un recolector de basura para la tabla de sesiones.
func (s *AnalyticsService) CleanExpiredSessions(ctx context.Context) {
	_ = s.repo.DeleteExpiredSessions(ctx, time.Now())
}

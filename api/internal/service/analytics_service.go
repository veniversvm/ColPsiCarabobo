// api/internal/service/analytics_service.go

// Package service implementa la lógica de negocio central del sistema.
//
// Este archivo en particular contiene el motor de Telemetría y Business Intelligence (BI).
// Está diseñado bajo el principio de "Impacto Cero": la recolección de métricas
// nunca debe ralentizar ni interrumpir el flujo principal del usuario final.
package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// AnalyticsService encapsula la lógica de negocio de telemetría,
// delegando la persistencia a un AnalyticsRepository (Dependency Inversion).
type AnalyticsService struct {
	repo domain.AnalyticsRepository
}

// NewAnalyticsService actúa como constructor para inyectar la dependencia del repositorio.
func NewAnalyticsService(repo domain.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// ── ESCRITURA (FIRE-AND-FORGET) ──────────────────────────────────────────────

// RecordLogin registra una huella de auditoría cada vez que un usuario accede al sistema.
func (s *AnalyticsService) RecordLogin(userID uuid.UUID, username, role, ip, userAgent string) {
	go func() {
		s.repo.CreateLoginEvent(domain.LoginEvent{
			UserID:    userID,
			Username:  username,
			Role:      role,
			IP:        ip,
			UserAgent: userAgent,
		})

		now := time.Now()
		s.repo.UpsertActiveSession(domain.ActiveSession{
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
		s.repo.DeleteActiveSession(userID)
	}()
}

// HeartbeatSession actualiza la marca de tiempo (LastSeen) de una sesión.
func (s *AnalyticsService) HeartbeatSession(userID uuid.UUID) {
	go func() {
		now := time.Now()
		s.repo.UpdateSessionHeartbeat(userID, now, now.Add(8*time.Hour))
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
		s.repo.CreateSearchEvent(domain.SearchEvent{
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
		s.repo.CreateProfileView(domain.ProfileView{
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
		s.repo.CreatePageView(view)
	}()
}

// CountRecentPageViews cuenta las visitas recientes de una sesión (debouncing).
func (s *AnalyticsService) CountRecentPageViews(sessionID string, since time.Time) (int64, error) {
	return s.repo.CountRecentPageViews(sessionID, since)
}

// ── LECTURA / DASHBOARD ──────────────────────────────────────────────────────

// DashboardStats es el Objeto de Proyección (Read Model) maestro.
type DashboardStats = domain.DashboardStats

// GetDashboardStats orquesta la recopilación de todas las métricas en tiempo real.
func (s *AnalyticsService) GetDashboardStats() (*DashboardStats, error) {
	return s.repo.GetDashboardStats()
}

// ── LIMPIEZA Y MANTENIMIENTO ─────────────────────────────────────────────────

// PurgeOldData asegura que el crecimiento de la base de datos sea sostenible.
func (s *AnalyticsService) PurgeOldData(olderThanDays int) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	s.repo.DeletePageViewsOlderThan(cutoff)
	s.repo.DeleteSearchEventsOlderThan(cutoff)
	s.repo.DeleteProfileViewsOlderThan(cutoff)
	// LoginEvent se conserva siempre (es auditoría)
}

// CleanExpiredSessions es un recolector de basura para la tabla de sesiones.
func (s *AnalyticsService) CleanExpiredSessions() {
	s.repo.DeleteExpiredSessions(time.Now())
}

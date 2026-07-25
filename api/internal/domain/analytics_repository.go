package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticsRepository define el contrato para persistencia de eventos de telemetría.
// Desacopla la capa de servicio del motor de ORM (GORM), permitiendo testing
// con mocks y cambiando la implementación de persistencia sin afectar el negocio.
type AnalyticsRepository interface {
	// Escritura de eventos
	CreateLoginEvent(event LoginEvent) error
	UpsertActiveSession(session ActiveSession) error
	DeleteActiveSession(userID uuid.UUID) error
	UpdateSessionHeartbeat(userID uuid.UUID, lastSeen, expiresAt time.Time) error
	CreateSearchEvent(event SearchEvent) error
	CreateProfileView(event ProfileView) error
	CreatePageView(view PageView) error
	CountRecentPageViews(sessionID string, since time.Time) (int64, error)

	// Lectura / Dashboard
	GetDashboardStats() (*DashboardStats, error)

	// Mantenimiento
	DeletePageViewsOlderThan(cutoff time.Time) error
	DeleteSearchEventsOlderThan(cutoff time.Time) error
	DeleteProfileViewsOlderThan(cutoff time.Time) error
	DeleteExpiredSessions(now time.Time) error
}

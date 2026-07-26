// Package domain defines the core business models and repository interfaces for the ColPsiCarabobo API.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticsRepository defines the persistence contract for analytics telemetry events.
// It decouples the service layer from the ORM, enabling mock-based testing and
// allowing the storage implementation to change without affecting business logic.
type AnalyticsRepository interface {
	CreateLoginEvent(event LoginEvent) error
	UpsertActiveSession(session ActiveSession) error
	DeleteActiveSession(userID uuid.UUID) error
	UpdateSessionHeartbeat(userID uuid.UUID, lastSeen, expiresAt time.Time) error
	CreateSearchEvent(event SearchEvent) error
	CreateProfileView(event ProfileView) error
	CreatePageView(view PageView) error
	CountRecentPageViews(sessionID string, since time.Time) (int64, error)

	GetDashboardStats() (*DashboardStats, error)

	DeletePageViewsOlderThan(cutoff time.Time) error
	DeleteSearchEventsOlderThan(cutoff time.Time) error
	DeleteProfileViewsOlderThan(cutoff time.Time) error
	DeleteExpiredSessions(now time.Time) error
}

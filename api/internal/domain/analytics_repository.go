// Package domain defines the core business models and repository interfaces for the ColPsiCarabobo API.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AnalyticsRepository defines the persistence contract for analytics telemetry events.
// It decouples the service layer from the ORM, enabling mock-based testing and
// allowing the storage implementation to change without affecting business logic.
//
// Todos los métodos reciben context.Context para propagar timeouts/cancelaciones
// y evitar que una query bloqueada mantenga una conexión a BD colgada.
type AnalyticsRepository interface {
	CreateLoginEvent(ctx context.Context, event LoginEvent) error
	UpsertActiveSession(ctx context.Context, session ActiveSession) error
	DeleteActiveSession(ctx context.Context, userID uuid.UUID) error
	UpdateSessionHeartbeat(ctx context.Context, userID uuid.UUID, lastSeen, expiresAt time.Time) error
	CreateSearchEvent(ctx context.Context, event SearchEvent) error
	CreateProfileView(ctx context.Context, event ProfileView) error
	CreatePageView(ctx context.Context, view PageView) error
	CountRecentPageViews(ctx context.Context, sessionID string, since time.Time) (int64, error)

	GetDashboardStats(ctx context.Context) (*DashboardStats, error)

	DeletePageViewsOlderThan(ctx context.Context, cutoff time.Time) error
	DeleteSearchEventsOlderThan(ctx context.Context, cutoff time.Time) error
	DeleteProfileViewsOlderThan(ctx context.Context, cutoff time.Time) error
	DeleteExpiredSessions(ctx context.Context, now time.Time) error
}

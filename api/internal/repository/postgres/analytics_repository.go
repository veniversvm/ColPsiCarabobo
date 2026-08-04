package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

type analyticsRepo struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository backed by the given database.
func NewAnalyticsRepository(db *gorm.DB) domain.AnalyticsRepository {
	return &analyticsRepo{db: db}
}

// CreateLoginEvent persists a new login event record.
func (r *analyticsRepo) CreateLoginEvent(ctx context.Context, event domain.LoginEvent) error {
	return r.db.WithContext(ctx).Create(&event).Error
}

// UpsertActiveSession creates or updates the active session for a user.
func (r *analyticsRepo) UpsertActiveSession(ctx context.Context, session domain.ActiveSession) error {
	return r.db.WithContext(ctx).Where(domain.ActiveSession{UserID: session.UserID}).
		Assign(session).
		FirstOrCreate(&session).Error
}

// DeleteActiveSession removes the active session record for a user.
func (r *analyticsRepo) DeleteActiveSession(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.ActiveSession{}).Error
}

// UpdateSessionHeartbeat updates the last-seen and expiry timestamps for a session.
func (r *analyticsRepo) UpdateSessionHeartbeat(ctx context.Context, userID uuid.UUID, lastSeen, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.ActiveSession{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"last_seen":  lastSeen,
			"expires_at": expiresAt,
		}).Error
}

// CreateSearchEvent persists a directory search event.
func (r *analyticsRepo) CreateSearchEvent(ctx context.Context, event domain.SearchEvent) error {
	return r.db.WithContext(ctx).Create(&event).Error
}

// CreateProfileView persists a profile view event.
func (r *analyticsRepo) CreateProfileView(ctx context.Context, event domain.ProfileView) error {
	return r.db.WithContext(ctx).Create(&event).Error
}

// CreatePageView persists a page view event.
func (r *analyticsRepo) CreatePageView(ctx context.Context, view domain.PageView) error {
	return r.db.WithContext(ctx).Create(&view).Error
}

// CountRecentPageViews returns the number of page views for a session since the given time.
func (r *analyticsRepo) CountRecentPageViews(ctx context.Context, sessionID string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PageView{}).
		Where("session_id = ? AND created_at >= ?", sessionID, since).
		Count(&count).Error
	return count, err
}

// GetDashboardStats returns aggregated analytics metrics for the admin dashboard.
func (r *analyticsRepo) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)
	monthStart := todayStart.AddDate(0, -1, 0)
	thirtyDays := todayStart.AddDate(0, 0, -30)
	fourteenDays := todayStart.AddDate(0, 0, -14)

	db := r.db.WithContext(ctx)

	// Logins
	db.Model(&domain.LoginEvent{}).Count(&stats.LoginsTotal)
	db.Model(&domain.LoginEvent{}).Where("created_at >= ?", todayStart).Count(&stats.LoginsToday)
	db.Model(&domain.LoginEvent{}).Where("created_at >= ?", weekStart).Count(&stats.LoginsThisWeek)
	db.Model(&domain.LoginEvent{}).Where("created_at >= ?", monthStart).Count(&stats.LoginsThisMonth)
	db.Model(&domain.LoginEvent{}).
		Where("created_at >= ?", todayStart).
		Distinct("user_id").Count(&stats.UniqueUsersToday)

	// Page Views
	db.Model(&domain.PageView{}).Count(&stats.PageViewsTotal)
	db.Model(&domain.PageView{}).Where("created_at >= ?", todayStart).Count(&stats.PageViewsToday)
	db.Model(&domain.PageView{}).Where("created_at >= ?", weekStart).Count(&stats.PageViewsThisWeek)
	db.Model(&domain.PageView{}).
		Where("created_at >= ?", todayStart).
		Distinct("session_id").Count(&stats.UniqueVisitorsToday)
	db.Model(&domain.PageView{}).
		Where("created_at >= ?", weekStart).
		Distinct("session_id").Count(&stats.UniqueVisitorsWeek)

	// Búsquedas
	db.Model(&domain.SearchEvent{}).Count(&stats.SearchesTotal)
	db.Model(&domain.SearchEvent{}).Where("created_at >= ?", todayStart).Count(&stats.SearchesToday)
	db.Model(&domain.SearchEvent{}).Where("created_at >= ?", weekStart).Count(&stats.SearchesThisWeek)

	// Profile Views
	db.Model(&domain.ProfileView{}).Count(&stats.ProfileViewsTotal)
	db.Model(&domain.ProfileView{}).Where("created_at >= ?", todayStart).Count(&stats.ProfileViewsToday)
	db.Model(&domain.ProfileView{}).Where("created_at >= ?", weekStart).Count(&stats.ProfileViewsWeek)

	// Sesiones activas
	db.Model(&domain.ActiveSession{}).
		Where("expires_at > ?", now).
		Count(&stats.ActiveSessionsNow)

	// Top especialidades
	db.Model(&domain.SearchEvent{}).
		Select(`psi_specialty_models.name as value, 
            COUNT(search_events.id) as count,
            psi_specialty_models.name`).
		Joins("left join psi_specialty_models on psi_specialty_models.id = search_events.specialty::int").
		Where("search_events.created_at >= ? AND search_events.specialty != ''", thirtyDays).
		Group("psi_specialty_models.id, psi_specialty_models.name").
		Where("psi_specialty_models.id IS NOT NULL AND psi_specialty_models.id != 0").
		Order("count DESC").
		Limit(11).
		Scan(&stats.TopSpecialties)

	// Top municipios
	db.Model(&domain.SearchEvent{}).
		Select("municipality as value, count(*) as count").
		Where("created_at >= ? AND municipality != ''", thirtyDays).
		Group("municipality").
		Order("count DESC").
		Limit(10).
		Scan(&stats.TopMunicipios)

	// Top términos
	db.Model(&domain.SearchEvent{}).
		Select("query as value, count(*) as count").
		Where("created_at >= ? AND query != ''", thirtyDays).
		Group("query").
		Order("count DESC").
		Limit(10).
		Scan(&stats.TopSearchTerms)

	// Top perfiles
	db.Model(&domain.ProfileView{}).
		Select(`psi_users.id as psi_id, 
                psi_users.first_name as name, 
                psi_users.last_name, 
                COUNT(profile_views.id) as count`).
		Joins("left join psi_users on psi_users.id = profile_views.psi_id").
		Where("psi_users.id IS NOT NULL").
		Group("psi_users.id, psi_users.first_name, psi_users.last_name").
		Order("count DESC").
		Limit(10).
		Scan(&stats.TopProfiles)

	// Tendencia logins
	db.Model(&domain.LoginEvent{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ?", fourteenDays).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&stats.LoginTrend)

	// Tendencia visitas
	db.Model(&domain.PageView{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ?", fourteenDays).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&stats.ViewTrend)

	return stats, nil
}

// DeletePageViewsOlderThan removes page view records older than the cutoff time.
func (r *analyticsRepo) DeletePageViewsOlderThan(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&domain.PageView{}).Error
}

// DeleteSearchEventsOlderThan removes search event records older than the cutoff time.
func (r *analyticsRepo) DeleteSearchEventsOlderThan(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&domain.SearchEvent{}).Error
}

// DeleteProfileViewsOlderThan removes profile view records older than the cutoff time.
func (r *analyticsRepo) DeleteProfileViewsOlderThan(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&domain.ProfileView{}).Error
}

// DeleteExpiredSessions removes active sessions that have expired.
func (r *analyticsRepo) DeleteExpiredSessions(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", now).Delete(&domain.ActiveSession{}).Error
}

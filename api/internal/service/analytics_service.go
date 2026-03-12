// api/internal/service/analytics_service.go

package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// ── ESCRITURA ────────────────────────────────────────────────────────────────

// RecordLogin llama esto desde tu handler de login, tras validar credenciales
func (s *AnalyticsService) RecordLogin(userID uuid.UUID, username, role, ip, userAgent string) {
	go func() {
		// 1. Guardar evento histórico
		s.db.Create(&domain.LoginEvent{
			UserID:    userID,
			Username:  username,
			Role:      role,
			IP:        ip,
			UserAgent: userAgent,
		})

		// 2. Upsert sesión activa (1 por usuario, renueva si ya existe)
		session := domain.ActiveSession{
			UserID:    userID,
			Username:  username,
			Role:      role,
			IP:        ip,
			LastSeen:  time.Now(),
			ExpiresAt: time.Now().Add(8 * time.Hour), // ajusta al TTL de tu JWT
		}
		s.db.Where(domain.ActiveSession{UserID: userID}).
			Assign(session).
			FirstOrCreate(&session)
	}()
}

// RecordLogout invalida la sesión activa
func (s *AnalyticsService) RecordLogout(userID uuid.UUID) {
	go func() {
		s.db.Where("user_id = ?", userID).Delete(&domain.ActiveSession{})
	}()
}

// HeartbeatSession actualiza LastSeen (llámalo en el middleware de auth)
func (s *AnalyticsService) HeartbeatSession(userID uuid.UUID) {
	go func() {
		s.db.Model(&domain.ActiveSession{}).
			Where("user_id = ?", userID).
			Updates(map[string]any{
				"last_seen":  time.Now(),
				"expires_at": time.Now().Add(8 * time.Hour),
			})
	}()
}

// RecordSearch llámalo desde tu handler de búsqueda del directorio
func (s *AnalyticsService) RecordSearch(
	query, specialty, municipality, state string,
	resultsCount int,
	userID *uuid.UUID,
	sessionID, ip string,
) {
	go func() {
		s.db.Create(&domain.SearchEvent{
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

// RecordProfileView llámalo desde tu handler de perfil público
func (s *AnalyticsService) RecordProfileView(psiID uuid.UUID, viewerID *uuid.UUID, sessionID, ip string) {
	go func() {
		s.db.Create(&domain.ProfileView{
			PsiID:     psiID,
			ViewerID:  viewerID,
			SessionID: sessionID,
			IP:        ip,
		})
	}()
}

// ── LECTURA / DASHBOARD ──────────────────────────────────────────────────────

type DashboardStats struct {
	// Logins
	LoginsTotal      int64 `json:"logins_total"`
	LoginsToday      int64 `json:"logins_today"`
	LoginsThisWeek   int64 `json:"logins_this_week"`
	LoginsThisMonth  int64 `json:"logins_this_month"`
	UniqueUsersToday int64 `json:"unique_users_today"`

	// Visitantes del portal
	PageViewsTotal      int64 `json:"page_views_total"`
	PageViewsToday      int64 `json:"page_views_today"`
	PageViewsThisWeek   int64 `json:"page_views_this_week"`
	UniqueVisitorsToday int64 `json:"unique_visitors_today"` // por session_id
	UniqueVisitorsWeek  int64 `json:"unique_visitors_week"`

	// Búsquedas
	SearchesTotal    int64 `json:"searches_total"`
	SearchesToday    int64 `json:"searches_today"`
	SearchesThisWeek int64 `json:"searches_this_week"`

	// Perfiles visitados
	ProfileViewsTotal int64 `json:"profile_views_total"`
	ProfileViewsToday int64 `json:"profile_views_today"`
	ProfileViewsWeek  int64 `json:"profile_views_week"`

	// Usuarios activos AHORA (sesiones no expiradas)
	ActiveSessionsNow int64 `json:"active_sessions_now"`

	// Top búsquedas (últimos 30 días)
	TopSpecialties []TopItem `json:"top_specialties"`
	TopMunicipios  []TopItem `json:"top_municipios"`
	TopSearchTerms []TopItem `json:"top_search_terms"`

	// Top perfiles visitados (últimos 30 días)
	TopProfiles []TopProfile `json:"top_profiles"`

	// Tendencia diaria de logins (últimos 14 días)
	LoginTrend []DailyCount `json:"login_trend"`

	// Tendencia diaria de visitas (últimos 14 días)
	ViewTrend []DailyCount `json:"view_trend"`
}

type TopItem struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
	Name  string `json:"name"`
}

type TopProfile struct {
	PsiID    string `json:"psi_id"`
	Name     string `json:"name"`
	LastName string `json:"last_name"`
	Count    int64  `json:"count"`
}

type DailyCount struct {
	Date  string `json:"date"` // "2025-01-15"
	Count int64  `json:"count"`
}

func (s *AnalyticsService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)
	monthStart := todayStart.AddDate(0, -1, 0)
	thirtyDays := todayStart.AddDate(0, 0, -30)
	fourteenDays := todayStart.AddDate(0, 0, -14)

	db := s.db

	// ── Logins ──────────────────────────────────────────────────────────────
	db.Model(&domain.LoginEvent{}).Count(&stats.LoginsTotal)
	db.Model(&domain.LoginEvent{}).Where("created_at >= ?", todayStart).Count(&stats.LoginsToday)
	db.Model(&domain.LoginEvent{}).Where("created_at >= ?", weekStart).Count(&stats.LoginsThisWeek)
	db.Model(&domain.LoginEvent{}).Where("created_at >= ?", monthStart).Count(&stats.LoginsThisMonth)
	db.Model(&domain.LoginEvent{}).
		Where("created_at >= ?", todayStart).
		Distinct("user_id").Count(&stats.UniqueUsersToday)

	// ── Page Views ───────────────────────────────────────────────────────────
	db.Model(&domain.PageView{}).Count(&stats.PageViewsTotal)
	db.Model(&domain.PageView{}).Where("created_at >= ?", todayStart).Count(&stats.PageViewsToday)
	db.Model(&domain.PageView{}).Where("created_at >= ?", weekStart).Count(&stats.PageViewsThisWeek)
	db.Model(&domain.PageView{}).
		Where("created_at >= ?", todayStart).
		Distinct("session_id").Count(&stats.UniqueVisitorsToday)
	db.Model(&domain.PageView{}).
		Where("created_at >= ?", weekStart).
		Distinct("session_id").Count(&stats.UniqueVisitorsWeek)

	// ── Búsquedas ────────────────────────────────────────────────────────────
	db.Model(&domain.SearchEvent{}).Count(&stats.SearchesTotal)
	db.Model(&domain.SearchEvent{}).Where("created_at >= ?", todayStart).Count(&stats.SearchesToday)
	db.Model(&domain.SearchEvent{}).Where("created_at >= ?", weekStart).Count(&stats.SearchesThisWeek)

	// ── Profile Views ────────────────────────────────────────────────────────
	db.Model(&domain.ProfileView{}).Count(&stats.ProfileViewsTotal)
	db.Model(&domain.ProfileView{}).Where("created_at >= ?", todayStart).Count(&stats.ProfileViewsToday)
	db.Model(&domain.ProfileView{}).Where("created_at >= ?", weekStart).Count(&stats.ProfileViewsWeek)

	// ── Sesiones activas AHORA ───────────────────────────────────────────────
	db.Model(&domain.ActiveSession{}).
		Where("expires_at > ?", now).
		Count(&stats.ActiveSessionsNow)

	// ── Top especialidades buscadas (últimos 30 días) ────────────────────────
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

	// ── Top municipios buscados ───────────────────────────────────────────────
	db.Model(&domain.SearchEvent{}).
		Select("municipality as value, count(*) as count").
		Where("created_at >= ? AND municipality != ''", thirtyDays).
		Group("municipality").
		Order("count DESC").
		Limit(10).
		Scan(&stats.TopMunicipios)

	// ── Top términos de búsqueda libre ───────────────────────────────────────
	db.Model(&domain.SearchEvent{}).
		Select("query as value, count(*) as count").
		Where("created_at >= ? AND query != ''", thirtyDays).
		Group("query").
		Order("count DESC").
		Limit(10).
		Scan(&stats.TopSearchTerms)

	// ── Top perfiles más visitados ───────────────────────────────────────────
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

	// ── Tendencia diaria de logins (últimos 14 días) ─────────────────────────
	db.Model(&domain.LoginEvent{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ?", fourteenDays).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&stats.LoginTrend)

	// ── Tendencia diaria de visitas (últimos 14 días) ────────────────────────
	db.Model(&domain.PageView{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ?", fourteenDays).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&stats.ViewTrend)

	return stats, nil
}

// PurgeOldData limpia registros antiguos para no inflar la BD indefinidamente
// Puedes llamarlo con un cron job mensual
func (s *AnalyticsService) PurgeOldData(olderThanDays int) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	s.db.Where("created_at < ?", cutoff).Delete(&domain.PageView{})
	s.db.Where("created_at < ?", cutoff).Delete(&domain.SearchEvent{})
	s.db.Where("created_at < ?", cutoff).Delete(&domain.ProfileView{})
	// LoginEvent se conserva siempre (es auditoría)
}

func (s *AnalyticsService) CleanExpiredSessions() {
	s.db.Where("expires_at < ?", time.Now()).Delete(&domain.ActiveSession{})
}

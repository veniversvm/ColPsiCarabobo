// api/internal/domain/analytics.go
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// 1. LOGIN EVENT — cada vez que alguien inicia sesión
// ---------------------------------------------------------------------------
type LoginEvent struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	Username  string         `gorm:"size:100"`
	Role      string         `gorm:"size:50"` // "psi" | "admin"
	IP        string         `gorm:"size:45"`
	UserAgent string         `gorm:"size:512"`
	CreatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ---------------------------------------------------------------------------
// 2. PAGE VIEW — cada visita a cualquier ruta (sin login requerido)
// ---------------------------------------------------------------------------
type PageView struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	Path      string     `gorm:"size:512;index"` // "/directorio", "/perfil/123"
	Method    string     `gorm:"size:10"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"` // nil si no está autenticado
	SessionID string     `gorm:"size:64;index"`   // cookie anónima
	IP        string     `gorm:"size:45"`
	Referer   string     `gorm:"size:512"`
	CreatedAt time.Time  `gorm:"index"`
}

// ---------------------------------------------------------------------------
// 3. SEARCH EVENT — cada búsqueda parametrizada en el directorio
// ---------------------------------------------------------------------------
type SearchEvent struct {
	ID uint `gorm:"primaryKey;autoIncrement"`
	// Parámetros de búsqueda que uses en tu directorio
	Query        string     `gorm:"size:255"` // texto libre
	Specialty    string     `gorm:"size:255;index"`
	Municipality string     `gorm:"size:255;index"`
	State        string     `gorm:"size:255;index"`
	ResultsCount int        // cuántos resultados devolvió
	UserID       *uuid.UUID `gorm:"type:uuid;index"`
	SessionID    string     `gorm:"size:64"`
	IP           string     `gorm:"size:45"`
	CreatedAt    time.Time  `gorm:"index"`
}

// ---------------------------------------------------------------------------
// 4. PROFILE VIEW — cada vez que se visita el perfil de un psicólogo
// ---------------------------------------------------------------------------
type ProfileView struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	PsiID     uuid.UUID  `gorm:"type:uuid;not null;index"` // perfil visto
	ViewerID  *uuid.UUID `gorm:"type:uuid;index"`          // nil si anónimo
	SessionID string     `gorm:"size:64"`
	IP        string     `gorm:"size:45"`
	CreatedAt time.Time  `gorm:"index"`
}

// ---------------------------------------------------------------------------
//  5. ACTIVE SESSION — sesiones activas en este momento
//     Se inserta en login, se actualiza con heartbeat, se marca expired en logout
//
// ---------------------------------------------------------------------------
type ActiveSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex"` // 1 sesión por usuario
	Username  string    `gorm:"size:100"`
	Role      string    `gorm:"size:50"`
	IP        string    `gorm:"size:45"`
	LastSeen  time.Time `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

// IsActive devuelve true si la sesión no ha expirado
func (s *ActiveSession) IsActive() bool {
	return time.Now().Before(s.ExpiresAt)
}

// ── Dashboard Stats (Read Model para BI) ─────────────────────────────────────

type DashboardStats struct {
	LoginsTotal         int64       `json:"logins_total"`
	LoginsToday         int64       `json:"logins_today"`
	LoginsThisWeek      int64       `json:"logins_this_week"`
	LoginsThisMonth     int64       `json:"logins_this_month"`
	UniqueUsersToday    int64       `json:"unique_users_today"`
	PageViewsTotal      int64       `json:"page_views_total"`
	PageViewsToday      int64       `json:"page_views_today"`
	PageViewsThisWeek   int64       `json:"page_views_this_week"`
	UniqueVisitorsToday int64       `json:"unique_users_today_page"`
	UniqueVisitorsWeek  int64       `json:"unique_visitors_week"`
	SearchesTotal       int64       `json:"searches_total"`
	SearchesToday       int64       `json:"searches_today"`
	SearchesThisWeek    int64       `json:"searches_this_week"`
	ProfileViewsTotal   int64       `json:"profile_views_total"`
	ProfileViewsToday   int64       `json:"profile_views_today"`
	ProfileViewsWeek    int64       `json:"profile_views_week"`
	ActiveSessionsNow   int64       `json:"active_sessions_now"`
	TopSpecialties      []TopItem   `json:"top_specialties"`
	TopMunicipios       []TopItem   `json:"top_municipios"`
	TopSearchTerms      []TopItem   `json:"top_search_terms"`
	TopProfiles         []TopProfile `json:"top_profiles"`
	LoginTrend          []DailyCount `json:"login_trend"`
	ViewTrend           []DailyCount `json:"view_trend"`
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
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

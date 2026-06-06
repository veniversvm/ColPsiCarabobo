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
	"gorm.io/gorm"
)

// AnalyticsService encapsula la conexión a la base de datos para la ingesta y
// lectura de eventos de telemetría.
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService actúa como constructor para inyectar la dependencia de GORM.
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// ── ESCRITURA (FIRE-AND-FORGET) ──────────────────────────────────────────────
// Todos los métodos de escritura lanzan una Goroutine anónima (go func()).
// Esto significa que el hilo principal HTTP retorna inmediatamente al cliente,
// mientras que la base de datos escribe el evento en segundo plano.

// RecordLogin registra una huella de auditoría cada vez que un usuario accede al sistema.
// Se invoca desde el handler de login, inmediatamente después de validar credenciales.
func (s *AnalyticsService) RecordLogin(userID uuid.UUID, username, role, ip, userAgent string) {
	go func() {
		// 1. Guardar evento histórico (Auditoría Inmutable)
		// Registra el "Quién, Qué, Dónde y Cuándo" para análisis forense de seguridad.
		s.db.Create(&domain.LoginEvent{
			UserID:    userID,
			Username:  username,
			Role:      role,
			IP:        ip,
			UserAgent: userAgent,
		})

		// 2. Upsert de sesión activa (Manejo de Estado)
		// Garantiza una relación 1:1 entre el usuario y su estado de conectividad.
		// Si el usuario ya tenía una sesión activa, se renuevan sus datos (Assign)
		// en lugar de crear un duplicado, previniendo inconsistencias en el Dashboard.
		session := domain.ActiveSession{
			UserID:    userID,
			Username:  username,
			Role:      role,
			IP:        ip,
			LastSeen:  time.Now(),
			ExpiresAt: time.Now().Add(8 * time.Hour), // Ajustado al TTL de expiración del JWT
		}
		s.db.Where(domain.ActiveSession{UserID: userID}).
			Assign(session).
			FirstOrCreate(&session)
	}()
}

// RecordLogout destruye la estampa de sesión activa del usuario.
// Mantiene el panel de "Usuarios Activos" preciso en tiempo real.
func (s *AnalyticsService) RecordLogout(userID uuid.UUID) {
	go func() {
		s.db.Where("user_id = ?", userID).Delete(&domain.ActiveSession{})
	}()
}

// HeartbeatSession actualiza la marca de tiempo (LastSeen) de una sesión.
// Diseñado para ser invocado desde el middleware de autenticación (por cada request),
// extiende dinámicamente la vida útil de la sesión mientras el usuario esté navegando.
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

// RecordSearch almacena los metadatos de una búsqueda realizada en el directorio.
// Inteligencia de Negocio: Permite descubrir qué especialidades o zonas geográficas
// tienen más demanda en el portal, ayudando al Colegio a tomar decisiones gremiales.
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

// RecordProfileView rastrea la popularidad individual de los profesionales.
// Permite calcular qué perfiles atraen más tráfico y quién los está viendo.
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

// ── LECTURA / DASHBOARD (DATA AGGREGATION) ───────────────────────────────────

// DashboardStats es el Objeto de Proyección (Read Model) maestro.
// Agrupa todas las métricas procesadas para ser consumidas en una sola llamada de red
// por el panel administrativo del frontend.
type DashboardStats struct {
	// Logins (Actividad del Sistema)
	LoginsTotal      int64 `json:"logins_total"`
	LoginsToday      int64 `json:"logins_today"`
	LoginsThisWeek   int64 `json:"logins_this_week"`
	LoginsThisMonth  int64 `json:"logins_this_month"`
	UniqueUsersToday int64 `json:"unique_users_today"`

	// Visitantes del portal (Tráfico Web)
	PageViewsTotal      int64 `json:"page_views_total"`
	PageViewsToday      int64 `json:"page_views_today"`
	PageViewsThisWeek   int64 `json:"page_views_this_week"`
	UniqueVisitorsToday int64 `json:"unique_visitors_today"` // Agrupado por session_id (Cookie)
	UniqueVisitorsWeek  int64 `json:"unique_visitors_week"`

	// Búsquedas (Interacción de Usuarios)
	SearchesTotal    int64 `json:"searches_total"`
	SearchesToday    int64 `json:"searches_today"`
	SearchesThisWeek int64 `json:"searches_this_week"`

	// Perfiles visitados (Engagement)
	ProfileViewsTotal int64 `json:"profile_views_total"`
	ProfileViewsToday int64 `json:"profile_views_today"`
	ProfileViewsWeek  int64 `json:"profile_views_week"`

	// Usuarios activos AHORA (sesiones no expiradas, concurrencia en vivo)
	ActiveSessionsNow int64 `json:"active_sessions_now"`

	// Top búsquedas (Ranking de Intereses de los últimos 30 días)
	TopSpecialties []TopItem `json:"top_specialties"`
	TopMunicipios  []TopItem `json:"top_municipios"`
	TopSearchTerms []TopItem `json:"top_search_terms"`

	// Top perfiles visitados (Ranking de Popularidad de los últimos 30 días)
	TopProfiles []TopProfile `json:"top_profiles"`

	// Gráficos de Tendencia (Time-Series Data para renderizar gráficos de líneas/barras)
	LoginTrend []DailyCount `json:"login_trend"`
	ViewTrend  []DailyCount `json:"view_trend"`
}

// Structs auxiliares para el tipado fuerte de las consultas de agregación SQL.
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
	Date  string `json:"date"` // Formato garantizado por SQL DATE(): "YYYY-MM-DD"
	Count int64  `json:"count"`
}

// GetDashboardStats orquesta la recopilación de todas las métricas en tiempo real.
//
// Optimización de BD: Se delegan las sumatorias (COUNT) y agrupaciones (GROUP BY)
// al motor SQL nativo (PostgreSQL), el cual es infinitamente más rápido para matemáticas
// de conjuntos que extraer los registros y contarlos en la memoria de Go.
func (s *AnalyticsService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Bounding Boxes de Tiempo (Ventanas de Filtrado)
	// Se calculan las medianoches exactas para garantizar que "Today" significa
	// desde las 00:00:00 y no "hace 24 horas exactas".
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
	// UniqueUsersToday usa DISTINCT para no contar múltiples logins de la misma persona en el mismo día.
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
	// Métrica en tiempo real (Concurrencia). Útil para evaluar carga del servidor.
	db.Model(&domain.ActiveSession{}).
		Where("expires_at > ?", now).
		Count(&stats.ActiveSessionsNow)

	// ── Top especialidades buscadas (últimos 30 días) ────────────────────────
	// Cruza eventos de búsqueda con la tabla maestra de especialidades (JOIN)
	// para resolver el nombre legible de la especialidad buscada por ID.
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
	// Cruza las visitas con la tabla de Psicólogos para obtener sus nombres reales.
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
	// Extrae la fecha sin horas mediante DATE() para agrupar los conteos por día calendario.
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

// ── LIMPIEZA Y MANTENIMIENTO (DATA RETENTION POLICIES) ───────────────────────

// PurgeOldData asegura que el crecimiento de la base de datos sea sostenible.
// Diseñado para invocarse mediante un Job programado (Cron Job).
//
// Diferenciación Crítica de Datos:
// Elimina métricas efímeras (navegación, búsquedas) más antiguas de X días, pero
// mantiene INTACTA la tabla LoginEvent, ya que los logs de autenticación son
// requisitos de auditoría de ciberseguridad a largo plazo.
func (s *AnalyticsService) PurgeOldData(olderThanDays int) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	s.db.Where("created_at < ?", cutoff).Delete(&domain.PageView{})
	s.db.Where("created_at < ?", cutoff).Delete(&domain.SearchEvent{})
	s.db.Where("created_at < ?", cutoff).Delete(&domain.ProfileView{})
	// LoginEvent se conserva siempre (es auditoría)
}

// CleanExpiredSessions es un recolector de basura (Garbage Collector) para la tabla de sesiones.
// Destruye registros de sesiones que superaron su TTL, evitando bloqueos (Locks) y
// consultas pesadas sobre sesiones que de todas formas ya caducaron.
func (s *AnalyticsService) CleanExpiredSessions() {
	s.db.Where("expires_at < ?", time.Now()).Delete(&domain.ActiveSession{})
}

package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupAnalyticsTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=colpsi_test sslmode=disable"
	}

	adminDSN := "host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable"
	tmpDb, _ := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
	err = db.AutoMigrate(
		&domain.LoginEvent{},
		&domain.PageView{},
		&domain.SearchEvent{},
		&domain.ProfileView{},
		&domain.ActiveSession{},
		&domain.PsiUserModel{},
		&domain.PsiSpecialtyModel{},
		&domain.TextModel{},
		&domain.PsiUserColData{},
		&domain.PsiUserPostGrade{},
		&domain.PsiUserSocialNetwork{},
	)
	require.NoError(t, err)

	return db
}

func TestAnalyticsRepo_ComprehensiveSuite(t *testing.T) {
	mainDB := setupAnalyticsTestDB(t)

	// Cleanup before tests
	mainDB.Exec("DELETE FROM active_sessions")
	mainDB.Exec("DELETE FROM login_events")
	mainDB.Exec("DELETE FROM page_views")
	mainDB.Exec("DELETE FROM search_events")
	mainDB.Exec("DELETE FROM profile_views")

	t.Run("CreateLoginEvent", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		event := domain.LoginEvent{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Username:  "test_user",
			Role:      "admin",
			IP:        "127.0.0.1",
			UserAgent: "Mozilla/5.0",
		}

		err := r.CreateLoginEvent(context.Background(), event)
		require.NoError(t, err)

		var count int64
		tx.Model(&domain.LoginEvent{}).Where("id = ?", event.ID).Count(&count)
		require.Equal(t, int64(1), count)
	})

	t.Run("UpsertActiveSession_Creates and Updates", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		userID := uuid.New()
		now := time.Now()

		session := domain.ActiveSession{
			UserID:    userID,
			Username:  "session_user",
			Role:      "psi",
			IP:        "192.168.1.1",
			LastSeen:  now,
			ExpiresAt: now.Add(30 * time.Minute),
		}

		err := r.UpsertActiveSession(context.Background(), session)
		require.NoError(t, err)

		var found domain.ActiveSession
		err = tx.Where("user_id = ?", userID).First(&found).Error
		require.NoError(t, err)
		require.Equal(t, "session_user", found.Username)

		// Upsert again — should update, not create duplicate
		session.Username = "session_user_updated"
		err = r.UpsertActiveSession(context.Background(), session)
		require.NoError(t, err)

		var count int64
		tx.Model(&domain.ActiveSession{}).Where("user_id = ?", userID).Count(&count)
		require.Equal(t, int64(1), count, "Upsert must not create duplicates")

		tx.Where("user_id = ?", userID).First(&found)
		require.Equal(t, "session_user_updated", found.Username)
	})

	t.Run("DeleteActiveSession", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		userID := uuid.New()
		now := time.Now()
		tx.Create(&domain.ActiveSession{
			UserID:    userID,
			Username:  "to_delete",
			Role:      "psi",
			IP:        "127.0.0.1",
			LastSeen:  now,
			ExpiresAt: now.Add(30 * time.Minute),
		})

		err := r.DeleteActiveSession(context.Background(), userID)
		require.NoError(t, err)

		var count int64
		tx.Model(&domain.ActiveSession{}).Where("user_id = ?", userID).Count(&count)
		require.Equal(t, int64(0), count)
	})

	t.Run("UpdateSessionHeartbeat", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		userID := uuid.New()
		now := time.Now()
		tx.Create(&domain.ActiveSession{
			UserID:    userID,
			Username:  "hb_user",
			Role:      "admin",
			IP:        "10.0.0.1",
			LastSeen:  now,
			ExpiresAt: now.Add(10 * time.Minute),
		})

		newLastSeen := now.Add(5 * time.Minute)
		newExpiry := now.Add(35 * time.Minute)

		err := r.UpdateSessionHeartbeat(context.Background(), userID, newLastSeen, newExpiry)
		require.NoError(t, err)

		var found domain.ActiveSession
		tx.Where("user_id = ?", userID).First(&found)
		require.WithinDuration(t, newLastSeen, found.LastSeen, time.Second)
		require.WithinDuration(t, newExpiry, found.ExpiresAt, time.Second)
	})

	t.Run("CreatePageView and CountRecentPageViews", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		sessionID := "sess_abc123"
		now := time.Now()

		err := r.CreatePageView(context.Background(), domain.PageView{
			Path:      "/directorio",
			Method:    "GET",
			SessionID: sessionID,
			IPHash:    "127.0.0.1",
			CreatedAt: now,
		})
		require.NoError(t, err)

		err = r.CreatePageView(context.Background(), domain.PageView{
			Path:      "/perfil/123",
			Method:    "GET",
			SessionID: sessionID,
			IPHash:    "127.0.0.1",
			CreatedAt: now,
		})
		require.NoError(t, err)

		count, err := r.CountRecentPageViews(context.Background(), sessionID, now.Add(-1*time.Minute))
		require.NoError(t, err)
		require.Equal(t, int64(2), count)

		count, err = r.CountRecentPageViews(context.Background(), sessionID, now.Add(1*time.Hour))
		require.NoError(t, err)
		require.Equal(t, int64(0), count, "Should not count views after the since time")
	})

	t.Run("CreateSearchEvent", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		event := domain.SearchEvent{
			Query:        "clínica",
			Specialty:    "1",
			Municipality: "Valencia",
			State:        "Carabobo",
			ResultsCount: 5,
			SessionID:    "sess_xyz",
			IPHash:       "127.0.0.1",
		}

		err := r.CreateSearchEvent(context.Background(), event)
		require.NoError(t, err)

		var count int64
		tx.Model(&domain.SearchEvent{}).Where("query = ?", "clínica").Count(&count)
		require.Equal(t, int64(1), count)
	})

	t.Run("CreateProfileView", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		psiID := uuid.New()
		event := domain.ProfileView{
			PsiID:     psiID,
			SessionID: "sess_pv",
			IPHash:    "10.0.0.1",
		}

		err := r.CreateProfileView(context.Background(), event)
		require.NoError(t, err)

		var count int64
		tx.Model(&domain.ProfileView{}).Where("psi_id = ?", psiID).Count(&count)
		require.Equal(t, int64(1), count)
	})

	t.Run("GetDashboardStats_EmptyDB", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		stats, err := r.GetDashboardStats(context.Background())
		require.NoError(t, err)
		require.NotNil(t, stats)
		require.Equal(t, int64(0), stats.LoginsTotal)
		require.Equal(t, int64(0), stats.PageViewsTotal)
		require.Equal(t, int64(0), stats.SearchesTotal)
		require.Equal(t, int64(0), stats.ActiveSessionsNow)
	})

	t.Run("GetDashboardStats_WithData", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		now := time.Now()
		userID := uuid.New()

		// Create a psi user for profile views top
		bio := domain.TextModel{ID: uuid.New(), Content: "bio"}
		tx.Create(&bio)
		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), CI: 9999, FPV: 9999, BornDate: now, Genre: "M",
			Nationality: "V", ContactEmail: "psi@t.com", ContactPhone: "123",
			FirstName: "Test", LastName: "User", BioTextID: bio.ID,
			Credentials: domain.Credentials{Username: "top_user", Email: "psi@t.com", IsActive: true},
		})

		// Seed login events
		tx.Create(&domain.LoginEvent{ID: uuid.New(), UserID: userID, Username: "u1", Role: "admin", CreatedAt: now})
		tx.Create(&domain.LoginEvent{ID: uuid.New(), UserID: uuid.New(), Username: "u2", Role: "psi", CreatedAt: now})

		// Seed page views
		tx.Create(&domain.PageView{Path: "/home", Method: "GET", SessionID: "s1", IPHash: "127.0.0.1", CreatedAt: now})

		// Seed search events
		tx.Create(&domain.SearchEvent{Query: "test", Specialty: "1", Municipality: "Valencia", ResultsCount: 3, SessionID: "s1", CreatedAt: now})

		// Seed active session
		tx.Create(&domain.ActiveSession{
			UserID: userID, Username: "u1", Role: "admin",
			IP: "127.0.0.1", LastSeen: now, ExpiresAt: now.Add(10 * time.Minute),
		})

		stats, err := r.GetDashboardStats(context.Background())
		require.NoError(t, err)
		require.Equal(t, int64(2), stats.LoginsTotal)
		require.Equal(t, int64(2), stats.LoginsToday)
		require.Equal(t, int64(1), stats.PageViewsTotal)
		require.Equal(t, int64(1), stats.SearchesTotal)
		require.Equal(t, int64(1), stats.ActiveSessionsNow)
	})

	t.Run("DeleteExpiredSessions", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		now := time.Now()
		expired := now.Add(-1 * time.Hour)
		future := now.Add(1 * time.Hour)

		tx.Create(&domain.ActiveSession{UserID: uuid.New(), Username: "expired1", Role: "psi", IP: "127.0.0.1", LastSeen: expired, ExpiresAt: expired})
		tx.Create(&domain.ActiveSession{UserID: uuid.New(), Username: "expired2", Role: "psi", IP: "127.0.0.1", LastSeen: expired, ExpiresAt: expired})
		tx.Create(&domain.ActiveSession{UserID: uuid.New(), Username: "active1", Role: "admin", IP: "127.0.0.1", LastSeen: now, ExpiresAt: future})

		err := r.DeleteExpiredSessions(context.Background(), now)
		require.NoError(t, err)

		var count int64
		tx.Model(&domain.ActiveSession{}).Count(&count)
		require.Equal(t, int64(1), count, "Only the active session should remain")
	})

	t.Run("DeleteEventsOlderThan", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAnalyticsRepository(tx)

		old := time.Now().Add(-48 * time.Hour)
		recent := time.Now()

		tx.Create(&domain.PageView{Path: "/old", Method: "GET", SessionID: "s1", CreatedAt: old})
		tx.Create(&domain.PageView{Path: "/recent", Method: "GET", SessionID: "s2", CreatedAt: recent})
		tx.Create(&domain.SearchEvent{Query: "old", SessionID: "s1", CreatedAt: old})
		tx.Create(&domain.SearchEvent{Query: "recent", SessionID: "s2", CreatedAt: recent})
		tx.Create(&domain.ProfileView{PsiID: uuid.New(), SessionID: "s1", CreatedAt: old})
		tx.Create(&domain.ProfileView{PsiID: uuid.New(), SessionID: "s2", CreatedAt: recent})

		err := r.DeletePageViewsOlderThan(context.Background(), time.Now().Add(-24 * time.Hour))
		require.NoError(t, err)
		var pvCount int64
		tx.Model(&domain.PageView{}).Count(&pvCount)
		require.Equal(t, int64(1), pvCount)

		err = r.DeleteSearchEventsOlderThan(context.Background(), time.Now().Add(-24 * time.Hour))
		require.NoError(t, err)
		var seCount int64
		tx.Model(&domain.SearchEvent{}).Count(&seCount)
		require.Equal(t, int64(1), seCount)

		err = r.DeleteProfileViewsOlderThan(context.Background(), time.Now().Add(-24 * time.Hour))
		require.NoError(t, err)
		var proCount int64
		tx.Model(&domain.ProfileView{}).Count(&proCount)
		require.Equal(t, int64(1), proCount)
	})
}

// TestAnalyticsRepo_ColumnasAguantanLoQueMandaElServicio es el contrato entre el
// servicio y el esquema.
//
// El 22001 que tumbó la telemetría entera no lo produjo un dato raro: lo produjo el
// dato routineño. El servicio convirtió la IP en una huella SHA-256 (64 hex) y la
// columna se quedó en varchar(45), el largo de una IPv6, así que TODOS los inserts
// de las tres tablas de visitantes dejaron de funcionar. Aquí el punto es que el
// valor que el servicio manda hoy entre en la columna: es el mismo control, pero
// con la base real en vez de un mock, que es donde este tipo de desajuste se
// esconden (los tests de servicio usan mocks, por eso no lo detectaron).
//
// OJO el ancho de esta base lo construye AutoMigrate desde el tag del struct, no
// desde migrations/: este test fija el contrato con la base, y el contrato con el
// archivo de migración lo fija TestAnalyticsRepo_AnchosCoherentesConLaMigracion.
func TestAnalyticsRepo_ColumnasAguantanLoQueMandaElServicio(t *testing.T) {
	db := setupAnalyticsTestDB(t)

	// Una huella real del servicio: 64 hex, ni uno más ni uno menos.
	const huella = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	// Un _sid de cookie _sid real: UUIDv7, 36 caracteres.
	const sid = "0192f8a1-2b3c-7d4e-8f90-a1b2c3d4e5f6"
	const origen = "https://colegio-psicologos-carabobo.com"

	t.Run("PageView entra completo", func(t *testing.T) {
		r := NewAnalyticsRepository(db)
		require.NoError(t, r.CreatePageView(context.Background(), domain.PageView{
			Path:      "/directorio",
			Method:    "GET",
			SessionID: sid,
			IPHash:    huella,
			Referer:   origen,
			CreatedAt: time.Now(),
		}))
	})

	t.Run("SearchEvent entra completo", func(t *testing.T) {
		r := NewAnalyticsRepository(db)
		require.NoError(t, r.CreateSearchEvent(context.Background(), domain.SearchEvent{
			Query:        "depresión",
			Specialty:    "psicología clínica",
			Municipality: "Valencia",
			State:        "Carabobo",
			ResultsCount: 12,
			SessionID:    sid,
			IPHash:       huella,
			CreatedAt:    time.Now(),
		}))
	})

	t.Run("ProfileView entra completo", func(t *testing.T) {
		r := NewAnalyticsRepository(db)
		require.NoError(t, r.CreateProfileView(context.Background(), domain.ProfileView{
			PsiID:     uuid.New(),
			SessionID: sid,
			IPHash:    huella,
			CreatedAt: time.Now(),
		}))
	})

	t.Run("LoginEvent conserva la IP en claro y entra en varchar(45)", func(t *testing.T) {
		// La excepción declarada: esta tabla es bitácora de seguridad y guarda la
		// dirección, no la huella. Si algún día la hashean, el §10.3 de los
		// Términos deja de ser cierto y este test avisa.
		r := NewAnalyticsRepository(db)
		require.NoError(t, r.CreateLoginEvent(context.Background(), domain.LoginEvent{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Username:  "admin1",
			Role:      "admin",
			IP:        "2001:db8::1", // IPv6 completa: el peor caso de los 45
			UserAgent: "Mozilla/5.0",
			CreatedAt: time.Now(),
		}))
	})

	t.Run("la huella se guarda completa, no truncada", func(t *testing.T) {
		// El síntoma del bug era un 22001, pero el riesgo espejo es peor: que la
		// columna "acepte" la huella y la recorte. Se comprueba que leyéndola
		// vuelve entera, porque los 64 hex son lo que permite contar visitantes
		// distintos.
		var guardada string
		require.NoError(t, db.Raw("SELECT ip_hash FROM page_views WHERE ip_hash = ?", huella).Scan(&guardada).Error)
		require.Equal(t, huella, guardada)
		require.Len(t, guardada, 64)
	})
}

// TestAnalyticsRepo_AnchosCoherentesConLaMigracion ata la base de pruebas al
// archivo de migración. AutoMigrate construye los anchos desde los tags del struct,
// así que un tag correcto con una migración que no existe (o al revés) pasaría este
// suite entero y llegaría a producción como el 22001 del 30-sep.
func TestAnalyticsRepo_AnchosCoherentesConLaMigracion(t *testing.T) {
	db := setupAnalyticsTestDB(t)

	esperados := []struct {
		tabla   string
		columna string
		ancho   int
		porque  string
	}{
		{"page_views", "ip_hash", 64, "hex de SHA-256 completo (20261001100000)"},
		{"search_events", "ip_hash", 64, "hex de SHA-256 completo (20261001100000)"},
		{"profile_views", "ip_hash", 64, "hex de SHA-256 completo (20261001100000)"},
		{"login_events", "ip", 45, "IP en claro, bitácora de seguridad"},
		{"active_sessions", "ip", 45, "IP en claro, bitácora de seguridad"},
		{"page_views", "path", 512, "URI de la petición"},
		{"page_views", "referer", 512, "origen reducido, no la URL completa"},
		{"page_views", "method", 10, "método HTTP"},
		{"search_events", "query", 255, "texto que teclea el visitante"},
		{"search_events", "session_id", 64, "cookie _sid"},
	}

	for _, e := range esperados {
		var ancho int
		err := db.Raw(
			"SELECT character_maximum_length FROM information_schema.columns WHERE table_name = ? AND column_name = ?",
			e.tabla, e.columna,
		).Scan(&ancho).Error
		require.NoError(t, err, "columna %s.%s", e.tabla, e.columna)
		require.Equal(t, e.ancho, ancho, "%s.%s debería ser varchar(%d): %s", e.tabla, e.columna, e.ancho, e.porque)
	}
}

// TestAnalyticsRepo_HashSQLDelBackfillEquivaleAFingerprint fija la fórmula del
// backfill documentado en 20261001100000_analytics_ip_hash.sql: si el SQL y Go no
// producen el mismo hex, las filas históricas quedan con huellas que no
// corresponden a nada y un visitante de antes y después del corte cuenta como dos.
//
// Se ejecuta la EXPRESIÓN LITERAL de la migración, no una con placeholders, a
// propósito: la primera versión de ese SQL usaba
// `sha256('<sal>' || '|' || ip)` y fallaba con
// `42883 function sha256(text) does not exist` (la concatenación de literales
// resuelve a text, y sha256() solo existe para bytea). Con parámetros el error es
// el mismo, así que un test con placeholders tampoco lo habría delatado — lo
// delata tener la expresión escrita como el operador la va a ejecutar.
//
// El filtro `~ '[.:]'` (IPv4 tiene punto, IPv6 dos puntos) es lo que hace el
// UPDATE idempotente; se comprueba con casos reales y con el caso límite de una
// huella ya hasheada, que NO debe volver a pasar por el hash.
func TestAnalyticsRepo_HashSQLDelBackfillEquivaleAFingerprint(t *testing.T) {
	db := setupAnalyticsTestDB(t)

	svc := service.NewAnalyticsService(nil)
	// resolveIPSalt cae a la constante por defecto cuando ANALYTICS_IP_SALT no
	// está; el test usa esa misma sal explícita. Si algún día el test corre con
	// la variable definida, el servicio usaría OTRA sal y esta comparación
	// mentiría, así que se aborta en vez de dar un verde falso.
	if os.Getenv("ANALYTICS_IP_SALT") != "" {
		t.Skip("ANALYTICS_IP_SALT está definida: este test verifica la fórmula, no la sal del entorno")
	}
	const sal = "colpsi-analytics-ipsalt-v1"

	ips := []string{
		"190.52.130.45",
		"127.0.0.1",
		"2001:db8::1",
		"::1",
		"172.17.0.1",
	}

	t.Run("el SQL produce el mismo hex que el servicio", func(t *testing.T) {
		for _, ip := range ips {
			var desdeSQL string
			sql := fmt.Sprintf(
				"SELECT encode(sha256(convert_to('%s' || '|' || '%s', 'UTF8')), 'hex')",
				strings.ReplaceAll(sal, "'", "''"), strings.ReplaceAll(ip, "'", "''"),
			)
			require.NoError(t, db.Raw(sql).Scan(&desdeSQL).Error)
			require.Equal(t, svc.FingerprintIP(ip), desdeSQL, "divergen para la IP %q", ip)
			require.Len(t, desdeSQL, 64)
		}
	})

	t.Run("el filtro selecciona IPs en claro y NO huellas", func(t *testing.T) {
		for _, ip := range ips {
			var coincide bool
			require.NoError(t, db.Raw("SELECT ? ~ '[.:]'", ip).Scan(&coincide).Error)
			require.True(t, coincide, "una IP en claro debe entrar al backfill: %q", ip)
		}

		// La huella ya hasheada no tiene punto ni dos puntos: una segunda pasada
		// del backfill no debe volver a hashearla.
		for _, ip := range ips {
			huella := svc.FingerprintIP(ip)
			var coincide bool
			require.NoError(t, db.Raw("SELECT ? ~ '[.:]'", huella).Scan(&coincide).Error)
			require.False(t, coincide, "una huella no debe volver a pasar por el hash: %q", huella)
		}
	})
}

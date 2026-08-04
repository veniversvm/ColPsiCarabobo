package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
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
			IP:        "127.0.0.1",
			CreatedAt: now,
		})
		require.NoError(t, err)

		err = r.CreatePageView(context.Background(), domain.PageView{
			Path:      "/perfil/123",
			Method:    "GET",
			SessionID: sessionID,
			IP:        "127.0.0.1",
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
			IP:           "127.0.0.1",
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
			IP:        "10.0.0.1",
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
		tx.Create(&domain.PageView{Path: "/home", Method: "GET", SessionID: "s1", IP: "127.0.0.1", CreatedAt: now})

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

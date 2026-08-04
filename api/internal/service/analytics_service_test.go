package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// =========================================================================
// MOCK: AnalyticsRepository
// =========================================================================

type mockAnalyticsRepo struct {
	domain.AnalyticsRepository

	CreateLoginEventFunc            func(ctx context.Context, event domain.LoginEvent) error
	UpsertActiveSessionFunc         func(ctx context.Context, session domain.ActiveSession) error
	DeleteActiveSessionFunc         func(ctx context.Context, userID uuid.UUID) error
	UpdateSessionHeartbeatFunc      func(ctx context.Context, userID uuid.UUID, lastSeen, expiresAt time.Time) error
	CreateSearchEventFunc           func(ctx context.Context, event domain.SearchEvent) error
	CreateProfileViewFunc           func(ctx context.Context, event domain.ProfileView) error
	CreatePageViewFunc              func(ctx context.Context, view domain.PageView) error
	CountRecentPageViewsFunc        func(ctx context.Context, sessionID string, since time.Time) (int64, error)
	GetDashboardStatsFunc           func(ctx context.Context) (*domain.DashboardStats, error)
	DeletePageViewsOlderThanFunc    func(ctx context.Context, cutoff time.Time) error
	DeleteSearchEventsOlderThanFunc func(ctx context.Context, cutoff time.Time) error
	DeleteProfileViewsOlderThanFunc func(ctx context.Context, cutoff time.Time) error
	DeleteExpiredSessionsFunc       func(ctx context.Context, now time.Time) error
}

func (m *mockAnalyticsRepo) CreateLoginEvent(ctx context.Context, event domain.LoginEvent) error {
	if m.CreateLoginEventFunc != nil {
		return m.CreateLoginEventFunc(ctx, event)
	}
	return nil
}

func (m *mockAnalyticsRepo) UpsertActiveSession(ctx context.Context, session domain.ActiveSession) error {
	if m.UpsertActiveSessionFunc != nil {
		return m.UpsertActiveSessionFunc(ctx, session)
	}
	return nil
}

func (m *mockAnalyticsRepo) DeleteActiveSession(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteActiveSessionFunc != nil {
		return m.DeleteActiveSessionFunc(ctx, userID)
	}
	return nil
}

func (m *mockAnalyticsRepo) UpdateSessionHeartbeat(ctx context.Context, userID uuid.UUID, lastSeen, expiresAt time.Time) error {
	if m.UpdateSessionHeartbeatFunc != nil {
		return m.UpdateSessionHeartbeatFunc(ctx, userID, lastSeen, expiresAt)
	}
	return nil
}

func (m *mockAnalyticsRepo) CreateSearchEvent(ctx context.Context, event domain.SearchEvent) error {
	if m.CreateSearchEventFunc != nil {
		return m.CreateSearchEventFunc(ctx, event)
	}
	return nil
}

func (m *mockAnalyticsRepo) CreateProfileView(ctx context.Context, event domain.ProfileView) error {
	if m.CreateProfileViewFunc != nil {
		return m.CreateProfileViewFunc(ctx, event)
	}
	return nil
}

func (m *mockAnalyticsRepo) CreatePageView(ctx context.Context, view domain.PageView) error {
	if m.CreatePageViewFunc != nil {
		return m.CreatePageViewFunc(ctx, view)
	}
	return nil
}

func (m *mockAnalyticsRepo) CountRecentPageViews(ctx context.Context, sessionID string, since time.Time) (int64, error) {
	if m.CountRecentPageViewsFunc != nil {
		return m.CountRecentPageViewsFunc(ctx, sessionID, since)
	}
	return 0, nil
}

func (m *mockAnalyticsRepo) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	if m.GetDashboardStatsFunc != nil {
		return m.GetDashboardStatsFunc(ctx)
	}
	return &domain.DashboardStats{}, nil
}

func (m *mockAnalyticsRepo) DeletePageViewsOlderThan(ctx context.Context, cutoff time.Time) error {
	if m.DeletePageViewsOlderThanFunc != nil {
		return m.DeletePageViewsOlderThanFunc(ctx, cutoff)
	}
	return nil
}

func (m *mockAnalyticsRepo) DeleteSearchEventsOlderThan(ctx context.Context, cutoff time.Time) error {
	if m.DeleteSearchEventsOlderThanFunc != nil {
		return m.DeleteSearchEventsOlderThanFunc(ctx, cutoff)
	}
	return nil
}

func (m *mockAnalyticsRepo) DeleteProfileViewsOlderThan(ctx context.Context, cutoff time.Time) error {
	if m.DeleteProfileViewsOlderThanFunc != nil {
		return m.DeleteProfileViewsOlderThanFunc(ctx, cutoff)
	}
	return nil
}

func (m *mockAnalyticsRepo) DeleteExpiredSessions(ctx context.Context, now time.Time) error {
	if m.DeleteExpiredSessionsFunc != nil {
		return m.DeleteExpiredSessionsFunc(ctx, now)
	}
	return nil
}

// =========================================================================
// TESTS: AnalyticsService
// =========================================================================

func TestAnalyticsService_RecordLogin(t *testing.T) {
	t.Run("llama a CreateLoginEvent y UpsertActiveSession", func(t *testing.T) {
		var loginCalled bool
		var sessionCalled bool
		var capturedEvent domain.LoginEvent
		var capturedSession domain.ActiveSession

		repo := &mockAnalyticsRepo{
			CreateLoginEventFunc: func(ctx context.Context, event domain.LoginEvent) error {
				loginCalled = true
				capturedEvent = event
				return nil
			},
			UpsertActiveSessionFunc: func(ctx context.Context, session domain.ActiveSession) error {
				sessionCalled = true
				capturedSession = session
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		userID := uuid.Must(uuid.NewV7())
		svc.RecordLogin(userID, "admin1", "admin", "127.0.0.1", "TestAgent/1.0")

		// Fire-and-forget: esperamos un poco
		time.Sleep(50 * time.Millisecond)

		require.True(t, loginCalled, "CreateLoginEvent debe ser llamado")
		require.True(t, sessionCalled, "UpsertActiveSession debe ser llamado")
		require.Equal(t, userID, capturedEvent.UserID)
		require.Equal(t, "admin1", capturedEvent.Username)
		require.Equal(t, "admin", capturedEvent.Role)
		require.Equal(t, "127.0.0.1", capturedEvent.IP)
		require.Equal(t, "TestAgent/1.0", capturedEvent.UserAgent)
		require.Equal(t, userID, capturedSession.UserID)
	})

	t.Run("error en CreateLoginEvent no panichea", func(t *testing.T) {
		repo := &mockAnalyticsRepo{
			CreateLoginEventFunc: func(ctx context.Context, event domain.LoginEvent) error {
				return errors.New("db connection lost")
			},
			UpsertActiveSessionFunc: func(ctx context.Context, session domain.ActiveSession) error {
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		// No debe panichear aunque el repo falle
		require.NotPanics(t, func() {
			svc.RecordLogin(uuid.Must(uuid.NewV7()), "u", "psi", "127.0.0.1", "agent")
		})
		time.Sleep(50 * time.Millisecond)
	})
}

func TestAnalyticsService_RecordLogout(t *testing.T) {
	t.Run("llama a DeleteActiveSession", func(t *testing.T) {
		var called bool
		var capturedID uuid.UUID

		repo := &mockAnalyticsRepo{
			DeleteActiveSessionFunc: func(ctx context.Context, userID uuid.UUID) error {
				called = true
				capturedID = userID
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		userID := uuid.Must(uuid.NewV7())
		svc.RecordLogout(userID)
		time.Sleep(50 * time.Millisecond)

		require.True(t, called, "DeleteActiveSession debe ser llamado")
		require.Equal(t, userID, capturedID)
	})
}

func TestAnalyticsService_HeartbeatSession(t *testing.T) {
	t.Run("llama a UpdateSessionHeartbeat con tiempos", func(t *testing.T) {
		var called bool
		repo := &mockAnalyticsRepo{
			UpdateSessionHeartbeatFunc: func(ctx context.Context, userID uuid.UUID, lastSeen, expiresAt time.Time) error {
				called = true
				require.False(t, lastSeen.IsZero(), "lastSeen no debe ser zero")
				require.False(t, expiresAt.IsZero(), "expiresAt no debe ser zero")
				require.True(t, expiresAt.After(lastSeen), "expiresAt debe ser despues de lastSeen")
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		svc.HeartbeatSession(uuid.Must(uuid.NewV7()))
		time.Sleep(50 * time.Millisecond)

		require.True(t, called, "UpdateSessionHeartbeat debe ser llamado")
	})
}

func TestAnalyticsService_RecordSearch(t *testing.T) {
	t.Run("llama a CreateSearchEvent con datos correctos", func(t *testing.T) {
		var called bool
		var capturedEvent domain.SearchEvent

		repo := &mockAnalyticsRepo{
			CreateSearchEventFunc: func(ctx context.Context, event domain.SearchEvent) error {
				called = true
				capturedEvent = event
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		userID := uuid.Must(uuid.NewV7())
		svc.RecordSearch("psicología", "clinica", "Valencia", "Carabobo", 10, &userID, "session123", "192.168.1.1")
		time.Sleep(50 * time.Millisecond)

		require.True(t, called)
		require.Equal(t, "psicología", capturedEvent.Query)
		require.Equal(t, "clinica", capturedEvent.Specialty)
		require.Equal(t, "Valencia", capturedEvent.Municipality)
		require.Equal(t, "Carabobo", capturedEvent.State)
		require.Equal(t, 10, capturedEvent.ResultsCount)
		require.NotNil(t, capturedEvent.UserID)
		require.Equal(t, "session123", capturedEvent.SessionID)
	})

	t.Run("userID nil para busquedas anonimas", func(t *testing.T) {
		var capturedEvent domain.SearchEvent
		repo := &mockAnalyticsRepo{
			CreateSearchEventFunc: func(ctx context.Context, event domain.SearchEvent) error {
				capturedEvent = event
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		svc.RecordSearch("test", "", "", "", 0, nil, "sess", "127.0.0.1")
		time.Sleep(50 * time.Millisecond)

		require.Nil(t, capturedEvent.UserID, "UserID debe ser nil para anonimos")
	})
}

func TestAnalyticsService_RecordProfileView(t *testing.T) {
	t.Run("llama a CreateProfileView", func(t *testing.T) {
		var called bool
		var capturedEvent domain.ProfileView

		repo := &mockAnalyticsRepo{
			CreateProfileViewFunc: func(ctx context.Context, event domain.ProfileView) error {
				called = true
				capturedEvent = event
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		psiID := uuid.Must(uuid.NewV7())
		viewerID := uuid.Must(uuid.NewV7())
		svc.RecordProfileView(psiID, &viewerID, "sess1", "10.0.0.1")
		time.Sleep(50 * time.Millisecond)

		require.True(t, called)
		require.Equal(t, psiID, capturedEvent.PsiID)
		require.NotNil(t, capturedEvent.ViewerID)
	})
}

func TestAnalyticsService_RecordPageView(t *testing.T) {
	t.Run("llama a CreatePageView", func(t *testing.T) {
		var called bool
		var capturedView domain.PageView

		repo := &mockAnalyticsRepo{
			CreatePageViewFunc: func(ctx context.Context, view domain.PageView) error {
				called = true
				capturedView = view
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		svc.RecordPageView(domain.PageView{
			Path:      "/directorio",
			Method:    "GET",
			SessionID: "abc123",
			IP:        "127.0.0.1",
		})
		time.Sleep(50 * time.Millisecond)

		require.True(t, called)
		require.Equal(t, "/directorio", capturedView.Path)
		require.Equal(t, "GET", capturedView.Method)
	})
}

func TestAnalyticsService_CountRecentPageViews(t *testing.T) {
	t.Run("retorna count del repo", func(t *testing.T) {
		repo := &mockAnalyticsRepo{
			CountRecentPageViewsFunc: func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
				return 42, nil
			},
		}
		svc := NewAnalyticsService(repo)

		count, err := svc.CountRecentPageViews(context.Background(), "session1", time.Now().Add(-30*time.Minute))
		require.NoError(t, err)
		require.Equal(t, int64(42), count)
	})

	t.Run("propaga error del repo", func(t *testing.T) {
		repo := &mockAnalyticsRepo{
			CountRecentPageViewsFunc: func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
				return 0, errors.New("db error")
			},
		}
		svc := NewAnalyticsService(repo)

		_, err := svc.CountRecentPageViews(context.Background(), "session1", time.Now())
		require.Error(t, err)
		require.Equal(t, "db error", err.Error())
	})
}

func TestAnalyticsService_GetDashboardStats(t *testing.T) {
	t.Run("retorna stats del repo", func(t *testing.T) {
		expectedStats := &domain.DashboardStats{
			LoginsTotal:     100,
			ActiveSessionsNow: 5,
		}
		repo := &mockAnalyticsRepo{
			GetDashboardStatsFunc: func(ctx context.Context) (*domain.DashboardStats, error) {
				return expectedStats, nil
			},
		}
		svc := NewAnalyticsService(repo)

		stats, err := svc.GetDashboardStats(context.Background())
		require.NoError(t, err)
		require.Equal(t, int64(100), stats.LoginsTotal)
		require.Equal(t, int64(5), stats.ActiveSessionsNow)
	})

	t.Run("propaga error del repo", func(t *testing.T) {
		repo := &mockAnalyticsRepo{
			GetDashboardStatsFunc: func(ctx context.Context) (*domain.DashboardStats, error) {
				return nil, errors.New("query failed")
			},
		}
		svc := NewAnalyticsService(repo)

		_, err := svc.GetDashboardStats(context.Background())
		require.Error(t, err)
	})
}

func TestAnalyticsService_PurgeOldData(t *testing.T) {
	t.Run("elimina datos antiguos sin tocar login events", func(t *testing.T) {
		var pageViewsCalled, searchEventsCalled, profileViewsCalled bool
		var loginEventsCalled bool

		repo := &mockAnalyticsRepo{
			DeletePageViewsOlderThanFunc: func(ctx context.Context, cutoff time.Time) error {
				pageViewsCalled = true
				return nil
			},
			DeleteSearchEventsOlderThanFunc: func(ctx context.Context, cutoff time.Time) error {
				searchEventsCalled = true
				return nil
			},
			DeleteProfileViewsOlderThanFunc: func(ctx context.Context, cutoff time.Time) error {
				profileViewsCalled = true
				return nil
			},
			CreateLoginEventFunc: func(ctx context.Context, event domain.LoginEvent) error {
				loginEventsCalled = true
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		svc.PurgeOldData(context.Background(), 30)

		require.True(t, pageViewsCalled, "DeletePageViewsOlderThan debe ser llamado")
		require.True(t, searchEventsCalled, "DeleteSearchEventsOlderThan debe ser llamado")
		require.True(t, profileViewsCalled, "DeleteProfileViewsOlderThan debe ser llamado")
		require.False(t, loginEventsCalled, "LoginEvents NO deben ser eliminados (audit trail)")
	})
}

func TestAnalyticsService_CleanExpiredSessions(t *testing.T) {
	t.Run("llama a DeleteExpiredSessions", func(t *testing.T) {
		var called bool
		repo := &mockAnalyticsRepo{
			DeleteExpiredSessionsFunc: func(ctx context.Context, now time.Time) error {
				called = true
				require.False(t, now.IsZero(), "now no debe ser zero")
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		svc.CleanExpiredSessions(context.Background())
		require.True(t, called)
	})

	t.Run("propaga error del repo", func(t *testing.T) {
		repo := &mockAnalyticsRepo{
			DeleteExpiredSessionsFunc: func(ctx context.Context, now time.Time) error {
				return errors.New("delete failed")
			},
		}
		svc := NewAnalyticsService(repo)

		require.NotPanics(t, func() {
			svc.CleanExpiredSessions(context.Background())
		})
	})
}

// =========================================================================
// TEST: ActiveSession.IsActive()
// =========================================================================

func TestActiveSession_IsActive(t *testing.T) {
	t.Run("sesion futura esta activa", func(t *testing.T) {
		session := &domain.ActiveSession{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		require.True(t, session.IsActive())
	})

	t.Run("sesion pasada no esta activa", func(t *testing.T) {
		session := &domain.ActiveSession{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		require.False(t, session.IsActive())
	})
}

// =========================================================================
// TEST: Concurrent fire-and-forget safety
// =========================================================================

func TestAnalyticsService_ConcurrentRecordCalls(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	repo := &mockAnalyticsRepo{
		CreateLoginEventFunc: func(ctx context.Context, event domain.LoginEvent) error {
			mu.Lock()
			callCount++
			mu.Unlock()
			return nil
		},
		UpsertActiveSessionFunc: func(ctx context.Context, session domain.ActiveSession) error {
			return nil
		},
		CreateSearchEventFunc: func(ctx context.Context, event domain.SearchEvent) error {
			mu.Lock()
			callCount++
			mu.Unlock()
			return nil
		},
		CreatePageViewFunc: func(ctx context.Context, view domain.PageView) error {
			mu.Lock()
			callCount++
			mu.Unlock()
			return nil
		},
		CreateProfileViewFunc: func(ctx context.Context, event domain.ProfileView) error {
			mu.Lock()
			callCount++
			mu.Unlock()
			return nil
		},
	}
	svc := NewAnalyticsService(repo)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.RecordLogin(uuid.Must(uuid.NewV7()), "u", "psi", "127.0.0.1", "a")
			svc.RecordSearch("q", "", "", "", 0, nil, "s", "127.0.0.1")
			svc.RecordPageView(domain.PageView{Path: "/", Method: "GET"})
			svc.RecordProfileView(uuid.Must(uuid.NewV7()), nil, "s", "127.0.0.1")
		}()
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	require.Equal(t, 200, callCount, "50 goroutines * 4 calls = 200 total")
	mu.Unlock()
}

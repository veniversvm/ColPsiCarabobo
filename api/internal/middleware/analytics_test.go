package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// mockAnalyticsRepoForMiddleware es un mock funcional para el repositorio de analytics.
// Se diferencia del mock en auth_test.go por permitir inyectar comportamiento.
type mockAnalyticsRepoForMiddleware struct {
	domain.AnalyticsRepository
	CreatePageViewFunc       func(view domain.PageView) error
	CountRecentPageViewsFunc func(sessionID string, since time.Time) (int64, error)
}

func (m *mockAnalyticsRepoForMiddleware) CreatePageView(view domain.PageView) error {
	if m.CreatePageViewFunc != nil {
		return m.CreatePageViewFunc(view)
	}
	return nil
}

func (m *mockAnalyticsRepoForMiddleware) CountRecentPageViews(sessionID string, since time.Time) (int64, error) {
	if m.CountRecentPageViewsFunc != nil {
		return m.CountRecentPageViewsFunc(sessionID, since)
	}
	return 0, nil
}

func (m *mockAnalyticsRepoForMiddleware) CreateLoginEvent(domain.LoginEvent) error   { return nil }
func (m *mockAnalyticsRepoForMiddleware) UpsertActiveSession(domain.ActiveSession) error { return nil }
func (m *mockAnalyticsRepoForMiddleware) DeleteActiveSession(uuid.UUID) error        { return nil }
func (m *mockAnalyticsRepoForMiddleware) UpdateSessionHeartbeat(uuid.UUID, time.Time, time.Time) error {
	return nil
}
func (m *mockAnalyticsRepoForMiddleware) CreateSearchEvent(domain.SearchEvent) error  { return nil }
func (m *mockAnalyticsRepoForMiddleware) CreateProfileView(domain.ProfileView) error  { return nil }
func (m *mockAnalyticsRepoForMiddleware) GetDashboardStats() (*domain.DashboardStats, error) {
	return &domain.DashboardStats{}, nil
}
func (m *mockAnalyticsRepoForMiddleware) DeletePageViewsOlderThan(time.Time) error    { return nil }
func (m *mockAnalyticsRepoForMiddleware) DeleteSearchEventsOlderThan(time.Time) error { return nil }
func (m *mockAnalyticsRepoForMiddleware) DeleteProfileViewsOlderThan(time.Time) error { return nil }
func (m *mockAnalyticsRepoForMiddleware) DeleteExpiredSessions(time.Time) error       { return nil }

func init() {
	// Prevenir panic por config.Envs nil en analytics.go
	if config.Envs == nil {
		config.Envs = &config.Config{Environment: "test"}
	}
}

// =========================================================================
// TEST: shouldSkip (función pura)
// =========================================================================

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		path   string
		expect bool
	}{
		{"/health", true},
		{"/health/ready", true},
		{"/favicon.ico", true},
		{"/static/app.js", true},
		{"/static/css/main.css", true},
		{"/assets/logo.png", true},
		{"/_build/index.html", true},
		{"/metrics", true},
		{"/api/users", false},
		{"/api/psi/123", false},
		{"/admin/dashboard", false},
		{"/", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := shouldSkip(tc.path)
			if result != tc.expect {
				t.Errorf("shouldSkip(%q) = %v, want %v", tc.path, result, tc.expect)
			}
		})
	}
}

// =========================================================================
// TEST: AnalyticsMiddleware
// =========================================================================

func TestAnalyticsMiddleware(t *testing.T) {
	t.Run("método no GET no registra page view", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var createCalled int64

		repo.CreatePageViewFunc = func(view domain.PageView) error {
			atomic.AddInt64(&createCalled, 1)
			return nil
		}

		app := fiber.New()
		app.Post("/api/data", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})

		req := httptest.NewRequest("POST", "/api/data", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		app.Test(req)
		time.Sleep(50 * time.Millisecond)

		if atomic.LoadInt64(&createCalled) != 0 {
			t.Error("POST no debería registrar page view")
		}
	})

	t.Run("path skipped no registra", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var createCalled int64

		repo.CreatePageViewFunc = func(view domain.PageView) error {
			atomic.AddInt64(&createCalled, 1)
			return nil
		}

		app := fiber.New()
		app.Get("/health", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})
		app.Get("/static/app.js", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.SendString("js")
		})
		app.Get("/metrics", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.SendString("metrics")
		})

		app.Test(httptest.NewRequest("GET", "/health", nil))
		app.Test(httptest.NewRequest("GET", "/static/app.js", nil))
		app.Test(httptest.NewRequest("GET", "/metrics", nil))
		time.Sleep(50 * time.Millisecond)

		if atomic.LoadInt64(&createCalled) != 0 {
			t.Error("Paths skipped no deberían registrar page views")
		}
	})

	t.Run("admin exento no registra", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var createCalled int64

		repo.CreatePageViewFunc = func(view domain.PageView) error {
			atomic.AddInt64(&createCalled, 1)
			return nil
		}

		app := fiber.New()
		// Cadena: admin injector → analytics → handler
		app.Get("/admin/dashboard",
			func(c *fiber.Ctx) error {
				c.Locals("admin", &domain.UserAdmin{ID: uuid.Must(uuid.NewV7())})
				return c.Next()
			},
			AnalyticsMiddleware(svc),
			func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"page": "dashboard"})
			},
		)

		app.Test(httptest.NewRequest("GET", "/admin/dashboard", nil))
		time.Sleep(50 * time.Millisecond)

		if atomic.LoadInt64(&createCalled) != 0 {
			t.Error("Admin no debería registrar page views")
		}
	})

	t.Run("primera visita genera cookie _sid", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)

		repo.CountRecentPageViewsFunc = func(sessionID string, since time.Time) (int64, error) {
			return 0, nil // primera visita
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"users": []string{}})
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}

		// Verificar que se setea la cookie _sid
		cookies := resp.Header.Values("Set-Cookie")
		found := false
		for _, cookie := range cookies {
			if strings.HasPrefix(cookie, "_sid=") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Debería setear cookie _sid en primera visita")
		}
	})

	t.Run("cookie existente reutiliza session ID", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		existingSID := uuid.Must(uuid.NewV7()).String()

		var capturedSID string
		repo.CountRecentPageViewsFunc = func(sessionID string, since time.Time) (int64, error) {
			capturedSID = sessionID
			return 0, nil
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Cookie", "_sid="+existingSID)
		app.Test(req)
		time.Sleep(50 * time.Millisecond)

		if capturedSID != existingSID {
			t.Errorf("Session ID = %q, want %q (cookie reutilizada)", capturedSID, existingSID)
		}
	})

	t.Run("registra page view con datos correctos", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var capturedView domain.PageView
		var mu sync.Mutex

		repo.CountRecentPageViewsFunc = func(sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(view domain.PageView) error {
			mu.Lock()
			capturedView = view
			mu.Unlock()
			return nil
		}

		app := fiber.New()
		app.Get("/api/psi/profile", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"profile": true})
		})

		req := httptest.NewRequest("GET", "/api/psi/profile", nil)
		req.Header.Set("Referer", "https://google.com")
		app.Test(req)
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if capturedView.Path != "/api/psi/profile" {
			t.Errorf("Path = %q, want /api/psi/profile", capturedView.Path)
		}
		if capturedView.Method != "GET" {
			t.Errorf("Method = %q, want GET", capturedView.Method)
		}
		if capturedView.Referer != "https://google.com" {
			t.Errorf("Referer = %q, want https://google.com", capturedView.Referer)
		}
	})

	t.Run("debouncing dentro de ventana no registra", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var createCount int64

		repo.CountRecentPageViewsFunc = func(sessionID string, since time.Time) (int64, error) {
			return 1, nil // Ya hay vista reciente → debouncing activo
		}
		repo.CreatePageViewFunc = func(view domain.PageView) error {
			atomic.AddInt64(&createCount, 1)
			return nil
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})

		sid := uuid.Must(uuid.NewV7()).String()

		// Primera request con cookie
		req1 := httptest.NewRequest("GET", "/api/users", nil)
		req1.Header.Set("Cookie", "_sid="+sid)
		app.Test(req1)
		time.Sleep(50 * time.Millisecond)

		// Segunda request con misma cookie (debouncing debería activarse)
		req2 := httptest.NewRequest("GET", "/api/users", nil)
		req2.Header.Set("Cookie", "_sid="+sid)
		app.Test(req2)
		time.Sleep(50 * time.Millisecond)

		// CountRecentPageViews retorna 1 → debería saltarse el registro
		// Pero la primera request también tiene count=0 al inicio (el mock siempre retorna 1)
		// Verificamos que no se registraron page views
		if atomic.LoadInt64(&createCount) != 0 {
			t.Error("Debouncing debería prevenir registro, pero CreatePageView fue llamado")
		}
	})

	t.Run("debouncing fuera de ventana sí registra", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var createCount int64

		callIdx := 0
		repo.CountRecentPageViewsFunc = func(sessionID string, since time.Time) (int64, error) {
			callIdx++
			if callIdx == 1 {
				return 0, nil // Primera request: sin vistas previas
			}
			return 0, nil // Segunda request: también sin vistas (ventana expiró)
		}
		repo.CreatePageViewFunc = func(view domain.PageView) error {
			atomic.AddInt64(&createCount, 1)
			return nil
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})

		sid := uuid.Must(uuid.NewV7()).String()

		// Primera request
		req1 := httptest.NewRequest("GET", "/api/users", nil)
		req1.Header.Set("Cookie", "_sid="+sid)
		app.Test(req1)
		time.Sleep(50 * time.Millisecond)

		// Simular paso de tiempo: CountRecentPageViews retorna 0
		// ya que la ventana de 30 min ya pasó
		req2 := httptest.NewRequest("GET", "/api/users", nil)
		req2.Header.Set("Cookie", "_sid="+sid)
		app.Test(req2)
		time.Sleep(50 * time.Millisecond)

		// Ambas requests deberían registrar page views
		count := atomic.LoadInt64(&createCount)
		if count != 2 {
			t.Errorf("CreatePageView llamado %d veces, want 2 (fuera de ventana)", count)
		}
	})

	t.Run("usuario logueado registra con userID", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		userID := uuid.Must(uuid.NewV7())
		var capturedView domain.PageView
		var mu sync.Mutex

		repo.CountRecentPageViewsFunc = func(sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(view domain.PageView) error {
			mu.Lock()
			capturedView = view
			mu.Unlock()
			return nil
		}

		app := fiber.New()
		// Simular middleware de auth que setea userID
		app.Get("/api/profile",
			func(c *fiber.Ctx) error {
				c.Locals("userID", userID)
				return c.Next()
			},
			AnalyticsMiddleware(svc),
			func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"profile": true})
			},
		)

		req := httptest.NewRequest("GET", "/api/profile", nil)
		app.Test(req)
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if capturedView.UserID == nil {
			t.Error("UserID debería estar seteado para usuario logueado")
		} else if *capturedView.UserID != userID {
			t.Errorf("UserID = %v, want %v", *capturedView.UserID, userID)
		}
	})

	t.Run("usuario anónimo registra sin userID", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var capturedView domain.PageView
		var mu sync.Mutex

		repo.CountRecentPageViewsFunc = func(sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(view domain.PageView) error {
			mu.Lock()
			capturedView = view
			mu.Unlock()
			return nil
		}

		app := fiber.New()
		app.Get("/api/directory", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"results": []string{}})
		})

		req := httptest.NewRequest("GET", "/api/directory", nil)
		app.Test(req)
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if capturedView.UserID != nil {
			t.Errorf("UserID debería ser nil para usuario anónimo, got %v", *capturedView.UserID)
		}
	})
}

// TestAnalyticsMiddleware_SkipPathsCoverage verifica paths adicionales
func TestAnalyticsMiddleware_SkipPathsCoverage(t *testing.T) {
	repo := &mockAnalyticsRepoForMiddleware{}
	svc := service.NewAnalyticsService(repo)
	var createCount int64

	repo.CountRecentPageViewsFunc = func(sessionID string, since time.Time) (int64, error) {
		return 0, nil
	}
	repo.CreatePageViewFunc = func(view domain.PageView) error {
		atomic.AddInt64(&createCount, 1)
		return nil
	}

	app := fiber.New()
	for _, path := range []string{"/assets/logo.png", "/_build/index.html"} {
		app.Get(path, AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})
	}

	app.Test(httptest.NewRequest("GET", "/assets/logo.png", nil))
	app.Test(httptest.NewRequest("GET", "/_build/index.html", nil))
	time.Sleep(50 * time.Millisecond)

	body, _ := io.ReadAll(httptest.NewRequest("GET", "/api/users", nil).Body)
	_ = body

	if atomic.LoadInt64(&createCount) != 0 {
		t.Error("Assets y _build no deberían registrar page views")
	}
}

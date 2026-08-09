package middleware

import (
	"context"
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
	CreatePageViewFunc       func(ctx context.Context, view domain.PageView) error
	CountRecentPageViewsFunc func(ctx context.Context, sessionID string, since time.Time) (int64, error)
}

func (m *mockAnalyticsRepoForMiddleware) CreatePageView(ctx context.Context, view domain.PageView) error {
	if m.CreatePageViewFunc != nil {
		return m.CreatePageViewFunc(ctx, view)
	}
	return nil
}

func (m *mockAnalyticsRepoForMiddleware) CountRecentPageViews(ctx context.Context, sessionID string, since time.Time) (int64, error) {
	if m.CountRecentPageViewsFunc != nil {
		return m.CountRecentPageViewsFunc(ctx, sessionID, since)
	}
	return 0, nil
}

func (m *mockAnalyticsRepoForMiddleware) CreateLoginEvent(context.Context, domain.LoginEvent) error { return nil }
func (m *mockAnalyticsRepoForMiddleware) UpsertActiveSession(context.Context, domain.ActiveSession) error {
	return nil
}
func (m *mockAnalyticsRepoForMiddleware) DeleteActiveSession(context.Context, uuid.UUID) error { return nil }
func (m *mockAnalyticsRepoForMiddleware) UpdateSessionHeartbeat(context.Context, uuid.UUID, time.Time, time.Time) error {
	return nil
}
func (m *mockAnalyticsRepoForMiddleware) CreateSearchEvent(context.Context, domain.SearchEvent) error  { return nil }
func (m *mockAnalyticsRepoForMiddleware) CreateProfileView(context.Context, domain.ProfileView) error  { return nil }
func (m *mockAnalyticsRepoForMiddleware) GetDashboardStats(context.Context) (*domain.DashboardStats, error) {
	return &domain.DashboardStats{}, nil
}
func (m *mockAnalyticsRepoForMiddleware) DeletePageViewsOlderThan(context.Context, time.Time) error    { return nil }
func (m *mockAnalyticsRepoForMiddleware) DeleteSearchEventsOlderThan(context.Context, time.Time) error { return nil }
func (m *mockAnalyticsRepoForMiddleware) DeleteProfileViewsOlderThan(context.Context, time.Time) error { return nil }
func (m *mockAnalyticsRepoForMiddleware) DeleteExpiredSessions(context.Context, time.Time) error       { return nil }

func init() {
	// Prevenir panic por config.Envs nil en analytics.go
	if config.Envs == nil {
		config.Envs = &config.Config{Environment: "test"}
	}
}

// testBrowserUA es un User-Agent realista para requests que deben pasar el
// filtro de datos (los bots y UA vacíos se excluyen del tracking).
const testBrowserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

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
// TEST: isBotUA (función pura)
// =========================================================================

func TestIsBotUA(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expect    bool
	}{
		{
			name:      "Googlebot detectado",
			userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expect:    true,
		},
		{
			name:      "Bingbot detectado",
			userAgent: "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
			expect:    true,
		},
		{
			name:      "AhrefsBot detectado",
			userAgent: "Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)",
			expect:    true,
		},
		{
			name:      "GPTBot detectado",
			userAgent: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; GPTBot/1.0; +https://openai.com/gptbot",
			expect:    true,
		},
		{
			name:      "Facebook external hit detectado",
			userAgent: "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
			expect:    true,
		},
		{
			name:      "curl detectado",
			userAgent: "curl/8.5.0",
			expect:    true,
		},
		{
			name:      "UA vacío considerado bot",
			userAgent: "",
			expect:    true,
		},
		{
			name:      "UA solo con espacios considerado bot",
			userAgent: "   ",
			expect:    true,
		},
		{
			name:      "Chrome real no es bot",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
			expect:    false,
		},
		{
			name:      "Firefox real no es bot",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:127.0) Gecko/20100101 Firefox/127.0",
			expect:    false,
		},
		{
			name:      "Safari iOS real no es bot",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
			expect:    false,
		},
		{
			name:      "subcadena dentro de otro token detectada",
			userAgent: "Mozilla/5.0 (compatible; NotGooglebot/1.0; +http://example.com/)",
			expect:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isBotUA(tc.userAgent)
			if result != tc.expect {
				t.Errorf("isBotUA(%q) = %v, want %v", tc.userAgent, result, tc.expect)
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

		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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

		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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

		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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

		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			return 0, nil // primera visita
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"users": []string{}})
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		req.Header.Set("User-Agent", testBrowserUA)
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
		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			capturedSID = sessionID
			return 0, nil
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Cookie", "_sid="+existingSID)
		req.Header.Set("User-Agent", testBrowserUA)
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

		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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
		req.Header.Set("User-Agent", testBrowserUA)
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

		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			return 1, nil // Ya hay vista reciente → debouncing activo
		}
		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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
		req1.Header.Set("User-Agent", testBrowserUA)
		app.Test(req1)
		time.Sleep(50 * time.Millisecond)

		// Segunda request con misma cookie (debouncing debería activarse)
		req2 := httptest.NewRequest("GET", "/api/users", nil)
		req2.Header.Set("Cookie", "_sid="+sid)
		req2.Header.Set("User-Agent", testBrowserUA)
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
		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			callIdx++
			if callIdx == 1 {
				return 0, nil // Primera request: sin vistas previas
			}
			return 0, nil // Segunda request: también sin vistas (ventana expiró)
		}
		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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
		req1.Header.Set("User-Agent", testBrowserUA)
		app.Test(req1)
		time.Sleep(50 * time.Millisecond)

		// Simular paso de tiempo: CountRecentPageViews retorna 0
		// ya que la ventana de 30 min ya pasó
		req2 := httptest.NewRequest("GET", "/api/users", nil)
		req2.Header.Set("Cookie", "_sid="+sid)
		req2.Header.Set("User-Agent", testBrowserUA)
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

		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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
		req.Header.Set("User-Agent", testBrowserUA)
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

		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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
		req.Header.Set("User-Agent", testBrowserUA)
		app.Test(req)
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if capturedView.UserID != nil {
			t.Errorf("UserID debería ser nil para usuario anónimo, got %v", *capturedView.UserID)
		}
	})
}

// TestAnalyticsMiddleware_BotSkip verifica que un User-Agent de bot no registra
// page views ni genera cookie de sesión (es el origen del ruido: los bots no
// persisten cookies, así que cada request contaba como "visita nueva").
func TestAnalyticsMiddleware_BotSkip(t *testing.T) {
	t.Run("Googlebot no registra ni crea _sid", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var createCount int64

		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
			atomic.AddInt64(&createCount, 1)
			return nil
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		if atomic.LoadInt64(&createCount) != 0 {
			t.Error("Googlebot no debería registrar page views")
		}
		if len(resp.Header.Values("Set-Cookie")) != 0 {
			t.Error("Googlebot no debería recibir cookie _sid")
		}
	})

	t.Run("UA vacío no registra ni crea _sid", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var createCount int64

		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
			atomic.AddInt64(&createCount, 1)
			return nil
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		if atomic.LoadInt64(&createCount) != 0 {
			t.Error("UA vacío no debería registrar page views")
		}
		if len(resp.Header.Values("Set-Cookie")) != 0 {
			t.Error("UA vacío no debería recibir cookie _sid")
		}
	})

	t.Run("navegador real sí registra y crea _sid", func(t *testing.T) {
		repo := &mockAnalyticsRepoForMiddleware{}
		svc := service.NewAnalyticsService(repo)
		var createCount int64

		repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
			return 0, nil
		}
		repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
			atomic.AddInt64(&createCount, 1)
			return nil
		}

		app := fiber.New()
		app.Get("/api/users", AnalyticsMiddleware(svc), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		if atomic.LoadInt64(&createCount) != 1 {
			t.Errorf("Chrome real debería registrar 1 page view, got %d", atomic.LoadInt64(&createCount))
		}
		if len(resp.Header.Values("Set-Cookie")) == 0 {
			t.Error("Chrome real debería recibir cookie _sid")
		}
	})
}

// TestAnalyticsMiddleware_SkipPathsCoverage verifica paths adicionales
func TestAnalyticsMiddleware_SkipPathsCoverage(t *testing.T) {
	repo := &mockAnalyticsRepoForMiddleware{}
	svc := service.NewAnalyticsService(repo)
	var createCount int64

	repo.CountRecentPageViewsFunc = func(ctx context.Context, sessionID string, since time.Time) (int64, error) {
		return 0, nil
	}
	repo.CreatePageViewFunc = func(ctx context.Context, view domain.PageView) error {
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

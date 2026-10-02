package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// testBrowserUA es un User-Agent de navegador real. Los tests que esperan que el
// evento SE guarde lo pasan; los que esperan que se descarte pasan uno de bot.
const testBrowserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// analyticsCountingRepo cuenta las escrituras que llegan al repositorio. Los
// métodos son called desde goroutines, así que los contadores son atómicos.
type analyticsCountingRepo struct {
	mockAnalyticsRepo
	searches  atomic.Int64
	profiles  atomic.Int64
	pageViews atomic.Int64
}

func (r *analyticsCountingRepo) CreateSearchEvent(ctx context.Context, event domain.SearchEvent) error {
	r.searches.Add(1)
	return nil
}

func (r *analyticsCountingRepo) CreateProfileView(ctx context.Context, event domain.ProfileView) error {
	r.profiles.Add(1)
	return nil
}

func (r *analyticsCountingRepo) CreatePageView(ctx context.Context, view domain.PageView) error {
	r.pageViews.Add(1)
	return nil
}

// settleAnalytics espera a que las escrituras fire-and-forget terminen. Los
// métodos de Record* lanzan una goroutine, así que el test tiene que darle
// margen; es el mismo 50 ms que usan los tests vecinos.
func settleAnalytics() { time.Sleep(50 * time.Millisecond) }

// TestIsBotUA es la tabla de la función pura. Viaja con la función al paquete
// service a propósito: el middleware la sigue usando, y su cobertura de
// comportamiento vive en TestAnalyticsMiddleware_BotSkip.
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
			userAgent: testBrowserUA,
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
			result := IsBotUA(tc.userAgent)
			if result != tc.expect {
				t.Errorf("IsBotUA(%q) = %v, want %v", tc.userAgent, result, tc.expect)
			}
		})
	}
}

// TestAnalyticsService_BotsNoEscriben es el contrato que faltaba: el ruido de
// máquina tiene que quedarse FUERA de las tablas, no dentro con un filtro de
// lectura en el panel. El filtro de bots vivía solo en AnalyticsMiddleware, y
// las búsquedas y las fichas —que se invocan desde los handlers, fuera del
// middleware— no lo heredaban: un monitor que pegaba cada 60 s al directorio
// llenaba search_events mientras page_views quedaba limpio. De ahí que
// "Búsquedas" fuera 40× "Visitas" y las dos cifras no fueran comparables.
func TestAnalyticsService_BotsNoEscriben(t *testing.T) {
	psiID := uuid.Must(uuid.NewV7())

	t.Run("RecordSearch con UA de bot no escribe", func(t *testing.T) {
		for _, ua := range []string{"curl/8.5.0", "", "UptimeRobot/2.0", "python-requests/2.32"} {
			repo := &analyticsCountingRepo{}
			svc := NewAnalyticsService(repo)

			svc.RecordSearch("depresion", "", "valencia", "", 3, nil, "sess-1", "190.52.130.45", ua)
			settleAnalytics()

			if got := repo.searches.Load(); got != 0 {
				t.Fatalf("UA %q escribió %d search_events; se esperaba 0", ua, got)
			}
		}
	})

	t.Run("RecordProfileView con UA de bot no escribe", func(t *testing.T) {
		for _, ua := range []string{"python-requests/2.32", "", "Googlebot/2.1"} {
			repo := &analyticsCountingRepo{}
			svc := NewAnalyticsService(repo)

			svc.RecordProfileView(psiID, nil, "sess-1", "190.52.130.45", ua)
			settleAnalytics()

			if got := repo.profiles.Load(); got != 0 {
				t.Fatalf("UA %q escribió %d profile_views; se esperaba 0", ua, got)
			}
		}
	})

	t.Run("RecordSearch con UA de navegador sí escribe", func(t *testing.T) {
		repo := &analyticsCountingRepo{}
		svc := NewAnalyticsService(repo)

		svc.RecordSearch("depresion", "", "valencia", "", 3, nil, "sess-1", "190.52.130.45", testBrowserUA)
		settleAnalytics()

		if got := repo.searches.Load(); got != 1 {
			t.Fatalf("un navegador real escribió %d search_events; se esperaba 1", got)
		}
	})

	t.Run("RecordProfileView con UA de navegador sí escribe", func(t *testing.T) {
		repo := &analyticsCountingRepo{}
		svc := NewAnalyticsService(repo)

		svc.RecordProfileView(psiID, nil, "sess-1", "190.52.130.45", testBrowserUA)
		settleAnalytics()

		if got := repo.profiles.Load(); got != 1 {
			t.Fatalf("un navegador real escribió %d profile_views; se esperaba 1", got)
		}
	})
}

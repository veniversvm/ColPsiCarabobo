package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// TestAnalyticsService_fingerprintIP comprueba que la IP nunca llega a la base
// tal cual, que es la razón de existir de analytics_privacy.go.
func TestAnalyticsService_fingerprintIP(t *testing.T) {
	svc := NewAnalyticsService(nil)

	t.Run("la IP no se guarda en claro", func(t *testing.T) {
		h := svc.fingerprintIP("190.52.130.45")
		assert.NotEqual(t, "190.52.130.45", h)
		assert.NotContains(t, h, "190.52")
		assert.Len(t, h, 64) // hex de SHA-256
	})

	t.Run("la misma IP produce la misma huella", func(t *testing.T) {
		// Requisito de los conteos de visitantes únicos: si la huella cambiara
		// entre llamadas, cada visita contaría como un visitante nuevo.
		assert.Equal(t, svc.fingerprintIP("190.52.130.45"), svc.fingerprintIP("190.52.130.45"))
	})

	t.Run("IP distintas producen huellas distintas", func(t *testing.T) {
		assert.NotEqual(t, svc.fingerprintIP("190.52.130.45"), svc.fingerprintIP("190.52.130.46"))
	})

	t.Run("sin IP no hay huella, no un hash vacío", func(t *testing.T) {
		assert.Empty(t, svc.fingerprintIP(""))
		assert.Empty(t, svc.fingerprintIP("   "))
	})

	t.Run("la sal separa contextos distintos", func(t *testing.T) {
		other := &AnalyticsService{ipSalt: "otra-sal"}
		assert.NotEqual(t, svc.fingerprintIP("190.52.130.45"), other.fingerprintIP("190.52.130.45"))
	})
}

// TestResolveIPSalt documenta el fallback: sin ANALYTICS_IP_SALT se usa una
// constante, para que un despliegue mal configurado no guarde IPs en claro.
func TestResolveIPSalt(t *testing.T) {
	require.NotEmpty(t, resolveIPSalt())
	assert.Equal(t, analyticsDefaultIPSalt, resolveIPSalt())
}

// TestRefererOrigin comprueba que del Referer solo sobrevive el origen, que es
// donde entra la URL completa con los parámetros de terceros.
func TestRefererOrigin(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		expected string
	}{
		{"se queda el origen", "https://www.google.com/search?q=depresion+carabobo", "https://www.google.com"},
		{"conserva el esquema", "http://facebook.com/psicologos", "http://facebook.com"},
		{"con puerto", "https://colegio.example.ve:8443/directorio", "https://colegio.example.ve:8443"},
		{"interno al Colegio", "https://colegio-psicologos-carabobo.com/directorio?q=ana", "https://colegio-psicologos-carabobo.com"},
		{"vacío", "", ""},
		{"solo espacios", "   ", ""},
		{"sin host", "/directorio", ""},
		{"no es una URL", "javascript:alert(1)", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := refererOrigin(c.in)
			assert.Equal(t, c.expected, got)
			// Invariante de privacidad: más allá del separador "://" no sobrevive
			// ni la ruta ni los parámetros de la URL original.
			assert.NotContains(t, got, "?")
			rest, found := strings.CutPrefix(got, "http://")
			if !found {
				rest, _ = strings.CutPrefix(got, "https://")
			}
			assert.NotContains(t, rest, "/")
		})
	}
}

// TestTrackPageViewSaneaAntesDePersistir es el test de contrato: aunque el
// middleware (o cualquier futuro llamador) maneje la IP y el Referer en claro,
// lo que llega al repositorio ya está reducido.
func TestTrackPageViewSaneaAntesDePersistir(t *testing.T) {
	var guardada domain.PageView
	creada := false

	repo := &mockAnalyticsRepo{
		CountRecentPageViewsFunc: func(context.Context, string, time.Time) (int64, error) { return 0, nil },
		CreatePageViewFunc: func(_ context.Context, v domain.PageView) error {
			guardada = v
			creada = true
			return nil
		},
	}
	svc := NewAnalyticsService(repo)

	svc.TrackPageView(t.Context(), domain.PageView{
		Path:      "/directorio",
		SessionID: "sess-1",
		IP:        "190.52.130.45",
		Referer:   "https://www.google.com/search?q=secreto",
	})

	require.True(t, creada, "la visita debió persistirse")
	assert.NotContains(t, guardada.IP, "190.52")
	assert.Equal(t, "https://www.google.com", guardada.Referer)
	assert.NotContains(t, guardada.Referer, "secreto")
	// Lo útil para la métrica se conserva intacto.
	assert.Equal(t, "/directorio", guardada.Path)
	assert.Equal(t, "sess-1", guardada.SessionID)
}

// TestRecordSearchHasheaIP cubre el segundo de los tres puntos de captura.
func TestRecordSearchHasheaIP(t *testing.T) {
	var guardada domain.SearchEvent
	listo := make(chan struct{})

	repo := &mockAnalyticsRepo{
		CreateSearchEventFunc: func(_ context.Context, e domain.SearchEvent) error {
			guardada = e
			close(listo)
			return nil
		},
	}
	svc := NewAnalyticsService(repo)

	svc.RecordSearch("depresion", "", "valencia", "", 0, nil, "sess-1", "190.52.130.45")
	<-listo

	assert.NotContains(t, guardada.IP, "190.52")
	// El texto de la búsqueda SÍ se conserva: el Colegio lo consulta en el panel
	// de métricas y así está declarado en los Términos.
	assert.Equal(t, "depresion", guardada.Query)
}

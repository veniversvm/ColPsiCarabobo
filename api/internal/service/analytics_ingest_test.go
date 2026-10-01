package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// TestClamp documenta el contrato del recorte: rune, no byte. Un recorte a bytes
// partiría un carácter UTF-8 por la mitad y Postgres rechazaría el varchar igual
// que con el 22001 que viene a sustituir, así que el error se movería de sitio en
// vez de desaparecer.
func TestClamp(t *testing.T) {
	t.Run("deja intacto lo que ya cabe", func(t *testing.T) {
		assert.Equal(t, "", clamp("", 10))
		assert.Equal(t, "corto", clamp("corto", 10))
		assert.Equal(t, "0123456789", clamp("0123456789", 10))
	})

	t.Run("recorta por caracteres, no por bytes", func(t *testing.T) {
		// "á" ocupa 2 bytes en UTF-8: con 4 bytes de entrada, un recorte a bytes
		// devolvería un string con el último carácter partido por la mitad.
		got := clamp("áéíóú", 3)
		assert.Equal(t, "áéí", got)
		assert.True(t, utf8.ValidString(got), "el recorte no puede partir un carácter")
		assert.Equal(t, 3, len([]rune(got)))
	})

	t.Run("no entra en pánico con max 0", func(t *testing.T) {
		assert.Equal(t, "", clamp("texto", 0))
	})
}

// TestAnalyticsService_RecortaLoQueVieneDelCliente es el test del bug hermano del
// que sufrió la huella de 64: los campos que puede manipular el cliente (ruta,
// método, cookie _sid, texto del buscador) no tienen ningún límite upstream, y sin
// el recorte cada uno de ellos bastaba para tumbar su INSERT con 22001 y perder
// la métrica en silencio.
func TestAnalyticsService_RecortaLoQueVieneDelCliente(t *testing.T) {
	const enorme = 4000

	t.Run("TrackPageView recorta ruta, método y _sid", func(t *testing.T) {
		var guardada domain.PageView
		repo := &mockAnalyticsRepo{
			CountRecentPageViewsFunc: func(context.Context, string, time.Time) (int64, error) { return 0, nil },
			CreatePageViewFunc: func(_ context.Context, v domain.PageView) error {
				guardada = v
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		svc.TrackPageView(t.Context(), domain.PageView{
			Path:      "/" + strings.Repeat("a", enorme),
			Method:    strings.Repeat("M", enorme),
			SessionID: strings.Repeat("s", enorme), // una _sid escrita a mano
			IPHash:    "190.52.130.45",
			Referer:   "https://" + strings.Repeat("b", enorme) + "/privado?token=x",
		})

		assert.Len(t, guardada.Path, analyticsMaxPath)
		assert.Len(t, guardada.Method, analyticsMaxMethod)
		assert.Len(t, guardada.SessionID, analyticsMaxSessionID)
		assert.Len(t, guardada.Referer, analyticsMaxReferer)
		assert.True(t, utf8.ValidString(guardada.Path))
		assert.True(t, utf8.ValidString(guardada.Referer))
	})

	t.Run("RecordSearch recorta el texto y los filtros", func(t *testing.T) {
		var guardada domain.SearchEvent
		repo := &mockAnalyticsRepo{
			CreateSearchEventFunc: func(_ context.Context, e domain.SearchEvent) error {
				guardada = e
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		svc.RecordSearch(
			strings.Repeat("q", enorme), strings.Repeat("s", enorme),
			strings.Repeat("m", enorme), strings.Repeat("e", enorme),
			0, nil, strings.Repeat("s", enorme), "190.52.130.45",
		)
		require.Eventually(t, func() bool { return guardada.Query != "" }, time.Second, 10*time.Millisecond)

		assert.Len(t, guardada.Query, analyticsMaxFilter)
		assert.Len(t, guardada.Specialty, analyticsMaxFilter)
		assert.Len(t, guardada.Municipality, analyticsMaxFilter)
		assert.Len(t, guardada.State, analyticsMaxFilter)
		assert.Len(t, guardada.SessionID, analyticsMaxSessionID)
		// El texto recortado sigue siendo el que tecleó el visitante: se conserva
		// verbatim (es lo que declara el §10.1 de los Términos), no se descarta.
		assert.Equal(t, strings.Repeat("q", analyticsMaxFilter), guardada.Query)
		// La huella no se recorta (eso la haría incorrecta): o mide el ancho de
		// la columna o el INSERT falla de forma visible.
		assert.Len(t, guardada.IPHash, analyticsMaxIPHash)
	})

	t.Run("RecordProfileView recorta la _sid", func(t *testing.T) {
		var guardada domain.ProfileView
		repo := &mockAnalyticsRepo{
			CreateProfileViewFunc: func(_ context.Context, e domain.ProfileView) error {
				guardada = e
				return nil
			},
		}
		svc := NewAnalyticsService(repo)

		svc.RecordProfileView(uuid.Must(uuid.NewV7()), nil, strings.Repeat("s", enorme), "190.52.130.45")
		require.Eventually(t, func() bool { return guardada.SessionID != "" }, time.Second, 10*time.Millisecond)

		assert.Len(t, guardada.SessionID, analyticsMaxSessionID)
	})
}

// TestAnalyticsService_ReportWriteErrorNoInunda verifica el otro compromiso: que el
// error se vea. Un throttle sin límite sería un arreglo a medias — taparía la
// ceguera creando una inundación de log justo cuando la base está caída, que es
// cuando más lo necesitas.
func TestAnalyticsService_ReportWriteErrorNoInunda(t *testing.T) {
	t.Run("la primera vez avisa y luego se calla", func(t *testing.T) {
		svc := NewAnalyticsService(nil)

		svc.reportWriteError("create_page_view", errors.New("value too long"))
		st := svc.writeErrState["create_page_view"]
		require.NotNil(t, st)
		assert.True(t, st.active, "el primer fallo debe registrarse")
		assert.Equal(t, 0, st.suppressed)

		// Diez fallos más dentro de la ventana: ninguno debe imprimirse.
		for i := 0; i < 10; i++ {
			svc.reportWriteError("create_page_view", errors.New("value too long"))
		}
		assert.Equal(t, 10, st.suppressed, "los fallos siguientes deben acumularse, no imprimirse")
		assert.True(t, time.Since(st.lastLog) < analyticsErrorLogInterval, "no debe reimprimir dentro de la ventana")
	})

	t.Run("pasada la ventana, una línea arrastrando el total de omitidas", func(t *testing.T) {
		svc := NewAnalyticsService(nil)
		svc.reportWriteError("create_search_event", errors.New("boom"))
		for i := 0; i < 5; i++ {
			svc.reportWriteError("create_search_event", errors.New("boom"))
		}

		st := svc.writeErrState["create_search_event"]
		// Se atrasa la ventana para simular que pasó el minuto.
		st.lastLog = time.Now().Add(-2 * analyticsErrorLogInterval)
		svc.reportWriteError("create_search_event", errors.New("boom"))

		assert.Equal(t, 0, st.suppressed, "la línea nueva arrastra el conteo de omitidas")
		assert.WithinDuration(t, time.Now(), st.lastLog, time.Second)
	})

	t.Run("cada operación lleva su propia cuenta", func(t *testing.T) {
		svc := NewAnalyticsService(nil)
		svc.reportWriteError("create_page_view", errors.New("a"))
		svc.reportWriteError("create_search_event", errors.New("b"))

		assert.Len(t, svc.writeErrState, 2)
		assert.True(t, svc.writeErrState["create_page_view"].active)
		assert.True(t, svc.writeErrState["create_search_event"].active)
	})

	t.Run("reportWriteOK avisa de la recuperación, y solo una vez", func(t *testing.T) {
		svc := NewAnalyticsService(nil)
		svc.reportWriteError("create_page_view", errors.New("a"))
		require.True(t, svc.writeErrState["create_page_view"].active)

		svc.reportWriteOK("create_page_view")
		assert.False(t, svc.writeErrState["create_page_view"].active, "tras recuperarse deja de estar en fallo")

		// Un OK con la operación ya sana no debe avisar (si no, cada visita
		// logging una advertencia de "volvió a funcionar").
		svc.reportWriteOK("create_page_view")
		assert.False(t, svc.writeErrState["create_page_view"].active)

		// Y si vuelve a fallar, avisa otra vez en vez de heredar el silencio.
		svc.reportWriteError("create_page_view", errors.New("a"))
		assert.True(t, svc.writeErrState["create_page_view"].active)
	})

	t.Run("reportWriteOK sobre un mapa nil no entra en panic", func(t *testing.T) {
		// analytics_privacy_test.go construye el servicio como literal; si un test
		// futuro lo hace y dispara una escritura, no debe reventar.
		svc := &AnalyticsService{ipSalt: "x"}
		svc.reportWriteOK("create_page_view")
	})
}

// TestAnalyticsService_ReportWriteErrorEsConcurrente es el guard de la carrera: las
// escrituras corren en goroutines propias, así que sin el mutex dos actualizaciones
// del mapa se pisan. Con -race este test vale por lo que vale.
func TestAnalyticsService_ReportWriteErrorEsConcurrente(t *testing.T) {
	svc := NewAnalyticsService(nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			op := "op_a"
			if i%2 == 0 {
				op = "op_b"
			}
			svc.reportWriteError(op, errors.New("boom"))
			svc.reportWriteOK(op)
		}(i)
	}
	wg.Wait()

	assert.Len(t, svc.writeErrState, 2)
}

// TestAnalyticsService_TrackPageViewRecortaYSigueDeduplicando comprueba que el
// recorte no rompió la deduplicación, que es lo que evita que las visitas se
// inflen.
func TestAnalyticsService_TrackPageViewRecortaYSigueDeduplicando(t *testing.T) {
	creadas := 0
	repo := &mockAnalyticsRepo{
		CountRecentPageViewsFunc: func(context.Context, string, time.Time) (int64, error) {
			if creadas > 0 {
				return 1, nil
			}
			return 0, nil
		},
		CreatePageViewFunc: func(context.Context, domain.PageView) error {
			creadas++
			return nil
		},
	}
	svc := NewAnalyticsService(repo)

	view := domain.PageView{Path: "/directorio", Method: "GET", SessionID: "sess-1", IPHash: "1.2.3.4"}
	svc.TrackPageView(t.Context(), view)
	svc.TrackPageView(t.Context(), view)

	assert.Equal(t, 1, creadas, "la segunda visita dentro de la ventana no debe escribir")
}

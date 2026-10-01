// Lo que hace que un evento de telemetría llegue a la base y se note cuando no
// llega. Son dos problemas que costaron una mañana de métricas congeladas y que
// nunca son errores "visibles" por el lado del cliente, porque las escrituras son
// fire-and-forget y la respuesta al visitante no depende de ellas:
//
//  1. El ancho de la columna. El 22001 "value too long for type character
//     varying(N)" no lo provoca el dato raro, lo provoca el dato LARGO que
//     cualquiera puede mandar: la cookie _sid la escribe el cliente, el texto del
//     buscador lo teclea el visitante, y el Referer y el URI los manda el cliente
//     sin ningún límite. Cada uno de esos INSERT se descartaba en silencio. Aquí
//     se recortan a los anchos que declara la migración, en un solo sitio, para
//     que el siguiente campo que se invente no repita el bug (y para que ampliar
//     una columna sea un cambio consciente, no una sorpresa en producción).
//
//  2. El silencio. Un error de escritura se registraba con "_ =" y se olvidaba.
//     Cuando el error era constante —una columna de 45 con una huella de 64— el
//     panel dejó de recibir datos durante un día entero sin que nadie lo notara
//     más allá de una línea de GORM en el log. Aquí el fallo se registra con
//     throttling: la telemetría es de altísimo volumen, así que una línea por
//     visita convertiría una base caída en una inundación de log, que es
//     justamente cuando más lo necesitas.
package service

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// Anchos de columna de las tablas de telemetría. Cada constante está atada a una
// columna de api/migrations (baseline para session_id/path/method/referer y
// 20261001100000_analytics_ip_hash.sql para ip_hash): si cambias una, cambias
// la otra, y el contrato lo comprueba
// TestAnalyticsRepo_ColumnasAguantanLoQueMandaElServicio.
const (
	analyticsMaxPath      = 512 // page_views.path
	analyticsMaxMethod    = 10  // page_views.method
	analyticsMaxReferer   = 512 // page_views.referer
	analyticsMaxSessionID = 64  // page_views.session_id, search_events.session_id, profile_views.session_id
	analyticsMaxFilter    = 255 // search_events.query / specialty / municipality / state
)

// analyticsMaxIPHash documenta el ancho de las columnas ip_hash
// (20261001100000_analytics_ip_hash.sql). No se usa para recortar —una huella
// recortada en silencio sería una huella incorrecta— sino para que el test que
// comprueba el contrato cite el número desde un solo sitio.
const analyticsMaxIPHash = 64

// analyticsErrorLogInterval es la ventana mínima entre dos líneas del mismo
// error de escritura. El primero sale completo; los siguientes, uno por minuto con
// el total de omitidos, para que un fallo sostenido se lea sin enterrar el log.
const analyticsErrorLogInterval = time.Minute

// analyticsWriteOpError es el estado de una operación que está fallando.
type analyticsWriteOpError struct {
	// lastLog es cuándo se escribió la última línea de este error.
	lastLog time.Time
	// suppressed cuenta los fallos que no se imprimieron desde esa línea.
	suppressed int
	// active marca que la operación está en fallo, para poder avisar una sola vez
	// cuando se recupere en lugar de callarse para siempre.
	active bool
}

// reportWriteError registra el fallo de una escritura analítica SIN propagarlo:
// la telemetría nunca debe romper la petición del visitante, ni siquiera para
// reportar. Throttle por operación —la primera vez sale la línea completa y
// mientras siga fallando como mucho una por minuto con el total de omitidas—,
// porque el error que hay que ver es "llevo un rato sin escribir", no 4.000
// líneas de lo mismo.
func (s *AnalyticsService) reportWriteError(op string, err error) {
	s.writeErrMu.Lock()
	if s.writeErrState == nil {
		s.writeErrState = make(map[string]*analyticsWriteOpError)
	}
	st, ok := s.writeErrState[op]
	if !ok {
		st = &analyticsWriteOpError{}
		s.writeErrState[op] = st
	}
	st.suppressed++
	now := time.Now()
	if st.active && now.Sub(st.lastLog) < analyticsErrorLogInterval {
		s.writeErrMu.Unlock()
		return
	}
	suppressed := st.suppressed - 1
	st.suppressed = 0
	st.active = true
	st.lastLog = now
	s.writeErrMu.Unlock()

	log.Error().Err(err).
		Str("component", "analytics").
		Str("op", op).
		Int("omitidos_desde_la_ultima_linea", suppressed).
		Msg("Error escribiendo un evento de telemetría (la petición NO se vio afectada)")
}

// reportWriteOK avisa, una sola vez por operación, que la telemetría volvió a
// escribir después de un fallo: sin esto, un incidente que se cisara se vería
// igual que uno que sigue abierto.
func (s *AnalyticsService) reportWriteOK(op string) {
	s.writeErrMu.Lock()
	st, ok := s.writeErrState[op]
	if !ok || !st.active {
		s.writeErrMu.Unlock()
		return
	}
	st.active = false
	s.writeErrMu.Unlock()

	log.Warn().
		Str("component", "analytics").
		Str("op", op).
		Msg("La telemetría volvió a escribir después de fallos previos")
}

// clamp recorta s a max caracteres runes (no bytes) para no partir un carácter
// UTF-8 por la mitad: Postgres rechaza un varchar con UTF-8 inválido, así que un
// recorte a bytes convertiría un 22001 en otro 22001. Recortar es preferible a
// rechazar el evento: una ruta o un término de búsqueda truncados siguen siendo
// métricas útiles, y el evento perdido no lo es.
func clamp(s string, max int) string {
	if s == "" {
		return ""
	}
	// Atajo rápido: si ya cabe en bytes, cabe en runes.
	if len(s) <= max {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

// clampPageView deja la visita dentro de los anchos de columna de page_views.
//
// Path, Method, SessionID y Referer los pone el cliente (el middleware los toma
// de la petición y de la cookie _sid, que es un valor arbitrario que cualquiera
// puede escribir) y ninguno tiene un límite upstream garantizado: una URL larga,
// un Referer enorme o un _sid manipulado bastaban para que el INSERT muriera con
// 22001. Recortarlos es correcto porque el valor recortado sigue siendo un dato
// útil.
//
// IPHash NO se recorta, a propósito: si algún día la huella no midiera 64, un
// recortearía en silencio y devolvería una huella VÁLIDA pero incorrecta (dos IPs
// distintas con el mismo prefijo), que es peor que un insert fallido. El 22001
// resultante lo reporta reportWriteError, y el ancho real lo fija
// TestAnalyticsService_FingerprintIP y TestAnalyticsRepo_AnchosCoherentesConLaMigracion.
func clampPageView(v domain.PageView) domain.PageView {
	v.Path = clamp(v.Path, analyticsMaxPath)
	v.Method = clamp(v.Method, analyticsMaxMethod)
	v.SessionID = clamp(v.SessionID, analyticsMaxSessionID)
	v.Referer = clamp(v.Referer, analyticsMaxReferer)
	return v
}

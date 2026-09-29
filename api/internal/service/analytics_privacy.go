// api/internal/service/analytics_privacy.go

// Privacidad de la telemetría de visitantes.
//
// La telemetría del sitio público (page_views, search_events, profile_views)
// registra datos de quien NO ha iniciado sesión. Publicar eso en los Términos
// y Condiciones obliga a que el código haga lo que el texto promete, así que
// aquí se concentran las dos reducciones que lo permiten:
//
//   - La dirección IP no se guarda: se guarda su HUELLA (SHA-256 con sal del
//     servidor). La huella sirve para contar visitantes distintos y no para leer
//     la dirección. NO es anonimización: una IPv4 tiene 2^32 valores, así que un
//     atacante con la base podría revertarla por fuerza bruta. Por eso la
//     retención (ANALYTICS_RETENTION_DAYS, 90 días) es la que acota de verdad el
//     daño, y por eso el texto legal dice "huella", nunca "anónimo".
//   - La cabecera Referer no se guarda completa: se guarda solo el origen
//     (esquema://host). La URL completa llega con los parámetros de campañas,
//     buscadores y redes sociales, es decir datos de terceros que el Colegio no
//     necesita para contar visitas.
//
// OJO — alcance deliberado: el login_event y el active_session NO se tocan. Son
// bitácora de seguridad de cuentas con sesión (admin y agremiados) y el panel los
// usa; se declaran aparte en la sección de seguridad de los términos.
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

// analyticsDefaultIPSalt es la sal por defecto cuando ANALYTICS_IP_SALT no está
// definida. No es un secreto: sirve para que la huella sea reproducible entre
// reinicios (que es lo que mantiene válidos los conteos de visitantes únicos),
// no para protegerla. Quien quiera una defensa adicional la define por entorno.
const analyticsDefaultIPSalt = "colpsi-analytics-ipsalt-v1"

// resolveIPSalt decide la sal con la que se calculan las huellas de IP.
func resolveIPSalt() string {
	if config.Envs != nil {
		if s := strings.TrimSpace(config.Envs.AnalyticsIPSalt); s != "" {
			return s
		}
	}
	return analyticsDefaultIPSalt
}

// fingerprintIP convierte una dirección IP en su huella, o "" si no hay
// dirección. Nunca devuelve la dirección original.
func (s *AnalyticsService) fingerprintIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(s.ipSalt + "|" + ip))
	return hex.EncodeToString(sum[:])
}

// refererOrigin reduce un Referer a su origen (esquema://host). Devuelve "" si
// no es una URL con host: lo que no se puede reducir, no se guarda.
func refererOrigin(referer string) string {
	referer = strings.TrimSpace(referer)
	if referer == "" {
		return ""
	}
	u, err := url.Parse(referer)
	if err != nil || u.Host == "" {
		return ""
	}
	scheme := u.Scheme
	if scheme == "" {
		return u.Host
	}
	return scheme + "://" + u.Host
}

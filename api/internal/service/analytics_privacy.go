// api/internal/service/analytics_privacy.go

// Privacidad de la telemetría de visitantes.
//
// La telemetría del sitio público (page_views, search_events, profile_views)
// registra datos de quien NO ha iniciado sesión. Publicar eso en los Términos
// y Condiciones obliga a que el código haga lo que el texto promete, así que
// aquí se concentran las dos reducciones que lo permiten:
//
//   - La dirección IP no se guarda: se guarda su HUELLA (SHA-256 con sal del
//     servidor) en la columna ip_hash de page_views, search_events y
//     profile_views. La huella sirve para contar visitantes distintos y no para
//     leer la dirección. NO es anonimización: una IPv4 tiene 2^32 valores, así que
//     un atacante con la base podría revertarla por fuerza bruta. Por eso la
//     retención (ANALYTICS_RETENTION_DAYS, 90 días) es la que acota de verdad el
//     daño, y por eso el texto legal dice "huella", nunca "anónimo". El ancho de
//     esa columna es 64 (el hex completo): 45 era el largo máximo de una IPv6 y
//     hacía fallar TODOS los inserts — ver migrations/20261001100000.
//   - La cabecera Referer no se guarda completa: se guarda solo el origen
//     (esquema://host). La URL completa llega con los parámetros de campañas,
//     buscadores y redes sociales, es decir datos de terceros que el Colegio no
//     necesita para contar visitas.
//
// OJO — alcance deliberado: el login_event y el active_session NO se tocan. Son
// bitácora de seguridad de cuentas con sesión (admin y agremiados) y el panel los
// usa; se declaran aparte en la sección de seguridad de los términos. Ahí la IP
// sigue en claro, y por eso su columna sigue siendo varchar(45). Lo mismo con
// psi_terms_acceptance.ip, que es el registro de quién aceptó los Términos y desde
// dónde. Aplicar la huella ahí sería falsear una prueba jurídica.
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

// FingerprintIP convierte una dirección IP en su huella, o "" si no hay
// dirección. Nunca devuelve la dirección original.
//
// Devuelve 64 hex exactos: eso es lo que cabe en la columna ip_hash
// (character varying(64), migración 20261001100000_analytics_ip_hash.sql) y lo que
// fija TestAnalyticsService_fingerprintIP. Si algún día cambiara el algoritmo, ese
// ancho dejaría de alcanzar y TODAS las escrituras de telemetría volverían a
// fallar en silencio, que es exactamente el bug que esa columna corrigió.
//
// Es pública porque es un contrato de dos puntas: la misma fórmula tiene que
// producir el histórico que documenta la migración (backfill en SQL con
// sha256(sal || '|' || ip)) y lo que escribe el servicio. La equivalencia entre
// ambas la fija TestAnalyticsRepo_HashSQLDelBackfillEquivaleAFingerprint.
func (s *AnalyticsService) FingerprintIP(ip string) string {
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

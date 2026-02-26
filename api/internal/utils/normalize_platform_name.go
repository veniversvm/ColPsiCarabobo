// api/internal/utils/normalize_platform_name.go
// Package utils provee herramientas transversales de soporte para la aplicación.
package utils

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// =========================================================================
// NORMALIZACIÓN DE REDES SOCIALES
// =========================================================================

// Definimos el caser globalmente para evitar asignaciones de memoria repetitivas
// y garantizar el máximo rendimiento en procesos de carga masiva.
var titleCaser = cases.Title(language.Und)

// platformVariants mapea una amplia gama de variantes, abreviaturas y errores
// ortográficos comunes a su forma canónica profesional.
// Esto permite que la base de datos mantenga una estética uniforme y coherente.
var platformVariants = map[string]string{
	"instagram": "Instagram", "ig": "Instagram", "insta": "Instagram", "instagran": "Instagram", "instgram": "Instagram",
	"facebook": "Facebook", "fb": "Facebook", "face": "Facebook", "facbook": "Facebook", "facebok": "Facebook", "fbk": "Facebook",
	"twitter": "X (Twitter)", "x": "X (Twitter)", "tw": "X (Twitter)", "twiter": "X (Twitter)", "twiiter": "X (Twitter)", "twttr": "X (Twitter)",
	"youtube": "YouTube", "yt": "YouTube", "yutube": "YouTube", "ytb": "YouTube", "youtub": "YouTube", "tube": "YouTube", "youtu.be": "YouTube", "youtube.com": "YouTube",
	"linkedin": "LinkedIn", "in": "LinkedIn", "linkdin": "LinkedIn", "lnkd": "LinkedIn",
	"tiktok": "TikTok", "tk": "TikTok", "ticktok": "TikTok", "tictok": "TikTok",
	"whatsapp": "WhatsApp", "wa": "WhatsApp", "wsp": "WhatsApp", "watsapp": "WhatsApp", "whatsap": "WhatsApp", "wa.me": "WhatsApp",
	"snapchat": "Snapchat", "snap": "Snapchat", "sc": "Snapchat", "snapc": "Snapchat",
	"pinterest": "Pinterest", "pin": "Pinterest", "pint": "Pinterest", "pinterst": "Pinterest",
	"reddit": "Reddit", "rd": "Reddit", "redit": "Reddit",
	"telegram": "Telegram", "tg": "Telegram", "t.me": "Telegram", "tele": "Telegram",
	"discord": "Discord", "dc": "Discord",
	"twitch": "Twitch", "twitchtv": "Twitch",
	"signal": "Signal", "wechat": "WeChat", "wc": "WeChat", "line": "Line", "viber": "Viber",
}

// NormalizePlatformName estandariza nombres de plataformas sociales y URLs básicas.
//
// LÓGICA DE PROCESAMIENTO:
// 1. Limpieza: Elimina espacios y convierte a minúsculas.
// 2. Diccionario: Busca en el mapa de variantes para una resolución instantánea O(1).
// 3. Patrones: Si no hay coincidencia exacta, busca sub-cadenas (útil para URLs pegadas).
// 4. Fallback: Si es un nombre desconocido, aplica formato Título (Ej: "Threads" -> "Threads").
func NormalizePlatformName(name string) string {
	// 1. Limpieza básica y normalización de caracteres
	clean := strings.ToLower(strings.TrimSpace(name))
	lookup := strings.ReplaceAll(clean, " ", "")

	// 2. Búsqueda directa en el mapa (Rendimiento optimizado)
	if normalized, ok := platformVariants[lookup]; ok {
		return normalized
	}

	// 3. Detección por coincidencia de cadena (Manejo de URLs o typos largos)
	// Esta sección previene que variaciones no mapeadas de dominios se pierdan.
	if strings.Contains(lookup, "youtu") {
		return "YouTube"
	}
	if strings.Contains(lookup, "instagr") {
		return "Instagram"
	}
	if strings.Contains(lookup, "facebo") || strings.Contains(lookup, "fb.com") {
		return "Facebook"
	}

	// 4. Fallback: Uso del caser global para nombres de plataformas emergentes o no listadas.
	if len(clean) > 0 {
		return titleCaser.String(clean)
	}

	return clean
}

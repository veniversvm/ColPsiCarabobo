package utils

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Definimos el caser globalmente para máximo rendimiento
var titleCaser = cases.Title(language.Und)

// platformVariants mapea variantes de nombres a su forma canónica.
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
func NormalizePlatformName(name string) string {
	// 1. Limpieza básica
	clean := strings.ToLower(strings.TrimSpace(name))
	lookup := strings.ReplaceAll(clean, " ", "")

	// 2. Búsqueda directa en el mapa (O(1))
	if normalized, ok := platformVariants[lookup]; ok {
		return normalized
	}

	// 3. Detección por coincidencia de cadena (URLs o errores largos)
	if strings.Contains(lookup, "youtu") {
		return "YouTube"
	}
	if strings.Contains(lookup, "instagr") {
		return "Instagram"
	}
	if strings.Contains(lookup, "facebo") || strings.Contains(lookup, "fb.com") {
		return "Facebook"
	}

	// 4. Fallback: Uso del caser global para nombres desconocidos
	if len(clean) > 0 {
		return titleCaser.String(clean)
	}

	return clean
}

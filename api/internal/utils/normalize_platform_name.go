// api/internal/utils/normalize_platform_name.go

// Package utils provee herramientas transversales de soporte para la aplicación.
//
// Esta sub-capa actúa como el motor de Calidad de Datos (Data Quality).
// Protege a la base de datos de la entropía generada por el texto libre introducido
// por los usuarios, estandarizando los formatos antes de su persistencia.
package utils

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// =========================================================================
// NORMALIZACIÓN DE REDES SOCIALES (DATA CANONICALIZATION)
// =========================================================================

// Optimización de Memoria (Global Allocation):
// Definimos el 'caser' globalmente en lugar de instanciarlo dentro de la función.
// Al inicializar `cases.Title` una sola vez en el arranque de la aplicación,
// evitamos asignaciones dinámicas repetitivas en el Heap de memoria. Esto reduce
// drásticamente la presión sobre el Garbage Collector (GC) durante procesos de
// carga masiva (ej. al importar miles de psicólogos desde un Excel).
var titleCaser = cases.Title(language.Und)

// platformVariants implementa el patrón de Diccionario en Memoria (Lookup Table).
//
// Propósito Arquitectónico (UI Consistency):
// Mapea una amplia gama de variantes, abreviaturas (ej. "ig") y errores ortográficos
// (ej. "instagran") a una única forma canónica profesional ("Instagram").
// Esto garantiza que el Frontend siempre reciba el mismo string exacto,
// permitiéndole mapear la palabra a un ícono vectorial (SVG) sin condicionales complejos.
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

// NormalizePlatformName estandariza los nombres de las plataformas sociales.
// Actúa como un Pipeline (Tubería) de 4 fases para la resolución del dato.
//
// Complejidad Algorítmica:
// Prioriza la resolución en tiempo constante O(1) usando Hash Maps, recurriendo a
// escaneos lineales O(N) solo como un método defensivo (Fallback).
func NormalizePlatformName(name string) string {
	// 1. Limpieza (Sanitización Base):
	// Minimiza la varianza inicial eliminando capitalizaciones caprichosas
	// y removiendo por completo los espacios en blanco inyectados por error.
	clean := strings.ToLower(strings.TrimSpace(name))
	lookup := strings.ReplaceAll(clean, " ", "")

	// 2. Resolución O(1) (Lookup Table):
	// Búsqueda directa en el mapa hash en memoria. Es la ruta más rápida y procesa
	// el 95% de los casos de uso esperados de manera instantánea.
	if normalized, ok := platformVariants[lookup]; ok {
		return normalized
	}

	// 3. Heurística Defensiva (Pattern Matching):
	// Si el usuario intentó pegar una URL cruda o cometió un error tipográfico
	// demasiado largo que escapa al mapa (ej. "https://www.instagr.am/mi_perfil"),
	// se detecta la huella dactilar de la cadena mediante sub-strings.
	if strings.Contains(lookup, "youtu") {
		return "YouTube"
	}
	if strings.Contains(lookup, "instagr") {
		return "Instagram"
	}
	if strings.Contains(lookup, "facebo") || strings.Contains(lookup, "fb.com") {
		return "Facebook"
	}

	// 4. Degradación Elegante (Graceful Degradation / Fallback):
	// Si el usuario introduce una red social totalmente nueva o emergente
	// (ej. "Bluesky" o "Mastodon") que no existe en el diccionario, el sistema
	// no lanza error ni destruye el dato. Aplica la capitalización universal
	// (`titleCaser`) para que, al menos, se almacene de forma visualmente estética.
	if len(clean) > 0 {
		return titleCaser.String(clean)
	}

	// Retorno vacío en caso de que la entrada solo contuviera espacios o caracteres nulos.
	return clean
}

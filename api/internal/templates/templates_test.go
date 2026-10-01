package templates

import (
	"bytes"
	"html/template"
	"io/fs"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// renderTemplate ejecuta una plantilla embebida con datos de ejemplo y retorna
// el HTML resultante. Falla si la plantilla no parsea o no ejecuta (regresión
// de sintaxis/HTML en los correos institucionales).
func renderTemplate(t *testing.T, name string, data interface{}) string {
	t.Helper()
	tmpl, err := template.ParseFS(FS, name)
	if err != nil {
		t.Fatalf("ParseFS(%q) falló: %v", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Execute(%q) falló: %v", name, err)
	}
	return buf.String()
}

func TestTemplates_RenderConDatosDeEjemplo(t *testing.T) {
	data := map[string]interface{}{
		"Name":      "Psicólogo de Prueba",
		"Email":     "psicologo@ejemplo.com",
		"Password":  "Temporal2026!",
		"LoginTime": "Fri, 01 Aug 2026 10:00:00 -0400",
		"SiteURL":   "https://colpsicarabobo.com",
		"Signature": "Administración ColPsiCarabobo",
	}

	for _, name := range []string{"login_psi.html", "welcome_psi.html", "login_admin.html", "welcome_admin.html", "reset_password_psi.html"} {
		out := renderTemplate(t, name, data)

		// Branding institucional: letra Ψ, nombre del colegio y bandera de Carabobo.
		for _, expected := range []string{"Ψ", "Colegio de Psicólogos", "del Estado Carabobo", "#1e3a8a", "#991b1b", "#15803d"} {
			if !strings.Contains(out, expected) {
				t.Errorf("%s no contiene el branding %q", name, expected)
			}
		}

		// Bandera real del estado Carabobo servida desde el dominio del remitente
		// (web/public/bandera-carabobo.png), no enlazada a un tercero.
		if !strings.Contains(out, "https://colpsicarabobo.com/bandera-carabobo.png") {
			t.Errorf("%s no incluye la bandera de Carabobo servida desde el dominio de la aplicación", name)
		}

		// Variables del mensaje inyectadas correctamente.
		for _, v := range []string{"Psicólogo de Prueba", "psicologo@ejemplo.com"} {
			if !strings.Contains(out, v) {
				t.Errorf("%s no inyectó el dato %q", name, v)
			}
		}

		// Firma y URL de la aplicación configurable desde .env.
		if !strings.Contains(out, "Administración ColPsiCarabobo") {
			t.Errorf("%s no inyectó la firma institucional", name)
		}
		if !strings.Contains(out, "https://colpsicarabobo.com") {
			t.Errorf("%s no inyectó la URL de la aplicación", name)
		}
	}
}

// TestTemplates_SinRecursosDeTerceros es el contrato de entregabilidad de los
// correos: ningún enlace ni recurso puede apuntar a un host ajeno al dominio de
// la aplicación, y ninguna imagen puede ser SVG.
//
// Por qué es un test y no una revisión manual: los filtros de spam (Gmail entre
// ellos) penalizan los enlaces que no coinciden con el dominio de envío y las
// imágenes alojadas fuera; además Gmail y Outlook no renderizan SVG, así que una
// bandera en SVG era invisible aunque el aviso pasara. Antes de este test la
// bandera venía de upload.wikimedia.org y el pie de un https://franhsabt-…lat
// hardcodeado, y ninguna de las dos cosas había roto nada visible: el correo se
// enviaba igual, solo bajaba la reputación del dominio de envío.
func TestTemplates_SinRecursosDeTerceros(t *testing.T) {
	const siteURL = "https://colegio-psicologos-carabobo.com"

	parsed, err := url.Parse(siteURL)
	if err != nil {
		t.Fatalf("url.Parse(%q) falló: %v", siteURL, err)
	}

	data := map[string]interface{}{
		"Name":      "Psicólogo de Prueba",
		"Email":     "psicologo@ejemplo.com",
		"Password":  "Temporal2026!",
		"LoginTime": "Fri, 01 Aug 2026 10:00:00 -0400",
		"ResetURL":  siteURL + "/reset-password?token=abc123",
		"SiteURL":   siteURL,
		"Signature": "Administración ColPsiCarabobo",
		"Title":     "Aviso de prueba",
		"Message":   "Contenido de prueba",
	}

	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatalf("no se pudo leer el FS embebido: %v", err)
	}

	// Toda URL absoluta del HTML renderizado, incluidos src/href de cualquier tag.
	absoluteURL := regexp.MustCompile(`https?://[^\s"'<>)]+`)

	checked := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}
		checked++

		out := renderTemplate(t, entry.Name(), data)

		for _, raw := range absoluteURL.FindAllString(out, -1) {
			u, perr := url.Parse(raw)
			if perr != nil {
				t.Errorf("%s: URL ilegible %q: %v", entry.Name(), raw, perr)
				continue
			}
			if !strings.EqualFold(u.Host, parsed.Host) {
				t.Errorf("%s enlaza a un host externo al dominio de la aplicación: %q (host %q)", entry.Name(), raw, u.Host)
			}
		}

		// Los clientes de correo no renderizan SVG: una imagen SVG no se ve.
		for _, match := range regexp.MustCompile(`(?i)<img[^>]*>`).FindAllString(out, -1) {
			if strings.Contains(match, ".svg") {
				t.Errorf("%s usa una imagen SVG, que Gmail y Outlook no renderizan: %s", entry.Name(), match)
			}
		}
	}

	if checked == 0 {
		t.Fatal("no se encontró ninguna plantilla embebida; el FS cambió y el test ya no protege nada")
	}
}

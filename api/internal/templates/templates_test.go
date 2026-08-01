package templates

import (
	"bytes"
	"html/template"
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
	}

	for _, name := range []string{"login_psi.html", "welcome_psi.html", "login_admin.html", "welcome_admin.html"} {
		out := renderTemplate(t, name, data)

		// Branding institucional: letra Ψ, nombre del colegio y bandera de Carabobo.
		for _, expected := range []string{"Ψ", "Colegio de Psicólogos", "del Estado Carabobo", "#1e3a8a", "#991b1b", "#15803d"} {
			if !strings.Contains(out, expected) {
				t.Errorf("%s no contiene el branding %q", name, expected)
			}
		}

		// Bandera real del estado Carabobo (SVG de Wikimedia, como el navbar del front).
		if !strings.Contains(out, "upload.wikimedia.org/wikipedia/commons/4/4b/Bandera_de_carabobo.svg") {
			t.Errorf("%s no incluye la imagen de la bandera de Carabobo", name)
		}

		// Variables del mensaje inyectadas correctamente.
		for _, v := range []string{"Psicólogo de Prueba", "psicologo@ejemplo.com"} {
			if !strings.Contains(out, v) {
				t.Errorf("%s no inyectó el dato %q", name, v)
			}
		}
	}
}

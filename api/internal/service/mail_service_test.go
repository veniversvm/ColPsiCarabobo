package service

import (
	"strings"
	"testing"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

// TestResolveSiteURL_ResuelveElDominioDelRemitente fija el contrato de la URL que
// se inyecta en los correos. La razón de que sea un test: durante meses la
// constante.siteURL estuvo fijada a un dominio de pruebas
// (https://franhsabt-testing-ground.lat) y ningún correo falló —simplemente los
// enlaces y las imágenes no coincidían con el dominio de envío, que es lo que los
// filtros de spam revisan para bajar la reputación del remitente.
func TestResolveSiteURL_ResuelveElDominioDelRemitente(t *testing.T) {
	original := config.Envs
	t.Cleanup(func() { config.Envs = original })

	cases := []struct {
		nombre   string
		appURL   string
		esperado string
	}{
		// APP_URL es la fuente de verdad: es el dominio del sitio y coincide con
		// el dominio de envío (SMTP_FROM=…@contact.colegio-psicologos-carabobo.com).
		{"usa APP_URL", "https://colegio-psicologos-carabobo.com", "https://colegio-psicologos-carabobo.com"},
		// Sin barra final: las plantillas componen rutas ("{{.SiteURL}}/bandera-…"),
		// y una barra doble produce "//bandera-…" y un host distinto en algunos clientes.
		{"quita la barra final", "https://colegio-psicologos-carabobo.com/", "https://colegio-psicologos-carabobo.com"},
		{"quita espacios", "  https://colegio-psicologos-carabobo.com  ", "https://colegio-psicologos-carabobo.com"},
		// Un host local no le sirve de nada a quien recibe el correo: cae a producción.
		{"localhost cae a producción", "http://localhost:3000", defaultSiteURL},
		{"127.0.0.1 cae a producción", "http://127.0.0.1:3000", defaultSiteURL},
		{"vacío cae a producción", "", defaultSiteURL},
		{"basura cae a producción", "no-es-una-url", defaultSiteURL},
	}

	for _, tc := range cases {
		t.Run(tc.nombre, func(t *testing.T) {
			config.Envs = &config.Config{AppURL: tc.appURL}

			if got := resolveSiteURL(); got != tc.esperado {
				t.Errorf("resolveSiteURL() con APP_URL=%q = %q, se esperaba %q", tc.appURL, got, tc.esperado)
			}

			// El contrato real: lo que la plantilla compone debe ser una URL del
			// dominio de la aplicación, sin doble barra.
			compuesta := resolveSiteURL() + "/bandera-carabobo.png"
			if !strings.HasPrefix(compuesta, tc.esperado+"/") {
				t.Errorf("la URL compuesta quedó %q, fuera del dominio %q", compuesta, tc.esperado)
			}
		})
	}
}

func TestMailService_Close_ShutdownsWorker(t *testing.T) {
	config.InitConfig()

	ms, err := NewMailService()
	if err != nil {
		t.Fatalf("NewMailService() failed: %v", err)
	}

	// Verify the worker is running by enqueuing a job (it will fail to send
	// but the worker should process it without panicking).
	_ = ms.SendEmail("test@example.com", "test", "welcome", nil)

	// Give the worker time to pick up the job
	time.Sleep(200 * time.Millisecond)

	// Close — should cancel context and close channel without panic
	ms.Close()

	// After Close, SendEmail should fail because the channel is closed
	// Give a tiny window for close to propagate
	time.Sleep(10 * time.Millisecond)

	err = ms.SendEmail("test@example.com", "test", "welcome", nil)
	if err == nil {
		t.Error("SendEmail after Close() should have returned an error")
	}
}

func TestMailService_NewMailService_InitialState(t *testing.T) {
	config.InitConfig()

	ms, err := NewMailService()
	if err != nil {
		t.Fatalf("NewMailService() failed: %v", err)
	}
	defer ms.Close()

	if ms.client == nil {
		t.Error("client should not be nil")
	}
	if ms.from == "" {
		t.Error("from should not be empty")
	}
	if ms.queue == nil {
		t.Error("queue should not be nil")
	}
	if ms.cancel == nil {
		t.Error("cancel should not be nil")
	}
}

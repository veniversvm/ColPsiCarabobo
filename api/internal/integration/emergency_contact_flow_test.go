// api/internal/integration/emergency_contact_flow_test.go
package integration

// Suite E2E de la Persona de Contacto para Emergencias.
//
// Verifica el flujo HTTP completo (router → middleware → handler → service →
// repo) en sus dos caras —auto-gestión del agremiado y moderación admin— y, sobre
// todo, la PRIVACIDAD: los contactos son datos de un TERCERO y no pueden aparecer
// en ningún endpoint público (directorio ni ficha pública).

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// peticion ejecuta una petición HTTP contra la app de prueba. `body` vacío
// significa "sin payload".
func peticion(t *testing.T, app *fiber.App, method, path, token, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

// leerCrudo devuelve el cuerpo completo de la respuesta como texto plano. Se usa
// para las aserciones de privacidad: buscar la cadena prohibida en el JSON crudo
// es más fuerte que comprobar que falte una clave concreta.
func leerCrudo(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	crudo, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(crudo)
}

func itoa(n int) string { return strconv.Itoa(n) }

func TestEmergencyFlow_PsiSelfManagement(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	psi := seedPsi(t, 210001, 31000001, "psiemergencia")
	app := buildTestApp(testDB)
	token := loginPsi(t, app, "psiemergencia", "Psi123!@#")

	var contactID string

	t.Run("Alta válida → 201", func(t *testing.T) {
		body := `{"name":"María Rodríguez","relationship":"madre","phone":"0412-1234567"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/psi/me/emergency", token, body)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("El contacto llega en GET /psi/me con el teléfono normalizado", func(t *testing.T) {
		resp := peticion(t, app, http.MethodGet, "/api/v1/psi/me/", token, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		lista, ok := data["emergency_contacts"].([]interface{})
		require.True(t, ok, "GET /psi/me debe incluir emergency_contacts")
		require.Len(t, lista, 1)
		contacto := lista[0].(map[string]interface{})
		contactID = contacto["id"].(string)
		require.Equal(t, "María Rodríguez", contacto["name"])
		require.Equal(t, "madre", contacto["relationship"])
		require.Equal(t, "04121234567", contacto["phone"], "el service normaliza el teléfono")
	})

	t.Run("Sin teléfono ni correo → 400 (regla de canal)", func(t *testing.T) {
		body := `{"name":"Juan Pérez","relationship":"padre"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/psi/me/emergency", token, body)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		data := decodeBody(t, resp)
		require.Contains(t, data["error"], "teléfono o correo")
	})

	t.Run("Sin nombre → 400", func(t *testing.T) {
		body := `{"name":"","relationship":"padre","phone":"04121234567"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/psi/me/emergency", token, body)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Correo inválido → 400", func(t *testing.T) {
		body := `{"name":"Juan","relationship":"padre","email":"esto-no-es-un-correo"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/psi/me/emergency", token, body)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("PATCH válido → 200", func(t *testing.T) {
		body := `{"relationship":"esposa"}`
		resp := peticion(t, app, http.MethodPatch, "/api/v1/psi/me/emergency/"+contactID, token, body)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("PATCH que elimina el único canal → 400", func(t *testing.T) {
		body := `{"phone":"","email":""}`
		resp := peticion(t, app, http.MethodPatch, "/api/v1/psi/me/emergency/"+contactID, token, body)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Migración de canal: teléfono → correo → 200", func(t *testing.T) {
		body := `{"phone":"","email":"nuevo@correo.com"}`
		resp := peticion(t, app, http.MethodPatch, "/api/v1/psi/me/emergency/"+contactID, token, body)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp = peticion(t, app, http.MethodGet, "/api/v1/psi/me/", token, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		contacto := data["emergency_contacts"].([]interface{})[0].(map[string]interface{})
		require.Equal(t, "", contacto["phone"])
		require.Equal(t, "nuevo@correo.com", contacto["email"])
	})

	t.Run("IDOR: otro agremiado no puede editar el contacto → 403", func(t *testing.T) {
		otro := seedPsi(t, 210002, 31000002, "psiemergencia2")
		otroToken := loginPsi(t, app, "psiemergencia2", "Psi123!@#")
		require.NotEqual(t, otro.ID, psi.ID)

		body := `{"name":"Intruso"}`
		resp := peticion(t, app, http.MethodPatch, "/api/v1/psi/me/emergency/"+contactID, otroToken, body)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
		data := decodeBody(t, resp)
		require.NotContains(t, data["error"], "María", "el error no debe filtrar datos del tercero")
	})

	t.Run("Cuota: el 4º contacto se rechaza → 403", func(t *testing.T) {
		// Ya hay 1; se llega a 3 y el cuarto debe rebotar.
		for i := 0; i < 2; i++ {
			body := `{"name":"Extra","relationship":"amigo","phone":"04120000000"}`
			resp := peticion(t, app, http.MethodPost, "/api/v1/psi/me/emergency", token, body)
			require.Equal(t, http.StatusCreated, resp.StatusCode)
		}
		body := `{"name":"Cuarto","relationship":"amigo","phone":"04120000000"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/psi/me/emergency", token, body)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
		data := decodeBody(t, resp)
		require.Contains(t, data["error"], "límite máximo")
	})

	t.Run("Sin token → 401", func(t *testing.T) {
		body := `{"name":"Sin Token","relationship":"amigo","phone":"04120000000"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/psi/me/emergency", "", body)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("DELETE → 200 y desaparece de la lista", func(t *testing.T) {
		resp := peticion(t, app, http.MethodDelete, "/api/v1/psi/me/emergency/"+contactID, token, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp = peticion(t, app, http.MethodGet, "/api/v1/psi/me/", token, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		lista := data["emergency_contacts"].([]interface{})
		require.Len(t, lista, 2, "solo debe quedar lo que no se borró")
		for _, item := range lista {
			require.NotEqual(t, contactID, item.(map[string]interface{})["id"])
		}
	})
}

func TestEmergencyFlow_Admin(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	psi := seedPsi(t, 210003, 31000003, "psiadminemerg")
	otroPsi := seedPsi(t, 210004, 31000004, "psiadminotro")
	app := buildTestApp(testDB)

	sudoToken := loginAdmin(t, app, "sudo", "Sudo123!@#")

	var contactID string

	t.Run("Sin token de admin → 404 enmascarado", func(t *testing.T) {
		body := `{"name":"María","relationship":"madre","phone":"04121234567"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/admin/psi/"+psi.ID.String()+"/emergency", "", body)
		require.Equal(t, http.StatusNotFound, resp.StatusCode, "ProtectedAdmin404 enmascara como 404")
	})

	t.Run("Sudo registra un contacto obtenido por el Colegio → 201", func(t *testing.T) {
		body := `{"name":"María Rodríguez","relationship":"madre","phone":"0412-1234567"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/admin/psi/"+psi.ID.String()+"/emergency", sudoToken, body)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("El contacto llega en la ficha del admin (GET /admin/psi/:id)", func(t *testing.T) {
		resp := peticion(t, app, http.MethodGet, "/api/v1/admin/psi/"+psi.ID.String(), sudoToken, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		lista, ok := data["emergency_contacts"].([]interface{})
		require.True(t, ok, "la ficha del admin debe incluir emergency_contacts")
		require.Len(t, lista, 1)
		contactID = lista[0].(map[string]interface{})["id"].(string)
	})

	t.Run("PATCH del admin → 200", func(t *testing.T) {
		body := `{"name":"María Rodríguez Costa","email":"maria@correo.com"}`
		resp := peticion(t, app, http.MethodPatch, "/api/v1/admin/psi/"+psi.ID.String()+"/emergency/"+contactID, sudoToken, body)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("IDOR: contacto de otra ficha → 403", func(t *testing.T) {
		body := `{"name":"Secreto"}`
		resp := peticion(t, app, http.MethodPatch, "/api/v1/admin/psi/"+otroPsi.ID.String()+"/emergency/"+contactID, sudoToken, body)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("RBAC: admin sin permisos → 403", func(t *testing.T) {
		seedAdmin(t, "lector-emerg@test.com", map[string]bool{})
		lectorToken := loginAdmin(t, app, "lector-emerg@test.com", "Admin123!@#")
		body := `{"name":"No autorizado","relationship":"amigo","phone":"04120000000"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/admin/psi/"+psi.ID.String()+"/emergency", lectorToken, body)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("RBAC: admin con CanUpdatePsi sí puede", func(t *testing.T) {
		seedAdmin(t, "secretaria-emerg@test.com", map[string]bool{"can_update_psi": true})
		secToken := loginAdmin(t, app, "secretaria-emerg@test.com", "Admin123!@#")
		body := `{"name":"Otro contacto","relationship":"hijo","phone":"04120000000"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/admin/psi/"+psi.ID.String()+"/emergency", secToken, body)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("ID no parseable en la ruta → 400", func(t *testing.T) {
		body := `{"name":"X","relationship":"amigo","phone":"04120000000"}`
		resp := peticion(t, app, http.MethodPost, "/api/v1/admin/psi/no-es-uuid/emergency", sudoToken, body)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("DELETE del admin → 200", func(t *testing.T) {
		resp := peticion(t, app, http.MethodDelete, "/api/v1/admin/psi/"+psi.ID.String()+"/emergency/"+contactID, sudoToken, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp = peticion(t, app, http.MethodGet, "/api/v1/admin/psi/"+psi.ID.String(), sudoToken, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		lista := data["emergency_contacts"].([]interface{})
		require.Len(t, lista, 1, "solo queda el segundo contacto")
	})
}

func TestEmergencyFlow_PrivacidadPublica(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	psi := seedPsi(t, 210005, 31000005, "psiprivado")
	app := buildTestApp(testDB)

	// Se inyecta un contacto directamente en la base de datos: la garantía de
	// privacidad no puede depender de que el endpoint lo haga bien, sino de que
	// NINGÚN endpoint público lo lea.
	require.NoError(t, testDB.Create(&domain.PsiUserEmergencyContact{
		ID:           uuid.Must(uuid.NewV7()),
		PsiUserID:    psi.ID,
		Name:         "Tercero Confidencial",
		Relationship: "madre",
		Phone:        "04129998888",
		Email:        "confidencial@correo.com",
	}).Error)

	t.Run("El directorio público no lo expone", func(t *testing.T) {
		resp := peticion(t, app, http.MethodGet, "/api/v1/psi/directory?q=Garcia", "", "")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		crudo := leerCrudo(t, resp)
		require.NotContains(t, crudo, "emergenc")
		require.NotContains(t, crudo, "Confidencial")
		require.NotContains(t, crudo, "04129998888")
		require.NotContains(t, crudo, "confidencial@correo.com")
	})

	t.Run("La ficha pública no lo expone", func(t *testing.T) {
		resp := peticion(t, app, http.MethodGet, "/api/v1/psi/"+itoa(psi.FPV), "", "")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		crudo := leerCrudo(t, resp)
		require.NotContains(t, crudo, "emergenc")
		require.NotContains(t, crudo, "Confidencial")
		require.NotContains(t, crudo, "04129998888")
		require.NotContains(t, crudo, "confidencial@correo.com")
	})

	t.Run("El sitemap público no lo expone", func(t *testing.T) {
		resp := peticion(t, app, http.MethodGet, "/api/v1/psi/public/sitemap-data", "", "")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		crudo := leerCrudo(t, resp)
		require.NotContains(t, crudo, "emergenc")
		require.NotContains(t, crudo, "Confidencial")
	})
}

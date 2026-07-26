package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPsiFlow_AdminCRUD(t *testing.T) {
	truncateAll(t)
	sudo := seedSudo(t)
	app := buildTestApp(testDB)
	_ = sudo

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	var psiID string
	t.Run("AdminCreatePsi", func(t *testing.T) {
		body := `{
			"username":"psicreato",
			"email":"psicreato@test.com",
			"password":"Psi123!@#",
			"first_name":"Maria",
			"last_name":"Garcia",
			"ci":20000001,
			"fpv":100001,
			"nationality":"V",
			"genre":"F",
			"born_date":"1990-01-01",
			"solvent":true,
			"proof_of_life":true,
			"is_active":true,
			"contact_phone":"+58-241-1234567",
			"contact_cell_phone":"+58-412-1234567"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/psi/create", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("AdminListPsi", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/psi/list", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		results := data["data"].([]interface{})
		require.Len(t, results, 1)
		psiMap := results[0].(map[string]interface{})
		psiID = psiMap["id"].(string)
	})

	t.Run("AdminGetPsiByID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/psi/"+psiID, nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AdminDeletePsi", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/psi/"+psiID, nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestPsiFlow_SelfManagement(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	psi := seedPsi(t, 200001, 30000001, "psiself")
	app := buildTestApp(testDB)

	token := loginPsi(t, app, "psiself", "Psi123!@#")

	t.Run("GetMe", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/me/", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		require.Equal(t, "psiself", data["username"])
	})

	t.Run("UpdateOwnProfile", func(t *testing.T) {
		body := `{"password":"Psi123!@#","contact_phone":"+58-241-9999999"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/psi/me/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Logout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/psi/me/logout", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("TokenInvalidAfterLogout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/me/", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	_ = psi
}

func TestPsiFlow_SocialNetworks(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	psi := seedPsi(t, 200002, 30000002, "psisocial")
	app := buildTestApp(testDB)

	token := loginPsi(t, app, "psisocial", "Psi123!@#")

	var netID string
	t.Run("AddSocialNetwork", func(t *testing.T) {
		body := `{"name":"Instagram","url":"https://instagram.com/test"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/psi/me/social", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("ListSocialNetworksViaGetMe", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/me/", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		nets := data["social_networks"].([]interface{})
		require.Len(t, nets, 1)
		netMap := nets[0].(map[string]interface{})
		netID = netMap["id"].(string)
	})

	t.Run("UpdateSocialNetwork", func(t *testing.T) {
		body := `{"url":"https://instagram.com/updated"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/psi/me/social/"+netID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeleteSocialNetwork", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/psi/me/social/"+netID, nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	_ = psi
}

func TestPsiFlow_PostGrades(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	psi := seedPsi(t, 200003, 30000003, "psipostgrade")
	app := buildTestApp(testDB)

	token := loginPsi(t, app, "psipostgrade", "Psi123!@#")

	t.Run("AddPostGrade", func(t *testing.T) {
		body := `{"title":"Maestria en Psicologia","university":"UC","graduation_year":2020,"type":"maestria"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/psi/me/postgrades", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	_ = psi
}

func TestPsiFlow_TokenInvalidation(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	psi := seedPsi(t, 200004, 30000004, "psiinvalidate")
	app := buildTestApp(testDB)

	token := loginPsi(t, app, "psiinvalidate", "Psi123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/me/", nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	_, err = app.Test(req)
	require.NoError(t, err)

	body := fmt.Sprintf(`{"password":"Psi123!@#","new_password_1":"NewPsi456!@#","new_password_2":"NewPsi456!@#"}`)
	req2 := httptest.NewRequest(http.MethodPatch, "/api/v1/psi/me/", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", authHeader(token))
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/psi/me/", nil)
	req3.Header.Set("Authorization", authHeader(token))
	resp3, err := app.Test(req3)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp3.StatusCode)

	_ = psi
}

func TestPsiFlow_LoginInvalidCredentials(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedPsi(t, 200005, 30000005, "psibadlogin")
	app := buildTestApp(testDB)

	body := `{"identifier":"psibadlogin","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/psi/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

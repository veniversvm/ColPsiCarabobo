package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminFlow_FullLifecycle(t *testing.T) {
	truncateAll(t)
	sudo := seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	t.Run("CreateAdmin", func(t *testing.T) {
		body := `{"username":"newadmin@test.com","email":"newadmin@test.com","password":"Strong123!@#","permissions":{"can_create_psi":true}}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("ListAdmins", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		total := int64(data["total"].(float64))
		require.GreaterOrEqual(t, total, int64(2))
	})

	t.Run("UpdateAdmin", func(t *testing.T) {
		var newAdminID string
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, _ := app.Test(req)
		data := decodeBody(t, resp)
		admins := data["data"].([]interface{})
		for _, a := range admins {
			am := a.(map[string]interface{})
			if am["username"] == "newadmin@test.com" {
				newAdminID = am["id"].(string)
				break
			}
		}
		require.NotEmpty(t, newAdminID)

		body := fmt.Sprintf(`{"id":"%s","email":"renamed@test.com"}`, newAdminID)
		req2 := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/update", strings.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", authHeader(token))
		resp2, err := app.Test(req2)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp2.StatusCode)
	})

	t.Run("DeleteAdmin", func(t *testing.T) {
		var targetID string
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, _ := app.Test(req)
		data := decodeBody(t, resp)
		admins := data["data"].([]interface{})
		for _, a := range admins {
			am := a.(map[string]interface{})
			if am["email"] == "renamed@test.com" {
				targetID = am["id"].(string)
				break
			}
		}
		require.NotEmpty(t, targetID)

		req2 := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/delete/"+targetID, nil)
		req2.Header.Set("Authorization", authHeader(token))
		resp2, err := app.Test(req2)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		req3 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
		req3.Header.Set("Authorization", authHeader(token))
		resp3, _ := app.Test(req3)
		data3 := decodeBody(t, resp3)
		total := int64(data3["total"].(float64))
		require.Equal(t, int64(1), total)
	})

	_ = sudo
}

func TestAdminFlow_CannotDeleteSudo(t *testing.T) {
	truncateAll(t)
	sudo := seedSudo(t)
	seedAdmin(t, "deleter", map[string]bool{
		"can_delete_admin": true,
	})
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "deleter", "Admin123!@#")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/delete/"+sudo.ID.String(), nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "Super Usuario")
}

func TestAdminFlow_CannotDeleteSelf(t *testing.T) {
	truncateAll(t)
	sudo := seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/delete/"+sudo.ID.String(), nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "propia cuenta")
}

func TestAdminFlow_HierarchyEnforced(t *testing.T) {
	truncateAll(t)
	seedAdmin(t, "limited@test.com", map[string]bool{
		"can_create_psi": true,
	})
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "limited@test.com", "Admin123!@#")

	body := `{"username":"newadmin@test.com","email":"newadmin@test.com","password":"Strong123!@#"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminFlow_InvalidCredentials(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	body := `{"identifier":"sudo","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAdminFlow_InvalidToken(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAdminFlow_DashboardStats(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminFlow_ListPagination(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedAdmin(t, "admin1@test.com", fullAdminPerms())
	seedAdmin(t, "admin2@test.com", fullAdminPerms())
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list?page=1&limit=2", nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Equal(t, float64(2), data["limit"])
	require.NotNil(t, data["data"])
}

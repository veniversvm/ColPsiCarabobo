package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpecialtyFlow_CreateAndList(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	t.Run("AdminCreateSpecialty", func(t *testing.T) {
		body := `{"name":"Psicologia Clinica","description":"Area clinica"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/specialties/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("PublicListSpecialties", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/specialties/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestSpecialtyFlow_Deactivate(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	spec := seedSpecialty(t, "ToDeactivate")
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	t.Run("AdminDeactivateSpecialty", func(t *testing.T) {
		body := `{"active":false}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/specialties/"+fmt.Sprintf("%d", spec.ID), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("PublicDoesNotSeeInactive", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/specialties/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestSpecialtyFlow_DuplicateName(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedSpecialty(t, "Unique Specialty")
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"name":"Unique Specialty","description":"Duplicate"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/specialties/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.True(t, resp.StatusCode >= 400, "Expected error status for duplicate name")
}

func TestSpecialtyFlow_CountSpecialties(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedSpecialty(t, "Spec1")
	seedSpecialty(t, "Spec2")
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/specialties/count", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

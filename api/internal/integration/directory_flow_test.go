package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectoryFlow_PublicSearch(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedPsi(t, 300001, 40000001, "dirpsi1")
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/directory", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDirectoryFlow_InactiveHidden(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	psi := seedPsi(t, 300002, 40000002, "dirpsi2")
	testDB.Model(psi).Update("is_active", false).Update("solvent", false)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/directory", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := decodeBody(t, resp)
	total := int64(data["total"].(float64))
	require.Equal(t, int64(0), total)
}

func TestDirectoryFlow_SitemapData(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedPsi(t, 300003, 40000003, "sitemappsi")
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/public/sitemap-data", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDirectoryFlow_SearchByName(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedPsi(t, 300004, 40000004, "mariapsi")
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/directory?q=Maria", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

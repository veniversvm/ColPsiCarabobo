package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

func TestNoStore_SetsCacheControl(t *testing.T) {
	app := fiber.New()
	app.Use(NoStore())
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "no-store", resp.Header.Get("Cache-Control"))
}

func TestSecurityHeaders_EmitsHeaders(t *testing.T) {
	config.Envs = &config.Config{HSTSMaxAge: 31536000}
	app := fiber.New()
	app.Use(SecurityHeaders())
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	require.Equal(t, "SAMEORIGIN", resp.Header.Get("X-Frame-Options"))
	require.NotEmpty(t, resp.Header.Get("Permissions-Policy"))
	require.Empty(t, resp.Header.Get("Strict-Transport-Security"),
		"HSTS no debe emitirse sobre HTTP")
}

func TestSecurityHeaders_HSTSOverHTTPS(t *testing.T) {
	config.Envs = &config.Config{HSTSMaxAge: 31536000}
	app := fiber.New()
	app.Use(SecurityHeaders())
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	hsts := resp.Header.Get("Strict-Transport-Security")
	require.Contains(t, hsts, "max-age=31536000")
	require.Contains(t, hsts, "includeSubDomains")
}

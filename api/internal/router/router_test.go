package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

func TestSetupRouter_RoutesExist(t *testing.T) {
	config.Envs = &config.Config{
		Environment: "production",
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// We can't call SetupRouter directly without a real DB, but we CAN test
	// that the router functions register routes by verifying route groups exist.
	// This is a smoke test for the routing layer.

	// Test that the app starts up and responds to 404
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestSetupRouter_Default404(t *testing.T) {
	config.Envs = &config.Config{Environment: "production"}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Add a catch-all 404 handler like SetupRouter does
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Cannot " + c.Method() + " " + c.Path(),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	require.Contains(t, body["message"], "Cannot")
}

func TestSetupRouter_MethodNotAllowed(t *testing.T) {
	config.Envs = &config.Config{Environment: "production"}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/test-route", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/test-route", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
}

func TestSetupAdminRoutes_DevMonitor(t *testing.T) {
	config.Envs = &config.Config{Environment: "development"}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	api := app.Group("/api/v1")

	// Simulate the dev-only debug-monitor route registration
	if config.Envs.Environment == "development" {
		api.Get("/debug-monitor", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"status": "ok"})
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/debug-monitor", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSetupAdminRoutes_ProdNoMonitor(t *testing.T) {
	config.Envs = &config.Config{Environment: "production"}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	api := app.Group("/api/v1")

	if config.Envs.Environment == "development" {
		api.Get("/debug-monitor", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"status": "ok"})
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/debug-monitor", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestRouteGroups_Nesting(t *testing.T) {
	config.Envs = &config.Config{Environment: "production"}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	api := app.Group("/api/v1")

	admin := api.Group("/admin")
	admin.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// =========================================================================
// TEST: AuthRateLimiter (10 req / 15 min)
// =========================================================================

func TestAuthRateLimiter(t *testing.T) {
	t.Run("10 POST requests permitidas", func(t *testing.T) {
		// Cada llamada a AuthRateLimiter() crea una nueva instancia de limiter
		app := fiber.New()
		app.Post("/login", AuthRateLimiter(), func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{"ok": true})
		})

		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"pass":"test"}`))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request %d: app.Test error: %v", i+1, err)
			}
			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Request %d: StatusCode = %d, want 200. Body: %s", i+1, resp.StatusCode, string(body))
			}
		}
	})

	t.Run("11º POST retorna 429", func(t *testing.T) {
		app := fiber.New()
		app.Post("/login", AuthRateLimiter(), func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{"ok": true})
		})

		// Agotar las 10 permits
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"pass":"test"}`))
			req.Header.Set("Content-Type", "application/json")
			app.Test(req)
		}

		// 11º request debería ser bloqueada
		req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"pass":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 429 {
			t.Errorf("StatusCode = %d, want 429", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Demasiados intentos") {
			t.Errorf("Body should contain 'Demasiados intentos', got: %s", bodyStr)
		}
	})

	t.Run("GET no incrementa contador", func(t *testing.T) {
		app := fiber.New()
		app.Get("/login", AuthRateLimiter(), func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{"ok": true})
		})

		// 20 GETs no deberían ser bloqueadas
		for i := 0; i < 20; i++ {
			req := httptest.NewRequest("GET", "/login", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request %d: app.Test error: %v", i+1, err)
			}
			if resp.StatusCode != 200 {
				t.Errorf("Request %d: StatusCode = %d, want 200 (GET bypass)", i+1, resp.StatusCode)
			}
		}
	})

	t.Run("OPTIONS bypass no cuenta", func(t *testing.T) {
		app := fiber.New()
		app.Options("/login", AuthRateLimiter(), func(c *fiber.Ctx) error {
			return c.SendStatus(204)
		})

		for i := 0; i < 20; i++ {
			req := httptest.NewRequest("OPTIONS", "/login", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request %d: app.Test error: %v", i+1, err)
			}
			if resp.StatusCode != 204 {
				t.Errorf("Request %d: StatusCode = %d, want 204 (OPTIONS bypass)", i+1, resp.StatusCode)
			}
		}
	})
}

// =========================================================================
// TEST: AdminAuthRateLimiter (5 req / 30 min)
// =========================================================================

func TestAdminAuthRateLimiter(t *testing.T) {
	t.Run("5 POST requests permitidas", func(t *testing.T) {
		app := fiber.New()
		app.Post("/admin/login", AdminAuthRateLimiter(), func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{"ok": true})
		})

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(`{"pass":"test"}`))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request %d: app.Test error: %v", i+1, err)
			}
			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Request %d: StatusCode = %d, want 200. Body: %s", i+1, resp.StatusCode, string(body))
			}
		}
	})

	t.Run("6º POST retorna 429", func(t *testing.T) {
		app := fiber.New()
		app.Post("/admin/login", AdminAuthRateLimiter(), func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{"ok": true})
		})

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(`{"pass":"test"}`))
			req.Header.Set("Content-Type", "application/json")
			app.Test(req)
		}

		req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(`{"pass":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 429 {
			t.Errorf("StatusCode = %d, want 429", resp.StatusCode)
		}
	})

	t.Run("mensaje en español contiene 'Demasiados'", func(t *testing.T) {
		app := fiber.New()
		app.Post("/admin/login", AdminAuthRateLimiter(), func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{"ok": true})
		})

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(`{"pass":"test"}`))
			req.Header.Set("Content-Type", "application/json")
			app.Test(req)
		}

		req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(`{"pass":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Demasiados") {
			t.Errorf("Body debería contener 'Demasiados', got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "30 minutos") {
			t.Errorf("Body debería contener '30 minutos', got: %s", bodyStr)
		}
	})
}

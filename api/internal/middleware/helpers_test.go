package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// =========================================================================
// TEST: GetAuthenticatedAdmin
// =========================================================================

func TestGetAuthenticatedAdmin(t *testing.T) {
	app := fiber.New()

	t.Run("admin válido retorna admin sin error", func(t *testing.T) {
		adminID := uuid.Must(uuid.NewV7())
		app.Get("/test-admin", func(c *fiber.Ctx) error {
			c.Locals("admin", &domain.UserAdmin{
				ID: adminID,
				Credentials: domain.Credentials{Username: "test_admin"},
			})
			admin, err := GetAuthenticatedAdmin(c)
			if err != nil {
				return err
			}
			return c.JSON(fiber.Map{"id": admin.ID.String(), "username": admin.Username})
		})

		req := httptest.NewRequest("GET", "/test-admin", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if !contains(bodyStr, adminID.String()) {
			t.Errorf("Response body should contain admin ID, got: %s", bodyStr)
		}
	})

	t.Run("admin missing retorna 401", func(t *testing.T) {
		app.Get("/test-admin-missing", func(c *fiber.Ctx) error {
			_, err := GetAuthenticatedAdmin(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
			}
			return nil
		})

		req := httptest.NewRequest("GET", "/test-admin-missing", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if !contains(string(body), "not authenticated") {
			t.Errorf("Body should contain error message, got: %s", string(body))
		}
	})

	t.Run("admin nil retorna 401", func(t *testing.T) {
		app.Get("/test-admin-nil", func(c *fiber.Ctx) error {
			c.Locals("admin", nil)
			_, err := GetAuthenticatedAdmin(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
			}
			return nil
		})

		req := httptest.NewRequest("GET", "/test-admin-nil", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", resp.StatusCode)
		}
	})

	t.Run("tipo incorrecto en locals retorna 401", func(t *testing.T) {
		app.Get("/test-admin-wrong-type", func(c *fiber.Ctx) error {
			c.Locals("admin", "this is a string, not an admin")
			_, err := GetAuthenticatedAdmin(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
			}
			return nil
		})

		req := httptest.NewRequest("GET", "/test-admin-wrong-type", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", resp.StatusCode)
		}
	})
}

// =========================================================================
// TEST: GetAuthenticatedPsi
// =========================================================================

func TestGetAuthenticatedPsi(t *testing.T) {
	app := fiber.New()

	t.Run("psi válido retorna psi sin error", func(t *testing.T) {
		psiID := uuid.Must(uuid.NewV7())
		app.Get("/test-psi", func(c *fiber.Ctx) error {
			c.Locals("psi_user", &domain.PsiUserModel{
				ID: psiID,
				Credentials: domain.Credentials{Username: "test_psi"},
			})
			psi, err := GetAuthenticatedPsi(c)
			if err != nil {
				return err
			}
			return c.JSON(fiber.Map{"id": psi.ID.String(), "username": psi.Username})
		})

		req := httptest.NewRequest("GET", "/test-psi", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if !contains(string(body), psiID.String()) {
			t.Errorf("Response body should contain psi ID, got: %s", string(body))
		}
	})

	t.Run("psi missing retorna 401", func(t *testing.T) {
		app.Get("/test-psi-missing", func(c *fiber.Ctx) error {
			_, err := GetAuthenticatedPsi(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
			}
			return nil
		})

		req := httptest.NewRequest("GET", "/test-psi-missing", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if !contains(string(body), "not authenticated") {
			t.Errorf("Body should contain error message, got: %s", string(body))
		}
	})

	t.Run("psi nil retorna 401", func(t *testing.T) {
		app.Get("/test-psi-nil", func(c *fiber.Ctx) error {
			c.Locals("psi_user", nil)
			_, err := GetAuthenticatedPsi(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
			}
			return nil
		})

		req := httptest.NewRequest("GET", "/test-psi-nil", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", resp.StatusCode)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

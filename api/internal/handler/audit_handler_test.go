package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// mockAuditRepo implementa domain.AuditRepository sin tocar Postgres.
type mockAuditRepo struct {
	listFunc func(ctx context.Context, f domain.AuditLogFilters) ([]domain.ApiChangeLog, int64, error)
}

func (m *mockAuditRepo) CreateBatch(context.Context, []domain.ApiChangeLog) error { return nil }
func (m *mockAuditRepo) List(ctx context.Context, f domain.AuditLogFilters) ([]domain.ApiChangeLog, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return nil, 0, nil
}
func (m *mockAuditRepo) Stats(context.Context, time.Time) ([]domain.AuditStat, error) { return nil, nil }
func (m *mockAuditRepo) PurgeOlderThan(context.Context, time.Time) (int64, error)      { return 0, nil }

// setupAuditApp construye un app Fiber con todas las rutas de auditoría y un
// admin inyectado en c.Locals (nil = sin sesión).
func setupAuditApp(admin *domain.UserAdmin) *fiber.App {
	svc := service.NewAuditService(&mockAuditRepo{})
	h := NewAuditHandler(svc)

	app := newTestApp()
	app.Use(func(c *fiber.Ctx) error {
		if admin != nil {
			c.Locals("admin", admin)
		}
		// admin == nil simula la ausencia de sesión (fallo del middleware de auth).
		return c.Next()
	})
	app.Get("/audit-logs", h.ListLogs)
	app.Get("/audit-logs/psi/:id", h.ListPsiLogs)
	app.Get("/audit-logs/export", h.ExportLogs)
	app.Get("/audit-logs/stats", h.StatsLogs)
	return app
}

func viewerAdmin() *domain.UserAdmin {
	a := testAdmin(uuid.New(), true, false)
	a.CanViewLogs = true
	return a
}

func exporterAdmin() *domain.UserAdmin {
	a := viewerAdmin()
	a.CanExportLogs = true
	return a
}

func TestAuditHandler_ListLogs(t *testing.T) {
	t.Run("sin_sesion_masked_404", func(t *testing.T) {
		app := setupAuditApp(nil)
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		body := decodeBody(resp)
		require.NotEmpty(t, body["message"])
	})

	t.Run("sin_can_view_logs_masked_404", func(t *testing.T) {
		app := setupAuditApp(testAdmin(uuid.New(), true, false))
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("sudo_ok", func(t *testing.T) {
		log := domain.ApiChangeLog{
			ID:            uuid.New(),
			Entity:        domain.AuditEntityPsi,
			EntityID:      uuid.New().String(),
			Action:        domain.AuditActionLogin,
			ActorUsername: "admin",
		}
		svc := service.NewAuditService(&mockAuditRepo{
			listFunc: func(_ context.Context, f domain.AuditLogFilters) ([]domain.ApiChangeLog, int64, error) {
				if f.Action != "" && f.Action != domain.AuditActionLogin {
					t.Fatalf("filtro de acción inesperado: %s", f.Action)
				}
				return []domain.ApiChangeLog{log}, 1, nil
			},
		})
		h := NewAuditHandler(svc)
		app := newTestApp()
		sudo := testAdmin(uuid.New(), true, true)
		app.Use(func(c *fiber.Ctx) error { c.Locals("admin", sudo); return c.Next() })
		app.Get("/audit-logs", h.ListLogs)

		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs?suceso=login", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		require.Equal(t, float64(1), res["total"])
		require.NotEmpty(t, res["data"])
	})

	t.Run("can_view_logs_ok_con_filtros", func(t *testing.T) {
		app := setupAuditApp(viewerAdmin())
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs?entidad=psi&suceso=update", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("fecha_invalida_400", func(t *testing.T) {
		app := setupAuditApp(viewerAdmin())
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs?desde=no-es-fecha", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuditHandler_ListPsiLogs(t *testing.T) {
	t.Run("id_invalido_masked_404", func(t *testing.T) {
		app := setupAuditApp(viewerAdmin())
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs/psi/no-uuid", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("fuerza_entidad_psi_y_entity_id", func(t *testing.T) {
		psiID := uuid.New()
		svc := service.NewAuditService(&mockAuditRepo{
			listFunc: func(_ context.Context, f domain.AuditLogFilters) ([]domain.ApiChangeLog, int64, error) {
				if f.Entity != domain.AuditEntityPsi {
					t.Fatalf("se esperaba entidad psi, recibió %s", f.Entity)
				}
				if f.EntityID != psiID.String() {
					t.Fatalf("se esperaba entity_id %s, recibió %s", psiID.String(), f.EntityID)
				}
				return nil, 0, nil
			},
		})
		h := NewAuditHandler(svc)
		app := newTestApp()
		app.Use(func(c *fiber.Ctx) error { c.Locals("admin", viewerAdmin()); return c.Next() })
		app.Get("/audit-logs/psi/:id", h.ListPsiLogs)

		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs/psi/"+psiID.String(), nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestAuditHandler_ExportLogs(t *testing.T) {
	t.Run("can_view_logs_sin_export_masked_404", func(t *testing.T) {
		app := setupAuditApp(viewerAdmin())
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs/export", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("can_export_logs_csv_ok", func(t *testing.T) {
		app := setupAuditApp(exporterAdmin())
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs/export", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		ct := resp.Header.Get("Content-Type")
		require.Contains(t, ct, "text/csv")
	})
}

func TestAuditHandler_StatsLogs(t *testing.T) {
	t.Run("sin_permiso_masked_404", func(t *testing.T) {
		app := setupAuditApp(testAdmin(uuid.New(), true, false))
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs/stats", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("can_view_logs_ok", func(t *testing.T) {
		app := setupAuditApp(viewerAdmin())
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/audit-logs/stats", nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
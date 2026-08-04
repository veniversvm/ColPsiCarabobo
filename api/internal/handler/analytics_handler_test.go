package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

func TestGetDashboardStats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		repo := &mockAnalyticsRepo{
			GetDashboardStatsFunc: func(ctx context.Context) (*domain.DashboardStats, error) {
				return &domain.DashboardStats{
					LoginsTotal:     42,
					ActiveSessionsNow: 3,
				}, nil
			},
		}
		svc := service.NewAnalyticsService(repo)
		h := NewAnalyticsHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupAdminRoute(fiber.MethodGet, "/dashboard/stats", h.GetDashboardStats, adminRepo, &mockPsiRepo{})

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/dashboard/stats", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.DashboardStats
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, int64(42), result.LoginsTotal)
		require.Equal(t, int64(3), result.ActiveSessionsNow)
	})

	t.Run("no_auth", func(t *testing.T) {
		svc := service.NewAnalyticsService(&mockAnalyticsRepo{})
		h := NewAnalyticsHandler(svc)

		adminRepo := &mockAdminRepo{}
		app := setupAdminRoute(fiber.MethodGet, "/dashboard/stats", h.GetDashboardStats, adminRepo, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/admin/dashboard/stats", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		// ProtectedAdmin404 returns 404 when no token
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

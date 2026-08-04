package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// dashboardStatsTimeout acota el dashboard (25+ consultas) para que una BD lenta
// no deje la conexión colgada: la request muere a los 10s.
const dashboardStatsTimeout = 10 * time.Second

// AnalyticsHandler handles HTTP requests for dashboard analytics and statistics.
type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler with the given service.
func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

// GetDashboardStats returns aggregated analytics metrics for the admin dashboard.
func (h *AnalyticsHandler) GetDashboardStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), dashboardStatsTimeout)
	defer cancel()

	stats, err := h.svc.GetDashboardStats(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al obtener estadísticas"})
	}
	return c.JSON(stats)
}

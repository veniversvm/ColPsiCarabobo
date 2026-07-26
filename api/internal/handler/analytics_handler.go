package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

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
	stats, err := h.svc.GetDashboardStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al obtener estadísticas"})
	}
	return c.JSON(stats)
}

// api/internal/handler/birthdays.go
package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetBirthdaysByAdmin godoc
// @Summary      Aviso de cumpleaños (Admin)
// @Description  Retorna los agremiados que cumplen años hoy o en los próximos días y han autorizado el aviso de cumpleaños (opt-in).
// @Security     BearerAuth
// @Tags         Administración - Psicólogos
// @Produce      json
// @Param        range query string false "rango: 'today' (hoy) o 'week' (próximos 7 días). Default: today"
// @Success      200  {array}   service.BirthdayInfoProjection
// @Failure      401  {object}  map[string]string "No autorizado"
// @Router       /admin/psi/birthdays [get]
func (h *PsiHandler) GetBirthdaysByAdmin(c *fiber.Ctx) error {
	rng := c.Query("range", "today")
	now := time.Now()
	var from, to time.Time
	switch rng {
	case "week":
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		to = from.AddDate(0, 0, 7)
	default:
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		to = from.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}

	birthdays, err := h.service.GetBirthdaysByRange(c.UserContext(), from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al consultar cumpleaños"})
	}

	return c.JSON(fiber.Map{
		"range":     rng,
		"data":      birthdays,
		"total":     len(birthdays),
		"updated_at": time.Now(),
	})
}

// api/internal/handler/psi_terms_handler.go
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// =========================================================================
// TÉRMINOS Y CONDICIONES — PARTE II (AUTOGESTIÓN DEL AGREMIADO)
// =========================================================================

// TermsHandler expone la versión vigente de los términos y registra la
// aceptación de la Parte II.
//
// Es un handler independiente de PsiHandler (no un método suyo) porque el
// servicio de términos no necesita el sanitize de HTML, el cliente S3 ni el
// envío de correos que arrastra PsiService.
type TermsHandler struct {
	service *service.TermsService
}

// NewTermsHandler construye el handler de términos.
func NewTermsHandler(svc *service.TermsService) *TermsHandler {
	return &TermsHandler{service: svc}
}

// GetTermsStatus godoc
// @Summary      Estado de aceptación de los Términos y Condiciones (Parte II)
// @Description  Devuelve la versión vigente de los términos y si el agremiado ya la aceptó, con su historial. El control del aviso es del cliente: si este endpoint falla, el portal sigue funcionando por completo.
// @Security     BearerAuth
// @Tags         Psicólogos - Términos
// @Produce      json
// @Success      200 {object} service.TermsStatus
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /psi/me/terms [get]
func (h *TermsHandler) GetTermsStatus(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	status, err := h.service.GetStatus(c.UserContext(), psi.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no fue posible consultar el estado de los términos"})
	}

	return c.JSON(status)
}

// AcceptTermsRequest es el cuerpo de POST /psi/me/terms.
type AcceptTermsRequest struct {
	// Version es la versión que el agremiado afirma haber leído. El backend solo
	// acepta la vigente (domain.TermsVersion): enviar otra devuelve 409.
	Version string `json:"version"`
}

// AcceptTerms godoc
// @Summary      Aceptar los Términos y Condiciones (Parte II)
// @Description  Registra la aceptación de la versión vigente. Es idempotente: repetirla no crea una fila nueva ni un segundo evento de bitácora. Aceptar una versión que no es la vigente devuelve 409.
// @Security     BearerAuth
// @Tags         Psicólogos - Términos
// @Accept       json
// @Produce      json
// @Param        request body AcceptTermsRequest true "Versión aceptada"
// @Success      200 {object} service.TermsStatus
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /psi/me/terms [post]
func (h *TermsHandler) AcceptTerms(c *fiber.Ctx) error {
	psi, err := middleware.GetAuthenticatedPsi(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req AcceptTermsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}
	if req.Version == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "falta la versión de los términos"})
	}

	status, err := h.service.Accept(
		c.UserContext(),
		psi.ID,
		req.Version,
		c.IP(),
		string(c.Request().Header.UserAgent()),
	)
	if err != nil {
		return termsError(c, err)
	}

	return c.JSON(status)
}

// termsError mapea los errores del servicio a respuestas HTTP.
func termsError(c *fiber.Ctx, err error) error {
	if errors.Is(err, domain.ErrTermsVersionInvalid) {
		// 409 y no 400: el cliente está bien formado, la versión simplemente no
		// es la que el servidor tiene vigente. Un 400 haría pensar al frontend
		// en un error de tipeo y no lo recargaría.
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":           err.Error(),
			"current_version": domain.TermsVersion,
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no fue posible registrar la aceptación"})
}

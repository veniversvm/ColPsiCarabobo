// api/internal/router/manual_router.go
package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/handler"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// SetupManualRoutes sirve los manuales PDF del proyecto. Los archivos están
// embebidos en el binario (paquete internal/manuales) y SOLO se accede con una
// sesión válida (admin O psicólogo). Prefijo propio /manuales (no reutilizado
// en otro lugar) para que el 401 real funcione — ver gotcha 16 de
// api/AGENTS.md: registrar bajo /admin heredaría el enmascarado 404.
func SetupManualRoutes(router fiber.Router, adminRepo domain.UserAdminRepository, psiRepo domain.PsiUserRepository, analyticsSvc *service.AnalyticsService) {
	h := handler.NewManualHandler()
	authMid := middleware.NewAuthMiddleware(adminRepo, psiRepo, analyticsSvc)

	manuales := router.Group("/manuales", middleware.NoStore(), authMid.ProtectedAnyAuth())
	manuales.Get("/:file", h.GetManual)
}
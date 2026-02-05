package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

type PsiHandler struct {
	service *service.PsiService
}

// NewPsiHandler inyecta el servicio de psicólogos en el controlador
func NewPsiHandler(svc *service.PsiService) *PsiHandler {
	return &PsiHandler{
		service: svc,
	}
}

// UploadCsv godoc
// @Summary      Importar psicólogos masivamente
// @Description  Procesa un archivo CSV y crea registros. Solo accesible para administradores autorizados.
// @Tags         Psychologists
// @Accept       multipart/form-data
// @Produce      json
// @Param        csv  formData  file  true  "Archivo CSV a procesar"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string "Retornado si el usuario no tiene permisos (Security by Obscurity)"
// @Router       /psi/upload-csv [post]
func (h *PsiHandler) UploadCsv(c *fiber.Ctx) error {
	fmt.Println("Entrando a UploadCsv")
	// 1. OBTENER EL ADMIN DESDE EL TOKEN (Locals)
	// Esto es lo que inyectó el middleware. Si llegamos aquí, ya está validado.
	admin, ok := c.Locals("admin").(*domain.UserAdmin)
	if !ok || admin == nil {
		// Si no hay admin en locals, algo falló en el middleware. 404 por seguridad.
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Cannot POST /api/v1/psi/upload-csv",
		})
	}

	// 2. BORRA TODO EL BLOQUE QUE HACÍA: c.FormValue("admin_id")
	// Ya no lo necesitamos.

	// 3. PROCESAR EL ARCHIVO CSV
	file, err := c.FormFile("csv")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El archivo CSV es requerido",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No se pudo abrir el archivo"})
	}
	defer src.Close()

	// 4. USAR EL ID DEL ADMIN RECUPERADO DEL TOKEN
	success, failed := h.service.ImportFromCSV(c.UserContext(), src, admin.ID)

	return c.JSON(fiber.Map{
		"success_count": success,
		"failed":        failed,
	})
}

// Los demás métodos (ListPsis, GetPsi, etc.) se implementarán siguiendo el mismo patrón...
func (h *PsiHandler) ListPsis(c *fiber.Ctx) error  { return nil }
func (h *PsiHandler) GetPsi(c *fiber.Ctx) error    { return nil }
func (h *PsiHandler) UpdatePsi(c *fiber.Ctx) error { return nil }

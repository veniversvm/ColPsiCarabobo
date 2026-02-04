package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
// @Description  Procesa un archivo CSV y crea registros de psicólogos y sus datos colegiales.
// @Tags         Psychologists
// @Accept       multipart/form-data
// @Produce      json
// @Param        csv       formData  file    true  "Archivo CSV a procesar"
// @Param        admin_id  formData  string  true  "ID del administrador que realiza la carga"
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  map[string]string
// @Router       /psi/upload-csv [post]
func (h *PsiHandler) UploadCsv(c *fiber.Ctx) error {
	// 1. Obtener el archivo del FormData
	file, err := c.FormFile("csv")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El archivo CSV es requerido en el campo 'csv'",
		})
	}

	// 2. Obtener y validar el ID del Administrador
	adminIDStr := c.FormValue("admin_id")
	if adminIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El 'admin_id' es requerido",
		})
	}

	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El formato del 'admin_id' no es un UUID válido",
		})
	}

	// 3. Abrir el archivo de forma segura
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo procesar el archivo enviado",
		})
	}
	defer src.Close()

	// 4. Delegar el procesamiento al Servicio (Business Logic)
	// Usamos c.UserContext() para propagar cancelaciones si el cliente corta la conexión
	success, failed := h.service.ImportFromCSV(c.UserContext(), src, adminID)

	// 5. Respuesta detallada
	status := http.StatusOK
	if success == 0 && len(failed) > 0 {
		status = http.StatusMultiStatus // 207 si hubo errores parciales
	}

	return c.Status(status).JSON(fiber.Map{
		"message":                  "Procesamiento de CSV finalizado",
		"success_count":            success,
		"number_of_failed_records": len(failed),
		"failed_records":           failed,
	})
}

// Los demás métodos (ListPsis, GetPsi, etc.) se implementarán siguiendo el mismo patrón...
func (h *PsiHandler) ListPsis(c *fiber.Ctx) error  { return nil }
func (h *PsiHandler) GetPsi(c *fiber.Ctx) error    { return nil }
func (h *PsiHandler) UpdatePsi(c *fiber.Ctx) error { return nil }

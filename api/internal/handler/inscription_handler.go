// api/internal/handler/inscription_handler.go
package handler

import (
	"errors"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

// InscriptionHandler expone los endpoints de pre-inscripción de profesionales.
type InscriptionHandler struct {
	svc *service.InscriptionService
}

// NewInscriptionHandler crea un nuevo InscriptionHandler.
func NewInscriptionHandler(svc *service.InscriptionService) *InscriptionHandler {
	return &InscriptionHandler{svc: svc}
}

// Límites de archivos
const (
	maxFileSize     = 5 << 20 // 5MB
	allowedImageMIME = "image/"
	allowedPdfMIME   = "application/pdf"
)

// CheckCI godoc
// @Summary      Validar unicidad de cédula (público)
// @Description  Verifica si una cédula ya está registrada en el sistema o tiene una solicitud activa.
// @Tags         Inscripcion
// @Produce      json
// @Param        ci query int true "Cédula de identidad"
// @Success      200 {object} request_structs.UniquenessCheckResponse
// @Failure      400 {object} map[string]string "error: parámetro inválido"
// @Router       /inscripcion/check-ci [get]
func (h *InscriptionHandler) CheckCI(c *fiber.Ctx) error {
	ci, err := strconv.Atoi(c.Query("ci"))
	if err != nil || ci <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cédula inválida"})
	}
	res, err := h.svc.CheckCI(c.UserContext(), ci)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error interno al validar la cédula"})
	}
	return c.JSON(res)
}

// CheckFPV godoc
// @Summary      Validar unicidad de FPV (público)
// @Description  Verifica si un número FPV ya está registrado en el sistema o tiene una solicitud activa.
// @Tags         Inscripcion
// @Produce      json
// @Param        fpv query int true "Número FPV"
// @Success      200 {object} request_structs.UniquenessCheckResponse
// @Failure      400 {object} map[string]string "error: parámetro inválido"
// @Router       /inscripcion/check-fpv [get]
func (h *InscriptionHandler) CheckFPV(c *fiber.Ctx) error {
	fpv, err := strconv.Atoi(c.Query("fpv"))
	if err != nil || fpv <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "número de FPV inválido"})
	}
	res, err := h.svc.CheckFPV(c.UserContext(), fpv)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error interno al validar el FPV"})
	}
	return c.JSON(res)
}

// Submit godoc
// @Summary      Enviar solicitud de pre-inscripción (público)
// @Description  Crea una solicitud de pre-inscripción en estado "pending" con sus archivos adjuntos.
// @Tags         Inscripcion
// @Accept       multipart/form-data
// @Produce      json
// @Param        cedula formData int true "Cédula"
// @Param        nacionalidad formData string true "Nacionalidad (V/E)"
// @Param        nombres formData string true "Nombres"
// @Param        apellidos formData string true "Apellidos"
// @Param        fpv formData int false "Número FPV"
// @Param        telefono formData string false "Teléfono"
// @Param        correo formData string true "Correo electrónico"
// @Param        fecha_nacimiento formData string false "Fecha de nacimiento (YYYY-MM-DD)"
// @Param        titulo_universidad formData string false "Universidad"
// @Param        titulo_fecha_graduacion formData string false "Fecha de graduación (YYYY-MM-DD)"
// @Param        titulo_mencion formData string false "Mención"
// @Param        titulo_registro_numero formData string false "Nº registro del título"
// @Param        titulo_registro_estado formData string false "Estado del registro"
// @Param        rif formData string false "RIF"
// @Param        foto formData file false "Foto tipo carnet"
// @Param        comprobante formData file false "Comprobante de pago"
// @Success      201 {object} map[string]string "message: Solicitud recibida correctamente"
// @Failure      400 {object} map[string]string "error: validación fallida"
// @Failure      409 {object} map[string]string "error: cédula o FPV duplicado"
// @Failure      500 {object} map[string]string "error: fallo interno"
// @Router       /inscripcion/submit [post]
func (h *InscriptionHandler) Submit(c *fiber.Ctx) error {
	req, err := h.parseSubmitForm(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	inscription, err := h.svc.Submit(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, service.ErrCIExists) || errors.Is(err, service.ErrFPVExists) || errors.Is(err, service.ErrEmailExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		log.Error().Err(err).Str("component", "inscription").Msg("Error al procesar solicitud de inscripción")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo procesar la solicitud, intente nuevamente"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Solicitud recibida correctamente. Recibirás un correo con tus credenciales en un plazo de 5 días hábiles.",
		"id":      inscription.ID,
	})
}

// parseSubmitForm lee y valida el formulario multipart de la solicitud.
func (h *InscriptionHandler) parseSubmitForm(c *fiber.Ctx) (*service.SubmitInscriptionRequest, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errors.New("formulario inválido")
	}

	req := &service.SubmitInscriptionRequest{
		Nacionalidad:   strings.ToUpper(strings.TrimSpace(first(form, "nacionalidad"))),
		Nombres:        strings.TrimSpace(first(form, "nombres")),
		Apellidos:      strings.TrimSpace(first(form, "apellidos")),
		SegundoNombre:  strings.TrimSpace(first(form, "segundo_nombre")),
		SegundoApellido: strings.TrimSpace(first(form, "segundo_apellido")),
		Genero:         strings.ToUpper(strings.TrimSpace(first(form, "genero"))),
		Telefono:       strings.TrimSpace(first(form, "telefono")),
		Correo:         strings.ToLower(strings.TrimSpace(first(form, "correo"))),
		TituloUniversidad:    strings.TrimSpace(first(form, "titulo_universidad")),
		TituloMencion:        strings.TrimSpace(first(form, "titulo_mencion")),
		TituloRegistroNumero: strings.TrimSpace(first(form, "titulo_registro_numero")),
		TituloRegistroEstado: strings.TrimSpace(first(form, "titulo_registro_estado")),
		TituloRegistroTomo:   strings.TrimSpace(first(form, "titulo_registro_tomo")),
		TituloRegistroFolio:  strings.TrimSpace(first(form, "titulo_registro_folio")),
		RIF:                  strings.TrimSpace(first(form, "rif")),
	}

	// Validaciones requeridas
	if req.Nombres == "" || req.Apellidos == "" {
		return nil, errors.New("nombres y apellidos son obligatorios")
	}
	if req.Nacionalidad != "V" && req.Nacionalidad != "E" {
		return nil, errors.New("la nacionalidad debe ser V o E")
	}

	cedula, err := strconv.Atoi(first(form, "cedula"))
	if err != nil || cedula <= 0 {
		return nil, errors.New("cédula inválida o faltante")
	}
	req.Cedula = cedula

	if fpvRaw := strings.TrimSpace(first(form, "fpv")); fpvRaw != "" {
		fpv, err := strconv.Atoi(fpvRaw)
		if err == nil && fpv > 0 {
			req.FPV = fpv
		}
	}

	if req.Correo == "" || !isValidEmail(req.Correo) {
		return nil, errors.New("correo electrónico inválido")
	}

	// Fechas opcionales
	if v := strings.TrimSpace(first(form, "fecha_nacimiento")); v != "" {
		t, err := parseDateStr(v)
		if err != nil {
			return nil, errors.New("fecha de nacimiento inválida")
		}
		req.FechaNacimiento = &t
	}
	if v := strings.TrimSpace(first(form, "titulo_fecha_graduacion")); v != "" {
		t, err := parseDateStr(v)
		if err != nil {
			return nil, errors.New("fecha de graduación inválida")
		}
		req.TituloFechaGraduacion = &t
	}

	// Archivos
	foto, err := fileFromForm(form, "foto")
	if err != nil {
		return nil, err
	}
	req.Foto = foto

	comprobante, err := fileFromForm(form, "comprobante")
	if err != nil {
		return nil, err
	}
	req.Comprobante = comprobante

	return req, nil
}

// fileFromForm extrae y valida un archivo multipart.
func fileFromForm(form *multipart.Form, field string) (*multipart.FileHeader, error) {
	files := form.File[field]
	if len(files) == 0 {
		return nil, errors.New("archivo " + field + " es obligatorio")
	}
	fh := files[0]
	if fh == nil {
		return nil, nil
	}
	if fh.Size > maxFileSize {
		return nil, errors.New("el archivo " + field + " supera el tamaño máximo de 5MB")
	}
	ct := fh.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, allowedImageMIME) && ct != allowedPdfMIME {
		return nil, errors.New("el archivo " + field + " debe ser una imagen o PDF")
	}
	return fh, nil
}

func first(form *multipart.Form, key string) string {
	if form != nil && form.Value != nil {
		if v, ok := form.Value[key]; ok && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

func parseDateStr(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// =========================================================================
// ADMIN
// =========================================================================

// List godoc
// @Summary      Listar solicitudes de inscripción (admin)
// @Tags         Administración - Inscripciones
// @Produce      json
// @Security     BearerAuth
// @Router       /admin/inscripciones/list [get]
func (h *InscriptionHandler) List(c *fiber.Ctx) error {
	filter := request_structs.InscriptionListFilter{
		Status: c.Query("status", "pending"),
		Q:      c.Query("q"),
	}
	if v, err := strconv.Atoi(c.Query("page", "1")); err == nil {
		filter.Page = v
	}
	if v, err := strconv.Atoi(c.Query("limit", "20")); err == nil {
		filter.Limit = v
	}

	res, err := h.svc.List(c.UserContext(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al listar solicitudes"})
	}
	return c.JSON(res)
}

// Detail godoc
// @Summary      Detalle de solicitud (admin)
// @Tags         Administración - Inscripciones
// @Produce      json
// @Security     BearerAuth
// @Router       /admin/inscripciones/{id} [get]
func (h *InscriptionHandler) Detail(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}
	dto, err := h.svc.Detail(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, service.ErrInscriptionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al obtener la solicitud"})
	}
	return c.JSON(dto)
}

// Approve godoc
// @Summary      Aprobar inscripción (admin)
// @Description  Aprueba la solicitud, crea el psicólogo con is_active=false y envía email con credenciales.
// @Tags         Administración - Inscripciones
// @Produce      json
// @Security     BearerAuth
// @Router       /admin/inscripciones/{id}/approve [post]
func (h *InscriptionHandler) Approve(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}
	res, err := h.svc.Approve(c.UserContext(), admin, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInscriptionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrInscriptionNotPending):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrCIExists), errors.Is(err, service.ErrFPVExists), errors.Is(err, service.ErrEmailExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al aprobar la inscripción"})
	}
	return c.JSON(res)
}

// Reject godoc
// @Summary      Rechazar inscripción (admin)
// @Description  Elimina la solicitud y sus archivos S3.
// @Tags         Administración - Inscripciones
// @Security     BearerAuth
// @Router       /admin/inscripciones/{id} [delete]
func (h *InscriptionHandler) Reject(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}
	err = h.svc.Reject(c.UserContext(), admin, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInscriptionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrInscriptionNotPending):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al rechazar la inscripción"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateNotes godoc
// @Summary      Guardar notas administrativas (admin)
// @Description  Persiste las notas de texto simple del administrador sobre la solicitud.
// @Tags         Administración - Inscripciones
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Router       /admin/inscripciones/{id}/notes [patch]
func (h *InscriptionHandler) UpdateNotes(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var body request_structs.UpdateNotesRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cuerpo inválido"})
	}

	if err := h.svc.UpdateNotes(c.UserContext(), id, body.Notes); err != nil {
		if errors.Is(err, service.ErrInscriptionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al guardar las notas"})
	}

	return c.JSON(fiber.Map{"message": "Notas guardadas"})
}

// SendEmailToApplicant godoc
// @Summary      Enviar correo al solicitante (admin)
// @Description  Envía un correo con asunto y mensaje del administrador al correo del solicitante.
// @Tags         Administración - Inscripciones
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Router       /admin/inscripciones/{id}/email [post]
func (h *InscriptionHandler) SendEmailToApplicant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var body request_structs.SendEmailToApplicantRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cuerpo inválido"})
	}
	body.Subject = strings.TrimSpace(body.Subject)
	body.Message = strings.TrimSpace(body.Message)
	if body.Subject == "" || body.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "asunto y mensaje son obligatorios"})
	}

	emailSent, err := h.svc.SendEmailToApplicant(c.UserContext(), id, body.Subject, body.Message)
	if err != nil {
		if errors.Is(err, service.ErrInscriptionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al enviar el correo"})
	}

	return c.JSON(request_structs.SendEmailToApplicantResponse{EmailSent: emailSent})
}

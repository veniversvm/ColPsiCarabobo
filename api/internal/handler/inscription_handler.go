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
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
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
	maxFileSize      = 5 << 20 // 5MB
	allowedImageMIME = "image/"
	allowedPdfMIME   = "application/pdf"
)

// GetStatus godoc
// @Summary      Estado de recepción de inscripciones (público)
// @Description  Indica si la pre-inscripción de profesionales está habilitada. Cuando está desactivada, el campo `message` explica el motivo público.
// @Tags         Inscripcion
// @Produce      json
// @Success      200 {object} domain.ReceptionSetting
// @Router       /inscripcion/status [get]
func (h *InscriptionHandler) GetStatus(c *fiber.Ctx) error {
	status, err := h.svc.ReceptionStatus(c.UserContext())
	if err != nil {
		log.Error().Err(err).Str("component", "inscription").Msg("Error al consultar estado de recepción de inscripciones")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error al consultar el estado"})
	}
	return c.JSON(status)
}

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

// CheckEmail godoc
// @Summary      Validar unicidad de correo (público)
// @Description  Verifica si un correo ya está registrado en el sistema o tiene una solicitud activa.
// @Tags         Inscripcion
// @Produce      json
// @Param        correo query string true "Correo electrónico"
// @Success      200 {object} request_structs.UniquenessCheckResponse
// @Failure      400 {object} map[string]string "error: parámetro inválido"
// @Router       /inscripcion/check-email [get]
func (h *InscriptionHandler) CheckEmail(c *fiber.Ctx) error {
	email := strings.TrimSpace(c.Query("correo"))
	if email == "" || !isValidEmail(email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "correo electrónico inválido"})
	}
	res, err := h.svc.CheckEmail(c.UserContext(), email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error interno al validar el correo"})
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
// @Param        service_address formData string false "Dirección del consultorio"
// @Param        municipality_carabobo formData string false "Municipio (Carabobo)"
// @Param        state_outside formData string false "Estado (fuera de Carabobo)"
// @Param        municipality_outside_carabobo formData string false "Municipio (fuera de Carabobo)"
// @Param        country formData string false "País (fuera de Venezuela)"
// @Param        service_modality_presencial formData bool false "Modalidad presencial"
// @Param        service_modality_distance formData bool false "Modalidad a distancia"
// @Param        service_modality_telephone formData bool false "Modalidad telefónica"
// @Param        primary_specialty_id formData int false "Área de trabajo principal (id del catálogo)"
// @Param        secondary_specialty_id formData int false "Área de trabajo secundaria (id del catálogo)"
// @Param        foto formData file false "Foto tipo carnet"
// @Param        comprobante formData file false "Comprobante de pago"
// @Param        doc_cedula formData file false "Foto de la cédula"
// @Param        doc_titulo formData file false "Foto del título"
// @Param        doc_rif formData file false "Foto del RIF"
// @Param        doc_otro formData file false "Foto de otro documento"
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
		var rdErr *service.ReceptionDisabledError
		if errors.As(err, &rdErr) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "reception_disabled",
				"message": rdErr.Error(),
			})
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
		Nacionalidad:                strings.ToUpper(strings.TrimSpace(first(form, "nacionalidad"))),
		Nombres:                     strings.TrimSpace(first(form, "nombres")),
		Apellidos:                   strings.TrimSpace(first(form, "apellidos")),
		SegundoNombre:               strings.TrimSpace(first(form, "segundo_nombre")),
		SegundoApellido:             strings.TrimSpace(first(form, "segundo_apellido")),
		Genero:                      strings.ToUpper(strings.TrimSpace(first(form, "genero"))),
		Telefono:                    strings.TrimSpace(first(form, "telefono")),
		Correo:                      strings.ToLower(strings.TrimSpace(first(form, "correo"))),
		TituloUniversidad:           strings.TrimSpace(first(form, "titulo_universidad")),
		TituloMencion:               strings.TrimSpace(first(form, "titulo_mencion")),
		TituloRegistroNumero:        strings.TrimSpace(first(form, "titulo_registro_numero")),
		TituloRegistroEstado:        strings.TrimSpace(first(form, "titulo_registro_estado")),
		TituloRegistroTomo:          strings.TrimSpace(first(form, "titulo_registro_tomo")),
		TituloRegistroFolio:         strings.TrimSpace(first(form, "titulo_registro_folio")),
		RIF:                         strings.TrimSpace(first(form, "rif")),
		ServiceAddress:              strings.TrimSpace(first(form, "service_address")),
		MunicipalityCarabobo:        strings.TrimSpace(first(form, "municipality_carabobo")),
		StateOutside:                strings.TrimSpace(first(form, "state_outside")),
		MunicipalityOutSideCarabobo: strings.TrimSpace(first(form, "municipality_outside_carabobo")),
		Country:                     strings.TrimSpace(first(form, "country")),
		ServiceModalityPresencial:   formBool(form, "service_modality_presencial"),
		ServiceModalityDistance:     formBool(form, "service_modality_distance"),
		ServiceModalityTelephone:    formBool(form, "service_modality_telephone"),
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

	// Campos obligatorios de la ficha (personales, académicos y ubicación).
	if err := service.ValidateFichaObligatoria(service.FichaObligatoria{
		SegundoApellido:         req.SegundoApellido,
		Genero:                  req.Genero,
		Telefono:                req.Telefono,
		FechaNacimientoPresente: req.FechaNacimiento != nil,
		TituloUniversidad:       req.TituloUniversidad,
		FechaGraduacionPresente: req.TituloFechaGraduacion != nil,
		TituloRegistroEstado:    req.TituloRegistroEstado,
		ServiceAddress:          req.ServiceAddress,
		MunicipalityCarabobo:    req.MunicipalityCarabobo,
		StateOutside:            req.StateOutside,
		MunicipalityOutside:     req.MunicipalityOutSideCarabobo,
		Country:                 req.Country,
	}); err != nil {
		return nil, err
	}

	// Áreas de trabajo (ids del catálogo de especialidades)
	if v := strings.TrimSpace(first(form, "primary_specialty_id")); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			id := uint32(n)
			req.PrimarySpecialtyID = &id
		}
	}
	if v := strings.TrimSpace(first(form, "secondary_specialty_id")); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			id := uint32(n)
			req.SecondarySpecialtyID = &id
		}
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

	// Fotos de documentos requeridos (cedula, titulo, rif, otro)
	for _, docType := range []string{"cedula", "titulo", "rif", "otro"} {
		fh, err := fileFromFormOptional(form, "doc_"+docType)
		if err != nil {
			return nil, err
		}
		if fh != nil {
			req.Documents = append(req.Documents, service.InscriptionDocumentUpload{
				DocumentType: docType,
				File:         fh,
			})
		}
	}

	return req, nil
}

// formBool interpreta un campo de formulario como boolean (ausente o "false" ⇒ false).
func formBool(form *multipart.Form, key string) bool {
	v := strings.ToLower(strings.TrimSpace(first(form, key)))
	return v != "" && v != "0" && v != "false" && v != "no"
}

// fileFromForm extrae y valida un archivo multipart obligatorio.
func fileFromForm(form *multipart.Form, field string) (*multipart.FileHeader, error) {
	fh, err := fileFromFormOptional(form, field)
	if err != nil {
		return nil, err
	}
	if fh == nil {
		return nil, errors.New("archivo " + field + " es obligatorio")
	}
	return fh, nil
}

// fileFromFormOptional extrae y valida un archivo multipart opcional (nil si ausente).
func fileFromFormOptional(form *multipart.Form, field string) (*multipart.FileHeader, error) {
	files := form.File[field]
	if len(files) == 0 || files[0] == nil {
		return nil, nil
	}
	fh := files[0]
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
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

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

	res, err := h.svc.List(c.UserContext(), admin, filter)
	if err != nil {
		return mapInscriptionErr(c, err, "error al listar solicitudes")
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
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}
	dto, err := h.svc.Detail(c.UserContext(), admin, id)
	if err != nil {
		return mapInscriptionErr(c, err, "error al obtener la solicitud")
	}
	return c.JSON(dto)
}

// UpdateFicha godoc
// @Summary      Editar ficha de inscripción (admin)
// @Description  Actualiza los campos escalares de la ficha (ubicación, modalidades, áreas, datos personales). Solo admins con permiso de edición de psicólogos.
// @Tags         Administración - Inscripciones
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Router       /admin/inscripciones/{id} [patch]
func (h *InscriptionHandler) UpdateFicha(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var body request_structs.UpdateInscriptionRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cuerpo inválido"})
	}

	dto, err := h.svc.UpdateFicha(c.UserContext(), admin, id, &body)
	if err != nil {
		return mapInscriptionErr(c, err, "error al actualizar la ficha")
	}
	return c.JSON(dto)
}

// UpdateFichaPhoto godoc
// @Summary      Reemplazar foto de la ficha (admin)
// @Description  Reemplaza la foto tipo carnet o el comprobante de pago de la ficha.
// @Tags         Administración - Inscripciones
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        kind formData string true "foto | comprobante"
// @Param        file formData file true "Archivo de imagen o PDF"
// @Router       /admin/inscripciones/{id}/photo [post]
func (h *InscriptionHandler) UpdateFichaPhoto(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	kind := strings.TrimSpace(c.FormValue("kind"))
	if kind != "foto" && kind != "comprobante" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "kind debe ser foto o comprobante"})
	}
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "archivo obligatorio"})
	}
	if file.Size > maxFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "el archivo supera el tamaño máximo de 5MB"})
	}
	ct := file.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, allowedImageMIME) && ct != allowedPdfMIME {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "el archivo debe ser una imagen o PDF"})
	}

	dto, err := h.svc.UpdateFichaPhoto(c.UserContext(), admin, id, kind, file)
	if err != nil {
		return mapInscriptionErr(c, err, "error al reemplazar el archivo")
	}
	return c.JSON(dto)
}

// AddInscriptionDocument godoc
// @Summary      Agregar o reemplazar foto de documento (admin)
// @Description  Agrega o reemplaza la foto de un documento de la ficha (cedula, titulo, rif, otro).
// @Tags         Administración - Inscripciones
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        document_type formData string true "cedula | titulo | rif | otro"
// @Param        file formData file true "Archivo de imagen o PDF"
// @Router       /admin/inscripciones/{id}/documents [post]
func (h *InscriptionHandler) AddInscriptionDocument(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	docType := strings.TrimSpace(c.FormValue("document_type"))
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "archivo obligatorio"})
	}
	if file.Size > maxFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "el archivo supera el tamaño máximo de 5MB"})
	}
	ct := file.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, allowedImageMIME) && ct != allowedPdfMIME {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "el archivo debe ser una imagen o PDF"})
	}

	dto, err := h.svc.AddInscriptionDocument(c.UserContext(), admin, id, docType, file)
	if err != nil {
		return mapInscriptionErr(c, err, "error al agregar el documento")
	}
	return c.JSON(dto)
}

// DeleteInscriptionDocument godoc
// @Summary      Eliminar foto de documento (admin)
// @Description  Elimina la foto de un documento de la ficha y su archivo en el bucket.
// @Tags         Administración - Inscripciones
// @Security     BearerAuth
// @Router       /admin/inscripciones/{id}/documents/{docId} [delete]
func (h *InscriptionHandler) DeleteInscriptionDocument(c *fiber.Ctx) error {
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}
	docID, err := uuid.Parse(c.Params("docId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de documento inválido"})
	}

	err = h.svc.DeleteInscriptionDocument(c.UserContext(), admin, id, docID)
	if err != nil {
		return mapInscriptionErr(c, err, "error al eliminar el documento")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// mapInscriptionErr traduce errores del service a códigos HTTP.
func mapInscriptionErr(c *fiber.Ctx, err error, fallback string) error {
	switch {
	case errors.As(err, new(*service.ValidationError)):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInscriptionNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInscriptionNotPending):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrCIExists), errors.Is(err, service.ErrFPVExists), errors.Is(err, service.ErrEmailExists):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, domain.ErrPermissionDenied):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fallback})
	}
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
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var body request_structs.UpdateNotesRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cuerpo inválido"})
	}

	if err := h.svc.UpdateNotes(c.UserContext(), admin, id, body.Notes); err != nil {
		return mapInscriptionErr(c, err, "error al guardar las notas")
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
	admin, err := middleware.GetAuthenticatedAdmin(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

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

	emailSent, err := h.svc.SendEmailToApplicant(c.UserContext(), admin, id, body.Subject, body.Message)
	if err != nil {
		return mapInscriptionErr(c, err, "error al enviar el correo")
	}

	return c.JSON(request_structs.SendEmailToApplicantResponse{EmailSent: emailSent})
}

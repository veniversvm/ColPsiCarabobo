// api/internal/service/inscription_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
// Este archivo contiene las operaciones del módulo de pre-inscripción de profesionales.
package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"golang.org/x/crypto/bcrypt"
)

// InscriptionService agrupa la lógica de negocio de pre-inscripción de profesionales.
type InscriptionService struct {
	repo        domain.InscriptionRepository
	psiRepo     domain.PsiUserRepository
	s3Client    *s3.S3Client
	mailService IMailService
	sanitizer   *bluemonday.Policy
}

// NewInscriptionService crea un nuevo InscriptionService.
func NewInscriptionService(repo domain.InscriptionRepository, psiRepo domain.PsiUserRepository, s3Client *s3.S3Client, mailService IMailService) *InscriptionService {
	return &InscriptionService{
		repo:        repo,
		psiRepo:     psiRepo,
		s3Client:    s3Client,
		mailService: mailService,
		sanitizer:   bluemonday.UGCPolicy(),
	}
}

// ErrCIExists se retorna cuando la cédula ya está registrada (solicitud o psi_users).
var ErrCIExists = errors.New("la cédula ya se encuentra registrada o tiene una solicitud activa")

// ErrFPVExists se retorna cuando el FPV ya está registrado.
var ErrFPVExists = errors.New("el número de FPV ya se encuentra registrado")

// ErrEmailExists se retorna cuando el correo ya está registrado en psi_users
// o tiene una solicitud pendiente.
var ErrEmailExists = errors.New("el correo electrónico ya se encuentra registrado")

// ErrInscriptionNotFound se retorna cuando no existe la solicitud.
var ErrInscriptionNotFound = errors.New("solicitud de inscripción no encontrada")

// ErrInscriptionNotPending se retorna al intentar aprobar/rechazar una solicitud no pendiente.
var ErrInscriptionNotPending = errors.New("la solicitud ya fue procesada")

// CheckCI verifica si una cédula ya está registrada en psi_users o tiene solicitud pendiente.
func (s *InscriptionService) CheckCI(ctx context.Context, ci int) (*request_structs.UniquenessCheckResponse, error) {
	exists, err := s.repo.CIInPsiUsers(ctx, ci)
	if err != nil {
		return nil, err
	}
	if exists {
		return &request_structs.UniquenessCheckResponse{
			Exists:  true,
			Message: "La cédula ya está registrada en el sistema",
		}, nil
	}

	pending, err := s.repo.ExistsPendingCI(ctx, ci)
	if err != nil {
		return nil, err
	}
	if pending {
		return &request_structs.UniquenessCheckResponse{
			Exists:  true,
			Message: "Ya existe una solicitud activa con esta cédula",
		}, nil
	}

	return &request_structs.UniquenessCheckResponse{Exists: false}, nil
}

// CheckFPV verifica si un FPV ya está registrado en psi_users o tiene solicitud pendiente.
func (s *InscriptionService) CheckFPV(ctx context.Context, fpv int) (*request_structs.UniquenessCheckResponse, error) {
	exists, err := s.repo.FPVInPsiUsers(ctx, fpv)
	if err != nil {
		return nil, err
	}
	if exists {
		return &request_structs.UniquenessCheckResponse{
			Exists:  true,
			Message: "El número de FPV ya está registrado en el sistema",
		}, nil
	}

	pending, err := s.repo.ExistsPendingFPV(ctx, fpv)
	if err != nil {
		return nil, err
	}
	if pending {
		return &request_structs.UniquenessCheckResponse{
			Exists:  true,
			Message: "Ya existe una solicitud activa con este número de FPV",
		}, nil
	}

	return &request_structs.UniquenessCheckResponse{Exists: false}, nil
}

// SubmitInscriptionRequest agrupa los campos ya validados de la solicitud pública.
type SubmitInscriptionRequest struct {
	Cedula         int
	Nacionalidad   string
	Nombres        string
	Apellidos      string
	SegundoNombre  string
	SegundoApellido string
	Genero         string
	FPV            int
	Telefono       string
	Correo         string
	FechaNacimiento *time.Time
	TituloUniversidad      string
	TituloFechaGraduacion  *time.Time
	TituloMencion          string
	TituloRegistroNumero   string
	TituloRegistroEstado   string
	TituloRegistroTomo     string
	TituloRegistroFolio    string
	RIF                    string
	Foto                   *multipart.FileHeader
	Comprobante            *multipart.FileHeader
}

// Submit crea una nueva solicitud de pre-inscripción con estado "pending".
func (s *InscriptionService) Submit(ctx context.Context, req *SubmitInscriptionRequest) (*domain.PsiInscriptionRequest, error) {
	// 1. Validar unicidad de cédula
	exists, err := s.repo.CIInPsiUsers(ctx, req.Cedula)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCIExists
	}
	pending, err := s.repo.ExistsPendingCI(ctx, req.Cedula)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, ErrCIExists
	}

	// 2. Validar unicidad de FPV (solo si se provee)
	if req.FPV > 0 {
		fpvExists, err := s.repo.FPVInPsiUsers(ctx, req.FPV)
		if err != nil {
			return nil, err
		}
		if fpvExists {
			return nil, ErrFPVExists
		}
		fpvPending, err := s.repo.ExistsPendingFPV(ctx, req.FPV)
		if err != nil {
			return nil, err
		}
		if fpvPending {
			return nil, ErrFPVExists
		}
	}

	// 2.1 Validar unicidad de correo
	if req.Correo != "" {
		emailExists, err := s.repo.EmailInPsiUsers(ctx, req.Correo)
		if err != nil {
			return nil, err
		}
		if emailExists {
			return nil, ErrEmailExists
		}
		emailPending, err := s.repo.ExistsPendingEmail(ctx, req.Correo)
		if err != nil {
			return nil, err
		}
		if emailPending {
			return nil, ErrEmailExists
		}
	}

	// 3. Subir archivos a S3 (primero, para poder hacer rollback si algo falla)
	fotoKey, err := s.uploadFile(req.Foto, "inscripciones/fotos")
	if err != nil {
		return nil, err
	}
	comprobanteKey, err := s.uploadFile(req.Comprobante, "inscripciones/comprobantes")
	if err != nil {
		// cleanup de la foto ya subida
		if s.s3Client != nil && fotoKey != "" {
			_ = s.s3Client.DeleteFile(context.Background(), fotoKey)
		}
		return nil, err
	}

	// 4. Construir la entidad
	inscription := &domain.PsiInscriptionRequest{
		Cedula:                req.Cedula,
		Nacionalidad:          req.Nacionalidad,
		Nombres:               s.sanitizer.Sanitize(req.Nombres),
		Apellidos:             s.sanitizer.Sanitize(req.Apellidos),
		SegundoNombre:         s.sanitizer.Sanitize(req.SegundoNombre),
		SegundoApellido:       s.sanitizer.Sanitize(req.SegundoApellido),
		Genero:                s.sanitizer.Sanitize(strings.ToUpper(req.Genero)),
		FPV:                   req.FPV,
		Telefono:              s.sanitizer.Sanitize(req.Telefono),
		Correo:                s.sanitizer.Sanitize(req.Correo),
		FechaNacimiento:       req.FechaNacimiento,
		TituloUniversidad:     s.sanitizer.Sanitize(req.TituloUniversidad),
		TituloFechaGraduacion: req.TituloFechaGraduacion,
		TituloMencion:         s.sanitizer.Sanitize(req.TituloMencion),
		TituloRegistroNumero:  s.sanitizer.Sanitize(req.TituloRegistroNumero),
		TituloRegistroEstado:  s.sanitizer.Sanitize(req.TituloRegistroEstado),
		TituloRegistroTomo:    s.sanitizer.Sanitize(req.TituloRegistroTomo),
		TituloRegistroFolio:   s.sanitizer.Sanitize(req.TituloRegistroFolio),
		RIF:                   s.sanitizer.Sanitize(req.RIF),
		FotoS3Key:             fotoKey,
		ComprobanteS3Key:      comprobanteKey,
		Status:                domain.InscriptionPending,
	}

	// 5. Persistir
	if err := s.repo.Create(ctx, inscription); err != nil {
		// Rollback de archivos S3 si falla la inserción
		if s.s3Client != nil {
			_ = s.s3Client.DeleteFile(context.Background(), fotoKey)
			_ = s.s3Client.DeleteFile(context.Background(), comprobanteKey)
		}
		return nil, MapDBError(err)
	}

	return inscription, nil
}

// uploadFile sube un archivo multipart al bucket usando UploadFile.
// La validación de MIME y tamaño se realiza en el handler antes de invocar este método.
func (s *InscriptionService) uploadFile(file *multipart.FileHeader, folder string) (string, error) {
	if s.s3Client == nil {
		return "", errors.New("servicio de almacenamiento no disponible")
	}
	return s.s3Client.UploadFile(context.Background(), file, folder)
}

// List lista solicitudes con filtros y paginación (admin).
func (s *InscriptionService) List(ctx context.Context, filter request_structs.InscriptionListFilter) (*request_structs.InscriptionListResponse, error) {
	items, total, err := s.repo.Search(ctx, filter)
	if err != nil {
		return nil, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}

	dtos := make([]request_structs.InscriptionListDTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, request_structs.InscriptionListDTO{
			ID:            items[i].ID,
			Cedula:        items[i].Cedula,
			Nombres:       items[i].Nombres,
			Apellidos:     items[i].Apellidos,
			FPV:           items[i].FPV,
			Correo:        items[i].Correo,
			Status:        string(items[i].Status),
			ControlNumber: items[i].ControlNumber,
			CreatedAt:     items[i].CreatedAt,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &request_structs.InscriptionListResponse{
		Items:      dtos,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// Detail retorna el detalle de una solicitud con URLs públicas de archivos (admin).
func (s *InscriptionService) Detail(ctx context.Context, id uuid.UUID) (*request_structs.InscriptionDetailDTO, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInscriptionNotFound
	}

	dto := &request_structs.InscriptionDetailDTO{
		ID:                    req.ID,
		Cedula:                req.Cedula,
		Nacionalidad:          req.Nacionalidad,
		Nombres:               req.Nombres,
		Apellidos:             req.Apellidos,
		SegundoNombre:         req.SegundoNombre,
		SegundoApellido:       req.SegundoApellido,
		Genero:                req.Genero,
		FPV:                   req.FPV,
		Telefono:              req.Telefono,
		Correo:                req.Correo,
		FechaNacimiento:       req.FechaNacimiento,
		TituloUniversidad:     req.TituloUniversidad,
		TituloFechaGraduacion: req.TituloFechaGraduacion,
		TituloMencion:         req.TituloMencion,
		TituloRegistroNumero:  req.TituloRegistroNumero,
		TituloRegistroEstado:  req.TituloRegistroEstado,
		TituloRegistroTomo:    req.TituloRegistroTomo,
		TituloRegistroFolio:   req.TituloRegistroFolio,
		RIF:                   req.RIF,
		Status:                string(req.Status),
		ControlNumber:         req.ControlNumber,
		Notes:                 req.Notes,
		PsiUserID:             req.PsiUserID,
		CreatedAt:             req.CreatedAt,
		UpdatedAt:             req.UpdatedAt,
	}

	if req.PsiUserID != nil {
		solvencies, err := s.psiRepo.GetSolvencies(ctx, *req.PsiUserID)
		if err == nil {
			dto.SolvencyCount = len(solvencies)
		}
	}

	if s.s3Client != nil {
		dto.FotoURL = s.s3Client.GetPublicURL(req.FotoS3Key)
		dto.ComprobanteURL = s.s3Client.GetPublicURL(req.ComprobanteS3Key)
	}

	return dto, nil
}

// Approve aprueba una solicitud: genera N° control, crea el expediente del
// psicólogo (is_active=false) y envía email con las credenciales.
func (s *InscriptionService) Approve(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID) (*request_structs.ApproveInscriptionResponse, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInscriptionNotFound
	}
	if req.Status != domain.InscriptionPending {
		return nil, ErrInscriptionNotPending
	}

	// 1. Generar número de control secuencial
	controlNumber, err := s.repo.NextControlNumber(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Generar credenciales
	username := transliterateUsername(req.Nombres, req.Apellidos)
	tempPassword := utils.GenerateSecureRandomString(12)
	hashed, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("error al procesar seguridad")
	}

	// 2.5 Validar unicidad de correo antes de crear el psicólogo, para
	// no depender del constraint único (evita errores 500 en el approve).
	if req.Correo != "" {
		emailExists, err := s.repo.EmailInPsiUsers(ctx, req.Correo)
		if err != nil {
			return nil, err
		}
		if emailExists {
			return nil, ErrEmailExists
		}
	}

	// 3. Construir expediente
	psiID := uuid.Must(uuid.NewV7())
	colDataID := uuid.Must(uuid.NewV7())
	audit := domain.AuditModel{
		CreateBy:   admin.Username,
		CreateById: &admin.ID,
		UpdateBy:   admin.Username,
		UpdateById: &admin.ID,
	}

	now := time.Now()
	bornDate := time.Time{}
	if req.FechaNacimiento != nil {
		bornDate = *req.FechaNacimiento
	}
	gradDate := time.Time{}
	if req.TituloFechaGraduacion != nil {
		gradDate = *req.TituloFechaGraduacion
	}

	psi := &domain.PsiUserModel{
		ID:         psiID,
		AuditModel: audit,
		Credentials: domain.Credentials{
			Key:                uuid.Must(uuid.NewV7()).String(),
			Username:           username,
			Email:              req.Correo,
			Password:           string(hashed),
			IsActive:           true, // Parametrizable: al aprobar la inscripción la cuenta nace activa
			MustChangePassword: true,
		},
		FirstName:           req.Nombres,
		SecondName:          req.SegundoNombre,
		LastName:            req.Apellidos,
		SecondLastName:      req.SegundoApellido,
		CI:                  req.Cedula,
		FPV:                 req.FPV,
		Nationality:         req.Nacionalidad,
		Genre:               req.Genero, // M / F (se adopta vacío si no se indicó)
		BornDate:            bornDate,
		AudioBookShellId:    psiID.String(), // Único por psi (constraint uni) — igual que en el import
		Solvent:             true, // Inscripción aprobada ⇒ psicólogo solvente
		ControlNumber:       fmt.Sprintf("%d", controlNumber),
		ProfilePictureS3Key: req.FotoS3Key, // La foto tipo carnet pasa a ser la foto de perfil
		ContactEmail:        req.Correo,
		ContactPhone:        req.Telefono,
		ContactCellPhone:    req.Telefono,
	}

	colData := &domain.PsiUserColData{
		ID:                     colDataID,
		PsiUserModelID:         psiID,
		AuditModel:             audit,
		GuildInscriptionDate:   now,
		UniversityUndergraduate: req.TituloUniversidad,
		GraduateDate:           gradDate,
		MentionUndergraduate:   req.TituloMencion,
		RegisterNumber:         parseRegisterNumber(req.TituloRegistroNumero),
		RegisterTitleState:     req.TituloRegistroEstado,
		RegisterTome:           req.TituloRegistroTomo,
		RegisterFolio:          req.TituloRegistroFolio,
	}

	// Solvencia del año de ingreso, consistente con la regla existente de
	// "solvente = año vigente pagado" y con el conteo de solvencias pagadas.
	solvencyDate := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	solvency := domain.PsiUserSolvency{
		ID:             uuid.Must(uuid.NewV7()),
		AuditModel:     audit,
		PsiUserModelID: psiID,
		Date:           solvencyDate,
	}

	if err := s.psiRepo.CreateWithColData(ctx, psi, colData, []domain.PsiUserSolvency{solvency}, []domain.PsiUserPostGrade{}); err != nil {
		return nil, MapDBError(err)
	}

	// 4. Marcar solicitud como aprobada y vincular el expediente creado
	req.Status = domain.InscriptionApproved
	req.ControlNumber = fmt.Sprintf("%d", controlNumber)
	req.PsiUserID = &psiID
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	// 6. Enviar email con credenciales (no bloqueante)
	emailSent := false
	if s.mailService != nil {
		mailData := map[string]interface{}{
			"Name":     req.Nombres,
			"Email":    req.Correo,
			"Password": tempPassword,
		}
		if err := s.mailService.SendEmail(req.Correo, "Bienvenido", "welcome_psi", mailData); err == nil {
			emailSent = true
		}
	}

	return &request_structs.ApproveInscriptionResponse{
		Message:       "Inscripción aprobada y psicólogo creado",
		PsiUserID:     psiID,
		ControlNumber: fmt.Sprintf("%d", controlNumber),
		EmailSent:     emailSent,
	}, nil
}

// Reject rechaza una solicitud: elimina archivos S3 y el registro.
func (s *InscriptionService) Reject(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID) error {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrInscriptionNotFound
	}
	if req.Status != domain.InscriptionPending {
		return ErrInscriptionNotPending
	}

	// Eliminar archivos S3
	if s.s3Client != nil {
		if req.FotoS3Key != "" {
			_ = s.s3Client.DeleteFile(context.Background(), req.FotoS3Key)
		}
		if req.ComprobanteS3Key != "" {
			_ = s.s3Client.DeleteFile(context.Background(), req.ComprobanteS3Key)
		}
	}

	return s.repo.Delete(ctx, id)
}

// UpdateNotes actualiza las notas administrativas de una solicitud (texto simple).
func (s *InscriptionService) UpdateNotes(ctx context.Context, id uuid.UUID, notes string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrInscriptionNotFound
	}
	return s.repo.UpdateNotes(ctx, id, s.sanitizer.Sanitize(notes))
}

// SendEmailToApplicant envía un correo al solicitante con el mensaje del admin.
// Retorna true si el correo fue encolado correctamente (envío no bloqueante).
func (s *InscriptionService) SendEmailToApplicant(ctx context.Context, id uuid.UUID, subject, message string) (bool, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, ErrInscriptionNotFound
	}
	if s.mailService == nil {
		return false, errors.New("servicio de correo no disponible")
	}

	err = s.mailService.SendEmail(req.Correo, subject, "notification", map[string]interface{}{
		"Title":   "Notificación del Colegio de Psicólogos",
		"Message": s.sanitizer.Sanitize(message),
	})
	return err == nil, err
}

// transliterateUsername genera un username en snake_case desde nombres y apellidos.
func transliterateUsername(nombres, apellidos string) string {
	base := strings.ToLower(normString(nombres+"."+apellidos))
	// Remover caracteres no alfanuméricos
	var sb strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		}
	}
	out := sb.String()
	if len(out) > 25 {
		out = out[:25]
	}
	return out
}

// normString elimina tildes de caracteres latinos.
func normString(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u",
		"ñ", "n", "Ñ", "n", "ü", "u", "Ü", "u",
		" ", "", ".", "",
	)
	return replacer.Replace(s)
}

// parseRegisterNumber convierte el número de registro a int (0 si no es numérico).
func parseRegisterNumber(value string) int {
	if value == "" {
		return 0
	}
	n := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

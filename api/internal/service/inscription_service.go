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

// canViewFicha indica si el admin puede ver información de la ficha de solicitudes
// (gestión de la información del psicólogo).
func canViewFicha(a *domain.UserAdmin) bool {
	return a.Sudo || a.CanUpdatePsi || a.CanCreatePsi || a.CanDeletePsi
}

// requireUpdateFicha exige permiso de edición de la información del psicólogo.
func requireUpdateFicha(a *domain.UserAdmin) error {
	if a.Sudo || a.CanUpdatePsi {
		return nil
	}
	return domain.ErrPermissionDenied
}

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

// CheckEmail verifica si un correo ya está registrado en psi_users o tiene
// solicitud pendiente.
func (s *InscriptionService) CheckEmail(ctx context.Context, email string) (*request_structs.UniquenessCheckResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	exists, err := s.repo.EmailInPsiUsers(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return &request_structs.UniquenessCheckResponse{
			Exists:  true,
			Message: "El correo electrónico ya está registrado en el sistema",
		}, nil
	}

	pending, err := s.repo.ExistsPendingEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if pending {
		return &request_structs.UniquenessCheckResponse{
			Exists:  true,
			Message: "Ya existe una solicitud activa con este correo electrónico",
		}, nil
	}

	return &request_structs.UniquenessCheckResponse{Exists: false}, nil
}

// SubmitInscriptionRequest agrupa los campos ya validados de la solicitud pública.
type SubmitInscriptionRequest struct {
	Cedula          int
	Nacionalidad    string
	Nombres         string
	Apellidos       string
	SegundoNombre   string
	SegundoApellido string
	Genero          string
	FPV             int
	Telefono        string
	Correo          string
	FechaNacimiento *time.Time

	TituloUniversidad      string
	TituloFechaGraduacion  *time.Time
	TituloMencion          string
	TituloRegistroNumero   string
	TituloRegistroEstado   string
	TituloRegistroTomo     string
	TituloRegistroFolio    string
	RIF                    string

	// Ficha cercana a la interna
	ServiceAddress              string
	MunicipalityCarabobo        string
	StateOutside                string
	MunicipalityOutSideCarabobo string
	Country                     string
	ServiceModalityPresencial   bool
	ServiceModalityDistance     bool
	ServiceModalityTelephone    bool
	PrimarySpecialtyID          *uint32
	SecondarySpecialtyID        *uint32

	Foto        *multipart.FileHeader
	Comprobante *multipart.FileHeader
	Documents   []InscriptionDocumentUpload
}

// InscriptionDocumentUpload agrupa la foto de un documento requerido.
type InscriptionDocumentUpload struct {
	DocumentType string
	File         *multipart.FileHeader
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
	uploadedKeys := []string{fotoKey}
	rollback := func() {
		if s.s3Client != nil {
			for _, k := range uploadedKeys {
				if k != "" {
					_ = s.s3Client.DeleteFile(context.Background(), k)
				}
			}
		}
	}

	comprobanteKey, err := s.uploadFile(req.Comprobante, "inscripciones/comprobantes")
	if err != nil {
		rollback()
		return nil, err
	}
	uploadedKeys = append(uploadedKeys, comprobanteKey)

	// 3.1 Validar las fotos de documentos requeridos (cedula, titulo, rif).
	requiredDocs := []domain.DocumentType{domain.DocumentCedula, domain.DocumentTitulo, domain.DocumentRif}
	seenTypes := make(map[string]bool)
	for _, d := range req.Documents {
		if seenTypes[d.DocumentType] {
			rollback()
			return nil, errors.New("no se puede adjuntar dos fotos del mismo documento")
		}
		seenTypes[d.DocumentType] = true
	}
	for _, dt := range requiredDocs {
		if !seenTypes[string(dt)] {
			rollback()
			return nil, fmt.Errorf("es obligatorio adjuntar la foto del documento: %s", string(dt))
		}
	}

	// 3.2 Subir las fotos de documentos a S3 (se insertan tras crear la solicitud).
	var docRows []domain.PsiInscriptionDocument
	for _, d := range req.Documents {
		dt := domain.DocumentType(d.DocumentType)
		if !dt.IsValid() {
			rollback()
			return nil, errors.New("tipo de documento inválido")
		}
		key, err := s.uploadFile(d.File, "inscripciones/documentos/"+string(dt))
		if err != nil {
			rollback()
			return nil, err
		}
		uploadedKeys = append(uploadedKeys, key)
		docRows = append(docRows, domain.PsiInscriptionDocument{
			ID:               uuid.Must(uuid.NewV7()),
			DocumentType:     dt,
			S3Key:            key,
			OriginalFilename: d.File.Filename,
		})
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
		ServiceAddress:        s.sanitizer.Sanitize(req.ServiceAddress),
		MunicipalityCarabobo:  s.sanitizer.Sanitize(req.MunicipalityCarabobo),
		StateOutside:          s.sanitizer.Sanitize(req.StateOutside),
		MunicipalityOutSideCarabobo: s.sanitizer.Sanitize(req.MunicipalityOutSideCarabobo),
		Country:                     s.sanitizer.Sanitize(req.Country),
		ServiceModalityPresencial:   req.ServiceModalityPresencial,
		ServiceModalityDistance:     req.ServiceModalityDistance,
		ServiceModalityTelephone:    req.ServiceModalityTelephone,
		PrimarySpecialtyID:          req.PrimarySpecialtyID,
		SecondarySpecialtyID:        req.SecondarySpecialtyID,
		FotoS3Key:             fotoKey,
		ComprobanteS3Key:      comprobanteKey,
		Status:                domain.InscriptionPending,
	}

	// 5. Persistir
	if err := s.repo.Create(ctx, inscription); err != nil {
		// Rollback de archivos S3 si falla la inserción
		rollback()
		return nil, MapDBError(err)
	}

	// 5.1 Persistir las fotos de documentos vinculadas a la solicitud
	for i := range docRows {
		docRows[i].InscriptionRequestID = inscription.ID
	}
	if len(docRows) > 0 {
		if err := s.repo.CreateDocuments(ctx, docRows); err != nil {
			_ = s.repo.Delete(ctx, inscription.ID)
			rollback()
			return nil, MapDBError(err)
		}
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
func (s *InscriptionService) List(ctx context.Context, admin *domain.UserAdmin, filter request_structs.InscriptionListFilter) (*request_structs.InscriptionListResponse, error) {
	if !canViewFicha(admin) {
		return nil, domain.ErrPermissionDenied
	}

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
func (s *InscriptionService) Detail(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID) (*request_structs.InscriptionDetailDTO, error) {
	if !canViewFicha(admin) {
		return nil, domain.ErrPermissionDenied
	}

	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInscriptionNotFound
	}

	// Fotos de documentos de la ficha
	docs, err := s.repo.ListDocumentsByRequestID(ctx, req.ID)
	if err != nil {
		return nil, err
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
		ServiceAddress:        req.ServiceAddress,
		MunicipalityCarabobo:  req.MunicipalityCarabobo,
		StateOutside:          req.StateOutside,
		MunicipalityOutSideCarabobo: req.MunicipalityOutSideCarabobo,
		Country:                     req.Country,
		ServiceModalityPresencial:   req.ServiceModalityPresencial,
		ServiceModalityDistance:     req.ServiceModalityDistance,
		ServiceModalityTelephone:    req.ServiceModalityTelephone,
		PrimarySpecialtyID:          req.PrimarySpecialtyID,
		SecondarySpecialtyID:        req.SecondarySpecialtyID,
		Documents:                   buildInscriptionDocumentDTOs(docs, s.s3Client),
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

// buildInscriptionDocumentDTOs convierte los documentos de la ficha en DTOs con
// URL pública resuelta.
func buildInscriptionDocumentDTOs(docs []domain.PsiInscriptionDocument, client *s3.S3Client) []request_structs.InscriptionDocumentDTO {
	out := make([]request_structs.InscriptionDocumentDTO, 0, len(docs))
	for i := range docs {
		url := ""
		if client != nil {
			url = client.GetPublicURL(docs[i].S3Key)
		}
		out = append(out, request_structs.InscriptionDocumentDTO{
			ID:               docs[i].ID,
			DocumentType:     string(docs[i].DocumentType),
			URL:              url,
			Title:            docs[i].Title,
			Notes:            docs[i].Notes,
			OriginalFilename: docs[i].OriginalFilename,
		})
	}
	return out
}

// Approve aprueba una solicitud: genera N° control, crea el expediente del
// psicólogo (is_active=false) y envía email con las credenciales.
func (s *InscriptionService) Approve(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID) (*request_structs.ApproveInscriptionResponse, error) {
	if !admin.Sudo && !admin.CanCreatePsi {
		return nil, domain.ErrPermissionDenied
	}

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
		ServiceAddress:      req.ServiceAddress,
		MunicipalityCarabobo: req.MunicipalityCarabobo,
		StateOutside:        req.StateOutside,
		MunicipalityOutSideCarabobo: req.MunicipalityOutSideCarabobo,
		Country:             req.Country,
		ServiceModalityPresencial: req.ServiceModalityPresencial,
		ServiceModalityDistance:   req.ServiceModalityDistance,
		ServiceModalityTelephone:  req.ServiceModalityTelephone,
		PrimarySpecialtyID:        req.PrimarySpecialtyID,
		SecondarySpecialtyID:      req.SecondarySpecialtyID,
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

	// 3.1 Migrar las fotos de documentos al expediente del psicólogo,
	// reutilizando las claves S3 ya subidas (best-effort: si falla alguna, el
	// documento permanece en la ficha para re-cargarlo manualmente; la
	// aprobación no se bloquea).
	s.migrateDocumentsToPsi(ctx, req.ID, psiID, admin)

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
	if !admin.Sudo && !admin.CanDeletePsi {
		return domain.ErrPermissionDenied
	}

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
		// Fotos de documentos de la ficha
		if docs, err := s.repo.ListDocumentsByRequestID(ctx, id); err == nil {
			for _, d := range docs {
				if d.S3Key != "" {
					_ = s.s3Client.DeleteFile(context.Background(), d.S3Key)
				}
			}
		}
	}

	// Eliminar filas de documentos de la ficha y luego la solicitud
	_ = s.repo.DeleteInscriptionDocumentsByRequestID(ctx, id)
	return s.repo.Delete(ctx, id)
}

// UpdateNotes actualiza las notas administrativas de una solicitud (texto simple).
func (s *InscriptionService) UpdateNotes(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID, notes string) error {
	if err := requireUpdateFicha(admin); err != nil {
		return err
	}
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrInscriptionNotFound
	}
	return s.repo.UpdateNotes(ctx, id, s.sanitizer.Sanitize(notes))
}

// SendEmailToApplicant envía un correo al solicitante con el mensaje del admin.
// Retorna true si el correo fue encolado correctamente (envío no bloqueante).
func (s *InscriptionService) SendEmailToApplicant(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID, subject, message string) (bool, error) {
	if err := requireUpdateFicha(admin); err != nil {
		return false, err
	}
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

// =========================================================================
// EDICIÓN ADMIN DE LA FICHA (solo admins con permiso de edición)
// =========================================================================

// UpdateFicha actualiza los campos escalares de la ficha con semántica de
// reemplazo (el formulario admin envía todos los campos visibles) y re-verifica
// la unicidad de cédula / FPV / correo excluyendo la propia solicitud.
func (s *InscriptionService) UpdateFicha(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID, req *request_structs.UpdateInscriptionRequest) (*request_structs.InscriptionDetailDTO, error) {
	if err := requireUpdateFicha(admin); err != nil {
		return nil, err
	}

	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInscriptionNotFound
	}

	// Unicidad (solo si el valor cambió)
	if req.Cedula != cur.Cedula {
		exists, err := s.repo.CIInPsiUsers(ctx, req.Cedula)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCIExists
		}
		pending, err := s.repo.ExistsPendingCIExcluding(ctx, req.Cedula, id)
		if err != nil {
			return nil, err
		}
		if pending {
			return nil, ErrCIExists
		}
	}

	if req.FPV > 0 && req.FPV != cur.FPV {
		exists, err := s.repo.FPVInPsiUsers(ctx, req.FPV)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrFPVExists
		}
		pending, err := s.repo.ExistsPendingFPVExcluding(ctx, req.FPV, id)
		if err != nil {
			return nil, err
		}
		if pending {
			return nil, ErrFPVExists
		}
	}

	if !strings.EqualFold(strings.TrimSpace(req.Correo), strings.TrimSpace(cur.Correo)) {
		exists, err := s.repo.EmailInPsiUsers(ctx, req.Correo)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailExists
		}
		pending, err := s.repo.ExistsPendingEmailExcluding(ctx, req.Correo, id)
		if err != nil {
			return nil, err
		}
		if pending {
			return nil, ErrEmailExists
		}
	}

	// Aplicar los campos (sanitización en barrera final).
	cur.Cedula = req.Cedula
	cur.Nacionalidad = s.sanitizer.Sanitize(req.Nacionalidad)
	cur.Nombres = s.sanitizer.Sanitize(req.Nombres)
	cur.Apellidos = s.sanitizer.Sanitize(req.Apellidos)
	cur.SegundoNombre = s.sanitizer.Sanitize(req.SegundoNombre)
	cur.SegundoApellido = s.sanitizer.Sanitize(req.SegundoApellido)
	cur.Genero = s.sanitizer.Sanitize(strings.ToUpper(req.Genero))
	cur.FPV = req.FPV
	cur.Telefono = s.sanitizer.Sanitize(req.Telefono)
	cur.Correo = s.sanitizer.Sanitize(strings.TrimSpace(req.Correo))
	cur.FechaNacimiento = parseOptionalDate(req.FechaNacimiento)
	cur.TituloUniversidad = s.sanitizer.Sanitize(req.TituloUniversidad)
	cur.TituloFechaGraduacion = parseOptionalDate(req.TituloFechaGraduacion)
	cur.TituloMencion = s.sanitizer.Sanitize(req.TituloMencion)
	cur.TituloRegistroNumero = s.sanitizer.Sanitize(req.TituloRegistroNumero)
	cur.TituloRegistroEstado = s.sanitizer.Sanitize(req.TituloRegistroEstado)
	cur.TituloRegistroTomo = s.sanitizer.Sanitize(req.TituloRegistroTomo)
	cur.TituloRegistroFolio = s.sanitizer.Sanitize(req.TituloRegistroFolio)
	cur.RIF = s.sanitizer.Sanitize(req.RIF)
	cur.ServiceAddress = s.sanitizer.Sanitize(req.ServiceAddress)
	cur.MunicipalityCarabobo = s.sanitizer.Sanitize(req.MunicipalityCarabobo)
	cur.StateOutside = s.sanitizer.Sanitize(req.StateOutside)
	cur.MunicipalityOutSideCarabobo = s.sanitizer.Sanitize(req.MunicipalityOutSideCarabobo)
	cur.Country = s.sanitizer.Sanitize(req.Country)
	cur.ServiceModalityPresencial = req.ServiceModalityPresencial
	cur.ServiceModalityDistance = req.ServiceModalityDistance
	cur.ServiceModalityTelephone = req.ServiceModalityTelephone
	cur.PrimarySpecialtyID = req.PrimarySpecialtyID
	cur.SecondarySpecialtyID = req.SecondarySpecialtyID

	if err := s.repo.Update(ctx, cur); err != nil {
		return nil, err
	}

	return s.Detail(ctx, admin, id)
}

// parseOptionalDate convierte una fecha "YYYY-MM-DD" a *time.Time (nil si vacío).
func parseOptionalDate(v *string) *time.Time {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(*v))
	if err != nil {
		return nil
	}
	return &t
}

// UpdateFichaPhoto reemplaza la foto tipo carnet o el comprobante de pago de la
// ficha, eliminando el archivo anterior del bucket.
func (s *InscriptionService) UpdateFichaPhoto(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID, kind string, file *multipart.FileHeader) (*request_structs.InscriptionDetailDTO, error) {
	if err := requireUpdateFicha(admin); err != nil {
		return nil, err
	}
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInscriptionNotFound
	}

	folder := "inscripciones/fotos"
	field := "foto"
	switch kind {
	case "foto":
		folder = "inscripciones/fotos"
		field = "foto"
	case "comprobante":
		folder = "inscripciones/comprobantes"
		field = "comprobante"
	default:
		return nil, errors.New("tipo de archivo inválido")
	}

	key, err := s.uploadFile(file, folder)
	if err != nil {
		return nil, err
	}

	// Reemplazar la clave, borrando el archivo anterior si hubo éxito.
	oldKey := ""
	if field == "foto" {
		oldKey = cur.FotoS3Key
		cur.FotoS3Key = key
	} else {
		oldKey = cur.ComprobanteS3Key
		cur.ComprobanteS3Key = key
	}

	if err := s.repo.Update(ctx, cur); err != nil {
		if s.s3Client != nil {
			_ = s.s3Client.DeleteFile(context.Background(), key)
		}
		return nil, err
	}
	if s.s3Client != nil && oldKey != "" && oldKey != key {
		_ = s.s3Client.DeleteFile(context.Background(), oldKey)
	}

	return s.Detail(ctx, admin, id)
}

// AddInscriptionDocument agrega o reemplaza la foto de un documento de la ficha
// (cedula, titulo, rif, otro). Un documento por categoría; el anterior se borra
// del bucket.
func (s *InscriptionService) AddInscriptionDocument(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID, docType string, file *multipart.FileHeader) (*request_structs.InscriptionDetailDTO, error) {
	if err := requireUpdateFicha(admin); err != nil {
		return nil, err
	}
	dt := domain.DocumentType(docType)
	if !dt.IsValid() {
		return nil, errors.New("tipo de documento inválido")
	}
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return nil, ErrInscriptionNotFound
	}

	existing, err := s.repo.ListDocumentsByRequestID(ctx, id)
	if err != nil {
		return nil, err
	}
	var prior *domain.PsiInscriptionDocument
	for i := range existing {
		if existing[i].DocumentType == dt {
			prior = &existing[i]
			break
		}
	}

	key, err := s.uploadFile(file, "inscripciones/documentos/"+string(dt))
	if err != nil {
		return nil, err
	}

	if prior != nil {
		// Reemplazo: actualizar la fila y borrar el archivo anterior.
		oldKey := prior.S3Key
		prior.S3Key = key
		prior.OriginalFilename = file.Filename
		if err := s.repo.UpdateInscriptionDocument(ctx, prior); err != nil {
			if s.s3Client != nil {
				_ = s.s3Client.DeleteFile(context.Background(), key)
			}
			return nil, err
		}
		if s.s3Client != nil && oldKey != "" && oldKey != key {
			_ = s.s3Client.DeleteFile(context.Background(), oldKey)
		}
	} else {
		doc := &domain.PsiInscriptionDocument{
			ID:                 uuid.Must(uuid.NewV7()),
			InscriptionRequestID: id,
			DocumentType:       dt,
			S3Key:              key,
			OriginalFilename:   file.Filename,
		}
		if err := s.repo.CreateDocuments(ctx, []domain.PsiInscriptionDocument{*doc}); err != nil {
			if s.s3Client != nil {
				_ = s.s3Client.DeleteFile(context.Background(), key)
			}
			return nil, MapDBError(err)
		}
	}

	return s.Detail(ctx, admin, id)
}

// DeleteInscriptionDocument elimina la foto de un documento de la ficha y su
// archivo en el bucket.
func (s *InscriptionService) DeleteInscriptionDocument(ctx context.Context, admin *domain.UserAdmin, id uuid.UUID, docID uuid.UUID) error {
	if err := requireUpdateFicha(admin); err != nil {
		return err
	}
	doc, err := s.repo.GetInscriptionDocumentByID(ctx, docID)
	if err != nil {
		return ErrInscriptionNotFound
	}
	if doc.InscriptionRequestID != id {
		return ErrInscriptionNotFound
	}
	if s.s3Client != nil && doc.S3Key != "" {
		_ = s.s3Client.DeleteFile(context.Background(), doc.S3Key)
	}
	return s.repo.DeleteInscriptionDocument(ctx, docID)
}

// migrateDocumentsToPsi migra las fotos de documentos de una solicitud al
// expediente del psicólogo creado al aprobar, reutilizando las claves S3.
// Es best-effort: ante un error loguea y deja la fila original para recarga
// manual; nunca bloquea la aprobación.
func (s *InscriptionService) migrateDocumentsToPsi(ctx context.Context, requestID, psiID uuid.UUID, admin *domain.UserAdmin) {
	docs, err := s.repo.ListDocumentsByRequestID(ctx, requestID)
	if err != nil {
		// Si no se puede ni listar, no tocar nada.
		return
	}

	ok := true
	for i := range docs {
		audit := domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		}
		psiDoc := &domain.PsiUserDocument{
			ID:               uuid.Must(uuid.NewV7()),
			AuditModel:       audit,
			PsiUserID:        psiID,
			DocumentType:     docs[i].DocumentType,
			S3Key:            docs[i].S3Key,
			Title:            docs[i].Title,
			Notes:            docs[i].Notes,
			Filename:         docs[i].OriginalFilename,
		}
		if err := s.psiRepo.CreateDocument(ctx, psiDoc); err != nil {
			ok = false
			// No borrar la fila de la ficha si falló la migración.
		}
	}

	// Solo eliminar las filas de la ficha si todas migraron correctamente.
	if ok {
		_ = s.repo.DeleteInscriptionDocumentsByRequestID(ctx, requestID)
	}
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

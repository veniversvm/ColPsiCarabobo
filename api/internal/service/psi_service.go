// api/internal/service/psi_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
// Este archivo contiene las operaciones centrales relacionadas con los psicólogos colegiados,
// incluyendo la importación masiva desde CSV, la gestión de perfiles públicos y la autenticación.

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"golang.org/x/crypto/bcrypt"
)

// PsiService es la estructura que implementa la lógica de negocio relacionada con los psicólogos.
type PsiService struct {
	repo        domain.PsiUserRepository
	s3Client    *s3.S3Client
	mailService IMailService // Inyectamos el servicio de correo para notificaciones
}

// NewPsiService es el constructor de PsiService, inyectando las dependencias necesarias.
func NewPsiService(repo domain.PsiUserRepository, s3Client *s3.S3Client, mailService IMailService) *PsiService {
	return &PsiService{
		repo:        repo,
		s3Client:    s3Client,
		mailService: mailService,
	}
}

// =========================================================================
// GESTIÓN MASIVA (CSV IMPORT)
// =========================================================================

// ImportFromCSV procesa la carga masiva de agremiados desde un flujo de datos CSV.
// Implementa seguridad mediante el hashing de contraseñas y garantiza la integridad
// mediante transacciones atómicas por cada registro.
func (s *PsiService) ImportFromCSV(ctx context.Context, reader io.Reader, adminID uuid.UUID) (int, []map[string]string) {
	csvReader := csv.NewReader(reader)
	_, _ = csvReader.Read() // Saltar cabeceras

	successCount := 0
	var failedRecords []map[string]string

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// 1. Hash de la contraseña (Seguridad Senior)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(record[2]), bcrypt.DefaultCost)

		// 2. Mapeo del Modelo Principal
		psi := &domain.PsiUserModel{
			AuditModel: domain.AuditModel{
				CreateById: &adminID,
				CreateBy:   "Admin_CSV_Import",
			},
			Username:                 record[0],
			Email:                    record[1],
			Password:                 string(hashedPassword),
			FirstName:                record[3],
			SecondName:               record[4],
			LastName:                 record[5],
			SecondLastName:           record[6],
			FPV:                      parseInt(record[7]),
			CI:                       parseInt(record[8]),
			Nationality:              record[10],
			BornDate:                 parseDate(record[11]),
			Genre:                    record[12],
			ContactEmail:             record[13],
			ShowContactEmail:         parseBool(record[14]),
			PublicPhone:              record[15],
			ShowPublicPhone:          parseBool(record[16]),
			ServiceAddress:           record[17],
			ShowPublicServiceAddress: parseBool(record[18]),
			Solvent:                  parseBool(record[19]),
		}

		// 3. Mapeo de Datos Colegiales
		colData := &domain.PsiUserColData{
			AuditModel: domain.AuditModel{
				CreateById: &adminID,
			},
			UniversityUndergraduate:     record[27],
			ShowUniversityUndergraduate: parseBool(record[14]), // Ejemplo de reuso de lógica
			GraduateDate:                parseDate(record[28]),
			MentionUndergraduate:        record[29],
			RegisterTitleState:          record[30],
			RegisterTitleDate:           parseDate(record[31]),
			RegisterNumber:              parseInt(record[32]),
			RegisterFolio:               record[33],
			RegisterTome:                record[34],
			GuildDirector:               parseBool(record[35]),
			SixtyFiveOrPlus:             parseBool(record[36]),
			DateOfLastSolvency:          parseDate(record[40]),
		}

		// 4. Persistencia mediante el Repositorio (Transaccional)
		err = s.repo.CreateWithColData(ctx, psi, colData)
		if err != nil {
			failedRecords = append(failedRecords, map[string]string{
				"fila":      record[0],
				"identidad": record[8], // CI
				"error":     err.Error(),
			})
			continue
		}

		// 5. Notificación de Bienvenida (No bloqueante)
		mailData := map[string]interface{}{
			"Name":     psi.Username,
			"Email":    psi.Email,
			"Password": record[2],
		}

		// Invocación dinámica y no-bloqueante
		if err := s.mailService.SendEmail(psi.Email, "Bienvenido a la plataforma Colegio de Psicólogos", "welcome_psi", mailData); err != nil {
			log.Printf("⚠️ Error al preparar el correo (pero el psi-user se creó): %v", err)
		}
		successCount++
	}

	return successCount, failedRecords
}

// =========================================================================
// AUTOGESTIÓN Y PRIVACIDAD (SELF-MANAGEMENT)
// =========================================================================

// UpdateProfileSelf permite al psicólogo actualizar sus datos de contacto y visibilidad.
// Implementa "Lazy Loading" para ColData: solo consulta y actualiza la tabla de datos
// colegiales si el usuario solicita cambios en esos campos específicos.
func (s *PsiService) UpdateProfileSelf(ctx context.Context, psi *domain.PsiUserModel, id uuid.UUID, req request_structs.PsiUserUpdateRequestSelf, file *multipart.FileHeader) (*domain.PsiUserModel, error) {

	// 1. VALIDACIÓN DE CONTRASEÑA (Security First)
	if err := bcrypt.CompareHashAndPassword([]byte(psi.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("contraseña actual incorrecta")
	}

	// 2. CAMBIO DE PASSWORD (Si se solicitó)
	if req.NewPassword1 != nil && *req.NewPassword1 != "" {
		if req.NewPassword2 == nil || *req.NewPassword1 != *req.NewPassword2 {
			return nil, errors.New("las nuevas contraseñas no coinciden")
		}
		hashed, _ := bcrypt.GenerateFromPassword([]byte(*req.NewPassword1), bcrypt.DefaultCost)
		psi.Password = string(hashed)
		// Al cambiar password, rotamos la Key para invalidar otras sesiones
		psi.Key = uuid.New().String()
	}

	// 3. MANEJO DE IMAGEN EN S3
	if file != nil {
		// Sanitizar y subir (usando tu utilidad ya existente)
		src, _ := file.Open()
		defer src.Close()

		cleanBytes, ext, contentType, err := utils.SanitizeImage(src)
		if err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("avatars/%s%s", psi.ID.String(), ext)
		newKey, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "avatars", filename, contentType)
		if err != nil {
			return nil, err
		}

		// Si tenía una foto anterior, borrarla de S3 para no dejar basura
		if psi.ProfilePictureS3Key != "" && psi.ProfilePictureS3Key != newKey {
			_ = s.s3Client.DeleteFile(ctx, psi.ProfilePictureS3Key)
		}
		psi.ProfilePictureS3Key = newKey
	}

	// A. Auditoría Automática
	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID

	// B. Mapeo de campos del Modelo Principal (PsiUserModel)
	if req.Username != nil {
		psi.Username = *req.Username
	}

	if req.Email != nil {
		psi.Email = *req.Email
	}

	if req.ContactEmail != nil {
		psi.ContactEmail = *req.ContactEmail
	}
	if req.ShowContactEmail != nil {
		psi.ShowContactEmail = *req.ShowContactEmail
	}
	if req.PublicPhone != nil {
		psi.PublicPhone = *req.PublicPhone
	}
	if req.ShowPublicPhone != nil {
		psi.ShowPublicPhone = *req.ShowPublicPhone
	}
	if req.ServiceAddress != nil {
		psi.ServiceAddress = *req.ServiceAddress
	}
	if req.ShowPublicServiceAddress != nil {
		psi.ShowPublicServiceAddress = *req.ShowPublicServiceAddress
	}

	// Ubicación
	if req.MunicipalityCarabobo != nil {
		psi.MunicipalityCarabobo = *req.MunicipalityCarabobo
	}
	if req.PhoneCarabobo != nil {
		psi.PhoneCarabobo = *req.PhoneCarabobo
	}
	if req.CelPhoneCarabobo != nil {
		psi.CelPhoneCarabobo = *req.CelPhoneCarabobo
	}
	if req.StateOutside != nil {
		psi.StateOutside = *req.StateOutside
	}
	if req.MunicipalityOutSideCarabobo != nil {
		psi.MunicipalityOutSideCarabobo = *req.MunicipalityOutSideCarabobo
	}
	if req.PhoneOutSideCarabobo != nil {
		psi.PhoneOutSideCarabobo = *req.PhoneOutSideCarabobo
	}
	if req.CelPhoneOutSideCarabobo != nil {
		psi.CelPhoneOutSideCarabobo = *req.CelPhoneOutSideCarabobo
	}

	// Profesional
	if req.PrimarySpecialty != nil {
		psi.PrimarySpecialty = *req.PrimarySpecialty
	}
	if req.SecondarySpecialty != nil {
		psi.SecondarySpecialty = *req.SecondarySpecialty
	}
	if req.MiniBio != nil {
		psi.MiniBio = *req.MiniBio
	}

	// C. Lógica Inteligente para ColData
	// Solo traemos y actualizamos ColData si hay cambios relevantes en el request
	var colDataToUpdate *domain.PsiUserColData

	hasColDataChanges := req.ShowUniversityUndergraduate != nil ||
		req.ShowGraduateDate != nil ||
		req.ShowMentionUndergraduate != nil

	if hasColDataChanges {
		// Recuperamos solo los datos colegiales (Lazy Load eficiente)
		currentColData, err := s.repo.GetPsiUserColData(ctx, psi.ID)
		if err != nil {
			return nil, err
		}

		// Aplicamos cambios
		if req.ShowUniversityUndergraduate != nil {
			currentColData.ShowUniversityUndergraduate = *req.ShowUniversityUndergraduate
		}
		if req.ShowGraduateDate != nil {
			currentColData.ShowGraduateDate = *req.ShowGraduateDate
		}
		if req.ShowMentionUndergraduate != nil {
			currentColData.ShowMentionUndergraduate = *req.ShowMentionUndergraduate
		}

		// Auditoría de ColData
		currentColData.UpdateBy = psi.Username
		currentColData.UpdateById = &psi.ID

		colDataToUpdate = currentColData
	}

	// D. Persistencia Transaccional
	err := s.repo.UpdatePublicProfile(ctx, psi, colDataToUpdate)
	if err != nil {
		return nil, err
	}

	return psi, nil
}

// =========================================================================
// CONSULTA PÚBLICA (DIRECTORY & PROFILE)
// =========================================================================

// GetPublicDirectory devuelve una lista paginada de mini-perfiles.
// Aplica normalización de género y oculta datos sensibles de solvencia al público.
func (s *PsiService) GetPublicDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) (interface{}, error) {
	// Normalizar paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 12
	}

	filter.Gender = strings.ToUpper(strings.TrimSpace(filter.Gender))
	if filter.Gender != "M" && filter.Gender != "F" {
		filter.Gender = ""
	}

	users, total, err := s.repo.SearchDirectory(ctx, filter)
	if err != nil {
		return nil, err
	}

	list := make([]request_structs.PsiMiniProfileDTO, 0, len(users))
	for _, u := range users {
		mini := request_structs.PsiMiniProfileDTO{
			ID:             u.ID,
			FirstName:      u.FirstName,
			LastName:       u.LastName,
			CI:             u.CI,
			FPV:            u.FPV,
			ProfilePicture: u.ProfilePictureS3Key,
			MiniBio:        u.MiniBio,
			// Solvent:        u.Solvent, // no mostrar al publico
		}

		// Añadimos las especialidades al mini perfil para que aparezcan en las "cards"
		mini.Specialties = []string{}
		if u.PrimarySpecialty != "" {
			mini.Specialties = append(mini.Specialties, u.PrimarySpecialty)
		}
		if u.SecondarySpecialty != "" {
			mini.Specialties = append(mini.Specialties, u.SecondarySpecialty)
		}

		list = append(list, mini)
	}

	return fiber.Map{
		"data":        list,
		"total":       total,
		"page":        filter.Page,
		"limit":       filter.Limit,
		"total_pages": (total + int64(filter.Limit) - 1) / int64(filter.Limit),
	}, nil
}

// GetPublicProfile construye la ficha técnica del psicólogo aplicando el "Escudo de Privacidad".
// Los datos personales (Email, Teléfono) y académicos (Postgrados) se ocultan dinámicamente
// según la configuración del usuario y su estatus de solvencia institucional.
func (s *PsiService) GetPublicProfile(ctx context.Context, id uuid.UUID) (*request_structs.PsiFullProfileDTO, error) {
	// 1. Obtener datos crudos de la DB (Con Preload de ColData y PostGrades)
	psi, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("psicólogo no encontrado")
	}

	// 2. Verificar si está activo en el sistema general
	if !psi.IsActive {
		return nil, errors.New("perfil no disponible")
	}

	// 3. Inicializar el DTO Público
	dto := &request_structs.PsiFullProfileDTO{
		ID:             psi.ID,
		FirstName:      psi.FirstName,
		LastName:       psi.LastName,
		FPV:            psi.FPV,
		Gender:         psi.Genre,
		ProfilePicture: psi.ProfilePictureS3Key,
		Solvent:        psi.Solvent,
		MiniBio:        psi.MiniBio,
		Specialties:    make([]string, 0),
		PostGrades:     make([]request_structs.PostGradeDTO, 0), // Inicializamos vacío
		SocialNetworks: make([]request_structs.SocialNetworkDTO, 0),
	}

	// --- LÓGICA DE PRIVACIDAD PERSONAL ---

	if psi.ShowContactEmail {
		dto.Email = psi.ContactEmail
	}
	if psi.ShowPublicPhone {
		dto.Phone = psi.PublicPhone
	}
	if psi.ShowPublicServiceAddress {
		dto.Address = psi.ServiceAddress
	}

	// Ubicación
	if psi.MunicipalityCarabobo != "" {
		dto.Location.State = "Carabobo"
		dto.Location.Municipality = psi.MunicipalityCarabobo
	} else {
		dto.Location.State = psi.StateOutside
		dto.Location.Municipality = psi.MunicipalityOutSideCarabobo
	}

	// Especialidades
	if psi.PrimarySpecialty != "" {
		dto.Specialties = append(dto.Specialties, psi.PrimarySpecialty)
	}
	if psi.SecondarySpecialty != "" {
		dto.Specialties = append(dto.Specialties, psi.SecondarySpecialty)
	}

	// Datos Universitarios de Pregrado (Según privacidad del usuario)
	if psi.ColData.ShowUniversityUndergraduate {
		dto.Undergraduate.University = psi.ColData.UniversityUndergraduate
	}
	if psi.ColData.ShowGraduateDate && !psi.ColData.GraduateDate.IsZero() {
		dto.Undergraduate.Date = psi.ColData.GraduateDate.Format("2006-01-02")
	}
	if psi.ColData.ShowMentionUndergraduate {
		dto.Undergraduate.Mention = psi.ColData.MentionUndergraduate
	}

	// --- MAPEO DE REDES SOCIALES ---
	// Generalmente, las redes sociales son públicas, pero si quieres
	// protegerlas por solvencia, envuélvelas en el "if psi.Solvent"
	for _, sn := range psi.SocialNetworks {
		dto.SocialNetworks = append(dto.SocialNetworks, request_structs.SocialNetworkDTO{
			// Usamos nuestra utilidad de normalización para que el front reciba nombres bonitos
			Name: sn.Name,
			URL:  sn.URL,
		})
	}

	// --- LÓGICA DE NEGOCIO INSTITUCIONAL (SOLVENCIA) ---

	// Solo incluimos los Postgrados si el psicólogo está solvente.
	// Si debe cuotas, el arreglo de PostGrades quedará vacío.
	if psi.Solvent {
		for _, pg := range psi.PostGrades {
			// Además, el postgrado debe estar marcado como activo (validado)
			if pg.Active {
				dto.PostGrades = append(dto.PostGrades, request_structs.PostGradeDTO{
					Title:       pg.Title,
					University:  pg.University,
					Year:        pg.GraduationYear,
					Description: pg.Description,
					PicOneURL:   pg.PicOneS3Key,
					PicTwoURL:   pg.PicTwoS3Key,
					PicThreeURL: pg.PicThreeS3Key,
					// No incluimos las fechas de creación o actualización para no saturar el perfil público
				})
			}
		}
	}

	return dto, nil
}

// =========================================================================
// AUTENTICACIÓN Y SESIÓN
// =========================================================================

// Login gestiona el acceso de psicólogos implementando "Key Rotation".
// Al iniciar sesión, se genera un nuevo secreto de firma que invalida físicamente
// cualquier token previo del usuario en otros dispositivos.
func (s *PsiService) Login(ctx context.Context, identifier, password string) (string, error) {
	// 1. Buscar usuario
	psi, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// 2. Verificar si está activo (Soft delete o ban)
	if !psi.IsActive {
		return "", errors.New("cuenta inactiva o suspendida")
	}

	// 3. Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(psi.Password), []byte(password)); err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// 4. ROTACIÓN DE SESIÓN (Seguridad Senior)
	// Generamos una nueva llave. Esto invalida tokens anteriores en otros dispositivos.
	newKey := uuid.New().String()
	psi.Key = newKey

	// Auditoría automática de login
	psi.UpdateBy = psi.Username
	psi.UpdateById = &psi.ID

	if err := s.repo.UpdateKey(ctx, psi); err != nil {
		return "", errors.New("error de sistema al iniciar sesión")
	}

	// 5. Generar JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": psi.ID.String(),
		"role":    "psi", // Rol específico para el middleware híbrido
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	// Notificacion de Login
	mailData := map[string]interface{}{
		"Name":      psi.Username,
		"Email":     psi.Email,
		"LoginTime": time.Now().Format(time.RFC1123),
	}

	// Invocación dinámica y no-bloqueante
	if err := s.mailService.SendEmail(psi.Email, "Colegio de Psicólogos de Carabobo - Inicio de sesión en la plataforma.", "login_psi", mailData); err != nil {
		log.Printf("⚠️ Error al preparar el correo (pero el psicólogo se logueó): %v", err)
	}

	// Firmar con la llave personal del usuario
	return token.SignedString([]byte(newKey))
}

// =========================================================================
// MÓDULO ACADÉMICO (CERTIFICADOS)
// =========================================================================

// AddPostGrade registra un título y gestiona la subida de hasta 3 documentos a S3.
// Implementa saneamiento de imágenes para eliminar metadatos y scripts maliciosos.
func (s *PsiService) AddPostGrade(ctx context.Context, psi *domain.PsiUserModel, req request_structs.CreatePostGradeRequest, files []*multipart.FileHeader) error {

	// Estructura base
	postGrade := &domain.PsiUserPostGrade{
		AuditModel: domain.AuditModel{
			CreateById: &psi.ID,
			CreateBy:   psi.Username,
			UpdateById: &psi.ID,
			UpdateBy:   psi.Username,
		},
		PsiUserID:      psi.ID,
		Title:          req.Title,
		University:     req.University,
		GraduationYear: req.GraduationYear,
		Description:    req.Description,
		Active:         true,
	}

	// Helper interno para procesar subidas
	uploadHelper := func(fh *multipart.FileHeader) (string, error) {
		if fh == nil {
			return "", nil
		}

		// 1. Abrir y Sanitizar (Re-encoding a JPEG/PNG limpio)
		src, err := fh.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		cleanBytes, ext, contentType, err := utils.SanitizeImage(src)
		if err != nil {
			return "", fmt.Errorf("error en imagen: %v", err)
		}

		// 2. Subir a S3
		filename := uuid.New().String() + ext
		// Guardamos en la carpeta 'certificates'
		return s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "certificates", filename, contentType)
	}

	// Procesar las 3 imágenes posibles
	// Asumimos que el slice 'files' viene en orden [pic1, pic2, pic3] desde el handler
	var err error
	if len(files) > 0 {
		postGrade.PicOneS3Key, err = uploadHelper(files[0])
	}
	if err != nil {
		return err
	}

	if len(files) > 1 {
		postGrade.PicTwoS3Key, err = uploadHelper(files[1])
	}
	if err != nil {
		return err
	}

	if len(files) > 2 {
		postGrade.PicThreeS3Key, err = uploadHelper(files[2])
	}
	if err != nil {
		return err
	}

	// Persistir en DB
	return s.repo.CreatePostGrade(ctx, postGrade)
}

// UpdatePostGrade permite editar un título y reemplazar sus imágenes.
// Implementa un reemplazo inteligente: si se sube una nueva imagen, se borra la anterior de S3 para evitar acumulación de archivos huérfanos.
// fileMap: mapa donde la clave es el campo (ej: "pic_one") y el valor es el archivo.
func (s *PsiService) UpdatePostGrade(ctx context.Context, psi *domain.PsiUserModel, pgID uuid.UUID, req request_structs.UpdatePostGradeRequest, fileMap map[string]*multipart.FileHeader) error {

	// 1. Obtener el registro actual
	pg, err := s.repo.GetPostGradeByID(ctx, pgID)
	if err != nil {
		return errors.New("título académico no encontrado")
	}

	// 2. SEGURIDAD: Verificar Propiedad (Ownership Check)
	// Impedir que el Psicólogo A edite el título del Psicólogo B
	if pg.PsiUserID != psi.ID {
		return errors.New("no tienes permiso para editar este registro")
	}

	// 3. Auditoría
	pg.UpdateBy = psi.Username
	pg.UpdateById = &psi.ID
	pg.UpdatedAt = time.Now()

	// 4. Actualización de Campos de Texto (Si vienen en el request)
	if req.Title != nil {
		pg.Title = *req.Title
	}
	if req.University != nil {
		pg.University = *req.University
	}
	if req.GraduationYear != nil {
		pg.GraduationYear = *req.GraduationYear
	}
	if req.Description != nil {
		pg.Description = *req.Description
	}

	// 5. GESTIÓN DE IMÁGENES (Reemplazo Inteligente)
	// Helper para no repetir código: Sube nueva -> Borra vieja -> Retorna nueva Key
	replaceImage := func(newFile *multipart.FileHeader, oldKey string) (string, error) {
		// A. Sanitizar y leer
		src, err := newFile.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		cleanBytes, ext, contentType, err := utils.SanitizeImage(src)
		if err != nil {
			return "", err
		}

		// B. Subir nueva
		filename := uuid.New().String() + ext
		newKey, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "certificates", filename, contentType)
		if err != nil {
			return "", err
		}

		// C. Borrar vieja (Si existía)
		if oldKey != "" {
			// No bloqueamos si falla el borrado, solo logueamos (idealmente)
			_ = s.s3Client.DeleteFile(ctx, oldKey)
		}
		return newKey, nil
	}

	// Aplicar para cada slot si viene un archivo
	if file, ok := fileMap["pic_one"]; ok {
		pg.PicOneS3Key, err = replaceImage(file, pg.PicOneS3Key)
		if err != nil {
			return err
		}
	}
	if file, ok := fileMap["pic_two"]; ok {
		pg.PicTwoS3Key, err = replaceImage(file, pg.PicTwoS3Key)
		if err != nil {
			return err
		}
	}
	if file, ok := fileMap["pic_three"]; ok {
		pg.PicThreeS3Key, err = replaceImage(file, pg.PicThreeS3Key)
		if err != nil {
			return err
		}
	}

	// 6. Persistir cambios
	return s.repo.UpdatePostGrade(ctx, pg)
}

// =========================================================================
// --- HELPERS DE CONVERSIÓN (Privados) ---
// =========================================================================
func parseInt(s string) int {
	val, _ := strconv.Atoi(strings.TrimSpace(s))
	return val
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "v" || s == "s"
}

func parseDate(s string) time.Time {
	layout := "2006-01-02" // Formato estándar del CSV que pasaste
	t, err := time.Parse(layout, strings.TrimSpace(s))
	if err != nil {
		return time.Time{} // Fecha cero si falla
	}
	return t
}

// api/internal/service/psi_user_admin_service.go

// Package service implementa la capa de lógica de negocio (Business Logic Layer).
// Este archivo contiene las operaciones administrativas de alto nivel para la gestión
// de los expedientes de los psicólogos colegiados.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// GESTIÓN DE EXPEDIENTES (LECTURA DETALLADA)
// =========================================================================

// GetPsiByIDAdmin retorna el expediente completo de un psicólogo sin restricciones de privacidad.
// Está diseñado exclusivamente para el uso del personal administrativo autorizado.
// Implementa una barrera de permisos RBAC interna para asegurar que solo personal de gestión acceda.
func (s *PsiService) GetPsiByIDAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) (*domain.PsiUserModel, error) {
	// Como es una operación de solo lectura para el panel interno,
	// verificamos que sea Sudo o que al menos tenga un permiso administrativo básico.
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, errors.New("permisos insuficientes para ver expedientes detallados")
	}

	psi, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return nil, errors.New("psicólogo no encontrado")
	}

	return psi, nil
}

// =========================================================================
// REGISTRO INDIVIDUAL (ESCRITURA ATÓMICA)
// =========================================================================

// CreatePsiByAdmin orquestador de la creación manual de un nuevo colegiado.
// Realiza el hashing de credenciales, el parseo de fechas de registro y la vinculación
// transaccional entre el perfil de usuario y sus datos de grado universitario.
func (s *PsiService) CreatePsiByAdmin(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreatePsiAdminRequest) error {
	// 1. Validar Permisos
	if !admin.CanCreatePsi && !admin.Sudo {
		return errors.New("no tienes permiso para registrar psicólogos")
	}

	// 2. Hash de Password
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	// 3. Parsear fechas (Helper de validación)
	bornDate, _ := time.Parse("2006-01-02", req.BornDate)
	gradDate, _ := time.Parse("2006-01-02", req.GraduateDate)
	solvDate, _ := time.Parse("2006-01-02", req.DateOfLastSolvency)
	regDate, _ := time.Parse("2006-01-02", req.RegisterTitleDate)

	psiID := uuid.New()

	// 4. Mapeo Modelo Principal
	psi := &domain.PsiUserModel{
		ID:                 psiID,
		Username:           req.Username,
		Email:              req.Email,
		Password:           string(hashed),
		FirstName:          req.FirstName,
		SecondName:         req.SecondName,
		LastName:           req.LastName,
		SecondLastName:     req.SecondLastName,
		CI:                 req.CI,
		FPV:                req.FPV,
		BornDate:           bornDate,
		Genre:              req.Genre,
		Nationality:        req.Nationality,
		Solvent:            req.Solvent,
		ProofOfLife:        req.ProofOfLife,
		IsActive:           req.IsActive,
		ContactEmail:       req.ContactEmail,
		PublicPhone:        req.PublicPhone,
		ServiceAddress:     req.ServiceAddress,
		PrimarySpecialty:   req.PrimarySpecialty,
		SecondarySpecialty: req.SecondarySpecialty,
		Key:                uuid.New().String(),
		AuditModel: domain.AuditModel{
			CreateBy: admin.Username, CreateById: &admin.ID,
			UpdateBy: admin.Username, UpdateById: &admin.ID,
		},
	}

	// 5. Mapeo ColData
	colData := &domain.PsiUserColData{
		PsiUserModelID:          psiID,
		UniversityUndergraduate: req.UniversityUndergraduate,
		GraduateDate:            gradDate,
		MentionUndergraduate:    req.MentionUndergraduate,
		RegisterNumber:          req.RegisterNumber,
		RegisterTitleState:      req.RegisterTitleState,
		RegisterTitleDate:       regDate,
		RegisterFolio:           req.RegisterFolio,
		RegisterTome:            req.RegisterTome,
		GuildDirector:           req.GuildDirector,
		SixtyFiveOrPlus:         req.SixtyFiveOrPlus,
		GuildCollaborator:       req.GuildCollaborator,
		PublicEmployee:          req.PublicEmployee,
		UniversityProfessor:     req.UniversityProfessor,
		DateOfLastSolvency:      solvDate,
		DoubleGuild:             req.DoubleGuild,
		CPSM:                    req.CPSM,
		AuditModel: domain.AuditModel{
			CreateBy: admin.Username, CreateById: &admin.ID,
			UpdateBy: admin.Username, UpdateById: &admin.ID,
		},
	}

	return s.repo.CreateWithColData(ctx, psi, colData)
}

// =========================================================================
// ACTUALIZACIÓN MAESTRA (CONTROL TOTAL)
// =========================================================================

// UpdatePsiByAdmin permite a un administrador modificar íntegramente el expediente.
// Soporta actualizaciones parciales (PATCH) mediante el uso de punteros en el DTO,
// garantizando que solo los campos enviados en el JSON sean alterados en la base de datos.
func (s *PsiService) UpdatePsiByAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID, req request_structs.UpdatePsiAdminRequest) error {
	// 1. VALIDACIÓN DE PERMISOS (RBAC)
	if !admin.CanUpdatePsi && !admin.Sudo {
		return errors.New("no tienes permiso para editar registros de psicólogos")
	}

	// 2. OBTENER REGISTRO ACTUAL
	// El repositorio GetByID ya debe incluir el Preload("ColData")
	psi, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("error al recuperar el psicólogo: %w", err)
	}

	// Helper local para parsear fechas de forma segura durante el mapeo
	parseDate := func(dateStr *string) time.Time {
		if dateStr == nil || *dateStr == "" {
			return time.Time{}
		}
		t, _ := time.Parse("2006-01-02", *dateStr)
		return t
	}

	// 3. MAPEO DE TABLA PRINCIPAL (PsiUserModel)

	// Identidad y Filiación
	if req.FirstName != nil {
		psi.FirstName = *req.FirstName
	}
	if req.SecondName != nil {
		psi.SecondName = *req.SecondName
	}
	if req.LastName != nil {
		psi.LastName = *req.LastName
	}
	if req.SecondLastName != nil {
		psi.SecondLastName = *req.SecondLastName
	}
	if req.Email != nil {
		psi.Email = *req.Email
	}
	if req.CI != nil {
		psi.CI = *req.CI
	}
	if req.FPV != nil {
		psi.FPV = *req.FPV
	}
	if req.Genre != nil {
		psi.Genre = *req.Genre
	}
	if req.Nationality != nil {
		psi.Nationality = *req.Nationality
	}
	if req.BornDate != nil {
		psi.BornDate = parseDate(req.BornDate)
	}

	// Estatus Administrativo
	if req.Solvent != nil {
		psi.Solvent = *req.Solvent
	}
	if req.ProofOfLife != nil {
		psi.ProofOfLife = *req.ProofOfLife
	}
	if req.IsActive != nil {
		psi.IsActive = *req.IsActive
	}

	// Datos de Contacto y Privacidad
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
	if req.ShowServiceAddress != nil {
		psi.ShowPublicServiceAddress = *req.ShowServiceAddress
	}

	// Ubicación (Carabobo / Exterior)
	if req.MunicipalityOutSideCarabobo != nil {
		psi.MunicipalityOutSideCarabobo = *req.MunicipalityOutSideCarabobo
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

	// Especialidades y Bio
	if req.PrimarySpecialty != nil {
		psi.PrimarySpecialty = *req.PrimarySpecialty
	}
	if req.SecondarySpecialty != nil {
		psi.SecondarySpecialty = *req.SecondarySpecialty
	}
	if req.MiniBio != nil {
		psi.MiniBio = *req.MiniBio
	}
	if req.FullBio != nil {
		// Aquí se podría implementar la lógica para actualizar el TextModel si fuera necesario
	}

	// 4. MAPEO DE TABLA RELACIONADA (PsiUserColData)

	// Información Académica de Pregrado
	if req.UniversityUndergraduate != nil {
		psi.ColData.UniversityUndergraduate = *req.UniversityUndergraduate
	}
	if req.GraduateDate != nil {
		psi.ColData.GraduateDate = parseDate(req.GraduateDate)
	}
	if req.MentionUndergraduate != nil {
		psi.ColData.MentionUndergraduate = *req.MentionUndergraduate
	}

	// Registro de Título
	if req.RegisterNumber != nil {
		psi.ColData.RegisterNumber = *req.RegisterNumber
	}
	if req.RegisterTitleState != nil {
		psi.ColData.RegisterTitleState = *req.RegisterTitleState
	}
	if req.RegisterTitleDate != nil {
		psi.ColData.RegisterTitleDate = parseDate(req.RegisterTitleDate)
	}
	if req.RegisterFolio != nil {
		psi.ColData.RegisterFolio = *req.RegisterFolio
	}
	if req.RegisterTome != nil {
		psi.ColData.RegisterTome = *req.RegisterTome
	}

	// Flags Profesionales y Gremiales
	if req.GuildDirector != nil {
		psi.ColData.GuildDirector = *req.GuildDirector
	}
	if req.SixtyFiveOrPlus != nil {
		psi.ColData.SixtyFiveOrPlus = *req.SixtyFiveOrPlus
	}
	if req.GuildCollaborator != nil {
		psi.ColData.GuildCollaborator = *req.GuildCollaborator
	}
	if req.PublicEmployee != nil {
		psi.ColData.PublicEmployee = *req.PublicEmployee
	}
	if req.UniversityProfessor != nil {
		psi.ColData.UniversityProfessor = *req.UniversityProfessor
	}
	if req.DoubleGuild != nil {
		psi.ColData.DoubleGuild = *req.DoubleGuild
	}
	if req.CPSM != nil {
		psi.ColData.CPSM = *req.CPSM
	}
	if req.DateOfLastSolvency != nil {
		psi.ColData.DateOfLastSolvency = parseDate(req.DateOfLastSolvency)
	}

	// 5. ACTUALIZACIÓN DE AUDITORÍA
	// Marcamos ambas entidades con la identidad del administrador que realiza el cambio.
	now := time.Now()

	psi.UpdateBy = admin.Username
	psi.UpdateById = &admin.ID
	psi.UpdatedAt = now

	psi.ColData.UpdateBy = admin.Username
	psi.ColData.UpdateById = &admin.ID
	psi.ColData.UpdatedAt = now

	// 6. PERSISTENCIA
	// Llamamos al repositorio para guardar ambas estructuras dentro de una transacción.
	if err := s.repo.Update(ctx, psi, &psi.ColData); err != nil {
		return fmt.Errorf("error al persistir los cambios: %w", err)
	}

	return nil
}

// =========================================================================
// DESTRUCCIÓN LÓGICA (SOFT DELETE)
// =========================================================================

// DeletePsiByAdmin realiza un borrado lógico del registro.
// El sistema preserva la integridad referencial para propósitos legales históricos,
// pero el usuario deja de existir para el motor de búsqueda y la autenticación.
func (s *PsiService) DeletePsiByAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) error {
	// 1. Validación de permisos (RBAC)
	if !admin.CanDeletePsi && !admin.Sudo {
		return errors.New("no tienes permiso para eliminar psicólogos")
	}

	// 2. Ejecutar borrado
	if err := s.repo.Delete(ctx, targetID); err != nil {
		return fmt.Errorf("error al eliminar el registro: %w", err)
	}

	return nil
}

// =========================================================================
// DASHBOARD Y LISTADOS
// =========================================================================

// GetAdminDirectory proporciona una vista de "Rayos X" del listado de miembros.
// Ignora filtros de solvencia y visibilidad pública, permitiendo al staff administrativo
// gestionar la morosidad y estados de actividad.
func (s *PsiService) GetAdminDirectory(ctx context.Context, admin *domain.UserAdmin, filter request_structs.PsiDirectoryFilterDTO) (interface{}, error) {
	// Seguridad: Validar que sea administrador autorizado
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, errors.New("permisos insuficientes para listar agremiados")
	}

	// Normalizar paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 12
	}

	users, total, err := s.repo.SearchAdmin(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Mapeo al DTO Administrativo
	list := make([]request_structs.PsiAdminListDTO, 0, len(users))
	for _, u := range users {
		list = append(list, request_structs.PsiAdminListDTO{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			CI:        u.CI,
			FPV:       u.FPV,
			Email:     u.Email,
			Solvent:   u.Solvent,
			IsActive:  u.IsActive,
		})
	}

	return fiber.Map{
		"data":        list,
		"total":       total,
		"page":        filter.Page,
		"limit":       filter.Limit,
		"total_pages": (total + int64(filter.Limit) - 1) / int64(filter.Limit),
	}, nil
}

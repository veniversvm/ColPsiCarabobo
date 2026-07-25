// api/internal/service/specialty_service.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Esta capa actúa como orquestador estricto entre los controladores (Handlers HTTP)
// y la persistencia (Repository). Su propósito arquitectónico es asegurar que
// las políticas de seguridad, auditoría y reglas de negocio del Colegio se
// cumplan incondicionalmente antes de afectar el estado de la base de datos.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// SpecialtyService gestiona las reglas de negocio para el Catálogo Maestro de Especialidades.
//
// Master Data Management (MDM):
// Las "Especialidades" o "Áreas de Trabajo" son datos maestros categóricos.
// Alterar este catálogo tiene efectos en cascada sobre los perfiles de todos los psicólogos
// y en los motores de búsqueda del portal. Este servicio blinda esas operaciones.
type SpecialtyService struct {
	repo domain.SpecialtyRepository
}

// NewSpecialtyService crea una nueva instancia del servicio inyectando sus dependencias (DI).
// Favorece el bajo acoplamiento al depender de la interfaz `domain.SpecialtyRepository`
// en lugar de una implementación concreta de base de datos.
func NewSpecialtyService(repo domain.SpecialtyRepository) *SpecialtyService {
	return &SpecialtyService{repo: repo}
}

// =========================================================================
// ESCRITURA Y GESTIÓN ADMINISTRATIVA
// =========================================================================

// Create registra una nueva especialidad en el catálogo maestro.
//
// Trazabilidad Forense (Audit Trail):
// Implementa la política de auditoría inicial asignando al administrador responsable
// en el mismo instante de la creación, garantizando que no existan registros "huérfanos"
// o anónimos en el sistema.
func (s *SpecialtyService) Create(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreateSpecialtyRequest) error {
	// SEGURIDAD (Gatekeeping): Validación de permisos granulares (RBAC).
	if !admin.CanCreateTags && !admin.Sudo {
		return domain.ErrInsufficientPerms
	}

	newSpec := &domain.PsiSpecialtyModel{
		Name:        req.Name,
		Description: req.Description,
		Active:      true,
		// Inicia campos de auditoría de forma atómica para rastreabilidad total.
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
	}

	return s.repo.Create(ctx, newSpec)
}

// Update coordina la mutación parcial de una especialidad (Semántica PATCH).
//
// Uso de Punteros:
// Permite al cliente enviar únicamente el campo que desea modificar (ej. solo el 'Name').
// Si el cliente envía un campo, este sobrescribe el estado en memoria; si no lo envía (nil),
// el servicio preserva el valor original traído desde la base de datos.
func (s *SpecialtyService) Update(ctx context.Context, admin *domain.UserAdmin, id uint32, req request_structs.UpdateSpecialtyRequest) error {
	// SEGURIDAD: Control de acceso para modificación
	if !admin.CanEditTags && !admin.Sudo {
		return domain.ErrInsufficientPerms
	}

	// Lectura Previa (Read-before-Write): Necesaria para no perder datos al hacer Save.
	spec, err := s.repo.GetByAdminID(ctx, id)
	if err != nil {
		return err
	}

	// Mapeo selectivo: solo se alteran los campos presentes explícitamente en el request.
	if req.Name != nil {
		spec.Name = *req.Name
	}
	if req.Description != nil {
		spec.Description = *req.Description
	}
	if req.Active != nil {
		spec.Active = *req.Active
	}

	// TRAZABILIDAD (Non-Repudiation): Garantiza que sepamos quién realizó el último
	// cambio y en qué milisegundo, sobreescribiendo las estampas previas.
	spec.UpdateBy = admin.Username
	spec.UpdateById = &admin.ID
	spec.UpdatedAt = time.Now()

	return s.repo.Update(ctx, spec)
}

// Delete ejecuta una eliminación lógica (Soft Delete).
//
// Preservación de Integridad Referencial:
// Protege al sistema de violaciones de Claves Foráneas (Foreign Keys). Si se eliminara
// físicamente la especialidad "Psicología Clínica", todos los psicólogos asociados a ella
// quedarían huérfanos o corrompidos. El Soft-Delete simplemente la "apaga" de las búsquedas futuras.
func (s *SpecialtyService) Delete(ctx context.Context, admin *domain.UserAdmin, id uint32) error {
	if !admin.CanDeleteTags && !admin.Sudo {
		return domain.ErrInsufficientPerms
	}

	return s.repo.Delete(ctx, id)
}

// =========================================================================
// CONSULTAS Y VISIBILIDAD
// =========================================================================

// GetSpecialties retorna la colección de taxonomías (Especialidades).
//
// Patrón de Falla Segura (Fail-Safe Defaults / Secure by Default):
// Si el rol que realiza la consulta no es un administrador autorizado (`isAdmin == false`),
// el sistema asume una postura defensiva e ignora cualquier parámetro que intente
// forzar la lectura de datos inactivos, sobrescribiendo el filtro a "active".
func (s *SpecialtyService) GetSpecialties(ctx context.Context, requestedStatus string, isAdmin bool) ([]domain.PsiSpecialtyModel, error) {
	finalStatus := "active"

	// Solo los administradores (contexto de confianza) pueden explorar estados 'inactive' o 'all'.
	if isAdmin {
		finalStatus = requestedStatus
	}

	return s.repo.GetAll(ctx, finalStatus)
}

// Count retorna estadísticas absolutas del volumen de registros en el catálogo.
//
// Control de Fuga de Información Telemétrica:
// Evita que un visitante anónimo o usuario no privilegiado averigüe cuántas especialidades
// "ocultas" o "deprecadas" tiene el Colegio, forzando la métrica a devolver únicamente
// el conteo de entidades públicas.
func (s *SpecialtyService) Count(ctx context.Context, active *bool, admin *domain.UserAdmin) (int64, error) {
	// Regla de Protección: Si el admin es nulo o carece de permisos de lectura, forzar 'solo activos'.
	if admin == nil || (!admin.Sudo && !admin.CanReadNotifications) {
		onlyActive := true
		return s.repo.Count(ctx, &onlyActive)
	}

	return s.repo.Count(ctx, active)
}

// GetByID recupera una especialidad validando la integridad básica del identificador.
// Diseñado para consumo general (Bypass explícito de validación 'active = false'
// en este punto, dependiendo de la implementación interna del repositorio).
func (s *SpecialtyService) GetByID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	if id < 1 {
		return nil, errors.New("ID de especialidad inválido")
	}
	return s.repo.GetByID(ctx, id, false)
}

// GetAllAdmin es una vista de "Rayos X" del catálogo completo.
// Permite a las interfaces del Panel de Control administrativo renderizar
// la totalidad de la base de datos ignorando los filtros de visibilidad pública.
func (s *SpecialtyService) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {
	return s.repo.GetAllAdmin(ctx)
}

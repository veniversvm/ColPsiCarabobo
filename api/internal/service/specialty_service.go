// api/internal/service/specialty_service.go

// Package service implementa la lógica de negocio central de la aplicación.
// Esta capa actúa como orquestador entre los controladores (Handlers) y la persistencia (Repository),
// asegurando que las reglas del Colegio se cumplan antes de afectar la base de datos.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// SpecialtyService gestiona las reglas de negocio para el catálogo maestro de especialidades.
// Se encarga de validar permisos administrativos y mantener la consistencia de la auditoría.
type SpecialtyService struct {
	repo domain.SpecialtyRepository
}

// NewSpecialtyService crea una nueva instancia del servicio inyectando sus dependencias (DI).
func NewSpecialtyService(repo domain.SpecialtyRepository) *SpecialtyService {
	return &SpecialtyService{repo: repo}
}

// =========================================================================
// ESCRITURA Y GESTIÓN ADMINISTRATIVA
// =========================================================================

// Create registra una nueva especialidad validando permisos específicos.
// Implementa la política de auditoría inicial asignando al administrador responsable
// tanto en la creación como en la primera edición.
func (s *SpecialtyService) Create(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreateSpecialtyRequest) error {
	// SEGURIDAD: Validación de permisos granulares (RBAC).
	if !admin.CanCreateTags && !admin.Sudo {
		return errors.New("no tienes permiso para crear especialidades")
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

// Update modifica una especialidad y actualiza obligatoriamente la auditoría.
// Soporta actualizaciones parciales (PATCH) mediante el uso de punteros en el DTO.
func (s *SpecialtyService) Update(ctx context.Context, admin *domain.UserAdmin, id uint32, req request_structs.UpdateSpecialtyRequest) error {
	if !admin.CanEditTags && !admin.Sudo {
		return errors.New("no tienes permiso para editar especialidades")
	}

	spec, err := s.repo.GetByAdminID(ctx, id)
	if err != nil {
		return err
	}

	// Mapeo selectivo: solo se alteran los campos presentes en el request.
	if req.Name != nil {
		spec.Name = *req.Name
	}
	if req.Description != nil {
		spec.Description = *req.Description
	}
	if req.Active != nil {
		spec.Active = *req.Active
	}

	// TRAZABILIDAD: Garantiza que sepamos quién realizó el último cambio y cuándo.
	spec.UpdateBy = admin.Username
	spec.UpdateById = &admin.ID
	spec.UpdatedAt = time.Now()

	return s.repo.Update(ctx, spec)
}

// Delete ejecuta una eliminación lógica (Soft Delete) validando permisos de destrucción.
func (s *SpecialtyService) Delete(ctx context.Context, admin *domain.UserAdmin, id uint32) error {
	if !admin.CanDeleteTags && !admin.Sudo {
		return errors.New("no tienes permiso para eliminar especialidades")
	}

	return s.repo.Delete(ctx, id)
}

// =========================================================================
// CONSULTAS Y VISIBILIDAD
// =========================================================================

// GetSpecialties aplica reglas de visibilidad según el rol del solicitante.
// Implementa una "Falla Segura": si no es admin, el sistema fuerza la visualización
// exclusiva de registros activos, evitando fugas de información interna.
func (s *SpecialtyService) GetSpecialties(ctx context.Context, requestedStatus string, isAdmin bool) ([]domain.PsiSpecialtyModel, error) {
	finalStatus := "active"

	// Solo los administradores pueden explorar estados 'inactive' o 'all'.
	if isAdmin {
		finalStatus = requestedStatus
	}

	return s.repo.GetAll(ctx, finalStatus)
}

// Count retorna estadísticas de registros aplicando filtros de privacidad.
// Los usuarios restringidos reciben el conteo de registros públicos únicamente.
func (s *SpecialtyService) Count(ctx context.Context, active *bool, admin *domain.UserAdmin) (int64, error) {
	// Regla de Protección: Si el admin es nulo o carece de permisos de lectura, forzar 'solo activos'.
	if admin == nil || (!admin.Sudo && !admin.CanReadNotifications) {
		onlyActive := true
		return s.repo.Count(ctx, &onlyActive)
	}

	return s.repo.Count(ctx, active)
}

// GetByID recupera una especialidad validando la integridad del identificador.
func (s *SpecialtyService) GetByID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	if id < 1 {
		return nil, errors.New("ID de especialidad inválido")
	}
	return s.repo.GetByID(ctx, id, false)
}

// GetAllAdmin es un método de conveniencia para obtener el catálogo completo bypassando filtros.
func (s *SpecialtyService) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {
	return s.repo.GetAllAdmin(ctx)
}

// Package service implementa la lógica de negocio central de la aplicación.
// Esta capa actúa como orquestador entre los controladores (Handlers) y la persistencia (Repository).
package service

import (
	"context"
	"errors"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// SpecialtyService gestiona las reglas de negocio para el catálogo de especialidades.
type SpecialtyService struct {
	repo domain.SpecialtyRepository
}

// NewSpecialtyService crea una nueva instancia del servicio inyectando sus dependencias.
func NewSpecialtyService(repo domain.SpecialtyRepository) *SpecialtyService {
	return &SpecialtyService{repo: repo}
}

// Create registra una nueva especialidad validando permisos administrativos.
// Implementa la política de auditoría inicial asignando al creador como primer editor.
func (s *SpecialtyService) Create(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreateSpecialtyRequest) error {
	// Seguridad: Validar que el administrador tenga permiso de gestión de etiquetas/catálogos.
	if !admin.CanCreateTags && !admin.Sudo {
		return errors.New("no tienes permiso para crear especialidades")
	}

	newSpec := &domain.PsiSpecialtyModel{
		Name:        req.Name,
		Description: req.Description,
		Active:      true,
		// Inicia campos de auditoría de forma atómica.
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

// GetSpecialties aplica reglas de visibilidad según el rol del solicitante.
// Si el usuario no es administrador, el sistema fuerza la visualización exclusiva de registros activos,
// protegiendo la integridad de la información pública.
func (s *SpecialtyService) GetSpecialties(ctx context.Context, requestedStatus string, isAdmin bool) ([]domain.PsiSpecialtyModel, error) {
	finalStatus := "active"

	// Solo los administradores pueden ver estados 'inactive' o 'all'.
	if isAdmin {
		finalStatus = requestedStatus
	}

	return s.repo.GetAll(ctx, finalStatus)
}

// Count retorna estadísticas de registros aplicando filtros de privacidad.
// Los usuarios restringidos reciben el conteo de registros públicos únicamente.
func (s *SpecialtyService) Count(ctx context.Context, active *bool, admin *domain.UserAdmin) (int64, error) {
	// Regla de Falla Segura: Si el admin es nulo o no tiene permisos, forzar 'solo activos'.
	if admin == nil || (!admin.Sudo && !admin.CanReadNotifications) {
		onlyActive := true
		return s.repo.Count(ctx, &onlyActive)
	}

	// El administrador autorizado puede ver el conteo total (active=nil), activos o inactivos.
	return s.repo.Count(ctx, active)
}

// GetByID recupera una especialidad validando la integridad del identificador.
func (s *SpecialtyService) GetByID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	if id < 1 {
		return nil, errors.New("ID de especialidad inválido")
	}
	return s.repo.GetByID(ctx, id)
}

// Update modifica una especialidad existente y actualiza obligatoriamente la auditoría.
// Utiliza punteros en el request para permitir actualizaciones parciales (PATCH).
func (s *SpecialtyService) Update(ctx context.Context, admin *domain.UserAdmin, id uint32, req request_structs.UpdateSpecialtyRequest) error {
	if !admin.CanEditTags && !admin.Sudo {
		return errors.New("no tienes permiso para editar especialidades")
	}

	spec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Mapeo selectivo de campos.
	if req.Name != nil {
		spec.Name = *req.Name
	}
	if req.Description != nil {
		spec.Description = *req.Description
	}
	if req.Active != nil {
		spec.Active = *req.Active
	}

	// FIX AUDITORÍA: Garantizar que sepamos quién realizó el último cambio.
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

// GetAllAdmin es un método de conveniencia para obtener el catálogo completo sin filtros de privacidad.
func (s *SpecialtyService) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {
	return s.repo.GetAllAdmin(ctx)
}

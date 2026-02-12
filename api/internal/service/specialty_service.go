package service

import (
	"context"
	"errors"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

type SpecialtyService struct {
	repo domain.SpecialtyRepository
}

func NewSpecialtyService(repo domain.SpecialtyRepository) *SpecialtyService {
	return &SpecialtyService{repo: repo}
}

func (s *SpecialtyService) Create(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreateSpecialtyRequest) error {
	if !admin.CanCreateTags && !admin.Sudo {
		return errors.New("no tienes permiso para crear especialidades")
	}

	newSpec := &domain.PsiSpecialtyModel{
		Name:        req.Name,
		Description: req.Description,
		Active:      true,
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
	}

	if err := s.repo.Create(ctx, newSpec); err != nil {
		return err
	}
	return nil
}

// GetSpecialties es para el directorio (Usa caché)
func (s *SpecialtyService) GetSpecialties(ctx context.Context, requestedStatus string, isAdmin bool) ([]domain.PsiSpecialtyModel, error) {
	// 1. REGLA DE ORO: Seguridad de acceso
	// Si no es admin, ignoramos su petición y forzamos 'active'
	finalStatus := "active"
	if isAdmin {
		finalStatus = requestedStatus // El admin puede elegir: active, inactive, all
	}

	// 4. Consultar al repo
	list, err := s.repo.GetAll(ctx, finalStatus)
	if err != nil {
		return nil, err
	}

	return list, nil

}

// Count retorna el conteo de especialidades aplicando reglas de visibilidad.
func (s *SpecialtyService) Count(ctx context.Context, active *bool, admin *domain.UserAdmin) (int64, error) {
	// 1. REGLA DE VISIBILIDAD PÚBLICA / RESTRICTA
	// Si no hay admin, o el admin no tiene rango suficiente:
	// FORZAMOS el conteo solo de registros activos.
	if admin == nil || (!admin.Sudo && !admin.CanReadNotifications) {
		onlyActive := true
		return s.repo.Count(ctx, &onlyActive)
	}

	// 2. REGLA DE VISIBILIDAD ADMINISTRATIVA
	// Si llegamos aquí, es un Admin autorizado. Respetamos su filtro:
	// - active = true  -> cuenta activas
	// - active = false -> cuenta inactivas
	// - active = nil   -> cuenta todas
	return s.repo.Count(ctx, active)
}

func (s *SpecialtyService) GetByID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	if id < 1 {
		return nil, errors.New("ID inválido")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *SpecialtyService) Update(ctx context.Context, admin *domain.UserAdmin, id uint32, req request_structs.UpdateSpecialtyRequest) error {
	if !admin.CanEditTags && !admin.Sudo {
		return errors.New("no tienes permiso para editar especialidades")
	}

	spec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != nil {
		spec.Name = *req.Name
	}
	if req.Description != nil {
		spec.Description = *req.Description
	}
	if req.Active != nil {
		spec.Active = *req.Active
	}

	err = s.repo.Update(ctx, spec)
	if err != nil {
		return err
	}
	return nil
}

func (s *SpecialtyService) Delete(ctx context.Context, admin *domain.UserAdmin, id uint32) error {
	if !admin.CanDeleteTags && !admin.Sudo {
		return errors.New("no tienes permiso para eliminar especialidades")
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *SpecialtyService) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {

	return s.repo.GetAllAdmin(ctx)
}

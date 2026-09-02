// api/internal/service/observaciones.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona el submódulo de Observaciones Internas: notas internas
// del Colegio sobre un psicólogo. Por diseño, el psicólogo NUNCA puede ver ni
// gestionar sus propias observaciones; el acceso es exclusivo del personal
// administrativo autorizado.
package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// AddObservacionesByAdmin registra una nueva observación interna sobre un psicólogo.
func (s *PsiService) AddObservacionesByAdmin(ctx context.Context, admin *domain.UserAdmin, psiID uuid.UUID, req request_structs.CreateObservacionesRequest) error {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return domain.ErrInsufficientPerms
	}

	// 2. INTEGRIDAD REFERENCIAL
	if _, err := s.repo.GetByID(ctx, psiID); err != nil {
		return domain.ErrPsiNotFound
	}

	// 3. NORMALIZACIÓN Y SANITIZACIÓN
	content := strings.TrimSpace(bluemonday.StrictPolicy().Sanitize(req.Content))
	if content == "" {
		return domain.ErrInvalidRequest
	}
	if len(content) > 10000 {
		return domain.ErrInvalidRequest
	}

	// 4. PERSISTENCIA CON AUDITORÍA DEL EJECUTOR
	entry := &domain.PsiObservations{
		ID: uuid.Must(uuid.NewV7()),
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
		PsiUserID: psiID,
		Content:   content,
	}

	return s.repo.CreateObservations(ctx, entry)
}

// ListObservacionesByPsiID recupera las observaciones internas de un psicólogo.
func (s *PsiService) ListObservacionesByPsiID(ctx context.Context, admin *domain.UserAdmin, psiID uuid.UUID) ([]domain.PsiObservations, error) {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, domain.ErrInsufficientPerms
	}

	// 2. INTEGRIDAD REFERENCIAL
	if _, err := s.repo.GetByID(ctx, psiID); err != nil {
		return nil, domain.ErrPsiNotFound
	}

	return s.repo.ListObservationsByPsiID(ctx, psiID)
}

// UpdateObservacionesByAdmin edita el contenido de una observación interna existente.
func (s *PsiService) UpdateObservacionesByAdmin(ctx context.Context, admin *domain.UserAdmin, entryID uuid.UUID, req request_structs.UpdateObservacionesRequest) error {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return domain.ErrInsufficientPerms
	}

	// 2. EXISTENCIA (404 si no está)
	if _, err := s.repo.GetObservationsByID(ctx, entryID); err != nil {
		return domain.ErrObservacionesNotFound
	}

	// 3. EL CONTENIDO ES OBLIGATORIO EN EL PATCH
	if req.Content == nil {
		return domain.ErrInvalidRequest
	}

	// 4. NORMALIZACIÓN Y SANITIZACIÓN
	content := strings.TrimSpace(bluemonday.StrictPolicy().Sanitize(*req.Content))
	if content == "" {
		return domain.ErrInvalidRequest
	}
	if len(content) > 10000 {
		return domain.ErrInvalidRequest
	}

	// 5. PERSISTENCIA CON AUDITORÍA DEL EJECUTOR
	return s.repo.UpdateObservations(ctx, entryID, content, admin.Username, admin.ID)
}

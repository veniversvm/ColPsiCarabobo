// api/internal/service/deontologia.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona el submódulo de Expediente Deontológico: entradas internas
// del Colegio sobre un psicólogo (expedientes disciplinarios del Tribunal).
// Por diseño, el psicólogo NUNCA puede ver ni gestionar su propio expediente;
// el acceso es exclusivo del personal administrativo autorizado.
package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// AddDeontologiaByAdmin registra una nueva entrada deontológica sobre un psicólogo.
func (s *PsiService) AddDeontologiaByAdmin(ctx context.Context, admin *domain.UserAdmin, psiID uuid.UUID, req request_structs.CreateDeontologiaRequest) error {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return domain.ErrInsufficientPerms
	}

	// 2. INTEGRIDAD REFERENCIAL
	// Asegura que el psicólogo exista (si no, ErrPsiNotFound) antes de insertar.
	if _, err := s.repo.GetByID(ctx, psiID); err != nil {
		return domain.ErrPsiNotFound
	}

	// 3. NORMALIZACIÓN Y SANITIZACIÓN
	// El contenido es texto plano: se recorta y se eliminan TODAS las etiquetas
	// HTML con StrictPolicy (la barrera final XSS es la API, ver AGENTS.md).
	content := strings.TrimSpace(bluemonday.StrictPolicy().Sanitize(req.Content))
	if content == "" {
		return domain.ErrInvalidRequest
	}
	if len(content) > 10000 {
		return domain.ErrInvalidRequest
	}

	// 4. PERSISTENCIA CON AUDITORÍA DEL EJECUTOR
	entry := &domain.PsiODeontologia{
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

	return s.repo.CreateDeontologia(ctx, entry)
}

// ListDeontologiaByPsiID recupera el expediente deontológico completo de un psicólogo.
func (s *PsiService) ListDeontologiaByPsiID(ctx context.Context, admin *domain.UserAdmin, psiID uuid.UUID) ([]domain.PsiODeontologia, error) {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return nil, domain.ErrInsufficientPerms
	}

	// 2. INTEGRIDAD REFERENCIAL
	if _, err := s.repo.GetByID(ctx, psiID); err != nil {
		return nil, domain.ErrPsiNotFound
	}

	return s.repo.ListDeontologiaByPsiID(ctx, psiID)
}

// DeleteDeontologiaByAdmin elimina lógicamente una entrada deontológica.
func (s *PsiService) DeleteDeontologiaByAdmin(ctx context.Context, admin *domain.UserAdmin, entryID uuid.UUID) error {
	// 1. VALIDACIÓN DE PERMISOS (Gatekeeping)
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
		return domain.ErrInsufficientPerms
	}

	// 2. EXISTENCIA (404 si no está)
	if _, err := s.repo.GetDeontologiaByID(ctx, entryID); err != nil {
		return domain.ErrDeontologiaNotFound
	}

	return s.repo.DeleteDeontologia(ctx, entryID)
}

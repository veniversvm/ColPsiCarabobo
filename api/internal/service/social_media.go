package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

const MaxSocialNetworks = 10

// AddSocialNetwork crea una red social ligada al psicólogo con límite de seguridad
func (s *PsiService) AddSocialNetwork(ctx context.Context, psi *domain.PsiUserModel, req request_structs.CreateSocialNetworkRequest) error {
	// 1. Validar límite de seguridad
	currentCount, err := s.repo.CountSocialNetworksByPsiID(ctx, psi.ID)
	if err != nil {
		return fmt.Errorf("error al verificar límite de redes sociales: %w", err)
	}

	if currentCount >= MaxSocialNetworks {
		return fmt.Errorf("límite de redes sociales alcanzado (%d)", MaxSocialNetworks)
	}

	// 2. Preparar el modelo
	network := &domain.PsiUserSocialNetwork{
		ID: uuid.New(), // Generamos un nuevo UUID para la red social
		AuditModel: domain.AuditModel{
			CreateBy:   psi.Username,
			CreateById: &psi.ID,
			UpdateBy:   psi.Username,
			UpdateById: &psi.ID,
		},
		PsiUserID: psi.ID,
		// Usamos la normalización que perfeccionamos antes
		Name:     utils.NormalizePlatformName(req.Name),
		URL:      strings.TrimSpace(req.URL),
		IsActive: true,
	}

	// 3. Persistir
	return s.repo.CreateSocialNetwork(ctx, network)
}

// UpdateSocialNetwork permite al psicólogo editar su propia red
func (s *PsiService) UpdateSocialNetwork(ctx context.Context, psi *domain.PsiUserModel, netID uuid.UUID, req request_structs.UpdateSocialNetworkRequest) error {
	network, err := s.repo.GetSocialNetworkByID(ctx, netID)
	if err != nil {
		return errors.New("red social no encontrada")
	}

	// Seguridad: Ownership
	if network.PsiUserID != psi.ID {
		return errors.New("no tienes permiso para editar esta red social")
	}

	network.UpdateBy = psi.Username
	network.UpdateById = &psi.ID

	if req.Name != nil {
		network.Name = utils.NormalizePlatformName(*req.Name)
	} // Normalización en update
	if req.URL != nil {
		network.URL = strings.TrimSpace(*req.URL)
	}
	if req.IsActive != nil {
		network.IsActive = *req.IsActive
	}

	return s.repo.UpdateSocialNetwork(ctx, network)
}

// DeleteSocialNetwork borrado lógico (usable por Admin o por el propio Psicólogo)
func (s *PsiService) DeleteSocialNetwork(ctx context.Context, executorRole string, executorID uuid.UUID, netID uuid.UUID) error {
	network, err := s.repo.GetSocialNetworkByID(ctx, netID)
	if err != nil {
		return errors.New("red social no encontrada")
	}

	// Validar jerarquía y propiedad
	if executorRole == "psi" {
		if network.PsiUserID != executorID {
			return errors.New("no puedes borrar una red social que no te pertenece")
		}
	} else if executorRole == "admin" {
		// Si quieres validar que el admin tenga permiso 'CanUpdatePsi', deberías pasar el modelo Admin y chequearlo.
		// Asumiremos que si llegó aquí por el router de admin, ya pasó la primera barrera.
	} else {
		return errors.New("rol no autorizado")
	}

	return s.repo.DeleteSocialNetwork(ctx, netID)
}

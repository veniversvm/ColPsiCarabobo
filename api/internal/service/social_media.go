// api/internal/service/social_media.go
// Package service implementa la lógica de negocio central de la aplicación.
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

// MaxSocialNetworks define el límite máximo de perfiles sociales por psicólogo.
// Previene el abuso de almacenamiento y mantiene la estética del perfil público.
const MaxSocialNetworks = 10

// =========================================================================
// GESTIÓN DE REDES SOCIALES
// =========================================================================

// AddSocialNetwork vincula una nueva red social al perfil del psicólogo.
// Implementa validación de cuotas y normalización automática de plataformas.
func (s *PsiService) AddSocialNetwork(ctx context.Context, psi *domain.PsiUserModel, req request_structs.CreateSocialNetworkRequest) error {

	// 1. CONTROL DE CUOTA: Evita que un perfil se sature de registros innecesarios.
	currentCount, err := s.repo.CountSocialNetworksByPsiID(ctx, psi.ID)
	if err != nil {
		return fmt.Errorf("error al verificar límite de redes sociales: %w", err)
	}

	if currentCount >= MaxSocialNetworks {
		return fmt.Errorf("límite de redes sociales alcanzado (%d)", MaxSocialNetworks)
	}

	// 2. PREPARACIÓN DEL MODELO:
	// Se integra la lógica de auditoría interna y la utilidad de normalización.
	network := &domain.PsiUserSocialNetwork{
		ID: uuid.Must(uuid.NewV7()),
		AuditModel: domain.AuditModel{
			CreateBy:   psi.Username,
			CreateById: &psi.ID,
			UpdateBy:   psi.Username,
			UpdateById: &psi.ID,
		},
		PsiUserID: psi.ID,
		// NORMALIZACIÓN: Transforma "ig" -> "Instagram" o "fb" -> "Facebook".
		Name:     utils.NormalizePlatformName(req.Name),
		URL:      strings.TrimSpace(req.URL),
		IsActive: true,
	}

	// 3. PERSISTENCIA
	return s.repo.CreateSocialNetwork(ctx, network)
}

// UpdateSocialNetwork permite la edición de una red social validando la propiedad.
// Implementa el principio de "Zero Trust" al verificar que el ID del psicólogo
// coincida con el dueño del recurso (Ownership).
func (s *PsiService) UpdateSocialNetwork(ctx context.Context, psi *domain.PsiUserModel, netID uuid.UUID, req request_structs.UpdateSocialNetworkRequest) error {
	network, err := s.repo.GetSocialNetworkByID(ctx, netID)
	if err != nil {
		return errors.New("red social no encontrada")
	}

	// SEGURIDAD: Previene que un psicólogo edite las redes sociales de otro (ID Spoofing).
	if network.PsiUserID != psi.ID {
		return errors.New("no tienes permiso para editar esta red social")
	}

	// Actualización de auditoría
	network.UpdateBy = psi.Username
	network.UpdateById = &psi.ID

	// Aplicación de cambios parciales (PATCH)
	if req.Name != nil {
		network.Name = utils.NormalizePlatformName(*req.Name)
	}
	if req.URL != nil {
		network.URL = strings.TrimSpace(*req.URL)
	}
	if req.IsActive != nil {
		network.IsActive = *req.IsActive
	}

	return s.repo.UpdateSocialNetwork(ctx, network)
}

// DeleteSocialNetwork gestiona el borrado lógico de redes sociales.
// Es un método polimórfico que valida el acceso basado tanto en la
// propiedad (Psicólogo) como en la autoridad (Administrador).
func (s *PsiService) DeleteSocialNetwork(ctx context.Context, executorRole string, executorID uuid.UUID, netID uuid.UUID) error {
	network, err := s.repo.GetSocialNetworkByID(ctx, netID)
	if err != nil {
		return errors.New("red social no encontrada")
	}

	// VALIDACIÓN DE JERARQUÍA Y PERMISOS:
	if executorRole == "psi" {
		// El psicólogo solo puede borrar sus propios registros.
		if network.PsiUserID != executorID {
			return errors.New("no puedes borrar una red social que no te pertenece")
		}
	} else if executorRole == "admin" {
		// El administrador tiene autoridad sobre el registro (sujeto a validación de middleware).
	} else {
		return errors.New("rol no autorizado")
	}

	return s.repo.DeleteSocialNetwork(ctx, netID)
}

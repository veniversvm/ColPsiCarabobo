// api/internal/service/social_media.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona el submódulo de Presencia Digital (Redes Sociales).
// Se encarga de aislar la validación de URLs, la estandarización visual de las
// plataformas y la protección de los registros contra modificaciones no autorizadas.
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

// MaxSocialNetworks define el límite estricto de perfiles sociales por psicólogo.
//
// Mitigación de Agotamiento de Recursos (Resource Exhaustion):
// Previene que un usuario malicioso o un script automatizado sature la base de datos
// creando infinitos registros relacionados a su perfil (Denegación de Servicio a nivel de BD).
// Secundariamente, asegura que la interfaz de usuario (UI) se mantenga estéticamente limpia.
const MaxSocialNetworks = 10

// =========================================================================
// GESTIÓN DE REDES SOCIALES
// =========================================================================

// AddSocialNetwork vincula una nueva red social al perfil del psicólogo.
// Implementa validación de cuotas (Quotas) y estandarización de datos (Data Normalization).
func (s *PsiService) AddSocialNetwork(ctx context.Context, psi *domain.PsiUserModel, req request_structs.CreateSocialNetworkRequest) error {

	// 1. CONTROL DE CUOTA (Defensive Programming)
	// Evalúa cuántos registros tiene actualmente el usuario antes de permitir una nueva inserción.
	currentCount, err := s.repo.CountSocialNetworksByPsiID(ctx, psi.ID)
	if err != nil {
		return fmt.Errorf("error al verificar límite de redes sociales: %w", err)
	}

	if currentCount >= MaxSocialNetworks {
		return fmt.Errorf("límite de redes sociales alcanzado (%d)", MaxSocialNetworks)
	}

	// 2. PREPARACIÓN DEL MODELO E INMUTABILIDAD
	// Se inyecta la lógica de auditoría usando los datos del creador.
	network := &domain.PsiUserSocialNetwork{
		ID: uuid.Must(uuid.NewV7()),
		AuditModel: domain.AuditModel{
			CreateBy:   psi.Username,
			CreateById: &psi.ID,
			UpdateBy:   psi.Username,
			UpdateById: &psi.ID,
		},
		PsiUserID: psi.ID,

		// NORMALIZACIÓN CENTRALIZADA:
		// Sin importar cómo lo envíe el frontend ("ig", "INSTA", "facebook"),
		// el servicio lo convierte a un estándar (ej: "Instagram") para garantizar
		// que los íconos rendericen correctamente en el cliente.
		Name:     utils.NormalizePlatformName(req.Name),
		URL:      strings.TrimSpace(req.URL),
		IsActive: true,
	}

	// 3. PERSISTENCIA
	return s.repo.CreateSocialNetwork(ctx, network)
}

// UpdateSocialNetwork permite la edición parcial (PATCH) de una red social.
//
// Prevención de Vulnerabilidad IDOR (Insecure Direct Object Reference):
// Valida explícitamente mediante el principio "Zero Trust" (Confianza Cero) que
// el UUID de la red social que se intenta modificar realmente pertenezca al usuario
// que está realizando la petición HTTP.
func (s *PsiService) UpdateSocialNetwork(ctx context.Context, psi *domain.PsiUserModel, netID uuid.UUID, req request_structs.UpdateSocialNetworkRequest) error {
	network, err := s.repo.GetSocialNetworkByID(ctx, netID)
	if err != nil {
		return errors.New("red social no encontrada")
	}

	// SEGURIDAD (Ownership Check):
	// Bloquea el intento si un psicólogo inyecta en la URL el UUID de una red social
	// perteneciente a otro colega (ID Spoofing).
	if network.PsiUserID != psi.ID {
		return domain.ErrSocialPermDenied
	}

	// Actualización de auditoría (Rastro Forense)
	network.UpdateBy = psi.Username
	network.UpdateById = &psi.ID

	// Aplicación de cambios parciales (Evaluación de Punteros)
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

// DeleteSocialNetwork gestiona la destrucción lógica o física de una red social.
//
// Control de Acceso Polimórfico:
// Es un método diseñado para ser invocado tanto por el propio psicólogo desde
// su panel de autogestión, como por un administrador desde el panel de moderación.
// Enruta las reglas de validación basándose en el rol del 'Ejecutor'.
func (s *PsiService) DeleteSocialNetwork(ctx context.Context, executorRole string, executorID uuid.UUID, netID uuid.UUID) error {
	network, err := s.repo.GetSocialNetworkByID(ctx, netID)
	if err != nil {
		return errors.New("red social no encontrada")
	}

	// VALIDACIÓN DE JERARQUÍA Y PERMISOS (RBAC Dinámico):
	if executorRole == "psi" {
		// Modo Autogestión: El psicólogo solo puede borrar sus propios registros (IDOR Prevention).
		if network.PsiUserID != executorID {
			return domain.ErrSocialOwnDenied
		}
	} else if executorRole == "admin" {
		// Modo Moderación: El administrador tiene autoridad global sobre el registro.
		// (La validación de si el admin tiene permisos para moderar perfiles ya fue
		// garantizada previamente en la capa de Middleware).
	} else {
		return errors.New("rol no autorizado")
	}

	return s.repo.DeleteSocialNetwork(ctx, netID)
}

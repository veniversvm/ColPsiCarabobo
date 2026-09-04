// api/internal/domain/inscription_repository.go
package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// InscriptionRepository define el contrato de persistencia para las solicitudes
// de pre-inscripción de profesionales (psi_inscription_requests).
type InscriptionRepository interface {
	// Create inserta una nueva solicitud de pre-inscripción.
	Create(ctx context.Context, req *PsiInscriptionRequest) error

	// GetByID recupera una solicitud por su UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*PsiInscriptionRequest, error)

	// Search lista solicitudes con filtros (status, búsqueda) y paginación.
	Search(ctx context.Context, filter request_structs.InscriptionListFilter) ([]PsiInscriptionRequest, int64, error)

	// ExistsPendingCI retorna true si existe una solicitud pendiente con esa cédula.
	ExistsPendingCI(ctx context.Context, ci int) (bool, error)

	// ExistsPendingFPV retorna true si existe una solicitud pendiente con ese FPV.
	ExistsPendingFPV(ctx context.Context, fpv int) (bool, error)

	// CIExistsDevuelve si la cédula ya está registrada en psi_users.
	CIInPsiUsers(ctx context.Context, ci int) (bool, error)

	// FPVInPsiUsers retorna si el FPV ya está registrado en psi_users.
	FPVInPsiUsers(ctx context.Context, fpv int) (bool, error)

	// NextControlNumber calcula el siguiente número de control secuencial
	// basado en el MAX(control_number numérico) de psi_users + 1.
	NextControlNumber(ctx context.Context) (int, error)

	// UpdateChangedStatus actualiza el estado de una solicitud (approve/reject)
	// junto con su número de control si corresponde.
	Update(ctx context.Context, req *PsiInscriptionRequest) error

	// Delete elimina físicamente una solicitud.
	Delete(ctx context.Context, id uuid.UUID) error
}

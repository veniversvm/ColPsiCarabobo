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

	// ExistsPendingEmail retorna true si existe una solicitud pendiente con ese correo.
	ExistsPendingEmail(ctx context.Context, email string) (bool, error)

	// ExistsPendingCIExcluding es como ExistsPendingCI pero ignora la solicitud
	// indicada (usado al editar la ficha por admin).
	ExistsPendingCIExcluding(ctx context.Context, ci int, excludeID uuid.UUID) (bool, error)

	// ExistsPendingFPVExcluding es como ExistsPendingFPV pero ignora la solicitud
	// indicada (usado al editar la ficha por admin).
	ExistsPendingFPVExcluding(ctx context.Context, fpv int, excludeID uuid.UUID) (bool, error)

	// ExistsPendingEmailExcluding es como ExistsPendingEmail pero ignora la
	// solicitud indicada (usado al editar la ficha por admin).
	ExistsPendingEmailExcluding(ctx context.Context, email string, excludeID uuid.UUID) (bool, error)

	// CIExistsDevuelve si la cédula ya está registrada en psi_users.
	CIInPsiUsers(ctx context.Context, ci int) (bool, error)

	// FPVInPsiUsers retorna si el FPV ya está registrado en psi_users.
	FPVInPsiUsers(ctx context.Context, fpv int) (bool, error)

	// EmailInPsiUsers retorna si el correo ya está registrado en psi_users.
	EmailInPsiUsers(ctx context.Context, email string) (bool, error)

	// NextControlNumber calcula el siguiente número de control secuencial
	// basado en el MAX(control_number numérico) de psi_users + 1.
	NextControlNumber(ctx context.Context) (int, error)

	// UpdateChangedStatus actualiza el estado de una solicitud (approve/reject)
	// junto con su número de control si corresponde.
	Update(ctx context.Context, req *PsiInscriptionRequest) error

	// UpdateNotes actualiza solo las notas administrativas de una solicitud.
	UpdateNotes(ctx context.Context, id uuid.UUID, notes string) error

	// Delete elimina físicamente una solicitud.
	Delete(ctx context.Context, id uuid.UUID) error

	// ── Documentos de la ficha ─────────────────────────────────────────────

	// CreateDocuments persiste las fotos de documentos de la solicitud.
	CreateDocuments(ctx context.Context, docs []PsiInscriptionDocument) error

	// ListDocumentsByRequestID recupera las fotos de documentos de la solicitud.
	ListDocumentsByRequestID(ctx context.Context, requestID uuid.UUID) ([]PsiInscriptionDocument, error)

	// GetInscriptionDocumentByID busca una foto de documento por su UUID.
	GetInscriptionDocumentByID(ctx context.Context, id uuid.UUID) (*PsiInscriptionDocument, error)

	// UpdateInscriptionDocument actualiza una foto de documento de la ficha.
	UpdateInscriptionDocument(ctx context.Context, doc *PsiInscriptionDocument) error

	// DeleteInscriptionDocument elimina físicamente la foto de un documento.
	DeleteInscriptionDocument(ctx context.Context, id uuid.UUID) error

	// DeleteInscriptionDocumentsByRequestID elimina las fotos de documentos
	// de una solicitud (al aprobar migran a psi_user_documents o se rechaza).
	DeleteInscriptionDocumentsByRequestID(ctx context.Context, requestID uuid.UUID) error
}

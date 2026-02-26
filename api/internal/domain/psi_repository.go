// Package domain define las entidades y los contratos de persistencia del sistema.
package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// PsiUserRepository define el contrato de abstracción para la gestión integral de psicólogos.
// Implementa operaciones complejas que abarcan desde el registro administrativo hasta
// la autogestión de perfiles, postgrados y redes sociales.
type PsiUserRepository interface {
	// =========================================================================
	// GESTIÓN DE PERFIL Y DATOS CORE
	// =========================================================================

	// CreateWithColData realiza una inserción atómica (transaccional) de un nuevo psicólogo
	// junto a sus datos colegiales iniciales. Garantiza que no existan registros huérfanos.
	CreateWithColData(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData) error

	// GetByID recupera un psicólogo mediante su UUID, realizando un Eager Loading
	// de sus relaciones: ColData, PostGrades y SocialNetworks.
	GetByID(ctx context.Context, id uuid.UUID) (*PsiUserModel, error)

	// GetByIdentifier busca un psicólogo por su identificador único (Username o Email).
	// Es el método principal para el proceso de autenticación/login.
	GetByIdentifier(ctx context.Context, identifier string) (*PsiUserModel, error)

	// GetPsiUserColData recupera exclusivamente la información colegial de un psicólogo.
	GetPsiUserColData(ctx context.Context, psiID uuid.UUID) (*PsiUserColData, error)

	// =========================================================================
	// ACTUALIZACIONES (MUTACIONES)
	// =========================================================================

	// Update guarda cambios en el modelo principal y/o los datos colegiales asociados.
	// Se utiliza principalmente en operaciones de edición administrativa total.
	Update(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData) error

	// UpdatePublicProfile actualiza la información que el psicólogo gestiona de sí mismo,
	// como datos de contacto, biografía y preferencias de visibilidad.
	UpdatePublicProfile(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData) error

	// UpdateKey actualiza la semilla de firma (Key) del usuario.
	// Vital para la rotación de sesiones y la invalidación de tokens JWT.
	UpdateKey(ctx context.Context, psi *PsiUserModel) error

	// Delete realiza una eliminación lógica (Soft Delete) del perfil del psicólogo.
	Delete(ctx context.Context, id uuid.UUID) error

	// =========================================================================
	// BÚSQUEDA Y DIRECTORIO
	// =========================================================================

	// SearchDirectory implementa la lógica de búsqueda para el directorio público.
	// Filtra por solvencia a menos que se realice una búsqueda explícita por identidad.
	SearchDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]PsiUserModel, int64, error)

	// Search realiza búsquedas genéricas con filtros dinámicos (map) y paginación.
	Search(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]PsiUserModel, int64, error)

	// SearchAdmin proporciona una vista de "Rayos X" del listado de psicólogos para administradores,
	// ignorando filtros de visibilidad pública o estatus de solvencia.
	SearchAdmin(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]PsiUserModel, int64, error)

	// =========================================================================
	// GESTIÓN ACADÉMICA (POSTGRADOS)
	// =========================================================================

	// CreatePostGrade registra un nuevo título académico asociado al psicólogo.
	CreatePostGrade(ctx context.Context, pg *PsiUserPostGrade) error

	// GetPostGradeByID recupera un registro académico específico por su ID.
	GetPostGradeByID(ctx context.Context, id uuid.UUID) (*PsiUserPostGrade, error)

	// UpdatePostGrade modifica los datos de un título existente (Título, Universidad, Soportes).
	UpdatePostGrade(ctx context.Context, pg *PsiUserPostGrade) error

	// =========================================================================
	// PRESENCIA DIGITAL (REDES SOCIALES)
	// =========================================================================

	// CreateSocialNetwork añade un nuevo enlace de red social al perfil.
	CreateSocialNetwork(ctx context.Context, sn *PsiUserSocialNetwork) error

	// GetSocialNetworkByID busca una red social específica por su UUID.
	GetSocialNetworkByID(ctx context.Context, id uuid.UUID) (*PsiUserSocialNetwork, error)

	// UpdateSocialNetwork modifica una red social existente.
	UpdateSocialNetwork(ctx context.Context, sn *PsiUserSocialNetwork) error

	// DeleteSocialNetwork elimina lógicamente una red social del perfil.
	DeleteSocialNetwork(ctx context.Context, id uuid.UUID) error

	// CountSocialNetworksByPsiID retorna la cantidad de redes registradas por el usuario.
	CountSocialNetworksByPsiID(ctx context.Context, psiID uuid.UUID) (int64, error)
}

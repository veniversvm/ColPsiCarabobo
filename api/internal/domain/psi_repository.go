// api/internal/domain/psi_repository.go
// Package domain define las entidades y los contratos de persistencia del sistema.
package domain

import (
	"context"
	"time"

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
	// `solvencies` permite sembrar el historial anual de solvencias (slice vacío si no aplica).
	CreateWithColData(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData, solvencies []PsiUserSolvency, postgrades []PsiUserPostGrade) error

	// GetByID recupera un psicólogo mediante su UUID, realizando un Eager Loading
	// de sus relaciones: ColData, PostGrades y SocialNetworks.
	GetByID(ctx context.Context, id uuid.UUID) (*PsiUserModel, error)

	// GetByFPV busca un psicólogo por su número de FPV, que es el identificador público
	// utilizado en las URLs del directorio. También realiza un Preload de relaciones.
	GetByFPV(ctx context.Context, id int) (PsiUserModel, error)

	// GetByIdentifier busca un psicólogo por su identificador único (Username o Email).
	// Es el método principal para el proceso de autenticación/login.
	GetByIdentifier(ctx context.Context, identifier string) (*PsiUserModel, error)

	// GetPsiUserColData recupera exclusivamente la información colegial de un psicólogo.
	GetPsiUserColData(ctx context.Context, psiID uuid.UUID) (*PsiUserColData, error)

	// Biografia full
	GetTextContentByID(ctx context.Context, id uuid.UUID) (string, error)

	// Validacion de username o email
	ValidateUniqueCredentials(ctx context.Context, username, email string, excludeID uuid.UUID) error

	// Uso exlusivo para indexacion de Google SEO
	GetSitemapData(ctx context.Context) ([]PsiUserModel, error)

	// GetAllForABSSync recupera (incluyendo soft-deleted) los datos mínimos
	// necesarios para reconciliar las cuentas de Audiobookshelf: CI, solvencia,
	// estado activo e id de la cuenta ABS.
	GetAllForABSSync(ctx context.Context) ([]PsiUserModel, error)

	// =========================================================================
	// ACTUALIZACIONES (MUTACIONES)
	// =========================================================================

	// Update guarda cambios en el modelo principal y/o los datos colegiales asociados.
	// Se utiliza principalmente en operaciones de edición administrativa total.
	Update(
		ctx context.Context,
		psi *PsiUserModel,
		colData *PsiUserColData,
		bioText *TextModel,
		solvencies []PsiUserSolvency,
	) error

	// UpdatePublicProfile actualiza la información que el psicólogo gestiona de sí mismo,
	// como datos de contacto, biografía y preferencias de visibilidad.
	UpdatePublicProfile(ctx context.Context, psi *PsiUserModel, colData *PsiUserColData, bioText *TextModel) error

	// UpdateKey actualiza la semilla de firma (Key) del usuario.
	// Vital para la rotación de sesiones y la invalidación de tokens JWT.
	UpdateKey(ctx context.Context, psi *PsiUserModel) error

	// ResetPassword actualiza únicamente la clave temporal del psicólogo:
	// hash bcrypt, rotación de la Key de sesión y fuerza de must_change_password.
	// Usado por el admin al reiniciar la contraseña de una cuenta.
	ResetPassword(ctx context.Context, psi *PsiUserModel) error

	// UpdateAudioBookShellID persiste el id de la cuenta del agremiado en
	// Audiobookshelf (campo AudioBookShellId del modelo).
	UpdateAudioBookShellID(ctx context.Context, psi *PsiUserModel) error

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
	// GESTION DE SOLVENCIAS
	// =========================================================================

	CreateSolvency(ctx context.Context, pg *PsiUserSolvency) error

	GetSolvencies(ctx context.Context, id uuid.UUID) ([]PsiUserSolvency, error)

	CreateOrUpdateSolvencies(ctx context.Context, solvencies []PsiUserSolvency) error

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

	// =========================================================================
	// EXPEDIENTE DEONTOLÓGICO
	// =========================================================================

	// CreateDeontologia registra una nueva entrada del expediente deontológico.
	// Acceso exclusivo del personal administrativo autorizado (el psicólogo NUNCA
	// puede ver ni gestionar su propio expediente disciplinario).
	CreateDeontologia(ctx context.Context, entry *PsiODeontologia) error

	// ListDeontologiaByPsiID recupera todas las entradas deontológicas de un
	// psicólogo, ordenadas de la más reciente a la más antigua.
	ListDeontologiaByPsiID(ctx context.Context, psiID uuid.UUID) ([]PsiODeontologia, error)

	// GetDeontologiaByID busca una entrada deontológica específica por su UUID.
	GetDeontologiaByID(ctx context.Context, id uuid.UUID) (*PsiODeontologia, error)

	// UpdateDeontologia actualiza el contenido (y la auditoría de edición) de una
	// entrada deontológica existente. Las entradas del expediente no se eliminan:
	// solo pueden corregirse.
	UpdateDeontologia(ctx context.Context, id uuid.UUID, content, updateBy string, updateById uuid.UUID) error

	// =========================================================================
	// OBSERVACIONES INTERNAS
	// =========================================================================

	// CreateObservations registra una nueva nota interna sobre un psicólogo.
	// Acceso exclusivo del personal administrativo (el psicólogo NUNCA la ve).
	CreateObservations(ctx context.Context, entry *PsiObservations) error

	// ListObservationsByPsiID recupera las notas internas de un psicólogo.
	ListObservationsByPsiID(ctx context.Context, psiID uuid.UUID) ([]PsiObservations, error)

	// GetObservationsByID busca una nota interna específica por su UUID.
	GetObservationsByID(ctx context.Context, id uuid.UUID) (*PsiObservations, error)

	// UpdateObservations actualiza el contenido (y la auditoría) de una nota interna.
	UpdateObservations(ctx context.Context, id uuid.UUID, content, updateBy string, updateById uuid.UUID) error

	// =========================================================================
	// CUMPLEAÑOS (AVISO AL ADMIN)
	// =========================================================================

	// GetBirthdays recupera los agremiados que cumplen años dentro del rango de fechas
	// [from, to] y que han autorizado el aviso de cumpleaños (birthday_notification).
	// La comparación se hace por mes/día (sin importar el año) para cubrir el cruce de año.
	GetBirthdays(ctx context.Context, from, to time.Time) ([]PsiUserModel, error)

	// =========================================================================
	// REGISTRO DIGITAL DE DOCUMENTOS
	// =========================================================================

	// ListDocuments recupera los documentos digitales del expediente de un
	// psicólogo, ordenados del más reciente al más antiguo.
	ListDocuments(ctx context.Context, psiID uuid.UUID) ([]PsiUserDocument, error)

	// GetDocument busca un documento digital específico por su UUID.
	GetDocument(ctx context.Context, id uuid.UUID) (*PsiUserDocument, error)

	// CreateDocument persiste un nuevo documento digital del expediente.
	CreateDocument(ctx context.Context, doc *PsiUserDocument) error

	// UpdateDocument guarda los cambios de metadatos (y auditoría) de un documento.
	UpdateDocument(ctx context.Context, doc *PsiUserDocument) error

	// DeleteDocument elimina lógicamente (soft delete) un documento del expediente.
	DeleteDocument(ctx context.Context, id uuid.UUID) error
}

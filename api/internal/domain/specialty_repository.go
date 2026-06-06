// api/internal/domain/specialty_repository.go
// Package domain define las reglas de negocio y los contratos de abstracción de datos.
package domain

import (
	"context"
)

// SpecialtyRepository define el contrato de abstracción para la gestión de especialidades.
// Sigue el principio de Dependency Inversion (DIP) de Clean Architecture, permitiendo que
// la lógica de negocio dependa de una interfaz y no de una implementación específica.
type SpecialtyRepository interface {
	// Create persiste una nueva especialidad en el almacenamiento de datos.
	// Recibe un contexto para gestionar cancelaciones y un puntero al modelo a crear.
	Create(ctx context.Context, s *PsiSpecialtyModel) error

	// GetAll recupera una lista de especialidades basada en un filtro de estado textual.
	// El parámetro 'status' permite segmentar resultados (ej. "active", "inactive", "all").
	GetAll(ctx context.Context, status string) ([]PsiSpecialtyModel, error)

	// GetByID busca una especialidad por su identificador único numérico.
	// Retorna un error de tipo 'not found' si el registro no existe o está inactivo.
	GetByID(ctx context.Context, id uint32, active bool) (*PsiSpecialtyModel, error)

	// mismo que el anterior pero tra todos al admin
	GetByAdminID(ctx context.Context, id uint32) (*PsiSpecialtyModel, error)

	// Update guarda los cambios realizados en una instancia de especialidad existente.
	Update(ctx context.Context, s *PsiSpecialtyModel) error

	// Delete realiza una desactivación lógica (soft-delete) del registro identificado por 'id'.
	// No elimina la fila físicamente para mantener la integridad referencial con los perfiles.
	Delete(ctx context.Context, id uint32) error

	// Count retorna el número de registros que coinciden con el estado de actividad proporcionado.
	// El uso de *bool (puntero) permite lógica tri-estatal:
	// - true: cuenta activas.
	// - false: cuenta inactivas.
	// - nil: cuenta el total sin filtrar.
	Count(ctx context.Context, actives *bool) (int64, error)

	// GetAllAdmin recupera el catálogo completo sin restricciones de visibilidad.
	// Método optimizado para paneles administrativos de gestión masiva.
	GetAllAdmin(ctx context.Context) ([]PsiSpecialtyModel, error)
}

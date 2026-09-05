// Package postgres proporciona la capa de persistencia (Data Access Layer)
// para el motor de base de datos PostgreSQL utilizando el ORM GORM.
//
// Este archivo encapsula la persistencia del módulo de Proyectos (Kanban).
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// kanbanRepo es la implementación concreta de [domain.KanbanRepository].
type kanbanRepo struct {
	db *gorm.DB
}

// NewKanbanRepository actúa como constructor (Factory) del repositorio de Kanban.
func NewKanbanRepository(db *gorm.DB) domain.KanbanRepository {
	return &kanbanRepo{db: db}
}

// =========================================================================
// PROYECTOS
// =========================================================================

func (r *kanbanRepo) CreateProject(ctx context.Context, p *domain.KanbanProject) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *kanbanRepo) GetProjectByID(ctx context.Context, id uuid.UUID) (*domain.KanbanProject, error) {
	var p domain.KanbanProject
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *kanbanRepo) ListProjects(ctx context.Context, adminID uuid.UUID, isMaster bool) ([]domain.KanbanProject, error) {
	var projects []domain.KanbanProject
	query := r.db.WithContext(ctx).Model(&domain.KanbanProject{}).Preload("Owner")

	if !isMaster {
		// Proyectos donde el admin es dueño o miembro.
		query = query.Where(
			"owner_id = ? OR id IN (SELECT project_id FROM kanban_project_members WHERE user_admin_id = ?)",
			adminID, adminID,
		)
	}

	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

func (r *kanbanRepo) UpdateProject(ctx context.Context, p *domain.KanbanProject) error {
	return r.db.WithContext(ctx).Model(p).Updates(map[string]interface{}{
		"name":         p.Name,
		"description":  p.Description,
		"update_by":    p.UpdateBy,
		"update_by_id": p.UpdateById,
		"updated_at":   p.UpdatedAt,
	}).Error
}

// DeleteProject elimina físicamente el proyecto y todo su contenido en cascada.
func (r *kanbanRepo) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cardIDs []uuid.UUID
		if err := tx.Model(&domain.KanbanCard{}).Where("project_id = ?", id).Pluck("id", &cardIDs).Error; err != nil {
			return err
		}
		if len(cardIDs) > 0 {
			if err := tx.Where("card_id IN ?", cardIDs).Delete(&domain.KanbanNote{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("project_id = ?", id).Delete(&domain.KanbanCard{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&domain.KanbanColumn{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&domain.KanbanMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.KanbanProject{}, "id = ?", id).Error
	})
}

// =========================================================================
// MIEMBROS
// =========================================================================

func (r *kanbanRepo) AddMember(ctx context.Context, m *domain.KanbanMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *kanbanRepo) GetMember(ctx context.Context, projectID, userAdminID uuid.UUID) (*domain.KanbanMember, error) {
	var m domain.KanbanMember
	if err := r.db.WithContext(ctx).
		First(&m, "project_id = ? AND user_admin_id = ?", projectID, userAdminID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *kanbanRepo) GetMemberByID(ctx context.Context, memberID uuid.UUID) (*domain.KanbanMember, error) {
	var m domain.KanbanMember
	if err := r.db.WithContext(ctx).
		First(&m, "id = ?", memberID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *kanbanRepo) ListMembers(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanMember, error) {
	var members []domain.KanbanMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

func (r *kanbanRepo) UpdateMember(ctx context.Context, m *domain.KanbanMember) error {
	return r.db.WithContext(ctx).Model(m).Updates(map[string]interface{}{
		"role":         m.Role,
		"update_by":    m.UpdateBy,
		"update_by_id": m.UpdateById,
		"updated_at":   m.UpdatedAt,
	}).Error
}

func (r *kanbanRepo) RemoveMember(ctx context.Context, memberID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.KanbanMember{}, "id = ?", memberID).Error
}

// =========================================================================
// COLUMNAS
// =========================================================================

func (r *kanbanRepo) CreateColumn(ctx context.Context, col *domain.KanbanColumn) error {
	return r.db.WithContext(ctx).Create(col).Error
}

func (r *kanbanRepo) GetColumn(ctx context.Context, id uuid.UUID) (*domain.KanbanColumn, error) {
	var col domain.KanbanColumn
	if err := r.db.WithContext(ctx).
		First(&col, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &col, nil
}

func (r *kanbanRepo) GetColumns(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanColumn, error) {
	var cols []domain.KanbanColumn
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("position ASC, created_at ASC").
		Find(&cols).Error
	return cols, err
}

func (r *kanbanRepo) UpdateColumn(ctx context.Context, col *domain.KanbanColumn) error {
	return r.db.WithContext(ctx).Model(col).Updates(map[string]interface{}{
		"title":        col.Title,
		"position":     col.Position,
		"update_by":    col.UpdateBy,
		"update_by_id": col.UpdateById,
		"updated_at":   col.UpdatedAt,
	}).Error
}

// DeleteColumn elimina la columna junto con sus tarjetas y notas.
func (r *kanbanRepo) DeleteColumn(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cardIDs []uuid.UUID
		if err := tx.Model(&domain.KanbanCard{}).Where("column_id = ?", id).Pluck("id", &cardIDs).Error; err != nil {
			return err
		}
		if len(cardIDs) > 0 {
			if err := tx.Where("card_id IN ?", cardIDs).Delete(&domain.KanbanNote{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("column_id = ?", id).Delete(&domain.KanbanCard{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.KanbanColumn{}, "id = ?", id).Error
	})
}

// =========================================================================
// TARJETAS Y TABLERO
// =========================================================================

// GetBoard devuelve las columnas del proyecto con sus tarjetas y notas ordenadas.
func (r *kanbanRepo) GetBoard(ctx context.Context, projectID uuid.UUID) ([]domain.KanbanColumn, error) {
	var columns []domain.KanbanColumn
	err := r.db.WithContext(ctx).
		Preload("Cards.Notes", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Cards", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC, created_at ASC")
		}).
		Where("project_id = ?", projectID).
		Order("position ASC, created_at ASC").
		Find(&columns).Error
	return columns, err
}

func (r *kanbanRepo) CreateCard(ctx context.Context, card *domain.KanbanCard) error {
	return r.db.WithContext(ctx).Create(card).Error
}

func (r *kanbanRepo) GetCard(ctx context.Context, id uuid.UUID) (*domain.KanbanCard, error) {
	var card domain.KanbanCard
	if err := r.db.WithContext(ctx).
		First(&card, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *kanbanRepo) GetCards(ctx context.Context, columnID uuid.UUID) ([]domain.KanbanCard, error) {
	var cards []domain.KanbanCard
	err := r.db.WithContext(ctx).
		Where("column_id = ?", columnID).
		Order("position ASC, created_at ASC").
		Find(&cards).Error
	return cards, err
}

func (r *kanbanRepo) UpdateCard(ctx context.Context, card *domain.KanbanCard) error {
	return r.db.WithContext(ctx).Model(card).Updates(map[string]interface{}{
		"column_id":    card.ColumnID,
		"title":        card.Title,
		"description":  card.Description,
		"position":     card.Position,
		"update_by":    card.UpdateBy,
		"update_by_id": card.UpdateById,
		"updated_at":   card.UpdatedAt,
	}).Error
}

// DeleteCard elimina la tarjeta junto con sus notas.
func (r *kanbanRepo) DeleteCard(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("card_id = ?", id).Delete(&domain.KanbanNote{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.KanbanCard{}, "id = ?", id).Error
	})
}

// =========================================================================
// NOTAS
// =========================================================================

func (r *kanbanRepo) CreateNote(ctx context.Context, n *domain.KanbanNote) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *kanbanRepo) GetNote(ctx context.Context, id uuid.UUID) (*domain.KanbanNote, error) {
	var n domain.KanbanNote
	if err := r.db.WithContext(ctx).
		First(&n, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *kanbanRepo) DeleteNote(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.KanbanNote{}, "id = ?", id).Error
}

func (r *kanbanRepo) CountNotes(ctx context.Context, cardID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.KanbanNote{}).
		Where("card_id = ?", cardID).
		Count(&count).Error
	return count, err
}

// =========================================================================
// CONTEO AGREGADO
// =========================================================================

func (r *kanbanRepo) CountMembersByProject(ctx context.Context) (map[uuid.UUID]int64, error) {
	type row struct {
		ProjectID uuid.UUID
		Total     int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).Model(&domain.KanbanMember{}).
		Select("project_id, COUNT(*) AS total").
		Group("project_id").Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int64, len(rows))
	for _, rw := range rows {
		result[rw.ProjectID] = rw.Total
	}
	return result, nil
}

func (r *kanbanRepo) CountCardsByProject(ctx context.Context) (map[uuid.UUID]int64, error) {
	type row struct {
		ProjectID uuid.UUID
		Total     int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).Model(&domain.KanbanCard{}).
		Select("project_id, COUNT(*) AS total").
		Group("project_id").Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int64, len(rows))
	for _, rw := range rows {
		result[rw.ProjectID] = rw.Total
	}
	return result, nil
}

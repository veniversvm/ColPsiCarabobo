// api/internal/service/kanban_service.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona el módulo de Proyectos (tableros Kanban) del panel
// administrativo: creación de proyectos, delegación de permisos (viewer/editor),
// permiso master global (Sudo o CanManageProjects) y las reglas de las notas
// (máximo 10 por tarjeta, 500 caracteres cada una).
package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/microcosm-cc/bluemonday"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// Limitaciones de negocio del Kanban.
const (
	// MaxNotesPerCard es el número máximo de notas permitidas por tarjeta.
	MaxNotesPerCard = 10
	// MaxNoteLengthChars es el número máximo de caracteres por nota.
	MaxNoteLengthChars = 500
	maxProjectNameLen  = 120
	maxColumnTitleLen  = 120
	maxCardTitleLen    = 200
	maxCardDescLen     = 2000
	maxProjectDescLen  = 500
)

// defaultColumnTitles son las columnas que se crean al fundar un proyecto.
var defaultColumnTitles = []string{"Por hacer", "En progreso", "Hecho"}

// kanbanAccess representa el nivel de acceso efectivo de un admin sobre un proyecto.
type kanbanAccess int

const (
	kanbanNoAccess kanbanAccess = iota
	kanbanView
	kanbanEdit
	kanbanManage
)

// KanbanService orquesta las reglas de negocio del módulo Kanban.
type KanbanService struct {
	repo      domain.KanbanRepository
	adminRepo domain.UserAdminRepository
}

// NewKanbanService crea una instancia del servicio con sus dependencias.
func NewKanbanService(repo domain.KanbanRepository, adminRepo domain.UserAdminRepository) *KanbanService {
	return &KanbanService{repo: repo, adminRepo: adminRepo}
}

// isMaster determina si un admin tiene el permiso global "master" de proyectos.
func isProjectMaster(admin *domain.UserAdmin) bool {
	return admin != nil && (admin.Sudo || admin.CanManageProjects)
}

// isUniqueConstraintViolation detecta errores de constraint única de PostgreSQL (SQLSTATE 23505).
func isUniqueConstraintViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// sanitizeText limpia y recorta texto libre de usuario.
func sanitizeKanbanText(s string, maxRunes int) string {
	s = bluemonday.StrictPolicy().Sanitize(s)
	s = strings.TrimSpace(s)
	if maxRunes <= 0 {
		return s
	}
	// Recorta por runas para respetar caracteres multibyte (ñ, acentos, emojis).
	if r := utf8.RuneCountInString(s); r > maxRunes {
		for i, c := range s {
			if i >= maxRunes {
				return s[:i]
			}
			_ = c
		}
	}
	return s
}

// resolveProjectAccess carga el proyecto, valida que el admin pueda acceder y
// devuelve el nivel de acceso efectivo.
func (s *KanbanService) resolveProjectAccess(ctx context.Context, admin *domain.UserAdmin, projectID uuid.UUID) (*domain.KanbanProject, kanbanAccess, error) {
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, kanbanNoAccess, domain.ErrProjectNotFound
	}

	project.IsMaster = isProjectMaster(admin)

	// Master global: accede y administra cualquier proyecto.
	if project.IsMaster {
		project.MyRole = domain.MemberRoleEditor
		return project, kanbanManage, nil
	}

	// El dueño administra su propio proyecto.
	if project.OwnerID == admin.ID {
		project.MyRole = domain.MemberRoleEditor
		project.IsOwner = true
		return project, kanbanManage, nil
	}

	// Miembros invitados: viewer (solo lectura) o editor.
	member, err := s.repo.GetMember(ctx, projectID, admin.ID)
	if err != nil {
		return nil, kanbanNoAccess, domain.ErrNotProjectMember
	}

	project.MyRole = member.Role
	if member.Role == domain.MemberRoleEditor {
		return project, kanbanEdit, nil
	}
	return project, kanbanView, nil
}

// =========================================================================
// PROYECTOS
// =========================================================================

// CreateProject registra un nuevo proyecto (el admin autenticado es el dueño)
// y siembra las columnas por defecto del tablero.
func (s *KanbanService) CreateProject(ctx context.Context, admin *domain.UserAdmin, req request_structs.CreateProjectRequest) (*domain.KanbanProject, error) {
	name := sanitizeKanbanText(req.Name, maxProjectNameLen)
	if name == "" {
		return nil, errors.New("el nombre del proyecto es obligatorio")
	}
	description := sanitizeKanbanText(req.Description, maxProjectDescLen)

	project := &domain.KanbanProject{
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
		Name:        name,
		Description: description,
		OwnerID:     admin.ID,
		MyRole:      domain.MemberRoleEditor,
		IsOwner:     true,
	}

	if err := s.repo.CreateProject(ctx, project); err != nil {
		return nil, err
	}

	// Columnas por defecto del tablero.
	for i, title := range defaultColumnTitles {
		column := &domain.KanbanColumn{
			AuditModel: domain.AuditModel{
				CreateBy:   admin.Username,
				CreateById: &admin.ID,
				UpdateBy:   admin.Username,
				UpdateById: &admin.ID,
			},
			ProjectID: project.ID,
			Title:     title,
			Position:  i,
		}
		if err := s.repo.CreateColumn(ctx, column); err != nil {
			return nil, err
		}
	}

	return project, nil
}

// ListProjects devuelve los proyectos accesibles para el admin con su rol.
func (s *KanbanService) ListProjects(ctx context.Context, admin *domain.UserAdmin) ([]domain.KanbanProject, error) {
	isMaster := isProjectMaster(admin)
	projects, err := s.repo.ListProjects(ctx, admin.ID, isMaster)
	if err != nil {
		return nil, err
	}

	// Conteos agregados para las tarjetas del listado.
	memberCounts, _ := s.repo.CountMembersByProject(ctx)
	cardCounts, _ := s.repo.CountCardsByProject(ctx)

	for i := range projects {
		p := &projects[i]
		p.IsMaster = isMaster
		p.MemberCount = memberCounts[p.ID]
		p.CardCount = cardCounts[p.ID]
		if isMaster {
			p.MyRole = domain.MemberRoleEditor
			continue
		}
		if p.OwnerID == admin.ID {
			p.IsOwner = true
			p.MyRole = domain.MemberRoleEditor
			continue
		}
		if member, err := s.repo.GetMember(ctx, p.ID, admin.ID); err == nil {
			p.MyRole = member.Role
		}
	}

	return projects, nil
}

// UpdateProject muta los metadatos de un proyecto (solo dueño o master).
func (s *KanbanService) UpdateProject(ctx context.Context, admin *domain.UserAdmin, projectID uuid.UUID, req request_structs.UpdateProjectRequest) error {
	project, access, err := s.resolveProjectAccess(ctx, admin, projectID)
	if err != nil {
		return err
	}
	if access < kanbanManage {
		return domain.ErrInsufficientPerms
	}

	if req.Name != nil {
		name := sanitizeKanbanText(*req.Name, maxProjectNameLen)
		if name == "" {
			return errors.New("el nombre del proyecto es obligatorio")
		}
		project.Name = name
	}
	if req.Description != nil {
		project.Description = sanitizeKanbanText(*req.Description, maxProjectDescLen)
	}

	project.UpdateBy = admin.Username
	project.UpdateById = &admin.ID
	project.UpdatedAt = time.Now()
	return s.repo.UpdateProject(ctx, project)
}

// DeleteProject elimina un proyecto completo (solo dueño o master).
func (s *KanbanService) DeleteProject(ctx context.Context, admin *domain.UserAdmin, projectID uuid.UUID) error {
	_, access, err := s.resolveProjectAccess(ctx, admin, projectID)
	if err != nil {
		return err
	}
	if access < kanbanManage {
		return domain.ErrInsufficientPerms
	}
	return s.repo.DeleteProject(ctx, projectID)
}

// GetBoard devuelve el tablero completo (columnas con tarjetas y notas).
func (s *KanbanService) GetBoard(ctx context.Context, admin *domain.UserAdmin, projectID uuid.UUID) (*domain.KanbanProject, []domain.KanbanColumn, kanbanAccess, error) {
	project, access, err := s.resolveProjectAccess(ctx, admin, projectID)
	if err != nil {
		return nil, nil, kanbanNoAccess, err
	}

	columns, err := s.repo.GetBoard(ctx, projectID)
	if err != nil {
		return nil, nil, kanbanNoAccess, err
	}

	project.MemberCount, _ = s.countMembers(ctx, projectID)
	project.CardCount = countCards(columns)

	return project, columns, access, nil
}

func (s *KanbanService) countMembers(ctx context.Context, projectID uuid.UUID) (int64, error) {
	members, err := s.repo.ListMembers(ctx, projectID)
	if err != nil {
		return 0, err
	}
	return int64(len(members)), nil
}

func countCards(columns []domain.KanbanColumn) int64 {
	var total int64
	for i := range columns {
		total += int64(len(columns[i].Cards))
	}
	return total
}

// =========================================================================
// MIEMBROS
// =========================================================================

// ListMembers devuelve los miembros de un proyecto (acceso de lectura en adelante).
func (s *KanbanService) ListMembers(ctx context.Context, admin *domain.UserAdmin, projectID uuid.UUID) ([]domain.KanbanMember, error) {
	if _, access, err := s.resolveProjectAccess(ctx, admin, projectID); err != nil {
		return nil, err
	} else if access < kanbanView {
		return nil, domain.ErrInsufficientPerms
	}
	return s.repo.ListMembers(ctx, projectID)
}

// AddMember invita a un administrador al proyecto con un rol (viewer/editor).
func (s *KanbanService) AddMember(ctx context.Context, admin *domain.UserAdmin, projectID uuid.UUID, req request_structs.AddMemberRequest) error {
	_, access, err := s.resolveProjectAccess(ctx, admin, projectID)
	if err != nil {
		return err
	}
	if access < kanbanManage {
		return domain.ErrInsufficientPerms
	}
	role := domain.MemberRole(req.Role)
	if !role.IsValid() {
		return domain.ErrInvalidMemberRole
	}
	if req.UserAdminID == uuid.Nil || req.UserAdminID == admin.ID {
		return errors.New("selecciona un administrador válido para invitar")
	}

	// Verifica que el objetivo exista como admin.
	if _, err := s.adminRepo.GetByID(ctx, req.UserAdminID); err != nil {
		return errors.New("el administrador invitado no existe")
	}

	member := &domain.KanbanMember{
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
		ProjectID:   projectID,
		UserAdminID: req.UserAdminID,
		Role:        role,
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		if isUniqueConstraintViolation(err) {
			return domain.ErrMemberAlreadyExists
		}
		return err
	}
	return nil
}

// UpdateMember cambia el rol de un miembro del proyecto.
func (s *KanbanService) UpdateMember(ctx context.Context, admin *domain.UserAdmin, memberID uuid.UUID, req request_structs.UpdateMemberRequest) error {
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return domain.ErrNotProjectMember
	}
	if _, access, err := s.resolveProjectAccess(ctx, admin, member.ProjectID); err != nil {
		return err
	} else if access < kanbanManage {
		return domain.ErrInsufficientPerms
	}
	if req.Role == nil {
		return domain.ErrInvalidMemberRole
	}
	role := domain.MemberRole(*req.Role)
	if !role.IsValid() {
		return domain.ErrInvalidMemberRole
	}

	member.Role = role
	member.UpdateBy = admin.Username
	member.UpdateById = &admin.ID
	member.UpdatedAt = time.Now()
	return s.repo.UpdateMember(ctx, member)
}

// RemoveMember expulsa a un miembro del proyecto.
func (s *KanbanService) RemoveMember(ctx context.Context, admin *domain.UserAdmin, memberID uuid.UUID) error {
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return domain.ErrNotProjectMember
	}
	if _, access, err := s.resolveProjectAccess(ctx, admin, member.ProjectID); err != nil {
		return err
	} else if access < kanbanManage {
		return domain.ErrInsufficientPerms
	}
	return s.repo.RemoveMember(ctx, memberID)
}

// =========================================================================
// COLUMNAS
// =========================================================================

// CreateColumn añade una columna al final del tablero.
func (s *KanbanService) CreateColumn(ctx context.Context, admin *domain.UserAdmin, projectID uuid.UUID, req request_structs.CreateColumnRequest) (*domain.KanbanColumn, error) {
	_, access, err := s.resolveProjectAccess(ctx, admin, projectID)
	if err != nil {
		return nil, err
	}
	if access < kanbanManage {
		return nil, domain.ErrInsufficientPerms
	}

	title := sanitizeKanbanText(req.Title, maxColumnTitleLen)
	if title == "" {
		return nil, errors.New("el título de la columna es obligatorio")
	}

	existing, err := s.repo.GetColumns(ctx, projectID)
	if err != nil {
		return nil, err
	}

	column := &domain.KanbanColumn{
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
		ProjectID: projectID,
		Title:     title,
		Position:  len(existing),
	}

	if err := s.repo.CreateColumn(ctx, column); err != nil {
		return nil, err
	}
	return column, nil
}

// UpdateColumn renombra o reposiciona una columna.
func (s *KanbanService) UpdateColumn(ctx context.Context, admin *domain.UserAdmin, columnID uuid.UUID, req request_structs.UpdateColumnRequest) error {
	column, err := s.repo.GetColumn(ctx, columnID)
	if err != nil {
		return domain.ErrColumnNotFound
	}
	if _, access, err := s.resolveProjectAccess(ctx, admin, column.ProjectID); err != nil {
		return err
	} else if access < kanbanManage {
		return domain.ErrInsufficientPerms
	}

	if req.Title != nil {
		title := sanitizeKanbanText(*req.Title, maxColumnTitleLen)
		if title == "" {
			return errors.New("el título de la columna es obligatorio")
		}
		column.Title = title
	}
	if req.Position != nil {
		column.Position = *req.Position
	}

	column.UpdateBy = admin.Username
	column.UpdateById = &admin.ID
	column.UpdatedAt = time.Now()
	return s.repo.UpdateColumn(ctx, column)
}

// DeleteColumn elimina una columna junto con sus tarjetas y notas.
func (s *KanbanService) DeleteColumn(ctx context.Context, admin *domain.UserAdmin, columnID uuid.UUID) error {
	column, err := s.repo.GetColumn(ctx, columnID)
	if err != nil {
		return domain.ErrColumnNotFound
	}
	if _, access, err := s.resolveProjectAccess(ctx, admin, column.ProjectID); err != nil {
		return err
	} else if access < kanbanManage {
		return domain.ErrInsufficientPerms
	}
	return s.repo.DeleteColumn(ctx, columnID)
}

// =========================================================================
// TARJETAS
// =========================================================================

// CreateCard añade una tarjeta a una columna del tablero.
func (s *KanbanService) CreateCard(ctx context.Context, admin *domain.UserAdmin, projectID uuid.UUID, req request_structs.CreateCardRequest) (*domain.KanbanCard, error) {
	_, access, err := s.resolveProjectAccess(ctx, admin, projectID)
	if err != nil {
		return nil, err
	}
	if access < kanbanEdit {
		return nil, domain.ErrInsufficientPerms
	}

	// La columna debe pertenecer al proyecto.
	column, err := s.repo.GetColumn(ctx, req.ColumnID)
	if err != nil || column.ProjectID != projectID {
		return nil, domain.ErrColumnNotFound
	}

	title := sanitizeKanbanText(req.Title, maxCardTitleLen)
	if title == "" {
		return nil, errors.New("el título de la tarjeta es obligatorio")
	}
	description := sanitizeKanbanText(req.Description, maxCardDescLen)

	cards, err := s.repo.GetCards(ctx, req.ColumnID)
	if err != nil {
		return nil, err
	}

	card := &domain.KanbanCard{
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
		ProjectID:   projectID,
		ColumnID:    req.ColumnID,
		Title:       title,
		Description: description,
		Position:    len(cards),
	}

	if err := s.repo.CreateCard(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}

// UpdateCard edita y/o mueve una tarjeta entre columnas del mismo proyecto.
func (s *KanbanService) UpdateCard(ctx context.Context, admin *domain.UserAdmin, cardID uuid.UUID, req request_structs.UpdateCardRequest) error {
	card, err := s.repo.GetCard(ctx, cardID)
	if err != nil {
		return domain.ErrCardNotFound
	}
	_, access, err := s.resolveProjectAccess(ctx, admin, card.ProjectID)
	if err != nil {
		return err
	}
	if access < kanbanEdit {
		return domain.ErrInsufficientPerms
	}

	if req.Title != nil {
		title := sanitizeKanbanText(*req.Title, maxCardTitleLen)
		if title == "" {
			return errors.New("el título de la tarjeta es obligatorio")
		}
		card.Title = title
	}
	if req.Description != nil {
		card.Description = sanitizeKanbanText(*req.Description, maxCardDescLen)
	}

	movingColumn := req.ColumnID != nil && *req.ColumnID != card.ColumnID
	if movingColumn {
		// La nueva columna pertenece al mismo proyecto.
		column, err := s.repo.GetColumn(ctx, *req.ColumnID)
		if err != nil || column.ProjectID != card.ProjectID {
			return domain.ErrColumnNotFound
		}
		if req.Position == nil {
			cards, err := s.repo.GetCards(ctx, *req.ColumnID)
			if err != nil {
				return err
			}
			card.Position = len(cards)
		}
		card.ColumnID = *req.ColumnID
	}
	if req.Position != nil {
		card.Position = *req.Position
	}

	card.UpdateBy = admin.Username
	card.UpdateById = &admin.ID
	card.UpdatedAt = time.Now()
	return s.repo.UpdateCard(ctx, card)
}

// DeleteCard elimina una tarjeta y sus notas.
func (s *KanbanService) DeleteCard(ctx context.Context, admin *domain.UserAdmin, cardID uuid.UUID) error {
	card, err := s.repo.GetCard(ctx, cardID)
	if err != nil {
		return domain.ErrCardNotFound
	}
	if _, access, err := s.resolveProjectAccess(ctx, admin, card.ProjectID); err != nil {
		return err
	} else if access < kanbanEdit {
		return domain.ErrInsufficientPerms
	}
	return s.repo.DeleteCard(ctx, cardID)
}

// =========================================================================
// NOTAS
// =========================================================================

// CreateNote añade una nota a una tarjeta respetando los límites de negocio:
// máximo 10 notas por tarjeta y 500 caracteres por nota.
func (s *KanbanService) CreateNote(ctx context.Context, admin *domain.UserAdmin, cardID uuid.UUID, req request_structs.CreateNoteRequest) (*domain.KanbanNote, error) {
	card, err := s.repo.GetCard(ctx, cardID)
	if err != nil {
		return nil, domain.ErrCardNotFound
	}
	if _, access, err := s.resolveProjectAccess(ctx, admin, card.ProjectID); err != nil {
		return nil, err
	} else if access < kanbanEdit {
		return nil, domain.ErrInsufficientPerms
	}

	content := strings.TrimSpace(bluemonday.StrictPolicy().Sanitize(req.Content))
	if content == "" {
		return nil, errors.New("la nota no puede estar vacía")
	}
	if utf8.RuneCountInString(content) > MaxNoteLengthChars {
		return nil, domain.ErrNoteTooLong
	}

	// Límite duro: 10 notas por tarjeta (validado en servidor, no en el cliente).
	count, err := s.repo.CountNotes(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if count >= MaxNotesPerCard {
		return nil, domain.ErrNoteLimitReached
	}

	note := &domain.KanbanNote{
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
		CardID:  cardID,
		Content: content,
	}

	if err := s.repo.CreateNote(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

// DeleteNote elimina una nota. Solo su autor, el dueño del proyecto o un master.
func (s *KanbanService) DeleteNote(ctx context.Context, admin *domain.UserAdmin, noteID uuid.UUID) error {
	note, err := s.repo.GetNote(ctx, noteID)
	if err != nil {
		return domain.ErrNoteNotFound
	}

	card, err := s.repo.GetCard(ctx, note.CardID)
	if err != nil {
		return domain.ErrCardNotFound
	}

	_, access, err := s.resolveProjectAccess(ctx, admin, card.ProjectID)
	if err != nil {
		return err
	}
	if access < kanbanEdit {
		return domain.ErrInsufficientPerms
	}

	// Un editor solo puede borrar sus propias notas; el dueño/master cualquiera.
	isAuthor := note.CreateById != nil && *note.CreateById == admin.ID
	if access < kanbanManage && !isAuthor {
		return domain.ErrInsufficientPerms
	}

	return s.repo.DeleteNote(ctx, noteID)
}

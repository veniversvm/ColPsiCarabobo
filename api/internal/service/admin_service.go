// api/internal/service/admin_service.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	domain "github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	repo  domain.UserAdminRepository
	cache *cache.Cache
}

func NewAdminService(repo domain.UserAdminRepository) *AdminService {
	return &AdminService{
		repo:  repo,
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

////////////////////////////////////////////////////////////
///////////////////////// LOGIN ////////////////////////////
////////////////////////////////////////////////////////////

func (s *AdminService) Login(ctx context.Context, identifier, password string) (string, error) {

	admin, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	if !admin.IsActive {
		return "", errors.New("la cuenta está desactivada")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", errors.New("credenciales inválidas")
	}

	newKey := uuid.New().String()
	admin.Key = newKey

	if err := s.repo.Update(ctx, admin); err != nil {
		return "", errors.New("error al procesar inicio de sesión")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": admin.ID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"role":    "admin",
	})

	return token.SignedString([]byte(newKey))
}

////////////////////////////////////////////////////////////
//////////////////////// LISTADO ///////////////////////////
////////////////////////////////////////////////////////////

func (s *AdminService) GetAdmins(
	ctx context.Context,
	active *bool,
	search string,
	page, limit int,
) (interface{}, error) {

	cacheKey := fmt.Sprintf("admins_l:%d_p:%d_s:%s_a:%v", limit, page, search, active)

	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	admins, total, err := s.repo.List(ctx, active, search, page, limit)
	if err != nil {
		return nil, err
	}

	result := fiber.Map{
		"data":        admins,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	}

	s.cache.Set(cacheKey, result, cache.DefaultExpiration)

	return result, nil
}

func (s *AdminService) GetRepo() domain.UserAdminRepository {
	return s.repo
}

////////////////////////////////////////////////////////////
//////////////////// PERMISSION ENGINE /////////////////////
////////////////////////////////////////////////////////////

type permissionUpdate struct {
	name       string
	requested  *bool
	current    bool
	updaterHas bool
	setTarget  func(bool)
}

func buildPermissionMatrix(
	req request_structs.AdminPermissionsDTO,
	target *domain.UserAdmin,
	updater domain.UserAdmin,
) []permissionUpdate {

	return []permissionUpdate{
		{"Crear Psi", req.CanCreatePsi, target.CanCreatePsi, updater.CanCreatePsi, func(v bool) { target.CanCreatePsi = v }},
		{"Update Psi", req.CanUpdatePsi, target.CanUpdatePsi, updater.CanUpdatePsi, func(v bool) { target.CanUpdatePsi = v }},
		{"Delete Psi", req.CanDeletePsi, target.CanDeletePsi, updater.CanDeletePsi, func(v bool) { target.CanDeletePsi = v }},
		{"Crear Admin", req.CanCreateAdmin, target.CanCreateAdmin, updater.CanCreateAdmin, func(v bool) { target.CanCreateAdmin = v }},
		{"Update Admin", req.CanUpdateAdmin, target.CanUpdateAdmin, updater.CanUpdateAdmin, func(v bool) { target.CanUpdateAdmin = v }},
		{"Delete Admin", req.CanDeleteAdmin, target.CanDeleteAdmin, updater.CanDeleteAdmin, func(v bool) { target.CanDeleteAdmin = v }},
		{"Publish", req.CanPublish, target.CanPublish, updater.CanPublish, func(v bool) { target.CanPublish = v }},
		{"Update Publish", req.CanUpdatePublish, target.CanUpdatePublish, updater.CanUpdatePublish, func(v bool) { target.CanUpdatePublish = v }},
		{"Delete Publish", req.CanDeletePublish, target.CanDeletePublish, updater.CanDeletePublish, func(v bool) { target.CanDeletePublish = v }},
		{"Send Notifications", req.CanSendNotifications, target.CanSendNotifications, updater.CanSendNotifications, func(v bool) { target.CanSendNotifications = v }},
		{"Manage Notifications", req.CanManageNotifications, target.CanManageNotifications, updater.CanManageNotifications, func(v bool) { target.CanManageNotifications = v }},
		{"Read Notifications", req.CanReadNotifications, target.CanReadNotifications, updater.CanReadNotifications, func(v bool) { target.CanReadNotifications = v }},
		{"Create Tags", req.CanCreateTags, target.CanCreateTags, updater.CanCreateTags, func(v bool) { target.CanCreateTags = v }},
		{"Edit Tags", req.CanEditTags, target.CanEditTags, updater.CanEditTags, func(v bool) { target.CanEditTags = v }},
		{"Delete Tags", req.CanDeleteTags, target.CanDeleteTags, updater.CanDeleteTags, func(v bool) { target.CanDeleteTags = v }},
	}
}

////////////////////////////////////////////////////////////
/////////////////////// CREATE //////////////////////////////
////////////////////////////////////////////////////////////

func (s *AdminService) CreateAdmin(
	ctx context.Context,
	creator domain.UserAdmin,
	req request_structs.CreateAdminRequest,
) error {

	if !creator.CanCreateAdmin && !creator.Sudo {
		return errors.New("permisos insuficientes para crear administradores")
	}

	newAdmin := &domain.UserAdmin{
		AuditModel: domain.AuditModel{
			ID:         uuid.New(),
			CreateBy:   creator.Username,
			CreateById: &creator.ID,
			UpdateBy:   creator.Username,
			UpdateById: &creator.ID,
		},
		Username: req.Username,
		Email:    req.Email,
		IsActive: true,
		Key:      uuid.New().String(),
		Sudo:     false,
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error procesando seguridad de la cuenta")
	}
	newAdmin.Password = string(hashed)

	matrix := buildPermissionMatrix(req.Permissions, newAdmin, creator)

	if !creator.Sudo {
		for _, perm := range matrix {
			if perm.requested != nil && *perm.requested && !perm.updaterHas {
				return fmt.Errorf("no puedes otorgar el permiso: %s", perm.name)
			}
		}
	}

	for _, perm := range matrix {
		if perm.requested != nil {
			perm.setTarget(*perm.requested)
		}
	}

	err = s.repo.Create(ctx, newAdmin)
	if err != nil {
		if strings.Contains(err.Error(), "idx_user_admins_unique_sudo") {
			return errors.New("ya existe un usuario SUDO")
		}
		return err
	}

	s.cache.Flush()
	return nil
}

////////////////////////////////////////////////////////////
//////////////////////// UPDATE ////////////////////////////
////////////////////////////////////////////////////////////

func (s *AdminService) UpdateAdmin(
	ctx context.Context,
	updater domain.UserAdmin,
	req request_structs.UpdateAdminRequest,
) error {

	target, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return errors.New("administrador no encontrado")
	}

	if !updater.Sudo {

		if target.Sudo {
			return errors.New("no puedes editar a un Super Usuario")
		}

		matrix := buildPermissionMatrix(req.Permissions, target, updater)

		for _, perm := range matrix {
			if perm.requested != nil &&
				*perm.requested != perm.current &&
				!perm.updaterHas {
				return fmt.Errorf("no tienes rango para modificar: %s", perm.name)
			}
		}
	}

	target.UpdateBy = updater.Username
	target.UpdateById = &updater.ID

	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Email != nil {
		target.Email = *req.Email
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}

	if req.Password != nil && *req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		target.Password = string(hashed)
		target.Key = uuid.New().String()
	}

	matrix := buildPermissionMatrix(req.Permissions, target, updater)
	for _, perm := range matrix {
		if perm.requested != nil {
			perm.setTarget(*perm.requested)
		}
	}

	if err := s.repo.Update(ctx, target); err != nil {
		return err
	}

	s.cache.Flush()
	return nil
}

////////////////////////////////////////////////////////////
//////////////////////// DELETE ////////////////////////////
////////////////////////////////////////////////////////////

func (s *AdminService) DeleteAdmin(
	ctx context.Context,
	updater *domain.UserAdmin,
	targetID uuid.UUID,
) error {

	if updater.ID == targetID {
		return errors.New("no puedes eliminar tu propia cuenta")
	}

	target, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return errors.New("administrador no encontrado")
	}

	if !updater.Sudo {
		if !updater.CanDeleteAdmin {
			return errors.New("no tienes permiso para eliminar administradores")
		}
		if target.Sudo {
			return errors.New("no puedes eliminar un Super Usuario")
		}
	}

	if err := s.repo.Delete(ctx, targetID); err != nil {
		return err
	}

	s.cache.Flush()
	return nil
}

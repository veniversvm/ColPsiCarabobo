// api/internal/service/admin_service.go

// Package service implementa la capa de orquestación de lógica de negocio.
// El AdminService centraliza la gestión de identidad, autorización jerárquica
// y optimización de rendimiento para el staff administrativo.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	domain "github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// AdminService encapsula las dependencias necesarias para la administración.
// Utiliza un patrón de repositorio para la persistencia y un caché en memoria
// para reducir la latencia en operaciones de lectura masiva.
type AdminService struct {
	repo  domain.UserAdminRepository
	cache *cache.Cache
}

// NewAdminService inicializa el servicio con una política de caché de 5 minutos
// y limpieza automática de registros expirados cada 10 minutos.
func NewAdminService(repo domain.UserAdminRepository) *AdminService {
	return &AdminService{
		repo:  repo,
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

// =========================================================================
// GESTIÓN DE SESIÓN Y AUTENTICACIÓN
// =========================================================================

// Login procesa la autenticación de administradores.
// Implementa "Key Rotation Security": Cada inicio de sesión exitoso genera un nuevo
// secreto UUID en la base de datos que invalida físicamente cualquier JWT
// emitido con anterioridad para este usuario (Single Session Enforcement).
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

	// Renovación de la Key para invalidar tokens anteriores y proporcionar una capa extra de seguridad
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

// =========================================================================
// LECTURA Y RENDIMIENTO (CACHE-ASIDE PATTERN)
// =========================================================================

// GetAdmins recupera una colección paginada de administradores.
// Utiliza una clave de caché compuesta por los filtros de búsqueda para garantizar
// que los resultados en memoria sean consistentes con los criterios de filtrado.
func (s *AdminService) GetAdmins(
	ctx context.Context,
	active *bool,
	search string,
	page, limit int,
) (interface{}, error) {

	// Generación de llave de caché determinística
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

// GetRepo expone la interfaz de persistencia para uso en middlewares de autorización dinámica.
func (s *AdminService) GetRepo() domain.UserAdminRepository {
	return s.repo
}

// =========================================================================
// MOTOR DE VALIDACIÓN DE PERMISOS (MATRIX ENGINE)
// =========================================================================

// permissionUpdate define una estructura de mapeo para validaciones masivas de RBAC.
type permissionUpdate struct {
	name       string
	requested  *bool
	current    bool
	updaterHas bool
	setTarget  func(bool)
}

// buildPermissionMatrix construye un mapa declarativo de todos los flags de permisos.
// Esta técnica evita el uso de Reflection y garantiza que añadir nuevos permisos
// en el futuro solo requiera una línea de código adicional.
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

// =========================================================================
// OPERACIONES DE ESCRITURA Y CONTROL JERÁRQUICO
// =========================================================================

// CreateAdmin registra un nuevo miembro del staff administrativo.
// Aplica el Principio de Menor Privilegio: Un administrador no-Sudo tiene prohibido
// otorgar permisos que él mismo no posea explícitamente.
func (s *AdminService) CreateAdmin(
	ctx context.Context,
	creator domain.UserAdmin,
	req request_structs.CreateAdminRequest,
) error {

	// Validación de autoridad base
	if !creator.CanCreateAdmin && !creator.Sudo {
		return errors.New("permisos insuficientes para crear administradores")
	}

	// Sanitización y validación de formato
	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return errors.New("el formato del email es inválido")
	}

	// Validación de fortaleza de contraseña
	if !utils.IsStrongPassword(req.Password) {
		return errors.New("la contraseña no cumple con los estándares de seguridad")
	}

	// Construcción de la entidad con trazabilidad de auditoría inicial
	newAdmin := &domain.UserAdmin{
		AuditModel: domain.AuditModel{
			CreateBy:   creator.Username,
			CreateById: &creator.ID,
			UpdateBy:   creator.Username,
			UpdateById: &creator.ID,
		},
		Username: req.Username,
		Email:    req.Email,
		IsActive: true,
		Key:      uuid.New().String(),
		Sudo:     false, // Forzado a false; la elevación a Sudo es una operación externa a la API
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error procesando seguridad de la cuenta")
	}
	newAdmin.Password = string(hashed)

	// Procesamiento de seguridad jerárquica
	matrix := buildPermissionMatrix(req.Permissions, newAdmin, creator)

	// Regla: No puedes dar lo que no tienes (a menos que seas Sudo)
	if !creator.Sudo {
		for _, perm := range matrix {
			// Bloqueo si intenta dar un permiso (true) que el creador no tiene
			if perm.requested != nil && *perm.requested && !perm.updaterHas {
				return fmt.Errorf("no puedes otorgar el permiso: %s", perm.name)
			}
		}
	}

	// Aplicación definitiva de permisos validados
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

	s.cache.Flush() // Invalidez total para reflejar el nuevo usuario en listados
	return nil
}

////////////////////////////////////////////////////////////
//////////////////////// UPDATE ////////////////////////////
////////////////////////////////////////////////////////////

// UpdateAdmin gestiona la modificación parcial de administradores.
// Implementa lógica de protección jerárquica: Ningún administrador, excepto un Sudo,
// puede modificar datos de un usuario de rango Sudo.
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

		// Validación de escalada de privilegios durante la actualización
		matrix := buildPermissionMatrix(req.Permissions, target, updater)

		// Un administrador no puede alterar el estado de un permiso (darlo o quitarlo)
		// si él mismo no posee dicho privilegio.
		for _, perm := range matrix {
			if perm.requested != nil &&
				*perm.requested != perm.current &&
				!perm.updaterHas {
				return fmt.Errorf("no tienes rango para modificar: %s", perm.name)
			}
		}
	}

	// Trazabilidad mandataria de última modificación
	target.UpdateBy = updater.Username
	target.UpdateById = &updater.ID

	// Actualización selectiva (Campos no nulos en el DTO)
	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Email != nil {
		target.Email = *req.Email
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}

	// Si se cambia el password, se genera una nueva Key para forzar el cierre de sesión en otros dispositivos
	if req.Password != nil && *req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		target.Password = string(hashed)
		target.Key = uuid.New().String()
	}

	// Aplicación de mutaciones de permisos validadas
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

// DeleteAdmin ejecuta la baja lógica (Soft Delete) de un registro administrativo.
// Implementa protecciones críticas: impide el auto-suicidio de cuenta y protege la
// inmutabilidad del Super Usuario ante personal de staff.
func (s *AdminService) DeleteAdmin(
	ctx context.Context,
	updater *domain.UserAdmin,
	targetID uuid.UUID,
) error {

	// Prevención de bloqueo accidental del propio operador
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

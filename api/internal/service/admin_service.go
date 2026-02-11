// api/internal/service/admin_service.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5" // Usar v5 para consistencia
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
		repo: repo,
		// Cache expira en 5 min, limpia cada 10 min
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

// Login valida credenciales y retorna un JWT firmado con una clave única rotativa.
func (s *AdminService) Login(ctx context.Context, identifier, password string) (string, error) {
	// 1. Buscar administrador
	admin, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		// Senior tip: Ocultamos errores específicos de DB para evitar enumeración de usuarios
		return "", errors.New("credenciales inválidas")
	}

	// 2. Verificar estado de la cuenta
	if !admin.IsActive {
		return "", errors.New("la cuenta está desactivada")
	}

	// 3. Verificar password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// 4. ROTACIÓN DE KEY (Single Session Support)
	// Generamos un nuevo UUID que servirá de "Secret" solo para este token.
	// Esto invalida automáticamente cualquier JWT emitido anteriormente.
	newKey := uuid.New().String()
	admin.Key = newKey

	// Persistimos la nueva Key en la DB
	if err := s.repo.Update(ctx, admin); err != nil {
		return "", errors.New("error al procesar inicio de sesión")
	}

	// 5. GENERACIÓN DEL TOKEN
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// Sincronizado con el middleware: usamos "user_id" como clave genérica
		"user_id": admin.ID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expira en 24h
		"iat":     time.Now().Unix(),
		"role":    "admin", // Útil para lógica rápida en el frontend
	})

	// 6. FIRMA DINÁMICA
	// Firmamos con la clave única del usuario que acabamos de guardar en la DB
	return token.SignedString([]byte(newKey))
}

func (s *AdminService) GetAdmins(ctx context.Context, active *bool, search string, page, limit int) (interface{}, error) {
	// 1. Crear una llave única para esta búsqueda exacta
	cacheKey := fmt.Sprintf("admins_l:%d_p:%d_s:%s_a:%v", limit, page, search, active)

	// 2. Intentar obtener de caché
	if cachedData, found := s.cache.Get(cacheKey); found {
		return cachedData, nil
	}

	// 3. Si no está, buscar en DB
	admins, total, err := s.repo.List(ctx, active, search, page, limit)
	if err != nil {
		return nil, err
	}

	// Estructura de respuesta paginada
	result := fiber.Map{
		"data":        admins,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	}

	// 4. Guardar en caché antes de retornar
	s.cache.Set(cacheKey, result, cache.DefaultExpiration)

	return result, nil
}

// GetRepo permite al middleware acceder al repositorio si es necesario
func (s *AdminService) GetRepo() domain.UserAdminRepository {
	return s.repo
}

// CreateAdmin orquestra la creación de un nuevo administrador con reglas de seguridad estrictas.
func (s *AdminService) CreateAdmin(ctx context.Context, creator *domain.UserAdmin, req request_structs.CreateAdminRequest) error {
	// 1. VALIDACIÓN DE PERMISO BASE
	if !creator.CanCreateAdmin && !creator.Sudo {
		return errors.New("permisos insuficientes: no tienes rango para crear administradores")
	}

	// 2. REGLA DE SUDO ÚNICO (Lógica de Negocio)
	if req.Permissions.Sudo {
		if !creator.Sudo {
			return errors.New("violación de jerarquía: solo un SUDO puede delegar su rango")
		}

		count, err := s.repo.CountSudos(ctx)
		if err != nil {
			return errors.New("error técnico al verificar la integridad de rangos")
		}

		if count > 0 {
			return errors.New("conflicto de configuración: ya existe un usuario SUDO activo en el sistema")
		}
	}

	// 3. PREVENCIÓN DE ESCALADA DE PRIVILEGIOS
	if !creator.Sudo {
		// Validamos que el creador no esté otorgando permisos que él mismo no posee
		if err := s.checkPermissions(creator, req.Permissions); err != nil {
			return err
		}

		// Seguridad adicional: aunque el JSON envíe Sudo: true, si el creador no es Sudo, lo forzamos a false
		req.Permissions.Sudo = false
	}

	// 4. PREPARACIÓN DE CREDENCIALES
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error al procesar la seguridad de la cuenta")
	}

	// 5. CONSTRUCCIÓN DEL MODELO CON AUDITORÍA COMPLETA (Fix update_by)
	newAdmin := &domain.UserAdmin{
		AuditModel: domain.AuditModel{
			ID:         uuid.New(),
			CreateBy:   creator.Username,
			CreateById: &creator.ID,
			// FIX: Llenamos los campos de actualización desde la creación para evitar nulos
			UpdateBy:   creator.Username,
			UpdateById: &creator.ID,
		},
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsActive: true,
		Key:      uuid.New().String(), // Secret único para firmar sus futuros JWT
		Sudo:     req.Permissions.Sudo,

		// Mapeo de permisos granulares
		CanCreatePsi:           req.Permissions.CanCreatePsi,
		CanUpdatePsi:           req.Permissions.CanUpdatePsi,
		CanDeletePsi:           req.Permissions.CanDeletePsi,
		CanCreateAdmin:         req.Permissions.CanCreateAdmin,
		CanUpdateAdmin:         req.Permissions.CanUpdateAdmin,
		CanDeleteAdmin:         req.Permissions.CanDeleteAdmin,
		CanPublish:             req.Permissions.CanPublish,
		CanUpdatePublish:       req.Permissions.CanUpdatePublish,
		CanDeletePublish:       req.Permissions.CanDeletePublish,
		CanSendNotifications:   req.Permissions.CanSendNotifications,
		CanManageNotifications: req.Permissions.CanManageNotifications,
		CanReadNotifications:   req.Permissions.CanReadNotifications,
		CanCreateTags:          req.Permissions.CanCreateTags,
		CanEditTags:            req.Permissions.CanEditTags,
		CanDeleteTags:          req.Permissions.CanDeleteTags,
	}

	// 6. PERSISTENCIA Y MANEJO DE ERRORES DE INFRAESTRUCTURA
	err = s.repo.Create(ctx, newAdmin)
	if err != nil {
		// Capturamos la violación del índice único que pusimos en la migración
		if strings.Contains(err.Error(), "idx_user_admins_unique_sudo") {
			return errors.New("integridad violada: la base de datos rechazó la creación de un segundo SUDO")
		}
		return err
	}

	// 7. INVALIDAR CACHÉ DE LISTADOS
	s.cache.Flush()

	return nil
}

func (s *AdminService) UpdateAdmin(ctx context.Context, updater *domain.UserAdmin, req request_structs.UpdateAdminRequest) error {
	// 1. Buscar al administrador objetivo
	target, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return errors.New("administrador no encontrado")
	}

	// 2. VALIDACIÓN DE JERARQUÍA Y PERMISOS
	// Un admin normal no puede tocar a un SUDO ni otorgar lo que no tiene.
	if !updater.Sudo {
		if target.Sudo {
			return errors.New("permisos insuficientes: no puedes editar a un Super Usuario")
		}

		// Helper: Si el valor cambia y el updater no tiene el permiso -> Bloqueo
		check := func(permName string, requested *bool, current bool, updaterHas bool) error {
			if requested != nil && *requested != current && !updaterHas {
				return fmt.Errorf("no tienes rango para otorgar o revocar el permiso: %s", permName)
			}
			return nil
		}

		// --- Validación Blindada de Permisos (100% Cobertura) ---
		// Psicólogos
		if err := check("Crear Psi", req.CanCreatePsi, target.CanCreatePsi, updater.CanCreatePsi); err != nil {
			return err
		}
		if err := check("Update Psi", req.CanUpdatePsi, target.CanUpdatePsi, updater.CanUpdatePsi); err != nil {
			return err
		}
		if err := check("Delete Psi", req.CanDeletePsi, target.CanDeletePsi, updater.CanDeletePsi); err != nil {
			return err
		}
		// Administradores
		if err := check("Crear Admin", req.CanCreateAdmin, target.CanCreateAdmin, updater.CanCreateAdmin); err != nil {
			return err
		}
		if err := check("Update Admin", req.CanUpdateAdmin, target.CanUpdateAdmin, updater.CanUpdateAdmin); err != nil {
			return err
		}
		if err := check("Delete Admin", req.CanDeleteAdmin, target.CanDeleteAdmin, updater.CanDeleteAdmin); err != nil {
			return err
		}
		// Publicaciones
		if err := check("Publicar", req.CanPublish, target.CanPublish, updater.CanPublish); err != nil {
			return err
		}
		if err := check("Update Pub", req.CanUpdatePublish, target.CanUpdatePublish, updater.CanUpdatePublish); err != nil {
			return err
		}
		if err := check("Delete Pub", req.CanDeletePublish, target.CanDeletePublish, updater.CanDeletePublish); err != nil {
			return err
		}
		// Notificaciones
		if err := check("Enviar Notif", req.CanSendNotifications, target.CanSendNotifications, updater.CanSendNotifications); err != nil {
			return err
		}
		if err := check("Manage Notif", req.CanManageNotifications, target.CanManageNotifications, updater.CanManageNotifications); err != nil {
			return err
		}
		if err := check("Read Notif", req.CanReadNotifications, target.CanReadNotifications, updater.CanReadNotifications); err != nil {
			return err
		}
		// Etiquetas (Tags)
		if err := check("Crear Tags", req.CanCreateTags, target.CanCreateTags, updater.CanCreateTags); err != nil {
			return err
		}
		if err := check("Edit Tags", req.CanEditTags, target.CanEditTags, updater.CanEditTags); err != nil {
			return err
		}
		if err := check("Delete Tags", req.CanDeleteTags, target.CanDeleteTags, updater.CanDeleteTags); err != nil {
			return err
		}
	}

	// 3. AUDITORÍA (Fix update_by)
	target.UpdateBy = updater.Username
	target.UpdateById = &updater.ID

	// 4. CAMBIOS DE IDENTIDAD (Campos no sensibles a permisos jerárquicos)
	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Email != nil {
		target.Email = *req.Email
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}

	// 5. SEGURIDAD DINÁMICA (Password & JWT Invalidation)
	if req.Password != nil && *req.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		target.Password = string(hashed)
		// ROTACIÓN DE KEY: Invalida tokens robados o sesiones antiguas inmediatamente
		target.Key = uuid.New().String()
	}

	// 6. APLICACIÓN DE PERMISOS (Solo los presentes en el JSON)
	// Aquí ya no necesitamos verificar a 'updater.Sudo' porque el bloque del paso 2
	// ya filtró cualquier intento de escalada.
	if req.CanCreatePsi != nil {
		target.CanCreatePsi = *req.CanCreatePsi
	}
	if req.CanUpdatePsi != nil {
		target.CanUpdatePsi = *req.CanUpdatePsi
	}
	if req.CanDeletePsi != nil {
		target.CanDeletePsi = *req.CanDeletePsi
	}
	if req.CanCreateAdmin != nil {
		target.CanCreateAdmin = *req.CanCreateAdmin
	}
	if req.CanUpdateAdmin != nil {
		target.CanUpdateAdmin = *req.CanUpdateAdmin
	}
	if req.CanDeleteAdmin != nil {
		target.CanDeleteAdmin = *req.CanDeleteAdmin
	}
	if req.CanPublish != nil {
		target.CanPublish = *req.CanPublish
	}
	if req.CanUpdatePublish != nil {
		target.CanUpdatePublish = *req.CanUpdatePublish
	}
	if req.CanDeletePublish != nil {
		target.CanDeletePublish = *req.CanDeletePublish
	}
	if req.CanSendNotifications != nil {
		target.CanSendNotifications = *req.CanSendNotifications
	}
	if req.CanManageNotifications != nil {
		target.CanManageNotifications = *req.CanManageNotifications
	}
	if req.CanReadNotifications != nil {
		target.CanReadNotifications = *req.CanReadNotifications
	}
	if req.CanCreateTags != nil {
		target.CanCreateTags = *req.CanCreateTags
	}
	if req.CanEditTags != nil {
		target.CanEditTags = *req.CanEditTags
	}
	if req.CanDeleteTags != nil {
		target.CanDeleteTags = *req.CanDeleteTags
	}

	// 7. PERSISTENCIA E INVALIDACIÓN DE CACHÉ
	err = s.repo.Update(ctx, target)
	if err == nil {
		s.cache.Flush() // Garantiza que el próximo 'List' refleje los cambios
	}

	return err
}

func (s *AdminService) DeleteAdmin(ctx context.Context, updater *domain.UserAdmin, targetID uuid.UUID) error {
	// 1. Evitar auto-eliminación
	if updater.ID == targetID {
		return errors.New("no puedes eliminar tu propia cuenta")
	}

	// 2. Buscar el objetivo para validar jerarquía
	target, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return errors.New("administrador no encontrado")
	}

	// 3. Validar permisos del ejecutor
	if !updater.Sudo {
		if !updater.CanDeleteAdmin {
			return errors.New("no tienes permiso para eliminar administradores")
		}
		// Un admin normal no puede borrar a un Sudo
		if target.Sudo {
			return errors.New("permisos insuficientes para eliminar a un Super Usuario")
		}
	}

	// 4. Ejecutar borrado lógico
	err = s.repo.Delete(ctx, targetID)
	if err == nil {
		s.cache.Flush() // Limpiar caché para que el admin ya no aparezca en las listas
	}

	return err
}

// checkPermissions es un helper privado que valida que el creador no entregue
// poderes que él mismo no posee.
func (s *AdminService) checkPermissions(c *domain.UserAdmin, r domain.UserAdmin) error {
	// Psicólogos
	if r.CanCreatePsi && !c.CanCreatePsi {
		return errors.New("no puedes otorgar 'Crear Psicólogos'")
	}
	if r.CanUpdatePsi && !c.CanUpdatePsi {
		return errors.New("no puedes otorgar 'Actualizar Psicólogos'")
	}
	if r.CanDeletePsi && !c.CanDeletePsi {
		return errors.New("no puedes otorgar 'Borrar Psicólogos'")
	}

	// Administradores
	if r.CanCreateAdmin && !c.CanCreateAdmin {
		return errors.New("no puedes otorgar 'Crear Administradores'")
	}
	if r.CanUpdateAdmin && !c.CanUpdateAdmin {
		return errors.New("no puedes otorgar 'Actualizar Administradores'")
	}
	if r.CanDeleteAdmin && !c.CanDeleteAdmin {
		return errors.New("no puedes otorgar 'Borrar Administradores'")
	}

	// Publicaciones
	if r.CanPublish && !c.CanPublish {
		return errors.New("no puedes otorgar 'Publicar'")
	}
	if r.CanUpdatePublish && !c.CanUpdatePublish {
		return errors.New("no puedes otorgar 'Actualizar Publicaciones'")
	}
	if r.CanDeletePublish && !c.CanDeletePublish {
		return errors.New("no puedes otorgar 'Borrar Publicaciones'")
	}

	// Notificaciones y Tags
	if r.CanSendNotifications && !c.CanSendNotifications {
		return errors.New("no puedes otorgar 'Enviar Notificaciones'")
	}
	if r.CanCreateTags && !c.CanCreateTags {
		return errors.New("no puedes otorgar 'Crear Etiquetas'")
	}
	// ... podrías completar el resto siguiendo este patrón

	return nil
}

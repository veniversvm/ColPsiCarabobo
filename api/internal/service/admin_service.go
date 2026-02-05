// api/internal/service/admin_service.go

package service

import (
	"context"
	"errors"
	"fmt"
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

func (s *AdminService) CreateAdmin(ctx context.Context, creator *domain.UserAdmin, req request_structs.CreateAdminRequest) error {
	// 1. Validar que el creador tenga el permiso base para crear otros admins
	if !creator.CanCreateAdmin && !creator.Sudo {
		return errors.New("no tienes permiso para crear administradores")
	}

	// 2. VALIDACIÓN DE ESCALADA DE PRIVILEGIOS (Senior Logic)
	// Solo el SUDO puede asignar permisos que él mismo no tenga explícitamente.
	// Si no es sudo, verificamos campo por campo.
	if !creator.Sudo {
		if req.Permissions.CanCreatePsi && !creator.CanCreatePsi {
			return errors.New("no puedes asignar 'Crear Psicólogos' si no lo posees")
		}
		if req.Permissions.CanUpdatePsi && !creator.CanUpdatePsi {
			return errors.New("no puedes asignar 'Actualizar Psicólogos' si no lo posees")
		}
		if req.Permissions.CanDeletePsi && !creator.CanDeletePsi {
			return errors.New("no puedes asignar 'Borrar Psicólogos' si no lo posees")
		}
		if req.Permissions.CanCreateAdmin && !creator.CanCreateAdmin {
			return errors.New("no puedes asignar 'Crear Admins' si no lo posees")
		}
		// ... Repetir para el resto de los flags ...

		// Un admin normal NUNCA puede crear un SUDO
		if req.Permissions.Sudo {
			return errors.New("solo un Super Usuario puede crear otro Super Usuario")
		}
	}

	// 3. Hashear password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	// 4. Preparar el nuevo modelo
	newAdmin := &domain.UserAdmin{
		AuditModel: domain.AuditModel{
			CreateBy:   creator.Username,
			CreateById: &creator.ID,
		},
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsActive: true,
		Key:      uuid.New().String(), // Key inicial

		// Asignar permisos validados
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

	return s.repo.Create(ctx, newAdmin)
}

func (s *AdminService) UpdateAdmin(ctx context.Context, updater *domain.UserAdmin, req request_structs.UpdateAdminRequest) error {
	// 1. Buscar al administrador que se quiere editar
	target, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return errors.New("administrador no encontrado")
	}

	// 2. Si no es SUDO, validar jerarquía de permisos
	if !updater.Sudo {
		// No puede editar a un SUDO si él no lo es
		if target.Sudo {
			return errors.New("no tienes rango para editar a un Super Usuario")
		}

		// Función auxiliar para validar cada permiso
		// Si el updater no tiene el permiso, el valor nuevo DEBE ser igual al valor actual
		check := func(permName string, requested *bool, current bool, updaterHas bool) error {
			if requested != nil && *requested != current && !updaterHas {
				return fmt.Errorf("no tienes permiso para otorgar o revocar: %s", permName)
			}
			return nil
		}

		// Validar cada flag sensible
		if err := check("Crear Psicólogos", req.CanCreatePsi, target.CanCreatePsi, updater.CanCreatePsi); err != nil {
			return err
		}
		if err := check("Borrar Psicólogos", req.CanDeletePsi, target.CanDeletePsi, updater.CanDeletePsi); err != nil {
			return err
		}
		if err := check("Gestionar Admins", req.CanCreateAdmin, target.CanCreateAdmin, updater.CanCreateAdmin); err != nil {
			return err
		}
		// ... repetir para los demás campos según sea necesario ...
	}

	// 3. Aplicar cambios permitidos
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
		hashed, _ := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		target.Password = string(hashed)
		// IMPORTANTE: Si cambia password, rotamos Key para cerrar sesiones viejas
		target.Key = uuid.New().String()
	}

	// Actualizar permisos si pasaron la validación
	if req.CanCreatePsi != nil {
		target.CanCreatePsi = *req.CanCreatePsi
	}
	if req.CanDeletePsi != nil {
		target.CanDeletePsi = *req.CanDeletePsi
	}
	if req.CanCreateAdmin != nil {
		target.CanCreateAdmin = *req.CanCreateAdmin
	}
	// ... (actualizar resto de campos)

	// 4. Persistir y Limpiar Caché
	err = s.repo.Update(ctx, target)
	if err == nil {
		s.cache.Flush() // Invalida el caché de búsqueda para reflejar cambios
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

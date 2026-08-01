// api/internal/service/admin_service.go

// Package service implementa la Capa de Casos de Uso (Use Cases) o Lógica de Negocio.
// El AdminService funciona como el "Motor de Reglas", centralizando la gestión de identidad,
// la jerarquía de autorización (Role-Based Access Control / RBAC) y las optimizaciones de
// rendimiento (Caché en Memoria) para el staff administrativo.
package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
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

// AdminService encapsula las dependencias necesarias para las operaciones administrativas.
// Utiliza inyección de dependencias (DI) para conectarse a repositorios y servicios externos.
// Además, mantiene un `go-cache` interno para reducir la carga de la Base de Datos en consultas frecuentes.
type AdminService struct {
	repo        domain.UserAdminRepository
	cache       *cache.Cache
	mailService IMailService // Inyección de servicio de mensajería (Event-Driven)
}

// NewAdminService inicializa el servicio administrativo.
// Configura un TTL (Time-To-Live) base de 5 minutos y un ciclo de "Garbage Collection"
// de 10 minutos para purgar claves expiradas, optimizando el uso de memoria RAM.
func NewAdminService(repo domain.UserAdminRepository, mailService IMailService) *AdminService {
	return &AdminService{
		repo:        repo,
		mailService: mailService,
		cache:       cache.New(5*time.Minute, 10*time.Minute),
	}
}

// =========================================================================
// GESTIÓN DE SESIÓN Y AUTENTICACIÓN
// =========================================================================

// Login procesa la autenticación de miembros del staff.
//
// Implementa "Key Rotation Security" (Invalidación Activa de Sesiones):
// En lugar de emitir un JWT estático, cada inicio de sesión exitoso genera
// un nuevo UUID (`newKey`) y lo guarda en la base de datos como secreto del usuario.
// Esto garantiza el patrón "Single Session Enforcement" (Solo 1 dispositivo activo a la vez):
// al cambiar el Key, cualquier JWT robado o emitido anteriormente en otro dispositivo
// queda criptográficamente inservible de manera instantánea.
func (s *AdminService) Login(ctx context.Context, identifier, password string) (string, *domain.UserAdmin, error) {

	// Sanitización de entrada (Evita fallos por Capitalization en DB)
	lowercased := strings.ToLower(identifier)

	admin, err := s.repo.GetByIdentifier(ctx, lowercased)
	if err != nil {
		// Mensaje genérico: Prevención de fuga de información (Username Enumeration Attack)
		return "", nil, errors.New("credenciales inválidas")
	}

	if !admin.IsActive {
		return "", nil, errors.New("la cuenta está desactivada")
	}

	// Costosa validación criptográfica (Prevención de ataques Timing)
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", nil, errors.New("credenciales inválidas")
	}

	// Renovación de la Key para invalidar tokens anteriores y proporcionar una capa extra de seguridad
	newKey := uuid.Must(uuid.NewV7()).String()
	admin.Key = newKey

	// Se persiste el nuevo Key en la base de datos
	if err := s.repo.Update(ctx, admin); err != nil {
		return "", nil, errors.New("error al procesar inicio de sesión")
	}

	// Emisión del Token JWT firmado con el nuevo Key (UUID v7)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": admin.ID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"role":    "admin",
	})

	// Preparamos payload para la notificación por correo (Alerta de Seguridad)
	mailData := map[string]interface{}{
		"Name":      admin.Username,
		"Email":     admin.Email,
		"LoginTime": time.Now().Local().Format(time.RFC1123),
	}

	// Invocación dinámica y no-bloqueante del servicio de mensajería.
	// Si el servidor SMTP falla o el servicio no está disponible, la
	// autenticación sigue adelante ("Graceful Degradation").
	if s.mailService != nil {
		if err := s.mailService.SendEmail(admin.Email, "Inicio de sesión en el sistema", "login_admin", mailData); err != nil {
			log.Warn().Err(err).Str("component", "admin_service").Msg("Error al preparar el correo (pero el admin se creó)")
		}
	}

	signed, err := token.SignedString([]byte(newKey))
	return signed, admin, err
}

// Logout invalida la sesión del administrador a nivel de servidor (Stateful Logout).
// Al vaciar la Key, el middleware rechazará cualquier request futuro con el JWT anterior.
func (s *AdminService) Logout(ctx context.Context, admin *domain.UserAdmin) error {
	admin.Key = ""
	admin.UpdateBy = admin.Username
	admin.UpdateById = &admin.ID
	return s.repo.UpdateKey(ctx, admin)
}

// =========================================================================
// LECTURA Y RENDIMIENTO (CACHE-ASIDE PATTERN)
// =========================================================================

// GetAdmins recupera una colección paginada de miembros del staff.
//
// Patrón Cache-Aside Inteligente:
// Esta función construye una clave (cacheKey) determinística, combinando de forma
// única los parámetros de paginación y búsqueda (`limits`, `page`, `search`, `active`).
// Esto garantiza que dos usuarios buscando cosas diferentes no crucen sus cachés,
// mientras acelera enormemente respuestas a búsquedas repetitivas.
func (s *AdminService) GetAdmins(
	ctx context.Context,
	active *bool,
	search string,
	page, limit int,
) (interface{}, error) {

	// Sanitización estricta de límites (Prevención DoS)
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	search = utils.CleanAlphaNumeric(search)

	// Generación de llave de caché determinística (Hashing Lógico)
	cacheKey := fmt.Sprintf("admins_l:%d_p:%d_s:%s_a:%v", limit, page, search, active)

	// Intento de lectura Rápida (Cache Hit)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Lectura Lenta a Base de Datos (Cache Miss)
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

	// Se almacena el resultado calculado en memoria (Set) con la expiración por defecto.
	s.cache.Set(cacheKey, result, cache.DefaultExpiration)

	return result, nil
}

// GetRepo expone la interfaz de persistencia.
// Se utiliza principalmente para inyectar este repositorio en la inicialización
// de middlewares (como el AuthMiddleware) sin generar ciclos de importación.
func (s *AdminService) GetRepo() domain.UserAdminRepository {
	return s.repo
}

// =========================================================================
// MOTOR DE VALIDACIÓN DE PERMISOS (MATRIX ENGINE)
// =========================================================================

// permissionUpdate es una estructura interna (Data Transfer Object auxiliar)
// que define las reglas y el mapeo de memoria necesarios para iterar de manera segura
// sobre un conjunto de banderas booleanas de permisos.
type permissionUpdate struct {
	name       string
	requested  *bool
	current    bool
	updaterHas bool
	setTarget  func(bool) // Función de callback (Closure) para mutar el struct de forma segura
}

// buildPermissionMatrix construye una matriz plana de evaluación de seguridad.
//
// Patrón de Diseño: Data-Driven Validation.
// Transforma una validación caótica de docenas de condicionales `if` en un Array
// ordenado que puede iterarse de forma predecible. Es altamente performante ya
// que evita el uso del paquete `reflect` de Go (que es lento) usando funciones lambda (Closures).
// Para escalar el sistema, solo se añade una nueva tupla en esta lista.
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

// CreateAdmin registra a un nuevo integrante del personal de administración.
//
// Implementa el Principio de Escalada Restringida:
// Un administrador común no puede crear otro administrador con privilegios superiores
// a los suyos propios (previniendo "Privilege Escalation"). Esta regla solo puede
// ser saltada por el usuario `Sudo`.
func (s *AdminService) CreateAdmin(
	ctx context.Context,
	creator domain.UserAdmin,
	req request_structs.CreateAdminRequest,
) error {

	// 1. Autorización Base (Gatekeeping)
	if !creator.CanCreateAdmin && !creator.Sudo {
		return errors.New("permisos insuficientes para crear administradores")
	}

	// 2. Sanitización y Validación Básica de Formatos
	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return errors.New("el formato del email es inválido")
	}

	if !utils.IsStrongPassword(req.Password) {
		return errors.New("la contraseña no cumple con los estándares de seguridad")
	}

	// Este paso normaliza caracteres antes de insertar (ej: elimina espacios en blanco inyectados)
	validate_email, err := utils.ParseAndValidateEmail(req.Username)
	if err != nil {
		return errors.New("email inválido")
	}

	// 3. Ensamblaje del Dominio y Trazabilidad (Audit Trail)
	newAdmin := &domain.UserAdmin{
		AuditModel: domain.AuditModel{
			CreateBy:   creator.Username,
			CreateById: &creator.ID,
			UpdateBy:   creator.Username,
			UpdateById: &creator.ID,
		},
		Credentials: domain.Credentials{
			Username: req.Username,
			Email:    validate_email,
			IsActive: true,
			Key:      uuid.Must(uuid.NewV7()).String(),
		},
		Sudo:     false, // Regla Dura: Sudo no puede heredarse ni crearse por API, requiere intervención directa.
	}

	// Hashing Criptográfico seguro utilizando bcrypt.
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error procesando seguridad de la cuenta")
	}
	newAdmin.Password = string(hashed)

	// 4. Procesamiento Dinámico de la Matriz de Permisos
	matrix := buildPermissionMatrix(req.Permissions, newAdmin, creator)

	// Validación: No puedes delegar permisos que tú mismo no posees (Protección de Escalada).
	if !creator.Sudo {
		for _, perm := range matrix {
			// Si el permiso en el Request vino en 'true', y el usuario creador lo tiene en 'false' -> Bloqueo.
			if perm.requested != nil && *perm.requested && !perm.updaterHas {
				return fmt.Errorf("no puedes otorgar el permiso: %s", perm.name)
			}
		}
	}

	// Aplicación: Inyección definitiva de permisos pre-validados.
	for _, perm := range matrix {
		if perm.requested != nil {
			perm.setTarget(*perm.requested)
		}
	}

	// 5. Persistencia y Manejo de Errores Específicos
	err = s.repo.Create(ctx, newAdmin)
	if err != nil {
		if strings.Contains(err.Error(), "idx_user_admins_unique_sudo") {
			return domain.ErrSudoExists
		}
		return err
	}

	// 6. Notificación y Bienvenida Asíncrona
	mailData := map[string]interface{}{
		"Name":     newAdmin.Username,
		"Email":    newAdmin.Email,
		"Password": req.Password,
	}

	if s.mailService != nil {
		if err := s.mailService.SendEmail(newAdmin.Email, "Bienvenido al Colegio de Psicólogos", "welcome_admin", mailData); err != nil {
			log.Warn().Err(err).Str("component", "admin_service").Msg("Error al preparar el correo (pero el admin se creó)")
		}
	}

	// 7. Mantenimiento del Caché (Purge Completo)
	// Al insertar un nuevo registro, el paginado cacheado del listado es inválido. Se limpia preventivamente.
	s.cache.Flush()
	return nil
}

////////////////////////////////////////////////////////////
//////////////////////// UPDATE ////////////////////////////
////////////////////////////////////////////////////////////

// UpdateAdmin gestiona la modificación parcial de perfil y permisos del staff.
//
// Implementa un sistema de Inmunidad Jerárquica:
// Ningún administrador, sin importar cuántos permisos tenga, puede editar
// o modificar los permisos del usuario raíz (`Sudo`).
func (s *AdminService) UpdateAdmin(
	ctx context.Context,
	updater domain.UserAdmin,
	req request_structs.UpdateAdminRequest,
) error {

	target, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return errors.New("administrador no encontrado")
	}

	// 1. Control de Autoridad: Reglas de Inmunidad
	if !updater.Sudo {
		if target.Sudo {
			return errors.New("no puedes editar a un Super Usuario")
		}

		// Construimos la matriz para comprobar si el usuario intenta modificar permisos ajenos.
		matrix := buildPermissionMatrix(req.Permissions, target, updater)

		// Un administrador no puede alterar el estado de un permiso (ya sea encenderlo o apagarlo)
		// de un compañero, si él mismo carece de autoridad sobre ese módulo.
		for _, perm := range matrix {
			if perm.requested != nil &&
				*perm.requested != perm.current &&
				!perm.updaterHas {
				return fmt.Errorf("no tienes rango para modificar: %s", perm.name)
			}
		}
	}

	// 2. Trazabilidad mandataria de última modificación (Auditoría Ciega)
	target.UpdateBy = updater.Username
	target.UpdateById = &updater.ID

	// 3. Mutación Parcial Segura (Patching)
	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Email != nil {
		validate_email, err := utils.ParseAndValidateEmail(*req.Email)
		if err != nil {
			return fmt.Errorf("email inválido")
		}
		target.Email = validate_email
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}

	// Si se cambia la contraseña, la sesión actual se destruye generando un nuevo Key.
	// Esto cierra sesiones comprometidas de forma inmediata.
	if req.Password != nil && *req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		target.Password = string(hashed)
		target.Key = uuid.Must(uuid.NewV7()).String()
	}

	// 4. Aplicación de mutaciones de permisos validadas en el paso 1
	matrix := buildPermissionMatrix(req.Permissions, target, updater)
	for _, perm := range matrix {
		if perm.requested != nil {
			perm.setTarget(*perm.requested)
		}
	}

	// 5. Persistencia y Purga
	if err := s.repo.Update(ctx, target); err != nil {
		return err
	}

	s.cache.Flush() // Limpiar vistas cacheadas del listado de personal
	return nil
}

////////////////////////////////////////////////////////////
//////////////////////// DELETE ////////////////////////////
////////////////////////////////////////////////////////////

// DeleteAdmin ejecuta la baja lógica (Soft Delete) del registro.
//
// Protecciones de Negocio Implementadas:
//  1. Evita Auto-Bloqueo: Un administrador no puede eliminarse a sí mismo en un ataque de pánico.
//  2. Inmunidad Estructural: Nadie, excepto el sistema mismo, puede borrar al Sudo,
//     previniendo que el sistema entero quede irrecuperable (lockout total).
func (s *AdminService) DeleteAdmin(
	ctx context.Context,
	updater *domain.UserAdmin,
	targetID uuid.UUID,
) error {

	// Prevención de "kamikaze" o cierre involuntario de sesión permanente
	if updater.ID == targetID {
		return errors.New("no puedes eliminar tu propia cuenta")
	}

	target, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return errors.New("administrador no encontrado")
	}

	// Chequeo Dual de Rango
	if !updater.Sudo {
		if !updater.CanDeleteAdmin {
			return errors.New("no tienes permiso para eliminar administradores")
		}
		if target.Sudo {
			return errors.New("no puedes eliminar un Super Usuario")
		}
	}

	// Aplicación del borrado lógico delegada al repositorio
	if err := s.repo.Delete(ctx, targetID); err != nil {
		return err
	}

	// Purga obligatoria para eliminarlo de las grillas paginadas del dashboard
	s.cache.Flush()
	return nil
}

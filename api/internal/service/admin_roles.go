// admin_roles.go — Presets de roles del personal administrativo (RBAC-liviano).
//
// Son SOLO atajos: un rol autocompleta los 18 flags de permisos en la UI y su
// etiqueta se persiste en user_admins.role como metadato descriptivo.
// La autorización real SIEMPRE evalúa los booleanos individuales, nunca el rol.
package service

// RoleSlug identifica un preset de permisos.
type RoleSlug string

const (
	RoleSecretaria   RoleSlug = "secretaria"
	RoleComunicacion RoleSlug = "comunicacion"
	RoleSoporte      RoleSlug = "soporte"
	RoleProyectos    RoleSlug = "proyectos"
	RoleLector       RoleSlug = "lector"
	RoleCustom       RoleSlug = "personalizado"
)

// PermissionSet es la vista serializable de los 18 permisos como booleanos
// simples (los DTOs de admin usan *bool para distinguir "no enviado").
type PermissionSet struct {
	CanReadPsi           bool `json:"can_read_psi"`
	CanCreatePsi         bool `json:"can_create_psi"`
	CanUpdatePsi         bool `json:"can_update_psi"`
	CanDeletePsi         bool `json:"can_delete_psi"`
	CanCreateAdmin       bool `json:"can_create_admin"`
	CanUpdateAdmin       bool `json:"can_update_admin"`
	CanDeleteAdmin       bool `json:"can_delete_admin"`
	CanPublish           bool `json:"can_publish"`
	CanUpdatePublish     bool `json:"can_update_publish"`
	CanDeletePublish     bool `json:"can_delete_publish"`
	CanSendNotifications bool `json:"can_send_notifications"`
	CanManageNotifications bool `json:"can_manage_notifications"`
	CanReadNotifications bool `json:"can_read_notifications"`
	CanCreateTags        bool `json:"can_create_tags"`
	CanEditTags          bool `json:"can_edit_tags"`
	CanDeleteTags        bool `json:"can_delete_tags"`
	CanManageProjects    bool `json:"can_manage_projects"`
	CanManageTickets     bool `json:"can_manage_tickets"`
}

// RolePreset define un perfil de permisos predeterminado que la UI puede aplicar.
type RolePreset struct {
	Slug        string        `json:"slug"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Permissions PermissionSet `json:"permissions"`
}

// AdminRolePresets son los perfiles configurables del staff (fuente única).
//
// Nota de diseño: no existe un flag de "solo lectura a nivel de API para todo",
// por eso existe can_read_psi (ver listados y fichas sin poder editar). Los
// presets de solo consulta lo usan en lugar de can_update_psi.
var AdminRolePresets = []RolePreset{
	{
		Slug:        string(RoleSecretaria),
		Name:        "Secretaría",
		Description: "Gestión de colegiados e inscripciones: altas, bajas de agremiados y lectura de notificaciones.",
		Permissions: PermissionSet{
			CanReadPsi:         true,
			CanCreatePsi:       true,
			CanUpdatePsi:       true,
			CanReadNotifications: true,
		},
	},
	{
		Slug:        string(RoleComunicacion),
		Name:        "Comunicación",
		Description: "Contenido institucional: publicar noticias y enviar notificaciones; especialidades con moderación.",
		Permissions: PermissionSet{
			CanPublish:            true,
			CanUpdatePublish:      true,
			CanSendNotifications:  true,
			CanManageNotifications: true,
			CanReadNotifications:  true,
			CanCreateTags:         true,
		},
	},
	{
		Slug:        string(RoleSoporte),
		Name:        "Soporte",
		Description: "Atención de solicitudes: atender la cola de tickets y consultar fichas de los agremiados.",
		Permissions: PermissionSet{
			CanReadPsi:          true,
			CanReadNotifications: true,
			CanManageTickets:    true,
		},
	},
	{
		Slug:        string(RoleProyectos),
		Name:        "Proyectos",
		Description: "Administración de los proyectos (tableros Kanban) del panel.",
		Permissions: PermissionSet{
			CanManageProjects: true,
		},
	},
	{
		Slug:        string(RoleLector),
		Name:        "Lector",
		Description: "Solo consulta: ver el directorio de colegiados y leer notificaciones. No puede editar.",
		Permissions: PermissionSet{
			CanReadPsi:          true,
			CanReadNotifications: true,
		},
	},
}

// isValidRoleSlug valida que el rótulo provenga de un preset conocido (o el
// valor vacío / "personalizado", que la UI usa cuando los flags no coinciden
// con ningún preset). Nunca parte de la autorización: solo evita rótulos inventados.
func isValidRoleSlug(slug string) bool {
	switch RoleSlug(slug) {
	case RoleSecretaria, RoleComunicacion, RoleSoporte, RoleProyectos, RoleLector, RoleCustom:
		return true
	}
	return slug == ""
}
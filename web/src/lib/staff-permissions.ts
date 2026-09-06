// lib/staff-permissions.ts
// Constantes, tipos y presets de permisos del staff, compartidos entre
// crear/editar administrador. Fuente única en el frontend: el backend
// (admin_roles.go) es quien manda; este módulo refleja la misma estructura.
export interface PermissionState {
  can_read_psi: boolean;
  can_create_psi: boolean;
  can_update_psi: boolean;
  can_delete_psi: boolean;
  can_create_admin: boolean;
  can_update_admin: boolean;
  can_delete_admin: boolean;
  can_publish: boolean;
  can_update_publish: boolean;
  can_delete_publish: boolean;
  can_send_notifications: boolean;
  can_manage_notifications: boolean;
  can_read_notifications: boolean;
  can_create_tags: boolean;
  can_edit_tags: boolean;
  can_delete_tags: boolean;
  can_manage_projects: boolean;
  can_manage_tickets: boolean;
}

export const TOTAL_PERMS = 18;

export const PERM_KEYS: (keyof PermissionState)[] = [
  "can_read_psi", "can_create_psi", "can_update_psi", "can_delete_psi",
  "can_create_admin", "can_update_admin", "can_delete_admin",
  "can_publish", "can_update_publish", "can_delete_publish",
  "can_send_notifications", "can_manage_notifications", "can_read_notifications",
  "can_create_tags", "can_edit_tags", "can_delete_tags",
  "can_manage_projects", "can_manage_tickets",
];

export const defaultPerms = (): PermissionState => ({
  can_read_psi: false, can_create_psi: false, can_update_psi: false, can_delete_psi: false,
  can_create_admin: false, can_update_admin: false, can_delete_admin: false,
  can_publish: false, can_update_publish: false, can_delete_publish: false,
  can_send_notifications: false, can_manage_notifications: false, can_read_notifications: false,
  can_create_tags: false, can_edit_tags: false, can_delete_tags: false,
  can_manage_projects: false, can_manage_tickets: false,
});

export const PERM_GROUPS = [
  { label: "Gestión de Colegiados", icon: "👤", color: "blue",
    perms: [{ key: "can_read_psi", label: "Ver" }, { key: "can_create_psi", label: "Crear" }, { key: "can_update_psi", label: "Editar" }, { key: "can_delete_psi", label: "Eliminar" }] },
  { label: "Gestión de Staff", icon: "🛡️", color: "purple",
    perms: [{ key: "can_create_admin", label: "Crear" }, { key: "can_update_admin", label: "Editar" }, { key: "can_delete_admin", label: "Eliminar" }] },
  { label: "Publicaciones", icon: "📰", color: "emerald",
    perms: [{ key: "can_publish", label: "Publicar" }, { key: "can_update_publish", label: "Editar" }, { key: "can_delete_publish", label: "Eliminar" }] },
  { label: "Notificaciones", icon: "🔔", color: "amber",
    perms: [{ key: "can_send_notifications", label: "Enviar" }, { key: "can_manage_notifications", label: "Gestionar" }, { key: "can_read_notifications", label: "Leer" }] },
  { label: "Especialidades / Tags", icon: "🏷️", color: "rose",
    perms: [{ key: "can_create_tags", label: "Crear" }, { key: "can_edit_tags", label: "Editar" }, { key: "can_delete_tags", label: "Eliminar" }] },
  { label: "Proyectos", icon: "📋", color: "indigo",
    perms: [{ key: "can_manage_projects", label: "Gestionar" }] },
  { label: "Tickets", icon: "🎫", color: "cyan",
    perms: [{ key: "can_manage_tickets", label: "Gestionar" }] },
] as const;

export const COLOR_MAP: Record<string, string> = {
  blue: "bg-blue-50 border-blue-200 text-blue-700", purple: "bg-purple-50 border-purple-200 text-purple-700",
  emerald: "bg-emerald-50 border-emerald-200 text-emerald-700", amber: "bg-amber-50 border-amber-200 text-amber-700",
  rose: "bg-rose-50 border-rose-200 text-rose-700", indigo: "bg-indigo-50 border-indigo-200 text-indigo-700",
  cyan: "bg-cyan-50 border-cyan-200 text-cyan-700",
};
export const ACTIVE_MAP: Record<string, string> = {
  blue: "bg-blue-700 border-blue-700 text-white", purple: "bg-purple-700 border-purple-700 text-white",
  emerald: "bg-emerald-700 border-emerald-700 text-white", amber: "bg-amber-700 border-amber-700 text-white",
  rose: "bg-rose-700 border-rose-700 text-white", indigo: "bg-indigo-700 border-indigo-700 text-white",
  cyan: "bg-cyan-700 border-cyan-700 text-white",
};

// Preset tal cual lo devuelve GET /admin/roles/presets (admin_roles.go).
export interface RolePreset {
  slug: string;
  name: string;
  description: string;
  permissions: Partial<Record<keyof PermissionState, boolean>>;
}

export const ROLE_LABELS: Record<string, string> = {
  secretaria: "Secretaría",
  comunicacion: "Comunicación",
  soporte: "Soporte",
  proyectos: "Proyectos",
  lector: "Lector",
  personalizado: "Personalizado",
};

// Devuelve el slug del preset cuyos permisos coinciden EXACTAMENTE con los
// actuales; null si no corresponde a ninguno (permisos "a medida").
export const findRoleForPerms = (perms: PermissionState, presets: RolePreset[]): string | null => {
  for (const p of presets) {
    const matches = PERM_KEYS.every(
      (key) => (perms[key] ?? false) === (p.permissions[key] ?? false)
    );
    if (matches) return p.slug;
  }
  return null;
};

// Aplica los permisos de un preset a un estado de permisos.
export const applyPresetToPerms = (perms: PermissionState, preset: RolePreset): PermissionState => {
  const next: PermissionState = { ...defaultPerms(), ...perms };
  PERM_KEYS.forEach((key) => { next[key] = preset.permissions[key] ?? false; });
  return next;
};

// Rótulo legible para el badge del listado (o texto vacío si no hay rol).
export const roleLabel = (slug: string | null | undefined): string =>
  slug ? (ROLE_LABELS[slug] ?? "—") : "—";
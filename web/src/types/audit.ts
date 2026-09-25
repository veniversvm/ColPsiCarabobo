// web/src/types/audit.ts
// Tipos del módulo "Auditoría" (bitácora de cambios a nivel de API).
// Espejan el modelo Go domain/api_change_log.model.go.

/** Entidades registradas en la bitácora. */
export type AuditEntity = "psi" | "staff" | "notificacion" | "auth" | "otro";

/** Cambio por campo: { campo: { from, to } }. */
export interface AuditChange {
  from?: unknown;
  to?: unknown;
}

export interface ApiChangeLog {
  id: string;
  entity: AuditEntity;
  entity_id: string;
  entity_label: string;
  action: string;
  actor_id?: string | null;
  actor_role?: string | null;
  actor_username?: string | null;
  ip?: string | null;
  user_agent?: string | null;
  /**
   * Diff por campo. La API lo persiste como jsonb: llega como objeto
   * `{ campo: { from, to } }`; también puede llegar serializado como string
   * (historias previas al fix de la columna).
   */
  changes: Record<string, AuditChange> | string | null;
  /** JSON libre (metadata del suceso): objeto o string serializado. */
  metadata: Record<string, unknown> | string | null;
  created_at: string;
  updated_at?: string | null;
}

export interface AuditListResponse {
  data: ApiChangeLog[];
  total: number;
  page: number;
  limit: number;
}

export interface AuditStat {
  entity: string;
  action: string;
  count: number;
}

export interface AuditStatsResponse {
  since: string;
  stats: AuditStat[];
}

/** Rótulos humanos de las acciones (fallback: el snake_case crudo). */
export const ACTION_LABELS: Record<string, string> = {
  create: "Creación",
  update: "Actualización",
  delete: "Eliminación",
  login: "Inicio de sesión",
  logout: "Cierre de sesión",
  reset_password: "Reinicio de contraseña",
  transfer_sudo: "Sucesión de SUDO",
  role_change: "Cambio de rol",
  update_permissions: "Cambio de permisos",
  change_state: "Cambio de estado",
  send: "Envío",
  cancel: "Cancelación",
  estado: "Cambio de estado",
  move: "Movimiento",
};

export const ENTITY_LABELS: Record<string, string> = {
  psi: "Psicólogo",
  staff: "Staff",
  notificacion: "Notificación",
  auth: "Autenticación",
  otro: "Otro",
  ticket: "Ticket",
  post: "Publicación",
  proyecto: "Proyecto",
  area: "Área",
  ficha_inscripcion: "Ficha de inscripción",
  config: "Configuración",
};

/** Traducción amigable de un suceso crudo (o el crudo si no hay rótulo). */
export const actionLabel = (a: string): string => ACTION_LABELS[a] ?? a;
/** Traducción amigable de una entidad cruda. */
export const entityLabel = (e: string): string => ENTITY_LABELS[e] ?? e;

// =============================================================================
// Rótulos de campos del diff (compartidos por el drawer y las tarjetas)
// =============================================================================

/** Rótulos legibles para los campos del diff más comunes. */
export const FIELD_LABELS: Record<string, string> = {
  // Identidad y credenciales
  first_name: "Primer nombre",
  second_name: "Segundo nombre",
  last_name: "Apellido",
  second_last_name: "Segundo apellido",
  fpv: "Nº FPV",
  ci: "Cédula",
  nationality: "Nacionalidad",
  control_number: "Nº de control",
  genre: "Género",
  solvent: "Solvencia",
  proof_of_life: "Fe de vida",
  is_active: "Activo",
  username: "Usuario",
  email: "Correo",
  // Contacto
  contact_phone: "Teléfono de contacto",
  contact_cell_phone: "Celular de contacto",
  contact_email: "Correo de contacto",
  service_address: "Dirección de servicio",
  // Ubicaciones y privacidad (auto-gestión)
  show_contact_email: "Mostrar correo de contacto",
  show_public_service_address: "Mostrar dirección de servicio",
  show_municipality_carabobo: "Mostrar municipio (Carabobo)",
  phone_carabobo: "Teléfono (Carabobo)",
  show_phone_carabobo: "Mostrar teléfono (Carabobo)",
  cel_phone_carabobo: "Celular (Carabobo)",
  show_cel_phone_carabobo: "Mostrar celular (Carabobo)",
  municipality_carabobo: "Municipio (Carabobo)",
  state_outside: "Estado (fuera)",
  show_state_outside: "Mostrar estado (fuera)",
  municipality_outside_carabobo: "Municipio (fuera)",
  show_municipality_outside_carabobo: "Mostrar municipio (fuera)",
  phone_outside_carabobo: "Teléfono (fuera de Carabobo)",
  show_phone_outside_carabobo: "Mostrar teléfono (fuera de Carabobo)",
  cel_phone_outside_carabobo: "Celular (fuera de Carabobo)",
  show_cel_phone_outside_carabobo: "Mostrar celular (fuera de Carabobo)",
  service_address_outside_carabobo: "Dirección (fuera de Carabobo)",
  show_public_service_address_outside_carabobo: "Mostrar dirección (fuera de Carabobo)",
  country: "País",
  phone_outside_venezuela: "Teléfono (fuera de Venezuela)",
  show_phone_outside_venezuela: "Mostrar teléfono (fuera de Venezuela)",
  cell_phone_outside_venezuela: "Celular (fuera de Venezuela)",
  show_cell_phone_outside_venezuela: "Mostrar celular (fuera de Venezuela)",
  service_address_outside_venezuela: "Dirección (fuera de Venezuela)",
  show_public_service_address_outside_venezuela: "Mostrar dirección (fuera de Venezuela)",
  // Perfil profesional
  primary_work_area: "Área de ejercicio principal",
  secondary_work_area: "Área de ejercicio secundaria",
  primary_specialty_id: "Especialidad principal",
  secondary_specialty_id: "Especialidad secundaria",
  service_modality_presencial: "Modalidad presencial",
  service_modality_distance: "Modalidad a distancia",
  service_modality_telephone: "Modalidad telefónica",
  show_service_modality: "Mostrar modalidad de servicio",
  mini_bio: "Mini biografía",
  // Privacidad académica (col_data)
  show_university_undergraduate: "Mostrar universidad de pregrado",
  show_graduate_date: "Mostrar fecha de egreso",
  show_mention_undergraduate: "Mostrar mención de pregrado",
  birthday_notification: "Aviso de cumpleaños",
  // Títulos académicos
  post_grade_title: "Título",
  post_grade_university: "Universidad",
  post_grade_graduation_year: "Año de graduación",
  post_grade_description: "Descripción",
  // Redes sociales
  social_name: "Plataforma",
  social_url: "URL",
  social_active: "Red activa",
  // Staff / RBAC
  role: "Rol",
  can_read_psi: "Permiso · Ver psi", can_create_psi: "Permiso · Crear psi", can_update_psi: "Permiso · Editar psi", can_delete_psi: "Permiso · Eliminar psi",
  can_create_admin: "Permiso · Crear staff", can_update_admin: "Permiso · Editar staff", can_delete_admin: "Permiso · Eliminar staff",
  can_publish: "Permiso · Publicar", can_update_publish: "Permiso · Editar publicaciones", can_delete_publish: "Permiso · Eliminar publicaciones",
  can_send_notifications: "Permiso · Enviar notificaciones", can_manage_notifications: "Permiso · Gestionar notificaciones", can_read_notifications: "Permiso · Leer notificaciones",
  can_create_tags: "Permiso · Crear tags", can_edit_tags: "Permiso · Editar tags", can_delete_tags: "Permiso · Eliminar tags",
  can_manage_projects: "Permiso · Proyectos", can_manage_tickets: "Permiso · Tickets",
  can_view_logs: "Permiso · Ver bitácora", can_export_logs: "Permiso · Exportar bitácora",
  status: "Estado",
};

/** Traducción amigable de un campo del diff (fallback: capitaliza el snake_case). */
export const fieldLabel = (k: string): string => FIELD_LABELS[k] ?? k.replaceAll("_", " ").replace(/^\w|\s\w/g, (c) => c.toUpperCase());

/**
 * Parsea `changes`/`metadata` de la bitácora de forma defensiva: la API los
 * manda como objeto (jsonb) o como string serializado (historias previas).
 * Nunca lanza; devuelve siempre un objeto plano.
 */
export const parseAuditJson = (v: unknown): Record<string, unknown> => {
  if (!v) return {};
  if (typeof v === "string") {
    try {
      const p = JSON.parse(v);
      return p && typeof p === "object" && !Array.isArray(p) ? (p as Record<string, unknown>) : {};
    } catch {
      return {};
    }
  }
  if (typeof v === "object" && !Array.isArray(v)) return v as Record<string, unknown>;
  return {};
};

/** Claves de un diff de cambios (para resúmenes compactos). */
export const auditChangesKeys = (changes: unknown): string[] => Object.keys(parseAuditJson(changes));
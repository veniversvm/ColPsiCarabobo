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
  /** JSON crudo: { campo: { from, to } } */
  changes: string;
  /** JSON crudo libre (metadata del suceso). */
  metadata: string;
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
};

export const ENTITY_LABELS: Record<string, string> = {
  psi: "Psicólogo",
  staff: "Staff",
  notificacion: "Notificación",
  auth: "Autenticación",
  otro: "Otro",
};

/** Traducción amigable de un suceso crudo (o el crudo si no hay rótulo). */
export const actionLabel = (a: string): string => ACTION_LABELS[a] ?? a;
/** Traducción amigable de una entidad cruda. */
export const entityLabel = (e: string): string => ENTITY_LABELS[e] ?? e;
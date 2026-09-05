// web/src/types/tickets.ts
// Tipos del módulo "Tickets de Solicitudes". Espejan los DTOs de la API Go
// (domain/ticket.model.go). A diferencia del resto del sistema, los IDs de
// este módulo son numéricos (auto-increment), decidido así en el backend.

export type TicketAuthorType = "admin" | "psi" | "system";

export interface TicketAdjunto {
  id: number;
  mensaje_id: number;
  original_name: string;
  mime_type: string;
  size_bytes: number;
  created_at: string;
  updated_at: string;
  /** URL pública resuelta (S3_PUBLIC_URL) — computada por la API. */
  url: string;
}

export interface TicketMensaje {
  id: number;
  ticket_id: number;
  author_type: TicketAuthorType;
  author_admin_id?: string | null;
  author_psi_id?: string | null;
  author_name?: string;
  message: string;
  created_at: string;
  updated_at: string;
  adjuntos?: TicketAdjunto[];
}

export interface TicketEstado {
  id: number;
  motivo_id: number;
  name: string;
  order: number;
  is_closed: boolean;
  created_at?: string;
  updated_at?: string;
  motivo?: TicketMotivo;
}

export interface TicketMotivo {
  id: number;
  name: string;
  description?: string;
  tickets_per_psi: number;
  created_at?: string;
  updated_at?: string;
  estados?: TicketEstado[];
}

export interface TicketStatusLog {
  id: number;
  ticket_id: number;
  previous_state_id?: number | null;
  new_state_id: number;
  changed_by_type: TicketAuthorType;
  changed_by_admin_id?: string | null;
  changed_by_psi_id?: string | null;
  reason?: string;
  created_at: string;
  updated_at: string;
  new_state?: TicketEstado;
  previous_state?: TicketEstado;
}

export interface Ticket {
  id: number;
  psi_user_id: string;
  motivo_id: number;
  estado_id: number;
  title: string;
  description: string;
  close_reason?: string;
  closed_by_type?: TicketAuthorType | "";
  closed_by_admin_id?: string | null;
  closed_by_psi_id?: string | null;
  closed_at?: string | null;
  motivo?: TicketMotivo;
  estado?: TicketEstado;
  mensajes?: TicketMensaje[];
  status_logs?: TicketStatusLog[];
  psi_first_name?: string;
  psi_last_name?: string;
  is_closed: boolean;
  created_at: string;
  updated_at: string;
  create_by?: string;
}

export interface TicketsListResponse {
  data: Ticket[];
  total: number;
  page: number;
  limit: number;
}

export interface MotivoConfigQueryResponse {
  data: TicketMotivo[];
}

export interface PendientesResponse {
  pendientes: number;
}

// ── Requests ──────────────────────────────────────────────────────────────

export interface CreateTicketMotivoRequest {
  name: string;
  description?: string;
  tickets_per_psi: number;
}

export interface UpdateTicketMotivoRequest {
  name?: string;
  description?: string;
  tickets_per_psi?: number;
}

export interface CreateTicketEstadoRequest {
  motivo_id: number;
  name: string;
  order: number;
  is_closed: boolean;
}

export interface UpdateTicketEstadoAdminRequest {
  estado_id: number;
  reason?: string;
}

export interface CloseTicketRequest {
  close_reason: string;
}

// ── Constantes (espeja el backend) ────────────────────────────────────────

/** Máximo de caracteres por comentario del psicólogo. */
export const MAX_PSI_MENSAJE_CHARS = 1000;
/** Máximo de caracteres por comentario del admin. */
export const MAX_ADMIN_MENSAJE_CHARS = 4000;
/** El psi no puede publicar más de 3 mensajes seguidos. */
export const MAX_PSI_CONSECUTIVE = 3;
/** Límites de la configuración. */
export const MAX_TICKET_TITLE_CHARS = 200;
export const MAX_TICKET_DESC_CHARS = 2000;
export const MAX_CLOSE_REASON_CHARS = 500;

// ── Helpers ───────────────────────────────────────────────────────────────

export function estadoColor(estado?: TicketEstado | null): string {
  if (!estado) return "bg-gray-100 text-gray-600";
  if (estado.is_closed) return "bg-red-100 text-red-700";
  return "bg-blue-100 text-blue-700";
}

export function formatTicketDate(value?: string | null): string {
  if (!value) return "";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString("es-VE", { year: "numeric", month: "short", day: "numeric" });
}

export function formatTicketDateTime(value?: string | null): string {
  if (!value) return "";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleString("es-VE", {
    year: "numeric", month: "short", day: "numeric", hour: "2-digit", minute: "2-digit",
  });
}

export function formatFileSize(bytes?: number | null): string {
  if (!bytes || bytes <= 0) return "";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
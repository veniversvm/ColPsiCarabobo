// web/src/types/notifications.ts
export type NotificationTargetType = "global" | "individual" | "group";
export type NotificationStatus = "pending" | "sent" | "failed" | "cancelled";

export interface NotificationFilter {
  id: string;
  notification_id: string;
  municipality?: string;
  state?: string;
  genre?: string;
  specialty_id?: number;
  solvent?: boolean;
  target_user_ids?: string;
}

export interface NotificationTarget {
  id: string;
  notification_id: string;
  psi_user_id: string;
  is_read: boolean;
  read_at?: string;
  email_sent: boolean;
  email_sent_at?: string;
  psi_user?: {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    ci?: number;
  };
}

export interface Notification {
  id: string;
  created_at?: string;
  updated_at?: string;
  create_by?: string;
  title: string;
  message: string;
  target_type: NotificationTargetType;
  sender_id: string;
  send_email: boolean;
  scheduled_at?: string;
  sent_at?: string;
  status: NotificationStatus;
  targets?: NotificationTarget[];
  filters?: NotificationFilter[];
  attachments?: NotificationAttach[];
}

export interface NotificationAttach {
  id: string;
  notification_id: string;
  name: string;
  s3_key: string;
  content_type?: string;
}

// ── Requests ───────────────────────────────────────────────────────────

export interface NotificationFilterDTO {
  municipality?: string;
  state?: string;
  genre?: string;
  specialty_id?: number;
  solvent?: boolean;
}

export interface CreateNotificationRequest {
  title: string;
  message: string;
  target_type: NotificationTargetType;
  send_email?: boolean;
  scheduled_at?: string;
  filters?: NotificationFilterDTO;
  target_user_ids?: string[];
}

// ── Responses ──────────────────────────────────────────────────────────

export interface PsiPreviewUser {
  id: string;
  name: string;
  email: string;
}

export interface PreviewResponse {
  total_recipients: number;
  recipients: PsiPreviewUser[];
}

export interface CreateNotificationResponse {
  id: string;
  status: NotificationStatus;
  total_sent?: number;
  total_emails?: number;
  scheduled_at?: string;
  message?: string;
}

export interface NotificationDetailResponse {
  notification: Notification;
  total_recipients: number;
  total_read: number;
  total_unread: number;
}

export interface UnreadCountResponse {
  unread_count: number;
}

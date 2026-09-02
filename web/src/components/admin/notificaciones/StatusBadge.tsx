// web/src/components/admin/notificaciones/StatusBadge.tsx
import { NotificationStatus } from "~/types/notifications";

const STYLES: Record<NotificationStatus, string> = {
  pending: "bg-amber-100 text-amber-700 border-amber-200",
  sent: "bg-green-100 text-green-700 border-green-200",
  failed: "bg-red-100 text-red-700 border-red-200",
  cancelled: "bg-gray-200 text-gray-600 border-gray-300",
};

const LABELS: Record<NotificationStatus, string> = {
  pending: "Programada",
  sent: "Enviada",
  failed: "Fallida",
  cancelled: "Cancelada",
};

export function StatusBadge({ status }: { status: NotificationStatus }) {
  return (
    <span
      class={`inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-black uppercase tracking-wider border ${STYLES[status] || STYLES.pending}`}
    >
      {LABELS[status] || status}
    </span>
  );
}

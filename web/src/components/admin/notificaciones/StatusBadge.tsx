// web/src/components/admin/notificaciones/StatusBadge.tsx
import { NotificationStatus } from "~/types/notifications";

const STYLES: Record<NotificationStatus, string> = {
  pending: "bg-amber-50 text-amber-700 border-amber-200",
  sent: "bg-emerald-50 text-emerald-700 border-emerald-200",
  failed: "bg-red-50 text-red-700 border-red-200",
  cancelled: "bg-slate-100 text-slate-600 border-slate-200",
};

const LABELS: Record<NotificationStatus, string> = {
  pending: "Programada",
  sent: "Enviada",
  failed: "Fallida",
  cancelled: "Cancelada",
};

export function StatusBadge({ status }: { status: NotificationStatus }) {
  const style = STYLES[status] || STYLES.pending;
  return (
    <span class={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded border text-[11px] font-medium whitespace-nowrap ${style}`}>
      <span class="w-1.5 h-1.5 rounded-full bg-current" />
      {LABELS[status] || status}
    </span>
  );
}
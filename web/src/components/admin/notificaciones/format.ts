// web/src/components/admin/notificaciones/format.ts
import { NotificationTargetType } from "~/types/notifications";

export function formatNotifDate(value?: string | null): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString("es-VE", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function targetTypeLabel(type?: NotificationTargetType | string): string {
  switch (type) {
    case "global":
      return "Global";
    case "individual":
      return "Individual";
    case "group":
      return "Por grupo";
    default:
      return type || "—";
  }
}

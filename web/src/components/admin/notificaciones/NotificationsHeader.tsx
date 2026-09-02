// web/src/components/admin/notificaciones/NotificationsHeader.tsx
import { A } from "@solidjs/router";

export function NotificationsHeader() {
  return (
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
      <div>
        <h1 class="text-2xl md:text-3xl font-black text-gray-800 uppercase tracking-tight">
          Notificaciones
        </h1>
        <p class="text-sm text-gray-500 font-medium mt-1">
          Envía avisos a agremiados (global, individual o por grupo)
        </p>
      </div>
      <A
        href="/admin/notificaciones/crear"
        class="inline-flex items-center justify-center gap-2 bg-blue-800 hover:bg-blue-900 text-white font-black px-6 py-3 rounded-2xl shadow-sm active:scale-95 transition-all"
      >
        <span>➕</span> Nueva Notificación
      </A>
    </div>
  );
}

// web/src/routes/admin/notificaciones/[id].tsx
import { createResource, For, Show, Suspense, ErrorBoundary } from "solid-js";
import { A, useParams } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import {
  NotificationDetailResponse,
  NotificationTarget,
} from "~/types/notifications";
import { StatusBadge, formatNotifDate, targetTypeLabel } from "~/components/admin/notificaciones";

export default function AdminNotificacionDetalle() {
  const params = useParams<{ id: string }>();

  const [detail] = createResource(
    () => apiGet<NotificationDetailResponse>(`/notifications/admin/${params.id}`)
  );
  const [targets] = createResource(
    () => apiGet<NotificationTarget[]>(`/notifications/admin/${params.id}/targets`)
  );

  const n = () => detail()?.notification;

  return (
    <main class="pb-20 animate-in fade-in duration-500 max-w-3xl">
      <A href="/admin/notificaciones" class="inline-flex items-center gap-1 text-sm font-bold text-blue-600 hover:text-blue-800 mb-4">
        ← Volver a Notificaciones
      </A>

      <ErrorBoundary fallback={(err, reset) => (
        <div class="bg-red-50 border border-red-200 p-8 rounded-3xl text-center">
          <p class="text-4xl mb-3">🚨</p>
          <p class="text-red-700 font-black mb-2">No se pudo cargar la notificación</p>
          <p class="text-sm text-red-600 mb-4">{getUserFacingError(err)}</p>
          <button onClick={reset} class="bg-red-600 text-white font-black px-6 py-2.5 rounded-xl text-sm">↻ Reintentar</button>
        </div>
      )}>
        <Suspense fallback={<div class="h-48 bg-white animate-pulse rounded-3xl border border-colpsi-border" />}>
          <Show when={n()}>
            <div class="bg-white rounded-3xl border border-colpsi-border shadow-sm p-6 md:p-8">
              <div class="flex items-center gap-2 flex-wrap mb-3">
                <StatusBadge status={n()!.status} />
                <span class="text-[10px] font-black uppercase tracking-wider text-gray-400 bg-gray-100 rounded-full px-2 py-0.5">
                  {targetTypeLabel(n()!.target_type)}
                </span>
                {n()!.send_email && <span class="text-[10px] font-black uppercase tracking-wider text-blue-600 bg-blue-50 rounded-full px-2 py-0.5">✉️ Email</span>}
              </div>

              <h1 class="text-2xl font-black text-gray-800 mb-2">{n()!.title}</h1>
              <p class="text-sm text-gray-500 mb-5 font-medium">Enviada por {n()!.create_by || "admin"}</p>

              <div class="prose prose-sm max-w-none text-gray-700 whitespace-pre-wrap mb-6">{n()!.message}</div>

              <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 border-t border-colpsi-border pt-5 text-center">
                <div>
                  <p class="text-2xl font-black text-gray-800">{detail()?.total_recipients ?? 0}</p>
                  <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">Destinatarios</p>
                </div>
                <div>
                  <p class="text-2xl font-black text-green-600">{detail()?.total_read ?? 0}</p>
                  <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">Leídos</p>
                </div>
                <div>
                  <p class="text-2xl font-black text-amber-500">{detail()?.total_unread ?? 0}</p>
                  <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">No leídos</p>
                </div>
                <div>
                  <p class="text-sm font-black text-gray-700 truncate">{formatNotifDate(n()!.scheduled_at || n()!.sent_at)}</p>
                  <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">Fecha</p>
                </div>
              </div>
            </div>
          </Show>
        </Suspense>
      </ErrorBoundary>

      {/* Destinatarios */}
      <section class="mt-6">
        <h2 class="text-lg font-black text-gray-800 mb-3">
          Destinatarios <span class="text-gray-400 text-base">({targets()?.length ?? 0})</span>
        </h2>

        <Suspense fallback={<div class="h-24 bg-white animate-pulse rounded-2xl border border-colpsi-border" />}>
          <Show when={(targets()?.length ?? 0) === 0} fallback={undefined}>
            <div class="bg-white rounded-2xl border border-colpsi-border p-8 text-center text-sm text-gray-500">
              Sin destinatarios registrados.
            </div>
          </Show>

          <div class="space-y-2">
            <For each={targets() ?? []}>
              {(t) => (
                <div class="bg-white rounded-xl border border-colpsi-border px-4 py-3 flex items-center justify-between">
                  <div class="min-w-0">
                    <p class="text-sm font-bold text-gray-800 truncate">
                      {t.psi_user?.first_name} {t.psi_user?.last_name || "—"}
                    </p>
                    <p class="text-xs text-gray-400 truncate">{t.psi_user?.email}</p>
                  </div>
                  <span
                    class={`shrink-0 text-[10px] font-black uppercase tracking-wider px-2.5 py-1 rounded-full ${
                      t.is_read ? "bg-green-100 text-green-700" : "bg-amber-100 text-amber-700"
                    }`}
                  >
                    {t.is_read ? "Leído" : "No leído"}
                  </span>
                </div>
              )}
            </For>
          </div>
        </Suspense>
      </section>
    </main>
  );
}

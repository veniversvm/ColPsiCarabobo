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
import { Icon } from "~/components/admin/ui/icons";

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
    <main class="space-y-4 pb-12 max-w-3xl">
      <div class="flex items-center gap-3 pb-4 border-b border-colpsi-border">
        <A href="/admin/notificaciones" class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors" title="Volver">
          <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
        </A>
        <div class="min-w-0">
          <h1 class="text-lg font-semibold text-colpsi-text truncate">{n()?.title ?? "Notificación"}</h1>
          <p class="text-sm text-colpsi-muted mt-0.5">Detalle del aviso enviado</p>
        </div>
      </div>

      <ErrorBoundary fallback={(err, reset) => (
        <div class="bg-white border border-colpsi-border p-8 rounded-lg text-center">
          <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-red-50 text-colpsi-red mb-4">
            <Icon name="alertTriangle" class="w-6 h-6" />
          </span>
          <p class="text-colpsi-text font-semibold mb-1">No se pudo cargar la notificación</p>
          <p class="text-sm text-colpsi-muted mb-5">{getUserFacingError(err)}</p>
          <button onClick={reset} class="inline-flex items-center gap-2 h-9 px-4 rounded-md bg-colpsi-red text-white font-semibold hover:opacity-90 transition-colors text-sm">
            <Icon name="refresh" class="w-4 h-4" />
            Reintentar
          </button>
        </div>
      )}>
        <Suspense fallback={<div class="h-48 bg-white animate-pulse rounded-lg border border-colpsi-border" />}>
          <Show when={n()}>
            <div class="bg-white rounded-lg border border-colpsi-border p-5">
              <div class="flex items-center gap-2 flex-wrap mb-3">
                <StatusBadge status={n()!.status} />
                <span class="inline-flex items-center text-[10px] font-medium uppercase tracking-wide text-colpsi-muted bg-colpsi-bg rounded px-2 py-0.5">
                  {targetTypeLabel(n()!.target_type)}
                </span>
                {n()!.send_email && <span class="inline-flex items-center gap-1 text-[10px] font-medium uppercase tracking-wide text-colpsi-blue bg-colpsi-blue/5 rounded px-2 py-0.5"><Icon name="mail" class="w-3 h-3" /> Email</span>}
              </div>

              <h2 class="text-xl font-semibold text-colpsi-text mb-2">{n()!.title}</h2>
              <p class="text-sm text-colpsi-muted mb-5 font-medium">Enviada por {n()!.create_by || "admin"}</p>

              <div class="prose prose-sm max-w-none text-colpsi-text whitespace-pre-wrap mb-6">{n()!.message}</div>

              <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 border-t border-colpsi-border pt-4 text-center">
                <div>
                  <p class="text-xl font-semibold text-colpsi-text">{detail()?.total_recipients ?? 0}</p>
                  <p class="text-[10px] font-semibold text-colpsi-muted uppercase tracking-wide">Destinatarios</p>
                </div>
                <div>
                  <p class="text-xl font-semibold text-emerald-600">{detail()?.total_read ?? 0}</p>
                  <p class="text-[10px] font-semibold text-colpsi-muted uppercase tracking-wide">Leídos</p>
                </div>
                <div>
                  <p class="text-xl font-semibold text-amber-500">{detail()?.total_unread ?? 0}</p>
                  <p class="text-[10px] font-semibold text-colpsi-muted uppercase tracking-wide">No leídos</p>
                </div>
                <div>
                  <p class="text-sm font-medium text-colpsi-text truncate">{formatNotifDate(n()!.scheduled_at || n()!.sent_at)}</p>
                  <p class="text-[10px] font-semibold text-colpsi-muted uppercase tracking-wide">Fecha</p>
                </div>
              </div>
            </div>
          </Show>
        </Suspense>
      </ErrorBoundary>

      {/* Destinatarios */}
      <section class="space-y-3">
        <h2 class="text-base font-semibold text-colpsi-text">
          Destinatarios <span class="text-colpsi-muted text-sm">({targets()?.length ?? 0})</span>
        </h2>

        <Suspense fallback={<div class="h-24 bg-white animate-pulse rounded-lg border border-colpsi-border" />}>
          <Show when={(targets()?.length ?? 0) === 0} fallback={undefined}>
            <div class="bg-white rounded-lg border border-colpsi-border p-8 text-center text-sm text-colpsi-muted">
              Sin destinatarios registrados.
            </div>
          </Show>

          <div class="space-y-2">
            <For each={targets() ?? []}>
              {(t) => (
                <div class="bg-white rounded-lg border border-colpsi-border px-4 py-3 flex items-center justify-between">
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-colpsi-text truncate">
                      {t.psi_user?.first_name} {t.psi_user?.last_name || "—"}
                    </p>
                    <p class="text-xs text-colpsi-muted truncate">{t.psi_user?.email}</p>
                  </div>
                  <span
                    class={`shrink-0 inline-flex items-center gap-1.5 text-[11px] font-medium px-2 py-0.5 rounded-md border ${
                      t.is_read ? "bg-emerald-50 text-emerald-700 border-emerald-200" : "bg-amber-50 text-amber-700 border-amber-200"
                    }`}
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-current" />
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
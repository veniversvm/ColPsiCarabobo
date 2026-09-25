// web/src/routes/admin/notificaciones/index.tsx
import { createResource, createSignal, For, Show, Suspense, ErrorBoundary } from "solid-js";
import { A, action, useAction } from "@solidjs/router";
import { apiDelete, apiGet } from "~/lib/api";
import { PaginatedResponse } from "~/types/admin";
import { Notification } from "~/types/notifications";
import { NotificationsHeader, StatusBadge, formatNotifDate, targetTypeLabel } from "~/components/admin/notificaciones";
import { Icon } from "~/components/admin/ui/icons";

const cancelNotification = action(async (id: string) => {
  "use server";
  return await apiDelete(`/notifications/admin/${id}`);
});

export default function AdminNotificacionesPage() {
  const runCancel = useAction(cancelNotification);

  const [page, setPage] = createSignal(1);
  const [search, setSearch] = createSignal("");
  const [cancelId, setCancelId] = createSignal<string | null>(null);
  const [busy, setBusy] = createSignal<string | null>(null);

  const [list, { refetch }] = createResource(
    () => apiGet<{ data: Notification[]; total: number; page: number }>(`/notifications/admin?page=${page()}&limit=15`),
    { initialValue: { data: [], total: 0, page: 1 } }
  );

  const items = () => list()?.data ?? [];

  const filtered = () => {
    const q = search().toLowerCase().trim();
    if (!q) return items();
    return items().filter((n) => (n.title || "").toLowerCase().includes(q));
  };

  const handleCancel = async () => {
    const id = cancelId();
    if (!id) return;
    setBusy(id);
    try {
      await runCancel(id);
      setCancelId(null);
      refetch();
    } finally {
      setBusy(null);
    }
  };

  const totalPages = Math.max(1, Math.ceil((list()?.total ?? 0) / 15));

  return (
    <main class="space-y-5">
      <NotificationsHeader />

      <div class="relative max-w-sm">
        <Icon name="search" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          placeholder="Buscar por título..."
          value={search()}
          onInput={(e) => setSearch(e.currentTarget.value)}
          class="h-9 w-full rounded-md border border-slate-300 bg-white pl-9 pr-3 text-sm text-colpsi-text outline-none transition-colors placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
        />
      </div>

      <ErrorBoundary fallback={(err, reset) => (
        <div class="bg-white border border-colpsi-border p-8 rounded-lg text-center">
          <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-red-50 text-colpsi-red mb-4">
            <Icon name="bell" class="w-6 h-6" />
          </span>
          <h2 class="text-lg font-semibold text-colpsi-text mb-2">Error de Conexión</h2>
          <p class="text-colpsi-muted text-sm mb-6 max-w-lg mx-auto">{err.toString()}</p>
          <button onClick={reset} class="inline-flex items-center gap-2 h-9 px-4 rounded-md bg-colpsi-red text-white font-semibold hover:opacity-90 transition-all text-sm">
            <Icon name="refresh" /> Intentar de nuevo
          </button>
        </div>
      )}>
        <Suspense fallback={
          <div class="space-y-2">
            <For each={[1, 2, 3, 4]}>{() => <div class="h-20 bg-white animate-pulse rounded-lg border border-colpsi-border" />}</For>
          </div>
        }>
          <Show when={!list.loading && items().length === 0}>
            <div class="border border-colpsi-border rounded-lg bg-white p-12 text-center">
              <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
                <Icon name="bell" class="w-6 h-6" />
              </span>
              <h3 class="font-semibold text-colpsi-text mb-1">Sin notificaciones</h3>
              <p class="text-sm text-colpsi-muted">Crea tu primera notificación para los agremiados.</p>
            </div>
          </Show>

          <div class="space-y-2">
            <For each={filtered()}>
              {(n) => (
                <A
                  href={`/admin/notificaciones/${n.id}`}
                  class="block bg-white rounded-lg border border-colpsi-border hover:border-colpsi-blue/40 transition-colors p-4"
                >
                  <div class="flex items-start justify-between gap-4">
                    <div class="min-w-0">
                      <div class="flex items-center gap-2 flex-wrap">
                        <h3 class="font-semibold text-colpsi-text truncate">{n.title}</h3>
                        <StatusBadge status={n.status} />
                        <span class="text-[10px] font-medium uppercase tracking-wide text-colpsi-muted bg-colpsi-bg rounded px-2 py-0.5">
                          {targetTypeLabel(n.target_type)}
                        </span>
                      </div>
                      <p class="text-sm text-colpsi-muted line-clamp-2 mt-1">{n.message}</p>
                      <div class="flex items-center gap-3 mt-2 text-[11px] text-colpsi-muted font-medium">
                        <span class="inline-flex items-center gap-1"><Icon name="clock" class="w-3.5 h-3.5" /> {formatNotifDate(n.scheduled_at || n.sent_at || n.created_at)}</span>
                        {n.send_email && <span class="inline-flex items-center gap-1"><Icon name="mail" class="w-3.5 h-3.5" /> Email</span>}
                      </div>
                    </div>
                    <Show when={n.status === "pending"}>
                      <button
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          setCancelId(n.id);
                        }}
                        disabled={busy() === n.id}
                        class="shrink-0 inline-flex items-center gap-1 h-7 px-2 rounded-md text-xs font-medium text-colpsi-red hover:bg-red-50 transition-colors disabled:opacity-50"
                      >
                        <Icon name="trash" class="w-3.5 h-3.5" />
                        Cancelar
                      </button>
                    </Show>
                  </div>
                </A>
              )}
            </For>
          </div>

          <Show when={(list()?.total ?? 0) > 15}>
            <div class="flex items-center justify-center gap-4 mt-6">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page() <= 1}
                class="h-8 px-3 bg-white border border-colpsi-border rounded-md text-xs font-medium text-colpsi-text hover:bg-colpsi-bg hover:text-colpsi-blue disabled:opacity-40 transition-colors"
              >
                ← Anterior
              </button>
              <span class="text-sm text-colpsi-muted font-medium">Página {page()} de {totalPages}</span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page() >= totalPages}
                class="h-8 px-3 bg-white border border-colpsi-border rounded-md text-xs font-medium text-colpsi-text hover:bg-colpsi-bg hover:text-colpsi-blue disabled:opacity-40 transition-colors"
              >
                Siguiente →
              </button>
            </div>
          </Show>
        </Suspense>
      </ErrorBoundary>

      <Show when={!!cancelId()}>
        <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm">
          <div class="bg-white rounded-lg p-6 max-w-sm w-full shadow-lg border border-colpsi-border">
            <span class="inline-flex h-10 w-10 items-center justify-center rounded-md bg-amber-50 text-amber-600 mb-3">
              <Icon name="clock" class="w-5 h-5" />
            </span>
            <h3 class="text-base font-semibold text-colpsi-text mb-2">¿Cancelar notificación?</h3>
            <p class="text-sm text-colpsi-muted mb-5">Solo se pueden cancelar notificaciones programadas aún pendientes.</p>
            <div class="flex gap-2">
              <button onClick={() => setCancelId(null)} class="flex-1 h-9 rounded-md border border-colpsi-border bg-white text-colpsi-text font-medium hover:bg-colpsi-bg transition-colors">
                No
              </button>
              <button onClick={handleCancel} disabled={!!busy()} class="flex-1 h-9 rounded-md bg-colpsi-red text-white font-semibold hover:opacity-90 transition-colors disabled:opacity-50">
                {busy() ? "..." : "Sí, cancelar"}
              </button>
            </div>
          </div>
        </div>
      </Show>
    </main>
  );
}
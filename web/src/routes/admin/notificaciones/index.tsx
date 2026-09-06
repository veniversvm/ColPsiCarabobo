// web/src/routes/admin/notificaciones/index.tsx
import { createResource, createSignal, For, Show, Suspense, ErrorBoundary } from "solid-js";
import { A, action, useAction } from "@solidjs/router";
import { apiDelete, apiGet } from "~/lib/api";
import { PaginatedResponse } from "~/types/admin";
import { Notification } from "~/types/notifications";
import { NotificationsHeader, StatusBadge, formatNotifDate, targetTypeLabel } from "~/components/admin/notificaciones";

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
    <main class="pb-20 animate-in fade-in duration-500">
      <NotificationsHeader />

      <div class="mb-4">
        <input
          placeholder="Buscar por título..."
          value={search()}
          onInput={(e) => setSearch(e.currentTarget.value)}
          class="w-full sm:max-w-sm bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm"
        />
      </div>

      <ErrorBoundary fallback={(err, reset) => (
        <div class="bg-red-50 border border-red-200 p-8 rounded-3xl text-center">
          <p class="text-4xl mb-4">🚨</p>
          <h2 class="text-xl font-black text-red-800 mb-2">Error de Conexión</h2>
          <p class="text-red-600 text-sm mb-6 max-w-lg mx-auto">{err.toString()}</p>
          <button onClick={reset} class="bg-red-600 text-white font-black px-6 py-2.5 rounded-xl hover:bg-red-700 active:scale-95 transition-all text-sm">
            ↻ Intentar de nuevo
          </button>
        </div>
      )}>
        <Suspense fallback={
          <div class="space-y-3">
            <For each={[1, 2, 3, 4]}>{() => <div class="h-24 bg-white animate-pulse rounded-2xl border border-colpsi-border" />}</For>
          </div>
        }>
          <Show when={!list.loading && items().length === 0}>
            <div class="bg-white rounded-3xl border border-colpsi-border p-12 text-center">
              <p class="text-5xl mb-4">🔔</p>
              <h3 class="text-lg font-black text-gray-700 mb-1">Sin notificaciones</h3>
              <p class="text-sm text-gray-500">Crea tu primera notificación para los agremiados.</p>
            </div>
          </Show>

          <div class="space-y-3">
            <For each={filtered()}>
              {(n) => (
                <A
                  href={`/admin/notificaciones/${n.id}`}
                  class="block bg-white rounded-3xl border border-colpsi-border shadow-sm hover:shadow-md hover:border-blue-200 transition-all p-5"
                >
                  <div class="flex items-start justify-between gap-4">
                    <div class="min-w-0">
                      <div class="flex items-center gap-2 flex-wrap">
                        <h3 class="font-black text-gray-800 truncate">{n.title}</h3>
                        <StatusBadge status={n.status} />
                        <span class="text-[10px] font-black uppercase tracking-wider text-gray-400 bg-gray-100 rounded-full px-2 py-0.5">
                          {targetTypeLabel(n.target_type)}
                        </span>
                      </div>
                      <p class="text-sm text-gray-500 line-clamp-2 mt-1">{n.message}</p>
                      <div class="flex items-center gap-3 mt-2 text-[11px] text-gray-400 font-semibold">
                        <span>🕒 {formatNotifDate(n.scheduled_at || n.sent_at || n.created_at)}</span>
                        {n.send_email && <span>✉️ Email</span>}
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
                        class="shrink-0 text-[11px] font-black text-red-500 hover:text-red-700 bg-red-50 hover:bg-red-100 px-3 py-1.5 rounded-lg transition-colors disabled:opacity-50"
                      >
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
                class="bg-white border border-gray-200 rounded-xl px-4 py-2 text-sm font-bold text-gray-600 hover:bg-colpsi-surface disabled:opacity-40 transition-colors"
              >
                ← Anterior
              </button>
              <span class="text-sm text-gray-500 font-bold">Página {page()} de {totalPages}</span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page() >= totalPages}
                class="bg-white border border-gray-200 rounded-xl px-4 py-2 text-sm font-bold text-gray-600 hover:bg-colpsi-surface disabled:opacity-40 transition-colors"
              >
                Siguiente →
              </button>
            </div>
          </Show>
        </Suspense>
      </ErrorBoundary>

      <Show when={!!cancelId()}>
        <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm">
          <div class="bg-white rounded-3xl p-6 max-w-sm w-full shadow-2xl">
            <p class="text-4xl mb-3">⏳</p>
            <h3 class="text-lg font-black text-gray-800 mb-2">¿Cancelar notificación?</h3>
            <p class="text-sm text-gray-500 mb-6">Solo se pueden cancelar notificaciones programadas aún pendientes.</p>
            <div class="flex gap-3">
              <button onClick={() => setCancelId(null)} class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-black py-2.5 rounded-xl transition-colors">
                No
              </button>
              <button onClick={handleCancel} disabled={!!busy()} class="flex-1 bg-red-600 hover:bg-red-700 text-white font-black py-2.5 rounded-xl transition-colors disabled:opacity-50">
                {busy() ? "..." : "Sí, cancelar"}
              </button>
            </div>
          </div>
        </div>
      </Show>
    </main>
  );
}

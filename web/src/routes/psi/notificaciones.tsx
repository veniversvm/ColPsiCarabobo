// web/src/routes/psi/notificaciones.tsx
import { createResource, createSignal, For, Show, Suspense, ErrorBoundary } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Notification, UnreadCountResponse } from "~/types/notifications";

export default function PsiNotificaciones() {
  const [unread] = createResource(() => apiGet<UnreadCountResponse>("/notifications/psi-user/unread-count"));
  const [list, { refetch }] = createResource(
    () => apiGet<{ data: Notification[]; total: number; page: number }>("/notifications/psi-user?page=1&limit=50"),
    { initialValue: { data: [], total: 0, page: 1 } }
  );

  // Marcar como leída: GET del detalle la marca en el backend.
  const [opened, setOpened] = createSignal<string | null>(null);
  const [openError, setOpenError] = createSignal("");

  const openDetail = async (n: Notification) => {
    setOpenError("");
    setOpened(null);
    try {
      await apiGet(`/notifications/psi-user/${n.id}`);
      setOpened(n.id);
      refetch();
      unread.refetch();
    } catch (e: any) {
      setOpenError(getUserFacingError(e));
    }
  };

  const items = () => list()?.data ?? [];

  return (
    <main class="bg-[#f8fafc] min-h-screen pb-24">
      <div class="bg-[#1e3a8a] pt-12 pb-20 px-6">
        <div class="max-w-3xl mx-auto">
          <A href="/psi" class="inline-flex items-center gap-1 text-blue-200 text-sm font-bold mb-4 hover:text-white">← Volver al Panel</A>
          <h1 class="text-white text-2xl font-bold flex items-center gap-3">
            🔔 Notificaciones
            <Show when={(unread()?.unread_count ?? 0) > 0}>
              <span class="bg-colpsi-yellow text-[#1e3a8a] text-xs font-black px-3 py-1 rounded-full">{unread()?.unread_count} nuevas</span>
            </Show>
          </h1>
          <p class="text-blue-200 text-sm mt-1">Comunicados y avisos del colegio</p>
        </div>
      </div>

      <div class="max-w-3xl mx-auto px-4 -mt-12 space-y-3">
        <ErrorBoundary fallback={(err) => (
          <div class="bg-white rounded-3xl p-8 text-center shadow-sm border border-gray-100">
            <p class="text-3xl mb-3">🚨</p>
            <p class="text-red-600 font-bold text-sm">{getUserFacingError(err)}</p>
          </div>
        )}>
          <Suspense fallback={
            <div class="space-y-3">
              <For each={[1, 2, 3]}>{() => <div class="h-24 bg-white animate-pulse rounded-3xl border border-gray-100" />}</For>
            </div>
          }>
            <Show when={!list.loading && items().length === 0}>
              <div class="bg-white rounded-3xl p-12 text-center shadow-sm border border-gray-100">
                <p class="text-5xl mb-4">📭</p>
                <h3 class="font-black text-gray-700">Sin notificaciones</h3>
                <p class="text-sm text-gray-500 mt-1">No tienes comunicados pendientes.</p>
              </div>
            </Show>

            <Show when={openError()}>
              <div class="bg-red-50 border border-red-200 text-red-700 text-sm font-semibold px-4 py-3 rounded-2xl">{openError()}</div>
            </Show>

            <For each={items()}>
              {(n) => {
                const isTarget = n.targets && n.targets.length > 0 ? n.targets[0] : null;
                const isRead = isTarget?.is_read ?? true;
                return (
                  <button
                    onClick={() => openDetail(n)}
                    class={`w-full text-left bg-white rounded-3xl p-5 shadow-sm border transition-all active:scale-[0.99] ${
                      isRead ? "border-gray-100" : "border-blue-300 ring-1 ring-blue-100"
                    }`}
                  >
                    <div class="flex items-start gap-3">
                      <div class={`w-2.5 h-2.5 rounded-full mt-2 shrink-0 ${isRead ? "bg-transparent" : "bg-blue-600"}`} />
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-2">
                          <h4 class={`text-sm font-bold truncate ${isRead ? "text-gray-600" : "text-gray-900"}`}>{n.title}</h4>
                        </div>
                        <p class={`text-xs mt-1 line-clamp-3 ${isRead ? "text-gray-400" : "text-gray-600"}`}>{n.message}</p>
                        <p class="text-[10px] text-gray-400 font-bold mt-2 uppercase tracking-wider">
                          {formatDate(n.sent_at || n.created_at)}
                        </p>
                      </div>
                      {isRead && opened() === n.id && <span class="text-green-600 shrink-0">✓</span>}
                    </div>
                  </button>
                );
              }}
            </For>
          </Suspense>
        </ErrorBoundary>
      </div>
    </main>
  );
}

function formatDate(value?: string | null): string {
  if (!value) return "";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString("es-VE", { year: "numeric", month: "short", day: "numeric" });
}

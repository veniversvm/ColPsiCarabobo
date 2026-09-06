// web/src/routes/psi/notificaciones.tsx
import { createResource, createSignal, For, Show, Suspense, onMount } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet, apiPatch } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Notification, UnreadCountResponse } from "~/types/notifications";

export default function PsiNotificaciones() {
  const [unread, { refetch: refetchUnread }] = createResource(
    () => apiGet<UnreadCountResponse>("/notifications/psi-user/unread-count"),
    { initialValue: { unread_count: 0 } }
  );

  // Primera página vía resource (SSR incluye el listado inicial); las páginas
  // siguientes se cargan con keyset pagination (cursor) sin recargar la ruta.
  const LIST_LIMIT = 20;
  const [list] = createResource(
    () => apiGet<{ data: Notification[]; total: number; page: number; next_cursor?: string | null }>(
      `/notifications/psi-user?limit=${LIST_LIMIT}`
    ),
    { initialValue: { data: [], total: 0, page: 1 } }
  );
  const [extra, setExtra] = createSignal<Notification[]>([]);
  const [nextCursor, setNextCursor] = createSignal<string | null>(null);
  const [listLoading, setListLoading] = createSignal(false);
  const [listDone, setListDone] = createSignal(false);

  onMount(() => setNextCursor(list()?.next_cursor ?? null));

  // Id de la notificación expandida (el mensaje se lee del propio item del listado,
  // así la apertura es inmediata y sin parpadeo: no espera a la red ni re-renderiza la lista).
  const [opened, setOpened] = createSignal<string | null>(null);
  const [openError, setOpenError] = createSignal("");
  // Notificaciones marcadas como leídas en esta sesión (reemplaza el auto-marcar al abrir).
  const [readIds, setReadIds] = createSignal<Set<string>>(new Set());

  // Abrir/cerrar (toggle) sin petición: leer la notificación NO la marca como leída.
  // El psicólogo debe pulsar "Marcar como leída" dentro del panel abierto.
  const toggleDetail = (n: Notification) => {
    setOpenError("");
    setOpened(opened() === n.id ? null : n.id);
  };

  const markRead = async (n: Notification) => {
    setOpenError("");
    // Optimista: refleja "leída" al instante sin esperar red ni recargar.
    setReadIds((prev) => {
      const next = new Set(prev);
      next.add(n.id);
      return next;
    });
    try {
      await apiPatch(`/notifications/psi-user/${n.id}/read`, {});
    } catch (e) {
      // Revertir el estado local si el servidor no lo confirmó.
      setReadIds((prev) => {
        const next = new Set(prev);
        next.delete(n.id);
        return next;
      });
      setOpenError(getUserFacingError(e));
      return;
    }
    try { await refetchUnread(); } catch (e) { /* eslint-disable-next-line no-console */ console.error(e); }
  };

  const loadMore = async () => {
    if (listLoading() || nextCursor() === null) return;
    setListLoading(true);
    try {
      const params = new URLSearchParams({ limit: String(LIST_LIMIT), cursor: nextCursor()! });
      const data = await apiGet<{ data: Notification[]; next_cursor?: string | null }>(
        `/notifications/psi-user?${params.toString()}`
      );
      const result = data?.data ?? [];
      setExtra((prev) => [...prev, ...result]);
      setNextCursor(data?.next_cursor ?? null);
      if (result.length < LIST_LIMIT) setListDone(true);
    } catch (e) {
      setOpenError(getUserFacingError(e));
    } finally {
      setListLoading(false);
    }
  };

  const items = () => [...(list()?.data ?? []), ...extra()];

  const isRead = (n: Notification) => readIds().has(n.id) || n.targets?.[0]?.is_read === true;

  return (
    <main class="bg-colpsi-bg min-h-screen pb-24">
      <div class="bg-heraldic pt-12 pb-20 px-6 shadow-inner">
        <div class="max-w-3xl mx-auto">
          <A href="/psi" class="inline-flex items-center gap-1 text-blue-200 text-sm font-bold mb-4 hover:text-white">← Volver al Panel</A>
          <h1 class="text-white text-2xl font-bold flex items-center gap-3">
            🔔 Notificaciones
            <Show when={(unread()?.unread_count ?? 0) > 0}>
              <span class="bg-colpsi-yellow text-colpsi-blue text-xs font-black px-3 py-1 rounded-full">{unread()?.unread_count} nuevas</span>
            </Show>
          </h1>
          <p class="text-blue-200 text-sm mt-1">Comunicados y avisos del colegio</p>
        </div>
      </div>

      <div class="max-w-3xl mx-auto px-4 -mt-12 space-y-3">
        <Suspense fallback={
          <div class="space-y-3">
            <For each={[1, 2, 3]}>{() => <div class="h-24 bg-white animate-pulse rounded-3xl border border-colpsi-border" />}</For>
          </div>
        }>
          <Show when={!list.loading && items().length === 0}>
            <div class="bg-white rounded-3xl p-12 text-center shadow-sm border border-colpsi-border">
              <p class="text-5xl mb-4">📭</p>
              <h3 class="font-black text-gray-700">Sin notificaciones</h3>
              <p class="text-sm text-colpsi-muted mt-1">No tienes comunicados pendientes.</p>
            </div>
          </Show>

          <Show when={openError()}>
            <div class="bg-red-50 border border-red-200 text-red-700 text-sm font-semibold px-4 py-3 rounded-2xl">{openError()}</div>
          </Show>

          <For each={items()}>
            {(n) => {
              const read = isRead(n);
              const expanded = opened() === n.id;
              return (
                <div>
                  <div
                    role="button"
                    tabIndex={0}
                    onClick={() => toggleDetail(n)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        toggleDetail(n);
                      }
                    }}
                    class={`w-full text-left bg-white rounded-3xl p-5 shadow-sm border transition-all active:scale-[0.99] cursor-pointer select-none ${
                      read ? "border-colpsi-border" : "border-blue-300 ring-1 ring-blue-100"
                    } ${expanded ? "rounded-b-none border-b-0" : ""}`}
                  >
                    <div class="flex items-start gap-3">
                      <div class={`w-2.5 h-2.5 rounded-full mt-2 shrink-0 ${read ? "bg-transparent" : "bg-blue-600"}`} />
                      <div class="min-w-0 flex-1">
                        <div class="flex flex-wrap items-center gap-2">
                          <h4 class={`flex-1 min-w-0 text-sm font-bold truncate ${read ? "text-gray-600" : "text-colpsi-text"}`}>{n.title}</h4>
                          {read ? (
                            <span class="inline-flex items-center gap-1 text-[10px] font-black px-2 py-0.5 rounded-full bg-colpsi-surface text-colpsi-muted uppercase tracking-wider">✓ Leída</span>
                          ) : (
                            <>
                              <span class="inline-flex items-center gap-1 text-[10px] font-black px-2 py-0.5 rounded-full bg-blue-50 text-colpsi-blue uppercase tracking-wider">● No leída</span>
                              <button
                                type="button"
                                onClick={(e) => { e.stopPropagation(); markRead(n); }}
                                class="inline-flex items-center gap-1 text-[10px] font-black px-2.5 py-0.5 rounded-full bg-colpsi-blue text-white uppercase tracking-wider transition-all hover:bg-colpsi-blue-light active:scale-[0.96]"
                              >
                                ✓ Marcar como leída
                              </button>
                            </>
                          )}
                        </div>
                        <p class={`text-xs mt-1 line-clamp-3 ${read ? "text-gray-400" : "text-gray-600"}`}>{n.message}</p>
                        <p class="text-[10px] text-gray-400 font-bold mt-2 uppercase tracking-wider">
                          {formatDate(n.sent_at || n.created_at)}
                        </p>
                      </div>
                      {expanded && <span class="text-blue-600 shrink-0 font-black">▲</span>}
                    </div>
                  </div>
                  <Show when={expanded}>
                    <div class="bg-blue-50/60 border border-blue-100 border-t-0 rounded-b-3xl -mt-px px-5 py-4">
                      <p class="text-sm text-gray-800 whitespace-pre-wrap">{n.message}</p>
                      <Show when={!read}>
                        <button
                          onClick={() => markRead(n)}
                          class="mt-4 inline-flex items-center gap-2 px-4 py-2 rounded-2xl bg-colpsi-blue text-white text-xs font-black uppercase tracking-widest transition-all hover:bg-colpsi-blue-light active:scale-[0.98]"
                        >
                          ✓ Marcar como leída
                        </button>
                      </Show>
                    </div>
                  </Show>
                </div>
              );
            }}
          </For>

          <Show when={nextCursor() !== null && !listDone()}>
            <div class="flex justify-center pt-2">
              <button
                type="button"
                onClick={() => loadMore()}
                disabled={listLoading()}
                class="inline-flex items-center gap-2 px-5 py-2.5 rounded-full bg-colpsi-blue text-white text-xs font-black uppercase tracking-widest transition-all hover:bg-colpsi-blue-light active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
              >
                {listLoading() ? "Cargando..." : "Cargar más"}
              </button>
            </div>
          </Show>
        </Suspense>
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

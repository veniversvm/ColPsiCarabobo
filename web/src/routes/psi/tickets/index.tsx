// web/src/routes/psi/tickets/index.tsx
// Mis tickets de solicitud: lista paginada con estado, área y motivo.
import { createResource, createMemo, createSignal, For, Show, Suspense } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import type { TicketsListResponse, Ticket } from "~/types/tickets";
import { estadoColor, formatTicketDate } from "~/types/tickets";

export default function PsiMisTickets() {
  const [page, setPage] = createSignal(1);
  const [tickets] = createResource(
    () => `page=${page()}`,
    async (_k) => {
      try {
        return await apiGet<TicketsListResponse>(`/psi/tickets?page=${page()}&limit=10`);
      } catch {
        return { data: [], total: 0, page: 1, limit: 10 };
      }
    },
    { initialValue: { data: [], total: 0, page: 1, limit: 10 } }
  );

  const pendingCount = createMemo(() => (tickets()?.data ?? []).filter((t) => !t.is_closed).length);
  const totalPages = () => Math.max(1, Math.ceil((tickets()?.total ?? 0) / 10));

  return (
    <main class="bg-colpsi-bg min-h-screen pb-24">
      <div class="bg-heraldic pt-12 pb-20 px-6">
        <div class="max-w-4xl mx-auto flex items-center justify-between">
          <div>
            <A href="/psi" class="inline-flex items-center gap-1 text-blue-200 text-sm font-bold mb-4 hover:text-white">← Volver al Panel</A>
            <h1 class="text-white text-2xl font-bold flex items-center gap-3">
              🎫 Mis Solicitudes
            </h1>
            <p class="text-blue-200 text-sm mt-1">Trámites y solicitudes ante el colegio</p>
          </div>
          <A
            href="/psi/tickets/crear"
            class="inline-flex items-center gap-2 bg-colpsi-yellow text-colpsi-blue font-black px-5 py-3.5 rounded-2xl shadow-lg hover:scale-105 active:scale-95 transition-all text-sm"
          >
            <span class="text-lg leading-none">＋</span> Nueva Solicitud
          </A>
        </div>
      </div>

      <div class="max-w-4xl mx-auto px-4 -mt-12 space-y-3">
        <Suspense fallback={
          <div class="space-y-3">
            <For each={[1, 2, 3]}>{() => <div class="h-28 bg-white animate-pulse rounded-3xl border border-gray-100" />}</For>
          </div>
        }>
          <Show when={!tickets.loading && (tickets()?.data ?? []).length === 0}>
            <div class="bg-white rounded-3xl p-12 text-center shadow-sm border border-gray-100">
              <p class="text-5xl mb-4">🗂️</p>
              <h3 class="font-black text-gray-700">No tienes solicitudes</h3>
              <p class="text-sm text-gray-500 mt-1">Crea tu primera solicitud para iniciar un trámite con el colegio.</p>
              <A
                href="/psi/tickets/crear"
                class="mt-6 inline-flex items-center gap-2 bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-black px-6 py-3.5 rounded-2xl shadow-lg hover:scale-105 active:scale-95 transition-all text-sm"
              >
                <span class="text-lg leading-none">＋</span> Nueva Solicitud
              </A>
            </div>
          </Show>

          <div class="space-y-3">
            <For each={tickets()?.data ?? []}>
              {(t: Ticket) => (
                <A href={`/psi/tickets/${t.id}`} class="block bg-white rounded-3xl p-5 shadow-sm border border-gray-100 hover:border-blue-200 hover:shadow-md transition-all group">
                  <div class="flex items-start gap-4">
                    <div class={`w-11 h-11 rounded-2xl flex items-center justify-center shrink-0 font-black text-lg ${t.is_closed ? "bg-gray-100 text-gray-400" : "bg-blue-50 text-colpsi-blue group-hover:bg-colpsi-yellow transition-colors"}`}>
                      {t.is_closed ? "✓" : "🎫"}
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2 mb-1">
                        <span class="text-[10px] text-gray-400 font-black">#{t.id}</span>
                        <span class={`text-[10px] font-black px-2.5 py-0.5 rounded-full uppercase tracking-wider ${estadoColor(t.estado)}`}>
                          {t.estado?.name ?? "Sin estado"}
                        </span>
                      </div>
                      <h4 class="font-bold text-gray-800 truncate group-hover:text-colpsi-blue transition-colors">{t.title}</h4>
                      <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1.5 text-[10px] text-gray-400 font-bold uppercase tracking-wider">
                        <span>🗂️ {t.motivo?.name ?? t.motivo_id}</span>
                        <span class="w-1 h-1 bg-gray-200 rounded-full" />
                        <span>{formatTicketDate(t.created_at)}</span>
                      </div>
                    </div>
                    <span class="text-colpsi-blue opacity-0 group-hover:opacity-100 transition-opacity font-black">→</span>
                  </div>
                </A>
              )}
            </For>
          </div>

          <Show when={(tickets()?.data ?? []).length > 0}>
            <div class="flex items-center justify-between bg-white rounded-2xl px-5 py-3 shadow-sm border border-gray-100 mt-4">
              <span class="text-xs font-bold text-gray-400 uppercase tracking-widest">
                Página {page()} de {totalPages()} · {tickets()?.total ?? 0} solicitudes
              </span>
              <Show when={pendingCount() > 0}>
                <span class="text-xs font-black text-blue-600">{pendingCount()} en curso</span>
              </Show>
              <div class="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page() === 1}
                  class="px-4 py-2 bg-white border border-gray-200 rounded-xl text-xs font-black text-gray-600 hover:border-blue-300 disabled:opacity-30 transition-all"
                >
                  ← Anterior
                </button>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages(), p + 1))}
                  disabled={page() === totalPages()}
                  class="px-4 py-2 bg-white border border-gray-200 rounded-xl text-xs font-black text-gray-600 hover:border-blue-300 disabled:opacity-30 transition-all"
                >
                  Siguiente →
                </button>
              </div>
            </div>
          </Show>
        </Suspense>
      </div>
    </main>
  );
}
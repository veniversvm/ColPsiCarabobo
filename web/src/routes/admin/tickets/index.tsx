// web/src/routes/admin/tickets/index.tsx
// Cola de tickets administrativa (FIFO): filtros por motivo/estado, búsqueda y
// paginación. Los abiertos se listan por orden de llegada.
import { createResource, createMemo, createSignal, For, Show, Suspense } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import type { TicketsListResponse, Ticket, TicketMotivo } from "~/types/tickets";
import { estadoColor, formatTicketDate } from "~/types/tickets";

export default function AdminTickets() {
  const [page, setPage] = createSignal(1);
  const [q, setQ] = createSignal("");
  const [motivoId, setMotivoId] = createSignal("");
  const [estadoId, setEstadoId] = createSignal("");
  const [soloAbiertos, setSoloAbiertos] = createSignal(true);

  const [motivosConfig] = createResource(() => apiGet<{ data: TicketMotivo[] }>("/admin/tickets/motivos"), {
    initialValue: { data: [] },
  });
  const motivos = () => motivosConfig()?.data ?? [];

  // Estados disponibles del motivo seleccionado
  const estadosDisponibles = createMemo(() => {
    const m = motivos().find((mo) => String(mo.id) === motivoId());
    return m?.estados ?? [];
  });

  const [tickets] = createResource(
    () => JSON.stringify({ page: page(), q: q(), motivoId: motivoId(), estadoId: estadoId(), soloAbiertos: soloAbiertos() }),
    async (_k) => {
      const params = new URLSearchParams();
      params.set("page", String(page()));
      params.set("limit", "10");
      if (soloAbiertos()) params.set("solo_abiertos", "true");
      else params.set("solo_abiertos", "false");
      if (q()) params.set("q", q());
      if (motivoId()) params.set("motivo_id", motivoId());
      if (estadoId()) params.set("estado_id", estadoId());
      try {
        return await apiGet<TicketsListResponse>(`/admin/tickets?${params.toString()}`);
      } catch {
        return { data: [], total: 0, page: 1, limit: 10 };
      }
    },
    { initialValue: { data: [], total: 0, page: 1, limit: 10 } }
  );

  const totalPages = () => Math.max(1, Math.ceil((tickets()?.total ?? 0) / 10));
  const fluidCount = createMemo(() => (tickets()?.total ?? 0));

  const resetFilters = () => {
    setQ("");
    setMotivoId("");
    setEstadoId("");
    setSoloAbiertos(true);
    setPage(1);
  };

  return (
    <main class="space-y-5">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 class="text-2xl font-black text-gray-800 flex items-center gap-2">🎫 Tickets de Solicitudes</h1>
          <p class="text-sm text-gray-400 mt-1">Cola FIFO: las solicitudes abiertas se atienden por orden de llegada.</p>
        </div>
        <A
          href="/admin/tickets/configuracion"
          class="inline-flex items-center gap-2 bg-white border border-gray-200 px-5 py-3 rounded-2xl font-black text-gray-600 hover:border-blue-300 hover:text-colpsi-blue transition-all text-xs uppercase tracking-widest shadow-sm"
        >
          ⚙️ Configuración
        </A>
      </div>

      {/* Filtros */}
      <div class="bg-white rounded-3xl p-5 shadow-sm border border-colpsi-border space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-10 gap-3">
          <div class="md:col-span-4 relative">
            <input
              value={q()}
              onInput={(e) => setQ(e.currentTarget.value)}
              placeholder="Buscar por título o descripción..."
              class="w-full pl-10 pr-4 py-3 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
            />
            <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-gray-400 text-sm">🔍</span>
          </div>

          <select
            value={motivoId()}
            onChange={(e) => { setMotivoId(e.currentTarget.value); setEstadoId(""); }}
            class="md:col-span-3 px-4 py-3 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-700 transition-all"
          >
            <option value="">Todos los motivos</option>
            <For each={motivos()}>
              {(m) => <option value={m.id}>{m.name}</option>}
            </For>
          </select>

          <select
            value={estadoId()}
            onChange={(e) => setEstadoId(e.currentTarget.value)}
            class="md:col-span-3 px-4 py-3 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-700 transition-all"
          >
            <option value="">Todos los estados</option>
            <For each={estadosDisponibles()}>
              {(e) => <option value={e.id}>{e.name}</option>}
            </For>
          </select>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-3">
            <button
              onClick={() => { setSoloAbiertos(!soloAbiertos()); setPage(1); }}
              class={`px-4 py-2 rounded-2xl text-xs font-black uppercase tracking-widest transition-all border-2 ${
                soloAbiertos()
                  ? "bg-emerald-50 border-emerald-200 text-emerald-700"
                  : "bg-white border-gray-200 text-gray-400 hover:border-gray-300"
              }`}
            >
              {soloAbiertos() ? "✓ Solo abiertos" : "Solo abiertos"}
            </button>
            <span class="text-xs text-gray-400 font-bold">(desactivar para ver todos, incluidos cerrados)</span>
          </div>
          {(q() || motivoId() || estadoId()) && (
            <button
              onClick={resetFilters}
              class="text-xs font-black text-red-500 hover:text-red-700 uppercase tracking-widest"
            >
              ✕ Limpiar filtros
            </button>
          )}
        </div>
      </div>

      {/* Lista */}
      <Suspense fallback={
        <div class="space-y-3">
          <For each={[1, 2, 3]}>{() => <div class="h-24 bg-white animate-pulse rounded-3xl border border-colpsi-border" />}</For>
        </div>
      }>
        <Show when={!tickets.loading && (tickets()?.data ?? []).length === 0}>
          <div class="bg-white rounded-3xl p-12 text-center shadow-sm border border-colpsi-border">
            <p class="text-5xl mb-4">🗂️</p>
            <h3 class="font-black text-gray-700">Sin solicitudes que coincidan</h3>
            <p class="text-sm text-gray-500 mt-1">Ajusta los filtros o espera nuevas solicitudes de los psicólogos.</p>
          </div>
        </Show>

        <div class="space-y-3">
          <For each={tickets()?.data ?? []}>
            {(t: Ticket) => (
              <A href={`/admin/tickets/${t.id}`} class="block bg-white rounded-3xl p-5 shadow-sm border border-colpsi-border hover:border-blue-200 hover:shadow-md transition-all group">
                <div class="flex items-start gap-4">
                  <div class={`w-11 h-11 rounded-2xl flex items-center justify-center shrink-0 font-black text-lg ${t.is_closed ? "bg-gray-100 text-gray-400" : "bg-blue-50 text-colpsi-blue group-hover:bg-colpsi-yellow transition-colors"}`}>
                    {t.is_closed ? "✓" : "🎫"}
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex flex-wrap items-center gap-2 mb-1">
                      <span class="text-[10px] text-gray-400 font-black">#{t.id}</span>
                      <span class={`text-[10px] font-black px-2.5 py-0.5 rounded-full uppercase tracking-wider ${estadoColor(t.estado)}`}>
                        {t.estado?.name ?? "Sin estado"}
                      </span>
                      <span class="text-[10px] font-black px-2.5 py-0.5 rounded-full bg-gray-100 text-gray-500 uppercase tracking-wider">
                        {t.motivo?.name ?? t.motivo_id}
                      </span>
                    </div>
                    <h4 class="font-bold text-gray-800 truncate group-hover:text-colpsi-blue transition-colors">{t.title}</h4>
                    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1.5 text-[10px] text-gray-400 font-bold uppercase tracking-wider">
                      <span>👤 {[t.psi_first_name, t.psi_last_name].filter(Boolean).join(" ") || "Psicólogo/a"}</span>
                      <span class="w-1 h-1 bg-gray-200 rounded-full" />
                      <span>Recibida: {formatTicketDate(t.created_at)}</span>
                    </div>
                  </div>
                  <span class="text-colpsi-blue opacity-0 group-hover:opacity-100 transition-opacity font-black">→</span>
                </div>
              </A>
            )}
          </For>
        </div>

        <Show when={(tickets()?.data ?? []).length > 0}>
          <div class="flex items-center justify-between bg-white rounded-2xl px-5 py-3 shadow-sm border border-colpsi-border">
            <span class="text-xs font-bold text-gray-400 uppercase tracking-widest">
              Página {page()} de {totalPages()} · {fluidCount()} solicitudes
            </span>
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
    </main>
  );
}
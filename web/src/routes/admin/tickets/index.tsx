// web/src/routes/admin/tickets/index.tsx
// Cola de tickets administrativa (FIFO): filtros por motivo/estado, búsqueda y
// paginación. Los abiertos se listan por orden de llegada.
import { createResource, createMemo, createSignal, createEffect, onCleanup, For, Show, Suspense } from "solid-js";
import type { JSX } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import type { TicketsListResponse, Ticket, TicketMotivo } from "~/types/tickets";
import { estadoColor, formatTicketDate } from "~/types/tickets";

// ── Iconos inline (estilo lucide, currentColor) ───────────────────────────

const base = {
  fill: "none" as const,
  stroke: "currentColor",
  "stroke-width": 2,
  "stroke-linecap": "round" as const,
  "stroke-linejoin": "round" as const,
  viewBox: "0 0 24 24",
};

const IconChevron = (p: { class?: string }) => (
  <svg {...base} class={p.class ?? "w-4 h-4"}>
    <path d="m6 9 6 6 6-6" />
  </svg>
);

const IconSearch = (p: { class?: string }) => (
  <svg {...base} class={p.class ?? "w-4 h-4"}>
    <circle cx="11" cy="11" r="8" />
    <path d="m21 21-4.3-4.3" />
  </svg>
);

const IconSliders = (p: { class?: string }) => (
  <svg {...base} class={p.class ?? "w-4 h-4"}>
    <path d="M21 4h-7" />
    <path d="M10 4H3" />
    <path d="M21 12h-9" />
    <path d="M8 12H3" />
    <path d="M21 20h-5" />
    <path d="M12 20H3" />
    <path d="M14 2v4" />
    <path d="M8 10v4" />
    <path d="M16 18v4" />
  </svg>
);

const IconTicket = (p: { class?: string }) => (
  <svg {...base} class={p.class ?? "w-5 h-5"}>
    <path d="M2 9a3 3 0 0 1 0 6v2a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-2a3 3 0 0 1 0-6V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2Z" />
    <path d="M13 5v2" />
    <path d="M13 17v2" />
    <path d="M13 11v2" />
  </svg>
);

const IconCheckCircle = (p: { class?: string }) => (
  <svg {...base} class={p.class ?? "w-5 h-5"}>
    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
    <path d="m9 11 3 3L22 4" />
  </svg>
);

const IconUser = (p: { class?: string }) => (
  <svg {...base} class={p.class ?? "w-3.5 h-3.5"}>
    <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
    <circle cx="12" cy="7" r="4" />
  </svg>
);

const IconInbox = (p: { class?: string }) => (
  <svg {...base} class={p.class ?? "w-10 h-10"}>
    <path d="M22 12h-6l-2 3h-4l-2-3H2" />
    <path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" />
  </svg>
);

const IconArrowRight = (p: { class?: string }) => (
  <svg {...base} class={p.class ?? "w-4 h-4"}>
    <path d="M5 12h14" />
    <path d="m12 5 7 7-7 7" />
  </svg>
);

// Select de filtro con flecha propia (en vez del dropdown nativo).
function FilterSelect(props: {
  value: string;
  onChange: (v: string) => void;
  disabled?: boolean;
  placeholder: string;
  class?: string;
  children: JSX.Element;
}) {
  return (
    <div class={`relative ${props.class ?? ""}`}>
      <select
        value={props.value}
        onChange={(e) => props.onChange(e.currentTarget.value)}
        disabled={props.disabled}
        class={`w-full appearance-none cursor-pointer pl-4 pr-10 py-3 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none transition-all text-sm font-semibold focus:border-colpsi-yellow disabled:opacity-50 disabled:cursor-not-allowed ${
          props.value ? "text-colpsi-text" : "text-colpsi-muted"
        }`}
      >
        <option value="">{props.placeholder}</option>
        {props.children}
      </select>
      <span
        class={`pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 text-colpsi-blue ${props.disabled ? "opacity-50" : ""}`}
      >
        <IconChevron class="w-4 h-4" />
      </span>
    </div>
  );
}

export default function AdminTickets() {
  const [page, setPage] = createSignal(1);
  const [inputValue, setInputValue] = createSignal("");
  const [debouncedQuery, setDebouncedQuery] = createSignal("");
  const [motivoId, setMotivoId] = createSignal("");
  const [estadoId, setEstadoId] = createSignal("");
  const [soloAbiertos, setSoloAbiertos] = createSignal(true);
  const [cached, setCached] = createSignal<TicketsListResponse | undefined>(undefined);

  let debounceTimer: ReturnType<typeof setTimeout> | undefined;

  // La búsqueda se aplica 600ms después de escribir (evita refetch por tecla).
  const applySearch = (v: string) => {
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => { setDebouncedQuery(v); setPage(1); }, 600);
  };
  onCleanup(() => { if (debounceTimer) clearTimeout(debounceTimer); });

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
    () => JSON.stringify({ page: page(), q: debouncedQuery(), motivoId: motivoId(), estadoId: estadoId(), soloAbiertos: soloAbiertos() }),
    async (_k) => {
      const params = new URLSearchParams();
      params.set("page", String(page()));
      params.set("limit", "10");
      if (soloAbiertos()) params.set("solo_abiertos", "true");
      else params.set("solo_abiertos", "false");
      if (debouncedQuery()) params.set("q", debouncedQuery());
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

  // Mantener el último resultado visible mientras se refresca (evita parpadeo).
  createEffect(() => {
    const d = tickets();
    if (d) setCached(d);
  });

  const display = () => cached() ?? tickets();

  const totalPages = () => Math.max(1, Math.ceil((display()?.total ?? 0) / 10));
  const fluidCount = createMemo(() => (display()?.total ?? 0));
  const hasFilters = () => Boolean(debouncedQuery() || motivoId() || estadoId());

  const resetFilters = () => {
    if (debounceTimer) clearTimeout(debounceTimer);
    setInputValue("");
    setDebouncedQuery("");
    setMotivoId("");
    setEstadoId("");
    setSoloAbiertos(true);
    setPage(1);
  };

  return (
    <main class="space-y-6">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <div class="h-11 w-11 rounded-2xl bg-blue-50 text-colpsi-blue flex items-center justify-center shrink-0">
            <IconTicket class="w-6 h-6" />
          </div>
          <div>
            <h1 class="text-2xl font-black text-colpsi-blue flex items-center gap-2">Tickets de Solicitudes</h1>
            <p class="text-sm text-colpsi-muted mt-1">Cola FIFO: las solicitudes abiertas se atienden por orden de llegada.</p>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <span class="bg-white border border-colpsi-border px-4 py-3 rounded-2xl text-xs font-black text-colpsi-muted uppercase tracking-widest shadow-sm">
            {fluidCount()} solicitudes
          </span>
          <A
            href="/admin/tickets/configuracion"
            class="inline-flex items-center gap-2 bg-white border border-colpsi-border px-5 py-3 rounded-2xl font-black text-colpsi-muted hover:text-colpsi-blue hover:border-colpsi-blue/40 transition-all text-xs uppercase tracking-widest shadow-sm"
          >
            <IconSliders class="w-4 h-4" />
            Configuración
          </A>
        </div>
      </div>

      {/* Filtros */}
      <div class="bg-white rounded-3xl p-5 shadow-sm border border-colpsi-border space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-[11px] uppercase tracking-widest font-black text-colpsi-blue">Filtros</span>
          {hasFilters() && (
            <button
              onClick={resetFilters}
              class="inline-flex items-center gap-1 text-xs font-black text-colpsi-red hover:text-red-700 uppercase tracking-widest transition-all"
            >
              ✕ Limpiar filtros
            </button>
          )}
        </div>

        <div class="grid grid-cols-1 md:grid-cols-10 gap-3">
          <div class="md:col-span-4 relative">
            <input
              value={inputValue()}
              onInput={(e) => {
                const v = e.currentTarget.value;
                setInputValue(v);
                applySearch(v);
              }}
              placeholder="Buscar por título o descripción..."
              class="w-full pl-10 pr-4 py-3 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-colpsi-text transition-all"
            />
            <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-colpsi-muted">
              <IconSearch class="w-4 h-4" />
            </span>
          </div>

          <FilterSelect
            class="md:col-span-3"
            value={motivoId()}
            onChange={(v) => { setMotivoId(v); setEstadoId(""); setPage(1); }}
            placeholder="Todos los motivos"
          >
            <For each={motivos()}>
              {(m) => <option value={m.id}>{m.name}</option>}
            </For>
          </FilterSelect>

          <FilterSelect
            class="md:col-span-3"
            value={estadoId()}
            disabled={!motivoId()}
            onChange={(v) => {
              setEstadoId(v);
              if (v) {
                const sel = estadosDisponibles().find((x) => String(x.id) === String(v));
                if (sel?.is_closed) setSoloAbiertos(false);
              }
              setPage(1);
            }}
            placeholder={motivoId() ? "Todos los estados" : "Selecciona un motivo"}
          >
            <For each={estadosDisponibles()}>
              {(e) => <option value={e.id}>{e.name}</option>}
            </For>
          </FilterSelect>
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <div class="flex gap-1 bg-colpsi-surface p-1 rounded-2xl border border-colpsi-border">
            <button
              onClick={() => { setSoloAbiertos(true); setPage(1); }}
              class={`px-4 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all ${
                soloAbiertos()
                  ? "bg-colpsi-blue text-white shadow"
                  : "text-colpsi-muted hover:bg-white hover:text-colpsi-blue"
              }`}
            >
              Abiertos
            </button>
            <button
              onClick={() => { setSoloAbiertos(false); setPage(1); }}
              class={`px-4 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all ${
                soloAbiertos()
                  ? "text-colpsi-muted hover:bg-white hover:text-colpsi-blue"
                  : "bg-colpsi-blue text-white shadow"
              }`}
            >
              Todos
            </button>
          </div>
          <span class="text-xs text-colpsi-muted font-bold">Incluye cerrados para ver el historial completo.</span>
        </div>
      </div>

      {/* Indicador de actualización */}
      <Show when={tickets.loading}>
        <div class="flex items-center justify-end gap-2">
          <div class="animate-spin h-3.5 w-3.5 border-2 border-colpsi-yellow border-t-transparent rounded-full" />
          <span class="text-[10px] font-black uppercase tracking-widest text-colpsi-muted">Actualizando…</span>
        </div>
      </Show>

      {/* Lista */}
      <Suspense fallback={
        <div class="space-y-3">
          <For each={[1, 2, 3]}>{() => <div class="h-24 bg-white animate-pulse rounded-3xl border border-colpsi-border" />}</For>
        </div>
      }>
        <Show when={!tickets.loading && (display()?.data ?? []).length === 0}>
          <div class="bg-white rounded-3xl p-12 text-center shadow-sm border border-colpsi-border">
            <div class="h-20 w-20 mx-auto rounded-2xl bg-colpsi-surface text-colpsi-muted flex items-center justify-center mb-4">
              <IconInbox class="w-10 h-10" />
            </div>
            <h3 class="font-black text-colpsi-text">Sin solicitudes que coincidan</h3>
            <p class="text-sm text-colpsi-muted mt-1">Ajusta los filtros o espera nuevas solicitudes de los psicólogos.</p>
          </div>
        </Show>

        {/* La lista se anima una sola vez al montar (los refetch no re-animan). */}
        <div class="space-y-3 animate-in fade-in duration-300">
          <For each={display()?.data ?? []}>
            {(t: Ticket) => (
              <A
                href={`/admin/tickets/${t.id}`}
                class={`block bg-white rounded-3xl p-5 shadow-sm border border-colpsi-border hover:shadow-md active:scale-[0.99] transition-all group border-l-4 focus-visible:ring-2 focus-visible:ring-colpsi-blue/30 outline-none ${
                  t.is_closed ? "border-l-red-300" : "border-l-colpsi-blue"
                }`}
              >
                <div class="flex items-start gap-4">
                  <div class={`w-11 h-11 rounded-2xl flex items-center justify-center shrink-0 ${t.is_closed ? "bg-gray-100 text-gray-400" : "bg-blue-50 text-colpsi-blue group-hover:bg-colpsi-yellow group-hover:text-colpsi-blue-dark transition-colors"}`}>
                    {t.is_closed ? <IconCheckCircle class="w-5 h-5" /> : <IconTicket class="w-5 h-5" />}
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex flex-wrap items-center gap-2 mb-1">
                      <span class="text-[10px] font-black text-colpsi-blue">#{t.id}</span>
                      <span class={`text-[10px] font-black px-2.5 py-0.5 rounded-full uppercase tracking-wider ${estadoColor(t.estado)}`}>
                        {t.estado?.name ?? "Sin estado"}
                      </span>
                      <span class="text-[10px] font-black px-2.5 py-0.5 rounded-full bg-colpsi-surface text-colpsi-muted uppercase tracking-wider">
                        {t.motivo?.name ?? t.motivo_id}
                      </span>
                    </div>
                    <h4 class="font-bold text-colpsi-text truncate group-hover:text-colpsi-blue transition-colors">{t.title}</h4>
                    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1.5 text-[10px] text-colpsi-muted font-bold uppercase tracking-wider">
                      <span class="inline-flex items-center gap-1">
                        <IconUser class="w-3.5 h-3.5" />
                        {[t.psi_first_name, t.psi_last_name].filter(Boolean).join(" ") || "Psicólogo/a"}
                      </span>
                      <span class="w-1 h-1 bg-gray-200 rounded-full" />
                      <span>Recibida: {formatTicketDate(t.created_at)}</span>
                    </div>
                  </div>
                  <span class="text-colpsi-blue opacity-0 group-hover:opacity-100 transition-opacity self-center">
                    <IconArrowRight class="w-4 h-4" />
                  </span>
                </div>
              </A>
            )}
          </For>
        </div>

        <Show when={(display()?.data ?? []).length > 0}>
          <div class="flex flex-wrap items-center justify-between gap-3 bg-white rounded-2xl px-5 py-3 shadow-sm border border-colpsi-border">
            <span class="text-xs font-bold text-colpsi-muted uppercase tracking-widest">
              Página {page()} de {totalPages()} <span class="text-gray-300 mx-1">·</span> {fluidCount()} solicitudes
            </span>
            <div class="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page() === 1}
                class="px-4 py-2 bg-white border border-colpsi-border rounded-xl text-xs font-black text-colpsi-muted hover:border-colpsi-blue/40 hover:text-colpsi-blue disabled:opacity-30 transition-all"
              >
                ← Anterior
              </button>
              <Show when={totalPages() <= 7}>
                <For each={Array.from({ length: totalPages() }, (_, i) => i + 1)}>
                  {(n) => (
                    <button
                      onClick={() => setPage(n)}
                      class={`w-8 h-8 rounded-xl text-xs font-black transition-all border ${
                        n === page()
                          ? "bg-colpsi-blue text-white border-colpsi-blue"
                          : "bg-white text-colpsi-muted border-colpsi-border hover:border-colpsi-blue/40 hover:text-colpsi-blue"
                      }`}
                    >
                      {n}
                    </button>
                  )}
                </For>
              </Show>
              <button
                onClick={() => setPage((p) => Math.min(totalPages(), p + 1))}
                disabled={page() === totalPages()}
                class="px-4 py-2 bg-white border border-colpsi-border rounded-xl text-xs font-black text-colpsi-muted hover:border-colpsi-blue/40 hover:text-colpsi-blue disabled:opacity-30 transition-all"
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
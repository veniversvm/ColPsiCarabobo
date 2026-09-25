// web/src/routes/admin/tickets/index.tsx
// Cola de tickets administrativa (FIFO): filtros por motivo/estado, búsqueda y
// paginación. Los abiertos se listan por orden de llegada.
import { createResource, createMemo, createSignal, createEffect, onCleanup, For, Show, Suspense } from "solid-js";
import type { JSX } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import type { TicketsListResponse, Ticket, TicketMotivo } from "~/types/tickets";
import { estadoColor, formatTicketDate } from "~/types/tickets";
import { PageHeader } from "~/components/admin/ui/PageHeader";
import { Button } from "~/components/admin/ui/Button";
import { Badge } from "~/components/admin/ui/Badge";
import { Input } from "~/components/admin/ui/Input";
import { Icon } from "~/components/admin/ui/icons";

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
        class={`h-9 w-full appearance-none cursor-pointer rounded-md border bg-white pl-3 pr-9 text-sm outline-none transition-all focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 disabled:opacity-50 disabled:cursor-not-allowed ${
          props.value ? "border-slate-300 text-colpsi-text" : "border-slate-300 text-colpsi-muted"
        }`}
      >
        <option value="">{props.placeholder}</option>
        {props.children}
      </select>
      <span
        class={`pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-colpsi-muted ${props.disabled ? "opacity-50" : ""}`}
      >
        <Icon name="chevronDown" class="w-4 h-4" />
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

  const btnPage = (active: boolean) =>
    `w-8 h-8 rounded-md text-xs font-medium transition-all border ${
      active
        ? "bg-colpsi-blue text-white border-colpsi-blue"
        : "bg-white text-colpsi-muted border-colpsi-border hover:border-colpsi-blue/40 hover:text-colpsi-blue"
    }`;

  return (
    <main class="space-y-4">
      <PageHeader
        crumbs={[{ label: "Tickets" }]}
        title="Tickets de Solicitudes"
        description="Cola FIFO: las solicitudes abiertas se atienden por orden de llegada."
        actions={
          <div class="flex items-center gap-2">
            <Badge tone="neutral" dot={false} class="h-8">{fluidCount()} solicitudes</Badge>
            <A href="/admin/tickets/configuracion">
              <Button variant="secondary" size="md">
                <Icon name="sliders" class="w-4 h-4" />
                Configuración
              </Button>
            </A>
          </div>
        }
      />

      {/* Filtros */}
      <div class="border border-colpsi-border rounded-lg bg-white p-3 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted">Filtros</span>
          {hasFilters() && (
            <button
              onClick={resetFilters}
              class="inline-flex items-center gap-1 text-xs font-medium text-colpsi-red hover:text-red-700 transition-all"
            >
              <Icon name="trash" class="w-3.5 h-3.5" />
              Limpiar filtros
            </button>
          )}
        </div>

        <div class="grid grid-cols-1 md:grid-cols-10 gap-2">
          <div class="md:col-span-4 relative">
            <Icon name="search" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <Input
              value={inputValue()}
              onInput={(e) => {
                const v = e.currentTarget.value;
                setInputValue(v);
                applySearch(v);
              }}
              placeholder="Buscar por título o descripción..."
              class="pl-9"
            />
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
          <div class="inline-flex gap-1 p-1 rounded-md bg-colpsi-bg border border-colpsi-border">
            <button
              onClick={() => { setSoloAbiertos(true); setPage(1); }}
              class={`h-8 px-3 rounded-md text-xs font-medium transition-all border ${
                soloAbiertos()
                  ? "bg-white text-colpsi-blue border-colpsi-border shadow-sm"
                  : "bg-transparent text-colpsi-muted border-transparent hover:text-colpsi-blue hover:bg-white/60"
              }`}
            >
              Abiertos
            </button>
            <button
              onClick={() => { setSoloAbiertos(false); setPage(1); }}
              class={`h-8 px-3 rounded-md text-xs font-medium transition-all border ${
                soloAbiertos()
                  ? "bg-transparent text-colpsi-muted border-transparent hover:text-colpsi-blue hover:bg-white/60"
                  : "bg-white text-colpsi-blue border-colpsi-border shadow-sm"
              }`}
            >
              Todos
            </button>
          </div>
          <span class="text-xs text-colpsi-muted font-medium">Incluye cerrados para ver el historial completo.</span>
        </div>
      </div>

      {/* Indicador de actualización */}
      <Show when={tickets.loading}>
        <div class="flex items-center justify-end gap-2">
          <div class="animate-spin h-3.5 w-3.5 border-2 border-colpsi-yellow border-t-transparent rounded-full" />
          <span class="text-[10px] font-semibold uppercase tracking-wide text-colpsi-muted">Actualizando…</span>
        </div>
      </Show>

      {/* Lista */}
      <Suspense fallback={
        <div class="space-y-2">
          <For each={[1, 2, 3]}>{() => <div class="h-20 bg-white animate-pulse rounded-lg border border-colpsi-border" />}</For>
        </div>
      }>
        <Show when={!tickets.loading && (display()?.data ?? []).length === 0}>
          <div class="border border-colpsi-border rounded-lg bg-white p-12 text-center">
            <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
              <Icon name="inbox" class="w-6 h-6" />
            </span>
            <h3 class="font-semibold text-colpsi-text">Sin solicitudes que coincidan</h3>
            <p class="text-sm text-colpsi-muted mt-1">Ajusta los filtros o espera nuevas solicitudes de los psicólogos.</p>
          </div>
        </Show>

        {/* La lista se anima una sola vez al montar (los refetch no re-animan). */}
        <div class="space-y-2 animate-in fade-in duration-300">
          <For each={display()?.data ?? []}>
            {(t: Ticket) => (
              <A
                href={`/admin/tickets/${t.id}`}
                class={`block bg-white rounded-lg p-4 border transition-colors group focus-visible:ring-2 focus-visible:ring-colpsi-blue/30 outline-none ${
                  t.is_closed ? "border-colpsi-border opacity-70 hover:opacity-100" : "border-colpsi-border hover:border-colpsi-blue/40"
                }`}
              >
                <div class="flex items-start gap-3">
                  <div class={`w-8 h-8 rounded-md flex items-center justify-center shrink-0 ${t.is_closed ? "bg-colpsi-bg text-colpsi-muted" : "bg-colpsi-blue/10 text-colpsi-blue"}`}>
                    {t.is_closed ? <Icon name="checkCircle" class="w-4 h-4" /> : <Icon name="ticket" class="w-4 h-4" />}
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex flex-wrap items-center gap-2 mb-0.5">
                      <span class="text-[10px] font-semibold text-colpsi-blue">#{t.id}</span>
                      <Badge
                        dot={false}
                        class={`text-[10px] px-1.5 py-0 ${estadoColor(t.estado)}`}
                      >
                        {t.estado?.name ?? "Sin estado"}
                      </Badge>
                      <Badge tone="neutral" dot={false} class="text-[10px] px-1.5 py-0">
                        {t.motivo?.name ?? t.motivo_id}
                      </Badge>
                    </div>
                    <h4 class="font-medium text-colpsi-text truncate group-hover:text-colpsi-blue transition-colors">{t.title}</h4>
                    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1 text-[11px] text-colpsi-muted font-medium">
                      <span class="inline-flex items-center gap-1">
                        <Icon name="user" class="w-3.5 h-3.5" />
                        {[t.psi_first_name, t.psi_last_name].filter(Boolean).join(" ") || "Psicólogo/a"}
                      </span>
                      <span class="w-1 h-1 bg-slate-200 rounded-full" />
                      <span>Recibida: {formatTicketDate(t.created_at)}</span>
                    </div>
                  </div>
                  <span class="text-colpsi-blue opacity-0 group-hover:opacity-100 transition-opacity self-center">
                    <Icon name="arrowRight" class="w-4 h-4" />
                  </span>
                </div>
              </A>
            )}
          </For>
        </div>

        <Show when={(display()?.data ?? []).length > 0}>
          <div class="flex flex-wrap items-center justify-between gap-3 border border-colpsi-border rounded-lg bg-white px-4 py-2.5">
            <span class="text-xs font-medium text-colpsi-muted">
              Página {page()} de {totalPages()} <span class="text-slate-300 mx-1">·</span> {fluidCount()} solicitudes
            </span>
            <div class="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page() === 1}
                class="h-8 px-3 bg-white border border-colpsi-border rounded-md text-xs font-medium text-colpsi-text hover:bg-colpsi-bg hover:text-colpsi-blue disabled:opacity-30 transition-all"
              >
                ← Anterior
              </button>
              <Show when={totalPages() <= 7}>
                <For each={Array.from({ length: totalPages() }, (_, i) => i + 1)}>
                  {(n) => (
                    <button
                      onClick={() => setPage(n)}
                      class={btnPage(n === page())}
                    >
                      {n}
                    </button>
                  )}
                </For>
              </Show>
              <button
                onClick={() => setPage((p) => Math.min(totalPages(), p + 1))}
                disabled={page() === totalPages()}
                class="h-8 px-3 bg-white border border-colpsi-border rounded-md text-xs font-medium text-colpsi-text hover:bg-colpsi-bg hover:text-colpsi-blue disabled:opacity-30 transition-all"
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
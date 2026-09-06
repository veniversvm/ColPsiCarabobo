// web/src/routes/admin/inscripciones/index.tsx
import { createResource, createSignal, Show, For, createEffect, onCleanup } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { PaginationBar } from "~/components/ui/PaginationBar";
import type { InscriptionListResponse, InscriptionListItem } from "~/types/inscription";

const STATUS_TABS = [
  { value: "pending", label: "Pendientes" },
  { value: "approved", label: "Aprobadas" },
  { value: "rejected", label: "Rechazadas" },
  { value: "all", label: "Todas" },
];

export default function AdminInscripcionesList() {
  const navigate = useNavigate();
  const [page, setPage] = createSignal(1);
  const [limit, setLimit] = createSignal(20);
  const [status, setStatus] = createSignal("pending");
  const [inputValue, setInputValue] = createSignal("");
  const [debouncedQuery, setDebouncedQuery] = createSignal("");
  const [cached, setCached] = createSignal<InscriptionListResponse | undefined>(undefined);

  const [data, { refetch }] = createResource(
    () => ({ p: page(), l: limit(), s: status(), q: debouncedQuery() }),
    async (params) => {
      const parts = [`page=${params.p}`, `limit=${params.l}`, `status=${encodeURIComponent(params.s)}`];
      if (params.q) parts.push(`q=${encodeURIComponent(params.q)}`);
      return apiGet<InscriptionListResponse>(`/admin/inscripciones/list?${parts.join("&")}`);
    }
  );

  createEffect(() => {
    const d = data();
    if (d) setCached(d);
  });

  let timer: any;
  const handleSearch = (e: Event) => {
    const v = (e.target as HTMLInputElement).value;
    setInputValue(v);
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => { setDebouncedQuery(v); setPage(1); }, 600);
  };
  onCleanup(() => { if (timer) clearTimeout(timer); });

  const display = () => cached() ?? data();

  return (
    <div class="space-y-6 font-sans">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-2xl font-black text-colpsi-blue">Solicitudes de Inscripción</h1>
          <p class="text-sm text-gray-500 font-medium">Revisa y procesa las pre-inscripciones de nuevos profesionales.</p>
        </div>
      </div>

      {/* Tabs de estado */}
      <div class="flex flex-wrap gap-2 bg-white p-2 rounded-2xl border border-colpsi-border shadow-sm">
        <For each={STATUS_TABS}>
          {(tab) => (
            <button
              onClick={() => { setStatus(tab.value); setPage(1); }}
              class={`px-4 py-2 rounded-xl text-sm font-black transition-all ${
                status() === tab.value ? "bg-colpsi-blue text-white shadow" : "text-gray-500 hover:bg-gray-100"
              }`}
            >
              {tab.label}
            </button>
          )}
        </For>
      </div>

      {/* Búsqueda */}
      <div class="relative">
        <input
          value={inputValue()}
          onInput={handleSearch}
          placeholder="Buscar por nombre o cédula..."
          class="w-full px-5 py-3 bg-white border border-gray-200 rounded-2xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
        />
      </div>

      <div class="bg-white rounded-3xl shadow-premium border border-colpsi-border overflow-hidden">
        <PaginationBar
          page={page()}
          totalPages={display()?.total_pages ?? 1}
          limit={limit()}
          total={display()?.total ?? 0}
          onPrev={() => setPage((p) => p - 1)}
          onNext={() => setPage((p) => p + 1)}
          onLimitChange={(v) => { setLimit(v); setPage(1); }}
          isLoading={data.loading}
        />

        <div class="overflow-x-auto">
          <table class="w-full text-left">
            <thead class="bg-colpsi-surface border-b border-colpsi-border">
              <tr>
                <th class="px-5 py-3 text-xs font-black text-gray-400 uppercase tracking-widest">Cédula</th>
                <th class="px-5 py-3 text-xs font-black text-gray-400 uppercase tracking-widest">Nombre completo</th>
                <th class="px-5 py-3 text-xs font-black text-gray-400 uppercase tracking-widest">FPV</th>
                <th class="px-5 py-3 text-xs font-black text-gray-400 uppercase tracking-widest">Fecha</th>
                <th class="px-5 py-3 text-xs font-black text-gray-400 uppercase tracking-widest">Estado</th>
                <th class="px-5 py-3 text-xs font-black text-gray-400 uppercase tracking-widest">Acciones</th>
              </tr>
            </thead>
            <tbody>
              <Show
                when={!data.loading || cached()}
                fallback={
                  <tr><td colspan="6" class="p-16 text-center">
                    <div class="w-10 h-10 border-4 border-colpsi-blue border-t-transparent rounded-full animate-spin mx-auto" />
                  </td></tr>
                }
              >
                <For
                  each={display()?.items ?? []}
                  fallback={
                    <tr><td colspan="6" class="p-12 text-center text-sm font-bold text-gray-400">No hay solicitudes</td></tr>
                  }
                >
                  {(item: InscriptionListItem) => (
                    <tr class="border-b border-gray-50 hover:bg-colpsi-surface/60 transition-colors">
                      <td class="px-5 py-4 font-black text-gray-700">{item.cedula}</td>
                      <td class="px-5 py-4 font-bold text-gray-800">{item.nombres} {item.apellidos}</td>
                      <td class="px-5 py-4 text-gray-500">{item.fpv || "—"}</td>
                      <td class="px-5 py-4 text-xs text-gray-500">{new Date(item.created_at).toLocaleDateString()}</td>
                      <td class="px-5 py-4">
                        <StatusBadge status={item.status} />
                      </td>
                      <td class="px-5 py-4">
                        <button
                          onClick={() => navigate(`/admin/inscripciones/${item.id}`)}
                          class="px-4 py-1.5 bg-colpsi-blue text-white rounded-lg text-xs font-black hover:opacity-90 transition-opacity"
                        >
                          Ver
                        </button>
                      </td>
                    </tr>
                  )}
                </For>
              </Show>
            </tbody>
          </table>
        </div>

        <PaginationBar
          page={page()}
          totalPages={display()?.total_pages ?? 1}
          limit={limit()}
          total={display()?.total ?? 0}
          onPrev={() => setPage((p) => p - 1)}
          onNext={() => setPage((p) => p + 1)}
          onLimitChange={(v) => { setLimit(v); setPage(1); }}
          isLoading={data.loading}
        />
      </div>
    </div>
  );
}

function StatusBadge(props: { status: string }) {
  const map: Record<string, string> = {
    pending: "bg-amber-50 text-amber-700 border-amber-200",
    approved: "bg-green-50 text-green-700 border-green-200",
    rejected: "bg-red-50 text-red-700 border-red-200",
  };
  const label: Record<string, string> = {
    pending: "Pendiente",
    approved: "Aprobada",
    rejected: "Rechazada",
  };
  return (
    <span class={`px-3 py-1 rounded-full text-[11px] font-black border ${map[props.status] || "bg-gray-100 text-gray-700 border-gray-200"}`}>
      {label[props.status] || props.status}
    </span>
  );
}
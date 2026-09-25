// web/src/routes/admin/inscripciones/index.tsx
import { createResource, createSignal, Show, For, createEffect, onCleanup } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { PaginationBar } from "~/components/ui/PaginationBar";
import type { InscriptionListResponse, InscriptionListItem } from "~/types/inscription";
import { PageHeader } from "~/components/admin/ui/PageHeader";
import { Input } from "~/components/admin/ui/Input";
import { Badge } from "~/components/admin/ui/Badge";
import { Button } from "~/components/admin/ui/Button";
import { Icon } from "~/components/admin/ui/icons";

const STATUS_TABS = [
  { value: "pending", label: "Pendientes" },
  { value: "approved", label: "Aprobadas" },
  { value: "rejected", label: "Rechazadas" },
  { value: "all", label: "Todas" },
];

const STATUS_TONE: Record<string, "warning" | "success" | "danger"> = {
  pending: "warning",
  approved: "success",
  rejected: "danger",
};

const STATUS_LABEL: Record<string, string> = {
  pending: "Pendiente",
  approved: "Aprobada",
  rejected: "Rechazada",
};

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
    <div class="space-y-5">
      <PageHeader
        crumbs={[{ label: "Inscripciones" }]}
        title="Solicitudes de Inscripción"
        description="Revisa y procesa las pre-inscripciones de nuevos profesionales."
      />

      {/* Tabs de estado */}
      <div class="inline-flex flex-wrap gap-1 p-1 rounded-md bg-colpsi-bg border border-colpsi-border">
        <For each={STATUS_TABS}>
          {(tab) => (
            <button
              onClick={() => { setStatus(tab.value); setPage(1); }}
              class={`h-8 px-3 rounded-md text-sm font-medium transition-all ${
                status() === tab.value
                  ? "bg-white text-colpsi-blue border border-colpsi-border shadow-sm"
                  : "text-colpsi-muted hover:text-colpsi-blue hover:bg-white/60 border border-transparent"
              }`}
            >
              {tab.label}
            </button>
          )}
        </For>
      </div>

      {/* Búsqueda */}
      <div class="relative max-w-md">
        <Icon name="search" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <Input
          value={inputValue()}
          onInput={handleSearch}
          placeholder="Buscar por nombre o cédula..."
          class="pl-9"
        />
      </div>

      <div class="border border-colpsi-border rounded-lg bg-white overflow-hidden">
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
          <table class="w-full border-collapse">
            <thead>
              <tr>
                <th class="th-cell">Cédula</th>
                <th class="th-cell">Nombre completo</th>
                <th class="th-cell">FPV</th>
                <th class="th-cell">Fecha</th>
                <th class="th-cell">Estado</th>
                <th class="th-cell text-right">Acciones</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-colpsi-border">
              <Show
                when={!data.loading || cached()}
                fallback={
                  <tr><td colspan="6" class="py-16 text-center">
                    <div class="w-8 h-8 border-2 border-colpsi-blue border-t-transparent rounded-full animate-spin mx-auto" />
                  </td></tr>
                }
              >
                <For
                  each={display()?.items ?? []}
                  fallback={
                    <tr><td colspan="6" class="py-12 text-center text-sm font-medium text-colpsi-muted">No hay solicitudes</td></tr>
                  }
                >
                  {(item: InscriptionListItem) => (
                    <tr class="hover:bg-colpsi-bg/60 transition-colors group">
                      <td class="td-cell whitespace-nowrap font-medium">{item.cedula}</td>
                      <td class="td-cell min-w-[200px] font-medium">{item.nombres} {item.apellidos}</td>
                      <td class="td-cell text-colpsi-muted whitespace-nowrap">{item.fpv || "—"}</td>
                      <td class="td-cell text-colpsi-muted whitespace-nowrap">{new Date(item.created_at).toLocaleDateString()}</td>
                      <td class="td-cell">
                        <Badge tone={STATUS_TONE[item.status] ?? "neutral"}>
                          {STATUS_LABEL[item.status] ?? item.status}
                        </Badge>
                      </td>
                      <td class="td-cell text-right whitespace-nowrap">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => navigate(`/admin/inscripciones/${item.id}`)}
                          class="opacity-0 group-hover:opacity-100 focus:opacity-100"
                        >
                          Ver
                        </Button>
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
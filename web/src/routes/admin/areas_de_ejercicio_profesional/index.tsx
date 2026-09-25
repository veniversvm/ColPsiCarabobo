// web/src/routes/admin/areas_de_ejercicio_profesional/index.tsx
import { createResource, createSignal, For, Show, Suspense } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet, apiDelete, apiPatch } from "~/lib/api";
import { PageHeader } from "~/components/admin/ui/PageHeader";
import { Input } from "~/components/admin/ui/Input";
import { Button } from "~/components/admin/ui/Button";
import { Badge } from "~/components/admin/ui/Badge";
import { Icon } from "~/components/admin/ui/icons";

// ─── INTERFAZ ACTUALIZADA ─────────────────────────────────────────────────────
interface WorkArea {
  id: number;
  name: string;
  description: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  create_by: string;
  update_by: string;
}

const formatDate = (iso: string) => {
  if (!iso) return "";
  return new Date(iso).toLocaleDateString("es-VE", {
    day: "2-digit", month: "short", year: "numeric",
  });
};

export default function AdminAreasEjercicioPage() {
  const [search, setSearch] = createSignal("");
  const [filterActive, setFilterActive] = createSignal<"all" | "active" | "inactive">("all");
  const [confirmDelete, setConfirmDelete] = createSignal<number | null>(null);
  const [busy, setBusy] = createSignal<number | null>(null);

  // Llamada al catálogo maestro (Usamos el endpoint configurado en el backend)
  const [workAreas, { refetch }] = createResource(() =>
    apiGet<WorkArea[]>("/admin/specialties/all")
  );

  const list = () => {
    const data = workAreas();
    if (!data) return [];
    return Array.isArray(data) ? data : (data as any).data ?? [];
  };

  const filtered = () => {
    const q = search().toLowerCase().trim();
    return list().filter((wa: WorkArea) => {
      if (!wa) return false;
      if (filterActive() === "active" && !wa.active) return false;
      if (filterActive() === "inactive" && wa.active) return false;
      if (q && !wa.name.toLowerCase().includes(q) && !wa.description?.toLowerCase().includes(q)) return false;
      return true;
    });
  };

  const handleToggle = async (wa: WorkArea) => {
    setBusy(wa.id);
    try {
      await apiPatch(`/admin/specialties/${wa.id}`, { active: !wa.active });
      refetch();
    } catch (err: any) {
      console.error("Error al cambiar estado del área:", err);
    } finally {
      setBusy(null);
    }
  };

  const handleDelete = async (id: number) => {
    setBusy(id);
    try {
      await apiDelete(`/admin/specialties/${id}`);
      setConfirmDelete(null);
      refetch();
    } catch (err: any) {
      console.error("Error al eliminar el área:", err);
    } finally {
      setBusy(null);
    }
  };

  return (
    <main class="space-y-4 pb-12">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <PageHeader
        crumbs={[{ label: "Áreas de Ejercicio" }]}
        title="Áreas de Ejercicio Profesional"
        description="Catálogo maestro para la clasificación del desempeño de los agremiados."
        actions={
          <A href="/admin/areas_de_ejercicio_profesional/crear">
            <Button variant="primary" size="md">
              <Icon name="plus" />
              Nueva Área
            </Button>
          </A>
        }
      />

      {/* ── FILTROS ───────────────────────────────────────────────────────── */}
      <div class="flex flex-col md:flex-row gap-2">
        <div class="relative flex-1 min-w-[220px]">
          <Icon name="search" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <Input
            type="text"
            placeholder="Buscar por nombre o descripción de área..."
            value={search()}
            onInput={(e) => setSearch(e.currentTarget.value)}
            class="pl-9"
          />
        </div>

        <div class="inline-flex gap-1 p-1 rounded-md bg-colpsi-bg border border-colpsi-border self-start">
          {(["all", "active", "inactive"] as const).map((s) => (
            <button
              onClick={() => setFilterActive(s)}
              class={`h-8 px-3 rounded-md text-xs font-medium transition-all border ${
                filterActive() === s
                  ? "bg-white text-colpsi-blue border-colpsi-border shadow-sm"
                  : "bg-transparent text-colpsi-muted border-transparent hover:text-colpsi-blue hover:bg-white/60"
              }`}
            >
              {s === "all" ? "Todas" : s === "active" ? "Activas" : "Inactivas"}
            </button>
          ))}
        </div>
      </div>

      {/* ── LISTADO ───────────────────────────────────────────────────────── */}
      <Suspense fallback={
        <div class="space-y-2">
          <For each={[1, 2, 3]}>
            {() => <div class="h-24 bg-white animate-pulse rounded-lg border border-colpsi-border" />}
          </For>
        </div>
      }>
        <Show when={!workAreas.loading && list().length === 0}>
          <div class="text-center py-16 bg-white rounded-lg border border-dashed border-colpsi-border">
            <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
              <Icon name="tag" class="w-6 h-6" />
            </span>
            <p class="text-colpsi-muted font-medium">No hay áreas de ejercicio registradas</p>
            <A href="/admin/areas_de_ejercicio_profesional/crear" class="mt-3 inline-block text-colpsi-blue font-semibold text-sm hover:underline">
              Configurar primera área →
            </A>
          </div>
        </Show>

        <Show when={!workAreas.loading && list().length > 0 && filtered().length === 0}>
          <div class="text-center py-14 bg-white rounded-lg border border-colpsi-border">
            <p class="text-colpsi-muted font-medium">No se encontraron áreas con esos criterios</p>
          </div>
        </Show>

        <div class="space-y-2">
          <For each={filtered()}>
            {(area) => {
              const isBusy = () => busy() === area.id;
              return (
                <div class={`bg-white rounded-lg border p-4 flex flex-col md:flex-row md:items-center gap-4 transition-colors ${
                  area.active ? "border-colpsi-border hover:border-colpsi-blue/40" : "border-dashed border-colpsi-border opacity-60"
                }`}>
                  <div class={`w-10 h-10 rounded-md flex items-center justify-center font-bold text-base shrink-0 ${
                    area.active ? "bg-colpsi-blue/10 text-colpsi-blue" : "bg-colpsi-bg text-colpsi-muted"
                  }`}>
                    {area.name.charAt(0).toUpperCase()}
                  </div>

                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 flex-wrap">
                      <h3 class="font-semibold text-colpsi-text truncate">{area.name}</h3>
                      <Badge tone={area.active ? "success" : "neutral"}>
                        {area.active ? "Activa" : "Inactiva"}
                      </Badge>
                    </div>
                    <p class="text-sm text-colpsi-muted line-clamp-2 mt-0.5">
                      {area.description || <span class="italic text-slate-300">Sin descripción cargada en el sistema.</span>}
                    </p>
                    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-2 text-[11px] text-colpsi-muted font-medium">
                      <span><span class="text-slate-300">ID:</span> #{area.id}</span>
                      <span class="w-1 h-1 bg-slate-200 rounded-full" />
                      <span><span class="text-slate-300">Creada:</span> {formatDate(area.created_at)}</span>
                      <Show when={area.update_by}>
                        <span class="w-1 h-1 bg-slate-200 rounded-full" />
                        <span><span class="text-slate-300">Gestor:</span> {area.update_by}</span>
                      </Show>
                    </div>
                  </div>

                  <div class="flex items-center gap-1 self-start md:self-center">
                    <button
                      onClick={() => handleToggle(area)}
                      disabled={isBusy()}
                      title={area.active ? "Desactivar" : "Activar"}
                      class={`h-8 rounded-md px-2.5 text-[10px] font-semibold border transition-colors disabled:opacity-40 ${
                        area.active
                          ? "border-emerald-200 bg-white text-emerald-600 hover:bg-emerald-600 hover:text-white"
                          : "border-colpsi-border bg-white text-colpsi-muted hover:bg-slate-600 hover:text-white"
                      }`}
                    >
                      {isBusy() ? "..." : area.active ? "ON" : "OFF"}
                    </button>

                    <A
                      href={`/admin/areas_de_ejercicio_profesional/${area.id}`}
                      class="h-8 w-8 rounded-md flex items-center justify-center text-colpsi-muted border border-transparent hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors"
                      title="Editar parámetros"
                    >
                      <Icon name="pencil" class="w-4 h-4" />
                    </A>

                    <button
                      onClick={() => setConfirmDelete(area.id)}
                      disabled={isBusy()}
                      class="h-8 w-8 rounded-md flex items-center justify-center text-colpsi-muted border border-transparent hover:text-colpsi-red hover:bg-red-50 transition-colors disabled:opacity-40"
                      title="Eliminar"
                    >
                      <Icon name="trash" class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              );
            }}
          </For>
        </div>

        <Show when={list().length > 0}>
          <p class="text-center text-xs text-colpsi-muted font-medium py-1">
            Mostrando {filtered().length} de {list().length} áreas configuradas
          </p>
        </Show>
      </Suspense>

      {/* ── MODAL CONFIRMACIÓN BORRADO ─────────────────────────────────── */}
      <Show when={confirmDelete()}>
        <div
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
          onClick={(e) => { if (e.target === e.currentTarget) setConfirmDelete(null); }}
        >
          <div class="bg-white rounded-lg shadow-lg p-6 w-full max-w-sm border border-colpsi-border">
            <span class="inline-flex h-10 w-10 items-center justify-center rounded-md bg-red-50 text-colpsi-red mb-3">
              <Icon name="trash" class="w-5 h-5" />
            </span>
            <h2 class="text-base font-semibold text-colpsi-text mb-1">¿Eliminar área de ejercicio?</h2>
            <p class="text-colpsi-muted text-sm mb-5">
              Esta acción marcará el área como inactiva. Los psicólogos que la tengan asignada dejarán de mostrarla en el directorio público.
            </p>
            <div class="flex gap-2">
              <button
                onClick={() => setConfirmDelete(null)}
                class="flex-1 h-9 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors text-sm"
              >
                Cancelar
              </button>
              <button
                onClick={() => handleDelete(confirmDelete()!)}
                disabled={busy() === confirmDelete()}
                class="flex-1 h-9 rounded-md bg-colpsi-red text-white font-semibold hover:opacity-90 transition-colors text-sm disabled:opacity-60"
              >
                {busy() === confirmDelete() ? "Procesando..." : "Confirmar"}
              </button>
            </div>
          </div>
        </div>
      </Show>
    </main>
  );
}
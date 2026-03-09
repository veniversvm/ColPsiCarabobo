// routes/admin/especialidades/index.tsx
import { createResource, createSignal, For, Show, Suspense } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet, apiDelete, apiPatch } from "~/lib/api";

interface Specialty {
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

export default function AdminEspecialidadesPage() {
  const [search, setSearch] = createSignal("");
  const [filterActive, setFilterActive] = createSignal<"all" | "active" | "inactive">("all");
  const [confirmDelete, setConfirmDelete] = createSignal<number | null>(null);
  const [busy, setBusy] = createSignal<number | null>(null);

  const [specialties, { refetch }] = createResource(() =>
    apiGet<Specialty[]>("/admin/specialties/all")
  );

  const list = () => {
    const data = specialties();
    if (!data) return [];
    return Array.isArray(data) ? data : (data as any).data ?? [];
  };

  const filtered = () => {
    const q = search().toLowerCase().trim();
    return list().filter((s: Specialty) => {
      if (!s) return false;
      if (filterActive() === "active" && !s.active) return false;
      if (filterActive() === "inactive" && s.active) return false;
      if (q && !s.name.toLowerCase().includes(q) && !s.description?.toLowerCase().includes(q)) return false;
      return true;
    });
  };

  const handleToggle = async (s: Specialty) => {
    setBusy(s.id);
    try {
      await apiPatch(`/admin/specialties/${s.id}`, { active: !s.active });
      refetch();
    } catch (err: any) {
      console.error("Error toggling specialty:", err);
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
      console.error("Error deleting specialty:", err);
    } finally {
      setBusy(null);
    }
  };

  return (
    <main class="pb-20 animate-in fade-in duration-500">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
        <div>
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
            Especialidades
          </h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium">
            Catálogo de especialidades psicológicas del Colegio
          </p>
        </div>
        <A
          href="/admin/especialidades/crear"
          class="inline-flex items-center gap-2 bg-blue-800 hover:bg-blue-900 text-white font-black px-6 py-3 rounded-2xl shadow-lg hover:scale-105 active:scale-95 transition-all text-sm"
        >
          <span class="text-lg leading-none">＋</span>
          Nueva Especialidad
        </A>
      </div>

      {/* ── FILTROS ───────────────────────────────────────────────────────── */}
      <div class="flex flex-col md:flex-row gap-3 mb-6">
        <div class="relative flex-1">
          <svg class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
          </svg>
          <input
            type="text"
            placeholder="Buscar por nombre o descripción..."
            value={search()}
            onInput={(e) => setSearch(e.currentTarget.value)}
            class="w-full pl-10 pr-4 py-2.5 bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl outline-none text-sm text-gray-800 transition-all"
          />
        </div>

        <div class="flex gap-2">
          {(["all", "active", "inactive"] as const).map((s) => (
            <button
              onClick={() => setFilterActive(s)}
              class={`px-4 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide transition-all border-2 ${
                filterActive() === s
                  ? "bg-blue-800 text-white border-blue-800"
                  : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
              }`}
            >
              {s === "all" ? "Todas" : s === "active" ? "Activas" : "Inactivas"}
            </button>
          ))}
        </div>
      </div>

      {/* ── LISTADO ───────────────────────────────────────────────────────── */}
      <Suspense fallback={
        <div class="space-y-3">
          <For each={[1, 2, 3, 4, 5]}>
            {() => <div class="h-20 bg-white animate-pulse rounded-2xl border border-gray-100" />}
          </For>
        </div>
      }>
        <Show when={!specialties.loading && list().length === 0}>
          <div class="text-center py-20 bg-white rounded-3xl border border-gray-100">
            <p class="text-5xl mb-4">🏷️</p>
            <p class="text-gray-400 font-bold">No hay especialidades aún</p>
            <A href="/admin/especialidades/crear" class="mt-4 inline-block text-blue-600 font-black text-sm hover:underline">
              Crear la primera →
            </A>
          </div>
        </Show>

        <Show when={!specialties.loading && list().length > 0 && filtered().length === 0}>
          <div class="text-center py-16 bg-white rounded-3xl border border-gray-100">
            <p class="text-gray-400 font-bold">Ningún resultado para los filtros aplicados</p>
          </div>
        </Show>

        <div class="space-y-3">
          <For each={filtered()}>
            {(spec) => {
              const isBusy = () => busy() === spec.id;
              return (
                <article class={`bg-white rounded-2xl border-2 transition-all duration-200 overflow-hidden ${
                  spec.active ? "border-gray-100 hover:border-blue-100" : "border-dashed border-gray-200 opacity-70"
                }`}>
                  <div class="flex items-center gap-4 p-4 md:p-5">

                    {/* Icono */}
                    <div class={`flex-shrink-0 w-12 h-12 rounded-2xl flex items-center justify-center font-black text-lg border-2 ${
                      spec.active ? "bg-blue-50 border-blue-100 text-blue-600" : "bg-gray-50 border-gray-200 text-gray-400"
                    }`}>
                      🏷️
                    </div>

                    {/* Contenido */}
                    <div class="flex-1 min-w-0">
                      <div class="flex flex-wrap items-center gap-2 mb-1">
                        <h2 class="font-black text-gray-900 text-base">{spec.name}</h2>
                        <span class={`text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider ${
                          spec.active ? "bg-emerald-100 text-emerald-700" : "bg-gray-100 text-gray-500"
                        }`}>
                          {spec.active ? "Activa" : "Inactiva"}
                        </span>
                      </div>
                      <p class="text-gray-500 text-sm line-clamp-1">
                        {spec.description || <span class="italic text-gray-300">Sin descripción</span>}
                      </p>
                      <div class="flex items-center gap-3 mt-1.5 text-[11px] text-gray-400 font-medium">
                        <span>ID: <span class="font-bold text-gray-600">#{spec.id}</span></span>
                        <span>·</span>
                        <span>Creada {formatDate(spec.created_at)}</span>
                        <Show when={spec.update_by}>
                          <span>·</span>
                          <span>Editada por <span class="font-bold text-gray-600">{spec.update_by}</span></span>
                        </Show>
                      </div>
                    </div>

                    {/* Acciones */}
                    <div class="flex-shrink-0 flex items-center gap-2">
                      {/* Toggle activo */}
                      <button
                        onClick={() => handleToggle(spec)}
                        disabled={isBusy()}
                        title={spec.active ? "Desactivar" : "Activar"}
                        class={`w-9 h-9 rounded-xl flex items-center justify-center border-2 transition-all font-bold text-sm disabled:opacity-40 ${
                          spec.active
                            ? "border-emerald-200 bg-emerald-50 text-emerald-600 hover:bg-emerald-100"
                            : "border-gray-200 bg-gray-50 text-gray-400 hover:bg-gray-100"
                        }`}
                      >
                        {isBusy() ? "…" : spec.active ? "✓" : "○"}
                      </button>

                      {/* Editar */}
                      <A
                        href={`/admin/especialidades/${spec.id}`}
                        class="w-9 h-9 rounded-xl flex items-center justify-center border-2 border-blue-100 bg-blue-50 text-blue-600 hover:bg-blue-100 transition-all"
                        title="Editar"
                      >
                        ✏
                      </A>

                      {/* Eliminar */}
                      <button
                        onClick={() => setConfirmDelete(spec.id)}
                        disabled={isBusy()}
                        title="Eliminar"
                        class="w-9 h-9 rounded-xl flex items-center justify-center border-2 border-red-100 bg-red-50 text-red-400 hover:bg-red-100 hover:text-red-600 transition-all disabled:opacity-40"
                      >
                        🗑
                      </button>
                    </div>
                  </div>
                </article>
              );
            }}
          </For>
        </div>

        <Show when={list().length > 0}>
          <p class="text-center text-xs text-gray-400 font-bold mt-6">
            Mostrando {filtered().length} de {list().length} especialidades
          </p>
        </Show>
      </Suspense>

      {/* ── MODAL CONFIRMACIÓN BORRADO ─────────────────────────────────── */}
      <Show when={confirmDelete()}>
        <div
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
          onClick={(e) => { if (e.target === e.currentTarget) setConfirmDelete(null); }}
        >
          <div class="bg-white rounded-3xl shadow-2xl p-8 w-full max-w-sm border border-gray-100 text-center animate-in zoom-in-95 duration-200">
            <p class="text-4xl mb-4">🗑️</p>
            <h2 class="text-lg font-black text-gray-900 mb-2">¿Desactivar especialidad?</h2>
            <p class="text-gray-500 text-sm mb-6">Se realizará una eliminación lógica. La especialidad quedará inactiva.</p>
            <div class="flex gap-3">
              <button
                onClick={() => setConfirmDelete(null)}
                class="flex-1 px-4 py-3 rounded-2xl border-2 border-gray-200 font-black text-gray-600 hover:bg-gray-50 transition-all text-sm"
              >
                Cancelar
              </button>
              <button
                onClick={() => handleDelete(confirmDelete()!)}
                disabled={busy() === confirmDelete()}
                class="flex-1 px-4 py-3 rounded-2xl bg-red-600 text-white font-black hover:bg-red-700 active:scale-95 transition-all text-sm disabled:opacity-60"
              >
                {busy() === confirmDelete() ? "Eliminando..." : "Sí, desactivar"}
              </button>
            </div>
          </div>
        </div>
      </Show>

    </main>
  );
}
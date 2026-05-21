// web/src/routes/admin/areas_de_ejercicio_profesional/index.tsx
import { createResource, createSignal, For, Show, Suspense } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet, apiDelete, apiPatch } from "~/lib/api";

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
    <main class="pb-20 animate-in fade-in duration-500 font-sans">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8 bg-white p-8 rounded-[2.5rem] shadow-premium border border-gray-100">
        <div>
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
            Áreas de Ejercicio Profesional
          </h1>
          <p class="text-gray-400 text-sm mt-1 font-medium">
            Catálogo maestro para la clasificación del desempeño de los agremiados.
          </p>
        </div>
        <A
          href="/admin/areas_de_ejercicio_profesional/crear"
          class="inline-flex items-center gap-2 bg-blue-800 hover:bg-blue-900 text-white font-black px-6 py-4 rounded-2xl shadow-xl hover:scale-105 active:scale-95 transition-all text-sm uppercase tracking-wider"
        >
          <span class="text-xl leading-none">＋</span>
          Nueva Área
        </A>
      </div>

      {/* ── FILTROS ───────────────────────────────────────────────────────── */}
      <div class="flex flex-col md:flex-row gap-4 mb-8">
        <div class="relative flex-1">
          <div class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
            </svg>
          </div>
          <input
            type="text"
            placeholder="Buscar por nombre o descripción de área..."
            value={search()}
            onInput={(e) => setSearch(e.currentTarget.value)}
            class="w-full pl-12 pr-6 py-4 bg-white border-2 border-transparent focus:border-blue-500 rounded-2xl outline-none shadow-sm text-sm text-gray-800 transition-all placeholder:text-gray-400"
          />
        </div>

        <div class="flex bg-gray-100 p-1.5 rounded-2xl border border-gray-200">
          {(["all", "active", "inactive"] as const).map((s) => (
            <button
              onClick={() => setFilterActive(s)}
              class={`px-6 py-2.5 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all ${
                filterActive() === s
                  ? "bg-white text-blue-900 shadow-md"
                  : "text-gray-500 hover:text-gray-700"
              }`}
            >
              {s === "all" ? "Todas" : s === "active" ? "Activas" : "Inactivas"}
            </button>
          ))}
        </div>
      </div>

      {/* ── LISTADO ───────────────────────────────────────────────────────── */}
      <Suspense fallback={
        <div class="grid grid-cols-1 gap-4">
          <For each={[1, 2, 3]}>
            {() => <div class="h-28 bg-white animate-pulse rounded-3xl border border-gray-100" />}
          </For>
        </div>
      }>
        <Show when={!workAreas.loading && list().length === 0}>
          <div class="text-center py-24 bg-white rounded-[2.5rem] border-2 border-dashed border-gray-200">
            <p class="text-6xl mb-6">📂</p>
            <p class="text-gray-400 font-bold text-lg">No hay áreas de ejercicio registradas</p>
            <A href="/admin/areas_de_ejercicio_profesional/crear" class="mt-4 inline-block text-blue-600 font-black text-sm hover:underline uppercase tracking-widest">
              Configurar primera área →
            </A>
          </div>
        </Show>

        <Show when={!workAreas.loading && list().length > 0 && filtered().length === 0}>
          <div class="text-center py-20 bg-white rounded-[2.5rem] border border-gray-100">
            <p class="text-gray-400 font-bold">No se encontraron áreas con esos criterios</p>
          </div>
        </Show>

        <div class="grid grid-cols-1 gap-4">
          <For each={filtered()}>
            {(area) => {
              const isBusy = () => busy() === area.id;
              return (
                <article class={`bg-white rounded-3xl border-2 transition-all duration-300 ${
                  area.active ? "border-gray-50 shadow-sm hover:shadow-xl hover:border-blue-100" : "border-dashed border-gray-200 opacity-60 bg-gray-50/50"
                }`}>
                  <div class="flex flex-col md:flex-row md:items-center gap-6 p-6 md:p-8">

                    {/* Icono / Status */}
                    <div class={`w-14 h-14 rounded-2xl flex items-center justify-center font-black text-xl border-2 transition-colors ${
                      area.active ? "bg-blue-50 border-blue-100 text-blue-600" : "bg-gray-100 border-gray-200 text-gray-400"
                    }`}>
                      {area.name.charAt(0).toUpperCase()}
                    </div>

                    {/* Info */}
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-3 mb-2">
                        <h2 class="font-black text-blue-900 text-lg tracking-tight truncate">{area.name}</h2>
                        <span class={`text-[9px] font-black px-2.5 py-1 rounded-full uppercase tracking-widest ${
                          area.active ? "bg-emerald-100 text-emerald-700" : "bg-gray-200 text-gray-600"
                        }`}>
                          {area.active ? "Activa" : "Inactiva"}
                        </span>
                      </div>
                      <p class="text-gray-500 text-sm leading-relaxed mb-4">
                        {area.description || <span class="italic text-gray-300">Sin descripción cargada en el sistema.</span>}
                      </p>
                      
                      <div class="flex flex-wrap items-center gap-y-2 gap-x-4 text-[10px] text-gray-400 font-bold uppercase tracking-wider">
                        <span class="flex items-center gap-1.5">
                          <span class="text-gray-300">ID:</span> #{area.id}
                        </span>
                        <span class="w-1 h-1 bg-gray-200 rounded-full" />
                        <span class="flex items-center gap-1.5">
                           <span class="text-gray-300">Creada:</span> {formatDate(area.created_at)}
                        </span>
                        <Show when={area.update_by}>
                          <span class="w-1 h-1 bg-gray-200 rounded-full" />
                          <span class="flex items-center gap-1.5">
                            <span class="text-gray-300">Gestor:</span> {area.update_by}
                          </span>
                        </Show>
                      </div>
                    </div>

                    {/* Acciones */}
                    <div class="flex items-center gap-3 bg-gray-50/80 p-2 rounded-2xl border border-gray-100">
                      <button
                        onClick={() => handleToggle(area)}
                        disabled={isBusy()}
                        class={`w-11 h-11 rounded-xl flex items-center justify-center border-2 transition-all font-black text-sm disabled:opacity-40 ${
                          area.active
                            ? "border-emerald-200 bg-white text-emerald-600 hover:bg-emerald-600 hover:text-white"
                            : "border-gray-300 bg-white text-gray-400 hover:bg-gray-600 hover:text-white"
                        }`}
                      >
                        {isBusy() ? "..." : area.active ? "ON" : "OFF"}
                      </button>

                      <A
                        href={`/admin/areas_de_ejercicio_profesional/${area.id}`}
                        class="w-11 h-11 rounded-xl flex items-center justify-center border-2 border-blue-100 bg-white text-blue-600 hover:bg-blue-600 hover:text-white transition-all shadow-sm"
                        title="Editar parámetros"
                      >
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                          <path d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                        </svg>
                      </A>

                      <button
                        onClick={() => setConfirmDelete(area.id)}
                        disabled={isBusy()}
                        class="w-11 h-11 rounded-xl flex items-center justify-center border-2 border-red-100 bg-white text-red-500 hover:bg-red-600 hover:text-white transition-all shadow-sm disabled:opacity-40"
                      >
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                          <path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                  </div>
                </article>
              );
            }}
          </For>
        </div>

        <Show when={list().length > 0}>
          <div class="mt-10 pt-6 border-t border-gray-100 text-center">
             <p class="text-[10px] text-gray-400 font-black uppercase tracking-[0.2em]">
                Mostrando {filtered().length} de {list().length} áreas configuradas
             </p>
          </div>
        </Show>
      </Suspense>

      {/* ── MODAL CONFIRMACIÓN BORRADO ─────────────────────────────────── */}
      <Show when={confirmDelete()}>
        <div
          class="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-blue-900/40 backdrop-blur-md"
          onClick={(e) => { if (e.target === e.currentTarget) setConfirmDelete(null); }}
        >
          <div class="bg-white rounded-[2.5rem] shadow-2xl p-10 w-full max-w-md border border-gray-100 text-center animate-in zoom-in-95 duration-200">
            <div class="w-20 h-20 bg-red-50 text-red-500 rounded-3xl flex items-center justify-center mx-auto mb-6 text-4xl">
              ⚠️
            </div>
            <h2 class="text-xl font-black text-gray-900 mb-2 uppercase tracking-tight">¿Eliminar área de ejercicio?</h2>
            <p class="text-gray-500 text-sm mb-8 leading-relaxed">
              Esta acción marcará el área como inactiva. Los psicólogos que la tengan asignada dejarán de mostrarla en el directorio público.
            </p>
            <div class="flex gap-4">
              <button
                onClick={() => setConfirmDelete(null)}
                class="flex-1 px-6 py-4 rounded-2xl border-2 border-gray-100 font-black text-gray-400 hover:bg-gray-50 transition-all text-xs uppercase tracking-widest"
              >
                Cancelar
              </button>
              <button
                onClick={() => handleDelete(confirmDelete()!)}
                disabled={busy() === confirmDelete()}
                class="flex-1 px-6 py-4 rounded-2xl bg-red-600 text-white font-black hover:bg-red-700 active:scale-95 transition-all text-xs uppercase tracking-widest shadow-lg shadow-red-200 disabled:opacity-60"
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
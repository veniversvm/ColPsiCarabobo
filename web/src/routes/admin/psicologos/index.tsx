// web/src/routes/admin/psicologos/index.tsx
import { createResource, createSignal, For, Show, Suspense, onCleanup } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { PsiAdminListItem, PaginatedResponse } from "~/types/admin";
import { ImportCsvModal } from "~/components/admin/ImportCsvModal";

const PAGE_SIZE_OPTIONS = [10, 25, 50, 100];

// ── Componente de paginación reutilizable ─────────────────────────────────────
function PaginationBar(props: {
  page: number;
  totalPages: number;
  limit: number;
  total: number;
  onPrev: () => void;
  onNext: () => void;
  onLimitChange: (v: number) => void;
}) {
  return (
    <div class="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 bg-gray-50 border-t border-gray-100">
      {/* Info + selector de entradas */}
      <div class="flex items-center gap-3">
        <span class="text-xs font-bold text-gray-400 uppercase tracking-widest whitespace-nowrap">
          Página {props.page} de {props.totalPages}
          <span class="text-gray-300 mx-2">·</span>
          {props.total} registros
        </span>
        <select
          value={props.limit}
          onChange={(e) => props.onLimitChange(Number(e.currentTarget.value))}
          class="text-xs font-black text-gray-600 bg-white border border-gray-200 rounded-lg px-2 py-1.5 outline-none focus:border-blue-400 transition-all cursor-pointer"
        >
          <For each={PAGE_SIZE_OPTIONS}>
            {(size) => <option value={size}>{size} por página</option>}
          </For>
        </select>
      </div>

      {/* Botones */}
      <div class="flex gap-2">
        <button
          disabled={props.page === 1}
          onClick={props.onPrev}
          class="px-4 py-1.5 bg-white border border-gray-200 rounded-lg text-xs font-black text-gray-600 hover:border-blue-400 hover:text-blue-700 disabled:opacity-30 transition-all"
        >
          ← Anterior
        </button>

        {/* Números de página (max 5 visibles) */}
        <Show when={props.totalPages <= 7}>
          <For each={Array.from({ length: props.totalPages }, (_, i) => i + 1)}>
            {(n) => (
              <button
                onClick={() => { if (n !== props.page) props.onLimitChange === props.onLimitChange && null; }}
                class={`w-8 h-8 rounded-lg text-xs font-black transition-all border ${
                  n === props.page
                    ? "bg-blue-800 text-white border-blue-800"
                    : "bg-white text-gray-600 border-gray-200 hover:border-blue-400"
                }`}
              >
                {n}
              </button>
            )}
          </For>
        </Show>

        <button
          disabled={props.page === props.totalPages}
          onClick={props.onNext}
          class="px-4 py-1.5 bg-white border border-gray-200 rounded-lg text-xs font-black text-gray-600 hover:border-blue-400 hover:text-blue-700 disabled:opacity-30 transition-all"
        >
          Siguiente →
        </button>
      </div>
    </div>
  );
}

// ── Página principal ──────────────────────────────────────────────────────────
export default function AdminPsychologistsList() {
  const [page, setPage] = createSignal(1);
  const [limit, setLimit] = createSignal(10);
  const [inputValue, setInputValue] = createSignal("");
  const [debouncedQuery, setDebouncedQuery] = createSignal("");
  const [showImportModal, setShowImportModal] = createSignal(false);

  const [data, { refetch }] = createResource(
    () => ({ p: page(), q: debouncedQuery(), l: limit() }),
    async (params) =>
      apiGet<PaginatedResponse<PsiAdminListItem>>(
        `/admin/psi/list?page=${params.p}&limit=${params.l}&q=${encodeURIComponent(params.q)}`
      )
  );

  let timer: any;
  const handleSearchInput = (e: Event) => {
    const val = (e.target as HTMLInputElement).value;
    setInputValue(val);
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => { setDebouncedQuery(val); setPage(1); }, 600);
  };

  onCleanup(() => { if (timer) clearTimeout(timer); });

  const goToPrev = () => { setPage((p) => p - 1); window.scrollTo({ top: 0, behavior: "smooth" }); };
  const goToNext = () => { setPage((p) => p + 1); window.scrollTo({ top: 0, behavior: "smooth" }); };
  const changeLimit = (v: number) => { setLimit(v); setPage(1); };

  const paginationProps = () => ({
    page: page(),
    totalPages: data()?.total_pages ?? 1,
    limit: limit(),
    total: data()?.total ?? 0,
    onPrev: goToPrev,
    onNext: goToNext,
    onLimitChange: changeLimit,
  });

  return (
    <div class="space-y-6 animate-in fade-in duration-500">

      {/* ── HEADER ──────────────────────────────────────────────────────── */}
      <div class="flex flex-col md:flex-row justify-between md:items-end gap-4">
        <div>
          <h1 class="text-2xl font-black text-colpsi-blue">Gestión de Agremiados</h1>
          <p class="text-gray-500 text-sm mt-1">Base de datos maestra de profesionales colegiados.</p>
        </div>
        <div class="flex gap-3">
          <button
            onClick={() => setShowImportModal(true)}
            class="bg-white border-2 border-gray-200 text-gray-700 px-4 py-2.5 rounded-xl font-bold hover:bg-gray-50 transition-colors text-sm"
          >
            📥 Importar CSV
          </button>
          <A
            href="/admin/psicologos/crear"
            class="bg-colpsi-blue text-white px-4 py-2.5 rounded-xl font-bold hover:bg-blue-800 transition-colors text-sm shadow-md"
          >
            ➕ Nuevo Registro
          </A>
        </div>
      </div>

      {/* ── BÚSQUEDA ────────────────────────────────────────────────────── */}
      <div class="bg-white p-4 rounded-2xl shadow-sm border border-gray-100 flex items-center gap-3">
        <span class="text-xl ml-2">🔍</span>
        <input
          type="text"
          value={inputValue()}
          onInput={handleSearchInput}
          placeholder="Buscar por nombre, apellido, cédula o FPV..."
          class="flex-grow bg-transparent outline-none text-colpsi-text font-medium"
        />
        <Show when={data.loading}>
          <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-colpsi-yellow mr-2 flex-shrink-0" />
        </Show>
        <Show when={inputValue()}>
          <button
            onClick={() => { setInputValue(""); setDebouncedQuery(""); setPage(1); }}
            class="text-gray-400 hover:text-gray-600 font-black text-lg leading-none flex-shrink-0 mr-1"
            title="Limpiar búsqueda"
          >
            ×
          </button>
        </Show>
      </div>

      {/* ── TABLA ───────────────────────────────────────────────────────── */}
      <div class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">

        {/* Paginación ARRIBA */}
        <Show when={data() && data()!.total_pages >= 1}>
          <PaginationBar {...paginationProps()} />
        </Show>

        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-gray-50/50 border-b border-gray-100">
                <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest">Agremiado</th>
                <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest">Credenciales</th>
                <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest">Solvencia</th>
                <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest">Estatus</th>
                <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest text-right">Acciones</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
              <Suspense fallback={
                <tr>
                  <td colSpan="5" class="p-8 text-center text-gray-400 font-medium animate-pulse">
                    Cargando base de datos...
                  </td>
                </tr>
              }>
                <For
                  each={data()?.data}
                  fallback={
                    <tr>
                      <td colSpan="5" class="p-20 text-center text-gray-500 font-medium">
                        {debouncedQuery()
                          ? `Sin resultados para "${debouncedQuery()}"`
                          : "No hay registros en la base de datos."}
                      </td>
                    </tr>
                  }
                >
                  {(psi) => (
                    <tr class="hover:bg-blue-50/30 transition-colors group">
                      <td class="px-6 py-4">
                        <A href={`/admin/psicologos/${psi.id}`} class="flex items-center gap-3 group/link">
                          <div class="w-10 h-10 rounded-xl bg-colpsi-blue text-white flex items-center justify-center font-bold shadow-sm group-hover/link:bg-colpsi-yellow group-hover/link:text-colpsi-blue transition-colors flex-shrink-0">
                            {psi.first_name.charAt(0)}{psi.last_name.charAt(0)}
                          </div>
                          <div class="min-w-0">
                            <p class="font-bold text-colpsi-blue group-hover/link:underline truncate">{psi.first_name} {psi.last_name}</p>
                            <p class="text-xs text-gray-500 truncate">{psi.email}</p>
                          </div>
                        </A>
                      </td>
                      <td class="px-6 py-4">
                        <p class="text-sm font-bold text-gray-700">FPV: {psi.fpv}</p>
                        <p class="text-xs text-gray-500">CI: {psi.ci}</p>
                      </td>
                      <td class="px-6 py-4">
                        <Show
                          when={psi.solvent}
                          fallback={<span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-red-50 text-red-700 border border-red-100">Deudor</span>}
                        >
                          <span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-green-50 text-green-700 border border-green-100">Solvente</span>
                        </Show>
                      </td>
                      <td class="px-6 py-4">
                        <Show
                          when={psi.is_active}
                          fallback={<span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-gray-100 text-gray-600 border border-gray-200">Inactivo</span>}
                        >
                          <span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-blue-50 text-colpsi-blue border border-blue-100">Activo</span>
                        </Show>
                      </td>
                      <td class="px-6 py-4 text-right">
                        <A
                          href={`/admin/psicologos/${psi.id}`}
                          class="inline-flex items-center gap-2 px-3 py-1.5 bg-blue-50 text-colpsi-blue hover:bg-colpsi-yellow transition-colors rounded-lg text-xs font-bold"
                        >
                          Detalle ✏️
                        </A>
                      </td>
                    </tr>
                  )}
                </For>
              </Suspense>
            </tbody>
          </table>
        </div>

        {/* Paginación ABAJO */}
        <Show when={data() && data()!.total_pages >= 1}>
          <PaginationBar {...paginationProps()} />
        </Show>

      </div>

      {/* ── MODAL CSV ───────────────────────────────────────────────────── */}
      <Show when={showImportModal()}>
        <ImportCsvModal
          onClose={() => setShowImportModal(false)}
          onSuccess={() => { refetch(); setShowImportModal(false); }}
        />
      </Show>

    </div>
  );
}
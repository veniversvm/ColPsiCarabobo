import { createResource, createSignal, For, Show, Suspense, onCleanup } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { PsiAdminListItem, PaginatedResponse } from "~/types/admin";

export default function AdminPsychologistsList() {
  const [page, setPage] = createSignal(1);
  const [inputValue, setInputValue] = createSignal("");
  const [debouncedQuery, setDebouncedQuery] = createSignal("");

  const [data] = createResource(
    () => ({ p: page(), q: debouncedQuery() }),
    async (params) => {
      return await apiGet<PaginatedResponse<PsiAdminListItem>>(
        `/admin/psi/list?page=${params.p}&limit=10&q=${params.q}`
      );
    }
  );

  let timer: any;
  const handleSearchInput = (e: Event) => {
    const val = (e.target as HTMLInputElement).value;
    setInputValue(val);
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => {
      setDebouncedQuery(val);
      setPage(1);
    }, 800);
  };

  onCleanup(() => {
    if (timer) clearTimeout(timer);
  });

  return (
    <div class="space-y-6 animate-in fade-in duration-500">
      
      {/* HEADER Y ACCIONES */}
      <div class="flex flex-col md:flex-row justify-between md:items-end gap-4">
        <div>
          <h1 class="text-2xl font-black text-colpsi-blue">Gestión de Agremiados</h1>
          <p class="text-gray-500 text-sm mt-1">Base de datos maestra de profesionales colegiados.</p>
        </div>
        <div class="flex gap-3">
          <A href="/admin/psicologos/importar" class="bg-white border-2 border-gray-200 text-gray-700 px-4 py-2.5 rounded-xl font-bold hover:bg-gray-50 transition-colors text-sm">
            📥 Importar CSV
          </A>
          <A href="/admin/psicologos/crear" class="bg-colpsi-blue text-white px-4 py-2.5 rounded-xl font-bold hover:bg-blue-800 transition-colors text-sm shadow-md">
            ➕ Nuevo Registro
          </A>
        </div>
      </div>

      {/* BARRA DE BÚSQUEDA */}
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
           <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-colpsi-yellow mr-2"></div>
        </Show>
      </div>

      {/* TABLA DE DATOS */}
      <div class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
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
                <tr><td colSpan="5" class="p-8 text-center text-gray-400 font-medium animate-pulse">Cargando base de datos...</td></tr>
              }>
                <For each={data()?.data} fallback={
                  <tr><td colSpan="5" class="p-20 text-center text-gray-500 font-medium">No se encontraron registros que coincidan con la búsqueda.</td></tr>
                }>
                  {(psi) => (
                    <tr class="hover:bg-blue-50/30 transition-colors group">
                      <td class="px-6 py-4">
                        {/* MEJORA: El nombre y avatar ahora son un enlace al detalle */}
                        <A href={`/admin/psicologos/${psi.id}`} class="flex items-center gap-3 group/link">
                          <div class="w-10 h-10 rounded-xl bg-colpsi-blue text-white flex items-center justify-center font-bold shadow-sm group-hover/link:bg-colpsi-yellow group-hover/link:text-colpsi-blue transition-colors">
                            {psi.first_name.charAt(0)}{psi.last_name.charAt(0)}
                          </div>
                          <div>
                            <p class="font-bold text-colpsi-blue group-hover/link:underline">{psi.first_name} {psi.last_name}</p>
                            <p class="text-xs text-gray-500">{psi.email}</p>
                          </div>
                        </A>
                      </td>
                      <td class="px-6 py-4">
                        <p class="text-sm font-bold text-gray-700">FPV: {psi.fpv}</p>
                        <p class="text-xs text-gray-500">CI: {psi.ci}</p>
                      </td>
                      <td class="px-6 py-4">
                        <Show when={psi.solvent} fallback={<span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-red-50 text-red-700 border border-red-100">Deudor</span>}>
                          <span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-green-50 text-green-700 border border-green-100">Solvente</span>
                        </Show>
                      </td>
                      <td class="px-6 py-4">
                        <Show when={psi.is_active} fallback={<span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-gray-100 text-gray-600 border border-gray-200">Inactivo</span>}>
                          <span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-blue-50 text-colpsi-blue border border-blue-100">Activo</span>
                        </Show>
                      </td>
                      <td class="px-6 py-4 text-right">
                        <div class="flex items-center justify-end gap-2">
                          <A 
                            href={`/admin/psicologos/${psi.id}`} 
                            class="inline-flex items-center gap-2 px-3 py-1.5 bg-blue-50 text-colpsi-blue hover:bg-colpsi-yellow transition-colors rounded-lg text-xs font-bold"
                          >
                            <span>Detalle</span>
                            <span>✏️</span>
                          </A>
                        </div>
                      </td>
                    </tr>
                  )}
                </For>
              </Suspense>
            </tbody>
          </table>
        </div>

        {/* CONTROLES DE PAGINACIÓN */}
        <Show when={data() && data()!.total_pages > 1}>
          <div class="px-6 py-4 border-t border-gray-100 bg-gray-50 flex items-center justify-between">
            <span class="text-xs font-bold text-gray-400 uppercase tracking-widest">
              Página {data()?.page} de {data()?.total_pages}
            </span>
            <div class="flex gap-2">
              <button 
                disabled={page() === 1}
                onClick={() => { setPage(p => p - 1); window.scrollTo({top: 0, behavior: 'smooth'}); }}
                class="px-4 py-2 bg-white border border-gray-200 rounded-lg text-xs font-black text-gray-600 hover:border-colpsi-blue disabled:opacity-30 transition-all"
              >
                ← Anterior
              </button>
              <button 
                disabled={page() === data()?.total_pages}
                onClick={() => { setPage(p => p + 1); window.scrollTo({top: 0, behavior: 'smooth'}); }}
                class="px-4 py-2 bg-white border border-gray-200 rounded-lg text-xs font-black text-gray-600 hover:border-colpsi-blue disabled:opacity-30 transition-all"
              >
                Siguiente →
              </button>
            </div>
          </div>
        </Show>
      </div>
    </div>
  );
}
// web/src/routes/admin/psicologos/index.tsx
import { createResource, createSignal, onCleanup, Show, createEffect } from "solid-js";
import { apiGet } from "~/lib/api";
import { PsiAdminListItem, PaginatedResponse } from "~/types/admin";
import { ImportCsvModal } from "~/components/admin/ImportCsvModal";
import { PaginationBar } from "~/components/ui/PaginationBar";
import {
  PsychologistHeader,
  PsychologistSearchBar,
  PsychologistTable
} from "~/components/admin/psicologos";

export default function AdminPsychologistsList() {
  const [page, setPage] = createSignal(1);
  const [limit, setLimit] = createSignal(10);
  const [inputValue, setInputValue] = createSignal("");
  const [debouncedQuery, setDebouncedQuery] = createSignal("");
  const [showImportModal, setShowImportModal] = createSignal(false);
  
  // Caché para mantener los datos visibles durante la carga
  const [cachedData, setCachedData] = createSignal<PaginatedResponse<PsiAdminListItem> | undefined>(undefined);

  const [data, { refetch }] = createResource(
    () => ({ p: page(), q: debouncedQuery(), l: limit() }),
    async (params) =>
      apiGet<PaginatedResponse<PsiAdminListItem>>(
        `/admin/psi/list?page=${params.p}&limit=${params.l}&q=${encodeURIComponent(params.q)}`
      )
  );

  // Actualizar caché solo cuando hay datos nuevos
  createEffect(() => {
    const newData = data();
    if (newData) {
      setCachedData(newData);
    }
  });

  let timer: any;
  const handleSearchInput = (e: Event) => {
    const val = (e.target as HTMLInputElement).value;
    setInputValue(val);
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => { 
      setDebouncedQuery(val); 
      setPage(1); 
    }, 600);
  };

  const clearSearch = () => {
    setInputValue("");
    setDebouncedQuery("");
    setPage(1);
  };

  onCleanup(() => { if (timer) clearTimeout(timer); });

  const goToPrev = () => { 
    setPage((p) => p - 1); 
    window.scrollTo({ top: 0, behavior: "smooth" }); 
  };

  const goToNext = () => { 
    setPage((p) => p + 1); 
    window.scrollTo({ top: 0, behavior: "smooth" }); 
  };

  const changeLimit = (v: number) => { 
    setLimit(v); 
    setPage(1); 
  };

  // Usar datos cacheados para la UI
  const displayData = () => cachedData();
  
  const paginationProps = () => ({
    page: page(),
    totalPages: displayData()?.total_pages ?? 1,
    limit: limit(),
    total: displayData()?.total ?? 0,
    onPrev: goToPrev,
    onNext: goToNext,
    onLimitChange: changeLimit,
    // Indicador de carga para los botones
    isLoading: data.loading,
  });

  // Loading solo para la primera carga
  const initialLoading = () => data.loading && !cachedData();

  return (
    <div class="space-y-6 animate-in fade-in duration-500">

      {/* Header */}
      <PsychologistHeader onImportClick={() => setShowImportModal(true)} />

      {/* Búsqueda */}
      <PsychologistSearchBar
        value={inputValue()}
        onInput={handleSearchInput}
        onClear={clearSearch}
        loading={data.loading}
      />

      {/* Tabla con paginación */}
      <div class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">

        {/* Paginación superior */}
        <Show when={displayData() && displayData()!.total_pages >= 1}>
          <PaginationBar {...paginationProps()} />
        </Show>

        {/* Tabla - siempre muestra datos cacheados si existen */}
        <Show
          when={!initialLoading()}
          fallback={
            <div class="p-8 text-center text-gray-400 font-medium animate-pulse">
              Cargando base de datos...
            </div>
          }
        >
          <PsychologistTable
            data={displayData()?.data}
            loading={data.loading && !!cachedData()} // true durante refetch
            hasQuery={!!debouncedQuery()}
            query={debouncedQuery()}
          />
        </Show>

        {/* Paginación inferior */}
        <Show when={displayData() && displayData()!.total_pages >= 1}>
          <PaginationBar {...paginationProps()} />
        </Show>

      </div>

      {/* Modal CSV */}
      <Show when={showImportModal()}>
        <ImportCsvModal
          onClose={() => setShowImportModal(false)}
          onSuccess={() => { refetch(); setShowImportModal(false); }}
        />
      </Show>

    </div>
  );
}
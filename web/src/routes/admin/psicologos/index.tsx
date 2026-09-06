import { createResource, createSignal, onCleanup, Show, createEffect } from "solid-js";
import { isServer } from "solid-js/web";
import { apiGet } from "~/lib/api";
import { PsiAdminListItem, PaginatedResponse } from "~/types/admin";
// Cambiamos el nombre del componente de Importación (Asumiendo que actualizaste el componente a XLSX)
import { ImportXlsxModal } from "~/components/admin/ImportXlsxModal";
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

  // Filtros tri-estado ("" = todos) y especialidad
  const [solvent, setSolvent] = createSignal("");
  const [active, setActive] = createSignal("");
  const [gender, setGender] = createSignal("");
  const [specialty, setSpecialty] = createSignal("");

  // Catálogo de especialidades (mismo patrón que el directorio)
  const [workAreas] = createResource(
    () => !isServer,
    async (ready) => {
      if (!ready) return [];
      try { return await apiGet<any[]>("/specialties"); }
      catch { return []; }
    }
  );

  // Caché para mantener los datos visibles durante la carga (UX fluida)
  const [cachedData, setCachedData] = createSignal<PaginatedResponse<PsiAdminListItem> | undefined>(undefined);

  const [data, { refetch }] = createResource(
    () => ({
      p: page(), q: debouncedQuery(), l: limit(),
      solvent: solvent(), active: active(), gender: gender(), specialty: specialty(),
    }),
    async (params) => {
      const queryParts = [`page=${params.p}`, `limit=${params.l}`];
      if (params.q) queryParts.push(`q=${encodeURIComponent(params.q)}`);
      if (params.solvent) queryParts.push(`solvent=${params.solvent}`);
      if (params.active) queryParts.push(`active=${params.active}`);
      if (params.gender) queryParts.push(`gender=${params.gender}`);
      if (params.specialty) queryParts.push(`specialty=${params.specialty}`);
      return apiGet<PaginatedResponse<PsiAdminListItem>>(`/admin/psi/list?${queryParts.join("&")}`);
    }
  );

  // Actualizar caché solo cuando hay datos nuevos disponibles
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
    setSolvent("");
    setActive("");
    setGender("");
    setSpecialty("");
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

  const displayData = () => cachedData();
  
  const paginationProps = () => ({
    page: page(),
    totalPages: displayData()?.total_pages ?? 1,
    limit: limit(),
    total: displayData()?.total ?? 0,
    onPrev: goToPrev,
    onNext: goToNext,
    onLimitChange: changeLimit,
    isLoading: data.loading,
  });

  const initialLoading = () => data.loading && !cachedData();

  return (
    <div class="space-y-6 animate-in fade-in duration-500 font-sans">

      {/* Header - Asegúrate de que este componente reciba la prop onImportClick */}
      <PsychologistHeader 
        title="Gestión de Agremiados" 
        onImportClick={() => setShowImportModal(true)} 
      />

      {/* Barra de Búsqueda */}
      <PsychologistSearchBar
        placeholder="Buscar por Nombre, Cédula, FPV o Área de Desempeño..."
        value={inputValue()}
        onInput={handleSearchInput}
        onClear={clearSearch}
        loading={data.loading}
        solvent={solvent()}
        onSolventChange={(v) => { setSolvent(v); setPage(1); }}
        active={active()}
        onActiveChange={(v) => { setActive(v); setPage(1); }}
        gender={gender()}
        onGenderChange={(v) => { setGender(v); setPage(1); }}
        specialty={specialty()}
        onSpecialtyChange={(v) => { setSpecialty(v); setPage(1); }}
        workAreas={workAreas()}
      />

      {/* Tabla con paginación */}
      <div class="bg-white rounded-3xl shadow-premium border border-colpsi-border overflow-hidden">

        {/* Paginación superior */}
        <Show when={displayData() && displayData()!.total_pages >= 1}>
          <PaginationBar {...paginationProps()} />
        </Show>

        <Show
          when={!initialLoading()}
          fallback={
            <div class="p-20 text-center space-y-4">
              <div class="w-12 h-12 border-4 border-colpsi-blue border-t-transparent rounded-full animate-spin mx-auto" />
              <p class="text-gray-400 font-bold tracking-widest uppercase text-xs">
                Sincronizando Base de Datos...
              </p>
            </div>
          }
        >
          {/* 
            Nota: PsychologistTable debe estar actualizado internamente para 
            mostrar 'primary_work_area' en lugar de 'primary_specialty'
          */}
          <PsychologistTable
            data={displayData()?.data}
            loading={data.loading && !!cachedData()}
            hasQuery={!!debouncedQuery()}
            query={debouncedQuery()}
          />
        </Show>

        {/* Paginación inferior */}
        <Show when={displayData() && displayData()!.total_pages >= 1}>
          <PaginationBar {...paginationProps()} />
        </Show>

      </div>

      {/* Modal de Importación XLSX (Antes CSV) */}
      <Show when={showImportModal()}>
        <ImportXlsxModal
          onClose={() => setShowImportModal(false)}
          onSuccess={() => { 
            refetch(); 
            setShowImportModal(false); 
          }}
        />
      </Show>

    </div>
  );
}
// web/src/routes/directorio/index.tsx
import { createSignal, onCleanup, onMount, Suspense } from "solid-js";
import { createStore } from "solid-js/store";
import { apiGet } from "~/lib/api";
import { DirectoryPsychologist } from "~/types/psi";
import { SearchHeader } from "~/components/directory/SearchHeader";
import { FlagFooter } from "~/components/ui/FlagFooter";
import { LoadingScreen } from "~/components/ui/LoadingScreen";
import { Meta, Title } from "@solidjs/meta";
import { isServer } from "solid-js/web";
import { createResource, Show } from "solid-js";
import { clientOnly } from "@solidjs/start";

const ResultsGrid = clientOnly(
  () => import("~/components/directory/ResultsGrid")
    .then(m => ({ default: m.ResultsGrid }))
);

interface DirectoryResponse {
  data: DirectoryPsychologist[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

const LIMIT = 12; 
export const ssr = false;

// ── ESTADO GLOBAL (Caché en Memoria) ──────────────────────────────────────
const [query,      setQuery]      = createSignal("");
const [workArea,   setWorkArea]   = createSignal(""); 
const [location,   setLocation]   = createSignal("");
const [searchParams, setSearchParams] = createSignal({ q: "", area: "", loc: "" });

const [page,        setPage]        = createSignal(1);
const [total,       setTotal]       = createSignal(0);
const [allItems,    setAllItems]    = createStore<DirectoryPsychologist[]>([]);
const [hasMore,     setHasMore]     = createSignal(true);
const [isCached,    setIsCached]    = createSignal(false);
// ─────────────────────────────────────────────────────────────────────────

export default function DirectoryPage() {
  const [loading, setLoading] = createSignal(false);
  const [loadingMore, setLoadingMore] = createSignal(false);
  
  const [showLoading, setShowLoading] = createSignal(!isCached());

  const [workAreas] = createResource(
    () => !isServer,
    async (ready) => {
      if (!ready) return [];
      try { return await apiGet<any[]>("/specialties"); }
      catch { return []; }
    }
  );

  const executeSearch = async (params: { q: string, area: string, loc: string }) => {
    setLoading(true);
    try {
      const url = `/psi/directory?q=${encodeURIComponent(params.q)}&specialty=${params.area}&location=${encodeURIComponent(params.loc)}&limit=${LIMIT}&page=1`;
      const res = await apiGet<DirectoryResponse>(url);
      
      setAllItems(res.data ?? []);
      setTotal(res.total ?? 0);
      setPage(1);
      setHasMore((res.total_pages ?? 1) > 1);
      setIsCached(true); 
    } catch (err) {
      console.error("[directorio] error de búsqueda:", err);
      setAllItems([]);
      setHasMore(false);
    } finally {
      setLoading(false);
    }
  };

  onMount(() => {
    if (!isCached()) {
      executeSearch(searchParams());
      setTimeout(() => setShowLoading(false), 1);
    } else {
      setShowLoading(false);
    }
  });

  const loadMore = async () => {
    if (loadingMore() || !hasMore() || loading()) return;
    setLoadingMore(true);
    try {
      const params  = searchParams();
      const nextPage = page() + 1;
      const url = `/psi/directory?q=${encodeURIComponent(params.q)}&specialty=${params.area}&location=${encodeURIComponent(params.loc)}&limit=${LIMIT}&page=${nextPage}`;
      
      const res = await apiGet<DirectoryResponse>(url);
      
      setAllItems((prev) => [...prev, ...(res.data ?? [])]);
      setPage(nextPage);
      setHasMore(nextPage < (res.total_pages ?? 1));
    } catch (err) {
      console.error("[directorio] error scroll:", err);
    } finally {
      setLoadingMore(false);
    }
  };

  let sentinel: HTMLDivElement | undefined;
  onMount(() => {
    const observer = new IntersectionObserver(
      (entries) => { if (entries[0].isIntersecting) loadMore(); },
      { rootMargin: "400px" }
    );
    if (sentinel) observer.observe(sentinel);
    onCleanup(() => observer.disconnect());
  });

  const handleSearch = (e: Event) => {
    e.preventDefault();
    const newParams = { q: query(), area: workArea(), loc: location() };
    setSearchParams(newParams);
    
    setShowLoading(true);
    executeSearch(newParams).then(() => {
      setTimeout(() => setShowLoading(false), 800);
    });
  };

  // ── UTILIDADES PARA LA BARRA DE FILTROS ────────────────────────────────
  const hasActiveFilters = () => {
    const p = searchParams();
    return p.q !== "" || p.area !== "" || p.loc !== "";
  };

  const clearSearch = () => {
    // 1. Limpiamos los inputs visuales
    setQuery("");
    setWorkArea("");
    setLocation("");
    
    // 2. Limpiamos los parámetros de búsqueda activos
    const emptyParams = { q: "", area: "", loc: "" };
    setSearchParams(emptyParams);
    
    // 3. Ejecutamos la búsqueda limpia
    setShowLoading(true);
    executeSearch(emptyParams).then(() => {
      setTimeout(() => setShowLoading(false), 800);
    });
  };

  // Helper para buscar el nombre del Área según su ID para mostrarlo bonito en la barra
  const getWorkAreaName = (id: string) => {
    const areas = workAreas();
    if (!areas) return id;
    const found = areas.find((a: any) => String(a.id) === id);
    return found ? found.name : id;
  };

  return (
    <>
      <Title>Directorio de Psicólogos | COLPSI Carabobo</Title>
      <Meta name="description" content="Encuentra psicólogos colegiados en el estado Carabobo. Busca por nombre, área de desempeño o ubicación." />
      <Meta name="robots" content="index, follow" />

      <main class="min-h-screen bg-[#fcfcfc] pb-24 font-sans">
        <SearchHeader
          query={query()}
          workArea={workArea()}         
          location={location()}
          workAreas={workAreas()}       
          onQueryChange={setQuery}
          onWorkAreaChange={setWorkArea} 
          onLocationChange={setLocation}
          onSearch={handleSearch}
        />

        <div class="max-w-7xl mx-auto px-6 -mt-10 relative z-20">

          {/* ── BARRA DE FILTROS ACTIVOS ───────────────────────────────── */}
          <Show when={hasActiveFilters() && !showLoading()}>
            <div class="bg-white rounded-2xl shadow-premium border border-gray-100 p-4 mb-8 flex flex-wrap items-center justify-between gap-4 animate-in fade-in slide-in-from-top-4">
              
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-[10px] font-black text-gray-400 uppercase tracking-[0.2em] mr-2">
                  Filtros aplicados:
                </span>
                
                <Show when={searchParams().q}>
                  <span class="bg-blue-50 text-colpsi-blue text-xs font-bold px-3 py-1.5 rounded-xl border border-blue-100 flex items-center gap-1.5">
                    <span class="opacity-50">🔍</span> "{searchParams().q}"
                  </span>
                </Show>
                
                <Show when={searchParams().area}>
                  <span class="bg-emerald-50 text-emerald-700 text-xs font-bold px-3 py-1.5 rounded-xl border border-emerald-100 flex items-center gap-1.5">
                    <span class="opacity-50">🏷️</span> {getWorkAreaName(searchParams().area)}
                  </span>
                </Show>
                
                <Show when={searchParams().loc}>
                  <span class="bg-purple-50 text-purple-700 text-xs font-bold px-3 py-1.5 rounded-xl border border-purple-100 flex items-center gap-1.5">
                    <span class="opacity-50">📍</span> {searchParams().loc}
                  </span>
                </Show>
              </div>

              <button 
                onClick={clearSearch}
                class="text-[10px] font-black text-red-500 hover:text-red-700 hover:bg-red-50 px-4 py-2 rounded-xl transition-all uppercase tracking-widest border border-transparent hover:border-red-100 flex items-center gap-1"
              >
                <span class="text-sm">✕</span> Limpiar Búsqueda
              </button>

            </div>
          </Show>
          {/* ───────────────────────────────────────────────────────────── */}

          <Show when={showLoading()}>
            <LoadingScreen
              image="/psi_loading.png"
              imageAlt="COLPSI Carabobo"
              size={300}
              message="Consultando el Registro de Psicólogos..."
              submessage="Garantizando la idoneidad profesional"
            />
          </Show>

          <Show when={!showLoading()}>
            <Suspense fallback={<div class="text-center p-20 text-gray-400 font-bold">Cargando resultados...</div>}>
              <ResultsGrid
                psychologists={allItems}
                loading={loading()}
                loadingMore={loadingMore()}
                hasMore={hasMore()}
                total={total()}
              />
            </Suspense>
          </Show>

          <div ref={sentinel} class="h-10 w-full" />
        </div>

        <FlagFooter />
      </main>
    </>
  );
}
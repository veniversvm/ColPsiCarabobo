import { createSignal, onCleanup, onMount, Suspense, createEffect } from "solid-js";
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

const LIMIT = 10;
export const ssr = false;

export default function DirectoryPage() {
  // ── ESTADOS DE BÚSQUEDA (Renombrados a WorkArea) ────────────────────────
  const [query,      setQuery]      = createSignal("");
  const [workArea,   setWorkArea]   = createSignal(""); // Almacena el ID del área seleccionada
  const [location,   setLocation]   = createSignal("");
  
  // Parámetros consolidados para disparar la búsqueda
  const [searchParams, setSearchParams] = createSignal({ q: "", area: "", loc: "" });

  const [page,        setPage]        = createSignal(1);
  const [allItems,    setAllItems]    = createStore<DirectoryPsychologist[]>([]);
  const [hasMore,     setHasMore]     = createSignal(true);
  const [loadingMore, setLoadingMore] = createSignal(false);

  // ── FEEDBACK DE CARGA ──────────────────────────────────────────────────
  const [showLoading, setShowLoading] = createSignal(true);
  onMount(() => {
    setTimeout(() => setShowLoading(false), 1200);
  });

  // ── CARGA DE CATÁLOGO (Áreas de Desempeño) ─────────────────────────────
  const [workAreas] = createResource(
    () => !isServer,
    async (ready) => {
      if (!ready) return [];
      try { 
        // El endpoint sigue siendo /specialties en Go, pero aquí lo tratamos como áreas
        return await apiGet<any[]>("/specialties"); 
      }
      catch { return []; }
    }
  );

  // ── RECURSO DE RESULTADOS (Página 1) ──────────────────────────────────
  const [firstPage] = createResource(
    () => (!isServer ? searchParams() : null),
    async (params) => {
      if (!params) return null;
      try {
        setPage(1);
        setHasMore(true);
        
        // Mapeo: mandamos 'area' al parámetro 'specialty' que espera la API de Go
        const url = `/psi/directory?q=${encodeURIComponent(params.q)}&specialty=${params.area}&location=${encodeURIComponent(params.loc)}&limit=${LIMIT}&page=1`;
        
        const res = await apiGet<DirectoryResponse>(url);
        
        setAllItems(res.data ?? []);
        setHasMore((res.total_pages ?? 1) > 1);
        return res;
      } catch (err) {
        console.error("[directorio] error:", err);
        setAllItems([]);
        setHasMore(false);
        return null;
      }
    }
  );

  // Reiniciar loading visual cuando cambian los parámetros de búsqueda
  createEffect(() => {
    searchParams(); 
    setShowLoading(true);
    const timer = setTimeout(() => setShowLoading(false), 1200);
    onCleanup(() => clearTimeout(timer));
  });

  // ── SCROLL INFINITO (Cargar más) ──────────────────────────────────────
  const loadMore = async () => {
    if (loadingMore() || !hasMore()) return;
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
      { rootMargin: "300px" }
    );
    if (sentinel) observer.observe(sentinel);
    onCleanup(() => observer.disconnect());
  });

  const handleSearch = (e: Event) => {
    e.preventDefault();
    setSearchParams({ q: query(), area: workArea(), loc: location() });
  };

  return (
    <>
      <Title>Directorio de Psicólogos | COLPSI Carabobo</Title>
      <Meta name="description" content="Encuentra psicólogos colegiados en el estado Carabobo. Busca por nombre, área de desempeño o ubicación." />
      <Meta name="robots" content="index, follow" />

      <main class="min-h-screen bg-[#fcfcfc] pb-24 font-sans">
        <SearchHeader
          query={query()}
          workArea={workArea()}         // Prop actualizada
          location={location()}
          workAreas={workAreas()}       // Lista actualizada
          onQueryChange={setQuery}
          onWorkAreaChange={setWorkArea} // Handler actualizado
          onLocationChange={setLocation}
          onSearch={handleSearch}
        />

        <div class="max-w-7xl mx-auto px-6 -mt-10 relative z-20">

          <Show when={showLoading() || firstPage.loading}>
            <LoadingScreen
              image="/psi_loading.png"
              imageAlt="COLPSI Carabobo"
              size={300}
              message="Consultando el Registro de Psicólogos..."
              submessage="Garantizando la idoneidad profesional"
            />
          </Show>

          <Show when={!showLoading() && !firstPage.loading}>
            <Suspense fallback={<div class="text-center p-20">Cargando resultados...</div>}>
              <ResultsGrid
                psychologists={allItems}
                loading={false}
                loadingMore={loadingMore()}
                hasMore={hasMore()}
                total={firstPage()?.total}
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
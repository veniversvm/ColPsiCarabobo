// web/src/routes/directorio/index.tsx
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
  const [query,    setQuery]    = createSignal("");
  const [specialty, setSpecialty] = createSignal("");
  const [location, setLocation] = createSignal("");
  const [searchParams, setSearchParams] = createSignal({ q: "", spec: "", location: "" });

  const [page,        setPage]        = createSignal(1);
  const [allItems,    setAllItems]    = createStore<DirectoryPsychologist[]>([]);
  const [hasMore,     setHasMore]     = createSignal(true);
  const [loadingMore, setLoadingMore] = createSignal(false);

  // ── Tiempo muerto de 1.5s para apreciar el loading ───────────────────────
  const [showLoading, setShowLoading] = createSignal(true);
  onMount(() => {
    setTimeout(() => setShowLoading(false), 1500);
  });
  // ─────────────────────────────────────────────────────────────────────────

  const [specialties] = createResource(
    () => !isServer,
    async (ready) => {
      if (!ready) return [];
      try { return await apiGet<any[]>("/specialties"); }
      catch { return []; }
    }
  );

  const [firstPage] = createResource(
    () => (!isServer ? searchParams() : null),
    async (params) => {
      if (!params) return null;
      try {
        setPage(1);
        setHasMore(true);
        const res = await apiGet<DirectoryResponse>(
          `/psi/directory?q=${encodeURIComponent(params.q)}&specialty=${encodeURIComponent(params.spec)}&limit=${LIMIT}&page=1&location=${encodeURIComponent(params.location)}`
        );
        setAllItems(res.data ?? []);
        setHasMore((res.total_pages ?? 1) > 1);
        return res;
      } catch (err) {
        console.error("[directorio] error cargando página 1:", err);
        setAllItems([]);
        setHasMore(false);
        return null;
      }
    }
  );

  // Al hacer una nueva búsqueda, volver a mostrar el loading 1.5s
  createEffect(() => {
    searchParams(); // suscripción reactiva
    setShowLoading(true);
    setTimeout(() => setShowLoading(false), 1500);
  });

  const loadMore = async () => {
    if (loadingMore() || !hasMore()) return;
    setLoadingMore(true);
    try {
      const params  = searchParams();
      const nextPage = page() + 1;
      const res = await apiGet<DirectoryResponse>(
        `/psi/directory?q=${encodeURIComponent(params.q)}&specialty=${encodeURIComponent(params.spec)}&limit=${LIMIT}&page=${nextPage}&location=${encodeURIComponent(params.location)}`
      );
      setAllItems((prev) => [...prev, ...(res.data ?? [])]);
      setPage(nextPage);
      setHasMore(nextPage < (res.total_pages ?? 1));
    } catch (err) {
      console.error("[directorio] error cargando más:", err);
    } finally {
      setLoadingMore(false);
    }
  };

  let sentinel: HTMLDivElement | undefined;

  onMount(() => {
    const observer = new IntersectionObserver(
      (entries) => { if (entries[0].isIntersecting) loadMore(); },
      { rootMargin: "200px" }
    );
    if (sentinel) observer.observe(sentinel);
    onCleanup(() => observer.disconnect());
  });

  const handleSearch = (e: Event) => {
    e.preventDefault();
    setSearchParams({ q: query(), spec: specialty(), location: location() });
  };

  return (
    <>
      <Title>Directorio de Psicólogos | COLPSI Carabobo</Title>
      <Meta name="description" content="Encuentra psicólogos colegiados en el estado Carabobo, Venezuela. Busca por nombre, especialidad o ubicación." />
      <Meta name="robots" content="index, follow" />

      <main class="min-h-screen bg-[#fcfcfc] pb-24">
        <SearchHeader
          query={query()}
          specialty={specialty()}
          location={location()}
          specialties={specialties()}
          onQueryChange={setQuery}
          onSpecialtyChange={setSpecialty}
          onLocationChange={setLocation}
          onSearch={handleSearch}
        />

        <div class="max-w-7xl mx-auto px-6 -mt-10 relative z-20">

          {/* ── Loading institucional ──────────────────────────────────── */}
          <Show when={showLoading() || firstPage.loading}>
            <LoadingScreen
              image="/psi_loading.png"
              imageAlt="Colegio de Psicólogos de Carabobo"
              size={300}
              message="Buscando psicólogos..."
              submessage="Colegio de Psicólogos de Carabobo"
            />
          </Show>

          {/* ── Resultados — ocultos mientras carga ───────────────────── */}
          <Show when={!showLoading() && !firstPage.loading}>
            <Suspense fallback={
              <LoadingScreen
                image="/psi_loading.png"
                message="Cargando resultados..."
              />
            }>
              <ResultsGrid
                psychologists={allItems}
                loading={false}
                loadingMore={loadingMore()}
                hasMore={hasMore()}
                total={firstPage()?.total}
              />
            </Suspense>
          </Show>

          <div ref={sentinel} class="h-4" />
        </div>

        <FlagFooter />
      </main>
    </>
  );
}
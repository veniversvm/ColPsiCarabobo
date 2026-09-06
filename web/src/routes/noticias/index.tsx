// web/src/routes/noticias/index.tsx
import { createSignal, For, Show, onMount, onCleanup } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { LoadingScreen } from "~/components/ui/LoadingScreen";
import { bucketUrl as imgUrl } from "~/lib/bucket";

interface Post {
  id: string;
  title: string;
  short_description: string;
  image_url: string;
  created_at: string;
  create_by: string;
}

interface PaginatedPosts {
  data: Post[];
  page: number;
  total: number;
}

const toSlug = (title: string, id: string) => {
  const slugTitle = title
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^a-z0-9\s]/g, "")
    .trim()
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .slice(0, 55);
  const firstSegment = id.split('-')[0];
  return `${slugTitle}-${firstSegment}`;
};

const LIMIT = 10;

const formatDate = (iso: string) =>
  new Date(iso).toLocaleDateString("es-VE", {
    day: "numeric", month: "long", year: "numeric",
  });

// ─────────────────────────────────────────────────────────────────────────────
export default function PublicNoticiasPage() {
  const [posts,         setPosts]         = createSignal<Post[]>([]);
  const [page,          setPage]          = createSignal(1);
  const [loading,       setLoading]       = createSignal(false);
  const [hasMore,       setHasMore]       = createSignal(true);
  const [initialDone,   setInitialDone]   = createSignal(false);
  const [search,        setSearch]        = createSignal("");
  const [searchTimeout, setSearchTimeout] = createSignal<ReturnType<typeof setTimeout> | null>(null);

  let sentinelRef!: HTMLDivElement;
  let observer: IntersectionObserver | null = null;

  const loadPage = async (pageNum: number, q: string, replace = false) => {
    if (loading()) return;
    setLoading(true);
    try {
      const params = new URLSearchParams({ page: String(pageNum), limit: String(LIMIT) });
      if (q) params.set("search", q);
      const data = await apiGet<PaginatedPosts>(`/posts?${params.toString()}`);
      const result = data?.data ?? [];
      if (replace) {
        setPosts(result);
      } else {
        setPosts((prev) => [...prev, ...result]);
      }
      setHasMore(result.length === LIMIT);
    } catch (err) {
      console.error("Error cargando posts:", err);
      setHasMore(false);
    } finally {
      setLoading(false);
      setInitialDone(true);
    }
  };

  onMount(() => {
    loadPage(1, "");

    observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore() && !loading()) {
          const next = page() + 1;
          setPage(next);
          loadPage(next, search());
        }
      },
      { rootMargin: "200px" }
    );

    if (sentinelRef) observer.observe(sentinelRef);
  });

  onCleanup(() => {
    observer?.disconnect();
    const t = searchTimeout();
    if (t) clearTimeout(t);
  });

  const handleSearch = (q: string) => {
    setSearch(q);
    const prev = searchTimeout();
    if (prev) clearTimeout(prev);
    const t = setTimeout(() => {
      setPage(1);
      setHasMore(true);
      loadPage(1, q, true);
    }, 350);
    setSearchTimeout(t);
  };

  const hero = () => posts()[0];
  const rest = () => posts().slice(1);

  return (
    <main class="min-h-screen bg-colpsi-bg">

      {/* ── CABECERA ───────────────────────────────────────────────────────── */}
      <header class="bg-colpsi-navy text-white py-16 px-6 relative overflow-hidden">
        <div class="absolute inset-0 opacity-5"
          style="background-image: repeating-linear-gradient(45deg, #fff 0, #fff 1px, transparent 0, transparent 50%); background-size: 12px 12px;" />
        <div class="max-w-5xl mx-auto relative">
          <p class="text-xs font-black uppercase tracking-[0.3em] text-blue-300 mb-3">
            Colegio de Psicólogos de Carabobo
          </p>
          <h1 class="text-4xl md:text-5xl font-black leading-none uppercase tracking-tight mb-4">
            Noticias &<br />
            <span class="text-colpsi-yellow">Comunicados</span>
          </h1>
          <p class="text-blue-200 text-sm max-w-md leading-relaxed">
            Información institucional, convocatorias y novedades para la comunidad psicológica de Venezuela.
          </p>
        </div>
      </header>

      {/* ── CONTENIDO ──────────────────────────────────────────────────────── */}
      <div class="max-w-5xl mx-auto px-4 py-12">

        {/* ── Loading institucional — reemplaza el SkeletonLoader ─────────── */}
        <Show when={!initialDone()}>
          <LoadingScreen
            image="/noticias.png"
            imageAlt="Noticias COLPSI Carabobo"
            message="Cargando publicaciones..."
            submessage="Colegio de Psicólogos de Carabobo"
            size={300}
          />
        </Show>

        {/* Sin resultados */}
        <Show when={initialDone() && posts().length === 0 && !loading()}>
          <EmptyState hasSearch={search().length > 0} />
        </Show>

        <Show when={posts().length > 0}>

          {/* ── HERO ─────────────────────────────────────────────────────── */}
          <Show when={hero()}>
            {(post) => (
              <A href={`/noticias/${toSlug(post().title, post().id)}`} state={{ id: post().id }} class="group block mb-10">
                <article class="bg-white rounded-3xl overflow-hidden shadow-sm hover:shadow-xl transition-all duration-300 border border-colpsi-border grid md:grid-cols-2">
                  <div class="relative h-56 md:h-auto bg-gradient-to-br from-blue-900 to-blue-700 overflow-hidden">
                    <Show when={post().image_url}>
                      <img
                        src={imgUrl(post().image_url)}
                        alt={post().title}
                        class="w-full h-full object-cover opacity-90 group-hover:scale-105 transition-transform duration-500"
                        loading="eager"
                      />
                    </Show>
                    <div class="absolute top-4 left-4">
                      <span class="bg-yellow-400 text-blue-900 text-[10px] font-black uppercase tracking-widest px-3 py-1.5 rounded-full shadow">
                        Destacado
                      </span>
                    </div>
                  </div>
                  <div class="p-8 flex flex-col justify-center">
                    <p class="text-[11px] font-black text-blue-400 uppercase tracking-widest mb-3">
                      {formatDate(post().created_at)}
                    </p>
                    <h2 class="text-2xl font-black text-gray-900 leading-tight mb-3 group-hover:text-blue-800 transition-colors">
                      {post().title}
                    </h2>
                    <p class="text-gray-500 text-sm leading-relaxed line-clamp-3 mb-6">
                      {post().short_description}
                    </p>
                    <span class="inline-flex items-center gap-2 text-blue-700 font-black text-sm group-hover:gap-3 transition-all">
                      Leer más
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                      </svg>
                    </span>
                  </div>
                </article>
              </A>
            )}
          </Show>

          {/* ── GRILLA ───────────────────────────────────────────────────── */}
          <Show when={rest().length > 0}>
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
              <For each={rest()}>
                {(post) => <PostCard post={post} />}
              </For>
            </div>
          </Show>

        </Show>

        <div ref={sentinelRef} class="h-1" />

        {/* Spinner "cargando más" — scroll infinito, no reemplaza el loading inicial */}
        <Show when={loading() && initialDone()}>
          <div class="flex justify-center py-10">
            <div class="flex items-center gap-3 text-gray-400 text-sm font-bold">
              <div class="w-5 h-5 border-2 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
              Cargando más publicaciones...
            </div>
          </div>
        </Show>

        {/* Fin del feed */}
        <Show when={!hasMore() && initialDone() && posts().length > 0 && !loading()}>
          <div class="text-center py-10">
            <div class="inline-flex items-center gap-3 text-gray-400 text-xs font-bold">
              <div class="h-px w-12 bg-gray-200" />
              {posts().length} publicación{posts().length !== 1 ? "es" : ""} en total
              <div class="h-px w-12 bg-gray-200" />
            </div>
          </div>
        </Show>

      </div>
    </main>
  );
}

// ── Tarjeta ───────────────────────────────────────────────────────────────────
function PostCard(props: { post: Post }) {
  const { post } = props;
  return (
    <A href={`/noticias/${toSlug(post.title, post.id)}`} state={{ id: post.id }} class="group block">
      <article class="bg-white rounded-2xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 border border-colpsi-border h-full flex flex-col">
        <div class="relative h-44 bg-gradient-to-br from-blue-800 to-blue-600 overflow-hidden flex-shrink-0">
          <Show when={post.image_url}>
            <img
              src={imgUrl(post.image_url)}
              alt={post.title}
              class="w-full h-full object-cover opacity-90 group-hover:scale-105 transition-transform duration-500"
              loading="lazy"
            />
          </Show>
          <Show when={!post.image_url}>
            <div class="w-full h-full flex items-center justify-center opacity-20">
              <svg class="w-16 h-16 text-white" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
              </svg>
            </div>
          </Show>
        </div>
        <div class="p-5 flex flex-col flex-1">
          <p class="text-[10px] font-black text-blue-400 uppercase tracking-widest mb-2">
            {formatDate(post.created_at)}
          </p>
          <h3 class="font-black text-gray-900 text-base leading-snug mb-2 group-hover:text-blue-800 transition-colors line-clamp-2">
            {post.title}
          </h3>
          <p class="text-gray-500 text-xs leading-relaxed line-clamp-2 flex-1">
            {post.short_description || <span class="italic text-gray-300">Sin resumen</span>}
          </p>
          <div class="mt-4 pt-4 border-t border-colpsi-border flex items-center justify-between">
            <span class="text-[10px] text-gray-400 font-bold">COLPSI Carabobo</span>
            <span class="text-blue-600 font-black text-xs group-hover:translate-x-1 transition-transform inline-block">
              Leer →
            </span>
          </div>
        </div>
      </article>
    </A>
  );
}

// ── Estado vacío ──────────────────────────────────────────────────────────────
function EmptyState(props: { hasSearch: boolean }) {
  return (
    <div class="text-center py-24 bg-white rounded-3xl border border-colpsi-border">
      <p class="text-5xl mb-4">{props.hasSearch ? "🔍" : "📰"}</p>
      <h2 class="text-lg font-black text-gray-700 mb-2">
        {props.hasSearch ? "Sin resultados" : "Sin publicaciones aún"}
      </h2>
      <p class="text-gray-400 text-sm max-w-xs mx-auto">
        {props.hasSearch
          ? "Intenta con otras palabras clave."
          : "Pronto publicaremos noticias y comunicados institucionales."}
      </p>
    </div>
  );
}
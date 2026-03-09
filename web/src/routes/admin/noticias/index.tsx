// routes/admin/noticias/index.tsx
import { createResource, createSignal, For, Show, Suspense, ErrorBoundary } from "solid-js";
import { A, action, useAction } from "@solidjs/router";
import { apiDelete, apiGet, apiPatch } from "~/lib/api";

// ── Tipos locales ─────────────────────────────────────────────────────────────
interface Post {
  id: string;
  title: string;
  short_description: string;
  type: "public" | "psi";
  image_url: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  create_by: string;
}

// ── Acción de servidor: toggle is_active ──────────────────────────────────────
const togglePostActive = action(async (params: { id: string; is_active: boolean }) => {
  "use server";
  return await apiPatch(`/admin/posts/${params.id}`, { is_active: params.is_active });
});

// ── Acción de servidor: borrar post ──────────────────────────────────────────
const deletePost = action(async (id: string) => {
  "use server";
  return await apiDelete(`/admin/posts/${id}`);
});

// ── Helpers ───────────────────────────────────────────────────────────────────
const formatDate = (iso: string) => {
  if (!iso) return "";
  return new Date(iso).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
};

const TYPE_LABELS: Record<string, { label: string; color: string }> = {
  public: { label: "Público", color: "bg-emerald-100 text-emerald-700" },
  psi:    { label: "Solo Colegiados", color: "bg-blue-100 text-blue-700" },
};

// ─────────────────────────────────────────────────────────────────────────────
export default function AdminNoticiasPage() {
  const runToggle = useAction(togglePostActive);
  const runDelete = useAction(deletePost);

  // Filtros UI
  const [filterType, setFilterType] = createSignal<"all" | "public" | "psi">("all");
  const [filterActive, setFilterActive] = createSignal<"all" | "active" | "draft">("all");
  const [search, setSearch] = createSignal("");
  const [confirmDelete, setConfirmDelete] = createSignal<string | null>(null);
  const [busy, setBusy] = createSignal<string | null>(null);

  // Se asume endpoint /admin/posts para consistencia con los métodos de borrar/editar
  const [posts, { refetch }] = createResource(() => apiGet<Post[]>("/posts"));

  // Extractor seguro: Garantiza que SIEMPRE se trabaje con un Array
  // Útil si el backend devuelve: { data: [...] } en lugar de [...]
  const postList = () => {
    const data = posts();
    if (!data) return [];
    return Array.isArray(data) ? data : (data as any).data || (data as any).posts || [];
  };

  // Filtrado en cliente a prueba de nulos
  const filtered = () => {
    const q = search().toLowerCase().trim();
    return postList().filter((p: Post) => {
      if (!p) return false;
      if (filterType() !== "all" && p.type !== filterType()) return false;
      if (filterActive() === "active" && !p.is_active) return false;
      if (filterActive() === "draft" && p.is_active) return false;
      
      const title = (p.title || "").toLowerCase();
      const desc = (p.short_description || "").toLowerCase();
      
      if (q && !title.includes(q) && !desc.includes(q)) return false;
      return true;
    });
  };

  const handleToggle = async (post: Post) => {
    setBusy(post.id);
    try {
      await runToggle({ id: post.id, is_active: !post.is_active });
      refetch();
    } finally {
      setBusy(null);
    }
  };

  const handleDelete = async (id: string) => {
    setBusy(id);
    try {
      await runDelete(id);
      setConfirmDelete(null);
      refetch();
    } finally {
      setBusy(null);
    }
  };

  // ─── Render ───────────────────────────────────────────────────────────────
  return (
    <main class="pb-20 animate-in fade-in duration-500">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
        <div>
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
            Publicaciones
          </h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium">
            Gestión de noticias y comunicados del Colegio
          </p>
        </div>
        <A
          href="/admin/noticias/crear"
          class="inline-flex items-center gap-2 bg-blue-800 hover:bg-blue-900 text-white font-black px-6 py-3 rounded-2xl shadow-lg hover:scale-105 active:scale-95 transition-all text-sm"
        >
          <span class="text-lg leading-none">＋</span>
          Nueva Publicación
        </A>
      </div>

      {/* ── FILTROS ───────────────────────────────────────────────────────── */}
      <div class="flex flex-col md:flex-row gap-3 mb-6">
        {/* Búsqueda */}
        <div class="relative flex-1">
          <svg class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
          </svg>
          <input
            type="text"
            placeholder="Buscar por título o resumen..."
            value={search()}
            onInput={(e) => setSearch(e.currentTarget.value)}
            class="w-full pl-10 pr-4 py-2.5 bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl outline-none text-sm text-gray-800 transition-all"
          />
        </div>

        {/* Tipo */}
        <div class="flex gap-2">
          {(["all", "public", "psi"] as const).map((t) => (
            <button
              onClick={() => setFilterType(t)}
              class={`px-4 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide transition-all border-2 ${
                filterType() === t
                  ? "bg-blue-800 text-white border-blue-800"
                  : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
              }`}
            >
              {t === "all" ? "Todos" : t === "public" ? "Públicos" : "Colegiados"}
            </button>
          ))}
        </div>

        {/* Estado */}
        <div class="flex gap-2">
          {(["all", "active", "draft"] as const).map((s) => (
            <button
              onClick={() => setFilterActive(s)}
              class={`px-4 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide transition-all border-2 ${
                filterActive() === s
                  ? "bg-blue-800 text-white border-blue-800"
                  : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
              }`}
            >
              {s === "all" ? "Todos" : s === "active" ? "Publicados" : "Borradores"}
            </button>
          ))}
        </div>
      </div>

      {/* ── LISTADO CON MANEJO DE ERRORES ───────────────────────────────── */}
      <ErrorBoundary fallback={(err, reset) => (
        <div class="bg-red-50 border border-red-200 p-8 rounded-3xl text-center">
          <p class="text-4xl mb-4">🚨</p>
          <h2 class="text-xl font-black text-red-800 mb-2">Error de Conexión</h2>
          <p class="text-red-600 text-sm mb-6 max-w-lg mx-auto">{err.toString()}</p>
          <button 
            onClick={reset} 
            class="bg-red-600 text-white font-black px-6 py-2.5 rounded-xl hover:bg-red-700 active:scale-95 transition-all text-sm"
          >
            ↻ Intentar de nuevo
          </button>
        </div>
      )}>
        <Suspense fallback={
          <div class="space-y-4">
            <For each={[1, 2, 3, 4]}>
              {() => <div class="h-28 bg-white animate-pulse rounded-2xl border border-gray-100" />}
            </For>
          </div>
        }>
          {/* Sin posts en BD */}
          <Show when={!posts.loading && posts.state === "ready" && postList().length === 0}>
            <div class="text-center py-20 bg-white rounded-3xl border border-gray-100">
              <p class="text-5xl mb-4">📰</p>
              <p class="text-gray-400 font-bold">No hay publicaciones aún</p>
              <A href="/admin/noticias/crear" class="mt-4 inline-block text-blue-600 font-black text-sm hover:underline">
                Crear la primera →
              </A>
            </div>
          </Show>

          {/* Hay posts pero ninguno pasa el filtro */}
          <Show when={!posts.loading && posts.state === "ready" && postList().length > 0 && filtered().length === 0}>
            <div class="text-center py-16 bg-white rounded-3xl border border-gray-100">
              <p class="text-gray-400 font-bold">Ningún resultado para los filtros aplicados</p>
            </div>
          </Show>

          <div class="space-y-3">
            <For each={filtered()}>
              {(post) => {
                const typeInfo = TYPE_LABELS[post.type] ?? { label: post.type || "Desconocido", color: "bg-gray-100 text-gray-600" };
                const isBusy = () => busy() === post.id;

                return (
                  <article class={`bg-white rounded-2xl border-2 transition-all duration-200 overflow-hidden ${
                    post.is_active ? "border-gray-100 hover:border-blue-100" : "border-dashed border-gray-200 opacity-70"
                  }`}>
                    <div class="flex items-start gap-4 p-4 md:p-5">

                      {/* Imagen */}
                      <div class="flex-shrink-0 w-20 h-20 md:w-24 md:h-24 rounded-xl overflow-hidden bg-gray-100 border border-gray-200">
                        <Show
                          when={post.image_url}
                          fallback={
                            <div class="w-full h-full flex items-center justify-center text-gray-300 text-2xl">
                              📄
                            </div>
                          }
                        >
                          <img
                            src={`http://localhost:9000/colpsi-bucket/${post.image_url}`}
                            alt={post.title}
                            class="w-full h-full object-cover"
                            loading="lazy"
                          />
                        </Show>
                      </div>

                      {/* Contenido */}
                      <div class="flex-1 min-w-0">
                        <div class="flex flex-wrap items-center gap-2 mb-1.5">
                          <span class={`text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider ${typeInfo.color}`}>
                            {typeInfo.label}
                          </span>
                          <Show when={!post.is_active}>
                            <span class="text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider bg-amber-100 text-amber-700">
                              Borrador
                            </span>
                          </Show>
                        </div>

                        <h2 class="font-black text-gray-900 text-base leading-tight truncate">
                          {post.title}
                        </h2>

                        <p class="text-gray-500 text-sm mt-1 line-clamp-1">
                          {post.short_description || <span class="italic text-gray-300">Sin resumen</span>}
                        </p>

                        <div class="flex items-center gap-3 mt-2 text-[11px] text-gray-400 font-medium">
                          <span>Por <span class="font-bold text-gray-600">{post.create_by}</span></span>
                          <span>·</span>
                          <span>{formatDate(post.created_at)}</span>
                        </div>
                      </div>

                      {/* Acciones */}
                      <div class="flex-shrink-0 flex flex-col sm:flex-row items-center gap-2">
                        {/* Toggle publicado */}
                        <button
                          onClick={() => handleToggle(post)}
                          disabled={isBusy()}
                          title={post.is_active ? "Despublicar" : "Publicar"}
                          class={`w-9 h-9 rounded-xl flex items-center justify-center border-2 transition-all font-bold text-sm disabled:opacity-40 ${
                            post.is_active
                              ? "border-emerald-200 bg-emerald-50 text-emerald-600 hover:bg-emerald-100"
                              : "border-gray-200 bg-gray-50 text-gray-400 hover:bg-gray-100"
                          }`}
                        >
                          {isBusy() ? "…" : post.is_active ? "✓" : "○"}
                        </button>

                        {/* Editar */}
                        <A
                          href={`/admin/noticias/${post.id}`}
                          class="w-9 h-9 rounded-xl flex items-center justify-center border-2 border-blue-100 bg-blue-50 text-blue-600 hover:bg-blue-100 transition-all"
                          title="Editar"
                        >
                          ✏
                        </A>

                        {/* Eliminar */}
                        <button
                          onClick={() => setConfirmDelete(post.id)}
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

          {/* Contador */}
          <Show when={postList().length > 0}>
            <p class="text-center text-xs text-gray-400 font-bold mt-6">
              Mostrando {filtered().length} de {postList().length} publicaciones
            </p>
          </Show>
        </Suspense>
      </ErrorBoundary>

      {/* ── MODAL CONFIRMACIÓN BORRADO ─────────────────────────────────── */}
      <Show when={confirmDelete()}>
        <div
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
          onClick={(e) => { if (e.target === e.currentTarget) setConfirmDelete(null); }}
        >
          <div class="bg-white rounded-3xl shadow-2xl p-8 w-full max-w-sm border border-gray-100 text-center animate-in zoom-in-95 duration-200">
            <p class="text-4xl mb-4">🗑️</p>
            <h2 class="text-lg font-black text-gray-900 mb-2">¿Eliminar publicación?</h2>
            <p class="text-gray-500 text-sm mb-6">Esta acción no se puede deshacer.</p>
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
                {busy() === confirmDelete() ? "Eliminando..." : "Sí, eliminar"}
              </button>
            </div>
          </div>
        </div>
      </Show>

    </main>
  );
}
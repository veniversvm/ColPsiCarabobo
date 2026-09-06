// web/src/routes/admin/noticias/index.tsx
import { createResource, createSignal, For, Show, Suspense, ErrorBoundary } from "solid-js";
import { A, action, useAction } from "@solidjs/router";
import { apiDelete, apiGet, apiPatch } from "~/lib/api";
import { PaginatedResponse } from "~/types/admin";
import {
  Post,
  NoticiasHeader,
  NoticiasFilters,
  NoticiaCard,
  EmptyState,
  DeleteModal,
} from "~/components/admin/noticias";

// ── Acciones de servidor ──────────────────────────────────────────────────────
const togglePostStatus = action(async (params: { id: string; currentStatus: Post["status"] }) => {
  "use server";
  const newStatus = params.currentStatus === "published" ? "draft" : "published";
  return await apiPatch(`/admin/posts/${params.id}`, { status: newStatus });
});

const archivePost = action(async (id: string) => {
  "use server";
  return await apiPatch(`/admin/posts/${id}`, { status: "archived" });
});

// ─────────────────────────────────────────────────────────────────────────────
export default function AdminNoticiasPage() {
  const runToggle = useAction(togglePostStatus);
  const runDelete = useAction(archivePost);

  const [filterType,   setFilterType]   = createSignal<"all" | "public" | "psi">("all");
  const [filterStatus, setFilterStatus] = createSignal<"all" | Post["status"]>("all");
  const [search,       setSearch]       = createSignal("");
  const [confirmDelete, setConfirmDelete] = createSignal<string | null>(null);
  const [busy, setBusy] = createSignal<string | null>(null);

  const [posts, { refetch }] = createResource(() => apiGet<PaginatedResponse<Post>>("/posts?page=1&limit=50"));

  const postList = () => posts()?.data ?? [];

  const filtered = () => {
    const q = search().toLowerCase().trim();
    return postList().filter((p: Post) => {
      if (!p) return false;
      if (filterType()   !== "all" && p.type   !== filterType())   return false;
      if (filterStatus() !== "all" && p.status !== filterStatus()) return false;
      const title = (p.title || "").toLowerCase();
      const desc  = (p.short_description || "").toLowerCase();
      if (q && !title.includes(q) && !desc.includes(q)) return false;
      return true;
    });
  };

  const handleToggle = async (post: Post) => {
    setBusy(post.id);
    try {
      await runToggle({ id: post.id, currentStatus: post.status });
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

  return (
    <main class="pb-20 animate-in fade-in duration-500">

      <NoticiasHeader />

      <NoticiasFilters
        search={search}
        setSearch={setSearch}
        filterType={filterType}
        setFilterType={setFilterType}
        filterStatus={filterStatus}
        setFilterStatus={setFilterStatus}
      />

      <ErrorBoundary fallback={(err, reset) => (
        <div class="bg-red-50 border border-red-200 p-8 rounded-3xl text-center">
          <p class="text-4xl mb-4">🚨</p>
          <h2 class="text-xl font-black text-red-800 mb-2">Error de Conexión</h2>
          <p class="text-red-600 text-sm mb-6 max-w-lg mx-auto">{err.toString()}</p>
          <button onClick={reset} class="bg-red-600 text-white font-black px-6 py-2.5 rounded-xl hover:bg-red-700 active:scale-95 transition-all text-sm">
            ↻ Intentar de nuevo
          </button>
        </div>
      )}>
        <Suspense fallback={
          <div class="space-y-4">
            <For each={[1, 2, 3, 4]}>{() => <div class="h-28 bg-white animate-pulse rounded-2xl border border-colpsi-border" />}</For>
          </div>
        }>
          <Show when={!posts.loading && posts.state === "ready" && postList().length === 0}>
            <EmptyState type="no-posts" />
          </Show>

          <Show when={!posts.loading && posts.state === "ready" && postList().length > 0 && filtered().length === 0}>
            <EmptyState type="no-results" hasFilters={true} />
          </Show>

          <div class="space-y-3">
            <For each={filtered()}>
              {(post) => (
                <NoticiaCard
                  post={post}
                  isBusy={busy() === post.id}
                  onToggle={handleToggle}
                  onDelete={(id) => setConfirmDelete(id)}
                />
              )}
            </For>
          </div>

          <Show when={postList().length > 0}>
            <p class="text-center text-xs text-gray-400 font-bold mt-6">
              Mostrando {filtered().length} de {postList().length} publicaciones
            </p>
          </Show>
        </Suspense>
      </ErrorBoundary>

      <DeleteModal
        isOpen={!!confirmDelete()}
        isBusy={busy() === confirmDelete()}
        onConfirm={() => handleDelete(confirmDelete()!)}
        onCancel={() => setConfirmDelete(null)}
      />

    </main>
  );
}
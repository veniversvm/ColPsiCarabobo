// routes/admin/staff/index.tsx
import { createResource, createSignal, For, Show, Suspense } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { apiGet, apiDelete } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { roleLabel } from "~/lib/staff-permissions";

interface Admin {
  id: string;
  username: string;
  email: string;
  is_active: boolean;
  created_at: string;
  create_by: string;
  role?: string | null;
  // permisos
  can_read_psi: boolean;
  can_create_psi: boolean;
  can_update_psi: boolean;
  can_delete_psi: boolean;
  can_create_admin: boolean;
  can_update_admin: boolean;
  can_delete_admin: boolean;
  can_publish: boolean;
  can_update_publish: boolean;
  can_delete_publish: boolean;
  can_send_notifications: boolean;
  can_manage_notifications: boolean;
  can_read_notifications: boolean;
  can_create_tags: boolean;
  can_edit_tags: boolean;
  can_delete_tags: boolean;
  can_manage_projects: boolean;
  can_manage_tickets: boolean;
}

interface AdminListResponse {
  data: Admin[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

const formatDate = (iso: string) => {
  if (!iso) return "";
  return new Date(iso).toLocaleDateString("es-VE", { day: "2-digit", month: "short", year: "numeric" });
};

const countPerms = (a: Admin) =>
  [
    a.can_read_psi, a.can_create_psi, a.can_update_psi, a.can_delete_psi,
    a.can_create_admin, a.can_update_admin, a.can_delete_admin,
    a.can_publish, a.can_update_publish, a.can_delete_publish,
    a.can_send_notifications, a.can_manage_notifications, a.can_read_notifications,
    a.can_create_tags, a.can_edit_tags, a.can_delete_tags,
    a.can_manage_projects, a.can_manage_tickets,
  ].filter(Boolean).length;

export default function AdminStaffPage() {
  const navigate = useNavigate();
  const [search, setSearch] = createSignal("");
  const [filterActive, setFilterActive] = createSignal<"all" | "active" | "inactive">("all");
  const [confirmDelete, setConfirmDelete] = createSignal<Admin | null>(null);
  const [busy, setBusy] = createSignal<string | null>(null);
  const [deleteError, setDeleteError] = createSignal<string | null>(null);

  const [result, { refetch }] = createResource(
    () => search(),
    async (q) => {
      try {
        return await apiGet<AdminListResponse>(`/admin/list?limit=50&search=${encodeURIComponent(q)}`);
      } catch (err: any) {
        console.error("[staff] error:", err?.status, err?.message);
        return null;
      }
    }
  );

  const list = () => {
    const data = result()?.data;
    if (!data) return [];
    return data;
  };

  const filtered = () =>
    list().filter((a: Admin) => {
      if (filterActive() === "active" && !a.is_active) return false;
      if (filterActive() === "inactive" && a.is_active) return false;
      return true;
    });

  const handleDelete = async (admin: Admin) => {
    setBusy(admin.id);
    setDeleteError(null);
    try {
      await apiDelete(`/admin/delete/${admin.id}`);
      setConfirmDelete(null);
      refetch();
    } catch (err: any) {
        setDeleteError(getUserFacingError(err));
    } finally {
      setBusy(null);
    }
  };

  return (
    <main class="pb-20 animate-in fade-in duration-500">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
        <div>
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">Staff</h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium">Personal administrativo del sistema</p>
        </div>
        <A
          href="/admin/staff/crear"
          class="inline-flex items-center gap-2 bg-blue-800 hover:bg-blue-900 text-white font-black px-6 py-3 rounded-2xl shadow-lg hover:scale-105 active:scale-95 transition-all text-sm"
        >
          <span class="text-lg leading-none">＋</span>
          Nuevo Administrador
        </A>
      </div>

      {/* ── FILTROS ───────────────────────────────────────────────────────── */}
      <div class="flex flex-col md:flex-row gap-3 mb-6">
        <div class="relative flex-1">
          <svg class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
          </svg>
          <input
            type="text"
            placeholder="Buscar por usuario o email..."
            value={search()}
            onInput={(e) => setSearch(e.currentTarget.value)}
            class="w-full pl-10 pr-4 py-2.5 bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl outline-none text-sm text-gray-800 transition-all"
          />
        </div>
        <div class="flex gap-2">
          {(["all", "active", "inactive"] as const).map((s) => (
            <button
              onClick={() => setFilterActive(s)}
              class={`px-4 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide transition-all border-2 ${
                filterActive() === s
                  ? "bg-blue-800 text-white border-blue-800"
                  : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
              }`}
            >
              {s === "all" ? "Todos" : s === "active" ? "Activos" : "Inactivos"}
            </button>
          ))}
        </div>
      </div>

      {/* ── LISTADO ───────────────────────────────────────────────────────── */}
      <Suspense fallback={
        <div class="space-y-3">
          <For each={[1, 2, 3]}>{() => <div class="h-24 bg-white animate-pulse rounded-2xl border border-gray-100" />}</For>
        </div>
      }>
        <Show when={!result.loading && list().length === 0}>
          <div class="text-center py-20 bg-white rounded-3xl border border-gray-100">
            <p class="text-5xl mb-4">👤</p>
            <p class="text-gray-400 font-bold">No hay administradores registrados</p>
            <A href="/admin/staff/crear" class="mt-4 inline-block text-blue-600 font-black text-sm hover:underline">
              Crear el primero →
            </A>
          </div>
        </Show>

        <Show when={!result.loading && list().length > 0 && filtered().length === 0}>
          <div class="text-center py-16 bg-white rounded-3xl border border-gray-100">
            <p class="text-gray-400 font-bold">Ningún resultado para los filtros aplicados</p>
          </div>
        </Show>

        <div class="space-y-3">
          <For each={filtered()}>
            {(admin) => {
              const isBusy = () => busy() === admin.id;
              const permsCount = countPerms(admin);
              return (
                <article class={`bg-white rounded-2xl border-2 transition-all duration-200 overflow-hidden ${
                  admin.is_active ? "border-gray-100 hover:border-blue-100" : "border-dashed border-gray-200 opacity-70"
                }`}>
                  <div class="flex items-center gap-4 p-4 md:p-5">

                    {/* Avatar */}
                    <div class={`flex-shrink-0 w-12 h-12 rounded-2xl flex items-center justify-center font-black text-lg uppercase border-2 ${
                      admin.is_active ? "bg-blue-50 border-blue-100 text-blue-700" : "bg-gray-50 border-gray-200 text-gray-400"
                    }`}>
                      {admin.username.charAt(0)}
                    </div>

                    {/* Info */}
                    <div class="flex-1 min-w-0">
                      <div class="flex flex-wrap items-center gap-2 mb-1">
                        <h2 class="font-black text-gray-900 text-base">{admin.username}</h2>
                        <span class={`text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider ${
                          admin.is_active ? "bg-emerald-100 text-emerald-700" : "bg-gray-100 text-gray-500"
                        }`}>
                          {admin.is_active ? "Activo" : "Inactivo"}
                        </span>
                        <span class="text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider bg-blue-50 text-blue-600">
                          {permsCount}/18 permisos
                        </span>
                        <span class={`text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider ${
                          admin.role ? "bg-indigo-50 text-indigo-600" : "bg-gray-100 text-gray-400"
                        }`}>
                          {roleLabel(admin.role)}
                        </span>
                      </div>
                      <p class="text-gray-500 text-sm truncate">{admin.email}</p>
                      <div class="flex items-center gap-3 mt-1.5 text-[11px] text-gray-400 font-medium">
                        <Show when={admin.create_by}>
                          <span>Creado por <span class="font-bold text-gray-600">{admin.create_by}</span></span>
                          <span>·</span>
                        </Show>
                        <span>{formatDate(admin.created_at)}</span>
                      </div>
                    </div>

                    {/* Acciones */}
                    <div class="flex-shrink-0 flex items-center gap-2">
                      <A
                        href={`/admin/staff/${admin.id}`}
                        class="w-9 h-9 rounded-xl flex items-center justify-center border-2 border-blue-100 bg-blue-50 text-blue-600 hover:bg-blue-100 transition-all"
                        title="Editar"
                      >
                        ✏
                      </A>
                      <button
                        onClick={() => { setDeleteError(null); setConfirmDelete(admin); }}
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

        <Show when={list().length > 0}>
          <p class="text-center text-xs text-gray-400 font-bold mt-6">
            Mostrando {filtered().length} de {result()?.total ?? list().length} administradores
          </p>
        </Show>
      </Suspense>

      {/* ── MODAL CONFIRMACIÓN BORRADO ─────────────────────────────────── */}
      <Show when={confirmDelete()}>
        {(admin) => (
          <div
            class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
            onClick={(e) => { if (e.target === e.currentTarget) setConfirmDelete(null); }}
          >
            <div class="bg-white rounded-3xl shadow-2xl p-8 w-full max-w-sm border border-gray-100 text-center animate-in zoom-in-95 duration-200">
              <p class="text-4xl mb-4">🗑️</p>
              <h2 class="text-lg font-black text-gray-900 mb-1">¿Eliminar administrador?</h2>
              <p class="text-blue-700 font-black text-sm mb-1">{admin().username}</p>
              <p class="text-gray-500 text-sm mb-4">Esta acción es irreversible.</p>
              <Show when={deleteError()}>
                <div class="mb-4 p-3 rounded-xl bg-red-50 text-red-700 text-xs font-bold border border-red-200">
                  {deleteError()}
                </div>
              </Show>
              <div class="flex gap-3">
                <button
                  onClick={() => setConfirmDelete(null)}
                  class="flex-1 px-4 py-3 rounded-2xl border-2 border-gray-200 font-black text-gray-600 hover:bg-gray-50 transition-all text-sm"
                >
                  Cancelar
                </button>
                <button
                  onClick={() => handleDelete(admin())}
                  disabled={busy() === admin().id}
                  class="flex-1 px-4 py-3 rounded-2xl bg-red-600 text-white font-black hover:bg-red-700 active:scale-95 transition-all text-sm disabled:opacity-60"
                >
                  {busy() === admin().id ? "Eliminando..." : "Sí, eliminar"}
                </button>
              </div>
            </div>
          </div>
        )}
      </Show>

    </main>
  );
}
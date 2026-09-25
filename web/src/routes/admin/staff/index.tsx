// routes/admin/staff/index.tsx
import { createResource, createSignal, For, Show, Suspense } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { apiGet, apiDelete, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { roleLabel } from "~/lib/staff-permissions";
import { PageHeader } from "~/components/admin/ui/PageHeader";
import { Input } from "~/components/admin/ui/Input";
import { Badge } from "~/components/admin/ui/Badge";
import { Button } from "~/components/admin/ui/Button";
import { Icon } from "~/components/admin/ui/icons";

interface AdminMe {
  id: string;
  sudo: boolean;
  can_view_logs?: boolean;
}

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
  can_view_logs?: boolean;
  can_export_logs?: boolean;
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
    a.can_view_logs ?? false, a.can_export_logs ?? false,
  ].filter(Boolean).length;

const btnLinkBase =
  "inline-flex items-center justify-center gap-2 h-9 px-3.5 rounded-md text-sm font-semibold transition-colors outline-none focus:ring-2";

export default function AdminStaffPage() {
  const navigate = useNavigate();
  const [search, setSearch] = createSignal("");
  const [filterActive, setFilterActive] = createSignal<"all" | "active" | "inactive">("all");
  const [confirmDelete, setConfirmDelete] = createSignal<Admin | null>(null);
  const [busy, setBusy] = createSignal<string | null>(null);
  const [deleteError, setDeleteError] = createSignal<string | null>(null);

  // Quién es el SUDO actual (para mostrar el botón de sucesión y bloquear el
  // destinatario a sí mismo).
  const [me] = createResource<AdminMe | null>(async () => {
    try {
      return await apiGet<AdminMe>("/admin/me");
    } catch {
      return null;
    }
  });
  const isSudo = () => me()?.sudo ?? false;
  const canViewLogs = () => isSudo() || (me()?.can_view_logs ?? false);

  // Estado del modal de sucesión de SUDO.
  const [showSudo, setShowSudo] = createSignal(false);
  const [sudoTarget, setSudoTarget] = createSignal<string>("");
  const [sudoPassword, setSudoPassword] = createSignal("");
  const [sudoError, setSudoError] = createSignal<string | null>(null);
  const [sudoBusy, setSudoBusy] = createSignal(false);
  const [sudoSuccess, setSudoSuccess] = createSignal(false);

  const openSudoModal = () => {
    setSudoTarget("");
    setSudoPassword("");
    setSudoError(null);
    setSudoSuccess(false);
    setShowSudo(true);
  };

  const handleTransferSudo = async (e: Event) => {
    e.preventDefault();
    if (!sudoTarget()) { setSudoError("Selecciona el administrador destinatario."); return; }
    if (!sudoPassword()) { setSudoError("Confirma tu contraseña para continuar."); return; }

    setSudoBusy(true);
    setSudoError(null);
    try {
      await apiPost("/admin/transfer-sudo", {
        target_id: sudoTarget(),
        password: sudoPassword(),
      });
      setSudoSuccess(true);
      setSudoPassword("");
      setSudoTarget("");
      refetch();
    } catch (err: any) {
      setSudoError(getUserFacingError(err));
    } finally {
      setSudoBusy(false);
    }
  };

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
    <main class="space-y-5">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <PageHeader
        crumbs={[{ label: "Staff" }]}
        title="Staff"
        description="Personal administrativo del sistema"
        actions={
          <div class="flex items-center gap-2">
            <Show when={isSudo()}>
              <button
                onClick={openSudoModal}
                class={`${btnLinkBase} bg-amber-500 text-white hover:bg-amber-600 focus:ring-amber-500/30`}
                title="Ceder el rol de Super Usuario a otro administrador"
              >
                <Icon name="shield" />
                Ceder SUDO
              </button>
            </Show>
            <A
              href="/admin/staff/crear"
              class={`${btnLinkBase} bg-colpsi-blue text-white hover:bg-colpsi-blue-light focus:ring-colpsi-blue/30`}
            >
              <Icon name="plus" />
              Nuevo Administrador
            </A>
          </div>
        }
      />

      {/* ── FILTROS ───────────────────────────────────────────────────────── */}
      <div class="flex flex-col md:flex-row gap-2">
        <div class="relative flex-1 min-w-[220px]">
          <Icon name="search" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <Input
            type="text"
            placeholder="Buscar por usuario o email..."
            value={search()}
            onInput={(e) => setSearch(e.currentTarget.value)}
            class="pl-9"
          />
        </div>
        <div class="inline-flex gap-1 p-1 rounded-md bg-colpsi-bg border border-colpsi-border">
          {(["all", "active", "inactive"] as const).map((s) => (
            <button
              onClick={() => setFilterActive(s)}
              class={`h-8 px-3 rounded-md text-xs font-medium transition-all border ${
                filterActive() === s
                  ? "bg-white text-colpsi-blue border-colpsi-border shadow-sm"
                  : "bg-transparent text-colpsi-muted border-transparent hover:text-colpsi-blue hover:bg-white/60"
              }`}
            >
              {s === "all" ? "Todos" : s === "active" ? "Activos" : "Inactivos"}
            </button>
          ))}
        </div>
      </div>

      {/* ── LISTADO ───────────────────────────────────────────────────────── */}
      <div class="border border-colpsi-border rounded-lg bg-white overflow-hidden">
        <Suspense fallback={
          <div class="space-y-3 p-4">
            <For each={[1, 2, 3]}>{() => <div class="h-14 bg-white animate-pulse rounded-md border border-colpsi-border" />}</For>
          </div>
        }>
          <Show when={!result.loading && list().length === 0}>
            <div class="text-center py-16">
              <p class="text-colpsi-muted font-medium">No hay administradores registrados</p>
              <A href="/admin/staff/crear" class="mt-3 inline-block text-colpsi-blue font-semibold text-sm hover:underline">
                Crear el primero →
              </A>
            </div>
          </Show>

          <Show when={!result.loading && list().length > 0 && filtered().length === 0}>
            <div class="text-center py-14">
              <p class="text-colpsi-muted font-medium">Ningún resultado para los filtros aplicados</p>
            </div>
          </Show>

          <div class="overflow-x-auto">
            <table class="w-full border-collapse">
              <thead>
                <tr>
                  <th class="th-cell">Usuario</th>
                  <th class="th-cell">Rol</th>
                  <th class="th-cell">Permisos</th>
                  <th class="th-cell">Estatus</th>
                  <th class="th-cell">Creado</th>
                  <th class="th-cell text-right">Acciones</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-colpsi-border">
                <For each={filtered()}>
                  {(admin) => {
                    const isBusy = () => busy() === admin.id;
                    const permsCount = countPerms(admin);
                    return (
                      <tr class={`transition-colors hover:bg-colpsi-bg/60 ${admin.is_active ? "" : "opacity-60"}`}>
                        <td class="td-cell min-w-[220px]">
                          <div class="flex items-center gap-2.5">
                            <div class={`w-8 h-8 rounded-md flex items-center justify-center font-bold text-xs uppercase shrink-0 ${
                              admin.is_active ? "bg-colpsi-blue/10 text-colpsi-blue" : "bg-colpsi-bg text-colpsi-muted"
                            }`}>
                              {admin.username.charAt(0)}
                            </div>
                            <div class="min-w-0">
                              <A href={`/admin/staff/${admin.id}`} class="font-medium text-colpsi-text hover:text-colpsi-blue hover:underline truncate block">
                                {admin.username}
                              </A>
                              <p class="text-xs text-colpsi-muted truncate">{admin.email}</p>
                            </div>
                          </div>
                        </td>
                        <td class="td-cell">
                          <Badge tone={admin.role ? "info" : "neutral"}>
                            {roleLabel(admin.role)}
                          </Badge>
                        </td>
                        <td class="td-cell whitespace-nowrap">
                          <Badge tone="neutral" dot={false}>{permsCount}/20 permisos</Badge>
                        </td>
                        <td class="td-cell">
                          <Badge tone={admin.is_active ? "success" : "neutral"}>
                            {admin.is_active ? "Activo" : "Inactivo"}
                          </Badge>
                        </td>
                        <td class="td-cell text-colpsi-muted whitespace-nowrap">
                          <span class="block text-sm font-medium text-colpsi-text">{formatDate(admin.created_at)}</span>
                          <Show when={admin.create_by}>
                            <span class="text-xs text-colpsi-muted">Por {admin.create_by}</span>
                          </Show>
                        </td>
                        <td class="td-cell text-right whitespace-nowrap">
                          <div class="inline-flex items-center gap-1">
                            <Show when={canViewLogs()}>
                              <A
                                href={`/admin/auditoria?actor_id=${admin.id}`}
                                class="h-8 w-8 rounded-md flex items-center justify-center text-colpsi-muted border border-transparent hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors"
                                title="Ver actividad (bitácora)"
                              >
                                <Icon name="fileText" class="w-4 h-4" />
                              </A>
                            </Show>
                            <A
                              href={`/admin/staff/${admin.id}`}
                              class="h-8 w-8 rounded-md flex items-center justify-center text-colpsi-muted border border-transparent hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors"
                              title="Editar"
                            >
                              <Icon name="pencil" class="w-4 h-4" />
                            </A>
                            <button
                              onClick={() => { setDeleteError(null); setConfirmDelete(admin); }}
                              disabled={isBusy()}
                              title="Eliminar"
                              class="h-8 w-8 rounded-md flex items-center justify-center text-colpsi-muted border border-transparent hover:text-colpsi-red hover:bg-red-50 transition-colors disabled:opacity-40"
                            >
                              <Icon name="trash" class="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  }}
                </For>
              </tbody>
            </table>
          </div>

          <Show when={list().length > 0}>
            <p class="text-center text-xs text-colpsi-muted font-medium py-3 border-t border-colpsi-border bg-colpsi-bg">
              Mostrando {filtered().length} de {result()?.total ?? list().length} administradores
            </p>
          </Show>
        </Suspense>
      </div>

      {/* ── MODAL CONFIRMACIÓN BORRADO ─────────────────────────────────── */}
      <Show when={confirmDelete()}>
        {(admin) => (
          <div
            class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
            onClick={(e) => { if (e.target === e.currentTarget) setConfirmDelete(null); }}
          >
            <div class="bg-white rounded-lg shadow-lg p-6 w-full max-w-sm border border-colpsi-border">
              <span class="inline-flex h-10 w-10 items-center justify-center rounded-md bg-red-50 text-colpsi-red mb-3">
                <Icon name="trash" class="w-5 h-5" />
              </span>
              <h2 class="text-base font-semibold text-colpsi-text mb-1">¿Eliminar administrador?</h2>
              <p class="text-colpsi-blue font-semibold text-sm mb-1">{admin().username}</p>
              <p class="text-colpsi-muted text-sm mb-4">Esta acción es irreversible.</p>
              <Show when={deleteError()}>
                <div class="mb-4 p-3 rounded-md bg-red-50 text-red-700 text-xs font-medium border border-red-200">
                  {deleteError()}
                </div>
              </Show>
              <div class="flex gap-2">
                <button
                  onClick={() => setConfirmDelete(null)}
                  class="flex-1 h-9 px-4 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors text-sm"
                >
                  Cancelar
                </button>
                <button
                  onClick={() => handleDelete(admin())}
                  disabled={busy() === admin().id}
                  class="flex-1 h-9 px-4 rounded-md bg-colpsi-red text-white font-semibold hover:opacity-90 transition-colors text-sm disabled:opacity-60"
                >
                  {busy() === admin().id ? "Eliminando..." : "Sí, eliminar"}
                </button>
              </div>
            </div>
          </div>
        )}
      </Show>

    {/* ── MODAL SUCESIÓN SUDO ─────────────────────────────────── */}
      <Show when={showSudo()}>
        <div
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
          onClick={(e) => { if (e.target === e.currentTarget) setShowSudo(false); }}
        >
          <form onSubmit={handleTransferSudo} class="bg-white rounded-lg shadow-lg p-6 w-full max-w-md border border-colpsi-border">
            <span class="inline-flex h-10 w-10 items-center justify-center rounded-md bg-amber-50 text-amber-600 mb-3">
              <Icon name="shield" class="w-5 h-5" />
            </span>
            <h2 class="text-base font-semibold text-colpsi-text mb-1">Transferir el rol de Super Usuario</h2>
            <p class="text-colpsi-muted text-sm mb-5">
              El destinatario pasará a ser el único SUDO del sistema y tú quedarás como administrador normal. Esta acción es grave e irrevocable.
            </p>

            <label class="block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide mb-1">Destinatario</label>
            <select
              value={sudoTarget()}
              onChange={(e) => setSudoTarget(e.currentTarget.value)}
              class="w-full mb-4 h-9 bg-white border border-slate-300 rounded-md px-3 outline-none transition-all text-colpsi-text text-sm focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
            >
              <option value="">— Seleccionar administrador —</option>
              <For each={filtered().filter((a) => a.id !== me()?.id && a.is_active)}>
                {(a) => <option value={a.id}>{a.username} ({a.email})</option>}
              </For>
            </select>

            <label class="block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide mb-1">Contraseña del SUDO actual</label>
            <input
              type="password"
              placeholder="Confirma con tu contraseña..."
              value={sudoPassword()}
              onInput={(e) => setSudoPassword(e.currentTarget.value)}
              class="w-full mb-4 h-9 bg-white border border-slate-300 rounded-md px-3 outline-none transition-all text-colpsi-text text-sm focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
            />

            <Show when={sudoError()}>
              <div class="mb-4 p-3 rounded-md bg-red-50 text-red-700 text-xs font-medium border border-red-200">{sudoError()}</div>
            </Show>
            <Show when={sudoSuccess()}>
              <div class="mb-4 p-3 rounded-md bg-emerald-50 text-emerald-700 text-xs font-medium border border-emerald-200">
                ✓ Rol transferido. El destinatario ya es SUDO. Tu sesión se actualizará al recargar.
              </div>
            </Show>

            <div class="flex gap-2">
              <button
                type="button"
                onClick={() => setShowSudo(false)}
                class="flex-1 h-9 px-4 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors text-sm"
              >
                Cancelar
              </button>
              <button
                type="submit"
                disabled={sudoBusy() || sudoSuccess()}
                class="flex-1 h-9 px-4 rounded-md bg-amber-500 hover:bg-amber-600 disabled:opacity-60 text-white font-semibold transition-colors text-sm"
              >
                {sudoBusy() ? "Transfiriendo..." : "Transferir SUDO"}
              </button>
            </div>
          </form>
        </div>
      </Show>

    </main>
  );
}
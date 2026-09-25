// web/src/components/admin/proyectos/MembersModal.tsx
import { For, Show, createResource, createSignal } from "solid-js";
import { apiGet, apiPost, apiPatch, apiDelete } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Project, ProjectMember, ProjectMemberRole } from "~/types/projects";
import { Icon } from "~/components/admin/ui/icons";

interface StaffAdmin {
  id: string;
  username: string;
  email: string;
  is_active: boolean;
}
interface StaffListResponse {
  data: StaffAdmin[];
}

const ROLE_LABELS: Record<ProjectMemberRole, string> = {
  viewer: "Espectador",
  editor: "Editor",
};

export default function MembersModal(props: {
  project: Project;
  members: ProjectMember[];
  canManage: boolean;
  onClose: () => void;
  reload: () => void;
}) {
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [selected, setSelected] = createSignal<string>("");
  const [role, setRole] = createSignal<ProjectMemberRole>("editor");

  const [admins] = createResource<StaffAdmin[]>(() =>
    apiGet<StaffListResponse>("/admin/list?limit=100").then((r) => r.data)
  );

  const existingIds = () =>
    new Set([props.project.owner_id, ...props.members.map((m) => m.user_admin_id)]);

  const available = () =>
    (admins() ?? []).filter((a) => a.is_active && !existingIds().has(a.id));

  const invite = async () => {
    if (!selected() || busy()) return;
    setBusy(true);
    setError(null);
    try {
      await apiPost(`/admin/projects/${props.project.id}/members`, {
        user_admin_id: selected(),
        role: role(),
      });
      setSelected("");
      props.reload();
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  const changeRole = async (member: ProjectMember, r: ProjectMemberRole) => {
    setBusy(true);
    setError(null);
    try {
      await apiPatch(`/admin/projects/members/${member.id}`, { role: r });
      props.reload();
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  const remove = async (member: ProjectMember) => {
    if (!window.confirm(`¿Quitar a ${member.user?.username ?? "este miembro"} del proyecto?`)) return;
    setBusy(true);
    setError(null);
    try {
      await apiDelete(`/admin/projects/members/${member.id}`);
      props.reload();
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div
      class="fixed inset-0 z-50 flex items-end md:items-center justify-center p-0 md:p-4 bg-black/60 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && !busy() && props.onClose()}
    >
      <div class="bg-white w-full md:max-w-lg md:rounded-lg rounded-t-lg max-h-[88vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
        <div class="flex items-center justify-between px-5 py-3.5 border-b border-colpsi-border">
          <div>
            <h3 class="font-semibold text-colpsi-text">Miembros del proyecto</h3>
            <p class="text-xs text-colpsi-muted mt-0.5">Los miembros son administradores del colegio.</p>
          </div>
          <button onClick={props.onClose} class="inline-flex items-center justify-center h-8 w-8 rounded-md bg-colpsi-bg text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-border/60 transition-colors">
            <Icon name="x" class="w-4 h-4" />
          </button>
        </div>

        <div class="p-5 space-y-4">
          <Show when={error()}>
            <div class="p-3 rounded-md bg-red-50 text-red-700 text-sm font-medium border border-red-200">{error()}</div>
          </Show>

          <div class="rounded-md border border-colpsi-border divide-y divide-colpsi-border">
            <div class="flex items-center justify-between px-4 py-3">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-md bg-colpsi-blue/10 text-colpsi-blue flex items-center justify-center font-semibold">Ψ</div>
                <div>
                  <span class="font-medium text-sm text-colpsi-text flex items-center gap-2">
                    {props.project.owner?.username ?? props.project.create_by}
                    <span class="text-[10px] font-medium uppercase text-blue-700 bg-blue-50 px-2 py-0.5 rounded border border-blue-200">Dueño</span>
                  </span>
                  <span class="text-xs text-colpsi-muted">{props.project.owner?.email}</span>
                </div>
              </div>
            </div>

            <For each={props.members}>
              {(m) => (
                <div class="flex items-center justify-between gap-3 px-4 py-3">
                  <div class="flex items-center gap-3 min-w-0">
                    <div class="w-9 h-9 rounded-md bg-colpsi-bg text-colpsi-muted flex items-center justify-center font-medium shrink-0">
                      {(m.user?.username ?? "?").slice(0, 2).toUpperCase()}
                    </div>
                    <div class="min-w-0">
                      <span class="font-medium text-sm text-colpsi-text block truncate">{m.user?.username}</span>
                      <span class="text-xs text-colpsi-muted truncate block">{m.user?.email}</span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 shrink-0">
                    <Show when={props.canManage}>
                      <button
                        disabled={busy()}
                        onClick={() => changeRole(m, m.role === "editor" ? "viewer" : "editor")}
                        class={`text-[11px] font-medium uppercase tracking-wide px-2.5 py-1 rounded-md border transition-colors disabled:opacity-60 ${
                          m.role === "editor"
                            ? "bg-amber-50 text-amber-700 border-amber-200"
                            : "bg-slate-100 text-slate-600 border-slate-200 hover:bg-slate-200"
                        }`}
                        title="Cambiar rol"
                      >
                        {ROLE_LABELS[m.role]}
                      </button>
                      <button
                        disabled={busy()}
                        onClick={() => remove(m)}
                        class="inline-flex items-center justify-center h-8 w-8 rounded-md bg-white text-colpsi-red border border-red-200 hover:bg-red-50 disabled:opacity-60 transition-colors"
                        title="Quitar miembro"
                      >
                        <Icon name="x" class="w-3.5 h-3.5" />
                      </button>
                    </Show>
                    <Show when={!props.canManage}>
                      <span class="text-[11px] font-medium uppercase tracking-wide px-2.5 py-1 rounded-md bg-slate-100 text-slate-600 border border-slate-200">
                        {ROLE_LABELS[m.role]}
                      </span>
                    </Show>
                  </div>
                </div>
              )}
            </For>
          </div>

          <Show when={props.canManage}>
            <div class="rounded-md border border-dashed border-colpsi-border p-3.5">
              <p class="text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted mb-2.5">Invitar administrador</p>
              <div class="flex flex-col sm:flex-row gap-2">
                <select
                  value={selected()}
                  onChange={(e) => setSelected(e.currentTarget.value)}
                  class="flex-grow h-9 rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors"
                >
                  <option value="">Selecciona un admin...</option>
                  <For each={available()}>
                    {(a) => <option value={a.id}>{a.username} — {a.email}</option>}
                  </For>
                </select>
                <select
                  value={role()}
                  onChange={(e) => setRole(e.currentTarget.value as ProjectMemberRole)}
                  class="h-9 rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors"
                >
                  <option value="editor">Editor</option>
                  <option value="viewer">Espectador</option>
                </select>
                <button
                  disabled={!selected() || busy()}
                  onClick={invite}
                  class="h-9 px-5 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold text-sm disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {busy() ? "..." : "Invitar"}
                </button>
              </div>
            </div>
          </Show>
        </div>
      </div>
    </div>
  );
}
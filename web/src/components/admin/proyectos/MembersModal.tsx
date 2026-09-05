// web/src/components/admin/proyectos/MembersModal.tsx
import { For, Show, createResource, createSignal } from "solid-js";
import { apiGet, apiPost, apiPatch, apiDelete } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Project, ProjectMember, ProjectMemberRole } from "~/types/projects";

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
      <div class="bg-white w-full md:max-w-lg md:rounded-3xl rounded-t-3xl max-h-[88vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <div>
            <h3 class="font-black text-colpsi-blue text-lg">Miembros del proyecto</h3>
            <p class="text-xs text-gray-400">Los miembros son administradores del colegio.</p>
          </div>
          <button onClick={props.onClose} class="w-9 h-9 rounded-full bg-gray-100 text-gray-500 font-black hover:bg-gray-200">
            ✕
          </button>
        </div>

        <div class="p-6 space-y-4">
          <Show when={error()}>
            <div class="p-3 rounded-xl bg-red-50 text-red-700 text-sm font-bold border-l-4 border-red-500">{error()}</div>
          </Show>

          <div class="rounded-2xl border-2 border-gray-100 divide-y divide-gray-100">
            <div class="flex items-center justify-between px-4 py-3">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center font-black">Ψ</div>
                <div>
                  <span class="font-black text-sm text-gray-800 flex items-center gap-2">
                    {props.project.owner?.username ?? props.project.create_by}
                    <span class="text-[10px] font-black uppercase text-blue-600 bg-blue-50 px-2 py-0.5 rounded-full border border-blue-200">Dueño</span>
                  </span>
                  <span class="text-xs text-gray-400">{props.project.owner?.email}</span>
                </div>
              </div>
            </div>

            <For each={props.members}>
              {(m) => (
                <div class="flex items-center justify-between gap-3 px-4 py-3">
                  <div class="flex items-center gap-3 min-w-0">
                    <div class="w-9 h-9 rounded-full bg-gray-100 text-gray-600 flex items-center justify-center font-black shrink-0">
                      {(m.user?.username ?? "?").slice(0, 2).toUpperCase()}
                    </div>
                    <div class="min-w-0">
                      <span class="font-black text-sm text-gray-800 block truncate">{m.user?.username}</span>
                      <span class="text-xs text-gray-400 truncate block">{m.user?.email}</span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 shrink-0">
                    <Show when={props.canManage}>
                      <button
                        disabled={busy()}
                        onClick={() => changeRole(m, m.role === "editor" ? "viewer" : "editor")}
                        class={`text-[11px] font-black uppercase tracking-wide px-3 py-1.5 rounded-lg border transition-colors disabled:opacity-60 ${
                          m.role === "editor"
                            ? "bg-amber-100 text-amber-700 border-amber-300"
                            : "bg-gray-100 text-gray-500 border-gray-200 hover:bg-gray-200"
                        }`}
                        title="Cambiar rol"
                      >
                        {ROLE_LABELS[m.role]}
                      </button>
                      <button
                        disabled={busy()}
                        onClick={() => remove(m)}
                        class="w-8 h-8 rounded-lg bg-red-50 text-red-500 font-black hover:bg-red-100 disabled:opacity-60"
                        title="Quitar miembro"
                      >
                        ✕
                      </button>
                    </Show>
                    <Show when={!props.canManage}>
                      <span class="text-[11px] font-black uppercase tracking-wide px-3 py-1.5 rounded-lg bg-gray-100 text-gray-500 border border-gray-200">
                        {ROLE_LABELS[m.role]}
                      </span>
                    </Show>
                  </div>
                </div>
              )}
            </For>
          </div>

          <Show when={props.canManage}>
            <div class="rounded-2xl border-2 border-dashed border-gray-200 p-4">
              <p class="text-xs font-black uppercase tracking-widest text-gray-400 mb-3">Invitar administrador</p>
              <div class="flex flex-col sm:flex-row gap-2">
                <select
                  value={selected()}
                  onChange={(e) => setSelected(e.currentTarget.value)}
                  class="flex-grow bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-3 py-2.5 text-sm outline-none"
                >
                  <option value="">Selecciona un admin…</option>
                  <For each={available()}>
                    {(a) => <option value={a.id}>{a.username} — {a.email}</option>}
                  </For>
                </select>
                <select
                  value={role()}
                  onChange={(e) => setRole(e.currentTarget.value as ProjectMemberRole)}
                  class="bg-white border-2 border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none"
                >
                  <option value="editor">Editor</option>
                  <option value="viewer">Espectador</option>
                </select>
                <button
                  disabled={!selected() || busy()}
                  onClick={invite}
                  class="bg-blue-800 hover:bg-blue-900 text-white font-black px-5 py-2.5 rounded-xl disabled:opacity-50 disabled:cursor-not-allowed"
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
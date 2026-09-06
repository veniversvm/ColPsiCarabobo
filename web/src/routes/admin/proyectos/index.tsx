// web/src/routes/admin/proyectos/index.tsx
import { For, Show, Suspense, createResource, ErrorBoundary, createSignal } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { apiGet, apiDelete } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Project } from "~/types/projects";
import { canManageProject, ProjectMemberRole } from "~/types/projects";
import ConfirmModal from "~/components/admin/proyectos/ConfirmModal";

const ROLE_LABELS: Record<string, string> = {
  viewer: "Espectador",
  editor: "Editor",
  owner: "Dueño",
  master: "Master",
};

function RoleBadge(props: { project: Project }) {
  const p = () => props.project;
  let label = "Espectador";
  let cls = "bg-gray-100 text-gray-600 border-gray-300";
  if (p().is_master) {
    label = "Master";
    cls = "bg-purple-100 text-purple-700 border-purple-300";
  } else if (p().is_owner) {
    label = "Dueño";
    cls = "bg-blue-100 text-blue-700 border-blue-300";
  } else if (p().my_role) {
    label = ROLE_LABELS[p().my_role as ProjectMemberRole] ?? "Espectador";
    cls = p().my_role === "editor" ? "bg-amber-100 text-amber-700 border-amber-300" : "bg-gray-100 text-gray-600 border-gray-300";
  }
  return (
    <span class={`px-2.5 py-0.5 rounded-full text-[11px] font-black uppercase tracking-wider border ${cls}`}>{label}</span>
  );
}

function ProjectCard(props: { project: Project; onDelete: (p: Project) => void }) {
  const p = () => props.project;
  return (
    <div class="bg-white rounded-3xl border border-colpsi-border shadow-premium overflow-hidden flex flex-col hover:shadow-xl hover:-translate-y-0.5 transition-all">
      <A href={`/admin/proyectos/${p().id}`} class="flex flex-col flex-grow p-6 text-left group">
        <div class="flex items-start justify-between gap-3">
          <h3 class="font-black text-lg text-colpsi-blue leading-snug group-hover:text-blue-700 transition-colors line-clamp-2">
            {p().name}
          </h3>
          <RoleBadge project={p()} />
        </div>
        <Show when={p().description}>
          <p class="mt-2 text-sm text-gray-500 line-clamp-2">{p().description}</p>
        </Show>
        <div class="mt-5 pt-4 border-t border-colpsi-border flex items-center gap-5 text-xs font-bold text-gray-400">
          <span class="flex items-center gap-1">👥 {p().member_count}</span>
          <span class="flex items-center gap-1">🃏 {p().card_count} tarjetas</span>
          <span class="ml-auto text-[11px]">por {p().create_by || "—"}</span>
        </div>
      </A>
      <div class="px-6 pb-5 flex items-center justify-between">
        <span class="text-[11px] text-gray-300">{new Date(p().created_at).toLocaleDateString("es-VE")}</span>
        <Show when={canManageProject(p())}>
          <button
            onClick={() => props.onDelete(p())}
            class="text-xs font-black text-red-400 hover:text-red-600 transition-colors"
          >
            Eliminar
          </button>
        </Show>
      </div>
    </div>
  );
}

export default function ProyectosIndex() {
  const navigate = useNavigate();
  const [deleting, setDeleting] = createSignal<Project | null>(null);
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const [projects] = createResource<Project[]>(
    () => apiGet<{ data: Project[] }>("/admin/projects").then((r) => r.data)
  );

  const doDelete = async () => {
    const p = deleting();
    if (!p || busy()) return;
    setBusy(true);
    setError(null);
    try {
      await apiDelete(`/admin/projects/${p.id}`);
      setDeleting(null);
      projects.mutate((prev) => (prev ?? []).filter((x) => x.id !== p.id));
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div class="pb-20">
      <div class="flex flex-col gap-4 mb-8">
        <span class="text-xs font-black text-blue-400 uppercase tracking-widest">Panel administrativo</span>
        <div class="flex items-center justify-between flex-wrap gap-4">
          <div>
            <h1 class="text-3xl md:text-4xl font-black text-colpsi-blue">Proyectos</h1>
            <p class="text-sm text-gray-500 mt-1">Tableros Kanban colaborativos del colegio.</p>
          </div>
          <button
            onClick={() => navigate("/admin/proyectos/crear")}
            class="bg-blue-800 hover:bg-blue-900 text-white font-black px-6 py-4 rounded-2xl shadow-xl hover:scale-105 active:scale-95 transition-all"
          >
            + Nuevo Proyecto
          </button>
        </div>
      </div>

      <Show when={error()}>
        <div class="mb-6 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500 shadow-sm">
          {error()}
        </div>
      </Show>

      <ErrorBoundary fallback={<p class="text-sm text-red-500">No se pudieron cargar los proyectos.</p>}>
        <Suspense
          fallback={
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
              <For each={[1, 2, 3]}>
                {() => <div class="h-48 bg-white rounded-3xl animate-pulse shadow-premium" />}
              </For>
            </div>
          }
        >
          <Show when={projects() && projects()!.length === 0} fallback={null}>
            <div class="text-center py-20 bg-white rounded-3xl border-2 border-dashed border-gray-200">
              <div class="text-5xl mb-3">📋</div>
              <p class="font-black text-colpsi-blue text-lg">Aún no hay proyectos</p>
              <p class="text-sm text-gray-500 mt-1">Crea el primero para empezar a organizar el trabajo del colegio.</p>
            </div>
          </Show>
          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
            <For each={projects()}>
              {(p) => <ProjectCard project={p} onDelete={setDeleting} />}
            </For>
          </div>
        </Suspense>
      </ErrorBoundary>

      <Show when={deleting()}>
        <ConfirmModal
          title="Eliminar proyecto"
          message={`¿Seguro que quieres eliminar «${deleting()!.name}»? Se borrarán todas sus columnas, tarjetas y notas de forma definitiva.`}
          confirmLabel="Eliminar"
          danger
          busy={busy()}
          onConfirm={doDelete}
          onClose={() => !busy() && setDeleting(null)}
        />
      </Show>
    </div>
  );
}
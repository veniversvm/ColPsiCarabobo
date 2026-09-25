// web/src/routes/admin/proyectos/index.tsx
import { For, Show, Suspense, createResource, ErrorBoundary, createSignal } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { apiGet, apiDelete } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Project } from "~/types/projects";
import { canManageProject, ProjectMemberRole } from "~/types/projects";
import ConfirmModal from "~/components/admin/proyectos/ConfirmModal";
import { PageHeader } from "~/components/admin/ui/PageHeader";
import { Button } from "~/components/admin/ui/Button";
import { Icon } from "~/components/admin/ui/icons";

const ROLE_LABELS: Record<string, string> = {
  viewer: "Espectador",
  editor: "Editor",
  owner: "Dueño",
  master: "Master",
};

function RoleBadge(props: { project: Project }) {
  const p = () => props.project;
  let label = "Espectador";
  let cls = "bg-slate-100 text-slate-600 border-slate-200";
  if (p().is_master) {
    label = "Master";
    cls = "bg-purple-50 text-purple-700 border-purple-200";
  } else if (p().is_owner) {
    label = "Dueño";
    cls = "bg-blue-50 text-blue-700 border-blue-200";
  } else if (p().my_role) {
    label = ROLE_LABELS[p().my_role as ProjectMemberRole] ?? "Espectador";
    cls = p().my_role === "editor" ? "bg-amber-50 text-amber-700 border-amber-200" : "bg-slate-100 text-slate-600 border-slate-200";
  }
  return (
    <span class={`inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium border whitespace-nowrap ${cls}`}>{label}</span>
  );
}

function ProjectCard(props: { project: Project; onDelete: (p: Project) => void }) {
  const p = () => props.project;
  return (
    <div class="bg-white rounded-lg border border-colpsi-border overflow-hidden flex flex-col hover:border-colpsi-blue/40 transition-colors">
      <A href={`/admin/proyectos/${p().id}`} class="flex flex-col flex-grow p-4 text-left group">
        <div class="flex items-start justify-between gap-3">
          <h3 class="font-semibold text-base text-colpsi-text leading-snug group-hover:text-colpsi-blue transition-colors line-clamp-2">
            {p().name}
          </h3>
          <RoleBadge project={p()} />
        </div>
        <Show when={p().description}>
          <p class="mt-1.5 text-sm text-colpsi-muted line-clamp-2">{p().description}</p>
        </Show>
        <div class="mt-4 pt-3 border-t border-colpsi-border flex items-center gap-4 text-xs font-medium text-colpsi-muted">
          <span class="inline-flex items-center gap-1.5"><Icon name="users" class="w-3.5 h-3.5" /> {p().member_count}</span>
          <span class="inline-flex items-center gap-1.5"><Icon name="kanban" class="w-3.5 h-3.5" /> {p().card_count} tarjetas</span>
          <span class="ml-auto text-[11px] text-colpsi-muted/70">por {p().create_by || "—"}</span>
        </div>
      </A>
      <div class="px-4 pb-3.5 flex items-center justify-between">
        <span class="text-[11px] text-colpsi-muted/70">{new Date(p().created_at).toLocaleDateString("es-VE")}</span>
        <Show when={canManageProject(p())}>
          <button
            onClick={() => props.onDelete(p())}
            class="inline-flex items-center gap-1.5 text-xs font-semibold text-colpsi-red/80 hover:text-colpsi-red transition-colors"
          >
            <Icon name="trash" class="w-3.5 h-3.5" />
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
    <main class="space-y-4 pb-12">
      <PageHeader
        crumbs={[{ label: "Proyectos" }]}
        title="Proyectos"
        description="Tableros Kanban colaborativos del colegio."
        actions={
          <Button variant="primary" size="md" onClick={() => navigate("/admin/proyectos/crear")}>
            <Icon name="plus" />
            Nuevo Proyecto
          </Button>
        }
      />

      <Show when={error()}>
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">
          {error()}
        </div>
      </Show>

      <ErrorBoundary fallback={<p class="text-sm text-red-500">No se pudieron cargar los proyectos.</p>}>
        <Suspense
          fallback={
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              <For each={[1, 2, 3]}>
                {() => <div class="h-44 bg-white rounded-lg animate-pulse border border-colpsi-border" />}
              </For>
            </div>
          }
        >
          <Show when={projects() && projects()!.length === 0} fallback={null}>
            <div class="text-center py-16 bg-white rounded-lg border border-dashed border-colpsi-border">
              <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
                <Icon name="kanban" class="w-6 h-6" />
              </span>
              <p class="font-semibold text-colpsi-text">Aún no hay proyectos</p>
              <p class="text-sm text-colpsi-muted mt-1">Crea el primero para empezar a organizar el trabajo del colegio.</p>
            </div>
          </Show>
          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
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
    </main>
  );
}
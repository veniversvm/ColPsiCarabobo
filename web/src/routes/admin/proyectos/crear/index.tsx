// web/src/routes/admin/proyectos/crear/index.tsx
import { Show, createSignal } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Icon } from "~/components/admin/ui/icons";

const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

export default function CrearProyecto() {
  const navigate = useNavigate();
  const [name, setName] = createSignal("");
  const [description, setDescription] = createSignal("");
  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const canSubmit = () => name().trim().length > 0 && !saving();

  const submit = async () => {
    if (!canSubmit()) return;
    setSaving(true);
    setError(null);
    try {
      const idempotencyKey = crypto.randomUUID();
      const project = await apiPost<{ id: string }>("/admin/projects", {
        name: name().trim(),
        description: description().trim(),
      }, { headers: { "X-Idempotency-Key": idempotencyKey } });
      navigate(`/admin/proyectos/${project.id}`);
    } catch (err) {
      setError(getUserFacingError(err));
      window.scrollTo(0, 0);
      setSaving(false);
    }
  };

  return (
    <main class="space-y-4 pb-12 max-w-2xl mx-auto">
      <div class="flex items-center gap-3 pb-4 border-b border-colpsi-border">
        <button
          onClick={() => navigate("/admin/proyectos")}
          class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors"
          title="Volver"
        >
          <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
        </button>
        <div>
          <h1 class="text-lg font-semibold text-colpsi-text">Nuevo Proyecto</h1>
          <p class="text-sm text-colpsi-muted mt-0.5">Un tablero Kanban para organizar el trabajo del colegio.</p>
        </div>
      </div>

      <Show when={error()}>
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">
          {error()}
        </div>
      </Show>

      <div class="bg-white rounded-lg border border-colpsi-border p-5 space-y-4">
        <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3">
          Información del proyecto
        </h2>

        <div>
          <label class={labelClass}>Nombre del proyecto *</label>
          <input
            value={name()}
            onInput={(e) => setName(e.currentTarget.value)}
            placeholder="Ej. Organización de la Convención 2026"
            maxLength={120}
            class={IC}
          />
          <p class="text-xs text-colpsi-muted mt-1 text-right">{name().length}/120</p>
        </div>

        <div>
          <label class={labelClass}>Descripción</label>
          <textarea
            value={description()}
            onInput={(e) => setDescription(e.currentTarget.value)}
            placeholder="¿Qué se quiere lograr con este proyecto?"
            maxLength={500}
            rows={4}
            class={`${IC} resize-none min-h-24`}
          />
          <p class="text-xs text-colpsi-muted mt-1 text-right">{description().length}/500</p>
        </div>

        <div class="rounded-md bg-sky-50/50 border border-sky-100 p-3 text-sm text-sky-900 flex items-start gap-2.5">
          <Icon name="info" class="w-4 h-4 text-sky-600 shrink-0 mt-0.5" />
          <div>
            <p class="font-semibold mb-0.5">Al crear el proyecto se añaden 3 columnas por defecto:</p>
            <p class="text-sky-800">«Por hacer» · «En progreso» · «Hecho»</p>
            <p class="text-xs text-sky-700 mt-1.5">Podrás invitar a otros administradores como miembros (Espectador o Editor).</p>
          </div>
        </div>

        <div class="flex gap-2 pt-4 border-t border-colpsi-border">
          <button
            onClick={() => navigate("/admin/proyectos")}
            class="h-10 px-4 rounded-md bg-white text-colpsi-text border border-colpsi-border font-medium hover:bg-colpsi-bg transition-colors text-sm"
          >
            Cancelar
          </button>
          <button
            onClick={submit}
            disabled={!canSubmit()}
            class="flex-grow inline-flex items-center justify-center gap-2 h-10 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-sm"
          >
            <Icon name="plus" />
            {saving() ? "Creando..." : "Crear Proyecto"}
          </button>
        </div>
      </div>
    </main>
  );
}
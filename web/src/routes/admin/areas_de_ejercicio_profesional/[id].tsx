// web/src/routes/admin/areas_de_ejercicio_profesional/[id].tsx
import { createResource, createSignal, Show } from "solid-js";
import { useNavigate, useParams } from "@solidjs/router";
import { apiGet, apiPatch } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Icon } from "~/components/admin/ui/icons";
import { Badge } from "~/components/admin/ui/Badge";

const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

// ─── INTERFAZ ACTUALIZADA ─────────────────────────────────────────────────────
interface WorkArea {
  id: number;
  name: string;
  description: string;
  active: boolean;
  created_at: string;
  create_by: string;
}

export default function AdminEditarAreaEjercicioPage() {
  const params = useParams<{ id: string }>();
  const navigate = useNavigate();

  // Recurso para obtener los datos del área
  const [workArea] = createResource(
    () => params.id,
    async (id) => {
      try {
        // Mantenemos el endpoint técnico /specialties del backend
        return await apiGet<WorkArea>(`/specialties/${id}`);
      } catch (err: any) {
        console.error("[edit area] error:", err?.status, err?.message);
        return null;
      }
    }
  );

  const [name, setName] = createSignal("");
  const [description, setDescription] = createSignal("");
  const [active, setActive] = createSignal(true);
  const [initialized, setInitialized] = createSignal(false);

  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [success, setSuccess] = createSignal(false);

  // Sincroniza los datos de la DB con el estado del formulario local
  const initForm = (wa: WorkArea) => {
    if (initialized()) return;
    setName(wa.name ?? "");
    setDescription(wa.description ?? "");
    setActive(wa.active ?? true);
    setInitialized(true);
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!name().trim()) { setError("El nombre del área es obligatorio."); return; }

    setSaving(true);
    setError(null);
    setSuccess(false);

    try {
      await apiPatch(`/admin/specialties/${params.id}`, {
        name: name().trim(),
        description: description().trim(),
        active: active(),
      });

      setSuccess(true);
      window.scrollTo({ top: 0, behavior: "smooth" });
      // Redirigir al catálogo principal
      setTimeout(() => navigate("/admin/areas_de_ejercicio_profesional"), 1500);
    } catch (err: any) {
      setError(getUserFacingError(err));
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <main class="space-y-4 pb-12 max-w-3xl mx-auto">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex items-center gap-3 pb-4 border-b border-colpsi-border">
        <button
          onClick={() => navigate(-1)}
          class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors"
          title="Volver"
        >
          <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
        </button>
        <div class="flex-1 min-w-0">
          <h1 class="text-lg font-semibold text-colpsi-text">Editar Área de Ejercicio</h1>
          <p class="text-sm text-colpsi-muted mt-0.5 truncate">
            {workArea.loading ? "Consultando registro..." : workArea()?.name ?? "Cargando..."}
          </p>
        </div>
        <Show when={workArea()}>
          {(wa) => (
            <Badge tone={wa().active ? "success" : "neutral"}>
              {wa().active ? "Activa" : "Inactiva"}
            </Badge>
          )}
        </Show>
      </div>

      {/* ── ALERTS ───────────────────────────────────────────────────────── */}
      <Show when={error()}>
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">
          {error()}
        </div>
      </Show>
      <Show when={success()}>
        <div class="p-3 rounded-md bg-emerald-50 text-emerald-700 border border-emerald-200 text-sm font-medium">
          Cambios aplicados con éxito. Sincronizando catálogo...
        </div>
      </Show>

      {/* ── FORMULARIO ────────────────────────────────────────────────────── */}
      <Show 
        when={!workArea.loading} 
        fallback={<div class="h-64 bg-white animate-pulse rounded-lg border border-colpsi-border" />}
      >
        <Show when={workArea()}>
          {(wa) => {
            initForm(wa());
            return (
              <form onSubmit={handleSubmit} class="space-y-4">

                <section class="bg-white rounded-lg p-5 border border-colpsi-border space-y-4">
                  <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3">
                    Configuración del Área
                  </h2>

                  {/* Nombre del Área */}
                  <div>
                    <label class={labelClass}>Nombre de la Disciplina / Área</label>
                    <input
                      type="text"
                      required
                      maxLength={100}
                      value={name()}
                      onInput={(e) => setName(e.currentTarget.value)}
                      class={IC}
                      placeholder="Ej: Psicología Clínica"
                    />
                    <div class="flex justify-end mt-1.5">
                       <p class="text-xs text-colpsi-muted font-medium">{name().length}/100</p>
                    </div>
                  </div>

                  {/* Descripción */}
                  <div>
                    <label class={labelClass}>Descripción del Campo de Acción</label>
                    <textarea
                      rows={5}
                      maxLength={500}
                      value={description()}
                      onInput={(e) => setDescription(e.currentTarget.value)}
                      class={`${IC} resize-none min-h-32 leading-relaxed`}
                      placeholder="Define brevemente el alcance de esta área..."
                    />
                    <div class="flex justify-end mt-1.5">
                       <p class="text-xs text-colpsi-muted font-medium">{description().length}/500</p>
                    </div>
                  </div>

                  {/* Estado de Activación */}
                  <div>
                    <label class={labelClass}>Estado en el Sistema</label>
                    <div class="grid grid-cols-2 gap-2 mt-1">
                      <button
                        type="button"
                        onClick={() => setActive(true)}
                        class={`h-10 rounded-md text-xs font-semibold border transition-all ${
                          active() === true
                            ? "bg-emerald-600 text-white border-emerald-600"
                            : "bg-white text-colpsi-muted border-colpsi-border hover:border-emerald-300"
                        }`}
                      >
                        {active() === true ? "Área Activa" : "Activar"}
                      </button>
                      <button
                        type="button"
                        onClick={() => setActive(false)}
                        class={`h-10 rounded-md text-xs font-semibold border transition-all ${
                          active() === false
                            ? "bg-slate-600 text-white border-slate-600"
                            : "bg-white text-colpsi-muted border-colpsi-border hover:border-slate-300"
                        }`}
                      >
                        {active() === false ? "Inactiva" : "Desactivar"}
                      </button>
                    </div>
                    <p class="text-[11px] text-colpsi-muted mt-2 ml-1 leading-relaxed">
                      {active()
                        ? "Esta área está habilitada para ser seleccionada por psicólogos y filtrada en el directorio público."
                        : "Esta área dejará de aparecer en los buscadores. Los psicólogos que ya la tengan asignada mantendrán el registro pero no será público."}
                    </p>
                  </div>

                  {/* Auditoría */}
                  <Show when={wa().create_by}>
                    <div class="pt-4 border-t border-colpsi-border">
                      <div class="bg-colpsi-bg rounded-md p-3 border border-colpsi-border flex items-center justify-between gap-4 flex-wrap">
                        <p class="text-[10px] font-semibold text-colpsi-muted uppercase tracking-wide">Registro de Control</p>
                        <p class="text-[11px] text-colpsi-muted font-medium">
                          Creado por <span class="font-semibold text-colpsi-blue">{wa().create_by}</span> el {new Date(wa().created_at).toLocaleDateString("es-VE")}
                        </p>
                      </div>
                    </div>
                  </Show>
                </section>

                {/* ── BOTONES DE ACCIÓN ────────────────────────────────────── */}
                <div class="sticky bottom-4 z-50 flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => navigate(-1)}
                    class="h-10 px-4 rounded-md border border-colpsi-border bg-white text-sm font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    disabled={saving()}
                    class="inline-flex items-center justify-center gap-2 h-10 px-5 rounded-md bg-colpsi-blue text-white text-sm font-semibold transition-colors hover:bg-colpsi-blue-light disabled:opacity-60"
                  >
                    <Show when={saving()} fallback={<><Icon name="check" /><span>Guardar Cambios</span></>}>
                       <span class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                       <span>Guardando...</span>
                    </Show>
                  </button>
                </div>

              </form>
            );
          }}
        </Show>

        <Show when={!workArea.loading && workArea() === null}>
           <div class="text-center py-16 bg-white rounded-lg border border-colpsi-border">
              <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
                <Icon name="tag" class="w-6 h-6" />
              </span>
              <h2 class="text-base font-semibold text-colpsi-text mb-1">Área no encontrada</h2>
              <p class="text-sm text-colpsi-muted mb-5">El registro solicitado no existe o fue eliminado permanentemente.</p>
              <button
                onClick={() => navigate("/admin/areas_de_ejercicio_profesional")}
                class="inline-flex items-center gap-2 h-9 px-4 rounded-md bg-colpsi-blue text-white font-semibold hover:bg-colpsi-blue-light transition-colors text-sm"
              >
                Volver al Catálogo
              </button>
           </div>
        </Show>
      </Show>

    </main>
  );
}
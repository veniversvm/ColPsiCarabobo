// web/src/routes/admin/areas_de_ejercicio_profesional/[id].tsx
import { createResource, createSignal, Show } from "solid-js";
import { useNavigate, useParams } from "@solidjs/router";
import { apiGet, apiPatch } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Animate } from "~/components/ui/Motion";

const IC = "w-full bg-white border-2 border-gray-100 focus:border-blue-500 rounded-2xl px-5 py-3.5 outline-none transition-all text-gray-800 text-sm shadow-sm";
const labelClass = "block text-[10px] font-black text-gray-400 uppercase tracking-[0.2em] ml-2 mb-2";

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
    <main class="pb-28 max-w-3xl mx-auto font-sans">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex items-center gap-6 mb-10 bg-white p-8 rounded-[2.5rem] shadow-premium border border-gray-100">
        <button
          onClick={() => navigate(-1)}
          class="w-12 h-12 bg-gray-50 hover:bg-blue-50 text-blue-900 rounded-2xl font-black flex items-center justify-center transition-all flex-shrink-0 border-2 border-transparent hover:border-blue-100 shadow-sm"
          title="Volver"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div class="flex-1 min-w-0">
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
            Editar Área de Ejercicio
          </h1>
          <p class="text-gray-400 text-sm mt-1 font-medium truncate">
            {workArea.loading ? "Consultando registro..." : workArea()?.name ?? "Cargando..."}
          </p>
        </div>
        <Show when={workArea()}>
          {(wa) => (
            <span class={`text-[10px] font-black px-4 py-2 rounded-xl uppercase tracking-widest flex-shrink-0 border-2 ${
              wa().active 
                ? "bg-emerald-50 text-emerald-600 border-emerald-100" 
                : "bg-gray-50 text-gray-500 border-gray-100"
            }`}>
              {wa().active ? "Activa" : "Inactiva"}
            </span>
          )}
        </Show>
      </div>

      {/* ── ALERTS ───────────────────────────────────────────────────────── */}
      <Show when={error()}>
        <Animate variant="slide-top" class="mb-8 p-5 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-2 border-red-100 flex items-center gap-3">
          <span class="text-xl">⚠️</span> {error()}
        </Animate>
      </Show>
      <Show when={success()}>
        <Animate variant="slide-top" class="mb-8 p-5 rounded-2xl bg-emerald-50 text-emerald-800 font-bold text-sm border-2 border-emerald-100 flex items-center gap-3">
          <span class="text-xl">✅</span> Cambios aplicados con éxito. Sincronizando catálogo...
        </Animate>
      </Show>

      {/* ── FORMULARIO ────────────────────────────────────────────────────── */}
      <Show 
        when={!workArea.loading} 
        fallback={<div class="h-96 bg-white animate-pulse rounded-[2.5rem] border border-gray-100" />}
      >
        <Show when={workArea()}>
          {(wa) => {
            initForm(wa());
            return (
              <form onSubmit={handleSubmit} class="space-y-8">

                <section class="bg-white rounded-[2.5rem] p-8 md:p-10 shadow-premium border border-gray-100 space-y-8">
                  <div class="border-l-4 border-colpsi-yellow pl-4">
                    <h2 class="text-sm font-black text-blue-900 uppercase tracking-widest">
                      Configuración del Área
                    </h2>
                  </div>

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
                    <div class="flex justify-end mt-2">
                       <p class="text-[10px] text-gray-400 font-bold uppercase tracking-wider">{name().length}/100</p>
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
                      class={`${IC} resize-none leading-relaxed`}
                      placeholder="Define brevemente el alcance de esta área..."
                    />
                    <div class="flex justify-end mt-2">
                       <p class="text-[10px] text-gray-400 font-bold uppercase tracking-wider">{description().length}/500</p>
                    </div>
                  </div>

                  {/* Estado de Activación */}
                  <div>
                    <label class={labelClass}>Estado en el Sistema</label>
                    <div class="grid grid-cols-2 gap-4 mt-2">
                      <button
                        type="button"
                        onClick={() => setActive(true)}
                        class={`py-4 rounded-2xl text-xs font-black uppercase tracking-widest border-2 transition-all ${
                          active() === true
                            ? "bg-emerald-600 text-white border-emerald-600 shadow-lg shadow-emerald-200"
                            : "bg-white text-gray-400 border-gray-100 hover:border-emerald-200"
                        }`}
                      >
                        {active() === true ? "✓ Área Activa" : "Activar"}
                      </button>
                      <button
                        type="button"
                        onClick={() => setActive(false)}
                        class={`py-4 rounded-2xl text-xs font-black uppercase tracking-widest border-2 transition-all ${
                          active() === false
                            ? "bg-gray-600 text-white border-gray-600 shadow-lg shadow-gray-200"
                            : "bg-white text-gray-400 border-gray-100 hover:border-red-200"
                        }`}
                      >
                        {active() === false ? "✕ Inactiva" : "Desactivar"}
                      </button>
                    </div>
                    <p class="text-[10px] text-gray-400 mt-4 italic leading-relaxed px-2">
                      {active()
                        ? "Esta área está habilitada para ser seleccionada por psicólogos y filtrada en el directorio público."
                        : "Esta área dejará de aparecer en los buscadores. Los psicólogos que ya la tengan asignada mantendrán el registro pero no será público."}
                    </p>
                  </div>

                  {/* Auditoría */}
                  <Show when={wa().create_by}>
                    <div class="pt-6 border-t border-gray-50 mt-4">
                      <div class="bg-gray-50/50 rounded-2xl p-5 border border-gray-100 flex items-center justify-between">
                        <p class="text-[10px] text-gray-400 font-bold uppercase tracking-widest">Registro de Control</p>
                        <p class="text-[10px] text-gray-500 font-medium">
                          Creado por <span class="font-black text-blue-900">{wa().create_by}</span> el {new Date(wa().created_at).toLocaleDateString("es-VE")}
                        </p>
                      </div>
                    </div>
                  </Show>
                </section>

                {/* ── BOTONES DE ACCIÓN ────────────────────────────────────── */}
                <div class="sticky bottom-10 z-50 flex justify-end gap-4 px-4">
                  <button
                    type="button"
                    onClick={() => navigate(-1)}
                    class="bg-white text-gray-500 border-2 border-gray-100 px-8 py-4 rounded-2xl font-black hover:bg-gray-50 hover:text-gray-700 transition-all text-xs uppercase tracking-widest shadow-xl"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    disabled={saving()}
                    class="bg-blue-900 text-white px-10 py-4 rounded-2xl font-black shadow-2xl hover:bg-blue-800 active:scale-95 transition-all disabled:opacity-50 flex items-center gap-3 border-2 border-white/20 text-xs uppercase tracking-widest"
                  >
                    <Show when={saving()} fallback={<span>💾 Guardar Cambios</span>}>
                       <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                       <span>Guardando...</span>
                    </Show>
                  </button>
                </div>

              </form>
            );
          }}
        </Show>

        <Show when={!workArea.loading && workArea() === null}>
           <div class="text-center py-24 bg-white rounded-[2.5rem] border border-gray-100 shadow-premium">
              <div class="text-6xl mb-6">🔍</div>
              <h2 class="text-xl font-black text-blue-900 uppercase">Área no encontrada</h2>
              <p class="text-gray-500 mt-2 mb-8">El registro solicitado no existe o fue eliminado permanentemente.</p>
              <button
                onClick={() => navigate("/admin/areas_de_ejercicio_profesional")}
                class="bg-blue-50 text-blue-800 px-8 py-3 rounded-2xl font-black text-xs uppercase tracking-widest hover:bg-blue-100 transition-all"
              >
                Volver al Catálogo
              </button>
           </div>
        </Show>
      </Show>

    </main>
  );
}
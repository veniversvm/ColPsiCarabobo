// web/src/routes/admin/proyectos/crear/index.tsx
import { Show, createSignal } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";

const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
const labelClass = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";

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
    <div class="pb-20">
      <button onClick={() => navigate("/admin/proyectos")} class="text-xs font-bold text-blue-500 hover:text-blue-700 mb-4">
        ← Volver a proyectos
      </button>

      <div class="max-w-2xl mx-auto">
        <div class="flex items-center gap-4 mb-8">
          <div class="w-14 h-14 bg-blue-800 rounded-2xl flex items-center justify-center text-2xl text-white shadow-xl">
            📋
          </div>
          <div>
            <h1 class="text-3xl font-black text-colpsi-blue">Nuevo Proyecto</h1>
            <p class="text-sm text-gray-500 mt-1">Un tablero Kanban para organizar el trabajo del colegio.</p>
          </div>
        </div>

        <Show when={error()}>
          <div class="mb-6 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500 shadow-sm">
            {error()}
          </div>
        </Show>

        <div class="bg-white rounded-3xl border border-gray-100 shadow-premium p-6 md:p-8 space-y-6">
          <section>
            <h2 class="text-sm font-black text-colpsi-blue uppercase tracking-widest border-b border-gray-100 pb-3 mb-5">
              Información del proyecto
            </h2>

            <div class="mb-5">
              <label class={labelClass}>Nombre del proyecto *</label>
              <input
                value={name()}
                onInput={(e) => setName(e.currentTarget.value)}
                placeholder="Ej. Organización de la Convención 2026"
                maxLength={120}
                class={IC}
              />
            </div>

            <div class="mb-5">
              <label class={labelClass}>Descripción</label>
              <textarea
                value={description()}
                onInput={(e) => setDescription(e.currentTarget.value)}
                placeholder="¿Qué se quiere lograr con este proyecto?"
                maxLength={500}
                rows={4}
                class={`${IC} resize-none`}
              />
            </div>

            <div class="rounded-2xl bg-blue-50 border border-blue-100 p-4 text-sm text-blue-800">
              <p class="font-black mb-1">Al crear el proyecto se añaden 3 columnas por defecto:</p>
              <p class="text-blue-700">«Por hacer» · «En progreso» · «Hecho»</p>
              <p class="text-xs text-blue-500 mt-2">Podrás invitar a otros administradores como miembros (Espectador o Editor).</p>
            </div>
          </section>

          <div class="flex gap-3 pt-2 border-t border-gray-100">
            <button
              onClick={() => navigate("/admin/proyectos")}
              class="bg-white text-gray-600 border-2 border-gray-200 px-8 py-4 rounded-2xl font-black hover:bg-gray-50"
            >
              Cancelar
            </button>
            <button
              onClick={submit}
              disabled={!canSubmit()}
              class="flex-grow bg-blue-800 hover:bg-blue-900 text-white font-black px-10 py-4 rounded-2xl shadow-xl hover:scale-[1.02] active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
            >
              {saving() ? "Creando…" : "+ Crear Proyecto"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
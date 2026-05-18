// web/src/routes/admin/areas_de_ejercicio_profesional/crear/index.tsx
import { createSignal, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiPost } from "~/lib/api";

const IC = "w-full bg-white border-2 border-gray-100 focus:border-blue-500 rounded-2xl px-5 py-3.5 outline-none transition-all text-gray-800 text-sm shadow-sm";
const labelClass = "block text-[10px] font-black text-gray-400 uppercase tracking-[0.2em] ml-2 mb-2";

export default function AdminCrearAreaEjercicioPage() {
  const navigate = useNavigate();

  const [name, setName] = createSignal("");
  const [description, setDescription] = createSignal("");
  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!name().trim()) { 
      setError("El nombre del área es obligatorio."); 
      return; 
    }

    setSaving(true);
    setError(null);

    try {
      // Mantenemos el endpoint técnico /admin/specialties del backend
      await apiPost("/admin/specialties", {
        name: name().trim(),
        description: description().trim(),
      });
      
      // Redirigir al nuevo catálogo uniforme
      navigate("/admin/areas_de_ejercicio_profesional");
    } catch (err: any) {
      setError(err.message || "Error al registrar el área. Intenta de nuevo.");
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <main class="pb-28 animate-in fade-in duration-500 max-w-3xl mx-auto font-sans">

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
        <div>
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
            Nueva Área de Ejercicio
          </h1>
          <p class="text-gray-400 text-sm mt-1 font-medium">
            Define un nuevo campo de desempeño para los agremiados.
          </p>
        </div>
      </div>

      {/* ── ERROR ALERT ───────────────────────────────────────────────────── */}
      <Show when={error()}>
        <div class="mb-8 p-5 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-2 border-red-100 flex items-center gap-3 animate-in slide-in-from-top-2 duration-300">
          <span class="text-xl">⚠️</span> {error()}
        </div>
      </Show>

      <form onSubmit={handleSubmit} class="space-y-8">

        <section class="bg-white rounded-[2.5rem] p-8 md:p-10 shadow-premium border border-gray-100 space-y-8">
          <div class="border-l-4 border-colpsi-yellow pl-4">
            <h2 class="text-sm font-black text-blue-900 uppercase tracking-widest">
              Información del Catálogo
            </h2>
          </div>

          {/* Nombre del Área */}
          <div>
            <label class={labelClass}>Nombre de la Disciplina / Área</label>
            <input
              type="text"
              required
              maxLength={100}
              placeholder="Ej: Psicología Clínica, Organizacional, etc."
              value={name()}
              onInput={(e) => setName(e.currentTarget.value)}
              class={IC}
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
              placeholder="Describe brevemente el alcance y naturaleza de esta área de desempeño profesional..."
              value={description()}
              onInput={(e) => setDescription(e.currentTarget.value)}
              class={`${IC} resize-none leading-relaxed`}
            />
            <div class="flex justify-end mt-2">
               <p class="text-[10px] text-gray-400 font-bold uppercase tracking-wider">{description().length}/500</p>
            </div>
          </div>

          {/* Nota Informativa */}
          <div class="bg-blue-50/50 rounded-2xl p-5 border border-blue-100 flex items-start gap-4">
            <span class="text-xl">ℹ️</span>
            <p class="text-blue-800 text-xs font-medium leading-relaxed">
              Al crear esta área, se marcará como <span class="font-black uppercase">activa</span> por defecto. 
              Estará disponible inmediatamente para que los psicólogos la seleccionen en su perfil y sea visible en el directorio.
            </p>
          </div>
        </section>

        {/* ── BOTONES DE ACCIÓN ────────────────────────────────────────────── */}
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
            <Show when={saving()} fallback={
              <>
                <span class="text-lg">📂</span>
                <span>Registrar Área</span>
              </>
            }>
               <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
               <span>Procesando...</span>
            </Show>
          </button>
        </div>

      </form>
    </main>
  );
}
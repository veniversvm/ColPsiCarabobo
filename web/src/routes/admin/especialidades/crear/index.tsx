// routes/admin/especialidades/crear/index.tsx
import { createSignal } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiPost } from "~/lib/api";

const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
const labelClass = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";

export default function AdminCrearEspecialidadPage() {
  const navigate = useNavigate();

  const [name, setName] = createSignal("");
  const [description, setDescription] = createSignal("");
  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!name().trim()) { setError("El nombre es obligatorio."); return; }

    setSaving(true);
    setError(null);

    try {
      await apiPost("/admin/specialties", {
        name: name().trim(),
        description: description().trim(),
      });
      navigate("/admin/especialidades");
    } catch (err: any) {
      setError(err.message || "Error al crear. Intenta de nuevo.");
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <main class="pb-28 animate-in fade-in duration-500 max-w-2xl mx-auto">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex items-center gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
        <button
          onClick={() => navigate(-1)}
          class="w-10 h-10 bg-gray-50 hover:bg-gray-100 text-gray-600 rounded-full font-bold flex items-center justify-center transition-colors flex-shrink-0"
        >
          ←
        </button>
        <div>
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
            Nueva Especialidad
          </h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium">
            Agrega una especialidad al catálogo del Colegio
          </p>
        </div>
      </div>

      {/* ── ERROR ─────────────────────────────────────────────────────────── */}
      {error() && (
        <div class="mb-6 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500 shadow-sm">
          {error()}
        </div>
      )}

      <form onSubmit={handleSubmit} class="space-y-6">

        <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100 space-y-5">
          <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3">
            Información de la Especialidad
          </h2>

          {/* Nombre */}
          <div>
            <label class={labelClass}>Nombre <span class="text-red-400">*</span></label>
            <input
              type="text"
              required
              maxLength={100}
              placeholder="Ej. Psicología Clínica"
              value={name()}
              onInput={(e) => setName(e.currentTarget.value)}
              class={IC}
            />
            <p class="text-[10px] text-gray-400 mt-1 text-right">{name().length}/100</p>
          </div>

          {/* Descripción */}
          <div>
            <label class={labelClass}>Descripción</label>
            <textarea
              rows={4}
              maxLength={500}
              placeholder="Describe brevemente el área de especialización..."
              value={description()}
              onInput={(e) => setDescription(e.currentTarget.value)}
              class={`${IC} resize-none`}
            />
            <p class="text-[10px] text-gray-400 mt-1 text-right">{description().length}/500</p>
          </div>

          {/* Info */}
          <div class="bg-blue-50 rounded-2xl p-4 border border-blue-100">
            <p class="text-blue-700 text-xs font-bold">
              ℹ️ La especialidad se creará como <span class="font-black">activa</span> y estará disponible inmediatamente en el directorio de psicólogos.
            </p>
          </div>
        </section>

        {/* ── BOTONES ───────────────────────────────────────────────────── */}
        <div class="sticky bottom-6 z-50 flex justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate(-1)}
            class="bg-white text-gray-600 border-2 border-gray-200 px-6 py-4 rounded-2xl font-black hover:bg-gray-50 transition-all text-sm"
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={saving()}
            class="bg-blue-800 text-white px-10 py-4 rounded-2xl font-black shadow-2xl hover:scale-105 active:scale-95 transition-all disabled:opacity-70 flex items-center gap-3 border-2 border-white text-sm"
          >
            {saving() ? "CREANDO..." : "🏷️ CREAR ESPECIALIDAD"}
          </button>
        </div>

      </form>
    </main>
  );
}
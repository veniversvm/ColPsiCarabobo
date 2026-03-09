// routes/admin/especialidades/[id].tsx
import { createResource, createSignal, Show } from "solid-js";
import { useNavigate, useParams } from "@solidjs/router";
import { apiGet, apiPatch } from "~/lib/api";

const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
const labelClass = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";

interface Specialty {
  id: number;
  name: string;
  description: string;
  active: boolean;
  created_at: string;
  create_by: string;
}

export default function AdminEditarEspecialidadPage() {
  const params = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [specialty] = createResource(
    () => params.id,
    async (id) => {
      try {
        return await apiGet<Specialty>(`/specialties/${id}`);
      } catch (err: any) {
        console.error("[edit specialty] error:", err?.status, err?.message);
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

  const initForm = (s: Specialty) => {
    if (initialized()) return;
    setName(s.name ?? "");
    setDescription(s.description ?? "");
    setActive(s.active ?? true);
    setInitialized(true);
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!name().trim()) { setError("El nombre es obligatorio."); return; }

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
      setTimeout(() => navigate("/admin/especialidades"), 1200);
    } catch (err: any) {
      setError(err.message || "Error al guardar. Intenta de nuevo.");
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
        <div class="flex-1 min-w-0">
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
            Editar Especialidad
          </h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium truncate">
            {specialty.loading ? "Cargando..." : specialty()?.name ?? ""}
          </p>
        </div>
        <Show when={specialty()}>
          {(s) => (
            <span class={`text-[10px] font-black px-3 py-1.5 rounded-full uppercase tracking-widest flex-shrink-0 ${
              s().active ? "bg-emerald-100 text-emerald-700" : "bg-gray-100 text-gray-500"
            }`}>
              {s().active ? "Activa" : "Inactiva"}
            </span>
          )}
        </Show>
      </div>

      {/* ── FEEDBACK ──────────────────────────────────────────────────────── */}
      <Show when={error()}>
        <div class="mb-6 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500 shadow-sm">
          {error()}
        </div>
      </Show>
      <Show when={success()}>
        <div class="mb-6 p-4 rounded-2xl bg-emerald-50 text-emerald-800 font-bold text-sm border-l-4 border-emerald-500 shadow-sm">
          ✓ Especialidad actualizada correctamente. Redirigiendo...
        </div>
      </Show>

      {/* ── SKELETON ──────────────────────────────────────────────────────── */}
      <Show when={specialty.loading}>
        <div class="space-y-6 animate-pulse">
          <div class="bg-white rounded-3xl h-64 border border-gray-100" />
        </div>
      </Show>

      {/* ── NO ENCONTRADO ─────────────────────────────────────────────────── */}
      <Show when={!specialty.loading && specialty() === null}>
        <div class="text-center py-24 bg-white rounded-3xl border border-gray-100">
          <p class="text-5xl mb-4">😕</p>
          <h2 class="text-lg font-black text-gray-700 mb-2">Especialidad no encontrada</h2>
          <button
            onClick={() => navigate("/admin/especialidades")}
            class="inline-flex items-center gap-2 text-blue-700 font-black text-sm hover:underline mt-4"
          >
            ← Volver al listado
          </button>
        </div>
      </Show>

      {/* ── FORMULARIO ────────────────────────────────────────────────────── */}
      <Show when={specialty()}>
        {(s) => {
          initForm(s());
          return (
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
                    value={description()}
                    onInput={(e) => setDescription(e.currentTarget.value)}
                    class={`${IC} resize-none`}
                  />
                  <p class="text-[10px] text-gray-400 mt-1 text-right">{description().length}/500</p>
                </div>

                {/* Estado */}
                <div>
                  <label class={labelClass}>Estado</label>
                  <div class="flex gap-3 mt-1">
                    {([true, false] as const).map((val) => (
                      <button
                        type="button"
                        onClick={() => setActive(val)}
                        class={`flex-1 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide border-2 transition-all ${
                          active() === val
                            ? val
                              ? "bg-emerald-600 text-white border-emerald-600"
                              : "bg-gray-500 text-white border-gray-500"
                            : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
                        }`}
                      >
                        {val ? "✓ Activa" : "○ Inactiva"}
                      </button>
                    ))}
                  </div>
                  <p class="text-[10px] text-gray-400 mt-1 ml-1">
                    {active()
                      ? "Visible en el directorio y formularios de psicólogos."
                      : "Oculta en el directorio. Los psicólogos que ya la tenían asignada la conservan."}
                  </p>
                </div>

                {/* Metadatos */}
                <Show when={s().create_by}>
                  <div class="bg-gray-50 rounded-2xl p-4 border border-gray-100">
                    <p class="text-[11px] text-gray-400 font-medium">
                      Creada por <span class="font-black text-gray-600">{s().create_by}</span>
                      <Show when={s().created_at}>
                        {" · "}{new Date(s().created_at).toLocaleDateString("es-VE", { day: "2-digit", month: "short", year: "numeric" })}
                      </Show>
                    </p>
                  </div>
                </Show>
              </section>

              {/* ── BOTONES ─────────────────────────────────────────────── */}
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
                  {saving() ? "GUARDANDO..." : "💾 GUARDAR CAMBIOS"}
                </button>
              </div>

            </form>
          );
        }}
      </Show>

    </main>
  );
}
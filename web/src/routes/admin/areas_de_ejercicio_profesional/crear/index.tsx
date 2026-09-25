// web/src/routes/admin/areas_de_ejercicio_profesional/crear/index.tsx
import { createSignal, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Icon } from "~/components/admin/ui/icons";

const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

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
      const idempotencyKey = crypto.randomUUID();
      // Mantenemos el endpoint técnico /admin/specialties del backend
      await apiPost("/admin/specialties", {
        name: name().trim(),
        description: description().trim(),
      }, {
        headers: { "X-Idempotency-Key": idempotencyKey },
      });
      
      // Redirigir al nuevo catálogo uniforme
      navigate("/admin/areas_de_ejercicio_profesional");
    } catch (err: any) {
      setError(getUserFacingError(err));
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  // Función para filtrar caracteres especiales en tiempo real
  const handleNameInput = (e: Event) => {
    const input = e.currentTarget as HTMLInputElement;
    // Permite: letras, números, espacios, acentos, ñ, guiones (-) y barras (/)
    const sanitizedValue = input.value.replace(/[^a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s\/\-]/g, "");

    setName(sanitizedValue);
    // Forzamos el input visual para que borre el caracter inválido instantáneamente
    input.value = sanitizedValue;
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
        <div>
          <h1 class="text-lg font-semibold text-colpsi-text">Nueva Área de Ejercicio</h1>
          <p class="text-sm text-colpsi-muted mt-0.5">Define un nuevo campo de desempeño para los agremiados.</p>
        </div>
      </div>

      {/* ── ERROR ALERT ───────────────────────────────────────────────────── */}
      <Show when={error()}>
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">
          {error()}
        </div>
      </Show>

      <form onSubmit={handleSubmit} class="space-y-4">

        <section class="bg-white rounded-lg p-5 border border-colpsi-border space-y-4">
          <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3">
            Información del Catálogo
          </h2>

          {/* Nombre del Área */}
          <div>
            <label class={labelClass}>Nombre de la Disciplina / Área</label>
            <input
              type="text"
              required
              maxLength={100}
              placeholder="Ej: Psicología Clínica, Organizacional, etc."
              value={name()}
              onInput={handleNameInput}
              class={IC}
            />
            <div class="flex justify-between mt-1.5 px-1">
               <p class="text-[11px] text-colpsi-muted italic">No se admiten símbolos especiales.</p>
               <p class="text-xs text-colpsi-muted font-medium">{name().length}/100</p>
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
              class={`${IC} resize-none min-h-32 leading-relaxed`}
            />
            <div class="flex justify-end mt-1.5">
               <p class="text-xs text-colpsi-muted font-medium">{description().length}/500</p>
            </div>
          </div>

          {/* Nota Informativa */}
          <div class="bg-sky-50/50 rounded-md p-3 border border-sky-100 flex items-start gap-2.5">
            <Icon name="info" class="w-4 h-4 text-sky-600 shrink-0 mt-0.5" />
            <p class="text-sky-900 text-xs font-medium leading-relaxed">
              Al crear esta área, se marcará como <span class="font-semibold uppercase">activa</span> por defecto.
              Estará disponible inmediatamente para que los psicólogos la seleccionen en su perfil y sea visible en el directorio.
            </p>
          </div>
        </section>

        {/* ── BOTONES DE ACCIÓN ────────────────────────────────────────────── */}
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
            <Show when={saving()} fallback={
              <>
                <Icon name="check" />
                <span>Registrar Área</span>
              </>
            }>
               <span class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
               <span>Procesando...</span>
            </Show>
          </button>
        </div>

      </form>
    </main>
  );
}
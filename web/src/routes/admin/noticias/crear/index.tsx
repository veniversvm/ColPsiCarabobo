// web/src/routes/admin/noticias/crear/index.tsx
import { createSignal } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { RichTextEditor } from "~/components/ui/RichTextEditor";
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { Icon } from "~/components/admin/ui/icons";

// ── Acción multipart ──────────────────────────────────────────────────────────
// No usamos server action aquí porque el endpoint acepta multipart/form-data
// con un archivo adjunto. Lo enviamos directamente desde el cliente con fetch.

const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

// ─────────────────────────────────────────────────────────────────────────────
export default function AdminCrearNoticiaPage() {
  const navigate = useNavigate();

  // Una key por montaje — se regenera si el admin navega fuera y vuelve
  const idempotencyKey = crypto.randomUUID();

  // ── Estado del formulario ─────────────────────────────────────────────────
  const [title, setTitle] = createSignal("");
  const [shortDescription, setShortDescription] = createSignal("");
  const [content, setContent] = createSignal("");
  const [type, setType] = createSignal<"public" | "psi">("public");
  const [isActive, setIsActive] = createSignal(true);
  const [imageFile, setImageFile] = createSignal<File | null>(null);
  const [imagePreview, setImagePreview] = createSignal<string | null>(null);

  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  // ── Manejo de imagen ──────────────────────────────────────────────────────
  const handleImageChange = (e: Event) => {
    const file = (e.currentTarget as HTMLInputElement).files?.[0] ?? null;
    setImageFile(file);
    if (file) {
      const reader = new FileReader();
      reader.onload = (ev) => setImagePreview(ev.target?.result as string);
      reader.readAsDataURL(file);
    } else {
      setImagePreview(null);
    }
  };

  const clearImage = () => {
    setImageFile(null);
    setImagePreview(null);
  };

  // ── Submit ────────────────────────────────────────────────────────────────
  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!title().trim()) { setError("El título es obligatorio."); return; }
    if (!content().trim() || content() === "<p></p>") { setError("El contenido no puede estar vacío."); return; }

    setSaving(true);
    setError(null);

    try {
      const fd = new FormData();
      fd.append("title", title().trim());
      fd.append("short_description", shortDescription().trim());
      fd.append("content", content());
      fd.append("type", type());
      fd.append("is_active", String(isActive()));
      if (imageFile()) fd.append("image", imageFile()!);

      const { apiPost } = await import("~/lib/api");
      const { getUserFacingError } = await import("~/lib/errors");

      // La key va en el header — el middleware de Go la valida contra el user ID
      await apiPost("/admin/posts", fd, {
        headers: { "X-Idempotency-Key": idempotencyKey },
      });

      navigate("/admin/noticias");
    } catch (err: any) {
        setError(getUserFacingError(err));
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  // ─── Render ───────────────────────────────────────────────────────────────
  return (
    <main class="space-y-4 pb-12 max-w-4xl mx-auto">

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
          <h1 class="text-lg font-semibold text-colpsi-text">Nueva Publicación</h1>
          <p class="text-sm text-colpsi-muted mt-0.5">Crea una noticia o comunicado para la plataforma</p>
        </div>
      </div>

      {/* ── ERROR ─────────────────────────────────────────────────────────── */}
      {error() && (
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">
          {error()}
        </div>
      )}

      <form onSubmit={handleSubmit} class="space-y-4">

        {/* ══ BLOQUE 1: METADATOS ══════════════════════════════════════════ */}
        <section class="bg-white rounded-lg p-5 border border-colpsi-border space-y-4">
          <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3">
            Información General
          </h2>

          {/* Título */}
          <div>
            <label class={labelClass}>Título <span class="text-red-400">*</span></label>
            <input
              type="text"
              required
              maxLength={100}
              placeholder="Ej. Convocatoria ordinaria 2026"
              value={title()}
              onInput={(e) => setTitle(e.currentTarget.value)}
              class={IC}
            />
            <p class="text-xs text-colpsi-muted mt-1 text-right">{title().length}/100</p>
          </div>

          {/* Resumen */}
          <div>
            <label class={labelClass}>Resumen (snippet del feed)</label>
            <textarea
              rows={2}
              maxLength={250}
              placeholder="Breve descripción que aparece en la lista de noticias..."
              value={shortDescription()}
              onInput={(e) => setShortDescription(e.currentTarget.value)}
              class={`${IC} resize-none min-h-20`}
            />
            <p class="text-xs text-colpsi-muted mt-1 text-right">{shortDescription().length}/250</p>
          </div>

          {/* Tipo y estado en fila */}
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class={labelClass}>Audiencia</label>
              <div class="flex gap-2 mt-1">
                {(["public", "psi"] as const).map((t) => (
                  <button
                    type="button"
                    onClick={() => setType(t)}
                    class={`flex-1 h-9 rounded-md border text-xs font-semibold transition-colors ${
                      type() === t
                        ? t === "public"
                          ? "bg-emerald-600 text-white border-emerald-600"
                          : "bg-colpsi-blue text-white border-colpsi-blue"
                        : "bg-white text-colpsi-muted border-colpsi-border hover:border-colpsi-blue/40"
                    }`}
                  >
                    {t === "public" ? "Público" : "Colegiados"}
                  </button>
                ))}
              </div>
              <p class="text-[11px] text-colpsi-muted mt-1 ml-1">
                {type() === "public"
                  ? "Visible para cualquier visitante del sitio."
                  : "Solo psicólogos con sesión iniciada."}
              </p>
            </div>

            <div class="flex flex-col justify-center bg-colpsi-bg rounded-md px-4 py-3 border border-colpsi-border">
              <ToggleSwitch
                label="Publicar inmediatamente"
                checked={isActive()}
                onChange={(v) => setIsActive(v)}
              />
              <p class="text-[11px] text-colpsi-muted mt-2 ml-1">
                {isActive()
                  ? "Visible en la plataforma al guardar."
                  : "Se guardará como borrador (no visible)."}
              </p>
            </div>
          </div>
        </section>

        {/* ══ BLOQUE 2: IMAGEN DE PORTADA ══════════════════════════════════ */}
        <section class="bg-white rounded-lg p-5 border border-colpsi-border">
          <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3 mb-4">
            Imagen de Portada
          </h2>

          {imagePreview() ? (
            <div class="relative group rounded-lg overflow-hidden border border-colpsi-border">
              <img src={imagePreview()!} alt="Vista previa" class="w-full max-h-64 object-cover" />
              <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                <button
                  type="button"
                  onClick={clearImage}
                  class="inline-flex items-center gap-1.5 h-9 px-3 bg-white text-colpsi-red font-semibold rounded-md text-xs border border-colpsi-border shadow-sm hover:bg-red-50 transition-colors"
                >
                  <Icon name="trash" class="w-3.5 h-3.5" />
                  Quitar imagen
                </button>
              </div>
              <div class="absolute bottom-2 left-2 bg-black/60 text-white text-[10px] font-semibold px-2 py-1 rounded-md">
                {imageFile()?.name}
              </div>
            </div>
          ) : (
            <label class="flex flex-col items-center justify-center w-full h-44 border-2 border-dashed border-slate-300 rounded-lg bg-colpsi-bg hover:bg-white hover:border-colpsi-blue/40 transition-colors cursor-pointer group">
              <div class="flex flex-col items-center gap-2 text-colpsi-muted group-hover:text-colpsi-blue transition-colors">
                <svg class="w-10 h-10" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3 9.75h.008M3.375 3h17.25A.375.375 0 0121 3.375v17.25A.375.375 0 0120.625 21H3.375A.375.375 0 013 20.625V3.375A.375.375 0 013.375 3z" />
                </svg>
                <span class="font-semibold text-sm">Haz clic para subir imagen</span>
                <span class="text-[11px]">JPG, PNG, WebP · Máx. 5MB</span>
              </div>
              <input type="file" accept="image/*" class="hidden" onChange={handleImageChange} />
            </label>
          )}
        </section>

        {/* ══ BLOQUE 3: CONTENIDO ENRIQUECIDO ══════════════════════════════ */}
        <section class="bg-white rounded-lg p-5 border border-colpsi-border">
          <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3 mb-4">
            Contenido <span class="text-red-400">*</span>
          </h2>
          <RichTextEditor
            content={content()}
            onUpdate={(html) => setContent(html)}
          />
        </section>

        {/* ── BOTÓN FLOTANTE ─────────────────────────────────────────────── */}
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
            {saving() ? "Procesando..." : isActive() ? "Publicar ahora" : "Guardar borrador"}
          </button>
        </div>

      </form>
    </main>
  );
}
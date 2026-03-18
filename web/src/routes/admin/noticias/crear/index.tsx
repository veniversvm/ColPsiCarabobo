// web/src/routes/admin/noticias/crear/index.tsx
import { createSignal } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { RichTextEditor } from "~/components/ui/RichTextEditor";
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";

// ── Acción multipart ──────────────────────────────────────────────────────────
// No usamos server action aquí porque el endpoint acepta multipart/form-data
// con un archivo adjunto. Lo enviamos directamente desde el cliente con fetch.

const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
const labelClass = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";

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

      // La key va en el header — el middleware de Go la valida contra el user ID
      await apiPost("/admin/posts", fd, {
        headers: { "X-Idempotency-Key": idempotencyKey },
      });

      navigate("/admin/noticias");
    } catch (err: any) {
      setError(err.message || "Error al publicar. Intenta de nuevo.");
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  // ─── Render ───────────────────────────────────────────────────────────────
  return (
    <main class="pb-28 animate-in fade-in duration-500 max-w-4xl mx-auto">

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
            Nueva Publicación
          </h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium">
            Crea una noticia o comunicado para la plataforma
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

        {/* ══ BLOQUE 1: METADATOS ══════════════════════════════════════════ */}
        <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100 space-y-5">
          <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3">
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
            <p class="text-[10px] text-gray-400 mt-1 text-right">{title().length}/100</p>
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
              class={`${IC} resize-none`}
            />
            <p class="text-[10px] text-gray-400 mt-1 text-right">{shortDescription().length}/250</p>
          </div>

          {/* Tipo y estado en fila */}
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
            <div>
              <label class={labelClass}>Audiencia</label>
              <div class="flex gap-3 mt-1">
                {(["public", "psi"] as const).map((t) => (
                  <button
                    type="button"
                    onClick={() => setType(t)}
                    class={`flex-1 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide border-2 transition-all ${
                      type() === t
                        ? t === "public"
                          ? "bg-emerald-600 text-white border-emerald-600"
                          : "bg-blue-700 text-white border-blue-700"
                        : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
                    }`}
                  >
                    {t === "public" ? "🌐 Público" : "🔒 Colegiados"}
                  </button>
                ))}
              </div>
              <p class="text-[10px] text-gray-400 mt-1 ml-1">
                {type() === "public"
                  ? "Visible para cualquier visitante del sitio."
                  : "Solo psicólogos con sesión iniciada."}
              </p>
            </div>

            <div class="flex flex-col justify-center bg-gray-50 rounded-2xl px-5 py-4 border border-gray-100">
              <ToggleSwitch
                label="Publicar inmediatamente"
                checked={isActive()}
                onChange={(v) => setIsActive(v)}
              />
              <p class="text-[10px] text-gray-400 mt-2 ml-1">
                {isActive()
                  ? "Visible en la plataforma al guardar."
                  : "Se guardará como borrador (no visible)."}
              </p>
            </div>
          </div>
        </section>

        {/* ══ BLOQUE 2: IMAGEN DE PORTADA ══════════════════════════════════ */}
        <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
          <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3 mb-5">
            Imagen de Portada
          </h2>

          {imagePreview() ? (
            <div class="relative group rounded-2xl overflow-hidden border-2 border-gray-200">
              <img src={imagePreview()!} alt="Vista previa" class="w-full max-h-64 object-cover" />
              <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                <button
                  type="button"
                  onClick={clearImage}
                  class="bg-white text-red-600 font-black px-4 py-2 rounded-xl text-sm hover:bg-red-50 transition-all shadow"
                >
                  🗑 Quitar imagen
                </button>
              </div>
              <div class="absolute bottom-2 left-2 bg-black/60 text-white text-[10px] font-bold px-2 py-1 rounded-lg">
                {imageFile()?.name}
              </div>
            </div>
          ) : (
            <label class="flex flex-col items-center justify-center w-full h-44 border-2 border-dashed border-gray-300 rounded-2xl bg-gray-50 hover:bg-blue-50 hover:border-blue-300 transition-all cursor-pointer group">
              <div class="flex flex-col items-center gap-2 text-gray-400 group-hover:text-blue-500 transition-colors">
                <svg class="w-10 h-10" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3 9.75h.008M3.375 3h17.25A.375.375 0 0121 3.375v17.25A.375.375 0 0120.625 21H3.375A.375.375 0 013 20.625V3.375A.375.375 0 013.375 3z" />
                </svg>
                <span class="font-bold text-sm">Haz clic para subir imagen</span>
                <span class="text-[11px]">JPG, PNG, WebP · Máx. 5MB</span>
              </div>
              <input type="file" accept="image/*" class="hidden" onChange={handleImageChange} />
            </label>
          )}
        </section>

        {/* ══ BLOQUE 3: CONTENIDO ENRIQUECIDO ══════════════════════════════ */}
        <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
          <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3 mb-5">
            Contenido <span class="text-red-400">*</span>
          </h2>
          <RichTextEditor
            content={content()}
            onUpdate={(html) => setContent(html)}
          />
        </section>

        {/* ── BOTÓN FLOTANTE ─────────────────────────────────────────────── */}
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
            {saving() ? "PUBLICANDO..." : isActive() ? "🚀 PUBLICAR AHORA" : "💾 GUARDAR BORRADOR"}
          </button>
        </div>

      </form>
    </main>
  );
}
// routes/admin/noticias/[id].tsx
import { createResource, createSignal, Show } from "solid-js";
import { useNavigate, useParams } from "@solidjs/router";
import { RichTextEditor } from "~/components/ui/RichTextEditor";
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { apiGet, apiPatch } from "~/lib/api";

const BUCKET_URL = import.meta.env.VITE_BUCKET_URL || "http://localhost:9000/colpsi-bucket";
const imgUrl = (key: string) => (key ? `${BUCKET_URL}/${key}` : "");

const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
const labelClass = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";

interface PostDetail {
  id: string;
  title: string;
  short_description: string;
  type: "public" | "psi";
  is_active: boolean;
  image_url: string;
  text: { id: string; content: string };
}

// ─────────────────────────────────────────────────────────────────────────────
export default function AdminEditarNoticiaPage() {
  const params = useParams<{ id: string }>();
  const navigate = useNavigate();

  // Fetch directo desde cliente — sin server function
  const [post] = createResource(
    () => params.id,
    async (id) => {
      try {
        return await apiGet<PostDetail>(`/posts/${id}`);
      } catch (err: any) {
        console.error("[edit] error cargando post:", err?.status, err?.message);
        return null;
      }
    }
  );

  // Estado del formulario
  const [title, setTitle] = createSignal("");
  const [shortDescription, setShortDescription] = createSignal("");
  const [content, setContent] = createSignal("");
  const [type, setType] = createSignal<"public" | "psi">("public");
  const [isActive, setIsActive] = createSignal(true);
  const [imageFile, setImageFile] = createSignal<File | null>(null);
  const [imagePreview, setImagePreview] = createSignal<string | null>(null);
  const [initialized, setInitialized] = createSignal(false);

  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [success, setSuccess] = createSignal(false);

  const initForm = (p: PostDetail) => {
    if (initialized()) return;
    setTitle(p.title ?? "");
    setShortDescription(p.short_description ?? "");
    setContent(p.text?.content ?? "");
    setType(p.type ?? "public");
    setIsActive(p.is_active ?? true);
    setInitialized(true);
  };

  const currentImage = () =>
    imagePreview() ?? (post()?.image_url ? imgUrl(post()!.image_url) : null);

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

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!title().trim()) { setError("El título es obligatorio."); return; }
    if (!content().trim() || content() === "<p></p>") { setError("El contenido no puede estar vacío."); return; }

    setSaving(true);
    setError(null);
    setSuccess(false);

    try {
      const fd = new FormData();
      fd.append("title", title().trim());
      fd.append("short_description", shortDescription().trim());
      fd.append("content", content());
      fd.append("type", type());
      fd.append("is_active", String(isActive()));
      if (imageFile()) fd.append("image", imageFile()!);

      await apiPatch(`/admin/posts/${params.id}`, fd);

      setSuccess(true);
      window.scrollTo({ top: 0, behavior: "smooth" });
      setTimeout(() => navigate("/admin/noticias"), 1200);
    } catch (err: any) {
      setError(err.message || "Error al guardar. Intenta de nuevo.");
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

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
        <div class="flex-1 min-w-0">
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
            Editar Publicación
          </h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium truncate">
            {post.loading ? "Cargando..." : post()?.title ?? ""}
          </p>
        </div>
        <Show when={post()}>
          {(p) => (
            <span class={`text-[10px] font-black px-3 py-1.5 rounded-full uppercase tracking-widest flex-shrink-0 ${
              p().is_active ? "bg-emerald-100 text-emerald-700" : "bg-amber-100 text-amber-700"
            }`}>
              {p().is_active ? "Publicado" : "Borrador"}
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
          ✓ Publicación actualizada correctamente. Redirigiendo...
        </div>
      </Show>

      {/* ── SKELETON ──────────────────────────────────────────────────────── */}
      <Show when={post.loading}>
        <div class="space-y-6 animate-pulse">
          <div class="bg-white rounded-3xl h-64 border border-gray-100" />
          <div class="bg-white rounded-3xl h-44 border border-gray-100" />
          <div class="bg-white rounded-3xl h-96 border border-gray-100" />
        </div>
      </Show>

      {/* ── NO ENCONTRADO ─────────────────────────────────────────────────── */}
      <Show when={!post.loading && post() === null}>
        <div class="text-center py-24 bg-white rounded-3xl border border-gray-100">
          <p class="text-5xl mb-4">😕</p>
          <h2 class="text-lg font-black text-gray-700 mb-2">Publicación no encontrada</h2>
          <p class="text-gray-400 text-sm mb-6">Es posible que haya sido eliminada.</p>
          <button
            onClick={() => navigate("/admin/noticias")}
            class="inline-flex items-center gap-2 text-blue-700 font-black text-sm hover:underline"
          >
            ← Volver al listado
          </button>
        </div>
      </Show>

      {/* ── FORMULARIO ────────────────────────────────────────────────────── */}
      <Show when={post()}>
        {(p) => {
          initForm(p());
          return (
            <form onSubmit={handleSubmit} class="space-y-6">

              {/* ══ BLOQUE 1: METADATOS ════════════════════════════════════ */}
              <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100 space-y-5">
                <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3">
                  Información General
                </h2>

                <div>
                  <label class={labelClass}>Título <span class="text-red-400">*</span></label>
                  <input
                    type="text"
                    required
                    maxLength={100}
                    value={title()}
                    onInput={(e) => setTitle(e.currentTarget.value)}
                    class={IC}
                  />
                  <p class="text-[10px] text-gray-400 mt-1 text-right">{title().length}/100</p>
                </div>

                <div>
                  <label class={labelClass}>Resumen (snippet del feed)</label>
                  <textarea
                    rows={2}
                    maxLength={250}
                    value={shortDescription()}
                    onInput={(e) => setShortDescription(e.currentTarget.value)}
                    class={`${IC} resize-none`}
                  />
                  <p class="text-[10px] text-gray-400 mt-1 text-right">{shortDescription().length}/250</p>
                </div>

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
                      {type() === "public" ? "Visible para cualquier visitante." : "Solo psicólogos con sesión iniciada."}
                    </p>
                  </div>

                  <div class="flex flex-col justify-center bg-gray-50 rounded-2xl px-5 py-4 border border-gray-100">
                    <ToggleSwitch
                      label="Publicar"
                      checked={isActive()}
                      onChange={(v) => setIsActive(v)}
                    />
                    <p class="text-[10px] text-gray-400 mt-2 ml-1">
                      {isActive() ? "Visible en la plataforma." : "Guardado como borrador (no visible)."}
                    </p>
                  </div>
                </div>
              </section>

              {/* ══ BLOQUE 2: IMAGEN ══════════════════════════════════════ */}
              <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
                <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3 mb-5">
                  Imagen de Portada
                </h2>

                <Show
                  when={currentImage()}
                  fallback={
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
                  }
                >
                  <div class="relative group rounded-2xl overflow-hidden border-2 border-gray-200">
                    <img
                      src={currentImage()!}
                      alt="Vista previa"
                      class="w-full max-h-64 object-contain bg-gray-50"
                    />
                    <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-3">
                      <label class="bg-white text-blue-700 font-black px-4 py-2 rounded-xl text-sm hover:bg-blue-50 transition-all shadow cursor-pointer">
                        🖼 Cambiar imagen
                        <input type="file" accept="image/*" class="hidden" onChange={handleImageChange} />
                      </label>
                      <Show when={imagePreview()}>
                        <button
                          type="button"
                          onClick={clearImage}
                          class="bg-white text-red-600 font-black px-4 py-2 rounded-xl text-sm hover:bg-red-50 transition-all shadow"
                        >
                          ↩ Restaurar original
                        </button>
                      </Show>
                    </div>
                    <div class="absolute bottom-2 left-2 bg-black/60 text-white text-[10px] font-bold px-2 py-1 rounded-lg">
                      {imageFile()?.name ?? "Imagen actual"}
                    </div>
                  </div>
                </Show>
              </section>

              {/* ══ BLOQUE 3: CONTENIDO ════════════════════════════════════ */}
              <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
                <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3 mb-5">
                  Contenido <span class="text-red-400">*</span>
                </h2>
                <RichTextEditor
                  content={content()}
                  onUpdate={(html) => setContent(html)}
                />
              </section>

              {/* ── BOTONES FLOTANTES ──────────────────────────────────── */}
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
                  {saving() ? "GUARDANDO..." : isActive() ? "💾 GUARDAR CAMBIOS" : "💾 GUARDAR BORRADOR"}
                </button>
              </div>

            </form>
          );
        }}
      </Show>

    </main>
  );
}
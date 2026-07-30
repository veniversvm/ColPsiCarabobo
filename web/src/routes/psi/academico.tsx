// web/src/routes/psi/academico.tsx
import { createResource, For, Show, Suspense, createSignal } from "solid-js";
import { createStore } from "solid-js/store";
import { A } from "@solidjs/router";
import { apiGet, apiPost, apiPatch, apiDelete } from "~/lib/api";
import { bucketUrl } from "~/lib/bucket";
import { getUserFacingError } from "~/lib/errors";
import { FileUploader } from "~/components/ui/fileUploader";

export default function AcademicoPage() {
  // ── Datos ──────────────────────────────────────────────────────────────────
  const [profile, { refetch }] = createResource(() => apiGet<any>("/psi/me"));

  // ── Control de UI ──────────────────────────────────────────────────────────
  const [showForm,   setShowForm]   = createSignal(false);
  const [isEditing,  setIsEditing]  = createSignal<string | null>(null);
  const [loading,    setLoading]    = createSignal(false);
  const [formError,  setFormError]  = createSignal<string | null>(null);
  const [deleteId,   setDeleteId]   = createSignal<string | null>(null); // ID pendiente de confirmar
  const [deleteError, setDeleteError] = createSignal<string | null>(null);

  // ── Store del formulario ───────────────────────────────────────────────────
  const [form, setForm] = createStore({
    title:           "",
    university:      "",
    graduation_year: "",
    description:     "",
  });

  const [files, setFiles] = createSignal<{ [key: string]: File }>({});

  const handleFileChange = (e: Event, key: string) => {
    const target = e.target as HTMLInputElement;
    if (target.files?.[0]) {
      setFiles({ ...files(), [key]: target.files[0] });
    }
  };

  // ── Abrir formulario ───────────────────────────────────────────────────────
  const openEdit = (pg: any) => {
    setIsEditing(pg.id);
    setForm({
      title:           pg.post_grade_title,
      university:      pg.post_grade_university,
      graduation_year: pg.post_grade_graduation_year,
      description:     pg.post_grade_description || "",
    });
    setFiles({});
    setFormError(null);
    setShowForm(true);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const openCreate = () => {
    setIsEditing(null);
    setForm({ title: "", university: "", graduation_year: "", description: "" });
    setFiles({});
    setFormError(null);
    setShowForm(true);
  };

  // ── Submit ─────────────────────────────────────────────────────────────────
  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setLoading(true);
    setFormError(null);

    const formData = new FormData();
    formData.append("title",           form.title);
    formData.append("university",       form.university);
    formData.append("graduation_year",  form.graduation_year);
    formData.append("description",      form.description);
    if (files().pic_one)   formData.append("pic_one",   files().pic_one);
    if (files().pic_two)   formData.append("pic_two",   files().pic_two);
    if (files().pic_three) formData.append("pic_three", files().pic_three);

    try {
      if (isEditing()) {
        await apiPatch(`/psi/me/postgrades/${isEditing()}`, formData);
      } else {
        await apiPost("/psi/me/postgrades", formData);
      }
      setShowForm(false);
      refetch();
    } catch (err: any) {
        setFormError(getUserFacingError(err));
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setLoading(false);
    }
  };

  // ── Eliminar ───────────────────────────────────────────────────────────────
  const confirmDelete = async () => {
    const id = deleteId();
    if (!id) return;
    setDeleteError(null);
    try {
      await apiDelete(`/psi/me/postgrades/${id}`);
      setDeleteId(null);
      refetch();
    } catch (err: any) {
        setDeleteError(getUserFacingError(err));
    }
  };

  // ── Render ─────────────────────────────────────────────────────────────────
  return (
    <main class="bg-[#f8fafc] min-h-screen pb-24 font-sans">

      {/* ── HEADER ──────────────────────────────────────────────────────────── */}
      <div class="bg-colpsi-blue pt-10 pb-20 px-4 md:px-8 shadow-inner">
        <div class="max-w-4xl mx-auto flex items-center justify-between">
          <A href="/psi" class="text-white font-bold flex items-center gap-2 hover:text-colpsi-yellow transition-colors">
            <span>←</span> Panel
          </A>
          <button
            onClick={() => showForm() ? setShowForm(false) : openCreate()}
            class="bg-colpsi-yellow text-colpsi-blue px-6 py-2.5 rounded-full font-black text-sm shadow-lg active:scale-95 transition-all"
          >
            {showForm() ? "CERRAR" : "+ NUEVO TÍTULO"}
          </button>
        </div>
        <div class="max-w-4xl mx-auto mt-8">
          <h1 class="text-white text-3xl font-black">Formación Académica</h1>
          <p class="text-blue-200 mt-1 italic uppercase text-xs tracking-widest font-bold">Respaldo Profesional</p>
        </div>
      </div>

      <div class="max-w-4xl mx-auto px-4 md:px-8 -mt-10 space-y-8">

        {/* ── MODAL DE CONFIRMACIÓN DE ELIMINACIÓN ──────────────────────────── */}
        <Show when={deleteId()}>
          <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm animate-in fade-in duration-200">
            <div class="bg-white rounded-3xl p-8 shadow-2xl max-w-sm w-full mx-4 border border-gray-100">
              <div class="text-center mb-6">
                <span class="text-4xl">🗑️</span>
                <h3 class="text-lg font-black text-gray-800 mt-3">¿Eliminar este título?</h3>
                <p class="text-sm text-gray-500 mt-1">Esta acción es permanente y no se puede deshacer.</p>
              </div>

              <Show when={deleteError()}>
                <div class="mb-4 p-3 rounded-xl bg-red-50 text-red-700 text-sm font-bold border border-red-100">
                  {deleteError()}
                </div>
              </Show>

              <div class="flex gap-3">
                <button
                  onClick={() => { setDeleteId(null); setDeleteError(null); }}
                  class="flex-1 bg-gray-100 text-gray-700 py-3 rounded-2xl font-black hover:bg-gray-200 transition-colors"
                >
                  Cancelar
                </button>
                <button
                  onClick={confirmDelete}
                  class="flex-1 bg-red-500 text-white py-3 rounded-2xl font-black hover:bg-red-600 active:scale-95 transition-all"
                >
                  Eliminar
                </button>
              </div>
            </div>
          </div>
        </Show>

        {/* ── FORMULARIO ────────────────────────────────────────────────────── */}
        <Show when={showForm()}>
          <form onSubmit={handleSubmit} class="bg-white rounded-[2.5rem] p-8 shadow-2xl border-2 border-colpsi-yellow animate-in slide-in-from-top-4 duration-300">
            <h2 class="text-xl font-black text-colpsi-blue mb-6">
              {isEditing() ? "✏️ Editando Título" : "📜 Registrar Postgrado"}
            </h2>

            <Show when={formError()}>
              <div class="mb-6 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500">
                {formError()}
              </div>
            </Show>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="space-y-1">
                <label class="text-xs font-black text-gray-400 uppercase ml-2">Título Obtenido</label>
                <input type="text" required value={form.title}
                  onInput={(e) => setForm("title", e.currentTarget.value)}
                  class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-black text-gray-400 uppercase ml-2">Universidad</label>
                <input type="text" required value={form.university}
                  onInput={(e) => setForm("university", e.currentTarget.value)}
                  class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-black text-gray-400 uppercase ml-2">Año de Egreso</label>
                <input type="number" required value={form.graduation_year}
                  onInput={(e) => setForm("graduation_year", e.currentTarget.value)}
                  class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none" />
              </div>
              <div class="md:col-span-2 space-y-1">
                <label class="text-xs font-black text-gray-400 uppercase ml-2">Breve descripción</label>
                <textarea value={form.description}
                  onInput={(e) => setForm("description", e.currentTarget.value)}
                  class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none" />
              </div>
              <div class="md:col-span-2 space-y-4">
                <label class="text-xs font-black text-gray-400 uppercase ml-2">Soportes Digitales (PDF o Imagen)</label>
                <div class="grid grid-cols-3 gap-4">
                  <FileUploader id="pic_one"   label="Título" onChange={(e) => handleFileChange(e, "pic_one")}   file={files().pic_one} />
                  <FileUploader id="pic_two"   label="Notas"  onChange={(e) => handleFileChange(e, "pic_two")}   file={files().pic_two} />
                  <FileUploader id="pic_three" label="Extra"  onChange={(e) => handleFileChange(e, "pic_three")} file={files().pic_three} />
                </div>
              </div>
              <div class="md:col-span-2 pt-4">
                <button type="submit" disabled={loading()}
                  class="w-full bg-colpsi-blue text-white py-4 rounded-2xl font-black shadow-lg disabled:opacity-70">
                  {loading() ? "PROCESANDO..." : isEditing() ? "GUARDAR CAMBIOS" : "AÑADIR A MI PERFIL"}
                </button>
              </div>
            </div>
          </form>
        </Show>

        {/* ── LISTADO ───────────────────────────────────────────────────────── */}
        <Suspense fallback={<div class="h-40 bg-white animate-pulse rounded-[2.5rem]" />}>
          <div class="space-y-6">
            <For each={profile()?.post_grades ?? []}>
              {(pg) => (
                <div class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100 flex flex-col md:flex-row gap-6 relative group overflow-hidden">

                  <div class="absolute top-6 right-6 flex gap-2">
                    <button
                      onClick={() => openEdit(pg)}
                      class="w-10 h-10 bg-blue-50 text-colpsi-blue rounded-xl flex items-center justify-center hover:bg-colpsi-yellow transition-colors"
                    >✏️</button>
                    <button
                      onClick={() => { setDeleteError(null); setDeleteId(pg.id); }}
                      class="w-10 h-10 bg-red-50 text-red-500 rounded-xl flex items-center justify-center hover:bg-red-500 hover:text-white transition-colors"
                    >🗑️</button>
                  </div>

                  <div class="flex-grow space-y-3">
                    <div class="flex items-center gap-3">
                      <span class="bg-blue-100 text-colpsi-blue text-[10px] font-black px-2 py-1 rounded-md uppercase">Postgrado</span>
                      <span class="text-xs font-bold text-gray-400">{pg.post_grade_graduation_year}</span>
                    </div>
                    <h3 class="text-xl font-black text-colpsi-blue uppercase leading-tight">{pg.post_grade_title}</h3>
                    <p class="text-sm font-bold text-gray-500 italic">{pg.post_grade_university}</p>

                    <Show when={pg.post_grade_description}>
                      <p class="text-xs text-gray-400 line-clamp-2">{pg.post_grade_description}</p>
                    </Show>

                    <div class="flex gap-3 pt-4">
                      <For each={[pg.pic_one_url, pg.pic_two_url, pg.pic_three_url]}>
                        {(url) => (
                          <Show when={url}>
                            <a
                              href={bucketUrl(url)}
                              target="_blank" rel="noopener noreferrer"
                              class="w-16 h-20 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden hover:ring-2 hover:ring-colpsi-yellow transition-all"
                            >
                              <img src={bucketUrl(url)} class="w-full h-full object-cover" />
                            </a>
                          </Show>
                        )}
                      </For>
                    </div>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Suspense>
      </div>
    </main>
  );
}
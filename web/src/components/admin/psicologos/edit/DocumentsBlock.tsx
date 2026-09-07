// web/src/components/admin/psicologos/edit/DocumentsBlock.tsx

import { For, Show, createSignal } from "solid-js";
import { createStore } from "solid-js/store";
import { bucketUrl } from "~/lib/bucket";
import {
  DOCUMENT_TYPE_LABELS,
  DOCUMENT_TYPE_EMOJI,
  isPdf,
} from "~/lib/doc_tipos";
import { getUserFacingError } from "~/lib/errors";
import FlatDatePicker from "~/components/ui/FlatDatePicker";
import type { PsiUserDocument, PsiUserDocumentType } from "~/types/psi";

interface Props {
  entries: PsiUserDocument[] | undefined;
  onAdd: (payload: FormData) => Promise<void>;
  onUpdate: (docId: string, payload: FormData) => Promise<void>;
  onDelete: (docId: string) => Promise<void>;
}

const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
const FIELD = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";

const formatDate = (dateStr?: string | null) => {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return "";
  return d.toLocaleDateString("es-VE", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
};

export function DocumentsBlock(props: Props) {
  // ── Formulario de alta ─────────────────────────────────────────────────────
  const [showForm, setShowForm] = createSignal(false);
  const [form, setForm] = createStore<{
    title: string;
    document_type: PsiUserDocumentType;
    notes: string;
    document_date: string;
  }>({ title: "", document_type: "otro", notes: "", document_date: "" });
  const [file, setFile] = createSignal<File | null>(null);
  const [saving, setSaving] = createSignal(false);

  // ── Edición inline ─────────────────────────────────────────────────────────
  const [editingId, setEditingId] = createSignal<string | null>(null);
  const [edit, setEdit] = createStore<{
    title: string;
    document_type: PsiUserDocumentType;
    notes: string;
    document_date: string;
  }>({ title: "", document_type: "otro", notes: "", document_date: "" });
  const [editFile, setEditFile] = createSignal<File | null>(null);
  const [clearDate, setClearDate] = createSignal(false);
  const [savingEdit, setSavingEdit] = createSignal(false);

  const [error, setError] = createSignal<string | null>(null);

  const pickFile = (e: Event) => {
    const input = e.currentTarget as HTMLInputElement;
    setFile(input.files?.[0] ?? null);
  };

  const resetAddForm = () => {
    setForm({ title: "", document_type: "otro", notes: "", document_date: "" });
    setFile(null);
    setError(null);
  };

  const handleAdd = async (e: Event) => {
    e.preventDefault();
    const fd = new FormData();
    if (!file()) {
      setError("Debes adjuntar el archivo del documento (imagen o PDF).");
      return;
    }
    fd.append("file", file()!);
    fd.append("document_type", form.document_type || "otro");
    fd.append("title", form.title.trim());
    if (form.notes.trim()) fd.append("notes", form.notes.trim());
    if (form.document_date) fd.append("document_date", form.document_date);

    setSaving(true);
    setError(null);
    try {
      await props.onAdd(fd);
      resetAddForm();
      setShowForm(false);
    } catch (err: unknown) {
      setError(getUserFacingError(err));
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (doc: PsiUserDocument) => {
    if (!doc.id) return;
    setEditingId(doc.id);
    setEdit({
      title: doc.title ?? "",
      document_type: (doc.document_type as PsiUserDocumentType) ?? "otro",
      notes: doc.notes ?? "",
      document_date: (doc.document_date ?? "").slice(0, 10),
    });
    setEditFile(null);
    setClearDate(false);
    setError(null);
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditFile(null);
    setClearDate(false);
    setError(null);
  };

  const saveEdit = async (doc: PsiUserDocument) => {
    if (!doc.id || savingEdit()) return;
    const fd = new FormData();
    fd.append("title", edit.title.trim());
    fd.append("document_type", edit.document_type || "otro");
    if (edit.notes.trim()) fd.append("notes", edit.notes.trim());
    if (edit.document_date) fd.append("document_date", edit.document_date);
    if (clearDate()) fd.append("clear_document_date", "1");
    if (editFile()) fd.append("file", editFile()!);

    setSavingEdit(true);
    setError(null);
    try {
      await props.onUpdate(doc.id, fd);
      cancelEdit();
    } catch (err: unknown) {
      setError(getUserFacingError(err));
    } finally {
      setSavingEdit(false);
    }
  };

  const handleDelete = async (doc: PsiUserDocument) => {
    if (!doc.id || !confirm(`¿Eliminar "${doc.title}" del expediente?`)) return;
    setError(null);
    try {
      await props.onDelete(doc.id);
    } catch (err: unknown) {
      setError(getUserFacingError(err));
    }
  };

  return (
    <div>
      <p class="text-xs text-gray-500 mb-6">
        Registro digital del expediente del psicólogo (CI, título, RIF,
        comprobantes de solvencia, etc.). La administración puede cargar, editar
        y eliminar; el psicólogo solo puede consultarlos desde su portal.
        Acepta imágenes (se optimizan a WebP) y PDF (máx. 4 MB).
      </p>

      <Show when={error()}>
        <div class="mb-6 p-4 rounded-2xl bg-red-50 text-red-700 font-bold text-sm border border-red-100">
          {error()}
        </div>
      </Show>

      {/* ── LISTADO ───────────────────────────────────────────────────────── */}
      <Show when={(props.entries?.length ?? 0) > 0}>
        <div class="mb-6 space-y-3">
          <For each={props.entries}>
            {(doc: PsiUserDocument) => (
              <div class="bg-colpsi-surface hover:bg-white p-4 rounded-2xl border border-colpsi-border hover:border-blue-100 transition-colors">
                <Show
                  when={editingId() === doc.id}
                  fallback={
                    <div class="flex flex-col md:flex-row gap-4">
                      <a
                        href={bucketUrl(doc.document_url)}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="shrink-0 w-full md:w-24 h-32 md:h-24 bg-white rounded-xl border border-gray-200 overflow-hidden flex items-center justify-center hover:ring-2 hover:ring-colpsi-yellow transition-all"
                      >
                        <Show
                          when={!isPdf(doc)}
                          fallback={
                            <div class="flex flex-col items-center justify-center text-red-500">
                              <span class="text-2xl">📕</span>
                              <span class="text-[9px] font-black uppercase mt-1">
                                PDF
                              </span>
                            </div>
                          }
                        >
                          <img
                            src={bucketUrl(doc.document_url)}
                            alt={doc.title}
                            class="w-full h-full object-cover"
                          />
                        </Show>
                      </a>
                      <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2 flex-wrap">
                          <span class="bg-white px-2 py-0.5 rounded-lg text-[10px] font-black text-blue-800 shadow-sm border border-colpsi-border">
                            {DOCUMENT_TYPE_EMOJI[doc.document_type ?? "otro"]}{" "}
                            {DOCUMENT_TYPE_LABELS[doc.document_type ?? "otro"]}
                          </span>
                          {doc.title && (
                            <h3 class="text-sm font-black text-gray-800">
                              {doc.title}
                            </h3>
                          )}
                        </div>
                        <Show when={doc.notes}>
                          <p class="text-xs text-gray-500 mt-1 line-clamp-2">
                            {doc.notes}
                          </p>
                        </Show>
                        <div class="flex items-center gap-3 mt-2 flex-wrap text-[10px] font-bold text-gray-400">
                          {doc.filename && (
                            <span class="truncate max-w-[180px]">📎 {doc.filename}</span>
                          )}
                          {formatDate(doc.document_date) && (
                            <span>{formatDate(doc.document_date)}</span>
                          )}
                          {doc.create_by && <span>por {doc.create_by}</span>}
                        </div>
                      </div>
                      <div class="flex md:flex-col gap-2 shrink-0">
                        <button
                          onClick={() => startEdit(doc)}
                          class="text-gray-400 hover:text-blue-600 hover:bg-blue-50 p-2 rounded-xl transition-colors"
                          title="Editar"
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                            <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                          </svg>
                        </button>
                        <button
                          onClick={() => handleDelete(doc)}
                          class="text-gray-400 hover:text-red-500 hover:bg-red-50 p-2 rounded-xl transition-colors"
                          title="Eliminar"
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                            <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  }
                >
                  <div class="space-y-4">
                    <div class="flex items-center justify-between">
                      <h4 class="text-sm font-black text-blue-800 uppercase">
                        Editando: {doc.title}
                      </h4>
                      <button
                        onClick={cancelEdit}
                        class="text-xs font-bold text-gray-500 hover:text-gray-700 hover:bg-gray-200 p-2 rounded-xl transition-colors"
                      >
                        Cancelar
                      </button>
                    </div>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label class={FIELD}>Etiqueta</label>
                        <input
                          type="text"
                          required
                          value={edit.title}
                          onInput={(e) => setEdit("title", e.currentTarget.value)}
                          class={IC}
                        />
                      </div>
                      <div>
                        <label class={FIELD}>Categoría</label>
                        <select
                          value={edit.document_type}
                          onChange={(e) =>
                            setEdit("document_type", e.currentTarget.value as PsiUserDocumentType)
                          }
                          class={IC}
                        >
                          <For each={Object.entries(DOCUMENT_TYPE_LABELS)}>
                            {([value, label]) => (
                              <option value={value}>{label}</option>
                            )}
                          </For>
                        </select>
                      </div>
                      <div>
                        <label class={FIELD}>Fecha del documento</label>
                        <FlatDatePicker
                          value={edit.document_date}
                          onChange={(v) => setEdit("document_date", v)}
                          class={IC}
                        />
                        <label class="flex items-center gap-2 mt-2 text-xs font-bold text-gray-500 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={clearDate()}
                            onChange={(e) => setClearDate(e.currentTarget.checked)}
                          />
                          Vaciar fecha
                        </label>
                      </div>
                      <div>
                        <label class={FIELD}>Reemplazar archivo (opcional)</label>
                        <label class="flex items-center justify-center gap-2 border-2 border-dashed border-gray-300 rounded-xl px-4 py-2.5 cursor-pointer hover:border-blue-400 transition-colors bg-white text-sm text-gray-500 font-bold">
                          <input
                            type="file"
                            accept=".pdf,image/*"
                            class="sr-only"
                            onChange={(e) => {
                              const input = e.currentTarget;
                              setEditFile(input.files?.[0] ?? null);
                            }}
                          />
                          {editFile() ? "✓ " + editFile()!.name : "📎 Elegir archivo"}
                        </label>
                      </div>
                      <div class="md:col-span-2">
                        <label class={FIELD}>Notas internas</label>
                        <textarea
                          value={edit.notes}
                          onInput={(e) => setEdit("notes", e.currentTarget.value)}
                          rows={2}
                          class={IC}
                        />
                      </div>
                    </div>
                    <div class="flex justify-end">
                      <button
                        onClick={() => saveEdit(doc)}
                        disabled={savingEdit() || !edit.title.trim()}
                        class="bg-blue-800 text-white px-6 py-2.5 rounded-xl text-sm font-bold hover:bg-blue-900 active:scale-95 transition-all shadow-md disabled:opacity-70"
                      >
                        {savingEdit() ? "GUARDANDO..." : "GUARDAR CAMBIOS"}
                      </button>
                    </div>
                  </div>
                </Show>
              </div>
            )}
          </For>
        </div>
      </Show>

      <Show when={!props.entries?.length}>
        <p class="text-sm italic text-gray-400 mb-5">
          Sin documentos en el expediente.
        </p>
      </Show>

      {/* ── FORMULARIO DE ALTA ─────────────────────────────────────────────── */}
      <Show
        when={showForm()}
        fallback={
          <button
            onClick={() => { setShowForm(true); setError(null); }}
            class="bg-blue-800 text-white px-8 py-3 rounded-xl font-bold hover:bg-blue-900 active:scale-95 transition-all shadow-md"
          >
            + REGISTRAR DOCUMENTO
          </button>
        }
      >
        <form onSubmit={handleAdd} class="bg-blue-50/50 p-5 rounded-2xl border border-blue-100 space-y-4">
          <div class="flex items-center justify-between">
            <h4 class="text-sm font-black text-blue-800 uppercase">Nuevo documento</h4>
            <button
              type="button"
              onClick={() => { setShowForm(false); resetAddForm(); }}
              class="text-xs font-bold text-gray-500 hover:text-gray-700 hover:bg-gray-200 p-2 rounded-xl transition-colors"
            >
              Cerrar
            </button>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class={FIELD}>Etiqueta descriptiva</label>
              <input
                type="text"
                required
                placeholder="Ej: Cédula V-123456 anverso"
                value={form.title}
                onInput={(e) => setForm("title", e.currentTarget.value)}
                class={IC}
              />
            </div>
            <div>
              <label class={FIELD}>Categoría</label>
              <select
                value={form.document_type}
                onChange={(e) =>
                  setForm("document_type", e.currentTarget.value as PsiUserDocumentType)
                }
                class={IC}
              >
                <For each={Object.entries(DOCUMENT_TYPE_LABELS)}>
                  {([value, label]) => (
                    <option value={value}>{label}</option>
                  )}
                </For>
              </select>
            </div>
            <div class="md:col-span-2">
              <label class={FIELD}>Archivo (imagen o PDF · máx. 4 MB)</label>
              <label class="flex items-center gap-3 border-2 border-dashed border-gray-300 rounded-xl px-4 py-4 cursor-pointer hover:border-blue-400 transition-colors bg-white text-sm text-gray-500 font-bold">
                <input
                  type="file"
                  required
                  accept=".pdf,image/*"
                  class="sr-only"
                  onChange={pickFile}
                />
                {file() ? "✓ " + file()!.name : "📎 Seleccionar archivo"}
              </label>
            </div>
            <div>
              <label class={FIELD}>Fecha del documento (opcional)</label>
              <FlatDatePicker
                value={form.document_date}
                onChange={(v) => setForm("document_date", v)}
                class={IC}
              />
            </div>
            <div>
              <label class={FIELD}>Notas internas (opcional)</label>
              <input
                type="text"
                placeholder="Ej: Solicitar certificación al Ministerio"
                value={form.notes}
                onInput={(e) => setForm("notes", e.currentTarget.value)}
                class={IC}
              />
            </div>
          </div>
          <div class="flex justify-end">
            <button
              type="submit"
              disabled={saving()}
              class="bg-blue-800 text-white px-8 py-3 rounded-xl font-bold hover:bg-blue-900 active:scale-95 transition-all shadow-md disabled:opacity-70"
            >
              {saving() ? "SUBIR..." : "SUBIR DOCUMENTO"}
            </button>
          </div>
        </form>
      </Show>
    </div>
  );
}
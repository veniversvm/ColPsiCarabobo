// web/src/components/admin/psicologos/edit/DeontologiaBlock.tsx

import { For, Show, createSignal } from "solid-js";
import type { DeontologiaEntry } from "./types";

interface Props {
  entries: DeontologiaEntry[] | undefined;
  onAdd: (content: string) => Promise<void>;
  onUpdate: (entryId: string, content: string) => Promise<void>;
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return "";
  return new Date(dateStr).toLocaleDateString("es-VE", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
};

export function DeontologiaBlock(props: Props) {
  const [content, setContent] = createSignal("");
  const [saving, setSaving] = createSignal(false);

  // Edición inline
  const [editingId, setEditingId] = createSignal<string | null>(null);
  const [editContent, setEditContent] = createSignal("");
  const [savingEdit, setSavingEdit] = createSignal(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    const trimmed = content().trim();
    if (!trimmed || saving()) return;
    setSaving(true);
    try {
      await props.onAdd(trimmed);
      setContent("");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (entry: DeontologiaEntry) => {
    if (!entry.id) return;
    setEditingId(entry.id);
    setEditContent(entry.content);
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditContent("");
  };

  const saveEdit = async (entry: DeontologiaEntry) => {
    if (!entry.id || savingEdit()) return;
    const trimmed = editContent().trim();
    if (!trimmed) return;
    setSavingEdit(true);
    try {
      await props.onUpdate(entry.id, trimmed);
      cancelEdit();
    } finally {
      setSavingEdit(false);
    }
  };

  return (
    <div>
      <p class="text-xs text-gray-500 mb-6">
        Registro interno del Colegio. Confidencial: el psicólogo nunca ve estas entradas.
        Es un expediente histórico: las entradas se pueden corregir, pero no eliminar.
      </p>

      {/* Listado existente */}
      <Show when={(props.entries?.length ?? 0) > 0}>
        <div class="mb-6 space-y-3">
          <For each={props.entries}>
            {(entry: DeontologiaEntry) => (
              <div class="bg-colpsi-surface hover:bg-white p-4 rounded-2xl border border-colpsi-border hover:border-blue-100 transition-colors group">
                <Show
                  when={editingId() === entry.id}
                  fallback={
                    <div class="flex items-start justify-between gap-3">
                      <div class="flex-1 overflow-hidden">
                        <p class="text-sm text-gray-700 whitespace-pre-wrap break-words">
                          {entry.content}
                        </p>
                        <p class="text-xs text-gray-400 mt-2">
                          {formatDate(entry.created_at)}
                          <Show when={entry.create_by}>
                            {" · "}
                            {entry.create_by}
                          </Show>
                        </p>
                      </div>
                      <button
                        onClick={() => startEdit(entry)}
                        class="text-gray-400 hover:text-blue-600 hover:bg-blue-50 p-2 rounded-xl transition-colors shrink-0"
                        title="Editar"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                          <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                        </svg>
                      </button>
                    </div>
                  }
                >
                  <div>
                    <textarea
                      value={editContent()}
                      onInput={(e) => setEditContent(e.currentTarget.value)}
                      rows={3}
                      class="w-full bg-white border-2 border-blue-300 rounded-xl px-4 py-3 outline-none text-sm shadow-sm transition-all resize-y"
                    />
                    <div class="flex justify-end gap-2 mt-3">
                      <button
                        onClick={cancelEdit}
                        class="bg-gray-100 text-gray-600 px-4 py-2 rounded-xl text-sm font-bold hover:bg-gray-200 transition-colors"
                      >
                        Cancelar
                      </button>
                      <button
                        onClick={() => saveEdit(entry)}
                        disabled={savingEdit() || !editContent().trim()}
                        class="bg-blue-800 text-white px-5 py-2 rounded-xl text-sm font-bold hover:bg-blue-900 active:scale-95 transition-all shadow-md disabled:opacity-70"
                      >
                        {savingEdit() ? "..." : "GUARDAR"}
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
          Sin entradas deontológicas registradas.
        </p>
      </Show>

      {/* Formulario de alta */}
      <form onSubmit={handleSubmit} class="bg-blue-50/50 p-5 rounded-2xl border border-blue-100">
        <textarea
          placeholder="Describe el expediente, sanción o nota interna..."
          required
          value={content()}
          onInput={(e) => setContent(e.currentTarget.value)}
          rows={4}
          class="w-full bg-white border-2 border-transparent focus:border-blue-500 rounded-xl px-4 py-3 outline-none text-sm shadow-sm transition-all resize-y"
        />
        <div class="flex justify-end mt-3">
          <button
            type="submit"
            disabled={saving()}
            class="bg-blue-800 text-white px-8 py-3 rounded-xl font-bold hover:bg-blue-900 active:scale-95 transition-all shadow-md disabled:opacity-70"
          >
            {saving() ? "..." : "REGISTRAR ENTRADA"}
          </button>
        </div>
      </form>
    </div>
  );
}

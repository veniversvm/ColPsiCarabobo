// web/src/components/admin/psicologos/edit/DeontologiaBlock.tsx

import { For, Show, createSignal } from "solid-js";
import type { DeontologiaEntry } from "./types";

interface Props {
  entries: DeontologiaEntry[] | undefined;
  onAdd: (content: string) => Promise<void>;
  onDelete: (entryId: string) => Promise<void>;
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

  return (
    <div>
      <p class="text-xs text-gray-500 mb-6">
        Registro interno del Colegio. Confidencial: el psicólogo nunca ve estas entradas.
      </p>

      {/* Listado existente */}
      <Show when={(props.entries?.length ?? 0) > 0}>
        <div class="mb-6 space-y-3">
          <For each={props.entries}>
            {(entry: DeontologiaEntry) => (
              <div class="flex items-start justify-between gap-3 bg-gray-50 hover:bg-white p-4 rounded-2xl border border-gray-100 hover:border-blue-100 transition-colors group">
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
                  onClick={() => entry.id && props.onDelete(entry.id)}
                  class="text-gray-400 hover:text-red-500 hover:bg-red-50 p-2 rounded-xl transition-colors shrink-0"
                  title="Eliminar"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                  </svg>
                </button>
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

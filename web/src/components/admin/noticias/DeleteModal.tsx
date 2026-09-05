// web/src/components/admin/noticias/DeleteModal.tsx
import { Show } from "solid-js";

interface Props {
  isOpen: boolean;
  isBusy: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function DeleteModal(props: Props) {
  return (
    <Show when={props.isOpen}>
      <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
        onClick={(e) => { if (e.target === e.currentTarget) props.onCancel(); }}
      >
        <div class="bg-white rounded-3xl shadow-2xl p-8 w-full max-w-sm border border-gray-100 text-center animate-in zoom-in-95 duration-200">
          <p class="text-4xl mb-4">🗑️</p>
          <h2 class="text-lg font-black text-gray-900 mb-2">¿Archivar publicación?</h2>
          <p class="text-gray-500 text-sm mb-6">El post quedará oculto. Puedes restaurarlo desde el editor.</p>
          <div class="flex gap-3">
            <button
              onClick={props.onCancel}
              class="flex-1 px-4 py-3 rounded-2xl border-2 border-gray-200 font-black text-gray-600 hover:bg-gray-50 transition-all text-sm"
            >
              Cancelar
            </button>
            <button
              onClick={props.onConfirm}
              disabled={props.isBusy}
              class="flex-1 px-4 py-3 rounded-2xl bg-red-600 text-white font-black hover:bg-red-700 active:scale-95 transition-all text-sm disabled:opacity-60"
            >
              {props.isBusy ? "Eliminando..." : "Sí, archivar"}
            </button>
          </div>
        </div>
      </div>
    </Show>
  );
}
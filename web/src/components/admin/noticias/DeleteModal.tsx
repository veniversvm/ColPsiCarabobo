// web/src/components/admin/noticias/DeleteModal.tsx
import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";

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
        <div class="bg-white rounded-lg shadow-lg p-6 w-full max-w-sm border border-colpsi-border">
          <span class="inline-flex h-10 w-10 items-center justify-center rounded-md bg-red-50 text-colpsi-red mb-3">
            <Icon name="trash" class="w-5 h-5" />
          </span>
          <h2 class="text-base font-semibold text-colpsi-text mb-1">¿Archivar publicación?</h2>
          <p class="text-colpsi-muted text-sm mb-5">El post quedará oculto. Puedes restaurarlo desde el editor.</p>
          <div class="flex gap-2">
            <button
              onClick={props.onCancel}
              class="flex-1 h-9 px-4 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors text-sm"
            >
              Cancelar
            </button>
            <button
              onClick={props.onConfirm}
              disabled={props.isBusy}
              class="flex-1 h-9 px-4 rounded-md bg-colpsi-red text-white font-semibold hover:opacity-90 transition-colors text-sm disabled:opacity-60"
            >
              {props.isBusy ? "Eliminando..." : "Sí, archivar"}
            </button>
          </div>
        </div>
      </div>
    </Show>
  );
}
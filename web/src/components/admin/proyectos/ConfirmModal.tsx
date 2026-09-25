// web/src/components/admin/proyectos/ConfirmModal.tsx
import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";

export default function ConfirmModal(props: {
  title: string;
  message: string;
  confirmLabel?: string;
  danger?: boolean;
  busy?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}) {
  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && !props.busy && props.onClose()}
    >
      <div class="bg-white rounded-lg shadow-lg p-6 w-full max-w-sm border border-colpsi-border animate-in zoom-in-95">
        <span class={`inline-flex h-10 w-10 items-center justify-center rounded-md mb-3 ${props.danger ? "bg-red-50 text-colpsi-red" : "bg-colpsi-blue/10 text-colpsi-blue"}`}>
          <Icon name={props.danger ? "alertTriangle" : "info"} class="w-5 h-5" />
        </span>
        <h3 class="text-base font-semibold text-colpsi-text">{props.title}</h3>
        <p class="mt-1.5 text-sm text-colpsi-muted">{props.message}</p>
        <div class="mt-5 grid grid-cols-2 gap-2">
          <button
            onClick={props.onClose}
            disabled={props.busy}
            class="h-9 rounded-md bg-white text-colpsi-text border border-colpsi-border font-medium hover:bg-colpsi-bg disabled:opacity-60 transition-colors text-sm"
          >
            Cancelar
          </button>
          <button
            onClick={props.onConfirm}
            disabled={props.busy}
            class={`h-9 rounded-md font-semibold text-white disabled:opacity-60 transition-colors text-sm ${props.danger ? "bg-colpsi-red hover:opacity-90" : "bg-colpsi-blue hover:bg-colpsi-blue-light"}`}
          >
            {props.busy ? "..." : (props.confirmLabel ?? "Confirmar")}
          </button>
        </div>
      </div>
    </div>
  );
}
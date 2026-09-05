// web/src/components/admin/proyectos/ConfirmModal.tsx
import { Show } from "solid-js";

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
      <div class="bg-white rounded-3xl shadow-2xl p-6 w-full max-w-sm text-center animate-in zoom-in-95">
        <div class="w-14 h-14 mx-auto rounded-2xl bg-red-50 flex items-center justify-center text-2xl mb-3">
          ⚠️
        </div>
        <h3 class="text-lg font-black text-gray-800">{props.title}</h3>
        <p class="mt-2 text-sm text-gray-500">{props.message}</p>
        <div class="mt-6 grid grid-cols-2 gap-3">
          <button
            onClick={props.onClose}
            disabled={props.busy}
            class="bg-white text-gray-600 border-2 border-gray-200 rounded-xl font-black py-3 hover:bg-gray-50 disabled:opacity-60"
          >
            Cancelar
          </button>
          <button
            onClick={props.onConfirm}
            disabled={props.busy}
            class={`rounded-xl font-black py-3 text-white disabled:opacity-60 ${props.danger ? "bg-red-600 hover:bg-red-700" : "bg-blue-800 hover:bg-blue-900"}`}
          >
            {props.busy ? "..." : (props.confirmLabel ?? "Confirmar")}
          </button>
        </div>
      </div>
    </div>
  );
}
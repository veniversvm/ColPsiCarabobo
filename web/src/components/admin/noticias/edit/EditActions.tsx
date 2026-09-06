// web/src/components/admin/noticias/edit/EditActions.tsx
import { PostStatus } from "./types";

interface Props {
  saving: boolean;
  status: PostStatus;
  onCancel: () => void;
}

export function EditActions(props: Props) {
  const buttonText = () => {
    if (props.saving) return "GUARDANDO...";
    if (props.status === "published") return "💾 GUARDAR Y PUBLICAR";
    if (props.status === "scheduled") return "⏰ PROGRAMAR";
    return "💾 GUARDAR";
  };

  return (
    <div class="sticky bottom-6 z-50 flex justify-end gap-3">
      <button
        type="button"
        onClick={props.onCancel}
        class="bg-white text-gray-600 border-2 border-gray-200 px-6 py-4 rounded-2xl font-black hover:bg-colpsi-surface transition-all text-sm"
      >Cancelar</button>
      <button
        type="submit"
        disabled={props.saving}
        class="bg-blue-800 text-white px-10 py-4 rounded-2xl font-black shadow-2xl hover:scale-105 active:scale-95 transition-all disabled:opacity-70 flex items-center gap-3 border-2 border-white text-sm"
      >
        {buttonText()}
      </button>
    </div>
  );
}
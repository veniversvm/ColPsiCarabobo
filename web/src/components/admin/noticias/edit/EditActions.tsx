// web/src/components/admin/noticias/edit/EditActions.tsx
import { Show } from "solid-js";
import { PostStatus } from "./types";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  saving: boolean;
  status: PostStatus;
  onCancel: () => void;
}

export function EditActions(props: Props) {
  const buttonText = () => {
    if (props.saving) return "Guardando...";
    if (props.status === "published") return "Guardar y publicar";
    if (props.status === "scheduled") return "Programar";
    return "Guardar";
  };

  return (
    <div class="sticky bottom-4 z-50 flex justify-end gap-2">
      <button
        type="button"
        onClick={props.onCancel}
        class="h-10 px-4 rounded-md border border-colpsi-border bg-white text-sm font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors"
      >Cancelar</button>
      <button
        type="submit"
        disabled={props.saving}
        class="inline-flex items-center justify-center gap-2 h-10 px-5 rounded-md bg-colpsi-blue text-white text-sm font-semibold transition-colors hover:bg-colpsi-blue-light disabled:opacity-60"
      >
        <Show when={props.saving} fallback={<Icon name="check" />}>
          <span class="animate-spin h-4 w-4 border-2 border-white/40 border-t-white rounded-full" />
        </Show>
        {buttonText()}
      </button>
    </div>
  );
}
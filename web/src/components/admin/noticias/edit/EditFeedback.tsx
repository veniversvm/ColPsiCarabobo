// web/src/components/admin/noticias/edit/EditFeedback.tsx
import { Show } from "solid-js";

interface Props {
  error: string | null;
  success: boolean;
}

export function EditFeedback(props: Props) {
  return (
    <>
      <Show when={props.error}>
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">
          {props.error}
        </div>
      </Show>
      <Show when={props.success}>
        <div class="p-3 rounded-md bg-emerald-50 text-emerald-700 border border-emerald-200 text-sm font-medium">
          Publicación actualizada correctamente. Redirigiendo...
        </div>
      </Show>
    </>
  );
}
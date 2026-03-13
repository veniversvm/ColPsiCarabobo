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
        <div class="mb-6 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500 shadow-sm">
          {props.error}
        </div>
      </Show>
      <Show when={props.success}>
        <div class="mb-6 p-4 rounded-2xl bg-emerald-50 text-emerald-800 font-bold text-sm border-l-4 border-emerald-500 shadow-sm">
          ✓ Publicación actualizada correctamente. Redirigiendo...
        </div>
      </Show>
    </>
  );
}
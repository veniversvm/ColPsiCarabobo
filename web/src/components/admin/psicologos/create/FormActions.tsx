// web/src/components/admin/psicologos/create/FormActions.tsx
import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  saving: boolean;
}

export function FormActions(props: Props) {
  return (
    <button
      type="submit"
      disabled={props.saving}
      class="inline-flex items-center justify-center gap-2 h-10 px-5 rounded-md bg-colpsi-blue text-white text-sm font-semibold transition-colors hover:bg-colpsi-blue-light disabled:opacity-60 disabled:pointer-events-none disabled:cursor-not-allowed"
    >
      <Show when={props.saving} fallback={<Icon name="check" />}>
        <span class="animate-spin h-4 w-4 border-2 border-white/40 border-t-white rounded-full" />
      </Show>
      {props.saving ? "Procesando..." : "Registrar Expediente"}
    </button>
  );
}
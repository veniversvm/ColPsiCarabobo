// web/src/components/psi/profile/SaveButton.tsx
import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";

interface SaveButtonProps {
  saving: boolean;
  onClick?: () => void;
}

export function SaveButton(props: SaveButtonProps) {
  return (
    <div class="sticky bottom-6 z-50 flex justify-end px-2">
      <button
        type="submit"
        disabled={props.saving}
        class="inline-flex items-center gap-2 h-11 px-6 rounded-md bg-colpsi-blue text-white text-sm font-semibold shadow-sm hover:bg-colpsi-blue-light transition-colors disabled:opacity-70 disabled:cursor-not-allowed"
      >
        <Show
          when={props.saving}
          fallback={
            <>
              <Icon name="check" class="w-4 h-4" />
              Guardar cambios
            </>
          }
        >
          <svg class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Guardando...
        </Show>
      </button>
    </div>
  );
}
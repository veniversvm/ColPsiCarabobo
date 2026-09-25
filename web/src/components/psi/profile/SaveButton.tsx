// web/src/components/psi/profile/SaveButton.tsx
// Barra inferior sticky de guardado del perfil: feedback (éxito/error) a la
// izquierda y botón de guardar a la derecha. El estado "saving" y el mensaje
// los maneja el padre; aquí solo se muestra el spinner y se deshabilita el botón.
import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";

interface SaveButtonProps {
  saving: boolean;
  message?: { type: "success" | "error"; text: string } | null;
  onClick?: () => void;
}

export function SaveButton(props: SaveButtonProps) {
  return (
    <div class="sticky bottom-0 z-40 -mx-4 md:-mx-8 border-t border-colpsi-border bg-white/95 backdrop-blur-sm px-4 md:px-8 py-3">
      <div class="flex flex-col sm:flex-row sm:items-center gap-3">
        <Show
          when={props.message?.text}
          fallback={
            <span class="inline-flex items-center gap-2 text-xs text-colpsi-muted">
              <span class="w-1.5 h-1.5 rounded-full bg-colpsi-yellow" />
              Los cambios se guardan para todo el expediente al pulsar el botón.
            </span>
          }
        >
          <span
            class={`inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm font-medium ${
              props.message!.type === "success"
                ? "bg-emerald-50 text-emerald-700 border-emerald-200"
                : "bg-red-50 text-red-700 border-red-200"
            }`}
          >
            {props.message!.text}
          </span>
        </Show>

        <button
          type="submit"
          disabled={props.saving}
          class="inline-flex items-center gap-2 h-10 px-5 rounded-md bg-colpsi-blue text-white text-sm font-semibold hover:bg-colpsi-blue-light transition-colors disabled:opacity-70 disabled:cursor-not-allowed sm:ml-auto"
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
            <svg
              class="animate-spin h-4 w-4 text-white"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            Guardando...
          </Show>
        </button>
      </div>
    </div>
  );
}
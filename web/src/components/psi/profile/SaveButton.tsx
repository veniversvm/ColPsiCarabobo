// web/src/components/psi/profile/SaveButton.tsx
// Tarjeta de guardado del perfil, del mismo ancho que el notebook: incluye el
// campo "Contraseña Actual" (obligatoria para guardar) y el botón de guardar.
// El feedback (éxito/error) se muestra debajo, junto a la acción.
import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";
import { PasswordInputComponent } from "~/components/ui/PasswordInput";

interface SaveButtonProps {
  saving: boolean;
  password: string;
  message?: { type: "success" | "error"; text: string } | null;
  onPasswordChange: (value: string) => void;
  onClick?: () => void;
}

export function SaveButton(props: SaveButtonProps) {
  return (
    <div class="bg-white rounded-lg border border-colpsi-border overflow-hidden">
      <div class="p-4 md:p-5">
        <div class="flex flex-col md:flex-row md:items-end gap-4">
          <div class="flex-1 min-w-0 md:max-w-sm space-y-1">
            <label class="text-[11px] font-semibold text-colpsi-red uppercase tracking-wide ml-1 mb-1">
                Contraseña Actual{" "}
                <span class="lowercase font-medium text-colpsi-muted">
                  (obligatoria para guardar)
                </span>
              </label>

              <PasswordInputComponent
                required
                value={props.password}
                onInput={(e) => props.onPasswordChange(e.currentTarget.value)}
                class="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-colpsi-text placeholder:text-slate-400 outline-none transition-colors focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
                placeholder="Introduce tu contraseña para confirmar"
              />
            </div>

            <button
              type="submit"
              disabled={props.saving}
              class="inline-flex items-center gap-2 h-10 px-5 rounded-md bg-colpsi-blue text-white text-sm font-semibold hover:bg-colpsi-blue-light transition-colors disabled:opacity-70 disabled:cursor-not-allowed"
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

          <div class="mt-3">
            <Show
              when={props.message?.text}
              fallback={
                <span class="inline-flex items-center gap-2 text-xs text-colpsi-muted">
                  <span class="w-1.5 h-1.5 rounded-full bg-colpsi-yellow" />
                  Los cambios se guardan para todo el expediente al pulsar el
                  botón.
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
          </div>
        </div>
      </div>
  );
}
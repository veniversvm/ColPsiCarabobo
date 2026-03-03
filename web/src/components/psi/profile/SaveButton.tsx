// web/src/components/psi/profile/SaveButton.tsx
import { Show } from "solid-js";

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
        class="bg-colpsi-yellow text-colpsi-blue px-8 py-4 md:px-12 rounded-full font-black shadow-2xl shadow-yellow-500/30 hover:scale-105 active:scale-95 transition-all disabled:opacity-70 flex items-center gap-3 border-2 border-transparent focus:border-white focus:ring-4 focus:ring-yellow-500/50"
      >
        <Show 
          when={props.saving} 
          fallback={
            <>
              <span class="text-xl">💾</span>
              GUARDAR CAMBIOS
            </>
          }
        >
          <svg class="animate-spin h-5 w-5 text-colpsi-blue" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          GUARDANDO...
        </Show>
      </button>
    </div>
  );
}
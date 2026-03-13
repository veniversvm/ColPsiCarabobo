// web/src/components/admin/psicologos/PsychologistSearchBar.tsx
import { Show } from "solid-js";

interface Props {
  value: string;
  onInput: (e: Event) => void;
  onClear: () => void;
  loading: boolean;
}

export function PsychologistSearchBar(props: Props) {
  return (
    <div class="bg-white p-4 rounded-2xl shadow-sm border border-gray-100 flex items-center gap-3">
      <span class="text-xl ml-2">🔍</span>
      <input
        type="text"
        value={props.value}
        onInput={props.onInput}
        placeholder="Buscar por nombre, apellido, cédula o FPV..."
        class="flex-grow bg-transparent outline-none text-colpsi-text font-medium"
      />
      <Show when={props.loading}>
        <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-colpsi-yellow mr-2 flex-shrink-0" />
      </Show>
      <Show when={props.value}>
        <button
          onClick={props.onClear}
          class="text-gray-400 hover:text-gray-600 font-black text-lg leading-none flex-shrink-0 mr-1"
          title="Limpiar búsqueda"
        >
          ×
        </button>
      </Show>
    </div>
  );
}
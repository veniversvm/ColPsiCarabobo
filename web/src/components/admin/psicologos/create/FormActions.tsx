// web/src/components/admin/psicologos/create/FormActions.tsx
import { Show } from "solid-js";

interface Props {
  saving: boolean;
}

export function FormActions(props: Props) {
  return (
    <button 
      type="submit" 
      disabled={props.saving}
      class="bg-colpsi-blue text-white px-10 py-4 rounded-2xl font-black shadow-2xl hover:scale-105 active:scale-95 transition-all disabled:opacity-70 disabled:pointer-events-none disabled:cursor-not-allowed flex items-center gap-3 border-2 border-white"
    >
      {props.saving 
        ? <><span class="animate-spin">⏳</span> PROCESANDO...</> 
        : "✔️ REGISTRAR EXPEDIENTE"
      }
    </button>
  );
}
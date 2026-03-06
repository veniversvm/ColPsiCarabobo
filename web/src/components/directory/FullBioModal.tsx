import { Show, onCleanup, createEffect, onMount } from "solid-js";
import { isServer } from "solid-js/web";

interface FullBioModalProps {
  isOpen: boolean;
  onClose: () => void;
  content?: string;
  psychologistName: string;
}

export function FullBioModal(props: FullBioModalProps) {
  // Manejo seguro del scroll: Solo se ejecuta en el navegador
  createEffect(() => {
    // Escudo anti-SSR: si estamos en Deno, no hagas nada.
    if (isServer) return;

    if (props.isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = ""; // Usamos string vacío para restaurar el valor original
    }
  });

  // Limpieza segura al desmontar el componente (Solo en cliente)
  onCleanup(() => {
    if (!isServer && typeof document !== "undefined") {
      document.body.style.overflow = "";
    }
  });

  return (
    <Show when={props.isOpen && props.content}>
      <div 
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200"
        onClick={(e) => { if (e.target === e.currentTarget) props.onClose(); }}
      >
        <div class="bg-white rounded-[2rem] shadow-2xl w-full max-w-4xl max-h-[85vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200 border border-gray-100">
          
          {/* Header del Modal */}
          <div class="flex items-center justify-between px-6 py-4 md:px-8 md:py-6 border-b border-gray-100 bg-gray-50/50">
            <div>
              <h3 class="font-black text-colpsi-blue uppercase tracking-widest text-sm">Biografía Detallada</h3>
              {/* <p class="text-xs text-colpsi-muted mt-1 font-bold">Dr(a). {props.psychologistName}</p> */}
            </div>
            <button 
              onClick={props.onClose} 
              class="w-10 h-10 bg-white border border-gray-200 rounded-full text-gray-500 hover:text-colpsi-red hover:bg-red-50 hover:border-colpsi-red transition-all text-xl font-bold flex items-center justify-center shadow-sm"
              title="Cerrar"
            >
              ×
            </button>
          </div>
          
          {/* Contenido (Scrollable) */}
          <div class="overflow-y-auto p-6 md:p-10 bg-white">
            <div 
              class="prose prose-slate max-w-none text-colpsi-text prose-headings:text-colpsi-blue prose-a:text-colpsi-yellow hover:prose-a:text-colpsi-blue transition-colors prose-li:marker:text-colpsi-yellow prose-img:rounded-xl"
              innerHTML={props.content} 
            />
          </div>
          
        </div>
      </div>
    </Show>
  );
}
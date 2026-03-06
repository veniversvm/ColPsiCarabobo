// web/src/components/psi/PostGradeCard.tsx
import { Show } from "solid-js";
import { PostGrade } from "~/types/psi";

export function PostGradeCard(props: { postGrade: PostGrade; onClick: () => void }) {
  // Calculamos cuántas imágenes tiene adjuntas
  const docsCount = () => {
    return[props.postGrade.pic_one_url, props.postGrade.pic_two_url, props.postGrade.pic_three_url]
      .filter(url => url && url.trim() !== "").length;
  };

  return (
    <button 
      onClick={props.onClick}
      class="w-full text-left group flex items-center justify-between p-4 bg-white hover:bg-blue-50/50 rounded-2xl border border-gray-100 hover:border-colpsi-blue/30 transition-all cursor-pointer"
    >
      {/* Lado Izquierdo: Info textual */}
      <div class="flex-grow pr-4">
        <h4 class="font-black text-colpsi-text group-hover:text-colpsi-blue transition-colors text-sm md:text-base leading-tight">
          {props.postGrade.title}
        </h4>
        <p class="text-xs text-colpsi-muted mt-1 font-medium">
          {props.postGrade.university}
        </p>
        
        {/* Indicador sutil de documentos */}
        <Show when={docsCount() > 0}>
          <div class="mt-2 inline-flex items-center gap-1.5 text-[10px] font-bold text-colpsi-blue bg-blue-50 px-2.5 py-1 rounded-md">
            <span>📎</span> {docsCount()} soporte{docsCount() > 1 ? 's' : ''}
          </div>
        </Show>
      </div>

      {/* Lado Derecho: Año y Acción */}
      <div class="flex flex-col items-end shrink-0 gap-2">
        <Show when={props.postGrade.year}>
          <span class="bg-gray-50 text-gray-500 text-xs font-black px-3 py-1 rounded-xl border border-gray-100 group-hover:bg-white transition-colors">
            {props.postGrade.year}
          </span>
        </Show>
        <span class="text-[10px] text-colpsi-yellow font-black uppercase tracking-widest opacity-0 group-hover:opacity-100 transition-opacity translate-x-2 group-hover:translate-x-0 duration-300">
          Ver detalle →
        </span>
      </div>
    </button>
  );
}
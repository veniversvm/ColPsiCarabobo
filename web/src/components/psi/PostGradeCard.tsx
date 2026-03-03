// web/src/components/psi/PostGradeCard.tsx
// Componente para mostrar un postgrado en la lista
import { Show } from "solid-js";
import { PostGrade } from "~/types/psi";

interface PostGradeCardProps {
  postGrade: PostGrade;
  onClick: () => void;
}

export function PostGradeCard(props: PostGradeCardProps) {
  return (
    <div 
      class="cursor-pointer hover:bg-gray-50 p-3 rounded-xl transition-all active:scale-[0.99]"
      onClick={props.onClick}
    >
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-1 md:gap-2">
        <h4 class="font-bold text-gray-900 text-base md:text-lg">
          {props.postGrade.title}
        </h4>
        <span class="text-xs bg-colpsi-yellow/20 text-colpsi-blue px-2 py-1 rounded-full font-medium w-fit">
          {props.postGrade.year}
        </span>
      </div>
      <p class="text-colpsi-blue text-sm md:text-base">{props.postGrade.university}</p>
      
      {/* Badge de certificado */}
      <Show when={props.postGrade.pic_one_url}>
        <div class="mt-1 flex items-center gap-1">
          <span class="text-[10px] text-gray-400">📎 Certificado disponible</span>
        </div>
      </Show>
    </div>
  );
}
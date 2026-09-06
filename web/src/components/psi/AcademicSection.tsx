// web/src/components/psi/AcademicSection.tsx
import { Show, For } from "solid-js";
import { PostGrade, Undergraduate } from "~/types/psi";
import { PostGradeCard } from "./PostGradeCard";
import { UndergraduateCard } from "./UndergraduateCard";

interface AcademicSectionProps {
  undergraduate?: Undergraduate;
  postGrades?: PostGrade[];
  onPostGradeClick: (postGrade: PostGrade) => void;
  onUndergraduateClick?: () => void;
}

export function AcademicSection(props: AcademicSectionProps) {
  const hasUndergraduate = () => props.undergraduate?.university;
  const hasPostGrades = () => props.postGrades && props.postGrades.length > 0;

  return (
    <div class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-colpsi-border">
      <h3 class="text-xs md:text-sm font-black text-colpsi-blue uppercase tracking-widest mb-6 flex items-center gap-2 border-b-2 border-gray-50 pb-4">
        <span class="text-xl">🎓</span> Formación Académica
      </h3>
      
      {/* PREGRADO */}
      <Show when={hasUndergraduate()}>
        <div class="mb-6">
          <UndergraduateCard 
            undergraduate={props.undergraduate!} 
            onClick={props.onUndergraduateClick}
          />
        </div>
      </Show>

      {/* POSTGRADOS (Lista Compacta) */}
      <Show when={hasPostGrades()}>
        <div class="space-y-3 relative">
          <h4 class="text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-3 pl-1">
            Especializaciones y Títulos Adicionales
          </h4>
          
          {/* CONTENEDOR CON SCROLL OPTIMIZADO */}
          {/* Al ser tarjetas más delgadas, 350px es suficiente para ver 4-5 a la vez */}
          <div class="max-h-[350px] overflow-y-auto pr-1 pb-1 space-y-2 
                      scrollbar-thin scrollbar-thumb-colpsi-blue/10 hover:scrollbar-thumb-colpsi-blue/30 transition-colors">
            <For each={props.postGrades}>
              {(pg) => (
                <PostGradeCard 
                  postGrade={pg} 
                  onClick={() => props.onPostGradeClick(pg)} 
                />
              )}
            </For>
          </div>
          
          {/* Sombra inferior indicadora de scroll (Aparece solo si hay muchos) */}
          <Show when={props.postGrades!.length > 4}>
             <div class="absolute bottom-0 left-0 w-full h-8 bg-linear-to-t from-white to-transparent pointer-events-none rounded-b-2xl"></div>
          </Show>
        </div>
      </Show>

      {/* ESTADO VACÍO */}
      <Show when={!hasUndergraduate() && !hasPostGrades()}>
        <div class="bg-colpsi-surface p-8 rounded-3xl text-center border-2 border-dashed border-gray-200">
           <span class="text-3xl grayscale opacity-50 mb-2 block">🎓</span>
           <p class="text-gray-400 text-sm font-bold">Sin información académica pública</p>
        </div>
      </Show>
    </div>
  );
}
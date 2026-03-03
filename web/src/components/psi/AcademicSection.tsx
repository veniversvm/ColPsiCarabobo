// web/src/components/psi/AcademicSection.tsx
// Sección de formación académica completa
import { Show, For } from "solid-js";
import { PostGrade, Undergraduate } from "~/types/psi";
import { PostGradeCard } from "./PostGradeCard";

interface AcademicSectionProps {
  undergraduate?: Undergraduate;
  postGrades?: PostGrade[];
  onPostGradeClick: (postGrade: PostGrade) => void;
}

export function AcademicSection(props: AcademicSectionProps) {
  const hasUndergraduate = () => props.undergraduate?.university;
  const hasPostGrades = () => props.postGrades && props.postGrades.length > 0;

  return (
    <div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100">
      <h3 class="text-xs md:text-sm font-black text-colpsi-blue uppercase tracking-widest mb-4">
        Formación Académica
      </h3>
      
      {/* Pregrado */}
      <Show when={hasUndergraduate()}>
        <div class="mb-4 pb-4 border-b border-gray-100">
          <h4 class="font-bold text-gray-900 text-base md:text-lg">Psicólogo</h4>
          <p class="text-colpsi-blue text-sm md:text-base">{props.undergraduate?.university}</p>
          <Show when={props.undergraduate?.date || props.undergraduate?.mention}>
            <div class="text-xs md:text-sm text-gray-500 mt-1 flex flex-col md:flex-row gap-1 md:gap-4">
              <Show when={props.undergraduate?.date}>
                <span>Egreso: {props.undergraduate?.date}</span>
              </Show>
              <Show when={props.undergraduate?.mention}>
                <span>Mención: {props.undergraduate?.mention}</span>
              </Show>
            </div>
          </Show>
        </div>
      </Show>

      {/* Postgrados */}
      <Show when={hasPostGrades()}>
        <div class="space-y-4">
          <For each={props.postGrades}>
            {(pg) => (
              <PostGradeCard 
                postGrade={pg} 
                onClick={() => props.onPostGradeClick(pg)} 
              />
            )}
          </For>
        </div>
      </Show>

      {/* Sin información */}
      <Show when={!hasUndergraduate() && !hasPostGrades()}>
        <p class="text-gray-400 italic text-sm">Información académica no disponible</p>
      </Show>
    </div>
  );
}
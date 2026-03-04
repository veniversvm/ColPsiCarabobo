// web/src/components/psi/AcademicSection.tsx
// Sección de formación académica completa
import { Show, For } from "solid-js";
import { PostGrade, Undergraduate } from "~/types/psi";
import { PostGradeCard } from "./PostGradeCard";
import { UndergraduateCard } from "./UndergraduateCard";

interface AcademicSectionProps {
  undergraduate?: Undergraduate;
  postGrades?: PostGrade[];
  onPostGradeClick: (postGrade: PostGrade) => void;
  onUndergraduateClick?: () => void; // Opcional
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
        <div class="mb-6 pb-6 border-b border-gray-100 last:border-0 last:mb-0 last:pb-0">
          <UndergraduateCard 
            undergraduate={props.undergraduate!} 
            onClick={props.onUndergraduateClick}
          />
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
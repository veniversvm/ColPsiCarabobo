// web/src/components/directory/ResultsGrid.tsx
import { For, Show } from "solid-js";
import { PsychologistCard } from "./PsychologistCard";
import { DirectoryPsychologist } from "~/types/psi";
import { SearchingIndicator } from "../ui/SearchingIndicator";

interface ResultsGridProps {
  psychologists: DirectoryPsychologist[] | undefined;
  loading: boolean;
}

export function ResultsGrid(props: ResultsGridProps) {
  return (
    <Show 
      when={!props.loading}
      fallback={
        <div class="bg-white rounded-[2.5rem] shadow-premium border border-gray-100">
          <SearchingIndicator />
        </div>
      }
    >
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <For 
          each={props.psychologists} 
          fallback={
            <div class="col-span-full bg-white rounded-[2.5rem] p-20 text-center border-2 border-dashed border-gray-100">
              <span class="text-5xl mb-4 block">🧐</span>
              <h4 class="text-colpsi-blue font-black uppercase">No hay coincidencias</h4>
              <p class="text-colpsi-muted text-sm mt-2">Intenta con otro nombre o especialidad.</p>
            </div>
          }
        >
          {(psychologist) => (
            <PsychologistCard psychologist={psychologist} />
          )}
        </For>
      </div>
    </Show>
  );
}
// web/src/components/directory/ResultsGrid.tsx
import { For, Show } from "solid-js";
import { PsychologistCard } from "./PsychologistCard";
import { DirectoryPsychologist } from "~/types/psi";
import { SearchingIndicator } from "../ui/SearchingIndicator";
import { Animate } from "~/components/ui/Motion";

interface ResultsGridProps {
  psychologists: DirectoryPsychologist[];
  loading: boolean;
  loadingMore: boolean;
  hasMore: boolean;
  total?: number;
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
      <Show when={props.psychologists.length > 0 && props.total !== undefined}>
        <p class="text-s text-gray-400 font-bold mb-4 px-10">
          {props.total} psicólogos encontrados.
        </p>
      </Show>

      {/* Grid */}
      <Animate variant="slide-bottom" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
          {(psychologist) => <PsychologistCard psychologist={psychologist} />}
        </For>
      </Animate>

      {/* Spinner de carga adicional */}
      <Show when={props.loadingMore}>
        <div class="flex justify-center items-center py-10 gap-3 text-gray-400">
          <div class="w-6 h-6 border-2 border-gray-200 border-t-colpsi-blue rounded-full animate-spin" />
          <span class="text-sm font-bold">Cargando más psicólogos...</span>
        </div>
      </Show>

      {/* Mensaje de fin de lista */}
      <Show when={!props.hasMore && props.psychologists.length > 0 && !props.loadingMore}>
        <p class="text-center text-xs text-gray-300 font-bold py-8 uppercase tracking-widest">
          — Has visto todos los resultados —
        </p>
      </Show>
    </Show>
  );
}
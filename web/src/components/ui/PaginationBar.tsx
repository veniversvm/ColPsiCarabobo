// web/src/components/ui/PaginationBar.tsx
import { For, Show } from "solid-js";

const PAGE_SIZE_OPTIONS = [10, 25, 50, 100];

interface Props {
  page: number;
  totalPages: number;
  limit: number;
  total: number;
  onPrev: () => void;
  onNext: () => void;
  onLimitChange: (v: number) => void;
  onPageChange?: (page: number) => void;
  isLoading?: boolean; // Nuevo prop
}

export function PaginationBar(props: Props) {
  return (
    <div class="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 bg-gray-50 border-t border-gray-100">
      {/* Info + selector de entradas */}
      <div class="flex items-center gap-3">
        <span class="text-xs font-bold text-gray-400 uppercase tracking-widest whitespace-nowrap">
          Página {props.page} de {props.totalPages}
          <span class="text-gray-300 mx-2">·</span>
          {props.total} registros
        </span>
        <select
          value={props.limit}
          onChange={(e) => props.onLimitChange(Number(e.currentTarget.value))}
          disabled={props.isLoading}
          class="text-xs font-black text-gray-600 bg-white border border-gray-200 rounded-lg px-2 py-1.5 outline-none focus:border-blue-400 transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <For each={PAGE_SIZE_OPTIONS}>
            {(size) => <option value={size}>{size} por página</option>}
          </For>
        </select>
        
        {/* Indicador de carga inline */}
        <Show when={props.isLoading}>
          <div class="animate-spin rounded-full h-4 w-4 border-2 border-colpsi-yellow border-t-transparent" />
        </Show>
      </div>

      {/* Botones */}
      <div class="flex gap-2">
        <button
          disabled={props.page === 1 || props.isLoading}
          onClick={props.onPrev}
          class="px-4 py-1.5 bg-white border border-gray-200 rounded-lg text-xs font-black text-gray-600 hover:border-blue-400 hover:text-blue-700 disabled:opacity-30 transition-all flex items-center gap-1"
        >
          <Show when={props.isLoading && props.page > 1}>
            <span class="animate-spin inline-block h-3 w-3 border-2 border-gray-400 border-t-transparent rounded-full" />
          </Show>
          ← Anterior
        </button>

        {/* Números de página (cuando hay pocas páginas) */}
        <Show when={props.totalPages <= 7 && props.onPageChange}>
          <For each={Array.from({ length: props.totalPages }, (_, i) => i + 1)}>
            {(n) => (
              <button
                onClick={() => props.onPageChange?.(n)}
                disabled={props.isLoading}
                class={`w-8 h-8 rounded-lg text-xs font-black transition-all border ${
                  n === props.page
                    ? "bg-blue-800 text-white border-blue-800"
                    : "bg-white text-gray-600 border-gray-200 hover:border-blue-400"
                } disabled:opacity-30`}
              >
                {n}
              </button>
            )}
          </For>
        </Show>

        <button
          disabled={props.page === props.totalPages || props.isLoading}
          onClick={props.onNext}
          class="px-4 py-1.5 bg-white border border-gray-200 rounded-lg text-xs font-black text-gray-600 hover:border-blue-400 hover:text-blue-700 disabled:opacity-30 transition-all flex items-center gap-1"
        >
          Siguiente →
          <Show when={props.isLoading && props.page < props.totalPages}>
            <span class="animate-spin inline-block h-3 w-3 border-2 border-gray-400 border-t-transparent rounded-full" />
          </Show>
        </button>
      </div>
    </div>
  );
}
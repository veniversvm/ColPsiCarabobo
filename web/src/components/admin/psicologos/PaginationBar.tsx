// web/src/components/admin/shared/PaginationBar.tsx
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
}

export function PaginationBar(props: Props) {
  return (
    <div class="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 bg-colpsi-surface border-t border-colpsi-border">
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
          class="text-xs font-black text-gray-600 bg-white border border-gray-200 rounded-lg px-2 py-1.5 outline-none focus:border-blue-400 transition-all cursor-pointer"
        >
          <For each={PAGE_SIZE_OPTIONS}>
            {(size) => <option value={size}>{size} por página</option>}
          </For>
        </select>
      </div>

      {/* Botones */}
      <div class="flex gap-2">
        <button
          disabled={props.page === 1}
          onClick={props.onPrev}
          class="px-4 py-1.5 bg-white border border-gray-200 rounded-lg text-xs font-black text-gray-600 hover:border-blue-400 hover:text-blue-700 disabled:opacity-30 transition-all"
        >
          ← Anterior
        </button>

        {/* Números de página (cuando hay pocas páginas) */}
        <Show when={props.totalPages <= 7 && props.onPageChange}>
          <For each={Array.from({ length: props.totalPages }, (_, i) => i + 1)}>
            {(n) => (
              <button
                onClick={() => props.onPageChange?.(n)}
                class={`w-8 h-8 rounded-lg text-xs font-black transition-all border ${
                  n === props.page
                    ? "bg-blue-800 text-white border-blue-800"
                    : "bg-white text-gray-600 border-gray-200 hover:border-blue-400"
                }`}
              >
                {n}
              </button>
            )}
          </For>
        </Show>

        <button
          disabled={props.page === props.totalPages}
          onClick={props.onNext}
          class="px-4 py-1.5 bg-white border border-gray-200 rounded-lg text-xs font-black text-gray-600 hover:border-blue-400 hover:text-blue-700 disabled:opacity-30 transition-all"
        >
          Siguiente →
        </button>
      </div>
    </div>
  );
}
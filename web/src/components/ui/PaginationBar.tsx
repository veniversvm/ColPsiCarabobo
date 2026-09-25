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
    <div class="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-2.5 bg-colpsi-surface border-t border-colpsi-border">
      {/* Info + selector de entradas */}
      <div class="flex items-center gap-3">
        <span class="text-xs font-semibold text-colpsi-muted whitespace-nowrap">
          Página {props.page} de {props.totalPages}
          <span class="text-slate-300 mx-2">·</span>
          {props.total} registros
        </span>
        <select
          value={props.limit}
          onChange={(e) => props.onLimitChange(Number(e.currentTarget.value))}
          disabled={props.isLoading}
          class="h-8 rounded-md border border-slate-300 bg-white px-2 text-xs font-medium text-colpsi-text outline-none focus:border-colpsi-blue transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
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
          class="h-8 px-3 bg-white border border-colpsi-border rounded-md text-xs font-medium text-colpsi-text hover:bg-colpsi-bg hover:text-colpsi-blue disabled:opacity-40 transition-all flex items-center gap-1"
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
                class={`w-8 h-8 rounded-md text-xs font-medium transition-all border ${
                  n === props.page
                    ? "bg-colpsi-blue text-white border-colpsi-blue"
                    : "bg-white text-colpsi-muted border-colpsi-border hover:bg-colpsi-bg hover:text-colpsi-blue"
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
          class="h-8 px-3 bg-white border border-colpsi-border rounded-md text-xs font-medium text-colpsi-text hover:bg-colpsi-bg hover:text-colpsi-blue disabled:opacity-40 transition-all flex items-center gap-1"
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
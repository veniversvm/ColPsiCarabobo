// web/src/components/admin/ui/KpiStrip.tsx
// Bloque de métricas integrado: UN contenedor dividido por líneas finas,
// en lugar de tarjetas sueltas flotando con sombras.
import { For, Show } from "solid-js";

export interface KpiCellData {
  label: string;
  value: string | number;
  /** Línea secundaria (texto ya formateado) */
  sub?: string;
  /** Clase de color del punto semántico (ej: "bg-emerald-500") */
  dot?: string;
}

interface KpiStripProps {
  cells: KpiCellData[];
  /** Columnas en pantallas lg (2/3/4) */
  columns?: 2 | 3 | 4;
}

const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE");

const gridCols: Record<number, string> = {
  2: "lg:grid-cols-2",
  3: "lg:grid-cols-3",
  4: "lg:grid-cols-4",
};

export function KpiStrip(props: KpiStripProps) {
  const cols = () => props.columns ?? 4;
  return (
    <div class={`grid grid-cols-2 border-t border-l border-colpsi-border ${gridCols[cols()] ?? "lg:grid-cols-4"}`}>
      <For each={props.cells}>
        {(c) => (
          <div class="border-b border-r border-colpsi-border p-4">
            <div class="flex items-center gap-1.5">
              <Show when={c.dot}>
                <span class={`w-1.5 h-1.5 rounded-full ${c.dot}`} />
              </Show>
              <p class="text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted truncate">{c.label}</p>
            </div>
            <p class="mt-1.5 text-2xl font-semibold tabular-nums text-colpsi-text leading-none">
              {typeof c.value === "number" ? fmt(c.value) : c.value}
            </p>
            <Show when={c.sub}>
              <p class="mt-1 text-xs text-colpsi-muted truncate">{c.sub}</p>
            </Show>
          </div>
        )}
      </For>
    </div>
  );
}
// web/src/components/admin/dashboard/RankingList.tsx
// Ranking con barras. Contenido puro: el Panel padre (con divisiones) lo envuelve.
import { For, Show } from "solid-js";

interface TopItem { value: string; count: number; name: string }

interface RankingListProps {
  title: string;
  items: TopItem[];
  /** Clase Tailwind de la barra de progreso (default: azul institucional) */
  accent?: string;
}

const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE");

export function RankingList(props: RankingListProps) {
  const max = () => Math.max(...(props.items ?? []).map(i => i.count), 1);
  const bar = props.accent ?? "bg-colpsi-blue/50";
  return (
    <div class="p-4">
      <div class="flex items-center justify-between gap-2 mb-3">
        <p class="text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted truncate">{props.title}</p>
        <span class="text-[11px] text-colpsi-muted shrink-0">30 días</span>
      </div>
      <Show
        when={props.items?.length > 0}
        fallback={<p class="text-xs text-slate-400 italic text-center py-4">Sin datos aún</p>}
      >
        <div class="space-y-2.5">
          <For each={props.items?.slice(0, 8)}>
            {(item, i) => (
              <div class="flex items-center gap-3">
                <span class="text-[11px] font-medium text-slate-400 w-4 text-right shrink-0">{i() + 1}</span>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between gap-2 mb-0.5">
                    <span class="text-sm font-medium text-colpsi-text truncate">
                      {item?.name || item.value || "—"}
                    </span>
                    <span class="text-sm font-semibold text-colpsi-text tabular-nums">
                      {fmt(item.count)}
                    </span>
                  </div>
                  <div class="h-1.5 bg-slate-100 rounded-full overflow-hidden">
                    <div
                      class={`h-full rounded-full transition-all duration-700 ${bar}`}
                      style={{ width: `${(item.count / max()) * 100}%` }}
                    />
                  </div>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </div>
  );
}
// web/src/components/admin/dashboard/RankingList.tsx
import { For, Show } from "solid-js";

interface TopItem { value: string; count: number; name: string }

interface RankingListProps {
  title: string;
  icon: string;
  items: TopItem[];
}

const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE");

export function RankingList(props: RankingListProps) {
  const max = () => Math.max(...(props.items ?? []).map(i => i.count), 1);
  return (
    <div class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
      <p class="text-xs font-black text-gray-400 uppercase tracking-widest mb-4 flex items-center gap-2">
        <span>{props.icon}</span>
        {props.title}
        <span class="ml-auto text-gray-300">últimos 30 días</span>
      </p>
      <Show
        when={props.items?.length > 0}
        fallback={<p class="text-xs text-gray-300 italic text-center py-4">Sin datos aún</p>}
      >
        <div class="space-y-2.5">
          <For each={props.items?.slice(0, 8)}>
            {(item, i) => (
              <div class="flex items-center gap-3">
                <span class="text-[10px] font-black text-gray-300 w-4 text-right">{i() + 1}</span>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between mb-0.5">
                    <span class="text-xs font-bold text-gray-700 truncate max-w-[160px]">
                      {item?.name || item.value || "—"}
                    </span>
                    <span class="text-xs font-black text-colpsi-blue tabular-nums ml-2">
                      {fmt(item.count)}
                    </span>
                  </div>
                  <div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div
                      class="h-full bg-colpsi-blue/60 rounded-full transition-all duration-700"
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
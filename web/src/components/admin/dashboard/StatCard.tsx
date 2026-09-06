// web/src/components/admin/dashboard/StatCard.tsx
import { For, Show } from "solid-js";

interface StatCardProps {
  icon: string;
  label: string;
  value: number;
  sub?: { label: string; value: number }[];
  accent?: string;
  pulse?: boolean;
}

const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE");

export function StatCard(props: StatCardProps) {
  const accent = () => props.accent ?? "border-colpsi-yellow";
  return (
    <div class={`bg-white rounded-2xl p-5 shadow-sm border border-colpsi-border border-l-4 ${accent()} flex flex-col gap-3`}>
      <div class="flex items-center justify-between">
        <span class="text-2xl">{props.icon}</span>
        <Show when={props.pulse}>
          <span class="flex items-center gap-1.5 text-[10px] font-black text-green-600 uppercase tracking-widest">
            <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse inline-block" />
            En vivo
          </span>
        </Show>
      </div>
      <div>
        <p class="text-3xl font-black text-colpsi-blue tabular-nums">{fmt(props.value)}</p>
        <p class="text-xs font-bold text-gray-400 uppercase tracking-widest mt-0.5">{props.label}</p>
      </div>
      <Show when={props.sub && props.sub.length > 0}>
        <div class="flex flex-wrap gap-3 border-t border-colpsi-border pt-3">
          <For each={props.sub}>
            {(s) => (
              <div>
                <p class="text-sm font-black text-gray-700 tabular-nums">{fmt(s.value)}</p>
                <p class="text-[10px] text-gray-400 uppercase tracking-widest">{s.label}</p>
              </div>
            )}
          </For>
        </div>
      </Show>
    </div>
  );
}
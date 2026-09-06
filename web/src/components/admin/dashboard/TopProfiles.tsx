// web/src/components/admin/dashboard/TopProfiles.tsx
import { For, Show } from "solid-js";

interface TopProfile {
  psi_id:     string;
  first_name: string;
  last_name:  string;
  fpv:        number;
  count:      number;
}

interface TopProfilesProps {
  profiles: TopProfile[];
}

const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE");

export function TopProfiles(props: TopProfilesProps) {
  return (
    <Show when={props.profiles?.length > 0}>
      <div>
        <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
          Perfiles más Visitados (últimos 30 días)
        </h2>
        <div class="bg-white rounded-2xl p-5 shadow-sm border border-colpsi-border">
          <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 gap-3">
            <For each={props.profiles?.slice(0, 10)}>
              {(p, i) => (
                <div class="bg-colpsi-surface rounded-xl p-3 border border-colpsi-border flex flex-col gap-1">
                  <span class="text-[9px] font-black text-gray-300 uppercase">#{i() + 1}</span>
                  <span class="text-lg font-black text-colpsi-blue tabular-nums">{fmt(p.count)}</span>
                  <span class="text-xs font-bold text-gray-700 truncate">
                    {p.first_name} {p.last_name}
                  </span>
                  <span class="text-[9px] text-gray-400">FPV {p.fpv} · {fmt(p.count)} visitas</span>
                </div>
              )}
            </For>
          </div>
        </div>
      </div>
    </Show>
  );
}
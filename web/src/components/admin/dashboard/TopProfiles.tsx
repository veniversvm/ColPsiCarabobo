// web/src/components/admin/dashboard/TopProfiles.tsx
// Perfiles más visitados: grilla con líneas divisorias (no tarjetas sueltas).
import { For, Show } from "solid-js";
import { Panel } from "~/components/admin/ui/Panel";

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
      <section class="space-y-2">
        <h2 class="text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted">
          Perfiles más visitados — últimos 30 días
        </h2>
        <Panel flush>
          <div class="grid grid-cols-2 md:grid-cols-5 border-t border-l border-colpsi-border">
            <For each={props.profiles?.slice(0, 10)}>
              {(p, i) => (
                <div class="border-b border-r border-colpsi-border p-3 bg-white flex flex-col gap-0.5">
                  <span class={
                    i() === 0
                      ? "text-[10px] font-semibold text-colpsi-yellow-dark"
                      : i() < 3
                        ? "text-[10px] font-medium text-colpsi-blue/80"
                        : "text-[10px] font-medium text-slate-400"
                  }>
                    #{i() + 1}
                  </span>
                  <span class="text-lg font-semibold text-colpsi-blue tabular-nums">{fmt(p.count)}</span>
                  <span class="text-sm font-medium text-colpsi-text truncate">
                    {p.first_name} {p.last_name}
                  </span>
                  <span class="text-[11px] text-colpsi-muted">FPV {p.fpv}</span>
                </div>
              )}
            </For>
          </div>
        </Panel>
      </section>
    </Show>
  );
}
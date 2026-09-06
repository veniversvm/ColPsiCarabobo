// web/src/components/admin/psicologos/edit/AuditBlock.tsx

import { For, Show } from "solid-js";
import type { PsiProfile } from "./types";

interface Props {
  profile: PsiProfile | undefined;
}

export function AuditBlock(props: Props) {
  return (
    <div>
      {/* Postgrados */}
      <div class="bg-white p-5 rounded-2xl border border-colpsi-border mb-4">
        <h3 class="text-xs font-bold text-blue-800 mb-3">
          Postgrados ({props.profile?.post_grades?.length || 0})
        </h3>
        <ul class="text-sm text-gray-600 space-y-2">
          <Show when={!props.profile?.post_grades?.length}>
            <li class="italic text-gray-400">Ninguno registrado</li>
          </Show>
          <For each={props.profile?.post_grades}>
            {(pg: any) => (
              <li class="flex flex-col">
                <span class="font-semibold text-gray-800">{pg.post_grade_title}</span>
                <span class="text-xs text-gray-400">{pg.post_grade_graduation_year}</span>
              </li>
            )}
          </For>
        </ul>
      </div>

      {/* Metadatos */}
      <div class="bg-white rounded-2xl border border-colpsi-border p-4 grid grid-cols-2 md:grid-cols-4 gap-4 text-[11px] text-gray-500">
        <div>
          <p class="font-black uppercase tracking-widest text-[9px] text-gray-400 mb-0.5">ID Interno</p>
          <p class="font-mono truncate">{props.profile?.id}</p>
        </div>
        <div>
          <p class="font-black uppercase tracking-widest text-[9px] text-gray-400 mb-0.5">Creado por</p>
          <p>{props.profile?.create_by}</p>
        </div>
        <div>
          <p class="font-black uppercase tracking-widest text-[9px] text-gray-400 mb-0.5">Último update</p>
          <p>{props.profile?.update_by}</p>
        </div>
        <div>
          <p class="font-black uppercase tracking-widest text-[9px] text-gray-400 mb-0.5">Actualizado</p>
          <p>
            {props.profile?.updated_at
              ? new Date(props.profile.updated_at).toLocaleDateString("es-VE")
              : "—"}
          </p>
        </div>
      </div>
    </div>
  );
}
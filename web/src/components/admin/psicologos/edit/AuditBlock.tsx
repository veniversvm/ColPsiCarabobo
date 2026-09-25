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
      <div class="bg-white p-5 rounded-lg border border-colpsi-border mb-4">
        <h3 class="text-xs font-bold text-colpsi-blue mb-3">
          Postgrados ({props.profile?.post_grades?.length || 0})
        </h3>
        <ul class="text-sm text-gray-600 space-y-2">
          <Show when={!props.profile?.post_grades?.length}>
            <li class="italic text-colpsi-muted">Ninguno registrado</li>
          </Show>
          <For each={props.profile?.post_grades}>
            {(pg: any) => (
              <li class="flex flex-col">
                <span class="font-semibold text-colpsi-text">{pg.post_grade_title}</span>
                <span class="text-xs text-colpsi-muted">{pg.post_grade_graduation_year}</span>
              </li>
            )}
          </For>
        </ul>
      </div>

      {/* Metadatos */}
      <div class="bg-white rounded-lg border border-colpsi-border p-4 grid grid-cols-2 md:grid-cols-4 gap-4 text-[11px] text-colpsi-muted">
        <div>
          <p class="font-semibold uppercase tracking-wide text-[9px] text-colpsi-muted mb-0.5">ID Interno</p>
          <p class="font-mono truncate">{props.profile?.id}</p>
        </div>
        <div>
          <p class="font-semibold uppercase tracking-wide text-[9px] text-colpsi-muted mb-0.5">Creado por</p>
          <p>{props.profile?.create_by}</p>
        </div>
        <div>
          <p class="font-semibold uppercase tracking-wide text-[9px] text-colpsi-muted mb-0.5">Último update</p>
          <p>{props.profile?.update_by}</p>
        </div>
        <div>
          <p class="font-semibold uppercase tracking-wide text-[9px] text-colpsi-muted mb-0.5">Actualizado</p>
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
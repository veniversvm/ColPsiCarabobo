// web/src/components/admin/psicologos/edit/EditPageHeader.tsx

import { createResource, Show } from "solid-js";
import { useNavigate, A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import type { PsiProfile } from "./types";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  profile: PsiProfile | undefined;
}

interface AdminMePerms {
  sudo: boolean;
  can_view_logs?: boolean;
}

export function EditPageHeader(props: Props) {
  const navigate = useNavigate();

  // Solo cosmético: el backend sigue siendo la barrera real (404 enmascarado).
  const [me] = createResource<AdminMePerms | null>(async () => {
    try {
      return await apiGet<AdminMePerms>("/admin/me");
    } catch {
      return null;
    }
  });
  const canViewLogs = () => me()?.sudo || me()?.can_view_logs || false;

  return (
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 mb-6 pb-4 border-b border-colpsi-border">
      <div class="flex items-center gap-3">
        <button
          onClick={() => navigate(-1)}
          class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors shrink-0"
          title="Volver"
        >
          <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
        </button>
        <div>
          <h1 class="text-lg font-semibold text-colpsi-text">Expediente de Colegiado</h1>
          <p class="text-sm text-colpsi-muted font-medium mt-0.5">
            FPV: {props.profile?.fpv || "—"} · USUARIO: {props.profile?.username || "—"}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2 flex-wrap">
        <Show when={canViewLogs() && props.profile?.id}>
          <A
            href={`/admin/auditoria?psi_id=${props.profile!.id}`}
            class="inline-flex items-center gap-1.5 h-8 px-3 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:border-colpsi-blue/40 hover:text-colpsi-blue transition-colors text-xs font-semibold"
            title="Historial de cambios de este colegiado en la bitácora"
          >
            <Icon name="shield" class="w-3.5 h-3.5" />
            Ver bitácora
          </A>
        </Show>
        <Show when={props.profile?.solvent}>
          <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border border-emerald-200 bg-emerald-50 text-emerald-700 text-[11px] font-semibold uppercase tracking-wide"><Icon name="check" class="w-3 h-3" /> Solvente</span>
        </Show>
        <Show when={!props.profile?.is_active}>
          <span class="inline-flex items-center px-2 py-0.5 rounded-md border border-red-200 bg-red-50 text-colpsi-red text-[11px] font-semibold uppercase tracking-wide">Suspendido</span>
        </Show>
        <Show when={props.profile?.proof_of_life}>
          <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border border-blue-200 bg-blue-50 text-colpsi-blue text-[11px] font-semibold uppercase tracking-wide"><Icon name="checkCircle" class="w-3 h-3" /> Fe de Vida</span>
        </Show>
      </div>
    </div>
  );
}
// web/src/components/admin/psicologos/edit/EditPageHeader.tsx

import { createResource, Show } from "solid-js";
import { useNavigate, A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import type { PsiProfile } from "./types";

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
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-colpsi-border">
      <div class="flex items-center gap-4">
        <button
          onClick={() => navigate(-1)}
          class="w-10 h-10 bg-colpsi-surface hover:bg-gray-100 text-gray-600 rounded-full font-bold flex items-center justify-center transition-colors"
        >
          ←
        </button>
        <div>
          <h1 class="text-2xl font-black text-blue-800 uppercase">Expediente de Colegiado</h1>
          <p class="text-gray-500 text-sm font-bold tracking-widest mt-0.5">
            FPV: {props.profile?.fpv || "—"} · USUARIO: {props.profile?.username || "—"}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3 flex-wrap">
        <Show when={canViewLogs() && props.profile?.id}>
          <A
            href={`/admin/auditoria?psi_id=${props.profile!.id}`}
            class="bg-slate-100 hover:bg-slate-200 text-slate-700 px-3 py-1.5 rounded-lg text-xs font-black uppercase transition-colors"
            title="Historial de cambios de este colegiado en la bitácora"
          >
            🧾 Ver bitácora
          </A>
        </Show>
        <Show when={props.profile?.solvent}>
          <span class="bg-green-100 text-green-700 px-3 py-1.5 rounded-lg text-xs font-black uppercase">Solvente</span>
        </Show>
        <Show when={!props.profile?.is_active}>
          <span class="bg-red-100 text-red-700 px-3 py-1.5 rounded-lg text-xs font-black uppercase">Suspendido</span>
        </Show>
        <Show when={props.profile?.proof_of_life}>
          <span class="bg-blue-100 text-blue-700 px-3 py-1.5 rounded-lg text-xs font-black uppercase">Fe de Vida ✓</span>
        </Show>
      </div>
    </div>
  );
}
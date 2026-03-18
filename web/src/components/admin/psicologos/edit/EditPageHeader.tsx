// web/src/components/admin/psicologos/edit/EditPageHeader.tsx

import { Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import type { PsiProfile } from "./types";

interface Props {
  profile: PsiProfile | undefined;
}

export function EditPageHeader(props: Props) {
  const navigate = useNavigate();

  return (
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
      <div class="flex items-center gap-4">
        <button
          onClick={() => navigate(-1)}
          class="w-10 h-10 bg-gray-50 hover:bg-gray-100 text-gray-600 rounded-full font-bold flex items-center justify-center transition-colors"
        >
          ←
        </button>
        <Show when={props.profile?.profile_picture_url}>
          <img
            src={`http://localhost:9000/colpsi-bucket/${props.profile?.profile_picture_url}`}
            class="w-14 h-14 rounded-2xl object-cover border-2 border-gray-100 shadow"
            alt="Foto de perfil"
          />
        </Show>
        <div>
          <h1 class="text-2xl font-black text-blue-800 uppercase">Expediente de Colegiado</h1>
          <p class="text-gray-500 text-sm font-bold tracking-widest mt-0.5">
            FPV: {props.profile?.fpv || "—"} · USUARIO: {props.profile?.username || "—"}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3 flex-wrap">
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
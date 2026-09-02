// web/src/components/admin/psicologos/PsychologistTableRow.tsx
import { A } from "@solidjs/router";
import { Show } from "solid-js";
import { PsiAdminListItem } from "~/types/admin";

interface Props {
  psi: PsiAdminListItem;
}

export function PsychologistTableRow(props: Props) {
  const psi = () => props.psi;

  return (
    <tr class="hover:bg-blue-50/30 transition-colors group">
      <td class="px-6 py-4 whitespace-nowrap">
        <span class="text-sm font-black text-gray-800 font-mono">
          {psi().control_number || "—"}
        </span>
      </td>
      <td class="px-6 py-4">
        <A href={`/admin/psicologos/${psi().id}/detalle`} class="flex items-center gap-3 group/link">
          <div class="w-10 h-10 rounded-xl bg-colpsi-blue text-white flex items-center justify-center font-bold shadow-sm group-hover/link:bg-colpsi-yellow group-hover/link:text-colpsi-blue transition-colors flex-shrink-0">
            {psi().first_name.charAt(0)}{psi().last_name.charAt(0)}
          </div>
          <div class="min-w-0">
            <p class="font-bold text-colpsi-blue group-hover/link:underline truncate">
              {psi().first_name} {psi().last_name}
            </p>
            <p class="text-xs text-gray-500 truncate">{psi().email}</p>
          </div>
        </A>
      </td>
      <td class="px-6 py-4">
        <p class="text-sm font-bold text-gray-700">FPV: {psi().fpv}</p>
        <p class="text-xs text-gray-500">CI: {psi().ci}</p>
        <Show when={psi().age && psi().age! > 0}>
          <p class="text-xs text-gray-500 mt-0.5">Edad: {psi().age} años</p>
        </Show>
      </td>
      <td class="px-6 py-4">
        <Show
          when={psi().solvent}
          fallback={
            <span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-red-50 text-red-700 border border-red-100">
              Deudor
            </span>
          }
        >
          <span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-green-50 text-green-700 border border-green-100">
            Solvente
          </span>
        </Show>
      </td>
      <td class="px-6 py-4">
        <Show
          when={psi().is_active}
          fallback={
            <span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-gray-100 text-gray-600 border border-gray-200">
              Inactivo
            </span>
          }
        >
          <span class="px-2 py-1 rounded-md text-[10px] font-black uppercase bg-blue-50 text-colpsi-blue border border-blue-100">
            Activo
          </span>
        </Show>
      </td>
      <td class="px-6 py-4 text-right">
        <A
          href={`/admin/psicologos/${psi().id}/detalle`}
          class="inline-flex items-center gap-2 px-3 py-1.5 bg-blue-50 text-colpsi-blue hover:bg-colpsi-yellow transition-colors rounded-lg text-xs font-bold"
        >
          Detalle ✏️
        </A>
      </td>
    </tr>
  );
}
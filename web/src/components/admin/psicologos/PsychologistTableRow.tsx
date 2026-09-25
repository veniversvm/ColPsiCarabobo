// web/src/components/admin/psicologos/PsychologistTableRow.tsx
import { A } from "@solidjs/router";
import { Show } from "solid-js";
import { PsiAdminListItem } from "~/types/admin";
import { Badge } from "~/components/admin/ui/Badge";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  psi: PsiAdminListItem;
}

export function PsychologistTableRow(props: Props) {
  const psi = () => props.psi;

  return (
    <tr class="hover:bg-colpsi-bg/60 transition-colors group">
      <td class="td-cell font-mono text-sm font-medium text-colpsi-text whitespace-nowrap">
        {psi().control_number || "—"}
      </td>
      <td class="td-cell min-w-[220px]">
        <A href={`/admin/psicologos/${psi().id}/detalle`} class="flex items-center gap-2.5 group/link min-w-0">
          <div class="w-8 h-8 rounded-md bg-colpsi-blue/10 text-colpsi-blue flex items-center justify-center font-bold text-xs shrink-0">
            {psi().first_name.charAt(0)}{psi().last_name.charAt(0)}
          </div>
          <div class="min-w-0">
            <p class="font-medium text-colpsi-text group-hover/link:text-colpsi-blue group-hover/link:underline truncate">
              {psi().first_name} {psi().last_name}
            </p>
            <p class="text-xs text-colpsi-muted truncate">{psi().email}</p>
          </div>
        </A>
      </td>
      <td class="td-cell whitespace-nowrap">
        <p class="text-sm font-medium text-colpsi-text">FPV: {psi().fpv}</p>
        <p class="text-xs text-colpsi-muted">CI: {psi().ci}</p>
        <Show when={psi().age && psi().age! > 0}>
          <p class="text-xs text-colpsi-muted mt-0.5">Edad: {psi().age} años</p>
        </Show>
      </td>
      <td class="td-cell">
        <Badge tone={psi().solvent ? "success" : "danger"}>
          {psi().solvent ? "Solvente" : "Deudor"}
        </Badge>
      </td>
      <td class="td-cell">
        <Badge tone={psi().is_active ? "info" : "neutral"}>
          {psi().is_active ? "Activo" : "Inactivo"}
        </Badge>
      </td>
      <td class="td-cell text-right whitespace-nowrap">
        <A
          href={`/admin/psicologos/${psi().id}/detalle`}
          class="inline-flex items-center gap-1.5 h-7 px-2 rounded-md text-xs font-medium text-colpsi-muted opacity-0 group-hover:opacity-100 focus:opacity-100 hover:text-colpsi-blue hover:bg-white transition-all"
        >
          <Icon name="pencil" class="w-3.5 h-3.5" />
          Detalle
        </A>
      </td>
    </tr>
  );
}
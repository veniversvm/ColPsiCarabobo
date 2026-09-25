// web/src/components/admin/psicologos/edit/EditAlert.tsx

import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  message: { type: "success" | "error"; text: string } | null;
}

export function EditAlert(props: Props) {
  return (
    <Show when={props.message}>
      <div
        class={`mb-4 p-3 rounded-md text-sm font-medium flex items-start gap-2.5 animate-in slide-in-from-top-4 ${
          props.message?.type === "success"
            ? "bg-emerald-50 text-emerald-700 border border-emerald-200"
            : "bg-red-50 text-colpsi-red border border-red-200"
        }`}
      >
        <Icon
          name={props.message?.type === "success" ? "checkCircle" : "alertTriangle"}
          class={`w-4 h-4 shrink-0 mt-0.5 ${props.message?.type === "success" ? "text-emerald-600" : "text-colpsi-red"}`}
        />
        <div>
          <p class="font-semibold uppercase text-[11px] tracking-wide">
            {props.message?.type === "success" ? "Operación Exitosa" : "Alerta del Sistema"}
          </p>
          <p class="mt-0.5">{props.message?.text}</p>
        </div>
      </div>
    </Show>
  );
}
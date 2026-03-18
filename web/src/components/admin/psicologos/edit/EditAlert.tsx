// web/src/components/admin/psicologos/edit/EditAlert.tsx

import { Show } from "solid-js";

interface Props {
  message: { type: "success" | "error"; text: string } | null;
}

export function EditAlert(props: Props) {
  return (
    <Show when={props.message}>
      <div
        class={`mb-8 p-5 rounded-2xl font-bold text-sm shadow-sm border-l-4 flex items-start gap-3 animate-in slide-in-from-top-4 ${
          props.message?.type === "success"
            ? "bg-green-50 text-green-800 border-green-500"
            : "bg-red-50 text-red-800 border-red-500"
        }`}
      >
        <span class="text-2xl">{props.message?.type === "success" ? "✅" : "⚠️"}</span>
        <div>
          <p class="font-black uppercase text-xs tracking-wide">
            {props.message?.type === "success" ? "Operación Exitosa" : "Alerta del Sistema"}
          </p>
          <p class="mt-1 font-medium">{props.message?.text}</p>
        </div>
      </div>
    </Show>
  );
}
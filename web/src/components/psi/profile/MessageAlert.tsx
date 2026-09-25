// web/src/components/psi/profile/MessageAlert.tsx
import { Show } from "solid-js";

interface MessageAlertProps {
  type: "success" | "error" | null;
  text: string | null;
}

export function MessageAlert(props: MessageAlertProps) {
  return (
    <Show when={props.type && props.text}>
      <div class={`px-3.5 py-2.5 rounded-md border text-sm font-medium animate-in fade-in slide-in-from-top-2 ${
        props.type === 'success'
          ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
          : 'bg-red-50 text-red-700 border-red-200'
      }`}>
        {props.text}
      </div>
    </Show>
  );
}
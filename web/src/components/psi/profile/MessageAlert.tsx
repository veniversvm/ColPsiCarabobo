// web/src/components/psi/profile/MessageAlert.tsx
import { Show } from "solid-js";
import { Animate } from "~/components/ui/Motion";

interface MessageAlertProps {
  type: "success" | "error" | null;
  text: string | null;
}

export function MessageAlert(props: MessageAlertProps) {
  return (
    <Show when={props.type && props.text}>
      <Animate variant="slide-top" class={`p-4 rounded-2xl font-bold text-sm shadow-sm ${
        props.type === 'success'
          ? 'bg-green-50 text-green-700 border border-green-200'
          : 'bg-red-50 text-red-700 border border-red-200'
      }`}>
        {props.text}
      </Animate>
    </Show>
  );
}
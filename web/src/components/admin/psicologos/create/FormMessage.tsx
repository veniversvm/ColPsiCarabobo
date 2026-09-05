// web/src/components/admin/psicologos/create/FormMessage.tsx
import { Show } from "solid-js";

interface Props {
  type?: "success" | "error" | null;  // Acepta undefined también
  text: string;
  details?: any;
}

export function FormMessage(props: Props) {
  return (
    <Show when={props.type}>
      <div class={`p-5 rounded-2xl shadow-md border-l-4 animate-in slide-in-from-top-4 ${
        props.type === 'success' ? 'bg-green-50 border-green-500' : 'bg-red-50 border-red-500'
      }`}>
        <div class="flex items-start gap-3">
          <span class="text-2xl">{props.type === 'success' ? '✅' : '⚠️'}</span>
          <div>
            <p class={`font-black uppercase tracking-wide text-xs ${
              props.type === 'success' ? 'text-green-800' : 'text-red-800'
            }`}>
              {props.type === 'success' ? 'Operación Exitosa' : 'Alerta del Sistema'}
            </p>
            <p class={`text-sm font-medium mt-1 ${
              props.type === 'success' ? 'text-green-700' : 'text-red-700'
            }`}>
              {props.text}
            </p>
            <Show when={props.details?.error}>
              <pre class="mt-3 p-3 bg-red-100/50 rounded-lg text-xs font-mono text-red-900 overflow-x-auto">
                {JSON.stringify(props.details, null, 2)}
              </pre>
            </Show>
          </div>
        </div>
      </div>
    </Show>
  );
}
// web/src/components/admin/psicologos/create/FormMessage.tsx
import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  type?: "success" | "error" | null;  // Acepta undefined también
  text: string;
  details?: any;
}

export function FormMessage(props: Props) {
  return (
    <Show when={props.type}>
      <div class={`p-4 rounded-md border animate-in slide-in-from-top-4 ${
        props.type === 'success' ? 'bg-emerald-50 border-emerald-200' : 'bg-red-50 border-red-200'
      }`}>
        <div class="flex items-start gap-3">
          <span class={props.type === 'success' ? 'text-emerald-600' : 'text-red-600'}>
            <Icon name={props.type === 'success' ? 'checkCircle' : 'alertTriangle'} class="w-5 h-5" />
          </span>
          <div class="min-w-0">
            <p class={`text-xs font-semibold uppercase tracking-wide ${
              props.type === 'success' ? 'text-emerald-800' : 'text-red-800'
            }`}>
              {props.type === 'success' ? 'Operación Exitosa' : 'Alerta del Sistema'}
            </p>
            <p class={`text-sm font-medium mt-1 ${
              props.type === 'success' ? 'text-emerald-700' : 'text-red-700'
            }`}>
              {props.text}
            </p>
            <Show when={props.details?.error}>
              <pre class="mt-3 p-3 bg-red-100/50 rounded-md text-xs font-mono text-red-900 overflow-x-auto">
                {JSON.stringify(props.details, null, 2)}
              </pre>
            </Show>
          </div>
        </div>
      </div>
    </Show>
  );
}
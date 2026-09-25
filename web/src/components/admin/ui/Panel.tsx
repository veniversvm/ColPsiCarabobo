// web/src/components/admin/ui/Panel.tsx
// Panel base del admin: borde fino, sin sombra, esquinas ligeras (rounded-lg).
// flush=true quita el padding interno para tablas/grillas con divisiones.
import { JSX, Show } from "solid-js";

interface PanelProps {
  children: JSX.Element;
  /** Quita el padding interno (para tablas/grillas con divisiones internas) */
  flush?: boolean;
  /** Título opcional en la cabecera del panel */
  title?: string;
  /** Acciones alineadas a la derecha de la cabecera */
  actions?: JSX.Element;
  class?: string;
}

export function Panel(props: PanelProps) {
  return (
    <section class={`border border-colpsi-border rounded-lg bg-white ${props.flush ? "" : "p-4"} ${props.class ?? ""}`}>
      <Show when={props.title || props.actions}>
        <div class={`flex items-center justify-between gap-3 ${props.flush ? "px-4 pt-4 pb-3" : "pb-3"}`}>
          <Show when={props.title}>
            <h2 class="text-sm font-semibold text-colpsi-text">{props.title}</h2>
          </Show>
          <Show when={props.actions}>
            <div class="flex items-center gap-2">{props.actions}</div>
          </Show>
        </div>
      </Show>
      {props.children}
    </section>
  );
}
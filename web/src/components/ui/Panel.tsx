// web/src/components/ui/Panel.tsx

import { Show, createSignal } from "solid-js";

// Panel: un solo contenedor continuo que agrupa secciones colapsables.
// Cada PanelSection vive dentro de él, separada por un divisor,
// sin su propio fondo/borde/sombra para preservar el flujo visual de un solo panel.
export function Panel(props: { children: any; class?: string }) {
  return (
    <div
      class={`bg-white rounded-lg border border-colpsi-border overflow-hidden divide-y divide-colpsi-border ${
        props.class ?? ""
      }`}
    >
      {props.children}
    </div>
  );
}

// PanelSection: cabecera colapsable/expandible dentro de un Panel.
// Renderiza solo su título + chevron, y el cuerpo queda oculto salvo que
// esté abierto. No aporta fondo/borde propios (los provee el Panel padre).
export function PanelSection(props: {
  title: string;
  subtitle?: string;
  accent?: string;
  defaultOpen?: boolean;
  children: any;
}) {
  // `accent` se conserva por compatibilidad de llamadas, pero ya no pinta
  // ningún acento lateral: el estilo plano usa solo tipografía.
  const [open, setOpen] = createSignal(props.defaultOpen ?? false);
  return (
    <section>
      <button
        type="button"
        onClick={() => setOpen(!open())}
        aria-expanded={open()}
        class="w-full flex items-center justify-between gap-4 px-5 py-4 text-left hover:bg-colpsi-bg/70 transition-colors group"
      >
        <div class="min-w-0">
          <h2 class="text-sm font-semibold text-colpsi-text">
            {props.title}
          </h2>
          <Show when={props.subtitle}>
            <p class="text-xs text-colpsi-muted mt-0.5 font-normal leading-relaxed">
              {props.subtitle}
            </p>
          </Show>
        </div>
        <span
          class={`text-colpsi-muted transition-transform duration-300 shrink-0 ${
            open() ? "rotate-180" : ""
          }`}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </span>
      </button>
      <Show when={open()}>
        <div class="px-5 pb-5">{props.children}</div>
      </Show>
    </section>
  );
}
// web/src/components/ui/Notebook.tsx
// Pestañas horizontales estilo Odoo (notebook): barra gris clara con la pestaña
// activa en blanco y línea superior colpsi-blue conectada al panel de contenido.
// Cada NotebookPage hace lazy-mount: la pestaña inicial se hidrata desde el SSR y
// las demás se montan al activarse por primera vez; una vez visitadas permanecen
// en el DOM (ocultas) para conservar estado local (TipTap, flatpickr, selecciones
// temporales) y evitar que los editores se inicialicen con el contenedor oculto.
import {
  createContext,
  createEffect,
  createSignal,
  Show,
  For,
  useContext,
} from "solid-js";
import type { JSX } from "solid-js";
import { Icon, type IconName } from "~/components/admin/ui/icons";

interface NotebookContextValue {
  active: () => string;
}

const NotebookContext = createContext<NotebookContextValue>();

export function Notebook(props: {
  pages: Array<{ id: string; label: string; icon?: IconName }>;
  children: JSX.Element;
}) {
  const [active, setActive] = createSignal(props.pages[0]?.id ?? "");

  return (
    <NotebookContext.Provider value={{ active }}>
      <div class="bg-white rounded-lg border border-colpsi-border overflow-hidden">
        <div
          role="tablist"
          class="flex items-end overflow-x-auto bg-colpsi-bg/60 border-b border-colpsi-border"
        >
          <For each={props.pages}>
            {(page) => (
              <button
                type="button"
                role="tab"
                aria-selected={active() === page.id}
                onClick={() => setActive(page.id)}
                class={`relative px-4 py-2.5 text-sm whitespace-nowrap border-t-2 -mb-px transition-colors ${
                  active() === page.id
                    ? "bg-white font-semibold text-colpsi-text border-t-colpsi-blue"
                    : "border-t-transparent text-colpsi-muted hover:text-colpsi-blue hover:bg-white/60"
                }`}
              >
                {page.icon && (
                  <Icon name={page.icon} class="w-3.5 h-3.5 inline mr-1.5 align-[-2px]" />
                )}
                {page.label}
              </button>
            )}
          </For>
        </div>
        <div class="p-5">{props.children}</div>
      </div>
    </NotebookContext.Provider>
  );
}

export function NotebookPage(props: { id: string; children: JSX.Element }) {
  const ctx = useContext(NotebookContext);
  // La pestaña inicial se monta desde el primer render (SSR hidrata su contenido);
  // las demás se montan la primera vez que se activan y luego permanecen ocultas.
  const [visited, setVisited] = createSignal(ctx?.active() === props.id);
  createEffect(() => {
    if (ctx?.active() === props.id) setVisited(true);
  });
  return (
    <Show when={visited()}>
      <div role="tabpanel" class={ctx?.active() === props.id ? "" : "hidden"}>
        {props.children}
      </div>
    </Show>
  );
}
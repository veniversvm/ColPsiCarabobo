// web/src/components/ui/Notebook.tsx
// Pestañas horizontales estilo Odoo con color único por posición: cada pestaña
// toma un color institucional del gremio (ciclo de 6: azul, amarillo fuerte,
// navy, vinotinto, verde, azul-claro) en su TÍTULO y su FONDO para diferenciarse
// de un vistazo. Cada pestaña conserva siempre su color propio (tintado al 12% +
// título del color); la seleccionada se pinta con el degradado azul heráldico
// del escudo (navy → azul → azul-oscuro, título blanco) para que nunca quede
// demasiado clara.
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

// Colores institucionales intercalados por posición de pestaña. Institucionales
// de @theme (azul, amarillo fuerte, navy, azul-claro) + vinotinto y verde
// complementario. Se usan tintados al 12% en inactivas y como título propio.
const TAB_COLORS = [
  "#1e3a8a",  // azul
  "#facc15",  // amarillo fuerte
  "#0a174f",  // navy
  "#722f37",  // vinotinto
  "#166534",  // verde complementario
  "#1e40af",  // azul-claro
];

// Degradado azul heráldico del escudo (mismo de la @utility bg-heraldic en
// app.css): navy → azul → azul-oscuro. Fondo de la pestaña seleccionada.
const HERALDIC_GRADIENT =
  "linear-gradient(135deg, #0a174f 0%, #1e3a8a 52%, #172554 100%)";

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
            {(page, i) => {
              const isActive = () => active() === page.id;
              const t = TAB_COLORS[i() % TAB_COLORS.length];
              return (
                <button
                  type="button"
                  role="tab"
                  aria-selected={isActive()}
                  onClick={() => setActive(page.id)}
                  style={{
                    backgroundColor: isActive() ? undefined : `${t}1f`,
                    backgroundImage: isActive() ? HERALDIC_GRADIENT : undefined,
                    color: isActive() ? "#ffffff" : t,
                  }}
                  class={`relative px-4 py-2.5 text-sm whitespace-nowrap -mb-px transition-colors hover:brightness-95 ${
                    isActive() ? "font-semibold" : "font-medium"
                  }`}
                >
                  {page.icon && (
                    <Icon name={page.icon} class="w-3.5 h-3.5 inline mr-1.5 align-[-2px]" />
                  )}
                  {page.label}
                </button>
              );
            }}
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
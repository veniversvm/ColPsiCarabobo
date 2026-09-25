// web/src/components/ui/Notebook.tsx
// Pestañas horizontales estilo Odoo con color único por posición: cada pestaña
// toma un color institucional del gremio (ciclo de 6: azul, amarillo fuerte,
// navy, vinotinto, verde, azul-claro) en su TÍTULO y su FONDO para diferenciarse
// de un vistazo. El título conserva SIEMPRE su color (legible, con la versión
// oscurecida para el amarillo); la pestaña seleccionada se eleva en BLANCO
// (patrón Odoo clásico) con franja inferior en el degradado azul heráldico del
// escudo (navy → azul → azul-oscuro) y título en negrita, sin cambiarle el color.
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
// complementario. "bg" = fondo tintado; "title" = color del título SIEMPRE fijo
// (el amarillo usa #a16207, su versión oscurecida, para que se lea sobre fondos
// claros; el resto conserva su color exacto).
const TAB_COLORS = [
  { bg: "#1e3a8a", title: "#1e3a8a" },  // azul
  { bg: "#facc15", title: "#a16207" },  // amarillo fuerte (título oscurecido)
  { bg: "#0a174f", title: "#0a174f" },  // navy
  { bg: "#722f37", title: "#722f37" },  // vinotinto
  { bg: "#166534", title: "#166534" },  // verde complementario
  { bg: "#1e40af", title: "#1e40af" },  // azul-claro
];

// Degradado azul heráldico del escudo (mismo de la @utility bg-heraldic en
// app.css): navy → azul → azul-oscuro. Franja indicadora de la pestaña activa.
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
                    backgroundColor: isActive() ? "#ffffff" : `${t.bg}1f`,
                    color: t.title,
                  }}
                  class={`relative px-4 py-2.5 text-sm whitespace-nowrap -mb-px overflow-hidden transition-colors hover:brightness-95 ${
                    isActive() ? "font-semibold" : "font-medium"
                  }`}
                >
                  {page.icon && (
                    <Icon name={page.icon} class="w-3.5 h-3.5 inline mr-1.5 align-[-2px]" />
                  )}
                  {page.label}
                  <Show when={isActive()}>
                    <span
                      class="absolute inset-x-0 bottom-0 h-1"
                      style={{ backgroundImage: HERALDIC_GRADIENT }}
                    />
                  </Show>
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
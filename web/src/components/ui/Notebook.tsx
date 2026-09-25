// web/src/components/ui/Notebook.tsx
// Pestañas horizontales estilo Odoo con color único por posición: cada pestaña
// toma un color institucional del gremio (ciclo de 6: azul, amarillo fuerte,
// navy, vinotinto, verde, azul-claro) en su TÍTULO y su FONDO para diferenciarse
// de un vistazo. El título conserva SIEMPRE su color (legible, con la versión
// oscurecida para el amarillo). La pestaña seleccionada se marca fuerte: fondo
// degradado azul heráldico del escudo (navy → azul → azul-oscuro) y el título
// (con su icono) va dentro de una píldora clara del color de su pestaña.
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
// claros; el resto conserva su color exacto); "light" = versión clara del color,
// fondo de la píldora que envuelve el título en la pestaña seleccionada.
const TAB_COLORS = [
  { bg: "#1e3a8a", title: "#1e3a8a", light: "#dbeafe" },  // azul
  { bg: "#facc15", title: "#a16207", light: "#fef9c3" },  // amarillo fuerte (título oscurecido)
  { bg: "#0a174f", title: "#0a174f", light: "#e0e7ff" },  // navy
  { bg: "#722f37", title: "#722f37", light: "#fce7f3" },  // vinotinto
  { bg: "#166534", title: "#166534", light: "#dcfce7" },  // verde complementario
  { bg: "#1e40af", title: "#1e40af", light: "#bfdbfe" },  // azul-claro
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
                    backgroundColor: isActive() ? undefined : `${t.bg}1f`,
                    backgroundImage: isActive() ? HERALDIC_GRADIENT : undefined,
                    color: isActive() ? undefined : t.title,
                  }}
                  class={`relative px-4 text-sm whitespace-nowrap -mb-px overflow-hidden transition-colors hover:brightness-95 ${
                    isActive() ? "py-1.5 font-semibold" : "py-2.5 font-medium"
                  }`}
                >
                  <Show
                    when={isActive()}
                    fallback={
                      <>
                        {page.icon && (
                          <Icon name={page.icon} class="w-3.5 h-3.5 inline mr-1.5 align-[-2px]" />
                        )}
                        {page.label}
                      </>
                    }
                  >
                    <span
                      class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5"
                      style={{ backgroundColor: t.light, color: t.title }}
                    >
                      {page.icon && <Icon name={page.icon} class="w-3.5 h-3.5" />}
                      {page.label}
                    </span>
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
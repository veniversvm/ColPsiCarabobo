// web/src/components/ui/DropdownSelect.tsx
import { For, Show, createSignal, onCleanup, onMount } from "solid-js";
import { Portal, isServer } from "solid-js/web";

export interface DropdownOption {
  value: string;
  label: string;
  disabled?: boolean;
}

interface DropdownSelectProps {
  value: string;
  onChange: (value: string) => void;
  options: DropdownOption[];
  placeholder?: string;
  disabled?: boolean;
  loading?: boolean;
  loadingLabel?: string;
  buttonClass?: string;
  panelClass?: string;
}

interface PanelRect {
  top: number;
  left: number;
  width: number;
  maxHeight: number;
}

const MAX_ALTO_PANEL = 256; // lo que daba `max-h-64`
const GAP = 8; // lo que daba `mt-2`
const MARGEN_INFERIOR = 8;

export function DropdownSelect(props: DropdownSelectProps) {
  const [open, setOpen] = createSignal(false);
  const [rect, setRect] = createSignal<PanelRect | null>(null);
  let rootRef: HTMLDivElement | undefined;
  let buttonRef: HTMLButtonElement | undefined;
  let panelRef: HTMLDivElement | undefined;

  const selected = () => props.options.find((o) => o.value === props.value);

  // El panel se ancla al botón en coordenadas de viewport porque vive en un
  // `Portal` a `document.body` (ver el comentario del JSX): si se midiera con
  // offsets del contenedor, cada scroll lo despegaría de su botón.
  const medir = (): PanelRect | null => {
    if (!buttonRef) return null;
    const b = buttonRef.getBoundingClientRect();
    const top = b.bottom + GAP;
    return {
      top,
      left: b.left,
      width: b.width,
      maxHeight: Math.min(
        MAX_ALTO_PANEL,
        Math.max(0, window.innerHeight - top - MARGEN_INFERIOR),
      ),
    };
  };

  const cerrar = () => {
    setOpen(false);
    setRect(null);
  };

  const repintar = () => {
    if (open()) setRect(medir());
  };

  const toggle = () => {
    if (open()) {
      cerrar();
    } else {
      setRect(medir());
      setOpen(true);
    }
  };

  const handleClickOutside = (e: MouseEvent) => {
    const target = e.target as Node;
    const dentro = (el: HTMLElement | undefined) =>
      !!el && el.contains(target);
    // El panel está en `document.body`, o sea FUERA de `rootRef`: sin esta segunda
    // comprobación el `mousedown` sobre una opción cerraría el panel antes de que
    // la opción reciba su `click` y no se seleccionaría nada.
    if (!dentro(rootRef) && !dentro(panelRef)) cerrar();
  };

  onMount(() => {
    document.addEventListener("mousedown", handleClickOutside);
    // `capture: true` porque los contenedores scrolleables internos (p. ej. el
    // scroll infinito del directorio) no propagan `scroll` a `window`.
    window.addEventListener("scroll", repintar, { capture: true, passive: true });
    window.addEventListener("resize", repintar);
  });

  onCleanup(() => {
    if (typeof document !== "undefined") {
      document.removeEventListener("mousedown", handleClickOutside);
    }
    if (typeof window !== "undefined") {
      window.removeEventListener("scroll", repintar, true);
      window.removeEventListener("resize", repintar);
    }
  });

  const baseButtonClass =
    "w-full flex items-center justify-between gap-2 text-left outline-none transition-all cursor-pointer disabled:opacity-60 disabled:cursor-wait";

  const renderButton = () => {
    const label = props.loading
      ? props.loadingLabel || "Cargando..."
      : selected()
        ? selected()!.label
        : props.placeholder || "Seleccionar...";
    return (
      <>
        <span class="truncate">{label}</span>
        <svg
          class={`h-5 w-5 shrink-0 text-colpsi-blue opacity-30 transition-transform ${open() ? "rotate-180" : ""}`}
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fill-rule="evenodd"
            d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
            clip-rule="evenodd"
          />
        </svg>
      </>
    );
  };

  return (
    <div ref={rootRef} class="relative">
      <button
        ref={buttonRef}
        type="button"
        disabled={props.disabled || props.loading}
        onClick={toggle}
        onKeyDown={(e) => {
          if (e.key === "Escape") cerrar();
        }}
        aria-haspopup="listbox"
        aria-expanded={open()}
        class={`${props.buttonClass || ""} ${baseButtonClass}`}
      >
        {renderButton()}
      </button>

      {/* El panel va en un `Portal` a `document.body` porque el hero del directorio
          tiene `isolate` (obligatorio para que se vea la foto, gotcha 15 de
          web/AGENTS.md) y `overflow-hidden` (para recortar la imagen): cualquiera
          de los dos atrapa un panel `absolute` con `z-30` dentro de él, que es
          justo lo que lo escondía detrás de las tarjetas. Fuera del hero no hay
          stacking context ni recorte que lo ataquen, y `z-[10000]` lo clears por
          encima del navbar (`z-50`) y de cualquier overlay. */}
      <Show when={!isServer}>
        <Show when={open() && rect()}>
          {(r) => (
            <Portal mount={document.body}>
              <div
                ref={panelRef}
                class={`fixed z-[10000] bg-white rounded-2xl shadow-2xl border border-colpsi-border overflow-y-auto py-2 ${props.panelClass || ""}`}
                style={{
                  top: `${r().top}px`,
                  left: `${r().left}px`,
                  width: `${r().width}px`,
                  "max-height": `${r().maxHeight}px`,
                }}
              >
                <Show when={props.options.length === 0} fallback={null}>
                  <p class="px-5 py-3 text-sm text-gray-400">
                    {props.loading ? "Cargando..." : "Sin opciones"}
                  </p>
                </Show>
                <For each={props.options}>
                  {(option) => (
                    <button
                      type="button"
                      disabled={option.disabled}
                      onClick={() => {
                        props.onChange(option.value);
                        cerrar();
                      }}
                      classList={{
                        "bg-blue-50 text-colpsi-blue font-bold":
                          option.value === props.value,
                        "opacity-40 cursor-not-allowed": option.disabled,
                        "hover:bg-blue-50 text-colpsi-text":
                          option.value !== props.value && !option.disabled,
                      }}
                      class="w-full text-left px-5 py-2.5 text-sm transition-colors"
                    >
                      {option.label}
                    </button>
                  )}
                </For>
              </div>
            </Portal>
          )}
        </Show>
      </Show>
    </div>
  );
}

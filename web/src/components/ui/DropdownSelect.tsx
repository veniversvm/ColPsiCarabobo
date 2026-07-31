// web/src/components/ui/DropdownSelect.tsx
import { For, Show, createSignal, onCleanup, onMount } from "solid-js";

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

export function DropdownSelect(props: DropdownSelectProps) {
  const [open, setOpen] = createSignal(false);
  let rootRef: HTMLDivElement | undefined;

  const selected = () => props.options.find((o) => o.value === props.value);

  const handleClickOutside = (e: MouseEvent) => {
    if (rootRef && !rootRef.contains(e.target as Node)) setOpen(false);
  };

  onMount(() => {
    document.addEventListener("mousedown", handleClickOutside);
  });

  onCleanup(() => {
    if (typeof document !== "undefined") {
      document.removeEventListener("mousedown", handleClickOutside);
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
        type="button"
        disabled={props.disabled || props.loading}
        onClick={() => setOpen((v) => !v)}
        onKeyDown={(e) => {
          if (e.key === "Escape") setOpen(false);
        }}
        aria-haspopup="listbox"
        aria-expanded={open()}
        class={`${props.buttonClass || ""} ${baseButtonClass}`}
      >
        {renderButton()}
      </button>

      <Show when={open()}>
        <div
          class={`absolute left-0 right-0 z-30 mt-2 bg-white rounded-2xl shadow-2xl border border-gray-100 max-h-64 overflow-y-auto py-2 ${props.panelClass || ""}`}
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
                  setOpen(false);
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
      </Show>
    </div>
  );
}

// web/src/components/inscripcion/CheckField.tsx
//
// Campo de texto con validación asíncrona contra la API (unicidad de Cédula / FPV).
// - Se dispara tras 500ms de debounce o al perder el foco.
// - Estados visuales: normal / validando / válido / inválido.
import { createSignal, Show, onCleanup, onMount } from "solid-js";
import { apiGet } from "~/lib/api";
import type { UniquenessCheckResponse } from "~/types/inscription";

interface Props {
  label: string;
  required?: boolean;
  endpoint: string;   // p.ej. "/inscripcion/check-ci"
  param: string;      // p.ej. "ci"
  placeholder?: string;
  initialValue?: string;
  onValid: (value: string) => void;
  onInvalid: (message: string) => void;
  onChange?: (raw: string) => void;
}

export function CheckField(props: Props) {
  const [value, setValue] = createSignal(props.initialValue ?? "");
  const [state, setState] = createSignal<"idle" | "checking" | "valid" | "invalid">("idle");
  const [message, setMessage] = createSignal("");

  let timer: any;
  onCleanup(() => { if (timer) clearTimeout(timer); });

  const check = (raw: string) => {
    if (timer) clearTimeout(timer);
    const trimmed = raw.trim();
    setValue(raw);
    props.onChange?.(raw);

    if (!trimmed) {
      setState("idle");
      setMessage("");
      props.onInvalid("");
      return;
    }

    setState("checking");
    setMessage("Validando...");
    timer = setTimeout(async () => {
      try {
        const res = await apiGet<UniquenessCheckResponse>(
          `${props.endpoint}?${props.param}=${encodeURIComponent(trimmed)}`
        );
        if (res.exists) {
          setState("invalid");
          setMessage(res.message || "Ya se encuentra registrado");
          props.onInvalid(res.message || "Ya se encuentra registrado");
        } else {
          setState("valid");
          setMessage("Disponible");
          props.onValid(trimmed);
        }
      } catch {
        setState("idle");
        setMessage("");
        props.onInvalid("");
      }
    }, 500);
  };

  // Al restaurar un valor persistido, re-validarlo para recuperar su estado visual.
  onMount(() => {
    const v = props.initialValue ?? "";
    if (v && value() === v) check(v);
  });

  const borderColor = () => {
    const s = state();
    if (s === "checking") return "border-blue-400";
    if (s === "valid") return "border-green-400";
    if (s === "invalid") return "border-red-400";
    return "border-gray-200";
  };

  return (
    <label class="block">
      <span class="block text-sm font-bold text-gray-700 mb-1.5">
        {props.label} {props.required && <span class="text-red-500">*</span>}
      </span>
      <div class="relative">
        <input
          type="text"
          inputmode="numeric"
          value={value()}
          placeholder={props.placeholder}
          onInput={(e) => check((e.target as HTMLInputElement).value)}
          class={`w-full px-4 py-3 bg-white border ${borderColor()} rounded-xl outline-none transition-all focus:ring-2 focus:ring-blue-100`}
        />
        <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-black">
          <Show when={state() === "checking"}>
            <span class="animate-spin inline-block h-4 w-4 border-2 border-blue-400 border-t-transparent rounded-full" />
          </Show>
          <Show when={state() === "valid"}>
            <span class="text-green-500">✓</span>
          </Show>
          <Show when={state() === "invalid"}>
            <span class="text-red-500">✗</span>
          </Show>
        </span>
      </div>
      <Show when={message()}>
        <span class={`block text-xs mt-1 font-semibold ${state() === "invalid" ? "text-red-500" : "text-green-600"}`}>
          {message()}
        </span>
      </Show>
    </label>
  );
}
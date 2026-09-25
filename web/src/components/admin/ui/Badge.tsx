// web/src/components/admin/ui/Badge.tsx
// Micro-etiqueta de estado semántico (Activo, Pendiente, Error, …).
import { JSX, Show } from "solid-js";

type BadgeTone = "success" | "warning" | "danger" | "info" | "neutral";

const tones: Record<BadgeTone, string> = {
  success: "bg-emerald-50 text-emerald-700 border-emerald-200",
  warning: "bg-amber-50 text-amber-700 border-amber-200",
  danger:  "bg-red-50 text-red-700 border-red-200",
  info:    "bg-blue-50 text-blue-700 border-blue-200",
  neutral: "bg-slate-100 text-slate-600 border-slate-200",
};

interface BadgeProps {
  tone?: BadgeTone;
  /** Muestra el punto de estado a la izquierda (default: true) */
  dot?: boolean;
  children: JSX.Element;
  class?: string;
}

export function Badge(props: BadgeProps) {
  const tone = () => props.tone ?? "neutral";
  return (
    <span
      class={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded border text-xs font-medium whitespace-nowrap ${props.class === undefined ? tones[tone()] : ""} ${props.class ?? ""}`}
    >
      <Show when={props.dot !== false}>
        <span class="w-1.5 h-1.5 rounded-full bg-current" />
      </Show>
      {props.children}
    </span>
  );
}
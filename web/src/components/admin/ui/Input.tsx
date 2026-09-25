// web/src/components/admin/ui/Input.tsx
// Input estándar del admin: borde visible, focus azul institucional.
import { JSX, splitProps } from "solid-js";

interface InputProps extends JSX.InputHTMLAttributes<HTMLInputElement> {
  class?: string;
}

export function Input(props: InputProps) {
  const [local, rest] = splitProps(props, ["class"]);
  return (
    <input
      class={`h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-colpsi-text outline-none transition-colors placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 ${local.class ?? ""}`}
      {...rest}
    />
  );
}
// web/src/components/admin/ui/Button.tsx
// Set único de botones del admin: planos, rounded-md, sin sombras.
import { JSX, splitProps } from "solid-js";

type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";
type ButtonSize = "sm" | "md";

const variants: Record<ButtonVariant, string> = {
  primary:   "bg-colpsi-blue text-white hover:bg-colpsi-blue-light focus:ring-colpsi-blue/30",
  secondary: "border border-colpsi-border bg-white text-colpsi-text hover:bg-colpsi-bg focus:ring-slate-300/40",
  ghost:     "text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg focus:ring-slate-300/30",
  danger:    "bg-colpsi-red text-white hover:opacity-90 focus:ring-colpsi-red/30",
};

const sizes: Record<ButtonSize, string> = {
  sm: "h-8 px-2.5 text-xs",
  md: "h-9 px-3.5 text-sm",
};

interface ButtonProps extends JSX.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  class?: string;
  children: JSX.Element;
}

export function Button(props: ButtonProps) {
  const [local, rest] = splitProps(props, ["variant", "size", "class", "children"]);
  const variant = () => local.variant ?? "secondary";
  const size = () => local.size ?? "md";
  return (
    <button
      class={`inline-flex items-center justify-center gap-2 rounded-md font-medium outline-none transition-colors focus:ring-2 disabled:opacity-50 disabled:cursor-not-allowed ${variants[variant()]} ${sizes[size()]} ${local.class ?? ""}`}
      {...rest}
    >
      {local.children}
    </button>
  );
}
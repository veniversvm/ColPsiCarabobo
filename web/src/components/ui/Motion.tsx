// web/src/components/ui/Motion.tsx
// Acceso centralizado a solid-motionone (componentes <motion.*>) +
// presets reutilizables para las transiciones de entrada/salida del portal.
import { JSX, ParentProps } from "solid-js";
import { Motion, Presence, motion } from "solid-motionone";

export { Motion, Presence, motion };

export type AnimateVariant =
  | "fade"
  | "zoom"
  | "slide-top"
  | "slide-bottom"
  | "shake";

type VariantTarget = Record<string, string | number>;

interface AnimatePreset {
  initial: VariantTarget;
  animate: VariantTarget;
  transition: { duration: number; easing?: string };
}

export const presets: Record<AnimateVariant, AnimatePreset> = {
  fade: {
    initial: { opacity: 0 },
    animate: { opacity: 1 },
    transition: { duration: 0.3, easing: "ease-out" },
  },
  zoom: {
    initial: { opacity: 0, scale: 0.95 },
    animate: { opacity: 1, scale: 1 },
    transition: { duration: 0.2, easing: "ease-out" },
  },
  "slide-top": {
    initial: { opacity: 0, y: -10 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.3, easing: "ease-out" },
  },
  "slide-bottom": {
    initial: { opacity: 0, y: 10 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.3, easing: "ease-out" },
  },
  shake: {
    initial: { x: 0 },
    animate: { x: [0, -10, 10, -8, 8, -4, 4, 0] },
    transition: { duration: 0.5, easing: "ease-in-out" },
  },
};

interface AnimateProps extends ParentProps {
  variant: AnimateVariant;
  class?: string;
  exit?: VariantTarget;
  style?: JSX.CSSProperties;
  onClick?: (e: MouseEvent) => void;
}

export function Animate(props: AnimateProps) {
  const preset = presets[props.variant];
  return (
    <Motion.div
      class={props.class}
      style={props.style}
      initial={preset.initial}
      animate={preset.animate}
      transition={preset.transition}
      exit={props.exit ? { ...preset.initial, ...props.exit } : undefined}
      onClick={props.onClick}
    >
      {props.children}
    </Motion.div>
  );
}
// web/src/components/ui/RouteTransition.tsx
// Transición suave al navegar entre rutas usando la Web Animations API.
// IMPORTANTE: NO usar keyed <Show> para esta animación (rompe la navegación SPA);
// se anima el contenedor actual disparado por useLocation().pathname.
import { createEffect, JSX } from "solid-js";
import { useLocation } from "@solidjs/router";

export default function RouteTransition(props: { children: JSX.Element }) {
  const location = useLocation();
  let ref: HTMLDivElement | undefined;

  createEffect(() => {
    location.pathname;
    if (typeof window === "undefined") return;
    if (window.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches) return;
    queueMicrotask(() => {
      if (!ref || typeof ref.animate !== "function") return;
      ref.animate(
        [{ opacity: 0, transform: "translateY(4px)" }, { opacity: 1, transform: "translateY(0)" }],
        { duration: 220, easing: "ease-out" }
      );
    });
  });

  return <div ref={ref}>{props.children}</div>;
}
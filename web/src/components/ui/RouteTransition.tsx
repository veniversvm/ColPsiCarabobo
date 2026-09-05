// web/src/components/ui/RouteTransition.tsx
// Transición cross-fade nativa del navegador (View Transitions API) al navegar.
// Intercepta la navegación con useBeforeLeave y la envuelve en
// document.startViewTransition(): el navegador captura el snapshot de la página
// saliente y hace el cross-fade hacia la entrante sin clonar el DOM ni usar
// overlays (el clonado causaba parpadeo en franjas animadas y congelaba rutas
// con layout compartido como admin). Fallbacks seguros:
//   - Sin startViewTransition o prefers-reduced-motion → no preventDefault,
//     la navegación sigue normal (sin efecto, cero riesgo de romper el SPA).
// Componente pasivo (solo registra el listener); se monta desde app.tsx.
import { useBeforeLeave } from "@solidjs/router";

export default function RouteTransition() {
  let inTransition = false;

  useBeforeLeave((e) => {
    if (typeof window === "undefined") return;
    if (e.defaultPrevented || inTransition) return;
    if (typeof document.startViewTransition !== "function") return;
    if (window.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches) return;
    e.preventDefault();
    inTransition = true;
    const vt = document.startViewTransition(() => e.retry());
    void vt.finished.finally(() => {
      inTransition = false;
    });
  });

  return null;
}
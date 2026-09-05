// web/src/components/ui/RouteTransition.tsx
// Transición cross-fade nativa del navegador (View Transitions API) al navegar
// entre rutas PÚBLICAS. Dentro de los paneles con layout persistente (/admin,
// /psi) se deja la navegación SPA nativa sin transición: crossfadear la página
// completa ahí causaba parpadeo del sidebar/barras persistentes.
// Solo navegaciones que cruzan a/desde rutas públicas se envuelven en
// document.startViewTransition(). Fallbacks seguros:
//   - Sin startViewTransition o prefers-reduced-motion → navegación normal.
//   - Rutas sin path (from/to pueden ser undefined) → navegación normal.
// Componente pasivo; se monta desde app.tsx.
import { useBeforeLeave } from "@solidjs/router";

const PANEL_PREFIX = ["/admin", "/psi"];

function isPublicPath(path: string | undefined) {
  if (!path) return true;
  return !PANEL_PREFIX.some((p) => path === p || path.startsWith(p + "/"));
}

export default function RouteTransition() {
  let inTransition = false;

  useBeforeLeave((e) => {
    if (typeof window === "undefined") return;
    if (e.defaultPrevented || inTransition) return;
    if (typeof document.startViewTransition !== "function") return;
    if (window.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches) return;

    // Navegaciones dentro de paneles persistentes → sin transición.
    const fromPublic = isPublicPath(e.from?.path);
    const toPublic = isPublicPath(e.to?.path);
    if (!fromPublic && !toPublic) return;

    e.preventDefault();
    inTransition = true;
    const vt = document.startViewTransition(() => e.retry());
    void vt.finished.finally(() => {
      inTransition = false;
    });
  });

  return null;
}

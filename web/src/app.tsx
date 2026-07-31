// web/src/app.tsx
// Raíz de la aplicación: provee Meta, Router (FileRoutes), AuthProvider,
// Navbar global y ErrorBoundary con fallback OfflineAlert.
import { ErrorBoundary, onMount, Suspense, Show, createSignal, onCleanup, type JSX } from "solid-js";
import { Router, useLocation } from "@solidjs/router";
import { FileRoutes } from "@solidjs/start/router";
import { MetaProvider } from "@solidjs/meta";
import { AuthProvider } from "~/lib/auth";
import OfflineAlert from "~/components/ui/OfflineAlert";
import Navbar from "~/components/layaout/Navbar";
import "./app.css";

// Transición de ruta: cross-fade suave entre secciones.
// Al navegar, la página anterior se captura como overlay (debajo del Navbar,
// z-40) y se desvanece con la Web Animations API mientras la nueva entra con
// un fade+slide sutil. Sin dependencias; con prefers-reduced-motion no anima.
function RouteTransition(props: { children: JSX.Element }) {
  const location = useLocation();
  const [leaving, setLeaving] = createSignal<HTMLDivElement | null>(null);

  return (
    <>
      <Show when={location.pathname} keyed>
        {() => {
          let el: HTMLDivElement | null = null;
          onCleanup(() => {
            if (!el) return;
            const reduceMotion =
              typeof window !== "undefined" &&
              window.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches;
            if (reduceMotion) return;
            el.classList.remove("route-fade");
            setLeaving(el);
            const anim = el.animate(
              [{ opacity: 1 }, { opacity: 0 }],
              { duration: 250, easing: "ease-in" },
            );
            anim.onfinish = () => setLeaving(null);
          });
          return (
            <div class="route-fade" ref={(r) => { el = r; }}>
              {props.children}
            </div>
          );
        }}
      </Show>
      <Show when={leaving()}>
        {(oldEl) => (
          <div
            class="pointer-events-none fixed inset-0 z-40 overflow-y-auto"
            aria-hidden="true"
            ref={(layer) => {
              if (layer && oldEl()) layer.scrollTop = window.scrollY;
            }}
          >
            {oldEl()}
          </div>
        )}
      </Show>
    </>
  );
}

export default function App() {

  /**
   * Se imrpime correctamente, se ve desproporcionado solo en el editor
   */
  onMount(() => {
    const logo = `
       * * *
     *       *
    *  ,;;;,  *
     */;;-;;\\*
     /;/   \\;\\
    /);|)-(|;;\\
   ;;;/ \`"  \\;(;
   |(|\\_/|\\_/|;|
   |;|_|/^\\|_|;|
   |;;\\=:=:=/;)|
   |:;| : : |;:|
   |);\\ : : /;;|
   ;;;| _:_ |;(;
   \\;;\\  |  /;;/
    |(;\\   /;;|
     \\;;| |;;/
      |;| |;|
     .'\`-.-''.
    /   .-.  (\\
   |   Q   \\__)|
   '-.__   __.-'
        \`\`\`

    SANCTA MARIA 
        ORA
     PRO NOBIS
    `;
    console.log(logo);
  });

  return (
    <MetaProvider>
      <Router
        root={(props) => (
          <AuthProvider>
            <Navbar />
            <ErrorBoundary fallback={(err, reset) => {
              console.error("[ErrorBoundary]", err)
              return <OfflineAlert error={err} reset={reset} />
              }
            }>
              <Suspense>
                <RouteTransition>
                  {props.children}
                </RouteTransition>
              </Suspense>
            </ErrorBoundary>
          </AuthProvider>
        )}
      >
        <FileRoutes />
      </Router>
    </MetaProvider>
  );
}
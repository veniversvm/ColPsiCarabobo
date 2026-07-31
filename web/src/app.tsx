// web/src/app.tsx
// Raíz de la aplicación: provee Meta, Router (FileRoutes), AuthProvider,
// Navbar global y ErrorBoundary con fallback OfflineAlert.
import { ErrorBoundary, onMount, Suspense } from "solid-js";
import { Router } from "@solidjs/router";
import { FileRoutes } from "@solidjs/start/router";
import { MetaProvider } from "@solidjs/meta";
import { AuthProvider } from "~/lib/auth";
import OfflineAlert from "~/components/ui/OfflineAlert";
import Navbar from "~/components/layaout/Navbar";
import "./app.css";

export default function App() {
  onMount(() => {
    console.log("Colegio de Psicólogos de Carabobo — frontend");
  });

  return (
    <MetaProvider>
      <Router
        root={(props) => (
          <AuthProvider>
            <Navbar />
            <ErrorBoundary
              fallback={(err, reset) => {
                console.error("[ErrorBoundary]", err);
                return <OfflineAlert error={err} reset={reset} />;
              }}
            >
              <Suspense>{props.children}</Suspense>
            </ErrorBoundary>
          </AuthProvider>
        )}
      >
        <FileRoutes />
      </Router>
    </MetaProvider>
  );
}

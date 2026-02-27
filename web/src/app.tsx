// web/src/app.tsx
import { ErrorBoundary, Suspense } from "solid-js";
import { Router } from "@solidjs/router";
import { FileRoutes } from "@solidjs/start/router";
import { AuthProvider } from "~/lib/auth";
import OfflineAlert from "~/components/ui/OfflineAlert";
import Navbar from "~/components/layaout/Navbar";
import "./app.css";

export default function App() {
  return (
    <Router
      root={(props) => (
        <AuthProvider>
          <Navbar />
          {/* Este ErrorBoundary capturará cualquier fallo en las peticiones de datos */}
          <ErrorBoundary fallback={(err, reset) => <OfflineAlert error={err} reset={reset} />}>
            <Suspense>
              {props.children}
            </Suspense>
          </ErrorBoundary>
        </AuthProvider>
      )}
    >
      <FileRoutes />
    </Router>
  );
}
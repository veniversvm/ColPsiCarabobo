// web/src/app.tsx
import { ErrorBoundary, Suspense } from "solid-js";
import { Router } from "@solidjs/router";
import { FileRoutes } from "@solidjs/start/router";
import { MetaProvider } from "@solidjs/meta";
import { AuthProvider } from "~/lib/auth";
import OfflineAlert from "~/components/ui/OfflineAlert";
import Navbar from "~/components/layaout/Navbar";
import "./app.css";

export default function App() {
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
                {props.children}
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
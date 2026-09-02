// web/src/app.tsx
// Raíz de la aplicación: provee Meta, Router (FileRoutes), AuthProvider,
// Navbar global y ErrorBoundary con fallback OfflineAlert.
import { ErrorBoundary, onMount, Suspense } from "solid-js";
import { Router, action, useAction } from "@solidjs/router";
import { FileRoutes } from "@solidjs/start/router";
import { MetaProvider } from "@solidjs/meta";
import { deleteCookie } from "vinxi/http";
import { AuthProvider } from "~/lib/auth";
import OfflineAlert from "~/components/ui/OfflineAlert";
import Navbar from "~/components/layaout/Navbar";
import "./app.css";

// Server action que borra las cookies de sesión HttpOnly del frontend.
// Se define en el root (mismo patrón que los actions de las rutas) para que
// la limpieza corra en el servidor del frontend, que es quien posee la cookie.
const clearSessionCookiesAction = action(async () => {
  "use server";
  deleteCookie("jwt", { path: "/" });
  deleteCookie("user_data", { path: "/" });
});

export default function App() {
  onMount(() => {
    console.log("Colegio de Psicólogos de Carabobo — frontend");
  });

  return (
    <MetaProvider>
      <Router
        root={(props) => {
          const clearServer = useAction(clearSessionCookiesAction);
          return (
            <AuthProvider onServerLogout={clearServer}>
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
          );
        }}
      >
        <FileRoutes />
      </Router>
    </MetaProvider>
  );
}

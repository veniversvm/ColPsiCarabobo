// web/src/app.tsx
import { ErrorBoundary, onMount, Suspense } from "solid-js";
import { Router } from "@solidjs/router";
import { FileRoutes } from "@solidjs/start/router";
import { MetaProvider } from "@solidjs/meta";
import { AuthProvider } from "~/lib/auth";
import OfflineAlert from "~/components/ui/OfflineAlert";
import Navbar from "~/components/layaout/Navbar";
import "./app.css";

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
// web/src/entry-client.tsx
//
// Punto de entrada CSR: hidrata la app Solid en el elemento #app
// que genera `entry-server.tsx`.

// @refresh reload
import { mount, StartClient } from "@solidjs/start/client";

mount(() => <StartClient />, document.getElementById("app")!);

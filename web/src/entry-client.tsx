// web/src/entry-client.tsx
//
// Punto de entrada CSR. Hidrata la app Solid en el elemento #app que genera
// `entry-server.tsx`.
//
// Los paneles privados (/admin y /psi) se renderizan SOLO en el cliente (SPA):
// `entry-server.tsx` devuelve la carcasa vacía sin marcadores de hidratación,
// así que aquí se montan con `render` (no `hydrate`) para no intentar emparejar
// claves de hidratación inexistentes. El resto de rutas (públicas, con SSR)
// se hidratan con `mount`.

// @refresh reload
import { render } from "solid-js/web";
import { mount, StartClient } from "@solidjs/start/client";

const isSpaPath = (pathname: string) =>
  pathname === "/admin" ||
  pathname.startsWith("/admin/") ||
  pathname === "/psi" ||
  pathname.startsWith("/psi/");

const appRoot = document.getElementById("app")!;

if (isSpaPath(window.location.pathname)) {
  render(() => <StartClient />, appRoot);
} else {
  mount(() => <StartClient />, appRoot);
}
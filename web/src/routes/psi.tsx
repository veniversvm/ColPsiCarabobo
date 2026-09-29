// web/src/routes/psi.tsx
// Layout de las rutas `/psi/*`: envuelve todas las páginas del portal del
// psicólogo con el notificador de nuevas notificaciones (sondeo + sonido) y con
// el aviso de aceptación de las condiciones del portal del agremiado.
//
// El modal va AQUÍ y no dentro de cada página: el aviso tiene que aparecer sobre
// cualquier ruta del portal, incluida una que se recargue por F5 en
// `/psi/tickets/123`. Montarlo por página lo dejaría saltando según el
// recorrido.
//
// `activo` se apaga en `/psi/terminos`: esa página ya trae su propio panel de
// aceptación con el historial, y tener los dos a la vez es redundante.
import { Show, type JSX } from "solid-js";
import { useLocation } from "@solidjs/router";
import { NotificationsProvider } from "~/lib/notifications";
import AceptarTerminosModal from "~/components/terminos/AceptarTerminosModal";

export default function PsiLayout(props: { children: JSX.Element }) {
  const location = useLocation();
  const enPaginaTerminos = () => location.pathname.startsWith("/psi/terminos");

  return (
    <NotificationsProvider>
      {props.children}
      <Show when={!enPaginaTerminos()}>
        <AceptarTerminosModal activo={true} />
      </Show>
    </NotificationsProvider>
  );
}

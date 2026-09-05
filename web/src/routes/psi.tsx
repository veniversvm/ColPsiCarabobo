// web/src/routes/psi.tsx
// Layout de las rutas `/psi/*`: envuelve todas las páginas del portal del
// psicólogo con el notificador de nuevas notificaciones (sondeo + sonido).
import { JSX } from "solid-js";
import { NotificationsProvider } from "~/lib/notifications";

export default function PsiLayout(props: { children: JSX.Element }) {
  return <NotificationsProvider>{props.children}</NotificationsProvider>;
}
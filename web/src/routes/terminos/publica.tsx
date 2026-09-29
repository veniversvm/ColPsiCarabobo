// web/src/routes/terminos/publica.tsx → /terminos/publica
// URL ANTIGUA. El documento público vive en `/terminos`; esta ruta solo
// redirige para que un enlace guardado, marcado o copiado no quede en 404.
//
// Por qué redirect de cliente y no un 301: esta URL nunca llegó a producción
// (el documento no estaba desplegado), así que no hay links entrantes ni
// posicionamiento que preservar. Es el mismo patrón que ya usan `/psi/terminos`
// y `routes/reset-password.tsx`. Si alguna vez se publicara y pasara a tener
// enlaces, esto debería volverse un 301 de verdad.
//
// El slug en sí era una fuga: "publica" anuncia que hay una parte no pública.

import { createEffect } from "solid-js";
import { useNavigate } from "@solidjs/router";

export default function TerminosPublicaRedirect() {
  const navigate = useNavigate();

  createEffect(() => {
    navigate("/terminos", { replace: true });
  });

  // Placeholder mientras ocurre la redirección (un frame). No muestra nada del
  // documento para que no se vea un destello de la página anterior.
  return (
    <main class="min-h-screen bg-colpsi-bg" aria-busy="true">
      <span class="sr-only">Redirigiendo…</span>
    </main>
  );
}

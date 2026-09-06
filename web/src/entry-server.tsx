// web/src/entry-server.tsx
//
// Punto de entrada SSR (Deno). Monta el HTML base y aplica una Content
// Security Policy (meta tag) que:
// - Permite estilos/scripts inline (necesarios para el hydrate de Solid).
// - Permite `img-src` desde el origen del bucket S3/MinIO (BUCKET_ORIGIN),
//   extraído de `VITE_BUCKET_URL`. Si se cambia el bucket, actualizar la env
//   y reconstruir, porque la URL se inlinea en el bundle.
// - Fija `base-uri`/`form-action`. `frame-ancestors` NO va aquí: es ignorado en
//   un `<meta>`; el clickjacking se cubre con los headers de la API (X-Frame-Options).
// - `connect-src` incluye `ws://localhost:*`/`wss://localhost:*` para el HMR de
//   Vinxi en desarrollo (el websocket de `/_build`).

// @refresh reload
import { createHandler, StartServer } from "@solidjs/start/server";

// Origen del bucket S3/MinIO (puede ser http en desarrollo) para img-src del CSP
const BUCKET_ORIGIN =
  (import.meta.env.VITE_BUCKET_URL || "").match(/^https?:\/\/[^/]+/i)?.[0] || "";

const CSP = `default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:${BUCKET_ORIGIN ? ` ${BUCKET_ORIGIN}` : ""}; connect-src 'self' data: https: http://localhost:* ws://localhost:* wss://localhost:*; font-src 'self' data:; form-action 'self'; base-uri 'self'`;

// Los paneles privados (/admin y /psi) se renderizan SOLO en el cliente (SPA).
// En SSR se devuelve la carcasa vacía y el documento NO accede a props.children,
// así el App (y sus rutas con código solo-cliente, p.ej. el kanban) nunca se
// evalúa en el servidor. El navegador monta la ruta desde window.location.
const isSpaPath = (pathname: string) =>
  pathname === "/admin" ||
  pathname.startsWith("/admin/") ||
  pathname === "/psi" ||
  pathname.startsWith("/psi/");

export default createHandler((event) => {
  const pathname = new URL(event.request.url).pathname;
  const spa = isSpaPath(pathname);

  return (
    <StartServer
      document={(props) => (
        <html lang="es"> {/* <-- Cambiado a español */}
          <head>
            <meta charset="utf-8" />
            <meta name="viewport" content="width=device-width, initial-scale=1" />
            
            {/* Título de la pestaña */}
            <title>Colegio de Psicólogos del Estado Carabobo</title>

            {/* Content Security Policy */}
            <meta
              http-equiv="Content-Security-Policy"
              content={CSP}
            />
            
            {/* Ícono de la pestaña referenciando a la carpeta public */}
            <link rel="icon" href="/psi.png" />

            {/* Preload de la tipografía Inter autoalojada (ver MANUAL_ESTILO §4) */}
            <link rel="preload" href="/fonts/inter-latin.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />

            {props.assets}
          </head>
          <body>
            <div id="app">{spa ? null : props.children}</div>
            {props.scripts}
          </body>
        </html>
      )}
    />
  );
});
// web/src/entry-server.tsx
//
// Punto de entrada SSR (Deno). Monta el HTML base y aplica una Content
// Security Policy (meta tag) que:
// - Permite estilos/scripts inline (necesarios para el hydrate de Solid).
// - Permite `img-src` desde el origen del bucket S3/MinIO (BUCKET_ORIGIN),
//   extraído de `VITE_BUCKET_URL`. Si se cambia el bucket, actualizar la env
//   y reconstruir, porque la URL se inlinea en el bundle.
// - Bloquea `frame-ancestors` (clickjacking) y fija `base-uri`/`form-action`.

// @refresh reload
import { createHandler, StartServer } from "@solidjs/start/server";

// Origen del bucket S3/MinIO (puede ser http en desarrollo) para img-src del CSP
const BUCKET_ORIGIN =
  (import.meta.env.VITE_BUCKET_URL || "").match(/^https?:\/\/[^/]+/i)?.[0] || "";

const CSP = `default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:${BUCKET_ORIGIN ? ` ${BUCKET_ORIGIN}` : ""}; connect-src 'self' https: http://localhost:*; font-src 'self' data:; frame-ancestors 'none'; form-action 'self'; base-uri 'self'`;

export default createHandler(() => (
  <StartServer
    document={({ assets, children, scripts }) => (
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
          
          {assets}
        </head>
        <body>
          <div id="app">{children}</div>
          {scripts}
        </body>
      </html>
    )}
  />
));
// web/src/entry-server.tsx

// @refresh reload
import { createHandler, StartServer } from "@solidjs/start/server";

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
            content="default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https: http://localhost:*; font-src 'self' data:; frame-ancestors 'none'; form-action 'self'; base-uri 'self'"
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
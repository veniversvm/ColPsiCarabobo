# AGENTS.md — Frontend (`web/`)

Guía operativa específica del frontend para agentes AI. Complementa el
[`AGENTS.md` raíz](../AGENTS.md); ante conflicto, esta guía es la fuente de verdad
para todo lo que viva dentro de `web/`.

## Contexto rápido

- **SolidStart** (SolidJS + TypeScript), SSR + CSR, rutas por archivos en `src/routes/`.
- La API es Go y vive en `api/`; el frontend se comunica por HTTP (`src/lib/api.ts`).
- El stack es **isomórfico**: todo el código se ejecuta en Deno (SSR) y en el navegador
  (CSR). Cualquier API de navegador/DOM solo se usa en cliente (`if (!isServer)`).
- CSS con Tailwind v4; editor RTF con TipTap v3.

## Comandos (verificar SIEMPRE con `npm run build`)

```bash
npm run dev          # desarrollo con hot reload
npm run build        # build SSR → .output/  (usar tras cambios)
npm run start        # servir el build
deno task build      # igual que npm run build (lo usa el dockerfile)
```

> No hay typecheck configurado (TypeScript no está instalado); `npm run build`
> es la verificación. El build en Docker es Deno (`deno task build`); asegúrate
> de que el código compila con Deno, no solo con Node.

## Reglas críticas (gotchas que ya rompieron el proyecto)

1. **Server actions (`"use server"`)** — deben invocarse con `useAction(fn)` del
   `@solidjs/router`, **nunca** llamando la función directamente. Llamarla directo
   lanza `TypeError: this.r.singleFlight` (el router solo inyecta `this.r` vía
   `useAction`). Ejemplo correcto en `src/routes/admin-access.tsx:24`.

2. **`VITE_*` se inlinean en build** — `VITE_API_URL`, `VITE_BUCKET_URL`,
   `VITE_SITE_URL` se incrustan en el bundle. Cambiarlas en `.env` no tiene efecto
   hasta reconstruir (`docker compose build web`). `API_URL_INTERNAL` (SSR) sí se
   lee en runtime.

3. **Imágenes del bucket** — SIEMPRE usar `bucketUrl(key)` de `src/lib/bucket.ts`.
   Prohibido hardcodear `http://localhost:9000` o `s3:9000` en componentes. La
   función es idempotente (acepta keys y URLs completas). El CSP agrega el origen
   del bucket a `img-src` automáticamente.

4. **DOMPurify sin `.sanitize` en SSR** — En Deno (sin DOM) `dompurify` exporta una
   fábrica. Usa `sanitizeHtml()` de `src/lib/sanitize-html.ts` (fallback SSR
   incluido), no `DOMPurify` directamente. La barrera final XSS es la API Go
   (bluemonday).

5. **Token JWT isomórfico** — `src/lib/api.ts` decide la fuente: en SSR lee la
   cookie HttpOnly `jwt`; en cliente lee `sessionStorage.jwt`. No intentes leer
   cookies desde el navegador con `js-cookie` para el token (es HttpOnly); el
   token "visible" solo vive en `sessionStorage`.

6. **Cambios de autenticación afectan dos flujos** — portal psi (`login.tsx`) y
   admin (`admin-access.tsx`). Ambos usan `syncJwtCookie` (server action) tras
   login. Si tocas el flujo admin, revisa que el middleware Go
   (`ProtectedAdmin404`) siga recibiendo la cookie en las server actions.

7. **SSR vs Browser URLs** — Dentro de server actions, la API se llama por
   `API_URL_INTERNAL` (red Docker). En el navegador por `VITE_API_URL`. Nunca
   uses una en el otro contexto.

8. **El directorio solo muestra áreas del catálogo** — `PsychologistCard.tsx`
   pinta chips de `psychologist.specialties` tal cual llegan del API, y la API
   Go los resuelve SOLO desde el catálogo (`psi_specialty_models`, ver gotcha 11
   de `api/AGENTS.md`). Si el catálogo está vacío, las tarjetas salen sin chips:
   NO es un bug del front. La tarjeta usa `line-clamp-3` en el mini_bio para
   mantener alturas uniformes.

9. **`ContactCard` y el detalle `/directorio/[slug]`** — el detalle usa layout de
   2 columnas: `ProfileHeader` + `ContactCard` a la derecha (encima de "Perfil").
   La `ContactCard` apila las ubicaciones UNA debajo de la otra y marca cada una
   con un badge de tipo: `Carabobo` (azul), `Venezuela` (índigo) o
   `Internacional` (verde). Si añades una ubicación nueva, mantenlo en la lista
   vertical y dale su `tag`; no vuelvas al grid.

## Estructura

```
src/
├── routes/          → Rutas (FileRoutes). snake_case. `admin/`, `psi/`, `public/`
├── components/      → UI por dominio: `admin/`, `directory/`, `psi/`, `layaout/`, `ui/`
├── lib/             → Capa compartida isomórfica (api, auth, bucket, sanitize, errors)
├── lib/actions/     → Server actions de auth ("use server")
├── types/           → Tipos compartidos (psi, admin, auth, post)
├── entry-server.tsx → HTML base + CSP (incluye origen del bucket)
├── entry-client.tsx → Hydrate CSR
└── app.tsx          → Raíz (Router, AuthProvider, Navbar, ErrorBoundary)
```

## Librerías clave (`src/lib/`)

| Archivo | Responsabilidad |
|---------|-----------------|
| `api.ts` | `fetchApi`/`apiGet`/`apiPost`/`apiPatch`/`apiDelete`, token isomórfico, `ApiError` |
| `auth.tsx` | `AuthProvider`/`useAuth` (estado de sesión cliente, timer de expiración) |
| `actions/auth.ts` | Server actions de login/logout (cookie HttpOnly) |
| `bucket.ts` | `bucketUrl()` / `siteUrl()` — único punto para URLs de bucket/sitio |
| `sanitize-html.ts` | `sanitizeHtml()` — HTML seguro para innerHTML (DOMPurify + fallback SSR) |
| `sanitizer.ts` | Sanitización de inputs (teléfono, texto, email, contraseña) |
| `errors.ts` | `getUserFacingError()` — mensajes amigables, sin detalle interno |
| `utils.ts` | Slugs (`createProfileSlug`), helpers de dominio |

## Convenciones de código

- snake_case para archivos/rutas (español).
- No agregar comentarios salvo gotchas o decisiones (los headers de `src/lib/*`
  son el ejemplo).
- Preservar funcionamiento: cambios seguros e incrementales.
- Un solo commit por fix en la rama `docs`.
- Antes de entregar: `npm run build` debe compilar.

## Flujo de verificación rápida

1. `npm run build` — ¿compila?
2. Con Docker: `docker compose build web && docker compose up -d`.
3. `curl localhost:23000` → 200, y el CSP debe incluir el origen del bucket en
   `img-src` (verificable en el HTML).
4. Login admin (`/admin-access`) → el PATCH del dashboard deja de dar 404 si la
   cookie HttpOnly se está seteando (`curl` del RPC de `syncJwtCookie`).

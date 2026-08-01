# Frontend — SolidStart

> **[⬆ Raíz](../)** — `web/`

Sitio web del Colegio de Psicólogos del Estado Carabobo: páginas públicas,
portal del psicólogo, administración e inscripciones.

## Stack

| Área | Tecnología |
|------|-----------|
| Framework | SolidStart (SolidJS 1.x) + TypeScript |
| Renderizado | SSR + CSR (hydrate) |
| Ruteo | Basado en archivos (`src/routes/`) con `@solidjs/router` |
| Build/Servidor | Vinxi / Nitro, preset `deno-server` (`app.config.ts`) |
| CSS | Tailwind CSS v4 (`@tailwindcss/vite`) |
| Editor RTF | TipTap v3 (contenido de noticias/perfiles) |
| Auth | JWT en cookie HttpOnly (server actions) + copia en `sessionStorage` |
| Bucket | MinIO/S3, URLs centralizadas vía `src/lib/bucket.ts` |

## Requisitos

- **Node.js ≥ 22** (para Vinxi/Vite)
- **Deno ≥ 2.x** (para build SSR y servidor de producción)
- **API Go corriendo** en `http://localhost:28080` (ver [`api/README.md`](../api/README.md))
- **MinIO/S3** con el bucket accesible desde el navegador (`http://localhost:29000/colpsi-bucket` en dev)

## Setup local

```bash
# 1. Copiar variables de entorno
cp .env.example .env   # ajustar VITE_API_URL, VITE_BUCKET_URL, VITE_SITE_URL

# 2. Instalar dependencias
npm install

# 3. Desarrollo (hot reload)
npm run dev
```

## Scripts

```bash
npm run dev          # Desarrollo con hot reload (Vinxi)
npm run build        # Build de producción SSR → .output/
npm run start        # Sirve el build (Vinxi)

# Equivalentes con Deno (también los usa el dockerfile):
deno task dev
deno task build
deno task start      # Ejecuta .output/server/index.mjs con permisos granulares
```

> No hay typecheck configurado en el proyecto (TypeScript no está instalado);
> la verificación de tipos/errores de build es `npm run build`.

## Variables de entorno

| Variable | Uso | SSR/Browser |
|----------|-----|-------------|
| `VITE_API_URL` | URL base de la API para el navegador | Browser |
| `API_URL_INTERNAL` | URL base de la API desde el servidor (red Docker, opcional) | SSR |
| `VITE_BUCKET_URL` | URL pública del bucket en formato Path-Style (`{endpoint}/{bucket}`) | Ambos |
| `VITE_SITE_URL` | URL canónica del sitio (QR, robots.txt, sitemap) | Ambos |
| `PORT` | Puerto del servidor (default `3000`) | SSR |

> ⚠️ Las variables `VITE_*` se **inlinean en el bundle** en tiempo de build.
> Si cambias `VITE_BUCKET_URL` o `VITE_API_URL` y despliegas con Docker,
> debes reconstruir la imagen (`docker compose build web`). Ver `.env.example`.

## Arquitectura

### Ruteo y layout

```
src/routes/
├── index.tsx               → Inicio
├── directorio/             → Directorio público de psicólogos
├── noticias/               → Noticias
├── explorar.tsx, nosotros.tsx, inscripcion.tsx
├── login.tsx, admin-access.tsx
├── admin/                  → CRUD psicólogos, noticias, áreas, staff, dashboard
├── psi/                    → Portal del psicólogo (perfil, académico)
├── public/                 → Rutas de la vista pública (alias)
├── robots.txt.ts, sitemap.xml.ts
└── [...404].tsx
```

`src/app.tsx` es la raíz: `MetaProvider` → `Router` (FileRoutes) → `AuthProvider` →
`Navbar` + `ErrorBoundary` (fallback `OfflineAlert`).

### Autenticación (doble capa)

1. **Server action** (`src/lib/actions/auth.ts`, `syncJwtCookie` en `login.tsx`/`admin-access.tsx`):
   autentica contra Go, setea el JWT como cookie **HttpOnly** (`jwt`, `SameSite=Strict`).
   El browser nunca ve el token real.
2. **Estado cliente** (`src/lib/auth.tsx`): guarda una copia en `sessionStorage` + datos del
   usuario en la cookie `user_data` (para restaurar sesión y el timer de expiración).

`src/lib/api.ts` (`fetchApi`) resuelve el token según contexto:

- **SSR/Server actions** → lee la cookie `jwt` del request, usa `API_URL_INTERNAL`.
- **Navegador** → lee `sessionStorage.jwt`, usa `VITE_API_URL`.

Los errores de red se convierten en `ApiError(503, "OFFLINE_SERVICE")` y los de API en
`ApiError(status, message)`.

> ⚠️ Server actions de `@solidjs/router` deben invocarse con `useAction()`, **nunca**
> llamando la acción directamente (`this.r.singleFlight` lanza TypeError).

### Imágenes (bucket S3/MinIO)

- **Única vía**: `bucketUrl(key)` de `src/lib/bucket.ts`. Es idempotente: si la key ya es
  una URL absoluta (el API puede devolverla), la retorna tal cual.
- La API Go devuelve URLs públicas usando `S3_PUBLIC_URL` (default `http://localhost:29000`),
  de modo que el navegador nunca recibe el host interno Docker (`s3:9000`).
- El CSP (`src/entry-server.tsx`) agrega el origen del bucket a `img-src` automáticamente
  a partir de `VITE_BUCKET_URL`.

### Sanitización (XSS)

- **HTML**: `sanitizeHtml()` en `src/lib/sanitize-html.ts` — DOMPurify en cliente con lista
  blanca de tags/atributos; fallback regex en SSR (Deno sin DOM). Barrera definitiva: API Go
  (bluemonday) al persistir.
- **Inputs de formulario**: `src/lib/sanitizer.ts` (teléfonos, texto, email, contraseña).
- **Errores**: `src/lib/errors.ts` traduce errores internos a mensajes amigables.

## Seguridad aplicada

| Medida | Dónde |
|--------|-------|
| CSP por meta tag (`img-src` incluye bucket) | `src/entry-server.tsx` |
| JWT HttpOnly + `SameSite=Strict` | `src/lib/actions/auth.ts` |
| Sanitización HTML (DOMPurify/SSR) | `src/lib/sanitize-html.ts` |
| Errores genéricos al usuario | `src/lib/errors.ts` |
| `rel="noopener noreferrer"` en enlaces externos | componentes psi |
| `encodeURIComponent` en búsquedas | `src/routes/directorio/index.tsx` |
| Validación MIME + tamaño de XLSX | `ImportXlsxModal.tsx` |
| Idempotency key en creación | `areas/crear/index.tsx` |
| robots.txt dinámico | `src/routes/robots.txt.ts` |

## Docker

```bash
docker compose build web && docker compose up -d
```

- `dockerfile`: build multi-stage Deno (Debian builder → Alpine final), preset
  `deno-server`, servidor final con permisos granulares (`--allow-net`,
  `--allow-read=.output,public`, `--allow-env`).
- `docker-compose.yml`: conecta a la red externa `api_colpsi_network` y setea
  `API_URL_INTERNAL=http://colpsi_api:8080/api/v1` para las server actions.

## Convenciones

- Archivos y rutas en **snake_case** (minúsculas, guiones bajos) para español.
- Un solo commit por fix en la rama `docs`.
- Preservar el funcionamiento existente ("no modifique el funcionamiento").
- No agregar comentarios salvo que documenten decisión o gotcha (ver `src/lib/*`).

## Ver también

- [`AGENTS.md`](./AGENTS.md) — guía operativa para agentes AI (gotchas críticos).
- [`AGENTS.md` raíz](../AGENTS.md) — contexto general del proyecto.
- [`api/README.md`](../api/README.md) — documentación de la API Go.

**[⬆ Volver a raíz](../)**

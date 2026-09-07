# Despliegue — ColPsiCarabobo

Guía de pasos para desplegar la aplicación completa en el servidor de
producción (VPS Contabo, 4 vCPU / 8 GB).

## Arquitectura desplegada

Dos stacks de Docker Compose independientes que comparten la red Docker
`api_colpsi_network`:

| Stack | Compose | Contenido |
|-------|---------|-----------|
| API   | `api/docker-compose.yml` | Postgres 18, PgBouncer, migrador (Atlas), API Go (Fiber), MinIO, edge cache nginx, Valkey, Audiobookshelf, pgAdmin y MailHog (dev) |
| Web   | `web/docker-compose.yml` | Frontend SolidStart SSR (Deno), conecta a la red externa `api_colpsi_network` |

> El stack de web **requiere** que exista la red externa `api_colpsi_network`,
> que se crea automáticamente al levantar el stack de API. Levanta primero la API.

### Puertos publicados (host)

| Puerto | Servicio |
|--------|----------|
| 23000  | Frontend (web) |
| 28080  | API Go |
| 29000  | Edge cache nginx (imágenes públicas / bucket) |
| 29001  | Consola MinIO |
| 29002  | MinIO directo para SDK (dev local) |
| 5432   | PostgreSQL |
| 6432   | PgBouncer |

> Los puertos de DB/PgBouncer los fija `POSTGRES_HOST_PORT` / `PGBOUNCER_HOST_PORT`
> en el `api/.env` (este repo usa `5432`/`6432`). El default del compose es
> `25432`/`26432`, pero si el `.env` los define, el effective es el del `.env`.
| 26379  | Valkey |
| 21337  | Audiobookshelf |
| 25050  | pgAdmin (solo dev) |
| 21025 / 28025 | MailHog SMTP / UI (solo dev) |

## Requisitos previos en el servidor

- Docker Engine + Docker Compose v2
- Acceso al repo (SSH key): `git@github.com:veniversvm/ColPsiCarabobo.git`
- Certs TLS y reverse proxy (nginx/Caddy) delante de `23000` y `28080`, con
  `X-Forwarded-Proto: https` — **necesario** para que el header HSTS se emita
  (ver sección "Security headers").

## Paso 0 — Verificación local (antes de desplegar)

```bash
# API
cd api
go build ./... && go vet ./...
make test-all        # requiere el test DB (docker compose -f docker-compose.test.yml up -d db_test)

# Web
cd web
npm run build        # comprueba que el SSR compila
```

## Paso 1 — Obtener el código en el servidor

```bash
cd /ruta/al/deploy
git pull origin main
```

## Paso 2 — Desplegar la API

```bash
cd api

# 1. Configurar entorno (solo la primera vez)
cp .env.example .env
nano .env            # ajustar valores, ver checklist de producción abajo

# 2. Construir y levantar
docker compose up -d --build
```

El contenedor `migrador` aplica las migraciones de Atlas automáticamente al
arrancar (`docker compose up` espera a que termine antes de iniciar la API).

### Verificar la API

```bash
curl -s http://localhost:28080/live            # → {"status":"ok",...}
curl -s http://localhost:28080/ready           # → 200 cuando DB responde
curl -s http://localhost:28080/api/v1/psi/directory | head
# Las URLs de imágenes deben salir con el host público (no s3:9000)
```

## Paso 3 — Desplegar el frontend

```bash
cd web

# 1. Configurar entorno (solo la primera vez)
cp .env.example .env
nano .env
```

> ⚠️ `VITE_API_URL`, `VITE_BUCKET_URL` y `VITE_SITE_URL` se **inlinean en el
> bundle en tiempo de build**. Cambiarlas exige reconstruir la imagen.

```bash
# 2. Construir y levantar (reconstruye SIEMPRE si cambiaste .env o código)
docker compose build web
docker compose up -d
```

### Verificar el frontend

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:23000   # → 200
# El HTML debe incluir el CSP con el origen del bucket en img-src
curl -s http://localhost:23000 | grep -o 'Content-Security-Policy[^<]*' | head -1
```

## Paso 4 — Checklist de producción (variables de entorno)

### `api/.env`

| Variable | Valor esperado en prod |
|----------|------------------------|
| `APP_ENV` | `production` (desactiva Swagger y debug-monitor; seed admin con pass aleatoria) |
| `DB_HOST` / `DB_PORT` | `pgbouncer` / `5432` (lo fuerza el compose) |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | credenciales reales |
| `ADMIN_PASSWORD` | definirla para fijar la pass del super admin |
| `ALLOWED_ORIGINS` | dominio público del frontend (`https://www.tusitio.com`) |
| `S3_ENDPOINT` | `http://s3:9000` (interno, lo fuerza el compose) |
| `S3_PUBLIC_URL` | URL **pública** accesible desde el navegador p.ej. `https://www.tusitio.com` (si el edge cache está tras el dominio) o `http://localhost:29000` en dev |
| `JWT_LIBRARY_SECRET` | secreto largo y estable (`openssl rand -hex 32`) |
| `HSTS_MAX_AGE` | `31536000` (1 año, solo se emite sobre HTTPS) |
| `HSTS_PRELOAD` | `false` (activar solo tras registrarse en hstspreload.org) |
| `ABS_BASE_URL` | `http://audiobookshelf:80` (interno, lo fuerza el compose) |
| `ABS_PUBLIC_URL` | URL pública de la biblioteca virtual |
| `ABS_ADMIN_USERNAME/PASSWORD` | admin de aprovisionamiento ABS (en prod `root`) |
| `ABS_PASSWORD_SECRET` | `openssl rand -hex 24`, mantener privado |
| `ABS_SYNC_INTERVAL_HOURS` | `24` |
| `VALKEY_ADDR` | `valkey:6379` (lo fuerza el compose) |

### `web/.env`

| Variable | Valor esperado en prod |
|----------|------------------------|
| `VITE_API_URL` | `https://api.tusitio.com/api/v1` (URL pública para el navegador) |
| `VITE_BUCKET_URL` | `https://imagenes.tusitio.com/colpsi-bucket` |
| `VITE_SITE_URL` | `https://www.tusitio.com` |

> En el compose de web se fuerza `VITE_API_URL` y `API_URL_INTERNAL` a
> `http://colpsi_api:8080/api/v1` para la red interna (server actions SSR).

## Paso 5 — Comandos operativos

```bash
# Ver estado de los stacks
docker compose -f api/docker-compose.yml ps
docker compose -f web/docker-compose.yml ps

# Ver logs
docker compose -f api/docker-compose.yml logs -f api
docker compose -f web/docker-compose.yml logs -f web

# Redesplegar tras cambios de código
git pull origin main
docker compose -f api/docker-compose.yml up -d --build
docker compose -f web/docker-compose.yml build web && docker compose -f web/docker-compose.yml up -d

# Aplicar migraciones a mano (si el migrador no corrió)
docker compose -f api/docker-compose.yml run --rm migrador

# Reiniciar todo limpio (conserva volúmenes)
docker compose -f api/docker-compose.yml up -d --force-recreate
```

## Rollback

```bash
git log --oneline -5          # identificar el commit anterior bueno
git checkout <commit-anterior> -- api/ web/   # o reset en rama propia
docker compose -f api/docker-compose.yml up -d --build
docker compose -f web/docker-compose.yml build web && docker compose -f web/docker-compose.yml up -d
```

## Notas de seguridad (headers)

- **HSTS** solo se emite sobre HTTPS: el reverse proxy debe enviar
  `X-Forwarded-Proto: https` a la API, o el header no aparecerá.
- **CSP** del frontend se genera en `src/entry-server.tsx` e incluye
  automáticamente el origen del bucket en `img-src`.
- **Swagger** (`/swagger/`) y `/debug-monitor` solo existen con `APP_ENV=development`:
  en producción no se despliegan.

## Ver también

- [`api/README.md`](./api/README.md) — documentación de la API y servicios.
- [`web/README.md`](./web/README.md) — build, scripts y arquitectura del front.
- [`AGENTS.md`](./AGENTS.md) — convenciones y seguridad aplicada.

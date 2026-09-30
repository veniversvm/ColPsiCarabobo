# Despliegue — ColPsiCarabobo

Cómo está desplegado hoy el Colegio de Psicólogos de Carabobo en el server de
producción y cómo reproducirlo, actualizarlo y recuperarlo.

Todo lo que aparece aquí se verificó **contra el server en producción**
(2026-09-30). Los comandos marcados con ✅ se ejecutaron y devolvieron lo que
se indica. Este doc **no contiene secretos**: solo los nombres de variable y
el comando que genera cada valor.

---

## 1. Arquitectura desplegada

**Dos proyectos de Docker Compose independientes** que comparten la red externa
`api_colpsi_network`:

| Proyecto | Compose | Contiene |
|----------|---------|----------|
| API | `api/docker-compose.yml` | Postgres, PgBouncer, migrador Atlas, API Go (Fiber), MinIO, edge cache nginx, Valkey, Audiobookshelf, `abs-init`, `createbuckets`, pgAdmin y MailHog (solo dev) |
| Web | `web/docker-compose.yml` | Frontend SolidStart SSR sobre Deno |

> ⚠️ **`docker compose up -d` desde `api/` NO levanta el web.** Son dos
> proyectos distintos con nombres distintos. Si borras el proyecto web y solo
> ejecutas el `up` del stack de la API, el frontend se queda caído y el sitio
> devuelve 502. Hay que levantar los dos por separado.

El web depende de que la red externa ya exista, y la red la crea el stack de
la API al levantarse. **Orden obligatorio: API primero, web después.**

### 1.1 Servicios del stack de la API y su función

| Servicio | Función | Puerto interno |
|----------|---------|----------------|
| `colpsi_db` | PostgreSQL 18.6, datos reales | 5432 |
| `colpsi_pgbouncer` | Pool de conexiones (transaction mode) frente a Postgres | 5432 |
| `colpsi_migrador` | Aplica las 11 migraciones de Atlas y termina (`run once`) | — |
| `colpsi_api` | API Go — el backend que sirve el sitio | 8080 |
| `colpsi_s3` | MinIO, almacenamiento de archivos (bucket `colpsi-bucket`) | 9000 / 9001 |
| `colpsi_imgcache` | Edge cache nginx delante de MinIO — sirve las imágenes públicas | 80 |
| `colpsi_valkey` | Store del rate limiting (60 req/min por IP) | 6379 |
| `audiobookshelf` | Biblioteca virtual (ABS 2.37.1) | 80 |
| `colpsi_abs_init` | Crea el usuario `root` de ABS si no existe (`run once`) | — |
| `api-createbuckets-1` | Crea el bucket y lo hace público (`run once`) | — |
| `colpsi_pgadmin` | pgAdmin — **solo con `--profile dev`** | 80 |
| `colpsi_mailhog` | Mailpit/MailHog — **solo con `MAILPIT_PROFILE=dev`** | 1025 / 8025 |

Los tres `run once` (`migrador`, `abs_init`, `createbuckets`) terminan con
`Exited (0)` y es **lo correcto**: no son fallos.

`abs-init` es **bloqueante** para la API: `api` lo espera con
`service_completed_successfully`, porque `ABS_ADMIN_TOKEN` es un gate duro de
arranque (`cmd/api/main.go:75` → `log.Fatal`).

### 1.2 Volúmenes

| Volumen | Contenido |
|---------|-----------|
| `api_postgres_data` | El datadir de Postgres — **lo único irreemplazable** |
| `api_minio_data` | Objetos de MinIO (`.minio.sys` incluido: usuarios, políticas) |
| `api_valkey_data` | Rate limiting (volátil, se puede perder) |
| `api_imgcache_data` | Caché de imágenes (volátil, se puede perder) |

Además hay **bind mounts** en `api/biblioteca/`: `config/` (la base SQLite de
ABS), `books/` y `audiobooks/` (los PDFs). Son datos de la biblioteca y no
viven en un volumen: copiarlos es un `cp -a` del directorio.

---

## 2. Capas de exposición

Solo hay **una** superficie pública: el puerto 443 de nginx.

```
Internet ──443──▶ nginx ──┬── /               ──▶ 127.0.0.1:23000  web (SSR)
                         ├── /api/           ──▶ 127.0.0.1:28080  API Go
                         └── /colpsi-bucket/ ──▶ 127.0.0.1:29000  imgcache → MinIO

Internet ──443──▶ nginx ──/  (subdominio) ──▶ 127.0.0.1:21337  Audiobookshelf
```

Todo lo demás — Postgres, PgBouncer, Valkey, la consola de MinIO, el SDK de
MinIO y el propio SSR del frontend — está **atado a `127.0.0.1`** y no es
alcanzable desde la red. Un contenedor que necesite hablar con la API lo hace
por la red interna de Docker (`colpsi_api:8080`), no por el puerto publicado.

> ⚠️ **No reviertas el prefijo `127.0.0.1:` de los `ports:`.** Docker publica
> en `0.0.0.0` cuando no se indica IP de host. Con `0.0.0.0` una base de datos,
> una consola de MinIO y la API completa quedan accesibles por la IP del
> servidor, saltándose TLS y los headers de nginx. firewalld los tapa hoy, pero
> eso es una regla de firewall y no del contenedor: cambiar de zona o desactivar
> firewalld los publica.

---

## 3. Puertos

Todos en `127.0.0.1`. La columna "consumido por" indica quién los necesita
abiertos; si está vacía, solo se usan desde dentro de la red de Docker.

| Puerto | Servicio | Consumido por |
|--------|----------|---------------|
| 23000 | Frontend SSR | nginx (`location /`) |
| 28080 | API Go | nginx (`location /api/`), healthchecks |
| 29000 | imgcache (imágenes) | nginx (`location /colpsi-bucket/`) |
| 21337 | Audiobookshelf | nginx (subdominio `abs.`) |
| 25432 | PostgreSQL | pgAdmin / psql desde el host |
| 26432 | PgBouncer | cliente externo del host |
| 26379 | Valkey | inspección desde el host |
| 29001 | Consola MinIO | navegador desde el host |
| 29002 | MinIO directo (SDK) | `S3_ENDPOINT` en dev local |
| 25050 | pgAdmin | — (solo dev) |
| 21025 / 28025 | MailHog SMTP / UI | — (solo dev) |

✅ Verificado: los 9 puertos de producción responden en `127.0.0.1` y dan
`No route to host` desde fuera del server.

**Los puertos de DB y PgBouncer los fija el `.env`** vía
`POSTGRES_HOST_PORT` / `PGBOUNCER_HOST_PORT`. El default del compose es
`25432`/`26432` (producción); el `.env` de desarrollo los baja a
`5432`/`6432`. El valor efectivo es el del `.env` si está definido.

> ⚠️ La API **nunca** habla con MinIO por el puerto publicado. `S3_ENDPOINT`
> apunta a `http://s3:9000` (interior). Pasarlo por el edge cache rompe las
> firmas S3 v4 y da 403. Ver gotcha 2 de `api/AGENTS.md`.

---

## 4. Prerrequisitos del server

Valores medidos en el server de producción:

| Componente | Versión |
|------------|---------|
| OS | AlmaLinux 10.2 |
| Docker Engine | 29.8.1 |
| Docker Compose | 5.5 |
| nginx | 1.26.3 |
| PostgreSQL | 18.6 (contenedor) |
| SELinux | **Enforcing** |
| Ruta del repo | `/root/ColPsiCarabobo` |

Además:

- **firewalld** activo, con `http`, `https` y `ssh` permitidos en la zona
  `public`. Los puertos de loopback no necesitan reglas.
- **certbot** con `certbot-renew.timer` habilitado y certificados para los dos
  dominios (vencen el **2026-12-29**).

> ⚠️ **SELinux Enforcing bloquea el reverse proxy por defecto.** nginx corre en
> el dominio `httpd_t` y sin `httpd_can_network_connect` el kernel le niega el
> `connect()` a cualquier puerto, con `errno 13`. El síntoma es **502 en todos
> los upstreams a la vez**, incluso en los que responden bien por `curl` — lo
> confunde con "el servicio está caído" cuando en realidad es el proxy. El
> error en `/var/log/nginx/*.error.log` lo delata:
>
> ```
> connect() to 127.0.0.1:28080 failed (13: Permission denied) while connecting to upstream
> ```
>
> `Connection refused` (errno 111) sería "nadie escucha"; `13` es "no te dejo".
> Verificado en este server.

---

## 5. nginx y TLS

Configuración en `/etc/nginx/conf.d/`:

| Archivo | Sirve |
|---------|-------|
| `colegio-psicologos-carabobo.conf` | Dominio principal, `listen 443 ssl` + `listen 80` (redirect 301) |
| `abs.colegio-psicologos-carabobo.conf` | Subdominio de la biblioteca, mismo esquema |

### 5.1 Bloques `location` del dominio principal

| Location | Upstream | Qué sirve |
|----------|----------|-----------|
| `/` | `127.0.0.1:23000` | Frontend SSR (SolidStart) |
| `/api/` | `127.0.0.1:28080` | API Go |
| `/colpsi-bucket/` | `127.0.0.1:29000/colpsi-bucket/` | Imágenes del bucket vía edge cache |

Los tres mandan `Host`, `X-Real-IP`, `X-Forwarded-For` y `X-Forwarded-Proto`.

> `X-Forwarded-Proto: $scheme` es **necesario**: sin él la API no detecta que
> la petición vino por HTTPS y el header HSTS no se emite.

### 5.2 Comprobaciones tras tocar nginx

```bash
nginx -t                          # sintaxis
systemctl reload nginx            # recarga sin cortar conexiones
```

> ⚠️ Un error tipográfico muy fácil: `systemctl enable --now nginx.` (con punto
> final) falla con `Unit nginx..service does not exist` y **no** significa que
> falte nginx.

---

## 6. Estado actual de la instalación

Medido el 2026-09-30. **La instalación de producción está vacía** — es un
despliegue limpio, sin datos cargados todavía:

| Elemento | Cantidad |
|----------|----------|
| Tablas en el esquema | 42 |
| `psi_users` (agremiados) | **0** |
| `posts` (noticias) | **0** |
| `psi_inscription_requests` | **0** |
| `user_admins` | 1 (el super admin) |
| `psi_specialty_models` (áreas) | 7 |
| `api_change_logs` | 1 |
| Objetos en el bucket `colpsi-bucket` | **0** |
| PDFs en `biblioteca/books/` | **0** |
| Tamaño de la base | 10 MB |

El directorio público, las noticias y el buscador de áreas responden `200` pero
con listas vacías — **no es un fallo**. La biblioteca virtual está instalada y
operativa, pero sin catálogo: el primer agremiado solvente que entre dispara el
worker de sync y se le crea su cuenta en ABS.

> Esto importa para el backup: hoy lo único que vale la pena proteger es el
> volumen de Postgres (10 MB). La sección de backup está preparada para cuando
> haya contenido.

---

## 7. Bootstrap de un server nuevo

### 7.1 Código

```bash
cd /root/ColPsiCarabobo
git pull origin main
```

### 7.2 Secretos — se generan, no se copian

> ⚠️ **`api/.env.production.example` está en `.gitignore` a propósito.** Existe
> en el disco del server con los secretos **reales** de producción, no
> placeholders. Por eso no se commitea. Para un server nuevo, **genera los
> valores**, no copies ese archivo.

```bash
cd api
cp .env.example .env          # 44 variables, todas con valor por defecto
```

Genera cada secreto por separado (no reutilices ninguno entre entornos):

```bash
openssl rand -hex 32   # DB_PASSWORD, JWT_LIBRARY_SECRET, ANALYTICS_IP_SALT
openssl rand -hex 24   # ABS_PASSWORD_SECRET
openssl rand -base64 32   # AWS_SECRET_ACCESS_KEY
```

| Variable | Cómo se genera | ⚠️ No rotar sin más |
|----------|----------------|---------------------|
| `DB_PASSWORD` | `openssl rand -hex 32` | Postgres la tiene hasheada: rotarla sin cambiarla en la BD deja la API sin conectar |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | `openssl rand` | MinIO las persiste en `format.json` dentro del volumen: rotarlas sin reescribirlo deja el bucket inaccesible |
| `ADMIN_PASSWORD` | `openssl rand` | El seed solo corre si no hay admins; con admins existentes no hace nada |
| `JWT_LIBRARY_SECRET` | `openssl rand -hex 32` | Firma los tokens de la biblioteca |
| `ABS_PASSWORD_SECRET` | `openssl rand -hex 24` | Deriva la contraseña de cada agremiado: al rotarla, las cuentas existentes dejan de validar |
| `ANALYTICS_IP_SALT` | `openssl rand -hex 32` | Debe ser **estable entre reinicios**: si cambia, la misma IP genera otra huella y se rompen los conteos de visitantes únicos |
| `RESEND_API_KEY` | panel de resend.com | — |

### 7.3 Valores fijos en producción

| Variable | Valor |
|----------|-------|
| `APP_ENV` | `production` — desactiva Swagger y el debug-monitor |
| `DB_HOST` / `DB_PORT` | `pgbouncer` / `5432` (lo fuerza el compose) |
| `DB_NAME` / `DB_USER` | `colpsi_db` / `postgres` |
| `S3_ENDPOINT` | `http://s3:9000` (interno, lo fuerza el compose) |
| `S3_PUBLIC_URL` | `https://colegio-psicologos-carabobo.com/colpsi-bucket` |
| `ABS_BASE_URL` | `http://audiobookshelf:80` (interno, lo fuerza el compose) |
| `ABS_PUBLIC_URL` | `https://abs.colegio-psicologos-carabobo.com` |
| `VALKEY_ADDR` | `valkey:6379` (lo fuerza el compose) |
| `ALLOWED_ORIGINS` | `https://colegio-psicologos-carabobo.com` |
| `ABS_SYNC_INTERVAL_HOURS` | `24` |
| `HSTS_MAX_AGE` | `31536000` |
| `HSTS_PRELOAD` | `false` (activar solo tras registrarse en hstspreload.org) |

### 7.4 `web/.env`

| Variable | Valor |
|----------|-------|
| `VITE_API_URL` | `https://colegio-psicologos-carabobo.com/api/v1` |
| `VITE_BUCKET_URL` | `https://colegio-psicologos-carabobo.com/colpsi-bucket` |
| `VITE_SITE_URL` | `https://colegio-psicologos-carabobo.com` |

---

## 8. Despliegue de la API

```bash
cd /root/ColPsiCarabobo/api
docker compose up -d --build
```

✅ El orden de arranque lo resuelve el `depends_on` y se respeta: Postgres
healthy → PgBouncer → migrador y `abs-init` completan → API. Tarda ~40 s.

### 8.1 Arrancar el SELinux para nginx

```bash
setsebool -P httpd_can_network_connect on    # -P = permanente
systemctl enable --now nginx
```

### 8.2 Verificar

```bash
cd /root/ColPsiCarabobo/api
docker compose ps                              # colpsi_api debe decir (healthy)
docker compose logs api | grep -iE '"level":"(fatal|error)"|panic'
curl -s http://127.0.0.1:28080/live
curl -s http://127.0.0.1:28080/api/v1/psi/directory | head -c 200
```

Los tres `run once` deben terminar en `Exited (0)`:

```bash
docker compose ps -a | grep -E 'migrador|abs_init|createbuckets'
```

> Si `colpsi_abs_init` sale con `Exited (1)`, es que `ABS_ADMIN_PASSWORD` está
> vacía en el `.env` y ABS no tiene root todavía. Es deliberado: frena el
> arranque en vez de crear un root sin contraseña.

---

## 9. Despliegue del web

```bash
cd /root/ColPsiCarabobo/web
docker compose build web      # los VITE_* se hornean aquí
docker compose up -d
```

> ⚠️ **`VITE_*` se inlinean en el bundle en tiempo de build.** Cambiarlos en el
> `.env` no tiene efecto hasta reconstruir la imagen. Si no cambiaste código ni
> `.env`, `docker compose up -d` basta (recrea el contenedor, no la imagen).

### 9.1 Las dos URLs, y por qué confundirlas rompe el sitio

Es la diferencia entre "el sitio se ve bien" y "el sitio funciona":

| Variable | Cuándo se lee | Para qué | De dónde sale |
|----------|----------------|----------|---------------|
| `VITE_API_URL` | **En build** (`import.meta.env`) | Lo consume el **navegador** | `web/.env` (URL pública) |
| `API_URL_INTERNAL` | **En runtime** (`process.env`) | Lo consumen los **server actions** de SSR | `environment:` del compose (red interna) |

`web/src/lib/api.ts` resuelve `API_URL_INTERNAL || VITE_API_URL`, y como el
compose solo define `API_URL_INTERNAL`, la URL pública del `.env` es la que
viaja al bundle.

> ⚠️ **No pongas `VITE_API_URL` en el `environment:` del compose.** Ese bloque
> pisa al `env_file` y el valor acabaría horneado en el bundle del navegador.
> Con `http://colpsi_api:8080/api/v1` el sitio **se ve bien** (el SSR responde
> por `API_URL_INTERNAL`) pero todo fetch desde el cliente falla: la búsqueda
> del directorio, el login, las acciones del panel. El navegador no puede
> resolver un hostname de Docker. Es un fallo que solo aparece cuando el
> usuario interactúa.

### 9.2 Verificar

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:23000    # 200
curl -s http://127.0.0.1:23000 | grep -o 'Content-Security-Policy[^<]*' | head -1
# el bundle del navegador debe traer la URL pública, nunca colpsi_api:
docker exec colpsi_web sh -c 'grep -rl "colpsi_api" /app/.output/public/ || echo "OK: sin URL interna"'
```

---

## 10. Verificación de extremo a extremo

✅ Estado comprobado el 2026-09-30. **Los 11 endpoints públicos devuelven 200:**

```bash
for p in / /directorio /inscripcion /noticias /terminos /admin-access /psi \
         /api/v1/psi/directory /api/v1/specialties /colpsi-bucket/ /robots.txt; do
  printf "%-28s %s\n" "$p" "$(curl -s -o /dev/null -w '%{http_code}' https://colegio-psicologos-carabobo.com$p)"
done
curl -s -o /dev/null -w "%{http_code}\n" https://abs.colegio-psicologos-carabobo.com/
```

### 10.1 ⚠️ Los healthchecks NO se sirven por el dominio

`/live` y `/ready` están registrados en la **raíz** de la API, fuera de
`/api/`, así que nginx no los proxea: caen en `location /` y van al frontend.

| Ruta | Por el dominio | Contra `127.0.0.1:28080` |
|------|----------------|---------------------------|
| `/live` | **200 con el HTML del sitio** | 200 real |
| `/ready` | 404 | 200 real |

✅ Comprobado. **Monitorizar `/live` por el dominio da verde para siempre sin
llegar a tocar la API**: el frontend devuelve su página para rutas
desconocidas. Para healthcheck usa siempre el host:

```bash
curl -s http://127.0.0.1:28080/live
curl -s http://127.0.0.1:28080/ready
```

El healthcheck interno del contenedor sí es correcto: usa
`/api/v1/psi/directory`.

### 10.2 Ruido esperado

- **Headers duplicados.** La API y nginx emiten `X-Frame-Options`, HSTS y
  `Referrer-Policy`, así que llegan dos veces (6 headers duplicados en total).
  Inofensivo: el navegador aplica el más restrictivo.
- **Sin rate limit en nginx.** El límite de 60 req/min por IP vive **solo en la
  API**. nginx no tiene `limit_req` ni `limit_conn`.

---

## 11. Backup y restauración

> ⚠️ **Hoy no hay ningún backup configurado.** Los únicos cron del host son
> `clamav-cron.sh` y `rkhunter-cron.sh` (seguridad del SO), ninguno de datos.
> Si se pierde el volumen de Postgres, se pierde todo.

### 11.1 Qué respaldar

| Qué | Volumen / ruta | Criticidad |
|-----|----------------|------------|
| Base de datos | `api_postgres_data` | **Irreemplazable** |
| Objetos del bucket | `api_minio_data` | Alto (fotos, documentos) |
| Base de ABS | `api/biblioteca/config/absdatabase.sqlite` | Medio (progreso de lectura) |
| PDFs de la biblioteca | `api/biblioteca/books/` | Medio |
| `api_valkey_data`, `api_imgcache_data` | — | **No respaldar**: caché, se regeneran |

### 11.2 Dump de la base de datos

✅ Verificado: genera 100 KB en texto plano, 132 KB en formato custom.

```bash
cd /root/ColPsiCarabobo/api
mkdir -p /root/backups
PW=$(grep '^DB_PASSWORD=' .env | cut -d= -f2-)
docker compose exec -T -e PGPASSWORD="$PW" db \
  pg_dump -U postgres -d colpsi_db -Fc > /root/backups/colpsi_$(date +%F).dump
```

El formato `-Fc` (custom) va comprimido y permite `pg_restore` selectivo; el
texto plano sirve para inspeccionar el dump a ojo.

> El `-T` desactiva el TTY: sin él `docker compose exec` se queda esperando
> entrada. Ver gotcha de la sección 12.

### 11.3 Copia del bucket

✅ Verificado (`mc mirror` funciona; hoy copia 0 objetos porque el bucket está
vacío):

```bash
docker compose exec -T s3 sh -c '
  MC=/opt/bitnami/minio-client/bin/mc
  $MC alias set b http://s3:9000 "$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY" >/dev/null
  $MC mirror --quiet b/colpsi-bucket /tmp/bucket-backup
'
docker cp colpsi_s3:/tmp/bucket-backup /root/backups/bucket_$(date +%F)
```

Usa `mc mirror` a nivel de objeto, no un `tar` del volumen: MinIO escribe
mientras corre y un tar de un volumen vivo no es consistente.

### 11.4 Restaurar

```bash
cd /root/ColPsiCarabobo/api
PW=$(grep '^DB_PASSWORD=' .env | cut -d= -f2-)

# ⚠️ Para restaurar sobre una base con datos hay que recrear el schema.
docker compose stop api
docker compose exec -T -e PGPASSWORD="$PW" db \
  psql -U postgres -d postgres -c "DROP DATABASE colpsi_db;" -c "CREATE DATABASE colpsi_db;"
docker compose exec -T -e PGPASSWORD="$PW" db \
  pg_restore -U postgres -d colpsi_db --no-owner < /root/backups/colpsi_AAAA-MM-DD.dump
docker compose start api
```

> ⚠️ **Verifica el dump antes de necesitarlo.** Un backup no probado no es un
> backup:
>
> ```bash
> pg_restore --list /root/backups/colpsi_AAAA-MM-DD.dump | head
> ```

### 11.5 ⚠️ La copia debe salir del server

Un backup en el mismo disco no protege de la pérdida del disco ni de un
`rm -rf` accidental. Sincroniza `/root/backups/` a un destino externo:

```bash
rsync -avz --delete /root/backups/ usuario@destino:/backups/colpsi/
```

Y pruébalo: una copia que nunca se restauró es una hipótesis.

---

## 12. Gotchas y diagnóstico

### 12.1 Al desplegar

| Síntoma | Causa | Qué hacer |
|---------|-------|-----------|
| **502 en todos los upstreams a la vez** | SELinux: `httpd_can_network_connect` apagado | `setsebool -P httpd_can_network_connect on` + `systemctl reload nginx` |
| **502 solo en `/`** | El web no está corriendo | `cd web && docker compose up -d` — `api/` no lo levanta |
| `Unit nginx..service does not exist` | Punto final en el nombre | `systemctl enable --now nginx` (sin punto) |
| `Unit nginx. service does not exist` tras editar | Punto en el YAML | `listen 80;` sin punto |
| La API no arranca, `log.Fatal` en los logs | `ABS_ADMIN_TOKEN` vacío o inválido | Regenerar (§12.3) |
| El sitio se ve bien pero no busca ni deja iniciar sesión | `VITE_API_URL` horneado con la URL interna | Quitarlo del `environment:` del compose y **reconstruir** (§9.1) |
| 500 `0A000` / `08P01` "cached plan must not change result type" | PgBouncer con planes viejos tras una migración | Reiniciar `colpsi_pgbouncer` |
| 404 en rutas `/admin/*` | JWT que no llega, o key rotada por un re-login | No es que la ruta no exista (enmascarado a 404 por diseño) |
| 429 "Demasiadas solicitudes" en el panel | Rate limit (60/min por IP) agotado | `docker compose up -d valkey` — si Valkey está caído el límite cae a memoria y se resetea en cada reinicio |
| 403 al subir/ver una imagen | `S3_ENDPOINT` apuntando al edge cache | Debe ser `http://s3:9000` |
| `bind: address already in use` | Puerto ocupado en el host | `ss -tlnp` y busca el puerto |

### 12.2 Comandos que engañan

> ⚠️ **`docker compose run` y `docker compose exec` se comen el stdin.** En un
> script por SSH (`ssh 'bash -s' <<'EOF'`) se tragan el resto del script y
> parece que se cortó. Añade `</dev/null`.

> ⚠️ **`awk` no cuenta líneas de fichero.** `NR` cuenta *registros leídos*, no
> el número de línea del fichero. Para citar "línea 137" usa
> `grep -n`.

### 12.3 `ABS_ADMIN_TOKEN`: gate de arranque

`ABS_ADMIN_TOKEN` es un `log.Fatal` en `cmd/api/main.go:75`: si falta o es
inválido, **la API completa no levanta**. Se obtiene con un login a ABS:

```bash
cd /root/ColPsiCarabobo/api
PW=$(grep '^ABS_ADMIN_PASSWORD=' .env | cut -d= -f2-)
TOK=$(curl -s -X POST http://127.0.0.1:21337/login \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"root\",\"password\":\"$PW\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["user"]["token"])')

sed -i '/^ABS_ADMIN_TOKEN=/d' .env     # borra TODAS las duplicadas
echo "ABS_ADMIN_TOKEN=$TOK" >> .env
docker compose up -d api
```

> ⚠️ **Una sola ocurrencia de la variable.** Docker Compose toma el **último**
> valor cuando hay duplicados, así que una línea antigua al principio del
> `.env` manda sobre la correcta del final. El `sed -i` de arriba es
> deliberado por eso.

> El `accessToken` que emite ABS **no caduca**: los `server-settings` de ABS no
> definen `tokenExpiration`. Revocar acceso = desactivar la cuenta.

### 12.4 No tocar

| Thing | Por qué |
|-------|---------|
| `prefer_simple_protocol=true` en el DSN | Quitarlo reintroduce los 500 `0A000`/`08P01` de PgBouncer. Y PgBouncer debe seguir ignorando el parámetro en `IGNORE_STARTUP_PARAMETERS` |
| `user: "0:0"` en el servicio `s3` | Los volúmenes los creó `minio/minio` como root. Bitnami corre como UID 1001 y sin esa línea MinIO muere con `Unable to write to the backend` |
| `image: bitnamilegacy/minio` | `minio/minio` ya no se puede descargar (retiró su repo de Docker Hub) |
| `depends_on: abs-init` de la API | Sin él, la API arranca antes de que exista el root de ABS |

---

## 13. Actualizar el código

```bash
cd /root/ColPsiCarabobo
git pull origin main

cd api    && docker compose up -d --build    # reconstruye y recrea
cd ../web && docker compose build web && docker compose up -d

# siempre al final: confirmar que todo responde
cd /root/ColPsiCarabobo
for p in / /directorio /api/v1/psi/directory /colpsi-bucket/; do
  printf "%-28s %s\n" "$p" "$(curl -s -o /dev/null -w '%{http_code}' https://colegio-psicologos-carabobo.com$p)"
done
```

### 13.1 Rollback

```bash
cd /root/ColPsiCarabobo
git log --oneline -5
git checkout <commit-bueno> -- api/ web/
cd api    && docker compose up -d --build
cd ../web && docker compose build web && docker compose up -d
```

> ⚠️ **El rollback de código no deshace migraciones.** Atlas solo aplica hacia
> adelante. Si el commit bueno espera un esquema anterior, hay que restaurar
> desde un backup (§11.4) o escribir una migración compensatoria. Por eso el
> backup no es opcional.
>
> ⚠️ `git checkout <commit> -- api/ web/` deja el repo en un estado raro
> (working tree de un commit, HEAD de otro). Para volver atrás del todo:
> `git checkout main && git pull`.

---

## 14. Ver también

- [`api/AGENTS.md`](./api/AGENTS.md) — gotchas del backend (21 entradas)
- [`web/AGENTS.md`](./web/AGENTS.md) — gotchas del frontend (14 entradas)
- [`api/AGENTS.md`](./api/AGENTS.md) §14 — `prefer_simple_protocol` y PgBouncer
- [`docs/funcionalidad-app.md`](./docs/funcionalidad-app.md) — qué hace cada módulo
- [`AGENTS.md`](./AGENTS.md) — registro de la seguridad aplicada

# AGENTS.md — API (`api/`)

Guía operativa específica del backend para agentes AI. Complementa el
[`AGENTS.md` raíz](../AGENTS.md); ante conflicto, esta guía es la fuente de verdad
para todo lo que viva dentro de `api/`.

## Contexto rápido

- **Go + Fiber v2**, Clean Architecture: `router → handler → service → repository → DB`.
- PostgreSQL 18 + PgBouncer (transaction mode); migraciones con Atlas; ORM GORM.
- Storage S3/MinIO (`pkg/s3`); emails vía `internal/service/mail_service.go` (worker async).
- Config global en singleton `config.Envs`, cargada con `config.InitConfig()` al arrancar.
- La documentación Swagger se genera con `swag init` y se sirve en `/swagger/`.

## Comandos

```bash
# Tests (ver Makefile)
make test-unit         # unitarios, sin DB (rápidos)
make test-repo         # repositorios (PostgreSQL real, puerto 5433)
make test-integration  # full stack: DB + Fiber + JWT
make test-security     # suite de seguridad E2E
make test-all          # todos en serial (-p 1)
make test-race         # con race detector
make coverage          # reporte de cobertura
make coverage-html     # reporte HTML

# Verificación rápida
go build ./...         # compilar todo
go vet ./...           # análisis estático
go run cmd/api/main.go # arrancar en dev (requiere .env)

# Documentación
swag init -g cmd/api/main.go -o docs/   # regenerar Swagger
```

> Los tests se ejecutan con `-p 1` (serial) para evitar condiciones de carrera
> entre paquetes. Mocks propios (func-override), sin gomock.

## Reglas críticas (gotchas que ya rompieron el proyecto)

1. **`ProtectedAdmin404()` enmascara como 404, no 401** —
   `internal/middleware/auth.go:117`. Cualquier ruta `admin/*` sin JWT válido
   responde `404 {"message": "Cannot <METHOD> <path>"}`. Al depurar un
   "Cannot PATCH ..." del panel admin: el token no está llegando (revisar la
   cookie HttpOnly `jwt` que envían las server actions del frontend), NO es
   que la ruta no exista.
   Las únicas rutas admin con **401 real** son `/session/me` y
   `/session/validate` (grupo `ProtectedAdmin`, ver gotchas 15 y 16); los paths
   `/admin/me` y `/admin/validate` NO existen — se movieron (404 catch-all si
   alguien los llama).

   La clave de firma **es `admin.Key`**, no un secreto global: se resuelve por
   `user_id` en `validateToken`. Un re-login rota la key → los JWT firmados con
   la anterior dejan de validar (404). Al depurar, mira el `SELECT * FROM
   "user_admins" WHERE id = '...'` del log de SQL: si sale `rows:0`, el
   `user_id` del token no existe (un dígito mal transcrito produce exactamente
   el mismo 404 que un token caducado) y el problema está en el cliente, no en
   la ruta ni en la key.
2. **`S3_ENDPOINT` ≠ `S3_PUBLIC_URL`** — son intencionalmente distintos.
   - `S3_ENDPOINT`: interno, para el SDK, debe apuntar SIEMPRE directo a MinIO
     (en Docker: `http://s3:9000`; en dev local: `http://localhost:29002`).
     NUNCA apuntarlo al puerto del host donde vive el edge cache nginx
     (`imgcache`, host `29000`) y las firmas S3 v4 se rompen al pasar por el proxy (403).
   - `S3_PUBLIC_URL`: pública, para URLs que renderiza el navegador
     (en Docker: `http://localhost:29000`, servida por nginx).
   `GetPublicURL()` (`pkg/s3/s3.go`) usa SOLO la pública; nunca cambies esa
   función para usar el endpoint interno. No hardcodear hosts de imágenes.

3. **Barrera final XSS es la API** — todo HTML de usuario (`full_bio`, contenido
   de posts) se sanitiza con `bluemonday.UGCPolicy()` ANTES de persistir
   (`psi_service_self_management.go:223`, `psi_user_admin_service.go:422`,
   `post_service.go`). La sanitización del frontend es solo defensa en profundidad:
   NO confiar en ella, sanitizar siempre en el servicio.

4. **Seed de admin** (`pkg/database/seed.go`) — solo se crea si no hay admins.
   - `development`: usuario `admin` / pass `admin123` (logueada).
   - producción: pass aleatoria de 16 chars, NO hardcodear. La advertencia
     "Cámbiela al iniciar sesión" en logs es esperada.

5. **Nombres reales de env vars** — la tabla en `internal/config/env.config.go`
   manda: `AWS_S3_BUCKET`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`,
   `S3_ENDPOINT`, `S3_PUBLIC_URL`, `APP_ENV`, `VALKEY_ADDR`, `JWT_LIBRARY_SECRET`,
   `ABS_ADMIN_TOKEN`. No usar nombres obsoletos (`AWS_BUCKET`, `AWS_ACCESS_KEY`,
   `GOENV`) — el README anterior los listaba y no existían.
   Para los headers de seguridad: `HSTS_MAX_AGE` (segundos, default 31536000;
   HSTS solo se emite sobre HTTPS) y `HSTS_PRELOAD` (bool, default `false`).
   Para la biblioteca virtual (Audiobookshelf): `ABS_BASE_URL` (interna, en Docker
   `http://audiobookshelf:80`, NUNCA la pública), `ABS_PUBLIC_URL` (navegador),
   `ABS_ADMIN_USERNAME/PASSWORD` (admin de aprovisionamiento; en producción se usa
   `root`, ver gotcha 9) y `ABS_PASSWORD_SECRET` (deriva la clave de cada agremiado;
   `openssl rand -hex 24`). `ABS_SYNC_INTERVAL_HOURS` controla el worker de sync
   (ver gotcha 10).

6. **Config singleton** — `config.InitConfig()` debe ejecutarse al inicio de
   `main()` (`cmd/api/main.go:54`) antes de cualquier otro servicio. `config.Envs`
   es global; no recrearlo por handler.

7. **Worker de email es asíncrono y no bloquea** — `mail_service.go` procesa la
   cola en background con throttling/jittering anti-spam. MailHog solo existe en
   el profile `dev` del compose. Cuando no corre, los logs `mailhog: no such host`
   son **ruido no bloqueante** del worker, no fallos de la petición HTTP.

8. **Idempotencia en creación** — `POST /admin/psi/create` exige la cabecera
   `X-Idempotency-Key` (ventana 30 min). Si un test/script da duplicados o
   "replayed", es el middleware actuando. La respuesta reutilizada marca
   `X-Idempotent-Replayed: true`.

9. **Biblioteca virtual (Audiobookshelf)** — `GET /psi/me/audiobookshelf`
   (`psi_handler.go:GetAudiobookshelfAccess`) devuelve la URL de auto-login
   `{ABS_PUBLIC_URL}/login/?accessToken=...` SOLO a agremiados solventes
   (403 si no). El usuario ABS es el **correo en minúsculas** del agremiado
   (`AbsUsernameFor`); solo cae a `psi_<ci>` si no tiene correo, y esas
   cuentas legacy son precisamente las que el worker de sync desactiva. Se
   crea al vuelo vía la API de admin con la clave derivada por
   `passwordFor` = HMAC-SHA256(`ABS_PASSWORD_SECRET`, `"abs:"+usuario`) hex
   recortado a 32 chars (nunca se expone en claro). El `accessToken` de la URL
   lo emite **ABS**, no la API, y **no caduca**: los `server-settings` de ABS
   no definen `tokenExpiration` (= 0), así que revocar el acceso = desactivar
   la cuenta. Ojo: `POST /psi/login-library` firma un token de 30 días pero
   **el frontend nunca lo llama** (ruta muerta). El id de la cuenta ABS se
   persiste en `audio_book_shell_id` (`UpdateAudioBookShellID`). El acceso se
   sirve por `ABS_PUBLIC_URL`, mientras `ABS_BASE_URL` es la interna del SDK.
   El admin de aprovisionamiento (`ABS_ADMIN_USERNAME/PASSWORD`) debe tener rol
   admin en ABS; si `colpsi-bot` no funciona (401), usar `root` del propio ABS.
   ⚠️ El catálogo contiene **3 títulos, todos PDF** (`biblioteca/books`);
   `biblioteca/audiobooks` está vacía y los ítems llegan `mediaType: book`,
   `ebookFormat: pdf`, `numChapters: 0`, `coverPath: null`. **No hay nada que
   reproducir**: se descarga y se lee. No prometas "audiolibros" ni streaming
   en la documentación (ver `docs/manual-psiuser.typ` §9).

10. **Worker de sync ABS** (`cmd/api/abs_sync_worker.go`) — `runABSSyncLoop`
    se dispara UNA vez al arrancar y luego cada `ABS_SYNC_INTERVAL_HOURS`
    (default 24; `<=0` desactiva la sync). Ejecuta
    `PsiService.SyncAudiobookshelfAccounts`: crea en masa las cuentas `psi_<ci>`
    de los solventes que aún no tienen una y **desactiva** (SIEMPRE, no es
    configurable) las `psi_*` de insolventes/inactivos/soft-deleted. NUNCA toca
    usuarios ABS que no empiecen por `psi_` (protege admin/root). Usa la consulta
    `GetAllForABSSync` (Unscoped) para ver también los soft-deleted.

11. **El directorio y la ficha pública solo muestran áreas del catálogo** —
    `SearchDirectory` (`psi_repository.go`) hace LEFT JOIN a
    `psi_specialty_models`: primero por `primary/secondary_specialty_id` (FK)
    y, como fallback, por coincidencia EXACTA del string legacy
    (`sp1n.name = psi_users.primary_work_area`) — los legacy que no existan en
    el catálogo quedan fuera (COALESCE a vacío): un chip jamás muestra un área
    inventada. Los chips (`Specialties` del `PsiMiniProfileDTO`) solo se
    rellenan para psicólogos **solventes**: la navegación sin texto ya muestra
    solo solventes, y la búsqueda de texto incluye insolventes pero con la
    tarjeta SIN chips. La ficha pública (`GET /psi/:fpv`) resuelve `work_areas`
    con `ResolveSpecialtyNames(ctx, ids, legacy)` — FK primero, luego nombre
    exacto del legacy — y el DTO público NO expone `solvent` (la solvencia es
    invisible al visitante; un insolvente recibe solo el perfil de identidad).
    El filtro por área del directorio (`SearchDirectory`), el del panel
    (`SearchAdmin`) y el de las notificaciones masivas (`ResolveRecipients`)
    usan la **misma** expresión que el chip (`specialtyAreaFilterSQL`, ver
    gotcha 24) — el filtro comparaba solo por la FK mientras el chip resolvía
    por FK + nombre, y por eso no encontraba a casi nadie.

12. **Bitácora de auditoría (audit logs)** — `api_change_logs` se escribe de
    forma **diferida y best-effort** (ver `docs/audit-logs.md`): `Record()`
    encola en memoria (buffer 5000) y un worker persiste en lotes de 50 cada 1s;
    un fallo o cola llena jamás revierte la operación principal. `Record` se
    invoca SOLO tras persistencia exitosa (0 llamadas si el UPDATE falla).
    - **Contrato de claves**: las claves del diff son snake_case y coinciden con
      los tags JSON reales del modelo (ver `service/audit_helpers.go`). La UI las
      traduce con `FIELD_LABELS`; una clave fuera de contrato cae al crudo.
    - **Gates**: lectura con `Sudo || CanViewLogs`, exportación CSV con
      `Sudo || CanExportLogs`, siempre 404 enmascarado (nunca 403). No existe
      `can_purge_logs`: la purga por antigüedad es exclusiva de Sudo (cron diario
      en `main.go`, `AUDIT_LOG_RETENTION_DAYS` default 90, `<=0` desactiva).
    - **`DeletePostGrade` (`DELETE /psi/me/postgrades/:id`) es SOFT delete** vía
      `gorm.DeletedAt` (embebido en `PsiUserPostGrade`): respeta IDOR y limpia el
      certificado en S3 best-effort. No lo cambies a borrado físico.

13. **Sesión de admin por-fila: re-login/cambio de password/logout matan las
    sesiones previas** — cada `user_admins` tiene su propia `key` (embebida en su
    JWT) y un admin puede tener varias pestañas activas; `validateToken` valida
    contra `GetByID(admin_id)` comparando esa key. Por tanto, CUALQUIER rotación
    deja **todas** las sesiones de ESE admin inválidas (401/404 enmascarado), sin
    tocar a los otros admins:
    - **Login** (`admin_service.go:84`): re-login rota la key — anti-replay, por diseño.
    - **Update con `password`** (`admin_service.go:457`): rota la key del editado —
      incluye el auto-edit: si un admin cambia SU contraseña, se caen sus otras pestañas.
    - **Logout** (`admin_service.go:125`): deja la key en `""` → mata la sesión en
      todas las pestañas de ese admin.
    - **DeleteAdmin** (`admin_service.go:515+`): soft-delete → `GetByID` deja de
      resolver la fila → muere la sesión del borrado (auto-borrado bloqueado por RBAC).
    Al depurar un "me sacó de la sesión" sin pasar por logout: sospechar rotación
    de key primero, NO la UI. Lo que **NO** mata la sesión vigente: `is_active=false`
    (solo bloquea re-login en `admin_service.go:73`) y `TransferSudo` (degrada
    permisos en vivo sin tocar keys).

14. **NO quitar `prefer_simple_protocol=true` del DSN** (`pkg/database/postgres.go`)
    — es el fix de los 500 `SQLSTATE 0A000`/`08P01` ("cached plan must not change
    result type" / "prepared statement name is already in use") que producían
    PgBouncer (transaction mode) + pgx: pgx cacheaba prepared statements y el
    pooler reparte cada consulta entre conexiones de backend distintas, así que
    un plan viejo sobrevivía a migraciones/reinicios. Con protocolo simple el
    driver no prepara statements. ⚠️ **PARA DESPLEGARlo, PgBouncer necesita
    ignorar el parámetro**: pgx SÍ lo envía como startup parameter y PgBouncer
    (transaction mode) lo rechaza con `FATAL: unsupported startup parameter:
    prefer_simple_protocol (SQLSTATE 08P01)` → la API queda en crash-loop al
    arrancar. Ya está agregado a `IGNORE_STARTUP_PARAMETERS` en
    `docker-compose.yml` (`extra_float_digits,search_path,prefer_simple_protocol`);
    ignorarlo no cambia la semántica (el protocolo simple es decisión client-side
    del driver). La API respondía esos 500 como genéricos y el frontend los pintaba
    como "Conexión en pausa" a pantalla completa (el admin percibía "me sacó del
    panel"). Si necesitas prepared statements por rendimiento en el futuro, la vía
    correcta es conectar directo a Postgres (no por PgBouncer), no quitar el param.

15. **Validación de sesión admin: `GET /session/me` y `GET /session/validate`,
    NO bajo `/admin`** — el frontend espera un **401 real** para distinguir
    "sesión inválida/revocada" de "ruta inexistente" (ver gotcha 12 de
    `web/AGENTS.md`), así que estos endpoints viven en su prefijo propio
    `/session` con `ProtectedAdmin()` (gotcha 16: registro bajo `/admin`
    hereda el enmascarado 404 y el 401 jamás se sirve, como pasó desde el
    05-sep). Verificación:
    `curl -i http://localhost:28080/api/v1/session/me` (sin token) → **401**
    (y con token válido → 200). NO existen más `/admin/me` ni
    `/admin/validate`; un binario viejo o un cliente que los llame recibe 404.

16. **QUIRK de Fiber v2: un 2º `Group()` (o una ruta directa) sobre un prefijo
    ya usado hereda el middleware del PRIMER grupo** — reproducido con
    v2.52.11: tras `admin := router.Group("/admin", NoStore,
    ProtectedAdmin404)`, cualquier registro posterior bajo `/admin` (otro
    `Group("/admin", ProtectedAdmin)` o `router.Get("/admin/me", ...)` directo
    sobre el grupo padre) queda apilado con el stack del primer grupo y NUNCA
    ejecuta su propio middleware. Ejemplo real: el grupo `adminValidate`
    (401 explícito) fue **código muerto desde el 05-sep** — `/admin/me` y
    `/admin/validate` respondieron 404 enmascarado a pesar de registrar
    `ProtectedAdmin` (confirmado incluso con build `--no-cache` y con una
    reproducción mínima). La única forma limpia de mezclar 401/404 bajo un
    mismo dominio: **prefijos distintos** (por eso `/session/*`). Si una
    versión nueva de Fiber cambia esto, validalo con un mini-app de dos grupos
    sobre el mismo prefijo antes de volver a anidarlos.

17. **Limitador global 60 req/min por IP** (`cmd/api/main.go:294`,
    `app.Use(limiter.New(...))` sin `Next` original) — aplica a TODOS los
    métodos y rutas, clave por `c.IP()`. Sintomatología: el panel admin recibe
    429 "Demasiadas solicitudes" en `/session/me`, `/admin/tickets/
    pendientes-count` y en acciones (archivar publicación → el modal muestra
    error genérico). Causa raíz: **las preflights OPTIONS del CORS cross-
    origin consumían cuota** (cada fetch del navegador = OPTIONS + método
    real, 2 unidades por petición) y, en la red Docker, **todo el tráfico del
    host y del dev server comparte la misma IP** (bridge `172.x.0.1`) — una
    ráfaga legítima (polling de `/session/validate` 60s + pendientes 30s +
    navegación + curls de verificación) agota el bucket y se auto-bloquea
    60s. Fix: `Next` salta `OPTIONS` y `/live` + `/ready` (healthchecks).
    **No re-agregues** OPTIONS al conteo ni quites el `Next`; mantener Max 60/
    min como defensa anti-DoS. ⚠️ Si `colpsi_valkey` está caído, el storage
    cae a **in-memory** (`rate_limiter.go:50`, log "No se pudo conectar…
    Usando in-memory"): los contadores viven en el proceso y **un restart de
    la API los resetea**; con Valkey arriba son persistente y multi-instancia
    (levantar: `docker compose up -d valkey` desde `api/`).

18. **La IP no se guarda en claro: se guarda una huella, y la sal
    (`ANALYTICS_IP_SALT`) tiene que ser estable** — `service/analytics_privacy.go`
    (`fingerprintIP`, `refererOrigin`). La sustitución ocurre **dentro de los
    métodos del servicio** (`TrackPageView`, `RecordSearch`, `RecordProfileView`,
    `RecordPageView`), no en los puntos de llamada, para que todo llamador
    futuro herede la garantía. La sal se resuelve **una vez** en
    `NewAnalyticsService`: si `ANALYTICS_IP_SALT` cambia entre reinicios la
    misma IP produce otra huella y los conteos de visitantes únicos se rompen;
    vacío usa una constante por defecto con el mismo propósito (que un
    despliegue malo no guarde IP en claro en silencio). ⚠️ **Una huella NO es
    anonimización**: una IPv4 tiene 2³² valores y es reversible por fuerza
    bruta; lo que acota el daño es `ANALYTICS_RETENTION_DAYS` (90), que
    programa `PurgeOldData` al arrancar y con un ticker diario — la función
    existía con test unitario y **nadie la invocaba**. `login_events` y
    `active_sessions` **no se tocan** (bitácora de seguridad). El texto de
    búsqueda **sí se conserva verbatim** y el §10.1 de `/terminos` lo declara:
    si se cambia esto hay que cambiar el texto, y al revés (ver
    `docs/plan-terminos-condiciones.md`).

19. **`text_id` / `bio_text_id` NULL = el `TextModel` no existe todavía, y
    GORM lo trata como `uuid.Nil`** — los campos son `uuid.UUID` **por valor**,
    así que una columna NULL se carga como UUID cero, no como NULL: las
    publicaciones y agremiados creados antes de que existiera el `TextModel`
    quedaron huérfanos (2 de 5 noticias y 1 de 103 agremiados en local) y
    cualquier edición de su cuerpo fallaba. **Dos síntomas distintos según
    quién inventara el ID**:
    - `tx.Model(text).Updates(...)` con la primary key en cero → GORM no genera
      WHERE → **`WHERE conditions required`** (500).
    - Un UUID "nuevo" generado en el servicio → el UPDATE afecta **0 filas sin
      error** y el guardado posterior del FK revienta `fk_psi_users_full_bio`
      (SQLSTATE 23503), que además se filtraba al cliente.
    **Quien decide es el repositorio, nunca el servicio**: `postRepo.Update` y
    el helper `upsertBioText` (compartido por `psiRepo.Update` y
    `UpdatePublicProfile`) hacen `Create` + enlace cuando el ID llega en
    `uuid.Nil`, ordenando **texto→enlace** para no referenciar una fila
    inexistente; los servicios solo rellenan autoría. Regla general: si un
    `Model(&X{}).Update(...)` puede recibir la primary key en cero, ramifica
    por `== uuid.Nil` y `Create`; y jamás inventes el ID de una fila que aún no
    has insertado, porque el fallo se manifiesta dos puntos más tarde (FK) y
    enmascarado. Ambas columnas admiten NULL a propósito: no las hagas NOT NULL
    sin un backfill.
    OJO `UpdatePsiByAdmin` responde **403 con `err.Error()` para cualquier
    error** del servicio (contrato heredado, no lo cambies sin revisar el
    frontend), así que un error de persistencia debe salir enmascarado desde el
    servicio (`log.Error()` + `MapDBError` o mensaje genérico) — el fallback
    passthrough de `MapDBError` es intencional y está testeado, no lo "arregles"
    globalmente.
    OJO la fila de `text_models` **nunca actualiza `update_by`** (el `Updates`
    solo escribe `content`), aunque el servicio lo setee en memoria: la
    trazabilidad real está en `api_change_logs`.
20. **La extensión `unaccent` la crea una migración, NO un modelo — y por eso
    `atlas migrate diff` va a proponer borrarla** — el buscador del directorio
    y la búsqueda de agremiados la usan en **18 llamadas** de
    `internal/repository/postgres/psi_repository.go`
    (`unaccent(first_name) ILIKE unaccent(?)`, etc.); sin ella, Postgres
    responde `42883 function unaccent(character varying) does not exist` y,
    como parece un fallo de búsqueda y no de esquema, se depura donde no es.
    Hasta `20260930160000_unaccent_extension.sql` no la declaraba **ninguna**
    migración ni `AutoMigrate` (`pkg/database/migration.go` solo activa
    `pgcrypto`), así que solo existía en las bases montadas a mano. Ahora es
    una migración (`CREATE EXTENSION IF NOT EXISTS ... WITH SCHEMA public`).
    ⚠️ Como **ningún modelo de dominio la declara**, Atlas la marca como
    objeto huérfano y el `.sql` que genera incluirá
    `DROP EXTENSION unaccent`: **elimínalo antes de aplicar** (la migración lo
    advierte en su propio comentario). Esto es la misma trampa que el
    `DROP INDEX idx_posts_status_publish_at` del diff de la persona de
    contacto, y más general: **`migrate diff` compara los modelos contra las
    migraciones, no contra la realidad de la base**, así que no conoce los
    `CHECK`, los índices **únicos parciales** ni las columnas escritas a mano
    (`reject_reason`, `last_change_reason`, `can_view_logs`…) y los propone
    borrar. Revisa siempre el `.sql` generado. Al depurar un 42883 o un
    "objeto no existe" de una base recién creada, mira primero si el
    `migrations/README.md` lista la extensión u objeto que falta.
    Las revisiones aplicadas viven en el schema `atlas_schema_revisions` (no
    en una tabla de `public`); si el directorio y esa tabla discrepan aparece
    *checksum mismatch* y se sincroniza con
    `atlas migrate set <versión> --env gorm --url ...`.
21. **El `root` de Audiobookshelf lo crea el servicio `abs-init`, y su
    contraseña SOLO existe si `ABS_ADMIN_PASSWORD` está en el `.env`** — ABS
    genera la clave del `root` al crear su base por primera vez y la muestra
    **una única vez** en los logs; si ese log se pierde (recreate del
    contenedor, logrotate, un server nuevo) la clave queda desconocida y **no
    hay forma de recuperarla**. Como `ABS_ADMIN_TOKEN` es un gate duro de
    arranque (`cmd/api/main.go:75` → `log.Fatal`), ese bloqueo tumba la API
    entera. El servicio `abs-init` (`scripts/abs-init.sh`, imagen `nginx:alpine`
    ya presente en el stack) lo resuelve: espera a ABS, consulta
    `GET /status` (`isInit`) y solo si viene `false` llama a `POST /init` — el
    propio endpoint de bootstrap de ABS, que **rechaza con 500 si
    `Database.hasRootUser` ya es true**, así que nunca pisa una clave existente.
    `api` lo espera con `depends_on: abs-init → service_completed_successfully`.
    - **`ABS_PASS` vacío + `isInit:false` → exit 1 a propósito**: frena el
      arranque de la API en vez de dejar que cree un `root` sin clave (ABS
      avisa `Creating root user with no password`, y un root sin clave es
      accesible sin autenticación).
    - ⚠️ El `wget` de `nginx:alpine` es el de **BUSYBOX, no el de GNU**: solo
      acepta `--post-data` (con `--header` aparte). Con `--method=POST` /
      `--body-data` (sintaxis de GNU) el POST **no se envía**, `wget` imprime su
      usage y sale !=0; si el script no comprobara el código de salida, el
      servicio reportaría "root creado" sin haber hecho nada.
    - El script vive **en disco**, montado `:ro`, y no inline en el `entrypoint`:
      dentro de un bloque `>` de YAML los newlines se convierten en espacios y el
      shell depende de ese folding para no cortar la línea del POST.
    - ⚠️ Esto **no** arregla un ABS que ya tiene un `root` con clave desconocida
      (`/init` da 500 a propósito): para eso hay que borrar
      `biblioteca/config/absdatabase.sqlite` (se pierden las cuentas de ABS, que
      recrea el worker de sync, y el progreso de lectura; **no** los PDFs de
      `biblioteca/books/`).
22. **La huella de la IP no cabía en `varchar(45)`: la telemetría estaba muerta y
    nadie la notó** — el fix de privacidad (gotcha 18) sustituyó la IP en claro
    por `hex(sha256(sal|ip))`, que son **64** caracteres, en las columnas `ip` de
    `page_views`, `search_events` y `profile_views`. Esas columnas eran
    `character varying(45)` (el largo máximo de una IPv6) y nadie las amplió, así
    que **todos** los inserts de las tres tablas murieron con
    `22001 value too long for type character varying(45)` y las tres tablas
    dejaron de crecer durante un día: visitas, búsquedas, fichas consultadas y
    visitantes únicos congelados en el panel.
    - **Por qué no se notó**: las escrituras de telemetría son fire-and-forget y su
      error se descartaba con `_ =`. El único rastro era una línea de GORM por
      evento. Ahora cada error de escritura pasa por
      `reportWriteError` (`analytics_ingest.go`), con throttling (primera línea
      completa, luego 1/minuto con el total de omitidas) y una línea de
      "volvió a escribir" al recuperarse. **No vuelvas a `_ =` en un write
      fire-and-forget**: discardable para la petición no es lo mismo que invisible.
    - **La lección de fondo — el contenido de una columna es un contrato con el
      código, no una constante de negocio**: al cambiar QUÉ se guarda hay que
      revisar el ancho en la misma jugada, y si la columna se renombra hay que
      actualizar a la vez el `gorm:"column:..."` del struct, el `AutoMigrate` y la
      migración. Aquí las tres piezas quedaron en contextos distintos: el struct sí
      se editó, la migración no existía.
    - OJO la base de pruebas (`TEST_DB_DSN`, `make test-repo`) construye los
      anchos con `AutoMigrate` desde los tags, **no** desde `migrations/`: un tag
      correcto sin migración pasa ese suite entero y solo falla en producción
      (aquí pasó). Por eso el contrato real lo fijan
      `TestAnalyticsRepo_ColumnasAguantanLoQueMandaElServicio` y
      `TestAnalyticsRepo_AnchosCoherentesConLaMigracion` (este lee
      `information_schema.columns`).
    - `FingerprintIP` es **pública a propósito**: la misma fórmula tiene que
      producir el histórico del backfill en SQL y lo que escribe el servicio, y
      `TestAnalyticsRepo_HashSQLDelBackfillEquivaleAFingerprint` compara ambas con
      la expresión **literal** de la migración
      (`encode(sha256(convert_to('<sal>' || '|' || ip,'UTF8')),'hex')`). El
      `convert_to(..., 'UTF8')` **no es opcional**: `sha256()` solo existe para
      `bytea` y `sal || '|' || ip` resuelve a `text`, así que sin él el backfill
      muere con `42883 function sha256(text) does not exist` (lo pasó de verdad al
      escribirlo: por eso el test ejecuta la expresión literal de la migración y
      no una con placeholders). Si cambias el algoritmo, ese test falla y hay que
      rehacer el histórico.
    - ⚠️ **Backfill del histórico, una vez por base, fuera de la migración**: las
      filas anteriores al fix guardan la IP **en claro**, y el §10.1 de los
      Términos promete que no se guarda. El SQL está documentado al final de
      `20261001100000_analytics_ip_hash.sql`; sustituye `<sal>` por
      `ANALYTICS_IP_SALT` del entorno (o por `analyticsDefaultIPSalt` si no está
      definida) y **usa la misma sal que la API**, o el visitante de antes y de
      después del corte contará como dos personas distintas. El filtro
      `ip_hash ~ '[.:]'` es lo que lo hace idempotente (IPv4 tiene punto, IPv6 dos
      puntos, un hex de 64 no tiene ninguno).
    - `login_events`, `active_sessions` y `psi_terms_acceptance` **siguen con la IP
      en claro y varchar(45) a propósito**: bitácora de seguridad de cuentas con
      sesión (§10.3) y prueba de aceptación jurídica. No las "corrijas" por
      uniformidad: hashear la IP de una aceptación falsifica el registro.
23. **Todo enlace e imagen de los correos va en el dominio del remitente, y la
    URL sale de `resolveSiteURL()` (APP_URL), nunca de una constante** — dos
    avisos materiales de un verificador de spam, y ninguno fallaba visiblemente:
    el correo se enviaba, se veía bien, y lo único que pasaba es que se degradaba
    la reputación del dominio de envío.
    - **Aviso "las URLs no coinciden con el dominio de envío"**: `mail_service.go`
      tenía `const siteURL = "https://franhsabt-testing-ground.lat"` (un dominio de
      pruebas) inyectada a TODAS las plantillas como `{{.SiteURL}}`, y
      `psi_service_password_reset.go` repetía la misma constante como fallback.
      Lo grave era que convivían **dos dominios en el mismo correo**: el botón de
      reset salía de `APP_URL` (correcto) y el pie de la constante. Ahora hay una
      sola fuente (`resolveSiteURL()`: `APP_URL` con `TrimRight("/")`, y fallback
      a `defaultSiteURL` si está vacía o es local — un host local no le sirve de
      nada a quien recibe el correo).
    - **Aviso "las imágenes deberían estar en el dominio de envío"**: la bandera
      venía hotlinkeada de `upload.wikimedia.org` en 6 plantillas. Ahora es
      `{{.SiteURL}}/bandera-carabobo.png`, un PNG propio en `web/public/`.
      OJO el SVG de Commons **era un composed de 5 rasters embebidos**
      (`<image>` con `data:image`), no arte vectorial, y Gmail/Outlook **no
      renderizan SVG**: la bandera era invisible en el correo aunque el aviso
      pasara. El PNG salió del render oficial
      (`commons.wikimedia.org/wiki/Special:FilePath/Bandera_de_carabobo.svg?width=320`
      → 330x220, 12 KB); **no** lo conviertas con ImageMagick local: su delegate
      de SVG invoca `rsvg-convert`, que no está instalado.
    - **El contrato lo vigila `TestTemplates_SinRecursosDeTerceros`**
      (`templates_test.go`): recorre **todas** las plantillas embebidas, extrae
      cada URL absoluta del HTML renderizado y falla si el host no es el de
      `SiteURL`; también falla si algún `<img>` es `.svg`. Si añades una
      plantilla con un recurso externo, el test lo dice — no lo compliques con una
      lista de excepciones.
    - ⚠️ **`notification.html` tenía `{{.SiteURL}}.`** y eso también lo cazó el
      test: al autolinkear, el punto final se va con la URL y queda el host
      `colegio-psicologos-carabobo.com.` (un host distinto, y feo). La URL va
      ahora en un `<a href>` con la puntuación **fuera**. Si escribes una URL en
      el texto de un correo, sepárala de la puntuación.
    - ⚠️ **`web/public/` va DENTRO de la imagen** (no hay bind-mount en
      `web/docker-compose.yml`): un archivo nuevo ahí no aparece hasta
      `docker compose build web && docker compose up -d web`. Un `curl` a un asset
      nuevo contra el contenedor viejo devuelve 404 con el HTML del fallback SPA
      (no el asset), aunque el build local ya lo tenga en `.output/public/`.
    - Los correos ya entregados no se reescriben: el fix evita que el próximo envío
      vuelva a emitir esos avisos, pero la reputación de calentamiento que ya
      acumuló el dominio sigue como esté.
24. **El área que se MUESTRA y el área por la que se FILTRA son la MISMA
    expresión — y por eso viven en constantes, no repetidas** — el "filtro por
    área de desempeño" estaba roto en tres sitios a la vez y **nada fallaba**:
    `SearchDirectory`, `SearchAdmin` y `ResolveRecipients`
    (`notification_repo.go`, el selector de destinatarios de las notificaciones
    masivas) filtraban por `primary_specialty_id = ? OR
    secondary_specialty_id = ?`, o sea **solo por FK**, mientras la tarjeta
    mostraba el área resuelta con `COALESCE(sp1.name, sp1n.name, '')` (FK y, si
    falta, coincidencia EXACTA del legacy). En los datos la FK está poblada en
    **2 de 103 agremiados**: el directorio pintaba 13 chips "Clínica" y el filtro
    devolvía 2; las 7 tarjetas de "Neuropsicología" devolvían **0** (directorio
    vacío) y una notificación masiva dirigida a "Clínica" salía a 2 personas en vez
    de a 21. El síntoma era "el filtro no funciona", no un error: el chip y su
    propio filtro eran dos nociones distintas de "tener esta área".
    - **Regla**: el filtro compara contra `specialtyAreaFilterSQL`, que se arma
      con las mismas constantes (`specialtyPrimaryArea` / `specialtySecondaryArea`)
      que el SELECT del chip, y los 4 JOIN van en `applySpecialtyJoins` (usado por
      las tres consultas). El nombre de catálogo del ID entrante se resuelve con
      una subconsulta **no correlacionada** (Postgres la evalúa una vez), y
      `psi_specialty_models.name` tiene índice ÚNICO
      (`idx_psi_specialty_models_name`), así que los JOIN por nombre no pueden
      duplicar filas y el COALESCE del SELECT y el del WHERE devuelven lo mismo.
      Si añades un cuarto consumidor, usa las constantes: **copiar el SQL a mano
      es exactamente cómo drifted esto**.
    - ⚠️ **Al añadir esos JOIN a una consulta que ya existía, hay que CUALIFICAR
      `id` y `created_at`**: `psi_specialty_models` comparte `id`, `created_at`,
      `updated_at` y `deleted_at` con `psi_users`, así que un `Select("id, …")` o
      un `Order("created_at DESC")` sin calificar muere con
      `42702 column reference "id" is ambiguous` (y el `Order`, también en el
      `Count` que GORM arma aparte). En `SearchAdmin` quedaron
      `psi_users.id` / `psi_users.created_at`, y en `ResolveRecipients`,
      `Pluck("psi_users.id", …)`.
    - Un legacy que **no existe en el catálogo** (`Deportiva`, `Forense`,
      `Social`, `Organizacional` — 63 filas en local) sigue sin ser alcanzable por
      ningún filtro **y es lo correcto**: mostrarse un área inventada rompe
      gotcha 11. Es una limitación de **datos**, no del código: se arregla
      asignando la FK (o migrando el texto al nombre exacto del catálogo), no
      tocando el SQL.
    - El contrato lo fijan `TestPsiRepo_FiltroDeAreaCoincideConElChip` y
      `TestPsiRepo_FiltroDeAreaDelPanelAdminEsElMismo`
      (`psi_repo_specialty_filter_test.go`, requieren `make test-repo`): montan
      mitad con FK y mitad solo con legacy, y exigen que filtrar devuelva lo mismo
      que se muestra. Ojo al escribir fixtures de `PsiUserModel`: `bio_text_id` y
      `audio_book_shell_id` son FK/UNIQUE y sin valor rompen el INSERT (23503 /
      23505) — ver gotcha 19.

## TestKnownFlaky: TestGetAccess_ConcurrentSameUser

`internal/service/audiobookshelf_service_test.go` falla de forma intermitente
(~10% de las corridas) **y es esperado por diseño**: el test exige
`createCalls == 2` ("ambas goroutines intentaron crear") pero `GetAccess`
serializa por usuario, así que según el scheduler hay 1 o 2 llamadas. No es
una regresión ni una race (verificado con `-race -count=10`: limpio, y falla
igual en baseline sin los cambios del commit actual). Si lo ves fallar, no lo
persigas: relanza `make test-unit`.

Los 4 fallos de repositorio que sí son preexistentes y constantes
(`make test-repo`) son `TestAdminRepo_ComprehensiveSuite`,
`TestAdminRepo_Update_PreservesBoolean{False,True}` y
`TestPsiRepo_ComprehensiveSuite`.

## Estructura

```
api/
├── cmd/api/main.go        # Bootstrap: InitConfig → DB → S3 → servicios → router
├── cmd/exp/migrate/       # Generador de esquema Atlas desde modelos Go
├── internal/
│   ├── config/            # Singleton de env vars (config.Envs)
│   ├── domain/            # Modelos GORM + interfaces de repositorio
│   ├── handler/           # Adaptadores HTTP (con anotaciones Swagger)
│   ├── middleware/        # Auth (JWT), RateLimit, Idempotency, Analytics
│   ├── repository/postgres/  # Implementaciones GORM
│   ├── request_structs/   # DTOs + validación validator/v10
│   ├── router/            # Registro de rutas por dominio (psi, admin, ...)
│   ├── service/           # Lógica de negocio (+ README.md por módulo)
│   ├── templates/         # Emails HTML embebidos
│   └── utils/             # Helpers (slugs, sanitize docs, randoms)
├── pkg/
│   ├── database/          # Conexión GORM, migración, seed admin
│   └── s3/                # Cliente S3/MinIO + GetPublicURL
├── migrations/            # SQL versionado por Atlas (baseline + diffs)
└── docs/                  # Swagger/OpenAPI generado
```

## Convenciones

- Comentarios en español (los existentes son el estilo a seguir).
- Orden de un módulo nuevo: domain → repository → service → request_structs →
  handler → router → middleware → swagger → tests (ver `README.md`).
- Respetar el contrato con el frontend: las rutas admin devuelven 404 (no 403)
  y las URLs de imágenes salen con `S3_PUBLIC_URL`.
- Un solo commit por fix en la rama `docs`; preservar funcionamiento.

## Flujo de verificación rápida

1. `go build ./... && go vet ./...` — ¿compila sin advertencias?
   OJO: `go build ./...` desde `api/` falla con
   `db_data/pgdata: permission denied` (el datadir de Postgres en el árbol);
   usa `go build ./internal/... ./cmd/... ./pkg/...`.
2. `make test-unit` — ¿tests unitarios en verde?
   OJO: `TestGetAccess_ConcurrentSameUser` es flaky (ver
   [TestKnownFlaky](#testknownflaky-testgetaccess_concurrentsameuser)); relanza
   antes de culpar a tu cambio. `make test-repo` necesita
   `colpsi_test_db` arriba (`docker compose -f docker-compose.test.yml up -d`)
   y tiene 4 fallos preexistentes constantes.
3. Con Docker: `docker compose up -d` (api, db, pgbouncer, s3, valkey).
4. `curl http://localhost:28080/api/v1/psi/directory` → 200 con URLs de imágenes
   `http://localhost:29000/colpsi-bucket/...` (nunca `s3:9000`).
5. Login admin → `PATCH /api/v1/admin/psi/:id` con la cookie `jwt` → 200 (no el
   404 enmascarado del middleware).

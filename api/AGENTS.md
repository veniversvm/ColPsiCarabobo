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
2. `make test-unit` — ¿tests unitarios en verde?
3. Con Docker: `docker compose up -d` (api, db, pgbouncer, s3, valkey).
4. `curl http://localhost:28080/api/v1/psi/directory` → 200 con URLs de imágenes
   `http://localhost:29000/colpsi-bucket/...` (nunca `s3:9000`).
5. Login admin → `PATCH /api/v1/admin/psi/:id` con la cookie `jwt` → 200 (no el
   404 enmascarado del middleware).

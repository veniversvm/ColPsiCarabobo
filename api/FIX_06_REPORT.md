# FIX-06 Report — Hardcoded secrets como defaults en configuración

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-06 |
| **Hallazgo original** | CRIT-06 |
| **Archivos modificados** | `internal/config/env.config.go`, `cmd/api/main.go`, `pkg/database/seed.go`, `api/.env` |
| **Fecha de implementación** | 2026-07-24 |
| **Estado** | Completado |

---

## Problema

Tres secrets estaban hardcodeados como defaults en `env.config.go:110-111`:

| Secret | Default | Uso |
|--------|---------|-----|
| `JWT_LIBRARY_SECRET` | `"secret-for-psi-library"` | Clave HMAC para firmar JWTs de la biblioteca |
| `ABS_ADMIN_TOKEN` | `"deaful_ABS_ADMIN_TOKEN` | Token de autenticación contra Audiobookshelf |

Si no se configuraban las variables de entorno, la aplicación usaba valores predecibles que cualquier persona con acceso al código fuente podía conocer.

Adicionalmente, `seed.go:22` usaba `"admin123"` como password del super admin sin importar el entorno.

---

## Corrección

### 1. `internal/config/env.config.go:110-111` — Defaults vacíos

```go
// ANTES:
JwtLibrarySecret: getEnv("JWT_LIBRARY_SECRET", "secret-for-psi-library"),
AbsAdminToken:    getEnv("ABS_ADMIN_TOKEN", "deaful_ABS_ADMIN_TOKEN"),

// DESPUÉS:
JwtLibrarySecret: getEnv("JWT_LIBRARY_SECRET", ""),
AbsAdminToken:    getEnv("ABS_ADMIN_TOKEN", ""),
```

### 2. `cmd/api/main.go:47-52` — Validación al startup

```go
config.InitConfig()

if config.Envs.JwtLibrarySecret == "" {
    log.Fatal("❌ JWT_LIBRARY_SECRET no está configurado. Defina la variable de entorno.")
}
if config.Envs.AbsAdminToken == "" {
    log.Fatal("❌ ABS_ADMIN_TOKEN no está configurado. Defina la variable de entorno.")
}
```

La aplicación no arranca si los secrets no están definidos — falla rápido y ruidoso.

### 3. `pkg/database/seed.go:21-25` — Password condicional

```go
// ANTES:
defaultPass := "admin123"

// DESPUÉS:
var defaultPass string
if config.Envs.Environment == "development" {
    defaultPass = "admin123"
} else {
    defaultPass = utils.GenerateSecureRandomString(16)
}
```

- **Desarrollo:** password predecible `"admin123"` (fácil de recordar)
- **Producción:** password aleatoria de 16 caracteres via `utils.GenerateSecureRandomString`

### 4. `api/.env:40-41` — Secrets generados

```bash
JWT_LIBRARY_SECRET=S1Apui7IkH/3kE7quUc8Cb2gDK2y8qqwQUSW6qnuALk=
ABS_ADMIN_TOKEN=6CIuxwdLSqGmoN0w2YUCLHR62fpIxJSE7qQkpESSGdQ=
```

Generados con `openssl rand -base64 32`.

---

## Comportamiento resultante

| Entorno | `APP_ENV` | Secrets en `.env` | Resultado |
|---------|-----------|-------------------|-----------|
| Desarrollo | `development` | Sí | Arranca, seed usa `"admin123"` |
| Producción | `production` | Sí | Arranca, seed genera password aleatoria |
| Sin configurar | `production` (default) | No | `log.Fatal` — no arranca |

---

## Testing

| Verificación | Estado |
|---|---|
| `go vet ./internal/config/...` | Pass |
| `go vet ./cmd/api/...` | Pass |
| `go vet ./pkg/database/...` | Pass |
| `go test ./internal/middleware/...` | Pass (6/6) |
| `go test ./internal/utils/...` | 4/5 pass, 1 fallo pre-existente (FIX-33) |

---

## Seguridad

- **Fail-safe secure:** Si no se configuran los secrets, la app no arranca en vez de usar valores predecibles.
- **Principio de mínimo privilegio:** Cada entorno tiene secrets diferentes (generados aleatoriamente).
- **Separación de entornos:** El mismo `.env` no debe usarse entre desarrollo y producción.

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/config/env.config.go:110-111` | Defaults removidos |
| `cmd/api/main.go:47-52` | Validación de secrets |
| `pkg/database/seed.go:21-25` | Password condicional |
| `api/.env:40-41` | Secrets generados |
| `internal/utils/random_string.go` | Función `GenerateSecureRandomString` reutilizada |
| `SECURITY_FIX_PLAN.md` (FIX-06) | Plan de seguridad referenciado |

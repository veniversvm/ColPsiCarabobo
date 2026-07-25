# FIX-05 Report — Seed admin credentials expuestas en logs

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-05 |
| **Hallazgo original** | CRIT-05 |
| **Archivos modificados** | `pkg/database/seed.go`, `internal/config/env.config.go`, `api/.env` |
| **Fecha de implementación** | 2026-07-24 |
| **Estado** | Completado |

---

## Problema

La función `SeedAdmin` en `pkg/database/seed.go:65` logueaba la contraseña del super admin en texto plano en cada startup:

```go
// ANTES (línea 65):
log.Printf("ℹ️  User: %s | Pass: %s | ID: %s", admin.Username, defaultPass, admin.ID)
```

Cualquier persona con acceso a los logs del sistema (CloudWatch, systemd journal, archivos de log) podía ver las credenciales del super admin.

Además, el default de `APP_ENV` era `"development"`, lo que significaba que incluso en producción se mostraba la contraseña si no se configuraba explícitamente la variable.

---

## Corrección

### 1. `internal/config/env.config.go:95` — Default seguro

```go
// ANTES:
Environment: getEnv("APP_ENV", "development"),

// DESPUÉS:
Environment: getEnv("APP_ENV", "production"),
```

Si `APP_ENV` no está definida en el entorno, se asume producción (no se muestra la password).

### 2. `pkg/database/seed.go:64-68` — Log condicional

```go
// ANTES:
log.Println("✅ Super Admin creado exitosamente.")
log.Printf("ℹ️  User: %s | Pass: %s | ID: %s", admin.Username, defaultPass, admin.ID)
log.Println("⚠️  POR FAVOR CAMBIE LA CONTRASEÑA AL INICIAR SESIÓN")

// DESPUÉS:
log.Println("✅ Super Admin creado exitosamente.")
if config.Envs.Environment == "development" {
    log.Printf("ℹ️  [DEV] User: %s | Pass: %s | ID: %s", admin.Username, defaultPass, admin.ID)
} else {
    log.Printf("ℹ️  Super Admin creado — User: %s | ID: %s", admin.Username, admin.ID)
    log.Println("⚠️  La contraseña fue generada automáticamente. Cámbiela al iniciar sesión.")
}
```

### 3. `api/.env` — Variable explícita para desarrollo

```bash
APP_ENV=development  # Solo para desarrollo. En producción: eliminar esta línea o cambiar a "production"
```

---

## Comportamiento resultante

| Entorno | `APP_ENV` | Log del seed |
|---------|-----------|--------------|
| Desarrollo (con `.env`) | `development` | `[DEV] User: admin \| Pass: admin123 \| ID: ...` |
| Producción (sin `.env`) | `production` (default) | `Super Admin creado — User: admin \| ID: ...` |
| Producción (con `.env`) | `production` | `Super Admin creado — User: admin \| ID: ...` |

---

## Testing

| Verificación | Estado |
|---|---|
| `go vet ./internal/config/...` | Pass |
| `go vet ./pkg/database/...` | Pass |
| `go test ./internal/middleware/...` | Pass (6/6) |
| `go test ./internal/utils/...` | 4/5 pass, 1 fallo pre-existente (FIX-33) |

---

## Nota de seguridad

El default `"development"` original era una práctica insegura: si se olvidaba configurar `APP_ENV`, la contraseña se logueaba en producción. El cambio a default `"production"` aplica el principio de **fail-safe secure**: si no se configura explícitamente, el sistema asume el entorno más restrictivo.

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `pkg/database/seed.go:65` | Línea original con el leak |
| `internal/config/env.config.go:95` | Default de `APP_ENV` |
| `api/.env` | Variable de entorno para desarrollo |
| `SECURITY_FIX_PLAN.md` (FIX-05) | Plan de seguridad referenciado |

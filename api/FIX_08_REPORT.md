# Reporte FIX-08 — Rate Limiter con Valkey

**Fecha:** 25 de Julio, 2026
**Commit:** `64c2457`
**Severidad:** 🟠 ALTO

---

## Hallazgo

El rate limiter usaba `limiter.New()` de Fiber con almacenamiento in-memory. Esto significa:

1. **Sin persistencia:** Al reiniciar el servicio, todos los contadores de intentos se pierden
2. **Sin multi-instancia:** Si se escalan réplicas de Docker/K8s, cada instancia tiene su propio contador
3. **Reset automático:** Un restart del contenedor "libera" a un atacante

---

## Solución implementada

### Arquitectura

```
┌─ rate_limiter.go ──────────────────────────────────────────┐
│                                                             │
│  newRateLimiterStorage()                                    │
│  ├── sync.Once (inicialización lazy)                        │
│  ├── config.Envs.ValkeyAddr != "" ?                         │
│  │   ├── SÍ → valkey.New(addr) → fiber.Storage             │
│  │   │         ├── Panic recovery si conexión falla         │
│  │   │         └── Log: "[RATE-LIMIT] Modo: Valkey"         │
│  │   └── NO → nil (Fiber usa in-memory por defecto)        │
│  │             └── Log: "[RATE-LIMIT] Modo: in-memory"      │
│  └── Retorno: fiber.Storage o nil                           │
│                                                             │
│  AuthRateLimiter()     ──→ Storage: newRateLimiterStorage() │
│  AdminAuthRateLimiter()──→ Storage: newRateLimiterStorage() │
└─────────────────────────────────────────────────────────────┘
```

### Comportamiento por escenario

| Escenario | `VALKEY_ADDR` | Comportamiento |
|-----------|---------------|----------------|
| Desarrollo local | vacío | In-memory (como antes) |
| Docker compose | `valkey:6379` | Valkey persistente |
| Valkey cae | `valkey:6379` | Panic recovery → in-memory |
| Sin config | vacío | In-memory (cero breaking changes) |

---

## Archivos modificados

| Archivo | Cambio |
|---------|--------|
| `internal/config/env.config.go:50` | Nuevo campo `ValkeyAddr string` |
| `internal/config/env.config.go:110` | `getEnv("VALKEY_ADDR", "")` |
| `internal/middleware/rate_limiter.go` | Reescritura completa: `newRateLimiterStorage()` + Valkey store |
| `docker-compose.yml` | Nuevo servicio `valkey` + volumen `valkey_data` |
| `docker-compose.yml:84` | `VALKEY_ADDR=valkey:6379` en api env |

---

## Servicio Valkey en docker-compose.yml

```yaml
valkey:
  image: valkey/valkey:9.1-alpine    # ~18MB (vs Redis ~32MB)
  container_name: colpsi_valkey
  ports:
    - "6379:6379"
  volumes:
    - valkey_data:/data
  command: valkey-server --save 60 1 --loglevel warning
  networks:
    - colpsi_network
  restart: unless-stopped
  healthcheck:
    test: ["CMD", "valkey-cli", "ping"]
    interval: 10s
    timeout: 5s
    retries: 5
```

---

## Dependencias agregadas

```
github.com/gofiber/storage/valkey v0.3.1
├── github.com/valkey-io/valkey-go v1.0.76
└── github.com/gofiber/storage v1.3.3
```

---

## Por qué Valkey vs Redis

| Aspecto | Valkey | Redis |
|---------|--------|-------|
| License | BSD-3 (open source) | SSPL (restrictivo) |
| Docker image | ~18MB (alpine) | ~32MB (alpine) |
| Performance | 5-10% mejor en benchmarks | Estándar |
| Compatibilidad | 100% Redis protocol | Nativo |
| Community | Linux Foundation | Redis LTD |
| Costo | Gratis siempre | Gratis (pero license cambió) |

---

## Tests

| Paquete | Resultado |
|---------|-----------|
| `internal/middleware` | ✅ 17/17 pass |
| `pkg/job` | ✅ 4/4 pass |
| `internal/domain` | ✅ pass |
| `internal/service` | 2 fallos pre-existente |
| `internal/repository/postgres` | 4 fallos (sin DB local) |

**0 regresiones.**

---

## Actualización SECURITY_FIX_PLAN.md

```markdown
### FIX-08: Rate Limiter con Valkey (multi-instancia)

| Campo | Valor |
|-------|-------|
| **Estado** | ✅ IMPLEMENTADO |
| **Commit** | 64c2457 |
| **Almacenamiento** | Valkey (fallback in-memory) |
| **Docker** | Servicio valkey/valkey:9.1-alpine |
```

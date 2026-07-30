# 🛡️ Middleware de Seguridad (middleware/)

> **[⬆ internal](../)** — `api/internal/middleware/`

El perímetro de seguridad de la API. Cada request debe atravesar estos 4 módulos antes de llegar a los handlers.

## Arquitectura de Seguridad

```
                    ┌─────────────────────────────────────────────────┐
                    │              REQUEST entrante                   │
                    └──────────────────────┬──────────────────────────┘
                                           │
                    ┌──────────────────────▼──────────────────────────┐
                    │           1. ANALYTICS MIDDLEWARE               │
                    │  • Intercepta TODOS los requests                │
                    │  • Crea cookie de sesión (30min, UUID)          │
                    │  • Registra page views, búsquedas, logins       │
                    │  • Goroutines no-bloqueantes para DB writes     │
                    │  • Trackea: IP, user agent, path, referrer      │
                    └──────────────────────┬──────────────────────────┘
                                           │
                    ┌──────────────────────▼──────────────────────────┐
                    │              REQUEST ID                          │
                    │  • Asigna ID único al request                   │
                    │  • Propagado en headers de respuesta            │
                    └──────────────────────┬──────────────────────────┘
                                           │
                    ┌──────────────────────▼──────────────────────────┐
                    │                CORS                              │
                    │  • Control de origenes permitidos               │
                    │  • Headers permitidos                           │
                    └──────────────────────┬──────────────────────────┘
                                           │
                    ┌──────────────────────▼──────────────────────────┐
                    │         2. RATE LIMITER MIDDLEWARE              │
                    │  • Sliding window en memoria                    │
                    │  • Auth: 10 req/15min por IP                    │
                    │  • Admin: 5 req/30min por IP                    │
                    │  • 429 Too Many Requests + Retry-After          │
                    └──────────────────────┬──────────────────────────┘
                                           │
                    ┌──────────────────────▼──────────────────────────┐
                    │        3. IDEMPOTENCY MIDDLEWARE                 │
                    │  • Previene duplicados en POST/PUT/DELETE        │
                    │  • SHA-256(body + userId + path) como key       │
                    │  • TTL configurable para expiración             │
                    └──────────────────────┬──────────────────────────┘
                                           │
                    ┌──────────────────────▼──────────────────────────┐
                    │            4. AUTH MIDDLEWARE (per-route)       │
                    │  • JWT Bearer token validation                  │
                    │  • Admin 404-Hiding                             │
                    │  • Hybrid Auth (tokens opcionales)              │
                    │  • Psi User Protection                          │
                    │  • Extracción: userId, userRole, userEmail      │
                    └──────────────────────┬──────────────────────────┘
                                           │
                    ┌──────────────────────▼──────────────────────────┐
                    │              HANDLER / CONTROLLER                │
                    └─────────────────────────────────────────────────┘
```

## Capas de Seguridad

```
┌─────────────────────────────────────────────────────────────────┐
│  CAPA 4: AUTHORIZATION (Auth Middleware)                        │
│  ¿Quién eres? ¿Qué puedes hacer?                               │
│  • JWT validation, Role extraction, Admin 404-Hiding            │
├─────────────────────────────────────────────────────────────────┤
│  CAPA 3: IDEMPOTENCY (Idempotency Middleware)                   │
│  ¿Es este request único?                                        │
│  • SHA-256 dedup, TTL cache, scoped per user                    │
├─────────────────────────────────────────────────────────────────┤
│  CAPA 2: RATE LIMITING (Rate Limiter Middleware)                │
│  ¿Cuántas veces has pedido esto?                                │
│  • Sliding window, IP-based, 429 responses                      │
├─────────────────────────────────────────────────────────────────┤
│  CAPA 1: OBSERVABILITY (Analytics Middleware)                   │
│  ¿Qué está pasando?                                            │
│  • Session tracking, page views, event logging                  │
└─────────────────────────────────────────────────────────────────┘
```

## Módulos

### 1. Analytics Middleware (`analytics.go`)

Intercepta **todos** los requests para registrar actividad del usuario.

**Funcionalidades:**
- Crea cookie de sesión con UUID (expira a los 30 minutos)
- Registra automáticamente: page views, profile views, search events, login events
- Usa goroutines para writes a DB no-bloqueantes (fire-and-forget)
- Captura: IP del cliente, user agent, path solicitado, referrer, session ID

**Flujo:**
```
Request → Extraer/Crear session cookie → Capturar metadata → Goroutine → DB
```

**Eventos trackeados:**
| Evento         | Trigger                          |
|----------------|----------------------------------|
| `page_view`    | Cualquier request GET            |
| `profile_view` | Request a `/api/profiles/:id`    |
| `search`       | Request a `/api/search`          |
| `login`        | Request exitoso a `/api/login`   |

---

### 2. Auth Middleware (`auth.go`)

Validación JWT y control de acceso basado en roles.

**Funcionalidades:**

#### JWT Validation
- Extrae token del header `Authorization: Bearer <token>`
- Valida firma y expiración
- Si no hay token → behaviour depende del tipo de ruta

#### Admin 404-Hiding
```
 Usuario normal → GET /admin/dashboard → 404 Not Found (no 403)
 Admin         → GET /admin/dashboard → 200 OK
```
Los usuarios no-admin obtienen **404** (no 403) para rutas admin. Esto evita revelar la existencia de panel administrativo.

#### Hybrid Auth
Algunas rutas aceptan tokens opcionales:
- Sin token → comportamiento público (menos datos)
- Con token → comportamiento autenticado (datos completos)

#### Psi User Protection
Usuarios con rol `PSICOLOGO` solo pueden acceder a sus propios datos. El middleware verifica que el userId del token coincida con el recurso solicitado.

#### Roles
| Rol              | Descripción                        |
|------------------|------------------------------------|
| `ADMINISTRADOR`  | Acceso total, incluye rutas admin  |
| `SUPER_USUARIO`  | Acceso elevado sin panel admin     |
| `USUARIO`        | Acceso básico                      |

#### c.Locals() extraídos
```go
c.Locals("userId", user.ID)
c.Locals("userRole", user.Role)
c.Locals("userEmail", user.Email)
```

---

### 3. Idempotency Middleware (`idempotency.go`)

Previene operaciones duplicadas en mutaciones.

**Funcionalidades:**
- Aplica solo a **POST**, **PUT**, **DELETE** (no a GET)
- Genera key con SHA-256 de: `request body + user ID + path`
- **Scoped por usuario**: dos usuarios diferentes pueden enviar el mismo body sin conflicto
- TTL configurable: las entradas expiran después de un tiempo determinado

**Flujo:**
```
POST /api/appointments
    → SHA-256(body + userId + path) = "a1b2c3..."
    → ¿Existe en cache?
        → SÍ: 409 Conflict (duplicate detected)
        → NO: Guardar en cache + Continuar al handler
```

**Key generation:**
```go
key = SHA256(requestBody + "|" + userId + "|" + requestPath)
```

---

### 4. Rate Limiter Middleware (`rate_limiter.go`)

Limitación de tasa basada en IP con sliding window en memoria.

**Configuración:**

| Endpoint Pattern   | Límite            | Ventana  |
|--------------------|-------------------|----------|
| `/api/auth/*`      | 10 requests       | 15 min   |
| `/api/admin/*`     | 5 requests        | 30 min   |
| Todo lo demás      | Sin límite        | —        |

**Respuesta cuando se excede el límite:**
```json
{
  "error": "Too Many Requests",
  "retry_after": 845
}
```
- Status code: `429 Too Many Requests`
- Header: `Retry-After: <seconds>`

**Implementación:**
- Sliding window log (no fixed window) para distribución más uniforme
- Almacenamiento en memoria (se resetea al reiniciar el servidor)
- Tracking por IP del cliente

---

## Orden de Ejecución

```
Analytics → RequestID → CORS → RateLimiter → Idempotency → Auth (per-route)
```

**Nota:** `Auth` se aplica **per-route**, no globalmente. Cada grupo de rutas define si necesita auth público, hybrid o obligatorio.

## Archivos

| Archivo              | Descripción                          |
|----------------------|--------------------------------------|
| `analytics.go`       | Tracking de actividad y sesiones     |
| `auth.go`            | JWT validation y control de acceso   |
| `idempotency.go`     | Deduplicación de mutaciones          |
| `rate_limiter.go`    | Limitación de tasa por IP            |
| `auth_test.go`       | Tests para escenarios de auth        |

## Notas de Seguridad

- Los tokens JWT se validan en cada request (no se cachea la validación)
- El Admin 404-Hiding es transparente para el cliente (nunca ve 403)
- El Rate Limiter se resetea con el servidor (in-memory)
- Las goroutines de Analytics son fire-and-forget (no afectan la respuesta)
- La Idempotency key incluye el userId para evitar falsos positivos entre usuarios

**[⬆ Volver a internal](../)**

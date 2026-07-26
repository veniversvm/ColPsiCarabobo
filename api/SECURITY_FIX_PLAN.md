# 🔧 Plan de Corrección — Auditoría de Seguridad

**Fecha:** 24 de Julio, 2026
**Estado:** Pendiente de implementación
**Metodología:** Fix por fix, verificación post-fix, commits atómicos
**Metodología de análisis:** Verificación cruzada del reporte contra código fuente real (59 archivos .go)

---

## 📋 Tabla de Contenidos

- [Corrección del Reporte Original](#corrección-del-reporte-original)
- [Fase 1 — Críticos (antes de producción)](#fase-1--críticos-antes-de-producción)
- [Fase 2 — Altos (antes de escalar)](#fase-2--altos-antes-de-escalar)
- [Fase 3 — Medios (deuda técnica)](#fase-3--medios-deuda-técnica)
- [Fase 4 — Bajos (polish)](#fase-4--bajos-polish)
- [Verificación Final](#verificación-final)
- [Orden de Implementación](#orden-de-implementación)

---

## Corrección del Reporte Original

El reporte de auditoría contiene **1 error factual** que debe corregirse antes de implementar:

### CRIT-01: Descripción Incorrecta → Reescribir

**El reporte dice:**
> *"OptionalHybridAuth retorna nil, nil en el jwt.Keyfunc cuando el usuario existe en la DB. Esto causa que jwt.Parse use nil como clave HMAC, lo que significa que cualquier token firmado con cualquier string pasa la validación siempre que el user_id exista."*

**La realidad verificada en código (`auth.go:153-198`):**

```go
// Línea 165: Se llama a jwt.Parse con un Keyfunc
_, _ = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    // ...
    if role == "admin" {
        admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
        if err == nil {
            c.Locals("admin", admin)
            return []byte(admin.Key), nil   // ← LÍNEA 184: RETORNA LA KEY REAL
        }
    } else {
        psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
        if err == nil {
            c.Locals("psi_user", psi)
            return []byte(psi.Key), nil     // ← LÍNEA 190: RETORNA LA KEY REAL
        }
    }
    return nil, nil    // ← LÍNEA 193: solo si el DB lookup FALLA
})
return c.Next()
```

Cuando el usuario **existe** en la DB, las líneas 184/190 **sí retornan la key real**. El `nil, nil` de la línea 193 solo se alcanza si la consulta a la DB falla. Además, `jwt.Parse` con key `nil` **rechaza** el token (falla type assertion `nil.([]byte)` en `hmac.go:58-63`).

**El problema REAL es diferente pero igualmente grave:**

En `golang-jwt/v5`, el `Keyfunc` se ejecuta **antes** de la verificación de firma (`parser.go:85-107`). El Keyfunc ejecuta `c.Locals("admin", admin)` como side-effect **antes** de que la firma sea verificada. Como el resultado de `jwt.Parse` se descarta con `_, _` (línea 165), un token forjado con un `user_id` válido causaría que:

1. El Keyfunc busca el usuario en DB → lo encuentra
2. Setea `c.Locals("admin", admin)` ← **side effect ya ocurrió**
3. Retorna la key real del usuario
4. `jwt.Parse` verifica la firma con la key real → **falla** (firma inválida)
5. El error se descarta (`_, _`)
6. `c.Next()` se ejecuta con el admin **ya inyectado** en Locals

Un atacante puede forjar un token con `"user_id": "<victim_uuid>"` firmado con cualquier key, y el middleware inyectará la víctima en `c.Locals` aunque la firma sea inválida.

**Contraste con `ProtectedAdmin404`** (línea 110-144): Usa `validateToken()` que **retorna** el token y verifica `err != nil || !token.Valid` (línea 127). Si el token es inválido, rechaza la请求. Los side-effects ocurren pero nunca llegan al controller.

---

## Fase 1 — Críticos (antes de producción)

---

### FIX-01: `OptionalHybridAuth` — Side-effect-before-validation

| Campo | Valor |
|-------|-------|
| **Hallazgo** | CRIT-01 (corregido — ver arriba) |
| **Archivo** | `internal/middleware/auth.go:153-198` |
| **Severidad** | 🔴 CRÍTICO |
| **Riesgo** | Un atacante puede inyectar identidad de cualquier usuario con un token firmado con cualquier key |

**Código actual (resumen):**
```go
func (m *AuthMiddleware) OptionalHybridAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // ...
        _, _ = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Side effects (c.Locals) ocurren ANTES de verificación de firma
            admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
            if err == nil {
                c.Locals("admin", admin)  // ← inyecta ANTES de validar
                return []byte(admin.Key), nil
            }
            return nil, nil
        })
        return c.Next()  // ← nunca verifica si el token era válido
    }
}
```

**Fix propuesto — Reescribir la función completa:**
```go
func (m *AuthMiddleware) OptionalHybridAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            return c.Next()
        }
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // PASO 1: Parse + Verificar firma (sin side effects)
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Enforce HMAC-only (protección contra "alg: none")
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("método de firma inesperado: %v", token.Method)
            }
            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                return nil, fmt.Errorf("claims inválidos")
            }
            userID, _ := claims["user_id"].(string)
            role, _ := claims["role"].(string)
            uid, err := uuid.Parse(userID)
            if err != nil {
                return nil, fmt.Errorf("user_id inválido: %w", err)
            }
            // SIEMPRE retornar la key real (necesario para verificación)
            if role == "admin" {
                admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
                if err != nil {
                    return nil, fmt.Errorf("admin no encontrado: %w", err)
                }
                return []byte(admin.Key), nil
            }
            psi, err := m.psiRepo.GetByID(c.UserContext(), uid)
            if err != nil {
                return nil, fmt.Errorf("psi no encontrado: %w", err)
            }
            return []byte(psi.Key), nil
        })

        // PASO 2: Verificar que el token sea válido ANTES de inyectar identidad
        if err != nil || !token.Valid {
            return c.Next()  // token inválido → proceder como anónimo
        }

        // PASO 3: Ahora SÍ inyectar identidad (el token ya pasó verificación)
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            return c.Next()
        }
        role, _ := claims["role"].(string)
        userID, _ := claims["user_id"].(string)
        uid, err := uuid.Parse(userID)
        if err != nil {
            return c.Next()
        }

        if role == "admin" {
            if admin, err := m.adminRepo.GetByID(c.UserContext(), uid); err == nil {
                c.Locals("admin", admin)
            }
        } else {
            if psi, err := m.psiRepo.GetByID(c.UserContext(), uid); err == nil {
                c.Locals("psi_user", psi)
            }
        }

        return c.Next()
    }
}
```

**Patrón de verificación post-fix:**
```go
// internal/middleware/auth_test.go
func TestOptionalHybridAuth_RejectsForgedToken(t *testing.T) {
    // Crear token con key incorrecta
    forgedToken := createTestTokenWithKey("victim-uuid", "admin", "wrong-secret-key")

    // Ejecutar middleware
    app := fiber.New()
    app.Get("/test", mockOptionalAuth middleware.OptionalHybridAuth(), func(c *fiber.Ctx) error {
        admin := c.Locals("admin")
        if admin != nil {
            t.Error("admin debería ser nil con token forjado")
        }
        return c.SendStatus(200)
    })

    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer "+forgedToken)
    resp, _ := app.Test(req)

    if resp.StatusCode != 200 {
        t.Errorf("esperaba 200, obtuvo %d", resp.StatusCode)
    }
}
```

**Nota sobre handlers downstream:** Los únicos 2 routes que usan `OptionalHybridAuth` son `GET /posts/` (`ListPosts`) y `GET /posts/:id` (`GetPost`). Ambos **ya usan** comma-ok idiom correctamente y caen a `"public"` cuando no hay sesión. No se necesitan cambios en handlers.

```go
// posts_handler.go — AMBOS handlers ya son seguros:
role := "public"
if _, ok := c.Locals("admin").(*domain.UserAdmin); ok {
    role = "admin"
} else if _, ok := c.Locals("psi_user").(*domain.PsiUserModel); ok {
    role = "psi"
}
```

**Archivos modificados:**
- `internal/middleware/auth.go` — reescribir `OptionalHybridAuth()`
- `internal/middleware/auth_test.go` — agregar test de token forjado

---

### FIX-02: Password hardcodeada `"Colpsi2025!"` en CSV import

| Campo | Valor |
|-------|-------|
| **Hallazgo** | CRIT-02 |
| **Archivo** | `internal/service/psi_service.go:84` |
| **Severidad** | 🔴 CRÍTICO |
| **Riesgo** | Todos los usuarios importados comparten la misma contraseña conocida |

**Código actual (`psi_service.go:84-86`):**
```go
defaultPassword := "Colpsi2025!"
hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
hashedPassword := string(hashedPasswordBytes)
```

**Nota importante:** La versión XLSX (`psi_service_xlsx.go:68`) **ya usa** el patrón seguro:
```go
defaultPassword := utils.GenerateSecureRandomString(12)  // ← correcto
```

**Fix propuesto — Unificar ambas importaciones:**
```go
// psi_service.go:84 — REEMPLAZAR
defaultPassword := utils.GenerateSecureRandomString(16)

hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
if err != nil {
    auditLogger.Printf("❌ FILA %d | Error hasheando password: %v", rowIdx, err)
    failedRecords = append(failedRecords, map[string]string{
        "fila": strconv.Itoa(rowIdx), "nombre": fullName, "error": "Error interno de seguridad",
    })
    continue
}
hashedPassword := string(hashedPasswordBytes)
```

**Cambio en el email de bienvenida (líneas 163-173):**
```go
if psi.ProofOfLife && validEmail {
    go s.mailService.SendEmail(
        psi.Email,
        "Bienvenido(a) a la plataforma COLPSI Carabobo",
        "welcome_psi",
        map[string]interface{}{
            "Name":              psi.FirstName,
            "Email":             psi.Email,
            "Password":          defaultPassword,
            "MustChangePassword": true,  // ← NUEVO: indicar que debe cambiar
        },
    )
}
```

**Modelo — Agregar campo `MustChangePassword`:**
```go
// internal/domain/user.model.go — en PsiUserModel
MustChangePassword bool `gorm:"default:true" json:"-"`
```

**Migración:**
```sql
ALTER TABLE psi_users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT TRUE;
```

**Verificación:** Ejecutar importación CSV y verificar que:
1. Cada usuario tiene contraseña diferente
2. El email indica "debe cambiar contraseña"
3. Login con primera contraseña funciona

**Archivos modificados:**
- `internal/service/psi_service.go` — línea 84, líneas 163-173
- `internal/domain/user.model.go` — agregar campo
- `migrations/` — nueva migración Atlas

---

### FIX-03: JWT HMAC keys en texto plano en DB

| Campo | Valor |
|-------|-------|
| **Hallazgo** | CRIT-03 |
| **Archivos** | `internal/middleware/auth.go:124,219` + `internal/domain/user.model.go:25,72` + `internal/service/admin_service.go:271` |
| **Severidad** | 🔴 CRÍTICO |
| **Riesgo** | DB comprometida → tokens JWT válidos para cualquier usuario |

**Flujo actual (vulnerabilidad):**
```
1. Admin se crea → Key = uuid.NewV7() (UUID en texto plano)
2. Admin hace login → service retorna admin.Key (UUID raw)
3. JWT se firma con []byte(admin.Key) → UUID como HMAC secret
4. JWT se verifica con []byte(admin.Key) → funciona
5. Si atacante obtiene Key de la DB → puede firmar tokens válidos
```

**Fix propuesto — Hashear key con SHA-256 antes de almacenar:**

```go
// internal/service/admin_service.go — en Register/Create
import "crypto/sha256"
import "encoding/hex"

// Al crear usuario:
rawKey := uuid.Must(uuid.NewV7()).String()
hashedKey := sha256.Sum256([]byte(rawKey))
admin.Key = hex.EncodeToString(hashedKey[:])

// La key que se retorna para JWT signing es el hash:
return admin.Key, nil  // ← retorna hash, no UUID raw
```

**Impacto en auth.go (sin cambios necesarios):**
```go
// auth.go:100 — ya funciona porque usa el campo Key tal cual
return []byte(key), nil
// Ahora key es el hash hex, no el UUID raw
```

**Migración SQL para key existentes:**
```sql
-- Script de una sola vez para migrar keys existentes
-- ADVERTENCIA: Ejecutar solo una vez. Todos los tokens activos serán invalidados.
-- SEGURIDAD: Usar regex UUID para prevenir doble-hash si se ejecuta dos veces.
-- UUIDs tienen formato: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
-- Hashes SHA-256 hex son exactamente 64 chars. El WHERE excluye hashes previos.

-- Para user_admins:
UPDATE user_admins
SET key = encode(sha256(key::bytea), 'hex')
WHERE key ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Para psi_users:
UPDATE psi_users
SET key = encode(sha256(key::bytea), 'hex')
WHERE key ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';
```

**Nota sobre salts:** Para mayor seguridad, se podría usar `hmac.New(sha256.New, []byte(salt))` con un salt aleatorio. Pero SHA-256 simple ya elimina el problema principal (UUID en plaintext). El salt se puede agregar como mejora futura.

**Nota sobre idempotencia del migrate:** El regex UUID en el WHERE garantiza que ejecutar el script dos veces sea seguro (segunda ejecución no afecta filas — los hashes de 64 chars no matchean el patrón UUID).

**Archivos modificados:**
- `internal/service/admin_service.go` — función Register
- `internal/service/psi_service.go` — función ImportFromCSV (línea 124)
- `internal/service/psi_service_xlsx.go` — función ImportFromXLSX (línea ~150)
- `migrations/` — script de migración de keys

---

### FIX-04: Tags GORM erróneos

| Campo | Valor |
|-------|-------|
| **Hallazgo** | CRIT-04 |
| **Archivo** | `internal/domain/user.model.go:108-109` |
| **Severidad** | 🔴 CRÍTICO |
| **Riesgo** | GORM panic al generar migración o ejecutar AutoMigrate |

**Código actual:**
```go
// Línea 108: bool con tag de size (inválido)
ShowMunicipalityCarabobo bool   `gorm:"size:255" json:"show_municipality_carabobo"`

// Línea 109: string con default booleano (inválido)
PhoneCarabobo            string `gorm:"default:false" json:"phone_carabobo"`
```

**Fix:**
```go
// Línea 108: bool → default booleano
ShowMunicipalityCarabobo bool   `gorm:"default:false" json:"show_municipality_carabobo"`

// Línea 109: string → size apropiado para teléfono
PhoneCarabobo            string `gorm:"size:20" json:"phone_carabobo"`
```

**Verificación inmediata:**
```bash
# Debe completar sin panic
go run cmd/exp/migrate/main.go > /dev/null
```

**Archivos modificados:**
- `internal/domain/user.model.go` — líneas 108-109

---

### FIX-05: Seed admin credentials expuestas en logs

| Campo | Valor |
|-------|-------|
| **Hallazgo** | CRIT-05 |
| **Archivo** | `pkg/database/seed.go:65` |
| **Severidad** | 🔴 CRÍTICO |
| **Riesgo** | Credenciales del admin SUDO visibles en logs de la aplicación |

**Código actual (`seed.go:65`):**
```go
log.Printf("ℹ️  User: %s | Pass: %s | ID: %s", admin.Username, defaultPass, admin.ID)
```

Donde `defaultPass` es `"admin123"` (definido en línea 21).

**Fix:**
```go
// seed.go:65 — REEMPLAZAR
log.Printf("ℹ️  Admin seed creado — User: %s | ID: %s", admin.Username, admin.ID)
// NUNCA loguear contraseñas, ni siquiera en desarrollo
```

**Acción adicional:** Verificar que no hay otros logs de passwords en el codebase:
```bash
grep -rn "Pass:" pkg/ internal/ --include="*.go" | grep -v "_test.go"
grep -rn "password" pkg/ internal/ --include="*.go" | grep -i "log\|print\|fmt"
```

**Archivos modificados:**
- `pkg/database/seed.go` — línea 65

---

### FIX-06: Hardcoded secrets como defaults en configuración

| Campo | Valor |
|-------|-------|
| **Hallazgo** | CRIT-06 |
| **Archivo** | `internal/config/env.config.go:110-111` |
| **Severidad** | 🔴 CRÍTICO |
| **Riesgo** | Si no se configuran env vars en producción, la API arranca con tokens predecibles |

**Código actual:**
```go
// Línea 110-111
JwtLibrarySecret: getEnv("JWT_LIBRARY_SECRET", "secret-for-psi-library"),
AbsAdminToken:    getEnv("ABS_ADMIN_TOKEN", "deaful_ABS_ADMIN_TOKEN"),
```

Nota: `"deaful"` es un typo de `"default"` — claramente un placeholder de desarrollo.

**Fix propuesto:**
```go
// internal/config/env.config.go — al final de InitConfig()

func InitConfig() {
    // ... carga existente de variables ...

    // ═══ Validación de secrets críticos ═══
    if Envs.Environment == "production" {
        // En producción: FATAL si no están configurados
        if Envs.JwtLibrarySecret == "" || Envs.JwtLibrarySecret == "secret-for-psi-library" {
            log.Fatal("🔴 JWT_LIBRARY_SECRET debe configurarse en producción. " +
                "Use: export JWT_LIBRARY_SECRET=$(openssl rand -hex 64)")
        }
        if Envs.AbsAdminToken == "" || Envs.AbsAdminToken == "deaful_ABS_ADMIN_TOKEN" {
            log.Fatal("🔴 ABS_ADMIN_TOKEN debe configurarse en producción. " +
                "Use: export ABS_ADMIN_TOKEN=$(openssl rand -hex 64)")
        }
    } else {
        // En desarrollo: generar aleatoriamente si no están seteados
        if Envs.JwtLibrarySecret == "" || Envs.JwtLibrarySecret == "secret-for-psi-library" {
            Envs.JwtLibrarySecret = generateRandomSecret(64)
            log.Println("⚠️  JWT_LIBRARY_SECRET generado aleatoriamente (solo desarrollo)")
        }
        if Envs.AbsAdminToken == "" || Envs.AbsAdminToken == "deaful_ABS_ADMIN_TOKEN" {
            Envs.AbsAdminToken = generateRandomSecret(64)
            log.Println("⚠️  ABS_ADMIN_TOKEN generado aleatoriamente (solo desarrollo)")
        }
    }
}

// generateRandomSecret crea un secret cryptográficamente seguro
func generateRandomSecret(length int) string {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        log.Fatalf("Error generando secret aleatorio: %v", err)
    }
    return hex.EncodeToString(bytes)
}
```

**Import necesario:** `"crypto/rand"`, `"encoding/hex"`

**Archivos modificados:**
- `internal/config/env.config.go` — agregar validación + función helper

---

### FIX-07: Envío SMTP deshabilitado en código

| Campo | Valor |
|-------|-------|
| **Hallazgo** | CRIT-07 |
| **Archivo** | `internal/service/mail_service.go:198-200` |
| **Severidad** | 🔴 CRÍTICO |
| **Riesgo** | Usuarios nunca reciben credenciales, notificaciones de login no funcionan |

**Código actual (`mail_service.go:197-201`):**
```go
// comentado por el autor, presumiblemente para entornos de desarrollo/pruebas
// if err := s.client.DialAndSend(m); err != nil {
// 	return fmt.Errorf("fallo la conexión SMTP o el envío: %w", err)
// }
return nil
```

**Fix propuesto:**
```go
// mail_service.go:197-201 — DESCOMENTAR y mejorar
if err := s.client.DialAndSend(m); err != nil {
    log.Printf("⚠️  Error enviando email a %s vía %s: %v", to, s.config.SMTPHost, err)
    return fmt.Errorf("fallo el envío de email: %w", err)
}
return nil
```

**Verificación con MailHog/Mailpit (testing local):**
```yaml
# Agregar a docker-compose.yml para testing:
mailhog:
  image: mailhog/mailhog:latest
  ports:
    - "1025:1025"   # SMTP server
    - "8025:8025"   # Web UI (http://localhost:8025)
  environment:
    - MH_API_BIND_ADDR=0.0.0.0:8025
    - MH_SMTP_BIND_ADDR=0.0.0.0:1025
```

```bash
# Configurar .env para testing:
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
SMTP_FROM=info@colpsicarabobo.com
```

**Archivos modificados:**
- `internal/service/mail_service.go` — líneas 198-200
- `docker-compose.yml` — agregar servicio MailHog (opcional)

---

## Fase 2 — Altos (antes de escalar)

---

### FIX-08: Rate limiting in-memory no sobrevive restarts

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-01 |
| **Archivo** | `internal/middleware/rate_limiter.go` |
| **Severidad** | 🟠 ALTO |
| **Estado** | ✅ IMPLEMENTADO — Commit `64c2457` |
| **Riesgo** | Multi-instancia: cada instancia tiene su propio contador |

**Fix implementado — Valkey con fallback in-memory:**
- Nuevo campo `ValkeyAddr` en config (`VALKEY_ADDR` env var)
- Rate limiter usa `gofiber/storage/valkey` si `VALKEY_ADDR` está configurado
- Panic recovery si Valkey no está disponible → fallback a in-memory
- Nuevo servicio `valkey/valkey:9.1-alpine` en `docker-compose.yml`
- Cero breaking changes: sin `VALKEY_ADDR`, funciona como antes

**Detalle:** Ver `FIX_08_REPORT.md`

**Fix a corto plazo (sin Redis):**
```go
// internal/middleware/rate_limiter.go — agregar comment documentando la limitación

// ╔══════════════════════════════════════════════════════════════════╗
// ║  IMPORTANTE: Este rate limiter es IN-MEMORY.                   ║
// ║  No soporta despliegue multi-instancia (Docker replicas, K8s). ║
// ║  Para escalar, migrar a Redis-backed store.                    ║
// ║  Ver: https://github.com/gofiber/contrib/tree/main/ratelimit   ║
// ╚══════════════════════════════════════════════════════════════════╝
```

**Fix a mediano plazo (cuando se escale):** Migrar a `github.com/gofiber/contrib/ratelimit` con Redis:
```go
import "github.com/redis/go-redis/v9"

// Crear cliente Redis
rdb := redis.NewClient(&redis.Options{Addr: config.Envs.RedisAddr})

// Usar Redis store
app.Use(ratelimit.New(ratelimit.Config{
    Max:        10,
    Expiration: 15 * time.Minute,
    Storage:    ratelimit.NewRedisStore(rdb),
}))
```

**Archivos modificados:**
- `internal/middleware/rate_limiter.go` — documentación

---

### FIX-09: `debug-monitor` sin autenticación en producción

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-02 |
| **Archivo** | `internal/router/admin_router.go:33` |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Expone métricas internas del sistema a cualquiera |

**Código actual:**
```go
// admin_router.go:33 — registrado sin guard de ambiente
router.Get("/debug-monitor", monitor.New(monitor.Config{Title: "DEV ONLY - Monitor"}))
```

**Fix:**
```go
// admin_router.go — envolver con check de ambiente
if config.Envs.Environment == "development" {
    router.Get("/debug-monitor", monitor.New(monitor.Config{
        Title: "DEV ONLY - Monitor",
    }))
}
```

**Alternativa más segura:** Eliminarlo completamente y usar herramientas externas:
```go
// Si se necesita en desarrollo, proteger con auth:
adminGroup.Use(authMid.ProtectedAdmin404())
adminGroup.Get("/debug-monitor", monitor.New(monitor.Config{
    Title: "Admin Monitor",
}))
```

**Archivos modificados:**
- `internal/router/admin_router.go` — línea 33
- Import: `"github.com/veniversvm/ColPsiCarabobo/api/internal/config"`

---

### FIX-10: Panic en startup si SMTP falla

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-03 |
| **Archivo** | `internal/router/psi_router.go:20-23` |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Toda la app muere si SMTP no está disponible |

**Código actual (inconsistente entre archivos):**

```go
// psi_router.go:20-23 — PANIC (mata la app)
mailService, err := service.NewMailService()
if err != nil {
    panic("Error al inicializar el servicio de correo: " + err.Error())
}
```

```go
// admin_router.go:20-23 — GRACEFUL (continúa sin email)
mailSvc, err := service.NewMailService()
if err != nil {
    log.Printf("⚠️  Advertencia: No se pudo conectar al servidor SMTP: %v", err)
}
```

**Fix — Unificar a degradación graciosa:**
```go
// psi_router.go:20-23 — REEMPLAZAR panic por log
mailService, err := service.NewMailService()
if err != nil {
    log.Printf("⚠️  No se pudo inicializar el servicio de correo: %v", err)
    log.Printf("⚠️  El envío de correos estará deshabilitado hasta que SMTP esté disponible")
    // mailService será nil — los handlers que lo usen deben validar
}
```

**Verificar que todos los consumers de `mailService` manejan nil:**
```bash
grep -rn "mailService\." internal/service/ --include="*.go"
```

**Archivos modificados:**
- `internal/router/psi_router.go` — líneas 20-23

---

### FIX-11: Casteo inseguro de `c.Locals()` sin nil-check (16 instancias)

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-04 |
| **Archivos** | 4 handlers, 16 instancias |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Panic runtime si middleware no setea el Local |

**Crear helper functions reutilizables:**

**Nota sobre dependencia circular:** No existe circular dependency entre `handler/` y `middleware/`. `handler/` importa `middleware/` para usar los helpers, pero `middleware/` NO importa `handler/`. `helpers.go` importa solo `domain/` y `fiber`, lo cual es seguro.

```go
// internal/middleware/helpers.go — NUEVO ARCHIVO
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// GetAuthenticatedAdmin extrae el admin de Fiber Locals de forma segura.
// Retorna error 401 si no hay sesión válida.
func GetAuthenticatedAdmin(c *fiber.Ctx) (*domain.UserAdmin, error) {
    admin, ok := c.Locals("admin").(*domain.UserAdmin)
    if !ok || admin == nil {
        return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Sesión administrativa inválida o expirada",
        })
    }
    return admin, nil
}

// GetAuthenticatedPsi extrae el psicólogo de Fiber Locals de forma segura.
// Retorna error 401 si no hay sesión válida.
func GetAuthenticatedPsi(c *fiber.Ctx) (*domain.PsiUserModel, error) {
    psi, ok := c.Locals("psi_user").(*domain.PsiUserModel)
    if !ok || psi == nil {
        return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Sesión de psicólogo inválida o expirada",
        })
    }
    return psi, nil
}

// GetOptionalAdmin extrae el admin si existe (para auth opcional).
// Retorna (nil, nil) si no hay sesión — NO es error.
func GetOptionalAdmin(c *fiber.Ctx) *domain.UserAdmin {
    admin, _ := c.Locals("admin").(*domain.UserAdmin)
    return admin
}

// GetOptionalPsi extrae el psicólogo si existe (para auth opcional).
// Retorna (nil, nil) si no hay sesión — NO es error.
func GetOptionalPsi(c *fiber.Ctx) *domain.PsiUserModel {
    psi, _ := c.Locals("psi_user").(*domain.PsiUserModel)
    return psi
}
```

**Reemplazar cada instancia (ejemplos):**

```go
// ═══ psi_handler.go ═══

// ANTES (Línea 89 — PANIC si nil):
updater := c.Locals("psi_user").(*domain.PsiUserModel)

// DESPUÉS:
updater, err := middleware.GetAuthenticatedPsi(c)
if err != nil {
    return err  // ya retorna error 401
}

// ANTES (Línea 307 — PANIC si nil):
psi := c.Locals("psi_user").(*domain.PsiUserModel)

// DESPUÉS:
psi, err := middleware.GetAuthenticatedPsi(c)
if err != nil {
    return err
}
```

```go
// ═══ psi_user_admin.go ═══

// ANTES (Línea 29):
admin := c.Locals("admin").(*domain.UserAdmin)

// DESPUÉS:
admin, err := middleware.GetAuthenticatedAdmin(c)
if err != nil {
    return err
}
```

```go
// ═══ posts_handler.go ═══ (auth opcional)

// ANTES (Línea 34):
if _, ok := c.Locals("admin").(*domain.UserAdmin); ok {

// DESPUÉS (ya funciona, pero usar helper para consistencia):
if admin := middleware.GetOptionalAdmin(c); admin != nil {
```

**Instancias completas a cambiar:**

| # | Archivo | Línea | Tipo | Fix |
|---|---------|-------|------|-----|
| 1 | `psi_handler.go` | 89 | `psi_user` (requerido) | `GetAuthenticatedPsi` |
| 2 | `psi_handler.go` | 307 | `psi_user` (requerido) | `GetAuthenticatedPsi` |
| 3 | `psi_handler.go` | 339 | `psi_user` (requerido) | `GetAuthenticatedPsi` |
| 4 | `psi_handler.go` | 380 | `psi_user` (requerido) | `GetAuthenticatedPsi` |
| 5 | `psi_handler.go` | 432 | `psi_user` (requerido) | `GetAuthenticatedPsi` |
| 6 | `psi_handler.go` | 454 | `psi_user` (requerido) | `GetAuthenticatedPsi` |
| 7 | `psi_user_admin.go` | 29 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 8 | `psi_user_admin.go` | 80 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 9 | `psi_user_admin.go` | 109 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 10 | `psi_user_admin.go` | 158 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 11 | `psi_user_admin.go` | 186 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 12 | `posts_handler.go` | 80 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 13 | `posts_handler.go` | 147 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 14 | `specialty_handler.go` | 83 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 15 | `specialty_handler.go` | 113 | `admin` (requerido) | `GetAuthenticatedAdmin` |
| 16 | `specialty_handler.go` | 141 | `admin` (requerido) | `GetAuthenticatedAdmin` |

**Archivos modificados:**
- `internal/middleware/helpers.go` — NUEVO
- `internal/handler/psi_handler.go` — 6 puntos
- `internal/handler/psi_user_admin.go` — 5 puntos
- `internal/handler/posts_handler.go` — 2 puntos
- `internal/handler/specialty_handler.go` — 3 puntos

---

### FIX-12: Comparación de errores por string matching

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-05 |
| **Archivos** | `psi_handler.go:116,413`, `psi_user_admin.go:41`, `specialty_handler.go:91` |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Cambio de texto de error rompe handlers silenciosamente |

**Crear sentinel errors:**

```go
// internal/domain/errors.go — NUEVO ARCHIVO
package domain

import "errors"

// ═══ Errores de Autenticación ═══
var (
    ErrPasswordIncorrect  = errors.New("contraseña actual incorrecta")
    ErrInvalidCredentials = errors.New("credenciales inválidas")
    ErrAccountInactive    = errors.New("cuenta desactivada")
)

// ═══ Errores de Autorización ═══
var (
    ErrPermissionDenied   = errors.New("no tienes permiso para editar este registro")
    ErrInsufficientPerms  = errors.New("permisos insuficientes")
    ErrAdminRequired      = errors.New("se requiere acceso administrativo")
)

// ═══ Errores de Recursos ═══
var (
    ErrPsiNotFound        = errors.New("psicólogo no encontrado")
    ErrAdminNotFound      = errors.New("administrador no encontrado")
    ErrPostNotFound       = errors.New("publicación no encontrada")
    ErrSpecialtyNotFound  = errors.New("especialidad no encontrada")
)

// ═══ Errores de Negocio ═══
var (
    ErrUniqueViolation    = errors.New("registro duplicado")
    ErrEmailExists        = errors.New("el email ya está registrado")
    ErrCedulaExists       = errors.New("la cédula ya está registrada")
    ErrFPVExists          = errors.New("el número FPV ya está registrado")
    ErrMaxSocialNetworks  = errors.New("límite máximo de redes sociales alcanzado")
    ErrSudoCount          = errors.New("no se puede eliminar el último administrador con permisos SUDO")
)
```

**Fix en handlers:**

```go
// ANTES (psi_handler.go:116):
if err.Error() == "contraseña actual incorrecta" {

// DESPUÉS:
if errors.Is(err, domain.ErrPasswordIncorrect) {
```

```go
// ANTES (psi_handler.go:413):
if err.Error() == "no tienes permiso para editar este registro" {

// DESPUÉS:
if errors.Is(err, domain.ErrPermissionDenied) {
```

```go
// ANTES (psi_user_admin.go:41):
if err.Error() == "psicólogo no encontrado" {

// DESPUÉS:
if errors.Is(err, domain.ErrPsiNotFound) {
```

```go
// ANTES (specialty_handler.go:91):
if strings.Contains(err.Error(), "permiso") || strings.Contains(err.Error(), "rango") {

// DESPUÉS:
if errors.Is(err, domain.ErrInsufficientPerms) || errors.Is(err, domain.ErrPermissionDenied) {
```

**Fix en services (retornar sentinel errors):**
```bash
# Encontrar todos los return fmt.Errorf que deben cambiarse:
grep -rn 'return fmt.Errorf("' internal/service/ --include="*.go" | head -20
```

```go
// ANTES (en cualquier service):
return fmt.Errorf("contraseña actual incorrecta")

// DESPUÉS:
return domain.ErrPasswordIncorrect
```

**Archivos modificados:**
- `internal/domain/errors.go` — NUEVO
- `internal/handler/psi_handler.go` — puntos 116, 413
- `internal/handler/psi_user_admin.go` — punto 41
- `internal/handler/specialty_handler.go` — punto 91
- `internal/service/psi_service.go` — múltiples returns
- `internal/service/admin_service.go` — múltiples returns
- `internal/service/post_service.go` — múltiples returns
- `internal/service/specialty_service.go` — múltiples returns

---

### FIX-13: Logger GORM en modo Info en producción

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-06 |
| **Archivo** | `pkg/database/postgres.go:45` |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Logs excesivos + posible filtración de PII en queries SQL |

**Código actual:**
```go
// postgres.go:45
Logger: logger.Default.LogMode(logger.Info),
```

**Fix:**
```go
// postgres.go — importar config
import "github.com/veniversvm/ColPsiCarabobo/api/internal/config"

// En ConnectDB():
var logLevel logger.LogLevel
switch config.Envs.Environment {
case "production":
    logLevel = logger.Warn  // Solo errores y warnings lentos
case "development":
    logLevel = logger.Info  // Queries completas
default:
    logLevel = logger.Warn
}

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logLevel),
    // ... resto de config
})
```

**Archivos modificados:**
- `pkg/database/postgres.go` — línea 45

---

### FIX-14: Cookie analytics sin `Secure: true`

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-07 |
| **Archivo** | `internal/middleware/analytics.go:82-88` |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Cookie enviada por HTTP en texto plano |

**Código actual:**
```go
// analytics.go:82-88
c.Cookie(&fiber.Cookie{
    Name:     "_sid",
    Value:    sessionID,
    Expires:  time.Now().Add(365 * 24 * time.Hour),  // ← 365 días es excesivo
    HTTPOnly: true,
    SameSite: "Lax",
    // Falta: Secure: true
})
```

**Fix:**
```go
// analytics.go — importar config
import "github.com/veniversvm/ColPsiCarabobo/api/internal/config"

c.Cookie(&fiber.Cookie{
    Name:     "_sid",
    Value:    sessionID,
    Expires:  time.Now().Add(30 * 24 * time.Hour),  // ← reducir a 30 días
    HTTPOnly: true,
    Secure:   config.Envs.Environment == "production", // ← solo HTTPS en prod
    SameSite: "Lax",
})
```

**Archivos modificados:**
- `internal/middleware/analytics.go` — líneas 82-88

---

### FIX-15: Pool de conexiones DB no configurado

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-08 |
| **Archivo** | `pkg/database/postgres.go` |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Pool ilimitado puede agotar conexiones de PostgreSQL bajo carga |

**Código actual:**
```go
// postgres.go — después de gorm.Open, NO hay configuración de pool
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{...})
// ... retorna db inmediatamente
```

**Fix:**
```go
// postgres.go — después de gorm.Open
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    PrepareStmt: false,
    SkipDefaultTransaction: false,
    Logger: logger.Default.LogMode(logLevel),
})
if err != nil {
    return nil, fmt.Errorf("error al conectar con la base de datos: %w", err)
}

// ═══ Configurar pool de conexiones ═══
sqlDB, err := db.DB()
if err != nil {
    return nil, fmt.Errorf("error al obtener underlying DB: %w", err)
}

sqlDB.SetMaxOpenConns(10)               // Máximo 10 conexiones abiertas
sqlDB.SetMaxIdleConns(5)                // Mantener 5 en idle
sqlDB.SetConnMaxLifetime(1 * time.Hour) // Reciclar conexiones cada hora

log.Printf("✅ Pool de conexiones configurado: MaxOpen=%d, MaxIdle=%d, MaxLifetime=1h",
    10, 5)

return db, nil
```

**Archivos modificados:**
- `pkg/database/postgres.go` — agregar configuración pool

---

### FIX-16: Passwords en templates de email en texto plano

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-09 |
| **Archivos** | `internal/templates/welcome_psi.html:16`, `internal/templates/welcome_admin.html:15` |
| **Severidad** | 🟠 ALTO |
| **Estado** | ✅ IMPLEMENTADO |
| **Riesgo** | Contraseña temporal en texto plano en email sin indicar que debe cambiarse |

**Fix implementado:**
- `{{.TempPassword}}` → `{{.Password}}` (corregido bug funcional: password renderizaba vacío)
- Agregado aviso amarillo: "IMPORTANTE: Esta es una contraseña temporal..."
- Nuevo campo `MustChangePassword bool` en `Credentials` struct
- Marcado automático `true` en imports masivos + CreatePsiByAdmin
- Login retorna `must_change_password` en JSON response
- Validación `IsStrongPassword()` agregada en CreatePsiByAdmin
- Migración: `ALTER TABLE` para ambas tablas

**Detalle:** Ver `FIX_16_REPORT.md`

**Código actual (templates):**
```html
<!-- welcome_psi.html:16 -->
<p>Password : <strong>{{.TempPassword}}</strong></p>

<!-- welcome_admin.html:15 -->
<p>Password : <strong>{{.TempPassword}}</strong></p>
```

**Fix a corto plazo — Agregar indicación de cambio obligatorio:**
```html
<!-- welcome_psi.html — ACTUALIZAR -->
<div style="background-color: #fff3cd; border: 1px solid #ffc107; padding: 12px; border-radius: 6px; margin: 16px 0;">
    <p style="margin: 0; color: #856404;">
        <strong>Su contraseña temporal es:</strong> <code style="font-size: 16px;">{{.TempPassword}}</code>
    </p>
</div>
<p style="color: #dc3545; font-weight: bold;">
    IMPORTANTE: Debe cambiar su contraseña en el primer inicio de sesión por seguridad.
</p>
```

**Fix a mediano plazo — Link de establecimiento de contraseña:**
```go
// Generar token de una sola vez
resetToken := utils.GenerateSecureRandomString(32)
resetLink := fmt.Sprintf("https://colpsicarabobo.com/establecer-contrasena?token=%s", resetToken)

// Guardar en DB con expiración (24 horas)
// Enviar link en vez de contraseña
go s.mailService.SendEmail(
    psi.Email,
    "Establezca su contraseña - COLPSI Carabobo",
    "welcome_psi",
    map[string]interface{}{
        "Name":      psi.FirstName,
        "ResetLink": resetLink,
        "ExpiresIn": "24 horas",
    },
)
```

**Archivos modificados:**
- `internal/templates/welcome_psi.html` — actualizar HTML
- `internal/templates/welcome_admin.html` — actualizar HTML
- `internal/domain/user.model.go` — agregar campos `PasswordResetToken`, `PasswordResetExpires` (fix mediano plazo)

---

### FIX-17: S3 keys expuestas en respuestas API

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-10 |
| **Archivo** | `internal/domain/user.model.go` (múltiples campos) |
| **Severidad** | 🟠 ALTO |
| **Estado** | ✅ IMPLEMENTADO |
| **Riesgo** | Estructura del bucket S3 expuesta al cliente |

**Fix implementado:**
- Nuevo método `S3Client.GetPublicURL(key)` construye URL completa (`endpoint/bucket/key`)
- `PsiService.ResolvePsiModelURLs()` convierte todas las S3 keys de un modelo
- `PostService.resolvePostURLs()` convierte ImageS3Key de posts
- Wrapper nil-safe `publicURL()` para compatibilidad con tests
- 6 endpoints actualizados: directory, profile, admin view, /me, posts list, post detail

**Detalle:** Ver `FIX_17_REPORT.md`

**Campos a proteger:**

```go
// ═══ user.model.go ═══

// Línea 90 — PsiUserModel
// ANTES:
ProfilePictureS3Key string `gorm:"size:512" json:"profile_picture_url"`
// DESPUÉS:
ProfilePictureS3Key string `gorm:"size:512" json:"-"`
ProfilePictureURL   string `gorm:"-" json:"profile_picture_url"` // campo virtual

// ═══ user.model.go (PsiUserColData) ═══

// Líneas 179-181
// ANTES:
TitleImageOneS3Key   string `gorm:"size:512" json:"title_image_one_url"`
TitleImageTwoS3Key   string `gorm:"size:512" json:"title_image_two_url"`
TitleImageThreeS3Key string `gorm:"size:512" json:"title_image_three_url"`
// DESPUÉS:
TitleImageOneS3Key   string `gorm:"size:512" json:"-"`
TitleImageTwoS3Key   string `gorm:"size:512" json:"-"`
TitleImageThreeS3Key string `gorm:"size:512" json:"-"`
TitleImageOneURL     string `gorm:"-" json:"title_image_one_url"`
TitleImageTwoURL     string `gorm:"-" json:"title_image_two_url"`
TitleImageThreeURL   string `gorm:"-" json:"title_image_three_url"`

// ═══ user.model.go (PsiUserPostGrade) ═══

// Líneas 255-257
// ANTES:
PicOneS3Key   string `gorm:"size:512" json:"pic_one_url"`
PicTwoS3Key   string `gorm:"size:512" json:"pic_two_url"`
PicThreeS3Key string `gorm:"size:512" json:"pic_three_url"`
// DESPUÉS:
PicOneS3Key   string `gorm:"size:512" json:"-"`
PicTwoS3Key   string `gorm:"size:512" json:"-"`
PicThreeS3Key string `gorm:"size:512" json:"-"`
PicOneURL     string `gorm:"-" json:"pic_one_url"`
PicTwoURL     string `gorm:"-" json:"pic_two_url"`
PicThreeURL   string `gorm:"-" json:"pic_one_url"`

// ═══ text.model.go (Post) ═══

// Línea 64
// ANTES:
ImageS3Key string `gorm:"size:512" json:"image_url"`
// DESPUÉS:
ImageS3Key string `gorm:"size:512" json:"-"`
ImageURL   string `gorm:"-" json:"image_url"`
```

**Poblar campos virtuales en service/repository:**
```go
// En el service, después de obtener el registro:
if psi.ProfilePictureS3Key != "" {
    psi.ProfilePictureURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
        config.Envs.S3Bucket, config.Envs.S3Region, psi.ProfilePictureS3Key)
}
```

**Archivos modificados:**
- `internal/domain/user.model.go` — múltiples campos
- `internal/domain/text.model.go` — campo ImageS3Key
- `internal/service/psi_service.go` — poblar URLs virtuales
- `internal/service/post_service.go` — poblar URLs virtuales

---

### FIX-18: `ON DELETE NO ACTION` en todas las Foreign Keys

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-11 |
| **Archivo** | `migrations/20260604165811_init.sql` |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Hard delete en psi_users deja registros huérfanos en tablas relacionadas |

**Análisis de cada FK y fix recomendado:**

| FK Constraint | Tabla hija → padre | Relación | Fix |
|---------------|---------------------|----------|-----|
| `fk_psi_users_col_data` | col_data → psi_users | 1:1 | `ON DELETE CASCADE` |
| `fk_psi_users_post_grades` | post_grades → psi_users | N:1 | `ON DELETE CASCADE` |
| `fk_psi_users_social_networks` | social_networks → psi_users | N:1 | `ON DELETE CASCADE` |
| `fk_psi_users_solvencies` | solvency → psi_users | N:1 | `ON DELETE CASCADE` |
| `fk_posts_text` | posts → text_models | N:1 | `ON DELETE CASCADE` |
| `fk_psi_users_full_bio` | psi_users → text_models | N:1 | `ON DELETE SET NULL` |

**Migración Atlas:**
```sql
-- migrations/YYYYMMDD_security_fixes.sql

-- 1. col_data → psi_users (1:1 cascade)
ALTER TABLE psi_user_col_data
    DROP CONSTRAINT IF EXISTS fk_psi_users_col_data,
    ADD CONSTRAINT fk_psi_users_col_data
        FOREIGN KEY (psi_user_model_id)
        REFERENCES psi_users(id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE;

-- 2. post_grades → psi_users (N:1 cascade)
ALTER TABLE psi_user_post_grades
    DROP CONSTRAINT IF EXISTS fk_psi_users_post_grades,
    ADD CONSTRAINT fk_psi_users_post_grades
        FOREIGN KEY (psi_user_id)
        REFERENCES psi_users(id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE;

-- 3. social_networks → psi_users (N:1 cascade)
ALTER TABLE psi_user_social_networks
    DROP CONSTRAINT IF EXISTS fk_psi_users_social_networks,
    ADD CONSTRAINT fk_psi_users_social_networks
        FOREIGN KEY (psi_user_id)
        REFERENCES psi_users(id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE;

-- 4. solvency → psi_users (N:1 cascade)
ALTER TABLE psi_user_solvency
    DROP CONSTRAINT IF EXISTS fk_psi_users_solvencies,
    ADD CONSTRAINT fk_psi_users_solvencies
        FOREIGN KEY (psi_user_model_id)
        REFERENCES psi_users(id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE;

-- 5. posts → text_models (N:1 cascade)
ALTER TABLE posts
    DROP CONSTRAINT IF EXISTS fk_posts_text,
    ADD CONSTRAINT fk_posts_text
        FOREIGN KEY (text_id)
        REFERENCES text_models(id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE;

-- 6. psi_users → text_models (bio FK, set null)
ALTER TABLE psi_users
    DROP CONSTRAINT IF EXISTS fk_psi_users_full_bio,
    ADD CONSTRAINT fk_psi_users_full_bio
        FOREIGN KEY (bio_text_id)
        REFERENCES text_models(id)
        ON UPDATE NO ACTION
        ON DELETE SET NULL;
```

**Archivos modificados:**
- `migrations/` — nueva migración Atlas

---

### FIX-19: Doble idempotencia global + custom

| Campo | Valor |
|-------|-------|
| **Hallazgo** | HIGH-12 |
| **Archivos** | `cmd/api/main.go:156` + `internal/middleware/idempotency.go:112` |
| **Severidad** | 🟠 ALTO |
| **Riesgo** | Comportamiento indefinido por conflicto de capas |

**Capa 1 — Global (`main.go:156`):**
```go
app.Use(idempotency.New(idempotency.Config{
    Lifetime:  30 * time.Minute,
    KeyHeader: "X-Idempotency-Key",
}))
```

**Capa 2 — Custom per-route (`psi_router.go:42`):**
```go
adminGroup.Post(
    "/create",
    middleware.UserScopedIdempotency(idempotencyStore, 30*time.Minute),
    h.CreatePsiByAdmin,
)
```

**Fix — Eliminar la capa global (la custom es superior):**

```go
// cmd/api/main.go — ELIMINAR líneas 156-159:
// ╔══════════════════════════════════════════════════════════════╗
// ║ ELIMINADO: Idempotencia global de Fiber                     ║
// ║ Razón: Conflicta con UserScopedIdempotency custom.          ║
// ║ La implementación custom es superior porque:                ║
// ║   1. Es scoped por usuario (SHA-256 del body + userId)      ║
// ║   2. Previene duplicados por usuario, no por key header     ║
// ║   3. Usa store persistente (sync.RWMutex + map)             ║
// ╚══════════════════════════════════════════════════════════════╝
```

```go
// cmd/api/main.go — ELIMINAR estas líneas:
// app.Use(idempotency.New(idempotency.Config{
//     Lifetime:  30 * time.Minute,
//     KeyHeader: "X-Idempotency-Key",
// }))
```

**Archivos modificados:**
- `cmd/api/main.go` — eliminar líneas 156-159

---

## Fase 3 — Medios (deuda técnica)

---

### FIX-20: Split `PsiService` God Object (~1860 líneas)

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-01 |
| **Archivo** | `internal/service/psi_service.go` (1439 líneas, 18+ métodos) |
| **Estado** | ✅ IMPLEMENTADO — Split en 7 archivos por dominio funcional |

**Implementación: 7 archivos + tests (ver `FIX_20_REPORT.md`)**

| Nuevo servicio | Métodos a migrar | Responsabilidad |
|---------------|------------------|-----------------|
| `PsiAuthService` | `Login`, `LoginLibrary`, `Logout` | Autenticación y sesiones |
| `PsiProfileService` | `GetPublicProfile`, `GetPublicDirectory`, `UpdateProfileSelf`, `GetMe`, `GetPsiBioByID`, `GetPsiSOlvency`, `GetSitemapPsis`, `GetAdminDirectory`, `GetPsiByIDAdmin` | Perfil y directorio |
| `PsiImportService` | `ImportFromCSV`, `ImportFromXLSX`, `CreatePsiByAdmin`, `UpdatePsiByAdmin`, `DeletePsiByAdmin` | Importación y administración |
| `PsiAcademicService` | `AddPostGrade`, `UpdatePostGrade`, `AddSocialNetwork`, `UpdateSocialNetwork`, `DeleteSocialNetwork` | Académico y redes sociales |

**Cada nuevo servicio recibe las mismas dependencias:**
```go
type PsiAuthService struct {
    repo        domain.PsiUserRepository
    mailService IMailService
}

type PsiProfileService struct {
    repo        domain.PsiUserRepository
    s3Client    *s3.S3Client
    sanitizer   *bluemonday.Policy
}
```

**Archivos a crear:**
- `internal/service/psi_auth_service.go`
- `internal/service/psi_profile_service.go`
- `internal/service/psi_import_service.go`
- `internal/service/psi_academic_service.go`

**Archivos a modificar:**
- `internal/router/psi_router.go` — inyectar nuevos servicios
- `cmd/api/main.go` — instanciar nuevos servicios

---

### FIX-21: `mapDBErrorToHuman` duplicada

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-02 |
| **Archivos** | `psi_service.go:230` vs `error_mapper.go:18` |

**Fix:**
```bash
# 1. Eliminar la función local en psi_service.go (líneas 230-261)
# 2. Reemplazar todas las llamadas:
grep -n "mapDBErrorToHuman" internal/service/psi_service.go

# Reemplazar cada:
humanError := mapDBErrorToHuman(err)

# Por:
humanError := MapDBError(err).Error()
```

**Archivos modificados:**
- `internal/service/psi_service.go` — eliminar función duplicada, actualizar llamadas

---

### FIX-22: `println()` nativo en producción

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-03 |
| **Archivo** | `internal/service/psi_service.go:632` |

**Fix:**
```go
// ANTES:
println("debug:", someVar)

// DESPUÉS:
log.Printf("DEBUG: %v", someVar)
// O eliminar si es debug temporal
```

**Archivos modificados:**
- `internal/service/psi_service.go` — línea 632

---

### FIX-23: AnalyticsService acoplado a `*gorm.DB`

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-04 |
| **Archivo** | `internal/service/analytics_service.go:22` |

**Fix:** Crear interfaz y repository:
```go
// internal/domain/analytics_repository.go
type AnalyticsRepository interface {
    RecordPageView(view PageView) error
    RecordProfileView(view ProfileView) error
    RecordSearchEvent(event SearchEvent) error
    RecordLoginEvent(event LoginEvent) error
    GetDashboardStats() (map[string]interface{}, error)
}

// internal/repository/postgres/analytics_repository.go
type AnalyticsRepoImpl struct { db *gorm.DB }

// Actualizar AnalyticsService para recibir la interfaz
type AnalyticsService struct {
    repo domain.AnalyticsRepository
}
```

**Archivos a crear:**
- `internal/domain/analytics_repository.go`
- `internal/repository/postgres/analytics_repository.go`

**Archivos a modificar:**
- `internal/service/analytics_service.go`
- `internal/router/` — inyectar repository

---

### FIX-24: `fmt.Printf` de debug en repositorio

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-05 |
| **Archivos** | `psi_repository.go:329,344` |

**Fix:**
```bash
# Encontrar todos los fmt.Printf en repos:
grep -rn "fmt.Printf\|fmt.Println" internal/repository/ --include="*.go"
```

Eliminar o reemplazar con `log.Printf` condicional:
```go
if config.Envs.Environment == "development" {
    log.Printf("DEBUG repo: %s", query)
}
```

---

### FIX-25: FormFile ignorado sin error handling

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-06 |
| **Archivos** | `psi_handler.go:96`, `posts_handler.go:87` |

**Fix:** Ya se maneja correctamente (archivo opcional). Solo agregar logging para debugging:
```go
file, err := c.FormFile("image")
if err != nil && !errors.Is(err, fiber.ErrNotFound) {
    log.Printf("⚠️  Error leyendo archivo image: %v", err)
}
```

---

### FIX-26: Logging de PII

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-07 |
| **Archivo** | `psi_handler.go:229-230` |

**Fix:**
```go
// ANTES (psi_handler.go:229-230):
log.Printf("----- PSI Profile ----\n")
log.Printf("%v", psi)

// DESPUÉS:
log.Printf("PSI Profile requested: ID=%s, Username=%s", psi.ID, psi.Username)
// NUNCA loguear el objeto completo (contiene email, teléfono, dirección, etc.)
```

---

### FIX-27: Status codes como enteros crudos

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-08 |
| **Archivos** | Múltiples handlers |

**Fix:** Reemplazar enteros por constantes Fiber:
```go
// ANTES:
return c.Status(400).JSON(...)
return c.Status(403).JSON(...)
return c.Status(500).JSON(...)

// DESPUÉS:
return c.Status(fiber.StatusBadRequest).JSON(...)
return c.Status(fiber.StatusForbidden).JSON(...)
return c.Status(fiber.StatusInternalServerError).JSON(...)
```

```bash
# Encontrar todas las instancias:
grep -rn 'c.Status([0-9]' internal/handler/ --include="*.go"
```

---

### FIX-28: `LoginLibrary` sin analytics ni Swagger

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-09 |
| **Archivo** | `psi_handler.go:281` |

**Fix:** Agregar anotaciones Swagger y llamada a analytics:
```go
// LoginLibrary godoc
// @Summary      Login para Biblioteca (Audiobookshelf)
// @Description  Autentica y retorna token SSO para el microservicio de biblioteca.
// @Tags         Psicólogos - Auth
// @Accept       json
// @Produce      json
// @Param        request body request_structs.PsiLoginRequest true "Credenciales"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /psi/login-library [post]
func (h *PsiHandler) LoginLibrary(c *fiber.Ctx) error {
    // ... body existente ...

    // Agregar analytics
    h.analytics.RecordLogin(user.ID, user.Username, "psi_library", c.IP(), c.Get("User-Agent"))

    return c.JSON(...)
}
```

---

### FIX-29: Repos instanciados múltiples veces

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-10 |
| **Archivos** | `psi_router.go`, `post_router.go`, `admin_router.go` |

**Fix:** Instanciar repos una sola vez en `main.go` y pasarlos:
```go
// cmd/api/main.go — instanciar una vez
adminRepo := postgres.NewUserAdminRepository(db)
psiRepo := postgres.NewPsiUserRepository(db)
postRepo := postgres.NewPostRepository(db)
specialtyRepo := postgres.NewSpecialtyRepository(db)

// Pasar a InitRouter
router.InitRouter(app, adminRepo, psiRepo, postRepo, specialtyRepo, s3Client, analyticsService)
```

---

### FIX-30: `Save()` para updates parciales (sobreescribe zero-values)

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-11 |
| **Archivos** | `user_admin_repo.go:73`, `specialty_repo.go:103` |

**Fix:**
```go
// ANTES (sobreescribe campos no enviados con zero-values):
db.Save(&admin)

// DESPUÉS (solo actualiza campos no-zero):
db.Model(&admin).Updates(map[string]interface{}{
    "email":    admin.Email,
    "username": admin.Username,
    // ... solo campos que realmente cambiaron
})
```

---

### FIX-31: Búsqueda de especialidad por nombre, no por FK

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-12 |
| **Archivos** | `psi_repository.go:537-542` + `domain/user.model.go` + `migrations/` |
| **Complejidad** | ALTA — requiere 7 pasos de migración |

**Contexto:** `psi_users` NO tiene columna `specialty_id`. Las especialidades se almacenan como strings en `primary_work_area` y `secondary_work_area`. No hay FK a `specialties`. El fix requiere agregar la FK y migrar datos.

**Fix — 7 pasos:**

**Paso 1:** Agregar campo `SpecialtyID` al modelo:
```go
// internal/domain/user.model.go — en PsiUserModel
SpecialtyID *uint32 `gorm:"column:specialty_id" json:"specialty_id,omitempty"`
```

**Paso 2:** Migración SQL para agregar columna:
```sql
-- Paso 2a: Agregar columna nullable
ALTER TABLE psi_users ADD COLUMN specialty_id INTEGER;

-- Paso 2b: Crear tabla specialties si no existe
CREATE TABLE IF NOT EXISTS specialties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Paso 2c: Migrar datos existentes (mapear strings → IDs)
INSERT INTO specialties (name)
SELECT DISTINCT primary_work_area FROM psi_users
WHERE primary_work_area IS NOT NULL AND primary_work_area != ''
ON CONFLICT (name) DO NOTHING;

UPDATE psi_users p
SET specialty_id = s.id
FROM specialties s
WHERE p.primary_work_area = s.name;

-- Paso 2d: Agregar FK constraint
ALTER TABLE psi_users
    ADD CONSTRAINT fk_psi_users_specialty
    FOREIGN KEY (specialty_id) REFERENCES specialties(id)
    ON DELETE SET NULL;
```

**Paso 3:** Refactorizar repository para usar FK:
```go
// ANTES (psi_repository.go:537-542):
WHERE psi.primary_work_area = ?

// DESPUÉS:
WHERE psi.specialty_id = ?
```

**Paso 4:** Actualizar handlers de importación CSV/XLSX para resolver nombre → ID.

**Paso 5:** Actualizar `specialty_handler.go` para retornar `specialty_id` en listados.

**Paso 6:** Frontend: cambiar de filtrar por nombre a filtrar por `specialty_id`.

**Paso 7:** (Opcional futuro) Eliminar columnas `primary_work_area`/`secondary_work_area` una vez verificada la migración.

---

### FIX-32: `BoolFromForm()` duplicada

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-13 |
| **Archivos** | `psi_user.go:117` vs `utils/geo_venezuela.go:144` |

**Fix:** Eliminar la versión duplicada en `psi_user.go` y usar `utils.BoolFromForm()`.

---

### FIX-33: Test `SanitizeImage_Defensive` roto

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-14 |
| **Archivos** | `utils_test.go:28` vs `image_sanitizer.go:86` |

**Problema concreto:** El test en `utils_test.go:28` busca el substring `"invalid image"` pero el handler real en `image_sanitizer.go:86` retorna `"invalid image format"`.

```go
// utils_test.go:28 — test busca:
if !strings.Contains(err.Error(), "invalid image") {
    // ^^ substring match parcial — funciona

// image_sanitizer.go:86 — handler retorna:
return nil, fmt.Errorf("invalid image format")
```

**Fix:** El test actualmente pasa (porque `"invalid image"` es substring de `"invalid image format"`). Sin embargo, el mensaje de error es demasiado genérico. Mejorar ambos:
```go
// image_sanitizer.go:86 — hacer más específico:
return nil, fmt.Errorf("image validation failed: unsupported format")

// utils_test.go:28 — actualizar test:
if !strings.Contains(err.Error(), "unsupported format") {
    t.Errorf("esperaba error con 'unsupported format', obtuvo: %v", err)
}
```

**Verificación:** `go test ./internal/utils/ -v -run SanitizeImage_Defensive`

---

### FIX-34: `runtime.GOMAXPROCS(2)` hardcodeado

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-15 |
| **Archivo** | `cmd/api/main.go:43` |

**Fix:** Reemplazar `runtime.GOMAXPROCS(2)` por `runtime.GOMAXPROCS(0)` (que equivale a `runtime.NumCPU()`). Go 1.25+ usa automáticamente todos los cores disponibles, pero ser explícito con `0` es más claro que simplemente eliminar la línea.

```go
// ANTES:
runtime.GOMAXPROCS(2)

// DESPUÉS:
runtime.GOMAXPROCS(0) // 0 = usar todos los cores disponibles
```

**Nota Docker:** En Docker/Kubernetes, `runtime.NumCPU()` lee `cpu.cfs_quota_us` del cgroup, así que el valor será correcto incluso dentro de un container con CPU limit.

---

### FIX-35: Sin graceful shutdown

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-16 |
| **Archivos** | `cmd/api/main.go:187` + `internal/service/mail_service.go` + `internal/router/admin_router.go` + `internal/router/psi_router.go` |

**Problema adicional — MailService instanciado 2 veces:** `MailService` se crea en `admin_router.go:20` Y en `psi_router.go:20`, resultando en 2 pools SMTP separados y 2 workers de retry. Debe ser un singleton.

**Fix — MailService singleton + Close():**

```go
// internal/service/mail_service.go — agregar:

// Singleton instance
var mailServiceInstance *MailService
var mailServiceOnce sync.Once

func NewMailServiceSingleton(config MailConfig) *MailService {
    mailServiceOnce.Do(func() {
        mailServiceInstance = NewMailService(config)
    })
    return mailServiceInstance
}

// Close cierra el worker de retry y el ticker
func (s *MailService) Close() {
    if s.cancelRetry != nil {
        s.cancelRetry()
    }
    log.Println("📧 MailService closed")
}
```

```go
// internal/service/mail_service.go — MailService struct debe tener:
type MailService struct {
    smtpHost     string
    smtpPort     string
    smtpUser     string
    smtpPass     string
    senderEmail  string
    templateDir  string
    retryQueue   chan EmailJob
    maxRetries   int
    cancelRetry  context.CancelFunc  // ← NUEVO
}
```

**Fix — Graceful shutdown en main.go:**
```go
// cmd/api/main.go — al final de main()

// ═══ Graceful Shutdown ═══
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

<-quit

log.Println("🛑 Shutting down server...")

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := app.ShutdownWithContext(ctx); err != nil {
    log.Fatalf("❌ Server forced to shutdown: %v", err)
}

// Cerrar MailService (flush cola de retry)
mailService.Close()

// Cerrar pool de conexiones
sqlDB, err := db.DB()
if err == nil {
    sqlDB.Close()
}

log.Println("✅ Server exited gracefully")
```

**Fix — Unificar instanciación en main.go:**
```go
// cmd/api/main.go — crear MailService UNA vez, pasar a ambos routers:
mailService := service.NewMailServiceSingleton(service.MailConfig{
    Host:       cfg.SMTPHost,
    Port:       cfg.SMTPPort,
    Username:   cfg.SMTPUsername,
    Password:   cfg.SMTPPassword,
    FromEmail:  cfg.SMTPFromEmail,
    TemplateDir: "templates",
})

adminRouter.RegisterAdminRoutes(app, adminHandler, middleware, mailService)
psiRouter.RegisterPsiRoutes(app, psiHandler, middleware, mailService)
```

**Imports necesarios:** `"os"`, `"os/signal"`, `"syscall"`, `"context"`, `"sync"`

---

### FIX-36: `PsiObservations` y `PsiODeontologia` sin relación en padre

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-17 |
| **Archivo** | `internal/domain/user.model.go` |

**Fix:**
```go
// user.model.go — en PsiUserModel, agregar:
Observations  []PsiObservations  `gorm:"foreignKey:PsiUserID" json:"-"`
Deontologia   []PsiODeontologia  `gorm:"foreignKey:PsiUserID" json:"-"`
```

---

### FIX-37: `phone_carabobo` es `text DEFAULT 'false'` en migración

| Campo | Valor |
|-------|-------|
| **Hallazgo** | MED-18 |
| **Archivo** | `migrations/20260604165811_init.sql:262` |

**Fix:** Corregido por FIX-04 (tags GORM). Generar nueva migración Atlas que corrija el tipo de columna.

---

## Fase 4 — Bajos (polish)

---

### FIX-38: Typos en nombres de archivos

| Hallazgo | Fix | Comando |
|----------|-----|---------|
| LOW-01: `radom_string.go` | Renombrar a `random_string.go` | `git mv internal/utils/radom_string.go internal/utils/random_string.go` |
| LOW-02: `post_respository.go` | Renombrar a `post_repository.go` | `git mv internal/repository/postgres/post_respository.go internal/repository/postgres/post_repository.go` |

---

### FIX-39: Typos en código

| Hallazgo | Ubicación | Fix |
|----------|-----------|-----|
| LOW-03 | `PsiUSerSolvency` (user.model.go:214) | Renombrar a `PsiUserSolvency` |
| LOW-04 | `"emial invalido"` (admin_service.go:379) | Corregir a `"email inválido"` |
| LOW-05 | Log "BIO" al recuperar Solvencies (psi_user_admin.go:55) | Corregir a "Solvencies" |
| LOW-06 | Comentario "Empleado público" en Discapacity (user.model.go:196) | Corregir a "Discapacidad" |

---

### FIX-40: `Post` sin `TableName()` explícito

| Campo | Valor |
|-------|-------|
| **Hallazgo** | LOW-07 |

**Fix:**
```go
// text.model.go — agregar
func (Post) TableName() string { return "posts" }
```

---

### FIX-41: `GraduationYear` es `string` en vez de `int`

| Campo | Valor |
|-------|-------|
| **Hallazgo** | LOW-08 |

**Fix:**
```go
// user.model.go:250
// ANTES:
GraduationYear string `gorm:"size:50" json:"post_grade_graduation_year"`
// DESPUÉS:
GraduationYear int `json:"post_grade_graduation_year"`
```

**Nota:** Requiere migración Atlas + actualizar service que parsea el año.

---

### FIX-42: Inconsistencia de UUID generation

| Campo | Valor |
|-------|-------|
| **Hallazgo** | LOW-09 |

**Fix:** Estandarizar en `uuidv7()` (más eficiente que `gen_random_uuid()` para PKs):
```sql
-- En la definición de tablas de analytics, cambiar:
-- ANTES: default:gen_random_uuid()
-- DESPUÉS: default:uuidv7()
```

---

### FIX-43: `context.TODO()` en `ConnectS3()`

| Campo | Valor |
|-------|-------|
| **Hallazgo** | LOW-10 |

**Fix:**
```go
// pkg/s3/s3.go
// ANTES:
func InitS3() *S3Client {
    cfg, err := config.LoadDefaultConfig(context.TODO())

// DESPUÉS:
func InitS3(ctx context.Context) *S3Client {
    cfg, err := config.LoadDefaultConfig(ctx)
```

---

### FIX-44: `GetPresignedURL()` comentado con `/*`

| Campo | Valor |
|-------|-------|
| **Hallazgo** | LOW-11 |
| **Archivo** | `pkg/s3/upload.go:79-92` |

**Fix:** Implementar o eliminar:
```go
// Si se necesita, implementar con AWS SDK v2:
func (s *S3Client) GetPresignedURL(key string, expiry time.Duration) (string, error) {
    presignClient := s3.NewPresignClient(s.client)
    output, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    }, func(o *s3.PresignOptions) {
        o.Expires = expiry
    })
    if err != nil {
        return "", err
    }
    return output.URL, nil
}

// Si no se necesita: eliminar el código comentado
```

---

### FIX-45: Log con emojis en producción

| Campo | Valor |
|-------|-------|
| **Hallazgo** | LOW-12 |

**Fix (a largo plazo):** Migrar a logger estructurado:
```go
// Reemplazar log.Printf("✅ ...") por:
import "github.com/rs/zerolog/log"

log.Info().Str("component", "database").Msg("Connected to PostgreSQL")
log.Warn().Str("component", "config").Msg("No .env file found")
log.Error().Err(err).Str("component", "s3").Msg("Failed to connect")
```

---

### FIX-46: `log.Println("[DEBUG AUTH]")` en producción

| Campo | Valor |
|-------|-------|
| **Hallazgo** | LOW-13 |
| **Archivo** | `internal/middleware/auth.go:66,71,80` |

**Fix:**
```go
// ANTES:
log.Println("[DEBUG AUTH] ...")

// DESPUÉS (opción 1 — eliminar):
// (simplemente borrar la línea)

// DESPUÉS (opción 2 — conditional):
if config.Envs.Environment == "development" {
    log.Printf("[DEBUG AUTH] token method: %v", token.Method)
}
```

---

### FIX-47: `UserAdmin` y `PsiUserModel` comparten campos duplicados

| Campo | Valor |
|-------|-------|
| **Hallazgo** | LOW-14 |

**Fix (opcional, refactor futuro):** Crear struct embebido:
```go
// internal/domain/credentials.go — NUEVO
type Credentials struct {
    Username string `gorm:"size:25;unique;not null" json:"username"`
    Email    string `gorm:"size:255;unique;not null" json:"email"`
    Password string `gorm:"size:512;not null" json:"-"`
    Key      string `gorm:"size:512;" json:"-"`
}

// En UserAdmin:
type UserAdmin struct {
    ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
    AuditModel
    Credentials
    // ... permisos específicos
}

// En PsiUserModel:
type PsiUserModel struct {
    ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
    AuditModel
    Credentials
    // ... datos del psicólogo
}
```

**Nota:** Este fix requiere migración y es de bajo impacto. Hacer solo si se refactoriza el código.

---

## Rollback Plan

Si algún fix causa problemas en producción, cada fase se puede revertir independientemente:

| Fase | Fix | Rollback |
|------|-----|----------|
| 1a | FIX-04 (GORM tags) | Revertir cambios en `user.model.go`, regenerar migración Atlas |
| 1b | FIX-05 (seed log) | Revertir 1 línea en `seed.go` |
| 1c | FIX-06 (secrets) | Restaurar defaults hardcodeados (temporal, hasta configurar env vars) |
| 2a | FIX-01 (auth) | Revertir `auth.go` al código anterior, re-deploy |
| 2b | FIX-02 (password) | La columna `must_change_password` no causa daño si se deja, revertir solo el service |
| 2c | FIX-11 (helpers) | Los helpers son adicionales, revertir imports en handlers |
| 3 | FIX-07/18/19 | **ATENCIÓN:** Las migraciones SQL son irreversibles. Mantener backup de DB antes de ejecutar cualquier migración |
| 4 | FIX-20+ | Cada fix es independiente, revertir el archivo específico |

**Regla general:** Siempre hacer backup de la base de datos antes de ejecutar migraciones SQL (FIX-03, FIX-07, FIX-18, FIX-19, FIX-31).

---

## Regression Tests

Tests que deben pasar después de CADA fase de implementación para prevenir regressions:

```bash
# ═══ Fase 1 — Compilación y GORM ═══
go build ./cmd/api
go test ./internal/domain/... -v  # GORM tags

# ═══ Fase 2 — Auth ═══
go test ./internal/middleware/... -v  # OptionalHybridAuth + helpers
go test ./internal/handler/... -v     # nil-check en handlers

# ═══ Fase 3 — Migraciones ═══
# Ejecutar migración en DB de staging
atlas migrate apply --env local
# Verificar integridad referencial
psql -c "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_type='FOREIGN KEY';"

# ═══ Fase 4 — Full suite ═══
go test ./... -v -count=1
go vet ./...
```

**Tests nuevos a crear:**
- `TestOptionalHybridAuth_RejectsForgedToken` — FIX-01
- `TestOptionalHybridAuth_SideEffectBeforeValidation` — FIX-01
- `TestGetAuthenticatedAdmin_NilLocal` — FIX-11
- `TestGetAuthenticatedPsi_NilLocal` — FIX-11
- `TestMailService_Singleton` — FIX-35
- `TestSanitizeImage_Defensive` (corregido) — FIX-33

---

## Verificación Final

### Checklist post-implementación

```bash
# ═══ 1. Compilar sin errores ═══
go build ./cmd/api

# ═══ 2. Ejecutar tests ═══
go test ./... -v

# ═══ 3. Verificar que no hay panic en startup ═══
go run ./cmd/api &
sleep 5
curl -s http://localhost:8080/api/v1/specialties | head -c 200
kill %1

# ═══ 4. Verificar tags GORM ═══
go run cmd/exp/migrate/main.go > /dev/null 2>&1 && echo "✅ GORM OK" || echo "❌ GORM FAIL"

# ═══ 5. Verificar que OptionalHybridAuth valida firma ═══
go test ./internal/middleware/ -v -run TestOptional

# ═══ 6. Verificar que no se loguean passwords ═══
grep -rn "Pass:" pkg/database/seed.go
# Debe retornar vacío

# ═══ 7. Verificar que secrets no son predecibles ═══
grep -rn "secret-for-psi-library\|deaful_ABS" internal/config/
# Debe retornar vacío

# ═══ 8. Verificar que S3 keys no aparecen en JSON tags ═══
grep -rn 'json:".*s3\|json:".*url"' internal/domain/ | grep -v 'json:"-"'
# Debe retornar vacío (o solo campos virtuales)

# ═══ 9. Verificar que no hay panic points ═══
grep -rn 'c.Locals(".*").*(.*domain\.' internal/handler/ --include="*.go" | grep -v "ok :="
# Debe retornar vacío (todos deben usar ok-check)

# ═══ 10. Swagger actualizado ═══
swag init -g cmd/api/main.go -o docs/

# ═══ 11. Migración Atlas generada ═══
atlas migrate diff security_fixes --env local
atlas migrate apply --env local

# ═══ 12. Verificar que no hay fmt.Printf de debug ═══
grep -rn "fmt.Printf\|fmt.Println\|println(" internal/ --include="*.go" | grep -v "_test.go" | grep -v "//"
# Revisar cada resultado — solo debug temporal debe eliminarse
```

### Métricas de éxito

| Métrica | Antes | Después |
|---------|-------|---------|
| Hallazgos críticos | 7 | 0 |
| Hallazgos altos | 12 | 0 |
| Panic risk points | 16 | 0 |
| Hardcoded secrets | 3 | 0 |
| String error comparisons | 4+ | 0 |
| Tests unitarios | ~220 líneas | >500 líneas |

---

## Orden de Implementación Recomendado

| Paso | Fixes | Dependencias | Tiempo est. |
|------|-------|--------------|-------------|
| **1a** | FIX-04 (GORM tags) | Ninguna — desbloquea startup | 5 min |
| **1b** | FIX-05 (seed log) | Ninguna | 5 min |
| **1c** | FIX-06 (hardcoded secrets) | Ninguna | 30 min |
| **2a** | FIX-01 (OptionalHybridAuth) | Requiere refactor de auth | 2 hrs |
| **2b** | FIX-02 (password hardcodeada) | Campo MustChangePassword | 1 hr |
| **2c** | FIX-11 (nil-check locals) | Crear helpers en middleware | 2 hrs |
| **3** | FIX-10, FIX-09, FIX-13, FIX-14, FIX-15 | Independientes | 2 hrs |
| **4** | FIX-12 (sentinel errors) | Requiere domain/errors.go | 2 hrs |
| **5** | FIX-07, FIX-18, FIX-19 | Migraciones Atlas | 2 hrs |
| **6** | Fase 3 completa (FIX-20 a FIX-37) | Deuda técnica | 1-2 días |
| **7** | Fase 4 completa (FIX-38 a FIX-47) | Polish | 2-3 hrs |

**Tiempo total estimado:** 3-5 días para Fase 1+2, 1-2 semanas para todas las fases.

---

## Estado Actual — Key Lifecycle Management (25 Julio 2026)

**Commit:** `b511d43` | **Reporte completo:** `KEY_LIFECYCLE_REPORT.md`

### Completado

| Componente | Archivos | Estado |
|------------|----------|--------|
| UUID v7 en producción | 4 archivos (`admin_service.go`, `seed.go`, `s3/upload.go`, `analytics.go`) | ✅ 0 `uuid.New()` restantes |
| Empty key guard | `auth.go:130,260` | ✅ Admin 404 + Psi 401, sin crypto innecesaria |
| Admin logout | `admin_handler.go`, `admin_router.go`, `admin_service.go`, `user_admin_repo.go` | ✅ `POST /admin/logout` |
| PsiUser logout | `psi_service.go:1148` | ✅ key → `""` (eliminación, no rotación) |
| Cleanup job | `cmd/cleanup/main.go`, `pkg/job/key_cleanup.go` | ✅ Binario independiente, tick 30min |
| Tests nuevos | `pkg/job/key_cleanup_test.go` | ✅ 4 tests, todos pasan |
| Admin mock actualizado | `admin_service_test.go` | ✅ `UpdateKey` agregado |
| CRUDAdmin key rotation | `admin_service.go:406` | ✅ Ya usa `uuid.Must(uuid.NewV7())` |

### Pendiente

| Fix | Descripción | Prioridad |
|-----|-------------|-----------|
| CRIT-03 (hashing) | Hashear keys con SHA-256 antes de almacenar (seguridad DB comprometida) | Alta |
| docker-compose cleanup | Agregar servicio `cleanup` al `docker-compose.yml` | Media |
| Tests integración logout | Tests E2E para endpoints de logout | Media |

### Métricas actualizadas

| Métrica | Antes | Después |
|---------|-------|---------|
| `uuid.New()` en producción | 4 | **0** |
| `uuid.NewString()` en producción | 1 | **0** |
| Rutas de logout | 1 (solo PsiUser) | **2** (Admin + PsiUser) |
| Binarios | 1 (`cmd/api`) | **2** (`cmd/api` + `cmd/cleanup`) |
| Tests unitarios | ~15 | **~19** (+4 en pkg/job) |

---

*Plan original generado por análisis cruzado del reporte de auditoría contra el código fuente real (59 archivos .go).*
*Actualizado el 25 de Julio de 2026 tras implementación de Key Lifecycle Management.*
*Fecha: 24 de Julio, 2026*

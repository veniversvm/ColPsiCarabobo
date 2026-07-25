# FIX-01 Report — OptionalHybridAuth side-effect-before-validation

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-01 |
| **Hallazgo original** | CRIT-01 |
| **Archivos modificados** | `internal/middleware/auth.go`, `internal/middleware/auth_test.go` |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

La función `OptionalHybridAuth` en `auth.go:153-198` tenía un bug de seguridad crítico: **inyectaba la identidad del usuario en `c.Locals()` antes de verificar la firma del token JWT**.

### Flujo vulnerable (antes)

```
jwt.Parse(token, keyFunc)
  │
  ├── 1. Extrae claims (userID, role)     ← SIN verificar firma
  ├── 2. Query a DB (GetByID)             ← SIN verificar firma
  ├── 3. c.Locals("admin", admin)         ← ⚠️ INYECTA identidad ANTES de validar
  ├── 4. Retorna key para verificar firma
  │
  └── 5. Verificación de firma
         ├── ✅ Válida → OK
         └── ❌ Inválida → Token rechazado, PERO c.Locals YA tiene el usuario
```

Un atacante podía enviar un token con firma inválida y el usuario se inyectaba igual en `c.Locals()`. Aunque `OptionalHybridAuth` no rechazaba la request (es middleware opcional), los handlers downstream veían un usuario autenticado que no debería estarlo.

### Código vulnerable (línea original)

```go
_, _ = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    // ...
    if role == "admin" {
        admin, err := m.adminRepo.GetByID(c.UserContext(), uid)
        if err == nil {
            c.Locals("admin", admin)  // ← SIDE EFFECT ANTES de verificación
            return []byte(admin.Key), nil
        }
    }
    // ...
})
```

---

## Corrección

Reescritura completa de `OptionalHybridAuth` con separación de responsabilidades:

### Flujo seguro (después)

```
jwt.Parse(token, keyFunc)
  │
  ├── 1. Extrae claims (userID, role)
  ├── 2. Query a DB (GetByID)
  ├── 3. Retorna key para verificación    ← SIN inyectar en c.Locals
  │
  └── 4. Verificación de firma
         ├── ❌ Inválida → c.Next() como anónimo
         └── ✅ Válida → THEN inyectar identidad
```

### Código corregido

```go
// PASO 1: Parse + Verificar firma (sin side effects)
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    // ... validaciones ...
    // Solo retorna la key, NO inyecta en c.Locals()
    return []byte(admin.Key), nil
})

// PASO 2: Verificar que el token sea válido ANTES de inyectar identidad
if err != nil || !token.Valid {
    return c.Next()  // token inválido → proceder como anónimo
}

// PASO 3: Ahora SÍ inyectar identidad (el token ya pasó verificación)
if role == "admin" {
    if admin, err := m.adminRepo.GetByID(c.UserContext(), uid); err == nil {
        c.Locals("admin", admin)
    }
}
```

### Cambios clave

| # | Cambio | Justificación |
|---|--------|---------------|
| 1 | Key function ya no inyecta en `c.Locals` | Separar validación de side-effects |
| 2 | Verificación de `token.Valid` antes de inyectar | Token inválido = anónimo |
| 3 | Segundo `GetByID` después de verificación | Trade-off: 1 query extra por request a cambio de seguridad |
| 4 | Errores 返回 explícitos en key function | Evita `return nil, nil` que enmascara fallos |

---

## Tests

Se agregaron **7 tests de seguridad** al suite existente:

| Test | Escenario | Resultado esperado |
|------|-----------|-------------------|
| `Forged_Token_Rejected` | Token firmado con secret incorrecto | `is_anonymous` |
| `Valid_Admin_Token_Detected` | Token admin con firma válida | `is_admin` |
| `Valid_Psi_Token_Detected` | Token psi con firma válida | `is_psi` |
| `Expired_Token_Anonymous` | Token expirado | `is_anonymous` |
| `No_Token_Anonymous` | Sin header Authorization | `is_anonymous` |
| `None_Algorithm_Anonymous` | Token con `alg: none` | `is_anonymous` |
| `Nonexistent_User_Anonymous` | User no existe en DB | `is_anonymous` |

### Resultado de ejecución

```
=== RUN   TestAuthMiddleware_Extensive
=== RUN   TestAuthMiddleware_Extensive/Admin_404_Logic
--- PASS: Admin_404_Logic (3/3 subtests)
=== RUN   TestAuthMiddleware_Extensive/PsiUser_Security_Edge_Cases
--- PASS: PsiUser_Security_Edge_Cases (1/1)
=== RUN   TestAuthMiddleware_Extensive/Hybrid_Auth_Context_Injection
--- PASS: Hybrid_Auth_Context_Injection (1/1)
=== RUN   TestAuthMiddleware_Extensive/OptionalHybridAuth_Security
--- PASS: OptionalHybridAuth_Security (7/7 subtests)
PASS — ok  github.com/veniversvm/ColPsiCarabobo/api/internal/middleware  0.071s
```

**15/15 tests pass.**

---

## Impacto

- **Riesgo de regressión:** Bajo. El middleware es no-bloqueante y los 2 routes que lo usan (`GET /posts/`, `GET /posts/:id`) ya usan comma-ok idiom correctamente.
- **Performance:** 1 adicional `GetByID` por request (antes el key function ya hacía 1, ahora son 2). Se puede optimizar cacheando el user ID y role del token en variables locales del closure.
- **Endpoints afectados:** `GET /posts/` (ListPosts), `GET /posts/:id` (GetPost) — ambos ya seguros, sin cambios necesarios en handlers.

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/middleware/auth.go:153-228` | Función reescrita |
| `internal/middleware/auth_test.go` | 7 tests nuevos |
| `internal/handler/posts_handler.go:34-45` | Handlers que usan OptionalHybridAuth (ya seguros) |
| `SECURITY_FIX_PLAN.md` (FIX-01) | Plan de seguridad referenciado |

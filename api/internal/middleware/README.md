Entendido. Vamos a profundizar al máximo nivel de detalle técnico. Esta documentación está diseñada para que cualquier ingeniero senior o auditor de sistemas entienda la mecánica interna de cada línea de código en la capa de **Middlewares**.

---

# 🛡️ Documentación Técnica: Capa de Middlewares

[⬅ Volver al Índice Principal](../../README.md)

Esta capa constituye el **Perímetro de Seguridad y Telemetría** de la API. Aquí se gestionan las reglas de tráfico, la identidad y la integridad de los datos antes de que las peticiones toquen la lógica de negocio.

---

## 📊 1. Módulo: Analíticas de Tráfico (`analytics.go`)

**Archivo:** `internal/middleware/analytics.go`
**Propósito:** Capturar el comportamiento de los usuarios para Business Intelligence (BI) sin degradar la experiencia de usuario (UX).

### 🛠️ Desglose de Funciones

#### `shouldSkip(path string) bool`

* **Propósito:** Actúa como un filtro de ruido inicial.
* **Lógica:** Recorre el slice `skipPaths` y compara el prefijo de la ruta actual. Si coincide con rutas de sistema (como `/health` o `/metrics`), retorna `true`.
* **Variables:**
  * `skipPaths`: Lista negra de rutas que no generan valor analítico.

#### `AnalyticsMiddleware(db *gorm.DB) fiber.Handler`

* **Propósito:** Punto de entrada principal para el rastreo de visitas.
* **Lógica:**
  1. Llama a `c.Next()` para procesar la petición primero.
  2. Valida si la petición fue exitosa (2xx) y es de tipo `GET`.
  3. Extrae el `sessionID` desde la cookie `_sid`. Si no existe, genera un nuevo UUID y lo planta en el navegador del usuario con flags de seguridad (`HTTPOnly`, `SameSite: Lax`).
  4. Copia los datos del contexto de Fiber a variables locales (evita colisiones de memoria en hilos asíncronos).
  5. Dispara una **Goroutine** que consulta si el usuario ya registró actividad en la ventana de tiempo definida; si no, inserta un registro en la tabla `page_views`.
* **Variables:**
  * `visitWindow` (Constante): 30 minutos. Define el tiempo de "sesión única".
  * `sessionID`: UUID que vincula múltiples peticiones a un mismo dispositivo.
  * `path`, `method`, `ip`, `referer`: Metadatos capturados del request.

---

## 🔐 2. Módulo: Gestión de Identidad - IAM (`auth.go`)

**Archivo:** `internal/middleware/auth.go`
**Propósito:** Validar la legitimidad de las sesiones mediante criptografía JWT y proveer contexto de usuario a los controladores.

### 🛠️ Desglose de Funciones

#### `NewAuthMiddleware(a, p, analytics) *AuthMiddleware`

* **Propósito:** Constructor del objeto de middleware. Inyecta dependencias de repositorios y servicios para que el middleware pueda consultar la base de datos durante la validación.

#### `jwtError(c *fiber.Ctx, status int, message string) error`

* **Propósito:** Estandarizar las respuestas de error de seguridad. Asegura que el cliente reciba un JSON uniforme ante fallos de token.

#### `validateToken(c *fiber.Ctx, getKeyFunc) (*jwt.Token, error)`

* **Propósito:** El motor de validación criptográfica.
* **Lógica:** Extrae el token de la cabecera `Authorization`. Valida que el método de firma sea `HMAC`. Decodifica los `claims` para obtener el `user_id`. Finalmente, llama a `getKeyFunc` para obtener la clave secreta específica de ese usuario y validar la firma.

#### `ProtectedAdmin404() fiber.Handler`

* **Propósito:** Proteger rutas de staff administrativo.
* **Lógica:** Valida el token contra el repositorio de administradores.
* **Estrategia:** Si el token es inválido o el usuario no es admin, retorna `404 Not Found` en lugar de `401`. Esto oculta la existencia de rutas sensibles a atacantes. Inyecta el objeto `admin` en `c.Locals`.

#### `OptionalHybridAuth() fiber.Handler`

* **Propósito:** Proveer "identidad opcional" en rutas públicas.
* **Lógica:** Si hay un token, intenta identificar al usuario (admin o psi) e inyectarlo en el contexto. Si falla o no hay token, no bloquea la petición, permitiendo que el controlador decida qué mostrar.

#### `ProtectedPsiUser() fiber.Handler`

* **Propósito:** Bloquear acceso a rutas privadas del psicólogo.
* **Lógica:** Similar a la protección de admin, pero valida contra el repositorio de psicólogos. Si falla, retorna `401 Unauthorized`. Inyecta el objeto `psi_user` en `c.Locals`.

---

## 🔄 3. Módulo: Resiliencia e Idempotencia (`idempotency.go`)

**Archivo:** `internal/middleware/idempotency.go`
**Propósito:** Garantizar que operaciones críticas (como cobros o registros) no se ejecuten dos veces si el cliente reintenta la petición.

### 🛠️ Desglose de Funciones

#### `NewIdempotencyStore() *IdempotencyStore`

* **Propósito:** Inicializa el mapa de memoria y arranca la rutina de limpieza.
* **Variables:** `entries`: Mapa donde la llave es el hash de la petición y el valor es la respuesta cacheada.

#### `cleanup()`

* **Propósito:** Prevenir el agotamiento de memoria (OOM).
* **Lógica:** Un `Ticker` despierta cada 10 minutos para eliminar del mapa todas las entradas cuya fecha `expires` sea anterior a la actual.

#### `get(key string)` / `set(key string, ...)`

* **Propósito:** Métodos de acceso al mapa protegidos por `sync.RWMutex`.
* **Lógica:** `get` usa un bloqueo de lectura compartido; `set` usa un bloqueo de escritura exclusivo para garantizar la consistencia en entornos multihilo.

#### `UserScopedIdempotency(store, ttl) fiber.Handler`

* **Propósito:** Middleware que intercepta la petición.
* **Lógica:** Lee la cabecera `X-Idempotency-Key`. Si existe, genera una llave compuesta con el ID del usuario. Si la llave está en el `store`, devuelve la respuesta cacheada instantáneamente. Si no, procesa la petición y guarda el resultado en el `store` antes de responder.

#### `scopeKey(userID, rawKey) string`

* **Propósito:** Blindar el caché contra accesos cruzados.
* **Lógica:** Genera un `SHA-256` combinando el ID del usuario y la llave enviada por el cliente. Esto asegura que la llave "123" del Usuario A no devuelva la respuesta del Usuario B.

---

## 🚦 4. Módulo: Control de Tráfico (`rate_limiter.go`)

**Archivo:** `internal/middleware/rate_limiter.go`
**Propósito:** Mitigar ataques de Denegación de Servicio (DoS) y Fuerza Bruta.

### 🛠️ Desglose de Funciones

#### `AuthRateLimiter()`

* **Propósito:** Proteger el login de usuarios.
* **Configuración:** Permite 10 peticiones por cada 15 minutos. Usa la IP del cliente como llave.
* **Lógica:** Si se excede el límite, retorna un `429 Too Many Requests`. Ignora peticiones que no sean `POST` para no bloquear la navegación normal.

#### `AdminAuthRateLimiter()`

* **Propósito:** Hardening extremo para el panel Staff.
* **Configuración:** Restringe a 5 intentos cada 30 minutos. La política es más agresiva porque el compromiso de una cuenta admin es crítico.

---

## 🔬 5. Módulo: Pruebas de Seguridad (`auth_test.go`)

**Archivo:** `internal/middleware/auth_test.go`
**Propósito:** Validar que los middlewares de autenticación sean impenetrables.

### 🛠️ Desglose de Funciones

#### `generateTestToken(userID, role, secret, expiresAt)`

* **Propósito:** Helper de test para crear JWTs sintéticos con cualquier configuración (expirados, roles falsos, etc.) para probar las reacciones del middleware.

#### `TestAuthMiddleware_Extensive(t *testing.T)`

* **Propósito:** Suite de pruebas principal.
* **Casos de prueba:**
  1. **Admin_404_Logic:** Valida que un admin inexistente o token expirado resulte en 404.
  2. **Wrong_Signing_Method:** Intenta saltarse la seguridad usando el ataque "None Algorithm".
  3. **Hybrid_Auth:** Valida que la inyección de contexto funcione correctamente en rutas mixtas.

---

## 🔗 Navegación de Módulos

- [Ir a la Raíz de Documentación](../../README.md)
- [Explorar Lógica de Negocio (Services) ➡](../service/README.md)
- [Explorar Controladores (Handlers) ➡](../handler/README.md)

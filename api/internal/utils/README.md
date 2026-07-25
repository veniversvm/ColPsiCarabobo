# 🔧 Utilidades (utils/)

Funciones puras de utilidad sin dependencias de lógica de negocio. Todas son **stateless**, no acceden a DB, y pueden ser usadas desde cualquier capa de la aplicación.

## Funciones

### 1. `christ.go` — ASCII Art Logo

Branding en consola. Muestra el logo de ColPsi al iniciar el servidor.

```
   ╔═══════════════════════════════════╗
   ║          ColPsi Carabobo          ║
   ║    Sistema de Gestión Psicológica ║
   ╚═══════════════════════════════════╝
```

---

### 2. `clean_alpha_numeric.go` — Limpieza de Texto

Elimina todos los caracteres que no sean alfanuméricos. Implementado a nivel **rune** para ser seguro con Unicode.

```go
CleanAlphaNumeric("¡Hola, Mundo! 123") → "HolaMundo123"
CleanAlphaNumeric("María José © 2024") → "MaríaJos2024"
```

**Por qué rune-based:** Los strings en Go son bytes, pero Unicode puede usar múltiples bytes por carácter. Trabajar con runes garantiza que emojis, tildes y caracteres especiales se manejan correctamente.

---

### 3. `geo_venezuela.go` — Geografía Venezolana

Normalización de nombres de municipios y estados de Venezuela. Incluye helper `BoolFromForm` para convertir valores de formulario a booleano.

```go
NormalizeGeographicLocation("caracas") → "Caracas"
NormalizeGeographicLocation("MARACAIBO") → "Maracaibo"
BoolFromForm("true") → true
BoolFromForm("1") → true
BoolFromForm("") → false
```

**Casos de uso:**
- Normalizar input de usuario antes de guardar en DB
- Convertir checkboxes de formularios HTML a booleanos Go

---

### 4. `image_sanitizer.go` — Sanitización de Imágenes

Compresión y validación de imágenes subidas por usuarios.

**Pipeline:**
```
Imagen subida → Validar magic bytes → Verificar tamaño (5MB max) → Aplanar transparencia → Comprimir a WebP
```

**Funcionalidades:**
- **Compresión a WebP**: Reduce tamaño manteniendo calidad visual
- **Aplanar transparencia**: Convierte fondo transparente a blanco (necesario para profiles)
- **Validación por magic bytes**: No confía en la extensión del archivo, verifica los bytes reales del header
- **Límite de 5MB**: Rechaza archivos más grandes antes de procesar

**Seguridad:**
```go
// Malicious: "photo.jpg" but actually an executable
// Magic bytes check catches this → REJECTED
```

---

### 5. `no_empty_req.go` — Detección de Requests Vacíos

Usa **reflect** para detectar si un struct está completamente vacío. Previene actualizaciones accidentales que sobreescriben datos con valores vacíos.

```go
type UpdateProfile struct {
    Name  string
    Email string
    Phone string
}

IsRequestEmpty(UpdateProfile{}) → true
IsRequestEmpty(UpdateProfile{Name: "Juan"}) → false
```

**Por qué importante:** Sin esta validación, un `PATCH` con body `{}` sobreescribiría todos los campos del perfil con strings vacíos.

---

### 6. `normalize_platform_name.go` — Nombres de Plataformas

Normaliza abreviaturas y variantes de redes sociales a su nombre completo.

| Input | Output |
|-------|--------|
| `ig` | Instagram |
| `instagram` | Instagram |
| `fb` | Facebook |
| `facebook` | Facebook |
| `tw` | Twitter |
| `twitter` | Twitter |
| `yt` | YouTube |
| `youtube` | YouTube |
| `in` | LinkedIn |
| `linkedin` | LinkedIn |
| `tt` | TikTok |
| `tiktok` | TikTok |
| `wa` | WhatsApp |
| `whatsapp` | WhatsApp |
| `tg` | Telegram |
| `telegram` | Telegram |

---

### 7. `radom_string.go` — Strings Aleatorios Seguros

Generación de strings aleatorios usando `crypto/rand` (no `math/rand`).

```go
GenerateRandomString(16) → "k8Jd3mXp2nQ7vR4t"
GenerateRandomString(32) → "aB3cD5eF7gH9iJ1kL3mN5oP7qR9sT1uV"
```

**Seguridad:** Usa `crypto/rand` que es criptográficamente seguro. Adecuado para tokens, IDs de sesión, nonces, etc. No es predecible aunque se conozca el patrón de generación.

---

### 8. `secure_password.go` — Validación de Contraseñas

Validación de fortaleza de contraseñas con requisitos mínimos.

**Requisitos:**
- Mínimo 8 caracteres
- Al menos 1 mayúscula
- Al menos 1 minúscula
- Al menos 1 dígito
- Al menos 1 carácter especial (!@#$%^&* etc.)

```go
ValidatePasswordStrength("Abc123!") → true
ValidatePasswordStrength("abc123")  → false (sin mayúscula ni especial)
ValidatePasswordStrength("ABCD")    → false (sin minúscula ni dígito)
```

---

### 9. `validate_email.go` — Validación de Email

Validación de formato de email usando expresión regular.

```go
ValidateEmail("user@example.com") → true
ValidateEmail("invalid-email")    → false
ValidateEmail("@missing.com")     → false
```

---

### 10. `utils_test.go` — Tests

Suite de tests que cubre todas las funciones de utilidad.

```
TestCleanAlphaNumeric
TestNormalizeGeographicLocation
TestBoolFromForm
TestIsRequestEmpty
TestNormalizePlatformName
TestGenerateRandomString
TestValidatePasswordStrength
TestValidateEmail
```

## Patrón: Funciones Puras

Todas las funciones en utils siguen el principio de **funciones puras**:
- Sin efectos secundarios
- Sin acceso a base de datos
- Sin estado global
- Mismos inputs → mismos outputs
- Fácilmente testables

## Seguridad

| Función | Nota de Seguridad |
|---------|-------------------|
| `image_sanitizer.go` | Magic bytes validation, no confía en extensión |
| `radom_string.go` | `crypto/rand`, no `math/rand` |
| `secure_password.go` | Requisitos mínimos de fortaleza |
| `clean_alpha_numeric.go` | Rune-based para Unicode safety |

## Uso en otras capas

```
utils/ ← service layer (validaciones, transformaciones)
utils/ ← handler layer (input sanitization)
utils/ ← middleware (analytics, auth helpers)
```

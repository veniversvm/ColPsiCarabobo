
Este es el `README.md` detallado para el módulo de estructuras de petición y respuesta del proyecto COLPSI Carabobo.

---

# Request Structs & DTOs Module

## 📌 Descripción General

El paquete `request_structs` actúa como la **Capa de Presentación** y la **Primera Línea de Defensa** de la API. Su propósito fundamental es implementar el patrón de **Capa Anticorrupción (ACL)**, asegurando que los datos introducidos por los usuarios sean validados, normalizados y sanitizados antes de interactuar con la lógica de negocio (Dominio) o la base de datos (Persistencia).

Este módulo garantiza que cambios en el esquema de la base de datos no afecten los contratos públicos de la API y viceversa.

---

## 🏗️ Principios de Diseño

1. **Semántica PATCH Real:** Se utiliza el uso extensivo de **punteros (`*string`, `*bool`, `*int`)**. En Go, esto es crucial para distinguir entre:
   * `nil`: El campo no fue enviado (No actualizar).
   * `valor`: El campo fue enviado con un dato nuevo (Actualizar).
2. **Privacidad por Diseño (Privacy by Design):** Los DTOs están segmentados por roles. Un psicólogo no puede ver ni editar campos administrativos sensibles (ej. estatus de solvencia) a través de su propio DTO de actualización.
3. **Soporte Multipart/Form-Data:** Para permitir la carga de imágenes (WebP) junto con datos de texto, el sistema utiliza campos `Raw` (strings) y métodos `Getter` para parsear booleanos complejos que los frameworks de Go a menudo no detectan en formularios multipart.
4. **Sanitización Unicode:** Toda entrada de texto libre pasa por un proceso de limpieza que soporta caracteres internacionales (ñ, tildes) pero descarta caracteres de control y símbolos SQL peligrosos.

---

## 📂 Estructura del Módulo

### 1. Gestión de Psicólogos (`psi_user.go`, `psi_user_Admin_requests.go`)

Define cómo fluye la información de los agremiados en tres niveles:

* **Creación (Admin):** Registro masivo o manual con campos de identidad legal.
* **Autogestión (Psicólogo):** Permite al usuario editar su perfil y configurar su **Privacy Shield** (qué datos desea mostrar al público).
* **Lectura (Directorio):** DTOs optimizados para visualización masiva (MiniProfile) o detallada (FullProfile), implementando el filtrado dinámico de datos privados.

### 2. CMS y Noticias (`request_posts.go`)

Gestiona el ciclo de vida de las comunicaciones institucionales:

* Soporta estados de publicación: `draft`, `published`, `archived`, `scheduled`.
* Implementa RBAC a nivel de contenido: segmentación entre noticias públicas y exclusivas para agremiados.

### 3. Seguridad y Staff (`requests_admin.structs.go`)

Maneja la administración interna del Colegio:

* **Matriz de Permisos (ACL):** Control granular sobre quién puede crear psicólogos, editar noticias o ver métricas.
* Usa el principio de **Menor Privilegio** para inicializar nuevos administradores.

### 4. Áreas de Desempeño y Redes Sociales (`specialty_requests.go`, `social_media.go`)

Controla los catálogos maestros:

* **Áreas de Desempeño:** Evita la duplicidad de términos científicos en el catálogo.
* **Social Media:** Normalización de URLs y visibilidad de perfiles sociales (Instagram, LinkedIn, etc.).

### 5. Sanitización y Utilidades (`directory_filter_sanitizer.go`, `response_request.go`)

Contiene la lógica de seguridad para búsquedas:

* **Limpieza de Texto:** Previene inyecciones en cláusulas `ILIKE`.
* **DoS Prevention:** Aplica "Hard Limits" a la paginación (máx. 100 registros por query).
* **Métricas:** Estructuras ligeras para el conteo de estadísticas en el Dashboard.

---

## 🛠️ Detalles de Implementación Técnica

### Sanitización de Búsqueda

El método `cleanSearchString` utiliza `[]rune` en lugar de `bytes`. Esto es vital para el español, ya que caracteres como la **"ñ"** ocupan múltiples bytes; un truncamiento simple por bytes podría corromper el carácter Unicode.

```go
// Ejemplo de lógica aplicada
runes := []rune(result)
if len(runes) > maxLen {
    result = string(runes[:maxLen])
}
```

### El patrón "Raw Fields"

Debido a las limitaciones del protocolo HTTP al enviar formularios con archivos, los booleanos suelen dar problemas. Este módulo lo resuelve capturando el valor como string y procesándolo en el backend:

```go
func (r *UpdateRequest) ShowEmail() *bool {
    return utils.BoolFromForm(r.ShowEmailRaw) // Convierte "1", "true", "on" -> true
}
```

---

## 🛡️ Seguridad

* **Validación de Tags:** Se utiliza `validator/v10` para aplicar reglas como `required`, `email`, `oneof`, y `datetime`.
* **Capa de Proyección:** Los DTOs de respuesta aseguran que la contraseña (`Password`) y la clave de sesión (`Key`) nunca se serialicen en los JSON de salida mediante el tag `json:"-"`.

---

*Este módulo es mantenido por el equipo técnico de COLPSI Carabobo.*

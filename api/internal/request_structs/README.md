# 📝 Estructuras de Request (request_structs/)

> **[⬆ internal](../)** — `api/internal/request_structs/`
>
> **Anti-Corruption Layer** — DTOs que validan y sanitizan TODOS los datos entrantes.

## Design Principles

- **Pointer-based PATCH semantics** (`*string`, `*int`) — `nil` = no cambiar, vacío = limpiar
- **Privacy by design** — campos sensibles solo en structs internos
- **Multipart/form-data support** para subida de imágenes
- **Unicode sanitization** para todos los campos de texto

## Archivos

| Archivo | Descripción |
|---------|-------------|
| `psi_user.go` | `CreatePsiUserRequest`, `UpdatePsiUserRequest` (multipart), `GetPsiUserParams` |
| `psi_user_Admin_requests.go` | Operaciones admin a nivel de psicólogos |
| `request_posts.go` | `CreatePostRequest`, `UpdatePostRequest`, `GetPostsParams` |
| `specialty_requests.go` | `CreateSpecialtyRequest`, `UpdateSpecialtyRequest` |
| `social_media.go` | CRUD de redes sociales |
| `response_request.go` | Structs de respuesta estándar |
| `requests_admin.structs.go` | Auth admin (`LoginRequest`, `RegisterRequest`), CRUD admin |
| `directory_filter_sanitizer.go` | Sanitización de filtros de búsqueda en directorio |

## Validación

Usa tags de **validator/v10**: `required`, `min`, `max`, `email`, `oneof`

## Seguridad

Proyección JSON con `json:"-"` para campos internos que nunca deben exponerse al cliente.

**[⬆ Volver a internal](../)**

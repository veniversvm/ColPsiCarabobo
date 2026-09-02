# 🔗 Handlers HTTP (handler/)

> **[⬆ internal](../)** — `api/internal/handler/`
>
> Capa de interfaz HTTP. Cada handler traduce peticiones HTTP en llamadas a servicios y mapea las respuestas a JSON. Usa **Fiber v2** como framework HTTP.

## Arquitectura

Los handlers son el punto de entrada del sistema. Cada uno recibe su servicio inyectado por constructor (Dependency Injection) y sigue un patrón consistente:

```
HTTP Request → Handler → Parse → Validate → Service → Map Response → HTTP Response
```

Los handlers **NO** contienen lógica de negocio — solo orquestación.

---

## 📂 Archivos y Controladores

### `psi_handler.go` — PsiHandler

El controlador más grande del sistema. Gestiona el ciclo de vida completo de los psicólogos: autenticación, perfil público, directorio y autogestión.

| Método | Endpoint | Auth | Descripción |
|---|---|---|---|
| `Login` | `POST /psi/login` | No | Autentica psicólogo, retorna JWT |
| `LoginLibrary` | `POST /psi/login-library` | No | Login con acceso limitado a biblioteca |
| `Logout` | `POST /psi/logout` | Bearer (psi) | Invalida token, elimina sesión activa |
| `GetMe` | `GET /psi/me` | Bearer (psi) | Perfil completo del psicólogo autenticado |
| `UpdateOwnProfile` | `PATCH /psi/me` | Bearer (psi) | Autogestión: actualizar perfil + imágenes |
| `SearchDirectory` | `GET /psi/directory` | No | Directorio público con filtros avanzados |
| `GetPublicProfile` | `GET /psi/:id` | No | Ficha pública por FPV (con privacy shield) |
| `GetSitemapData` | `GET /psi/sitemap` | No | Datos para sitemap XML (SEO) |
| `UploadCsv` | `POST /admin/psi/upload-csv` | Bearer (admin) | Importación masiva desde CSV |
| `AddPostGrade` | `POST /psi/me/postgrades` | Bearer (psi) | Agregar postgrado con documentos |
| `UpdatePostGrade` | `PATCH /psi/me/postgrades/:id` | Bearer (psi) | Actualizar postgrado existente |
| `AddSocialNetwork` | `POST /psi/me/social` | Bearer (psi) | Agregar red social |
| `UpdateSocialNetwork` | `PATCH /psi/me/social/:id` | Bearer (psi) | Actualizar red social |
| `DeleteSocialNetwork` | `DELETE /psi/me/social/:id` | Bearer (psi/admin) | Soft delete de red social |

**Constructor:** `NewPsiHandler(svc *service.PsiService, analytics *service.AnalyticsService)`

---

### `psi_user_admin.go` — Operaciones Admin de Psicólogos

Endpoints exclusivos para administradores con acceso total a los registros de psicólogos (sin filtros de privacidad).

| Método | Endpoint | Auth | Descripción |
|---|---|---|---|
| `GetPsiByIDAdmin` | `GET /admin/psi/:id` | Bearer (admin) | Expediente completo: identidad + col_data + bio + solvencias + RRSS |
| `CreatePsiByAdmin` | `POST /admin/psi/create` | Bearer (admin) | Registro manual individual |
| `UpdatePsiByAdmin` | `PATCH /admin/psi/:id` | Bearer (admin) | Edición total de cualquier campo (PATCH) |
| `DeletePsiByAdmin` | `DELETE /admin/psi/:id` | Bearer (admin) | Soft delete con auditoría |
| `ListAllPsis` | `GET /admin/psi/list` | Bearer (admin) | Listado "Rayos X" — ignora solvencia e inactivos |

Endpoints adicionales de administración sobre psicólogos (observaciones
internas y cumpleaños):

| Método | Endpoint | Auth | Descripción |
|---|---|---|---|
| `GetObservacionesByAdmin` | `GET /admin/psi/:id/observaciones` | Bearer (admin) | Lista observaciones internas del psicólogo |
| `AddObservacionByAdmin` | `POST /admin/psi/:id/observaciones` | Bearer (admin) | Crea una observación interna |
| `UpdateObservacionByAdmin` | `PATCH /admin/psi/:id/observaciones/:entryId` | Bearer (admin) | Edita una observación interna |
| `GetBirthdaysByAdmin` | `GET /admin/psi/birthdays` | Bearer (admin) | Cumpleaños (opt-in) de hoy o de la semana `?range=today\|week` |

---

### `posts_handler.go` — PostHandler

Gestión del CMS (sistema de publicaciones). Soporta ciclo de vida completo: draft → scheduled → published → archived.

| Método | Endpoint | Auth | Descripción |
|---|---|---|---|
| `ListPosts` | `GET /posts` | No / psi / admin | Listado adaptativo según rol |
| `GetPost` | `GET /posts/:id` | No / psi / admin | Detalle con contenido HTML y ACL |
| `CreatePost` | `POST /admin/posts` | Bearer (admin) | Crear post con imagen + status |
| `UpdatePost` | `PATCH /admin/posts/:id` | Bearer (admin) | Edición parcial + cambio de status |
| `GetSiteMapHandler` | `GET /posts/sitemap` | No | Datos para sitemap XML |

**Constructor:** `NewPostHandler(svc *service.PostService)`

**Visibilidad por rol:**
- `public` → solo posts con status=publicados y type=public
- `psi` → posts públicos + type=psi (gremiales)
- `admin` → todos los estados

---

### `admin_handler.go` — AdminHandler

Autenticación y gestión de personal administrativo. Sistema RBAC con verificación de jerarquía.

| Método | Endpoint | Auth | Descripción |
|---|---|---|---|
| `Login` | `POST /auth/login` | No | Login administrativo, retorna JWT |
| `CreateAdmin` | `POST /admin/create` | Bearer (admin) | Crear nuevo admin (verifica jerarquía) |
| `GetAdmins` | `GET /admin/list` | Bearer (admin) | Listado paginado con búsqueda |
| `GetAdminByID` | `GET /admin/:id` | Bearer (admin) | Detalle de administrador |
| `UpdateAdmin` | `PATCH /admin/update` | Bearer (admin) | Modificar datos y permisos |
| `DeleteAdmin` | `DELETE /admin/delete/:id` | Bearer (admin) | Soft delete (sin auto-eliminación ni borrar SUDO) |

**Constructor:** `NewAdminHandler(svc *service.AdminService)`

---

### `specialty_handler.go` — SpecialtyHandler

Catálogo de especialidades psicológicas. Soporta soft-delete (desactivación lógica).

| Método | Endpoint | Auth | Descripción |
|---|---|---|---|
| `GetSpecialties` | `GET /specialties` | No / admin | Catálogo filtrado por estado |
| `GetSpecialtyByID` | `GET /specialties/:id` | No | Detalle de especialidad |
| `CountSpecialties` | `GET /specialties/count` | No / admin | Total con filtro opcional |
| `CreateSpecialty` | `POST /admin/specialties` | Bearer (admin) | Crear nueva especialidad |
| `UpdateSpecialty` | `PATCH /admin/specialties/:id` | Bearer (admin) | Modificar especialidad |
| `DeleteSpecialty` | `DELETE /admin/specialties/:id` | Bearer (admin) | Desactivar (soft-delete) |
| `GetAllAdmin` | `GET /admin/specialties/all` | Bearer (admin) | Catálogo completo sin filtros |

**Constructor:** `NewSpecialtyHandler(svc *service.SpecialtyService)`

---

### `analytics_handler.go` — AnalyticsHandler

Dashboard y reportes de métricas del sistema.

| Método | Endpoint | Auth | Descripción |
|---|---|---|---|
| `GetDashboardStats` | `GET /admin/dashboard/stats` | Bearer (admin) | Estadísticas agregadas del dashboard |

**Constructor:** `NewAnalyticsHandler(svc *service.AnalyticsService)`

---

## 🔄 Flujo Request → Response

### Flujo típico: Ver perfil público de psicólogo

```
1. GET /psi/12345
2. PsiHandler.GetPublicProfile(c)
   ├─ Parse: c.ParamsInt("id") → extrae FPV number
   ├─ Service: h.service.GetPublicProfile(ctx, fvp)
   │   ├─ Repository: GetByFPV(ctx, 12345)
   │   ├─ Aplica privacy shield (filtra campos según show_* flags)
   │   └─ Retorna PsiFullProfileDTO
   ├─ Analytics: h.analytics.RecordProfileView(psi_id, viewerID, sid, ip)
   └─ Response: 200 + JSON del perfil filtrado
```

### Flujo típico: Login de administrador

```
1. POST /auth/login  { "identifier": "admin", "password": "secret" }
2. AdminHandler.Login(c)
   ├─ Parse: c.BodyParser(&req) → LoginRequest{Identifier, Password}
   ├─ Service: h.service.Login(ctx, identifier, password)
   │   ├─ Repository: GetByIdentifier(ctx, "admin")
   │   ├─ bcrypt.CompareHashAndPassword(password, hash)
   │   ├─ Repository: UpdateKey(ctx, user) → rota semilla JWT
   │   └─ Genera JWT con claims: id, role, key
   └─ Response: 200 + { "message": "Bienvenido al sistema", "token": "eyJ..." }
```

### Flujo típico: Crear publicación

```
1. POST /admin/posts (multipart/form-data)
   ├─ title, short_description, content, type, status, publish_at, image
2. PostHandler.CreatePost(c)
   ├─ Auth: c.Locals("admin") → verifica Bearer token
   ├─ Parse: c.BodyParser(&req) + c.FormFile("image")
   ├─ Service: h.service.CreatePost(ctx, admin, req, file)
   │   ├─ Sube imagen a S3 → obtiene key
   │   ├─ Repository: Create(ctx, post, textModel) → transacción atómica
   │   └─ Retorna error o nil
   └─ Response: 201 + { "message": "Post creado exitosamente" }
```

### Flujo típico: Directorio público con analytics

```
1. GET /psi/directory?q=María&specialty=3&location=Valencia&page=1&limit=12
2. PsiHandler.SearchDirectory(c)
   ├─ Parse: c.QueryParser(&filter) → PsiDirectoryFilterDTO
   ├─ Sanitize: request_structs.SanitizeDirectoryFilter(filter)
   ├─ Service: h.service.GetPublicDirectory(ctx, filter)
   │   ├─ Repository: SearchDirectory(ctx, filter)
   │   │   ├─ Filtra por solvencia (solo solventes)
   │   │   ├─ Aplica búsqueda por nombre/CI/FPV
   │   │   ├─ Filtra por especialidad, ubicación, género
   │   │   └─ Paginación con conteo total
   │   └─ Retorna { data: [...], total, page, limit }
   ├─ Analytics: h.analytics.RecordSearch(query, specialty, location, ...)
   └─ Response: 200 + JSON con resultados
```

---

## 🔐 Autenticación y Autorización

Los handlers reciben los claims JWT decodificados por los middleware vía `c.Locals()`:

| Local Key | Tipo | Fuente |
|---|---|---|
| `"admin"` | `*domain.UserAdmin` | JWT middleware (admin) |
| `"psi_user"` | `*domain.PsiUserModel` | JWT middleware (psi) |
| `"userID"` | `uuid.UUID` | JWT middleware (analytics) |

**Patrón de verificación:**
```go
admin, ok := c.Locals("admin").(*domain.UserAdmin)
if !ok || admin == nil {
    return c.Status(fiber.StatusUnauthorized).JSON(...)
}
```

---

## 📤 Patrón de Respuesta

**Éxito:**
```json
// 200 OK
{ "data": [...], "total": 42, "page": 1, "limit": 12 }

// 201 Created
{ "message": "Recurso creado exitosamente" }
```

**Error:**
```json
// 400 Bad Request
{ "error": "JSON malformado" }

// 401 Unauthorized
{ "error": "Credenciales inválidas" }

// 403 Forbidden
{ "error": "Permisos insuficientes" }

// 404 Not Found
{ "error": "Registro no encontrado" }

// 500 Internal Server Error
{ "error": "Error interno del servidor" }
```

---

## 📁 Upload de Archivos

Los handlers de imágenes usan `multipart/form-data`:

| Endpoint | Campos de archivo |
|---|---|
| `PATCH /psi/me` | `profile_picture`, `title_image_one/two/three` |
| `POST /psi/me/postgrades` | `pic_one`, `pic_two`, `pic_three` |
| `POST /admin/posts` | `image` (portada) |
| `PATCH /admin/psi/:id` | `profile_picture`, `title_image_one/two/three` |

Las imágenes se suben a S3/MinIO. Solo se almacenan las S3 keys en la base de datos.

---

## 📊 Total de Endpoints

| Controlador | Endpoints | Públicos | Autenticados |
|---|---|---|---|
| PsiHandler | 15 | 4 | 11 |
| PostHandler | 5 | 3 | 2 |
| AdminHandler | 5 | 1 | 4 |
| SpecialtyHandler | 7 | 3 | 4 |
| AnalyticsHandler | 1 | 0 | 1 |
| **Total** | **33** | **11** | **22** |

**[⬆ Volver a internal](../)**

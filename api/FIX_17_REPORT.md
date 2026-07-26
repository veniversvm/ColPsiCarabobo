# Reporte FIX-17 — S3 Keys expuestas en API JSON

**Fecha:** 25 de Julio, 2026
**Severidad:** 🟠 ALTO

---

## Hallazgo

8 campos de S3 keys internas estaban expuestos en las respuestas JSON de la API:

| Campo | JSON tag | Ejemplo value (antes) |
|-------|----------|----------------------|
| `ProfilePictureS3Key` | `profile_picture_url` | `avatars/abc.jpg` |
| `TitleImageOneS3Key` | `title_image_one_url` | `titles/def.jpg` |
| `TitleImageTwoS3Key` | `title_image_two_url` | `titles/ghi.jpg` |
| `TitleImageThreeS3Key` | `title_image_three_url` | `titles/jkl.jpg` |
| `PicOneS3Key` | `pic_one_url` | `postgrades/mno.jpg` |
| `PicTwoS3Key` | `pic_two_url` | `postgrades/pqr.jpg` |
| `PicThreeS3Key` | `pic_three_url` | `postgrades/stu.jpg` |
| `ImageS3Key` | `image_url` | `posts/vwx.jpg` |

Los JSON tags decían `_url` pero los valores eran **keys internas de S3**, no URLs.
El frontend construía URLs por su cuenta con concatenación client-side en ~12 archivos.

---

## Solución implementada

### 1. S3Client — GetPublicURL

Nuevo método que construye la URL pública completa:

```go
func (s *S3Client) GetPublicURL(key string) string {
    if key == "" { return "" }
    return fmt.Sprintf("%s/%s/%s", appConfig.Envs.S3Endpoint, s.Bucket, key)
}
```

Resultado: `http://localhost:9000/colpsi-bucket/avatars/abc.jpg`

### 2. PsiService — ResolvePsiModelURLs + publicURL

- `ResolvePsiModelURLs(psi)` — convierte todas las S3 keys de un PsiUserModel (profile, titles, postgrades)
- `publicURL(key)` — wrapper nil-safe para tests que pasan s3Client=nil

### 3. PostService — resolvePostURLs

Convierte `ImageS3Key` de cada Post en URL pública.

### 4. Endpoints actualizados

| Endpoint | Dónde se resuelve |
|----------|-------------------|
| `GET /psi/directory` | `psi_service.go` — `GetPublicDirectory()` |
| `GET /psi/{fpv}` | `psi_service.go` — `GetPublicProfile()` |
| `GET /admin/psi/{id}` | `psi_user_admin_service.go` — `GetPsiByIDAdmin()` |
| `GET /psi/me` | `psi_handler.go` — `GetMe()` |
| `GET /posts` | `post_service.go` — `GetPostsList()` |
| `GET /posts/{id}` | `post_service.go` — `GetPostByID()` |

### Ejemplo de respuesta (antes vs después)

**Antes:**
```json
{
    "profile_picture_url": "avatars/abc.jpg",
    "title_image_one_url": "titles/def.jpg"
}
```

**Después:**
```json
{
    "profile_picture_url": "http://localhost:9000/colpsi-bucket/avatars/abc.jpg",
    "title_image_one_url": "http://localhost:9000/colpsi-bucket/titles/def.jpg"
}
```

---

## Archivos modificados

| Archivo | Cambio |
|---------|--------|
| `pkg/s3/s3.go` | Nuevo método `GetPublicURL()` |
| `internal/service/psi_service.go` | `ResolvePsiModelURLs()` + `publicURL()` + mapeo en DTOs |
| `internal/service/psi_user_admin_service.go` | Llamada a `ResolvePsiModelURLs()` en `GetPsiByIDAdmin` |
| `internal/service/post_service.go` | `resolvePostURLs()` + mapeo en list/get |
| `internal/handler/psi_handler.go` | Llamada a `ResolvePsiModelURLs()` en `GetMe` |

---

## Tests

| Paquete | Resultado |
|---------|-----------|
| `internal/service` | ✅ Todos pasan (incluyendo test con s3Client=nil) |
| `internal/domain` | ✅ PASS |
| `internal/middleware` | ✅ PASS |
| `pkg/job` | ✅ PASS |

**0 regresiones.**

---

## Frontend — Acción requerida

El frontend actualmente construye URLs así:
```tsx
<img src={`http://localhost:9000/colpsi-bucket/${post.image_url}`} />
```

Con este cambio, `image_url` ya es una URL completa. El frontend debe:
1. Eliminar la concatenación del bucket base
2. Usar directamente: `<img src={post.image_url} />`
3. Centralizar `imgUrl` en un solo archivo compartido

**Nota:** Los ~12 archivos del frontend que hardcodean `localhost:9000/colpsi-bucket` deben actualizarse.

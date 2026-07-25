# FIX-12 Report — Sentinel errors reemplazan string matching

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-12 |
| **Hallazgo original** | HIGH-05 |
| **Archivos modificados** | 13 archivos |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

Se comparaban errores por string matching (`err.Error() == "texto"`) en handlers. Si alguien cambiaba el mensaje de error en un service, el handler fallaba silenciosamente.

---

## Corrección

### Nuevo archivo: `internal/domain/errors.go`

14 sentinel errors para autenticación, autorización, recursos y negocio.

### Services — retornar sentinel errors

| Service | Error anterior | Sentinel |
|---------|---------------|----------|
| `psi_service.go:316` | `errors.New("contraseña actual incorrecta")` | `domain.ErrPasswordIncorrect` |
| `psi_service.go:1275` | `errors.New("no tienes permiso...")` | `domain.ErrPermissionDenied` |
| `psi_service.go:787` | `errors.New("psicólogo no encontrado")` | `domain.ErrPsiNotFound` |
| `psi_user_admin_service.go:41` | `errors.New("permisos insuficientes...")` | `domain.ErrInsufficientPerms` |
| `psi_user_admin_service.go:46` | `errors.New("psicólogo no encontrado")` | `domain.ErrPsiNotFound` |
| `social_media.go:90` | `errors.New("no tienes permiso...")` | `domain.ErrSocialPermDenied` |
| `social_media.go:127` | `errors.New("no puedes borrar...")` | `domain.ErrSocialOwnDenied` |
| `post_service.go:56` | `errors.New("no tienes permiso...")` | `domain.ErrPostPermDenied` |
| `admin_service.go:306` | `errors.New("ya existe un usuario SUDO")` | `domain.ErrSudoExists` |
| `specialty_service.go:50,80,117` | `errors.New("no tienes permiso...")` | `domain.ErrInsufficientPerms` |

### Handlers — usar `errors.Is()`

| Handler | Antes | Después |
|---------|-------|---------|
| `psi_handler.go:120` | `err.Error() == "contraseña..."` | `errors.Is(err, domain.ErrPasswordIncorrect)` |
| `psi_handler.go:426` | `err.Error() == "no tienes permiso..."` | `errors.Is(err, domain.ErrPermissionDenied)` |
| `psi_user_admin.go:43` | `err.Error() == "psicólogo no encontrado"` | `errors.Is(err, domain.ErrPsiNotFound)` |
| `specialty_handler.go:95` | `strings.Contains(err.Error(), "permiso")` | `errors.Is(err, domain.ErrInsufficientPerms)` |

### Tests — usar `errors.Is()`

| Test | Antes | Después |
|------|-------|---------|
| `specialty_service_test.go:108` | `err.Error() != tt.errMsg` | `errors.Is(err, tt.errIs)` |
| `social_media_test.go:115` | `err.Error() != "no tienes permiso..."` | `errors.Is(err, domain.ErrSocialPermDenied)` |
| `social_media_test.go:150` | `err.Error() != "no puedes borrar..."` | `errors.Is(err, domain.ErrSocialOwnDenied)` |
| `post_service_test.go:184` | `err.Error() != "no tienes permiso..."` | `errors.Is(err, domain.ErrPostPermDenied)` |

---

## Testing

- `go build ./...` — Pass
- `go vet ./...` — Pass
- Middleware tests: 15/15 Pass
- Utils tests: 4/4 Pass
- Service tests: pasan (2 failures pre-existentes no relacionados)

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/domain/errors.go` | NUEVO — sentinel errors |
| `internal/handler/psi_handler.go` | 2 puntos |
| `internal/handler/psi_user_admin.go` | 1 punto |
| `internal/handler/specialty_handler.go` | 1 punto + eliminado `strings` import |
| `internal/service/psi_service.go` | 3 returns |
| `internal/service/psi_user_admin_service.go` | 2 returns |
| `internal/service/social_media.go` | 2 returns |
| `internal/service/post_service.go` | 1 return |
| `internal/service/admin_service.go` | 1 return |
| `internal/service/specialty_service.go` | 3 returns |

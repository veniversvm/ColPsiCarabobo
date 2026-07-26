# FIX-30: `Save()` → `Updates()` para prevenir zero-value overwrites

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-30 |
| **Hallazgo** | MED-11 |
| **Severidad** | MEDIO |
| **Archivos** | `user_admin_repo.go`, `specialty_repo.go`, `psi_repository.go`, `post_repo.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

GORM `db.Save()` reemplaza **TODOS** los campos de una fila con los valores del modelo proporcionado, incluyendo zero-values. Cuando un caller pasa un modelo con campos no inicializados (ej. `IsActive: false` porque no se tocó, no porque el usuario lo desactivó), `Save()` persiste `false` en la base de datos, borrando el valor real.

Esto afecta especialmente a:
- **Campos booleanos**: `false` es zero-value → `Save()` lo persiste aunque el valor real sea `true`
- **Campos int**: `0` es zero-value → `Save()` sobreescribe `GraduationYear`
- **Campos string**: `""` es zero-value → riesgo menor (vacío generalmente es intencional)

---

## Llamadas `Save()` corregidas (8 de 9)

| # | Archivo | Método | Riesgo | Cambio |
|---|---------|--------|--------|--------|
| 1 | `user_admin_repo.go:77` | `Update` | **ALTO** — 14+ booleanos de permisos | `Save(user)` → `Updates(map)` con `gorm.Expr` para cada booleano |
| 2 | `specialty_repo.go:104` | `Update` | **ALTO** — `Active` bool | `Save(s)` → `Updates(map)` con `gorm.Expr` para `Active` |
| 3 | `psi_repository.go:181` | `Update` (bioText) | MEDIO — `Content` string | `Session+Save(bioText)` → `Model(bioText).Updates(map)` |
| 4 | `psi_repository.go:367` | `UpdatePublicProfile` (bioText) | MEDIO — `Content` string | `Save(bioText)` → `Model(bioText).Updates(map)` |
| 5 | `psi_repository.go:699` | `UpdatePostGrade` | **ALTO** — `Active` bool, `GraduationYear` int | `Save(pg)` → `Updates(map)` con `gorm.Expr` para `Active` |
| 6 | `psi_repository.go:751` | `UpdateSocialNetwork` | **ALTO** — `IsActive` bool | `Save(sn)` → `Updates(map)` con `gorm.Expr` para `IsActive` |
| 7 | `post_repo.go:134` | `Update` (Post) | **ALTO** — `Status`, `Type` strings | `Save(post)` → `Updates(map)` explícito |
| 8 | `post_repo.go:140` | `Update` (TextModel) | MEDIO — `Content` string | `Save(text)` → `Updates(map)` explícito |

### Llamada conservada

| # | Archivo | Método | Razón |
|---|---------|--------|-------|
| 9 | `psi_repository.go:707` | `CreateSolvency` | **Intencional upsert** — `Save()` es correcto aquí porque crea o reemplaza la solvencia |

---

## Técnica: `gorm.Expr` para booleanos

```go
// ANTES (Save sobreescribe todos los campos incluyendo zero-values):
return r.db.WithContext(ctx).Save(user).Error

// DESPUÉS (Updates con mapa explícito + gorm.Expr para booleanos):
return r.db.WithContext(ctx).Model(user).Updates(map[string]interface{}{
    "username":    user.Username,
    "email":       user.Email,
    "is_active":   gorm.Expr("?", user.IsActive),   // ← fuerza escritura de false
    "sudo":        gorm.Expr("?", user.Sudo),         // ← fuerza escritura de false
    "can_publish": gorm.Expr("?", user.CanPublish),   // ← fuerza escritura de false
    // ... todos los campos explícitos
}).Error
```

**Por qué `gorm.Expr`**: `Updates(map)` con un booleano `false` directamente (`"is_active": false`) sería ignorado por GORM porque `false` es zero-value. `gorm.Expr("?", false)` genera SQL literal `SET is_active = false` sin importar el valor.

---

## Pruebas

7 tests nuevos en `updates_safety_test.go` usando SQLite en memoria:

| Test | Valida |
|------|--------|
| `TestAdminRepo_Update_PreservesBooleanFalse` | 16 booleanos persistidos como `false` tras Update |
| `TestAdminRepo_Update_PreservesBooleanTrue` | Round-trip `false→true` preserva `true` |
| `TestSpecialtyRepo_Update_PreservesActiveFalse` | `Active=false` persistido correctamente |
| `TestPostRepo_Update_PreservesStatusAndType` | `Status` y `Type` strings persistidos |
| `TestPostRepo_Update_TextContent` | Contenido de TextModel actualizado vía Updates |
| `TestPsiRepo_UpdatePostGrade_PreservesActiveFalse` | `Active=false` y `Title` en PostGrade |
| `TestPsiRepo_UpdateSocialNetwork_PreservesIsActiveFalse` | `IsActive=false` en SocialNetwork |

---

## Verificación

- `go build ./...` → sin errores
- `go vet ./...` → sin warnings
- 7 tests nuevos → **PASS**
- Tests preexistentes → misma situación (2 fallos conocidos, 0 regressions)

---

## Archivos modificados

| Archivo | Líneas Cambiadas |
|---------|-----------------|
| `internal/repository/postgres/user_admin_repo.go` | `Update()`: 7→32 líneas |
| `internal/repository/postgres/specialty_repo.go` | `Update()`: 3→10 líneas |
| `internal/repository/postgres/psi_repository.go` | 4 métodos: bioText×2, PostGrade, SocialNetwork |
| `internal/repository/postgres/post_repo.go` | `Update()`: 14→30 líneas |
| `internal/repository/postgres/updates_safety_test.go` | **NUEVO** — 7 tests unitarios con SQLite |

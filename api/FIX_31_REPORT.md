# FIX-31: Agregar FK de áreas de trabajo a `psi_users`

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-31 |
| **Hallazgo** | MED-12 |
| **Severidad** | MEDIO |
| **Archivos** | `user.model.go`, `psi_repository.go`, `psi_user.go`, `psi_user_Admin_requests.go`, `psi_service_self_management.go`, `psi_user_admin_service.go`, `psi_service_directory.go`, `migrations/` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

Las especialidades (áreas de desempeño) de los psicólogos se almacenaban como strings crudos (`primary_work_area varchar(50)`) sin FK al catálogo `psi_specialty_models`. Las queries de búsqueda hacían una **doble consulta** (N+1):

1. `SELECT name FROM psi_specialty_models WHERE id = ?` — lookup del nombre
2. `WHERE primary_work_area = ? OR secondary_work_area = ?` — comparación de strings

Esto era frágil (cambiar el nombre de una especialidad rompía datos) e ineficiente.

---

## Solución

### Migración SQL (`20260725030000_fix31_specialty_fk.sql`)

```sql
-- 1. Agregar columnas nullable
ALTER TABLE psi_users ADD COLUMN primary_specialty_id INTEGER;
ALTER TABLE psi_users ADD COLUMN secondary_specialty_id INTEGER;

-- 2. Migrar datos existentes (strings → IDs)
UPDATE psi_users p SET primary_specialty_id = s.id FROM psi_specialty_models s WHERE p.primary_work_area = s.name;
UPDATE psi_users p SET secondary_specialty_id = s.id FROM psi_specialty_models s WHERE p.secondary_work_area = s.name;

-- 3. FK constraints (ON DELETE SET NULL)
ALTER TABLE psi_users ADD CONSTRAINT fk_psi_users_primary_specialty ...;
ALTER TABLE psi_users ADD CONSTRAINT fk_psi_users_secondary_specialty ...;
```

### Modelo Go (`user.model.go`)

```go
PrimarySpecialtyID   *uint32 `gorm:"column:primary_specialty_id" json:"primary_specialty_id,omitempty"`
SecondarySpecialtyID *uint32 `gorm:"column:secondary_specialty_id" json:"secondary_specialty_id,omitempty"`
```

### Query de búsqueda (el win principal)

```go
// ANTES (doble query — N+1):
r.db.Model(&domain.PsiSpecialtyModel{}).Select("name").Where("id = ?", filter.SpecialtyID).Scan(&specName)
query = query.Where("primary_work_area = ? OR secondary_work_area = ?", specName, specName)

// DESPUÉS (query directa por FK):
query = query.Where("primary_specialty_id = ? OR secondary_specialty_id = ?", filter.SpecialtyID, filter.SpecialtyID)
```

---

## Archivos modificados

| Archivo | Cambio |
|---------|--------|
| `internal/domain/user.model.go` | +2 campos FK (nullable) |
| `migrations/20260725030000_fix31_specialty_fk.sql` | **NUEVO** — ALTER TABLE + data migration + FK |
| `internal/repository/postgres/psi_repository.go` | Reemplazadas 2 queries N+1 por FK directa; +2 campos en SELECT y Update maps |
| `internal/request_structs/psi_user.go` | +2 campos en UpdateProfileSelfRequest y PsiFullProfileDTO |
| `internal/request_structs/psi_user_Admin_requests.go` | +2 campos en CreatePsiAdminRequest y UpdatePsiAdminRequest |
| `internal/service/psi_service_self_management.go` | Asignación de FKs en UpdateProfileSelf |
| `internal/service/psi_user_admin_service.go` | Asignación de FKs en update y create admin |
| `internal/service/psi_service_directory.go` | FKs en DTO del perfil público |
| `internal/repository/postgres/updates_safety_test.go` | +1 test (FK persist + filter) |

---

## Pruebas

- 8 tests SQLite unitarios → **PASS** (incluye nuevo `TestPsiUsers_SpecialtyFK_PersistAndFilter`)
- Build + vet → sin errores
- Tests preexistentes → misma situación (2 fallos conocidos, 0 regressions)

---

## Decisión de diseño

- **Two FKs** (primary + secondary) según feedback del usuario
- **Backwards compatible** — columnas string se mantienen sincronizadas
- **Frontend** — fuera de scope de este fix (backend only)
- **CSV/XLSX import** — fuera de scope (las áreas se setean manualmente por el usuario)

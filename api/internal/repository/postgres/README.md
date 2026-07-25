# 🗄️ Repositorio PostgreSQL (repository/postgres/)

> **Capa de acceso a datos** — implementa todas las interfaces del dominio usando GORM sobre PostgreSQL. Cada repositorio recibe `*gorm.DB` y expone métodos que la capa de servicio consume a través de interfaces.

## Arquitectura

```
┌─────────────────────────────────────────────────────────┐
│                   SERVICE LAYER                          │
│         (Llama a interfaces del dominio)                 │
└──────────────────────┬──────────────────────────────────┘
                       │ domain.PsiUserRepository
                       │ domain.PostRepository
                       │ domain.UserAdminRepository
                       │ domain.SpecialtyRepository
                       ▼
┌─────────────────────────────────────────────────────────┐
│             REPOSITORY/postgres/  (ESTE PAQUETE)         │
│                                                          │
│  ┌────────────────┐  ┌────────────────┐                 │
│  │ psiRepo        │  │ postRepo       │                 │
│  │ → psi_repository.go│ → post_repo.go│                 │
│  └────────────────┘  └────────────────┘                 │
│  ┌────────────────┐  ┌────────────────┐                 │
│  │ adminRepo      │  │ specialtyRepo  │                 │
│  │ → user_admin_repo.go│→ specialty_repo.go│             │
│  └────────────────┘  └────────────────┘                 │
│                                                          │
│  Patrón: Cada repo recibe *gorm.DB, retorna interfaz    │
│  Transacciones: db.Begin() → defer Rollback → ops → Commit│
│  Búsqueda: ILIKE unaccent() para texto en español       │
│  Privacidad: Select() para excluir columnas sensibles   │
└──────────────────────┬──────────────────────────────────┘
                       │ *gorm.DB (pool de conexiones)
                       ▼
┌─────────────────────────────────────────────────────────┐
│              PostgreSQL (GORM AutoMigrate)               │
└─────────────────────────────────────────────────────────┘
```

---

## 📁 Estructura de Archivos

```
internal/repository/postgres/
├── psi_repository.go      # CRUD de psicólogos, búsquedas, redes sociales, solvencias
├── post_repo.go           # CRUD de publicaciones CMS, sitemap, scheduled publish
├── user_admin_repo.go     # CRUD de administradores, sudo count, listados
├── specialty_repo.go      # CRUD de catálogo de especialidades
├── *_test.go              # Pruebas unitarias
└── README.md              # Este archivo
```

---

## 🔧 Patrón Base de Todos los Repositorios

### Constructor (Factory)

Cada repositorio sigue el mismo patrón de inicialización:

```go
type xxxRepo struct {
    db *gorm.DB
}

func NewXxxRepository(db *gorm.DB) domain.XxxRepository {
    return &xxxRepo{db: db}
}
```

- Recibe `*gorm.DB` (pool de conexiones inyectado desde `cmd/api/main.go`)
- Retorna la interfaz del dominio (no el struct concreto)
- **Nunca** se instancia directamente — siempre a través del constructor

### Patrón Transaccional

```go
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // 1. Operaciones...
    if err := tx.Create(...).Error; err != nil {
        return err  // Rollback automático
    }
    // 2. Más operaciones...
    return nil  // Commit automático
})
```

**Reglas:**
- `WithContext(ctx)` propaga timeouts y cancelaciones de la petición HTTP
- `defer tx.Rollback()` no es necesario — GORM lo maneja internamente
- Si任何`return err` ocurre dentro de la función, se ejecuta Rollback automáticamente
- Solo `return nil` al final ejecuta Commit

---

## 👤 PsiUserRepositoryImpl (`psi_repository.go`)

El repositorio más extenso — 824 líneas. Gestiona psicólogos, datos colegiales, solvencias, redes sociales, postgrados y búsquedas.

### CreateWithColData — Inserción Transaccional Atómica

```go
func (r *psiRepo) CreateWithColData(ctx, psi, colData, solvencies, postgrades) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 1. Crear bio vacía (satisface FK obligatoria)
        tx.Create(emptyBio)
        psi.BioTextID = emptyBio.ID

        // 2. Crear psicólogo
        tx.Create(psi)

        // 3. Vincular datos colegiales
        colData.PsiUserModelID = psi.ID
        tx.Create(colData)

        // 4. Crear solvencias (si existen)
        if solvencies.ID != uuid.Nil {
            tx.Create(&solvencies)
        }

        // 5. Crear postgrados (bulk insert)
        if len(postgrades) > 0 {
            tx.Create(&postgrades)
        }

        return nil
    })
}
```

**Orden crítico:** Bio → PsiUser → ColData → Solvencies → PostGrades. Cada paso depende del anterior por las foreign keys.

### GetByID — Eager Loading con Preload

```go
r.db.WithContext(ctx).
    Preload("ColData").
    Preload("PostGrades", func(db *gorm.DB) *gorm.DB {
        return db.Order("graduation_year DESC")
    }).
    Preload("SocialNetworks").
    Preload("FullBio").
    First(&psi, "id = ?", id)
```

**Relaciones cargadas:**
- `ColData` — datos colegiales (1:1)
- `PostGrades` — títulos de postgrado (1:N, ordenados por año descendente)
- `SocialNetworks` — redes sociales (1:N)
- `FullBio` — biografía extensa (1:1, TextModel)

### SearchDirectory — Búsqueda Pública con ILIKE unaccent

```go
concatName := "unaccent(COALESCE(first_name, '') || ' ' || COALESCE(second_name, '') || ' ' || COALESCE(last_name, '') || ' ' || COALESCE(second_last_name, ''))"

query = query.Where(
    r.db.Where(concatName+" ILIKE unaccent(?)", w).
        Or("CAST(ci AS TEXT) LIKE ?", w).
        Or("CAST(fpv AS TEXT) LIKE ?", w),
)
```

**Características:**
- **Tokenización multi-palabra:** "Francisco Hernandez" → ["Francisco", "Hernandez"]. Cada palabra se busca con AND lógico
- **unaccent():** Función de PostgreSQL que elimina tildes → "José" matchea "Jose"
- **Concatenación de 4 nombres:** firstName + secondName + lastName + secondLastName
- **Búsqueda numérica:** CI y FPV se buscan como strings con LIKE
- **Filtro de ubicación:** Respeta flags de privacidad (`show_municipality_carabobo = true`)
- **Ordenamiento:** Solventes primero, luego con foto, luego alfabético

### SearchAdmin — Búsqueda Administrativa sin Restricciones

A diferencia de `SearchDirectory`, este método:
- No filtra por solvencia
- No respeta flags de privacidad
- Incluye email y estado activo en la proyección
- Ordena por `created_at DESC` (más recientes primero)

### Update — Mapa de Campos Explícito

```go
updateMap := map[string]interface{}{
    "username":            psi.Username,
    "email":               psi.Email,
    "is_active":           gorm.Expr("?", psi.IsActive),
    "show_contact_email":  gorm.Expr("?", psi.ShowContactEmail),
    // ... 40+ campos
}
tx.Model(&domain.PsiUserModel{}).Where("id = ?", psi.ID).
    Omit("created_at", "create_by", "create_by_id").
    Updates(updateMap)
```

**Técnica:** `gorm.Expr("?", value)` para campos booleanos — fuerza a GORM a escribir `false` en la DB (sin esto, GORM ignora zero-values).

**Omit()** protege los campos de auditoría de creación — nunca se sobreescriben en actualizaciones.

### UpdatePublicProfile — Subconjunto Restringido

Similar a `Update()` pero con menos campos — el psicólogo no puede modificar CI, FPV, isActive, solvency. Incluye `Omit("ci", "fpv", "is_active", "solvent", ...)` como doble protección.

### Solvencias con OnConflict (Upsert)

```go
tx.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "psi_user_model_id"}, {Name: "date"}},
    DoUpdates: clause.AssignmentColumns([]string{"updated_at", "update_by", "update_by_id"}),
}).Create(&solvencies)
```

**Comportamiento:** Si ya existe una solvencia para el mismo psicólogo y fecha, actualiza los campos de auditoría en lugar de fallar o duplicar.

### ValidateUniqueCredentials

```go
func (r *psiRepo) ValidateUniqueCredentials(ctx, username, email, excludeID) error {
    // Busca por username (ILIKE, case-insensitive) excluyendo el propio registro
    // Busca por email (ILIKE, case-insensitive) excluyendo el propio registro
}
```

**Nota:** Usa `ILIKE` para que "Admin@Email.com" y "admin@email.com" se consideren duplicados.

### GetSitemapData — Proyección Mínima

```go
r.db.Select("first_name, last_name, fpv").
    Where("is_active = ? AND solvent = true").
    Find(&users)
```

Solo trae los campos necesarios para generar URLs del sitemap, optimizando transferencia.

---

## 📝 PostRepositoryImpl (`post_repo.go`)

Gestiona publicaciones del CMS con separación metadata/contenido.

### Create — Transacción de Dos Pasos

```go
func (r *postRepo) Create(ctx, post, text) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 1. Crear TextModel primero (para obtener ID autogenerado)
        tx.Create(text)
        // 2. Vincular TextID al Post y crear
        post.TextID = text.ID
        return tx.Create(post).Error
    })
}
```

**Orden:** Text primero porque Post necesita el `TextID` generated.

### GetByID — Preload de Text

```go
r.db.WithContext(ctx).Preload("Text").First(&post, "id = ?", id)
```

Carga el contenido extenso solo cuando se necesita (consulta individual, no listados).

### List — Filtros Acumulativos

```go
func (r *postRepo) List(ctx, filter, page, limit) ([]domain.Post, int64, error) {
    query := r.db.WithContext(ctx).Model(&domain.Post{})

    // Filtro por estado(s)
    if len(filter.Status) > 0 {
        query = query.Where("status IN ?", filter.Status)
    }

    // Filtro por tipo
    if filter.Type == "all_visible" {
        query = query.Where("type IN ?", []string{"public", "psi"})
    } else if filter.Type != "" {
        query = query.Where("type = ?", filter.Type)
    }

    // Búsqueda por título (ILIKE)
    if filter.Search != "" {
        query = query.Where("title ILIKE ?", "%"+filter.Search+"%")
    }

    // Conteo + paginación
    query.Count(&total)
    query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&posts)
}
```

**Importante:** NO carga `TextModel` en listados — optimización deliberada para no transferir HTML masivo.

### PublishScheduled — Worker Job

```go
func (r *postRepo) PublishScheduled(ctx) int64 {
    result := r.db.WithContext(ctx).
        Where("status = ? AND publish_at <= ?", domain.PostStatusScheduled, time.Now()).
        Updates(map[string]interface{}{
            "status":     domain.PostStatusPublished,
            "publish_at": nil,
        })
    return result.RowsAffected
}
```

Transiciona posts programados cuya fecha ya pasó. Retorno: cantidad de filas afectadas para logging.

### GetSitemapPosts — Proyección SEO

```go
r.db.Select("id, title, updated_at").
    Where("status = ?", "published").
    Where("type = ?", "public").
    Order("created_at DESC").
    Find(&posts)
```

Solo posts publicados de tipo público — excluye contenido gremial "psi".

---

## 🔐 UserAdminRepositoryImpl (`user_admin_repo.go`)

Repositorio ligero — 130 líneas. CRUD básico de administradores.

### GetByIdentifier — Login Flexible

```go
r.db.Where("username = ? OR email = ?", identifier, identifier).First(&admin)
```

Permite login con username O email — el usuario elige su credencial.

### List — Búsqueda ILIKE

```go
if search != "" {
    s := "%" + search + "%"
    query = query.Where("email ILIKE ? OR username ILIKE ?", s, s)
}
```

Búsqueda parcial case-insensitive sobre email y username.

### CountSudos — Safety Check

```go
r.db.Where("sudo = ? AND deleted_at IS NULL", true).Count(&count)
```

Cuenta administradores Sudo activos (excluye soft-deleted). Usado para validar que nunca se quede el sistema sin al menos un Sudo.

### Delete — Soft Delete

```go
r.db.Delete(&domain.UserAdmin{}, "id = ?", id)
```

GORM ejecuta `UPDATE users SET deleted_at = NOW() WHERE id = ?` gracias al campo `DeletedAt` heredado de `AuditModel`.

---

## 🏷️ SpecialtyRepositoryImpl (`specialty_repo.go`)

Catálogo de especialidades con protección de integridad referencial.

### GetAll — Filtro Tri-Estatal

```go
switch status {
case "active":
    query = query.Where("active = ?", true)
case "inactive":
    query = query.Where("active = ?", false)
// "all" u otro → sin filtro
}
query.Order("name asc")
```

Resultados siempre ordenados alfabéticamente para UX óptima en dropdowns.

### GetByID — Escudo Público

```go
func (r *specialtyRepo) GetByID(ctx, id, active) {
    query := r.db.Where("id = ?", id)
    if active {
        query = query.Where("active = ?", true)  // Solo activas
    }
    return query.First(&s).Error
}
```

Si `active = true`, fuerza la cláusula a nivel SQL — previene que endpoints públicos accedan a especialidades desactivadas adivinando IDs.

### GetByAdminID — Vista Admin

Sin filtro de active — acceso total para el panel de administración.

### Delete — Soft Delete por Mapa

```go
r.db.Model(&domain.PsiSpecialtyModel{}).
    Where("id = ?", id).
    Updates(map[string]interface{}{"active": false})
```

**No usa Delete()** — solo desactiva el flag `active`. Esto preserva la integridad referencial (los psicólogos asociados no quedan huérfanos).

### GetAllAdmin — Proyección sin Filtros

```go
r.db.Order("name asc").Find(&list)
```

Catálogo completo sin filtros de visibilidad — reservado para interfaces de administración.

---

## 🔍 Patrones SQL Clave

### ILIKE unaccent para Texto en Español

```sql
-- Búsqueda que ignore tildes y mayúsculas
WHERE unaccent(first_name) ILIKE unaccent('%josé%')
```

**Problema resuelto:** En español, "José" ≠ "Jose" en comparación exacta, pero funcionalmente son el mismo nombre. `unaccent()` elimina acentos antes de comparar.

### gorm.Expr para Booleanos

```go
"is_active": gorm.Expr("?", psi.IsActive),
```

**Problema resuelto:** GORM ignora zero-values (`false`, `""`, `0`) en `Updates()`. `gorm.Expr` fuerza la escritura del valor literal, incluyendo `false`.

### Preload para Eager Loading

```go
Preload("ColData").
Preload("PostGrades", func(db *gorm.DB) *gorm.DB {
    return db.Order("graduation_year DESC")
})
```

**Problema resuelto:** Sin Preload, GORM haría queries N+1. Con Preload, carga todas las relaciones en queries optimizadas.

### OnConflict (Upsert)

```go
tx.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "psi_user_model_id"}, {Name: "date"}},
    DoUpdates: clause.AssignmentColumns([]string{"updated_at", "update_by"}),
}).Create(&solvencies)
```

**Problema resuelto:** Insertar solvencias duplicadas. Si ya existe (misma psi + fecha), actualiza auditoría en lugar de fallar.

### Omit para Proteger Campos

```go
tx.Model(psi).Where("id = ?", psi.ID).
    Omit("created_at", "create_by", "create_by_id").
    Updates(updateMap)
```

**Problema resuelto:** Evita que la auditoría de creación se sobreescriba durante actualizaciones.

---

## 🔒 Seguridad a Nivel de Repositorio

| Mecanismo | Descripción |
|-----------|-------------|
| **Soft Delete** | Nunca se borran registros físicamente (preserva FK y auditoría) |
| **Omit de auditoría** | `created_at`, `create_by` protegidos contra sobreescritura |
| **gorm.Expr para bool** | Previene que `false` se interprete como "no modificar" |
| **WithContext** | Propaga timeouts del HTTP para prevenir queries infinitas |
| **Proyección Select** | En listados, solo selecciona columnas necesarias (reduce attack surface) |
| **Unique constraints** | Reforzadas a nivel DB (cedula, email, FPV únicos) |

---

## 🗃️ Tablas Principales

| Modelo GORM | Tabla PostgreSQL | Descripción |
|-------------|-----------------|-------------|
| `PsiUserModel` | `psi_users` | Psicólogos colegiados |
| `PsiUserColData` | `psi_user_col_data` | Datos colegiales (1:1 con psi) |
| `PsiUserPostGrade` | `psi_user_post_grades` | Títulos de postgrado (1:N) |
| `PsiUserSocialNetwork` | `psi_user_social_networks` | Redes sociales (1:N) |
| `PsiUSerSolvency` | `psi_user_solvencies` | Historial de solvencia (1:N) |
| `TextModel` | `text_models` | Contenido extenso (biografías, posts) |
| `Post` | `posts` | Publicaciones CMS |
| `UserAdmin` | `user_admins` | Administradores del sistema |
| `PsiSpecialtyModel` | `psi_specialty_models` | Catálogo de especialidades |
| `LoginEvent` | `login_events` | Log de auditoría de logins |
| `ActiveSession` | `active_sessions` | Sesiones activas en tiempo real |
| `PageView` | `page_views` | Vistas de página del portal |
| `SearchEvent` | `search_events` | Eventos de búsqueda del directorio |
| `ProfileView` | `profile_views` | Vistas de perfiles de psicólogos |

---

## 🔗 Dependencias

| Paquete | Uso |
|---------|-----|
| `gorm.io/gorm` | ORM principal |
| `gorm.io/gorm/clause` | OnConflict (upsert) |
| `github.com/google/uuid` | Identificadores únicos |
| `github.com/veniversvm/ColPsiCarabobo/api/internal/domain` | Interfaces y modelos |
| `github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs` | DTOs de filtro |

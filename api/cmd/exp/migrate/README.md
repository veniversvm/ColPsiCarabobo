# cmd/exp/migrate — Generador de Esquema para Atlas

> **[⬆ exp](../)** — `api/cmd/exp/migrate/`

Herramienta que genera el esquema SQL de referencia a partir de los modelos GORM, utilizado por **Atlas** para calcular migraciones automáticas.

## generateSchema()

La función principal `generateSchema()` realiza los siguientes pasos:

1. **Conexión SQLite temporal** — Abre una conexión a `:memory:` usando el dialecto SQLite de GORM
2. **Registro de modelos** — Registra todos los modelos del dominio con `db.AutoMigrate()`:
   - UserAdmin, PsiUserModel, PsiUserColData
   - TextModel, PsiSpecialtyModel, PsiUserSocialNetwork
   - PostGrade, PsiUserSolvency, PsiUserDeontologia
   - PsiUserAuditModel, UserAdminAuditModel
   - PageView, ProfileView, SearchEvent, LoginEvent, ActiveSession
3. **Exportación del esquema** — Ejecuta `db.Migrator().CurrentSchema()` y escribe el DDL resultante a stdout
4. **Limpieza** — Cierra la conexión SQLite temporal

## Integración con Atlas

El esquema generado se usa como referencia en la configuración de Atlas (`atlas.hcl`):

```hcl
diff {
  compare {
    schema "public" {
      from = "file://schema.sql"   ← Salida de esta herramienta
      to   = "docker://postgres"   ← Estado actual de la DB
    }
  }
}
```

**Workflow completo:**

```
Modificar modelos GORM en /domain/
        ↓
go run cmd/exp/migrate/main.go > schema.sql
        ↓
atlas migrate diff --env local
        ↓
atlas genera archivo de migración en migrations/
        ↓
atlas migrate apply --env local
```

**[⬆ Volver a exp](../)**

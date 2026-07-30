# cmd/exp — Herramientas Experimentales

> **[⬆ cmd](../)** — `api/cmd/exp/`

El directorio `cmd/exp/` contiene herramientas utilitarias y experimentales que no forman parte del servidor principal pero son esenciales para el workflow de desarrollo.

## Herramientas

### migrate/

Generador de esquema SQL para las migraciones de **Atlas**. Esta herramienta:

1. Crea una base de datos SQLite temporal en memoria
2. Registra todos los modelos del dominio (UserAdmin, PsiUserModel, TextModel, etc.)
3. Ejecuta `AutoMigrate` de GORM sobre el SQLite temporal
4. Vuelca el esquema resultante a stdout en formato SQL

**Flujo de trabajo con Atlas:**

```
Developer modifica modelos GORM
            ↓
go run cmd/exp/migrate/main.go > schema.sql
            ↓
atlas diff --to file://schema.sql
            ↓
Atlas genera el archivo de migración SQL
            ↓
Migración aplicada a PostgreSQL en staging/producción
```

**Uso:**
```bash
# Generar esquema de referencia
go run cmd/exp/migrate/main.go > schema.sql

# O usando make
make migrate-generate
```

> **Nota:** El SQLite se utiliza como base temporal porque GORM necesita un dialecto válido para generar el DDL. El esquema resultante es lo suficientemente cercano a PostgreSQL para que Atlas pueda calcular los diffs correctamente.

---

**Subdirectorios:**
| Directorio | Descripción |
|------------|-------------|
| [`migrate/`](./migrate/) | Generador de esquema SQL para Atlas |

**[⬆ Volver a cmd](../)**

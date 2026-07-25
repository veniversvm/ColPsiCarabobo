# 📦 Paquetes Compartidos (pkg/)

## Descripción

La capa `pkg/` contiene las **bibliotecas reutilizables externas** del proyecto. Estos paquetes proveen funcionalidad genérica — conexión a base de datos, almacenamiento en S3 — sin depender de ningún componente interno del negocio.

Cada paquete en `pkg/` es un módulo independiente que puede ser consumido por cualquier capa superior (`internal/`, `cmd/`), pero **nunca** importa desde `internal/`.

---

## 📁 Estructura

```
pkg/
├── database/          # Conexión PostgreSQL, seed y migraciones
│   ├── postgres.go
│   ├── seed.go
│   └── migration.go
└── s3/                # Cliente AWS S3/MinIO y subida de archivos
    ├── s3.go
    └── upload.go
```

---

## 📚 Tabla de Contenidos

| Paquete | Descripción | Archivos |
|---------|-------------|----------|
| [database](./database/) | Conexión GORM con PgBouncer, seeding de admin, AutoMigrate | `postgres.go`, `seed.go`, `migration.go` |
| [s3](./s3/) | Cliente AWS S3/MinIO y upload de imágenes con validación | `s3.go`, `upload.go` |

---

## 🏗️ Nota de Arquitectura

> **`pkg/` tiene CERO dependencias internas.**

```
┌─────────────────────────────────────────────────┐
│                  cmd/api/main.go                │
├─────────────────────────────────────────────────┤
│              internal/ (negocio)                │
│  handler / service / repository / model         │
├─────────────────────────────────────────────────┤
│               📦 pkg/ (compartido)              │
│       database/    s3/                          │
├─────────────────────────────────────────────────┤
│        Dependencias externas                    │
│   GORM · AWS SDK · PgBouncer · uuid            │
└─────────────────────────────────────────────────┘
```

- **Dependencias externas únicas**: GORM, AWS SDK v2, PgBouncer, generador de UUIDs.
- **Configuración**: Lee únicamente de `config/` (variables de entorno).
- **Regla estricta**: `pkg/` **NUNCA** ejecuta `import` desde `internal/`. Si necesitás lógica de negocio, ese código no pertenece a `pkg/`.

---

## 🚀 Uso

```go
// Conexión a base de datos
db := pkgdatabase.InitDatabase()
pkgdatabase.SeedAdminUsers(db)
pkgdatabase.InitMigrations(db)

// Upload de imagen a S3
url, err := pkgs3.UploadImage(file, "avatars")
```

---

## 🔒 Consideraciones

- Todas las credenciales (DB, S3) se leen **exclusivamente** de variables de entorno.
- La conexión a PostgreSQL usa **PgBouncer en modo transacción** para eficiencia de conexiones.
- El cliente S3 soporta tanto **AWS S3** como **MinIO** para desarrollo local.
- Los archivos subidos usan **nombres UUID** para prevenir path traversal y colisiones.
